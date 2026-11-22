// Package engine executes the workflow of a claimed run and records each step in run_steps (M-12).
package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polcortadellas/AREA/server/internal/node"
)

type Engine struct {
	pool     *pgxpool.Pool
	registry *node.Registry
}

func New(pool *pgxpool.Pool, registry *node.Registry) *Engine {
	return &Engine{pool: pool, registry: registry}
}

type run struct {
	id       uuid.UUID
	nodes    map[string]node.WorkflowNode
	children map[string][]node.Edge
	done     map[string]step // also holds the trigger under "trigger"
	recorded map[string]bool
}

// Execute walks the tree from the trigger. Nodes it never reaches, the branch
// not taken or everything after a failed step, are recorded as SKIPPED.
// An error makes the worker mark the run FAILED.
func (e *Engine) Execute(ctx context.Context, runID uuid.UUID) error {
	r := &run{id: runID, done: map[string]step{}, recorded: map[string]bool{}}
	var workflowID uuid.UUID
	var payload node.Values
	if err := e.pool.QueryRow(ctx,
		`SELECT workflow_id, trigger_payload FROM runs WHERE id = $1`, runID,
	).Scan(&workflowID, &payload); err != nil {
		return fmt.Errorf("load run: %w", err)
	}
	trigger, err := e.load(ctx, workflowID, r)
	if err != nil {
		return err
	}

	var schema node.Schema
	if t, ok := e.registry.Trigger(r.nodes[trigger].Type); ok {
		schema = t.OutputSchema()
	}
	r.done["trigger"] = step{outputs: payload, schema: schema}
	runErr := e.visit(ctx, r, trigger, payload, schema)

	for id := range r.nodes {
		if !r.recorded[id] {
			if err := e.record(ctx, r, id, "SKIPPED", nil, nil); err != nil {
				return err
			}
		}
	}
	return runErr
}

// visit runs node id (the trigger just records the payload it was given) and then its children.
func (e *Engine) visit(ctx context.Context, r *run, id string, triggerOut node.Values, triggerSchema node.Schema) error {
	n := r.nodes[id]
	next := r.children[id]

	var out node.Values
	var schema node.Schema
	var err error
	switch n.Kind {
	case node.KindTrigger:
		out, schema = triggerOut, triggerSchema
	case node.KindCondition:
		var config node.Values
		var ok bool
		if config, err = resolve(n.Config, r.done); err == nil {
			ok, err = evaluate(config)
		}
		out = node.Values{"result": ok}
		next = nil
		for _, edge := range r.children[id] {
			if edge.Branch == fmt.Sprint(ok) {
				next = append(next, edge)
			}
		}
	case node.KindReaction:
		reaction, found := e.registry.Reaction(n.Type)
		if !found {
			err = fmt.Errorf("unknown reaction %q", n.Type)
			break
		}
		schema = reaction.OutputSchema()
		var inputs node.Values
		if err := e.record(ctx, r, id, "RUNNING", nil, nil); err != nil {
			return err
		}
		if inputs, err = resolve(n.Config, r.done); err == nil {
			// shortcut: no credentials yet, reactions that need OAuth get them with M-14.
			out, err = reaction.Execute(ctx, inputs, node.Credentials{})
		}
	default:
		err = fmt.Errorf("unknown kind %q", n.Kind)
	}

	if err != nil {
		if recErr := e.record(ctx, r, id, "FAILED", nil, err); recErr != nil {
			return recErr
		}
		return fmt.Errorf("node %s: %w", id, err)
	}
	if err := e.record(ctx, r, id, "SUCCEEDED", out, nil); err != nil {
		return err
	}
	r.done[id] = step{outputs: out, schema: schema}

	for _, edge := range next {
		if err := e.visit(ctx, r, edge.To, nil, nil); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) load(ctx context.Context, workflowID uuid.UUID, r *run) (string, error) {
	rows, err := e.pool.Query(ctx,
		`SELECT id::text, kind, type, config FROM nodes WHERE workflow_id = $1`, workflowID)
	if err != nil {
		return "", fmt.Errorf("load nodes: %w", err)
	}
	r.nodes = map[string]node.WorkflowNode{}
	var trigger string
	for rows.Next() {
		var n node.WorkflowNode
		if err := rows.Scan(&n.ID, &n.Kind, &n.Type, &n.Config); err != nil {
			return "", fmt.Errorf("load nodes: %w", err)
		}
		r.nodes[n.ID] = n
		if n.Kind == node.KindTrigger {
			trigger = n.ID
		}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("load nodes: %w", err)
	}
	if trigger == "" {
		return "", fmt.Errorf("workflow %s has no trigger", workflowID)
	}

	rows, err = e.pool.Query(ctx,
		`SELECT from_node_id::text, to_node_id::text, coalesce(branch, '') FROM edges WHERE workflow_id = $1`, workflowID)
	if err != nil {
		return "", fmt.Errorf("load edges: %w", err)
	}
	r.children = map[string][]node.Edge{}
	for rows.Next() {
		var edge node.Edge
		if err := rows.Scan(&edge.From, &edge.To, &edge.Branch); err != nil {
			return "", fmt.Errorf("load edges: %w", err)
		}
		r.children[edge.From] = append(r.children[edge.From], edge)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("load edges: %w", err)
	}
	return trigger, nil
}

func (e *Engine) record(ctx context.Context, r *run, nodeID, status string, out node.Values, stepErr error) error {
	var output []byte
	if out != nil {
		var err error
		if output, err = json.Marshal(out); err != nil {
			return fmt.Errorf("encode output of %s: %w", nodeID, err)
		}
	}
	var errText *string
	if stepErr != nil {
		s := stepErr.Error()
		errText = &s
	}
	// shortcut: one attempt and no idempotency yet, F-04 retries and F-05 replays reuse this key.
	// A reaction is written RUNNING first and finished by the second call, which keeps started_at.
	if _, err := e.pool.Exec(ctx, `
		INSERT INTO run_steps (run_id, node_id, status, attempt, idempotency_key, output, error, started_at, finished_at)
		VALUES ($1, $2, $3::text,
		        CASE WHEN $3::text = 'SKIPPED' THEN 0 ELSE 1 END,
		        $6, $4, $5,
		        CASE WHEN $3::text = 'SKIPPED' THEN NULL ELSE now() END,
		        CASE WHEN $3::text IN ('RUNNING', 'SKIPPED') THEN NULL ELSE now() END)
		ON CONFLICT (run_id, node_id) DO UPDATE
		SET status = EXCLUDED.status, output = EXCLUDED.output, error = EXCLUDED.error, finished_at = EXCLUDED.finished_at`,
		r.id, nodeID, status, output, errText, r.id.String()+":"+nodeID); err != nil {
		return fmt.Errorf("record step %s: %w", nodeID, err)
	}
	r.recorded[nodeID] = true
	return nil
}
