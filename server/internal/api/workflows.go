// Package api holds the HTTP handlers of the api role.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polcortadellas/AREA/server/internal/auth"
	"github.com/polcortadellas/AREA/server/internal/node"
)

const (
	maxBody       = 1 << 20
	maxNameLen    = 100
	defaultRuns   = 20
	maxRunsPerReq = 100
)

type workflowInput struct {
	Name    string `json:"name"`
	Enabled *bool  `json:"enabled"`
	node.Workflow
}

type workflowSummary struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type workflowOut struct {
	workflowSummary
	node.Workflow
}

type runOut struct {
	ID             uuid.UUID      `json:"id"`
	WorkflowID     uuid.UUID      `json:"workflow_id"`
	Status         string         `json:"status"`
	TriggerPayload map[string]any `json:"trigger_payload"`
	Error          *string        `json:"error"`
	CreatedAt      time.Time      `json:"created_at"`
	StartedAt      *time.Time     `json:"started_at"`
	FinishedAt     *time.Time     `json:"finished_at"`
}

type stepOut struct {
	NodeID     string         `json:"node_id"`
	Status     string         `json:"status"`
	Attempt    int32          `json:"attempt"`
	Output     map[string]any `json:"output"`
	Error      *string        `json:"error"`
	StartedAt  *time.Time     `json:"started_at"`
	FinishedAt *time.Time     `json:"finished_at"`
}

type runDetail struct {
	runOut
	Steps []stepOut `json:"steps"`
}

// querier is a pool or a transaction.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func CreateWorkflow(pool *pgxpool.Pool, reg *node.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		in, ok := decodeWorkflow(w, r, reg)
		if !ok {
			return
		}
		userID, _ := auth.UserID(r.Context())
		enabled := in.Enabled == nil || *in.Enabled
		var id uuid.UUID
		err := pgx.BeginFunc(r.Context(), pool, func(tx pgx.Tx) error {
			if err := tx.QueryRow(r.Context(),
				`INSERT INTO workflows (user_id, name, enabled) VALUES ($1, $2, $3) RETURNING id`,
				userID, in.Name, enabled,
			).Scan(&id); err != nil {
				return err
			}
			return saveGraph(r.Context(), tx, id, in.Workflow)
		})
		if err != nil {
			internalError(w, "create workflow", err)
			return
		}
		writeWorkflow(w, r, pool, userID, id, http.StatusCreated)
	}
}

func GetWorkflow(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r, "workflowId")
		if !ok {
			return
		}
		userID, _ := auth.UserID(r.Context())
		writeWorkflow(w, r, pool, userID, id, http.StatusOK)
	}
}

// ReplaceWorkflow keeps the rows of nodes whose key is unchanged, so the run
// history of those nodes survives an edit.
func ReplaceWorkflow(pool *pgxpool.Pool, reg *node.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r, "workflowId")
		if !ok {
			return
		}
		in, ok := decodeWorkflow(w, r, reg)
		if !ok {
			return
		}
		userID, _ := auth.UserID(r.Context())
		err := pgx.BeginFunc(r.Context(), pool, func(tx pgx.Tx) error {
			tag, err := tx.Exec(r.Context(), `
				UPDATE workflows SET name = $3, enabled = coalesce($4, enabled), updated_at = now()
				WHERE id = $1 AND user_id = $2`, id, userID, in.Name, in.Enabled)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return errWorkflowNotFound
			}
			return saveGraph(r.Context(), tx, id, in.Workflow)
		})
		if errors.Is(err, errWorkflowNotFound) {
			notFound(w)
			return
		}
		if err != nil {
			internalError(w, "replace workflow", err)
			return
		}
		writeWorkflow(w, r, pool, userID, id, http.StatusOK)
	}
}

// UpdateWorkflow renames, pauses or resumes. The scheduler only polls enabled
// workflows, so pausing stops new runs.
func UpdateWorkflow(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r, "workflowId")
		if !ok {
			return
		}
		var patch struct {
			Name    *string `json:"name"`
			Enabled *bool   `json:"enabled"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&patch); err != nil || (patch.Name == nil && patch.Enabled == nil) ||
			(patch.Name != nil && !validName(*patch.Name)) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_body"})
			return
		}
		userID, _ := auth.UserID(r.Context())
		tag, err := pool.Exec(r.Context(), `
			UPDATE workflows SET name = coalesce($3, name), enabled = coalesce($4, enabled), updated_at = now()
			WHERE id = $1 AND user_id = $2`, id, userID, patch.Name, patch.Enabled)
		if err != nil {
			internalError(w, "update workflow", err)
			return
		}
		if tag.RowsAffected() == 0 {
			notFound(w)
			return
		}
		writeWorkflow(w, r, pool, userID, id, http.StatusOK)
	}
}

func DeleteWorkflow(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r, "workflowId")
		if !ok {
			return
		}
		userID, _ := auth.UserID(r.Context())
		tag, err := pool.Exec(r.Context(), `DELETE FROM workflows WHERE id = $1 AND user_id = $2`, id, userID)
		if err != nil {
			internalError(w, "delete workflow", err)
			return
		}
		if tag.RowsAffected() == 0 {
			notFound(w)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func ListWorkflows(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserID(r.Context())
		rows, err := pool.Query(r.Context(), `
			SELECT id, name, enabled, created_at, updated_at FROM workflows
			WHERE user_id = $1 ORDER BY created_at DESC`, userID)
		if err != nil {
			internalError(w, "list workflows", err)
			return
		}
		list, err := pgx.CollectRows(rows, pgx.RowToStructByPos[workflowSummary])
		if err != nil {
			internalError(w, "list workflows", err)
			return
		}
		writeJSON(w, http.StatusOK, orEmptyList(list))
	}
}

func ListRuns(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r, "workflowId")
		if !ok {
			return
		}
		limit := defaultRuns
		if s := r.URL.Query().Get("limit"); s != "" {
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 || n > maxRunsPerReq {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_limit"})
				return
			}
			limit = n
		}
		userID, _ := auth.UserID(r.Context())
		var owned bool
		if err := pool.QueryRow(r.Context(),
			`SELECT EXISTS (SELECT 1 FROM workflows WHERE id = $1 AND user_id = $2)`, id, userID,
		).Scan(&owned); err != nil {
			internalError(w, "list runs", err)
			return
		}
		if !owned {
			notFound(w)
			return
		}
		rows, err := pool.Query(r.Context(), `
			SELECT id, workflow_id, status, trigger_payload, error, created_at, started_at, finished_at
			FROM runs WHERE workflow_id = $1 ORDER BY created_at DESC LIMIT $2`, id, limit)
		if err != nil {
			internalError(w, "list runs", err)
			return
		}
		runs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[runOut])
		if err != nil {
			internalError(w, "list runs", err)
			return
		}
		writeJSON(w, http.StatusOK, orEmptyList(runs))
	}
}

func GetRun(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r, "runId")
		if !ok {
			return
		}
		userID, _ := auth.UserID(r.Context())
		rows, err := pool.Query(r.Context(), `
			SELECT r.id, r.workflow_id, r.status, r.trigger_payload, r.error, r.created_at, r.started_at, r.finished_at
			FROM runs r JOIN workflows w ON w.id = r.workflow_id
			WHERE r.id = $1 AND w.user_id = $2`, id, userID)
		if err != nil {
			internalError(w, "get run", err)
			return
		}
		run, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[runOut])
		if errors.Is(err, pgx.ErrNoRows) {
			notFound(w)
			return
		}
		if err != nil {
			internalError(w, "get run", err)
			return
		}
		rows, err = pool.Query(r.Context(), `
			SELECT n.key, s.status, s.attempt, s.output, s.error, s.started_at, s.finished_at
			FROM run_steps s JOIN nodes n ON n.id = s.node_id
			WHERE s.run_id = $1 ORDER BY s.started_at NULLS LAST, n.key`, id)
		if err != nil {
			internalError(w, "get run", err)
			return
		}
		steps, err := pgx.CollectRows(rows, pgx.RowToStructByPos[stepOut])
		if err != nil {
			internalError(w, "get run", err)
			return
		}
		writeJSON(w, http.StatusOK, runDetail{runOut: run, Steps: orEmptyList(steps)})
	}
}

// decodeWorkflow answers 400 or 422 itself and reports whether to go on.
func decodeWorkflow(w http.ResponseWriter, r *http.Request, reg *node.Registry) (workflowInput, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	var in workflowInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || !validName(in.Name) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_body"})
		return in, false
	}
	var invalid *node.InvalidWorkflowError
	if err := reg.ValidateWorkflow(in.Workflow); errors.As(err, &invalid) {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":    "invalid_workflow",
			"problems": invalid.Problems,
		})
		return in, false
	} else if err != nil {
		internalError(w, "validate workflow", err)
		return in, false
	}
	return in, true
}

func validName(name string) bool {
	return strings.TrimSpace(name) != "" && utf8.RuneCountInString(name) <= maxNameLen
}

// saveGraph makes the stored nodes and edges match wf. Nodes are matched by
// key: kept rows keep their run_steps, removed ones take theirs with them.
func saveGraph(ctx context.Context, tx pgx.Tx, workflowID uuid.UUID, wf node.Workflow) error {
	keys := make([]string, 0, len(wf.Nodes))
	for _, n := range wf.Nodes {
		keys = append(keys, n.ID)
		if n.Kind != node.KindTrigger {
			continue
		}
		// A changed trigger must not resume from the old trigger's cursor.
		if _, err := tx.Exec(ctx, `
			UPDATE workflows SET poll_cursor = NULL
			WHERE id = $1 AND NOT EXISTS (
				SELECT 1 FROM nodes WHERE workflow_id = $1 AND kind = 'trigger' AND type = $2 AND config = $3)`,
			workflowID, n.Type, configOf(n)); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM edges WHERE workflow_id = $1`, workflowID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`DELETE FROM nodes WHERE workflow_id = $1 AND NOT (key = ANY($2))`, workflowID, keys); err != nil {
		return err
	}
	rowIDs := make(map[string]uuid.UUID, len(wf.Nodes))
	for _, n := range wf.Nodes {
		var id uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO nodes (workflow_id, key, kind, type, config) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (workflow_id, key) DO UPDATE
			SET kind = EXCLUDED.kind, type = EXCLUDED.type, config = EXCLUDED.config
			RETURNING id`, workflowID, n.ID, n.Kind, n.Type, configOf(n)).Scan(&id); err != nil {
			return err
		}
		rowIDs[n.ID] = id
	}
	// ValidateWorkflow already checked that every edge joins two listed nodes.
	for _, e := range wf.Edges {
		if _, err := tx.Exec(ctx, `
			INSERT INTO edges (workflow_id, from_node_id, to_node_id, branch) VALUES ($1, $2, $3, NULLIF($4, ''))`,
			workflowID, rowIDs[e.From], rowIDs[e.To], e.Branch); err != nil {
			return err
		}
	}
	return nil
}

func configOf(n node.WorkflowNode) node.Values {
	if n.Config == nil {
		return node.Values{}
	}
	return n.Config
}

func loadWorkflow(ctx context.Context, q querier, userID, id uuid.UUID) (workflowOut, error) {
	var out workflowOut
	rows, err := q.Query(ctx, `
		SELECT id, name, enabled, created_at, updated_at FROM workflows WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return out, err
	}
	out.workflowSummary, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[workflowSummary])
	if errors.Is(err, pgx.ErrNoRows) {
		return out, errWorkflowNotFound
	}
	if err != nil {
		return out, err
	}
	rows, err = q.Query(ctx, `SELECT key, kind, type, config FROM nodes WHERE workflow_id = $1 ORDER BY key`, id)
	if err != nil {
		return out, err
	}
	if out.Nodes, err = pgx.CollectRows(rows, pgx.RowToStructByPos[node.WorkflowNode]); err != nil {
		return out, err
	}
	rows, err = q.Query(ctx, `
		SELECT f.key, t.key, coalesce(e.branch, '')
		FROM edges e JOIN nodes f ON f.id = e.from_node_id JOIN nodes t ON t.id = e.to_node_id
		WHERE e.workflow_id = $1 ORDER BY f.key, t.key`, id)
	if err != nil {
		return out, err
	}
	if out.Edges, err = pgx.CollectRows(rows, pgx.RowToStructByPos[node.Edge]); err != nil {
		return out, err
	}
	out.Nodes, out.Edges = orEmptyList(out.Nodes), orEmptyList(out.Edges)
	return out, nil
}

func writeWorkflow(w http.ResponseWriter, r *http.Request, q querier, userID, id uuid.UUID, status int) {
	wf, err := loadWorkflow(r.Context(), q, userID, id)
	if errors.Is(err, errWorkflowNotFound) {
		notFound(w)
		return
	}
	if err != nil {
		internalError(w, "load workflow", err)
		return
	}
	writeJSON(w, status, wf)
}

// pathID answers 404 for a malformed ID, like an unknown one.
func pathID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		notFound(w)
		return uuid.Nil, false
	}
	return id, true
}

// orEmptyList keeps empty lists as [] in JSON instead of null.
func orEmptyList[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func notFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
}

func internalError(w http.ResponseWriter, op string, err error) {
	slog.Error("api: "+op, "err", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
