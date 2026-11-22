package api_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polcortadellas/AREA/server/internal/api"
	"github.com/polcortadellas/AREA/server/internal/auth"
	"github.com/polcortadellas/AREA/server/internal/db"
)

type fakeQueue struct {
	runs []uuid.UUID
	err  error
}

func (q *fakeQueue) Enqueue(_ context.Context, id uuid.UUID) error {
	q.runs = append(q.runs, id)
	return q.err
}

func triggerPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	base := os.Getenv("DATABASE_URL")
	if base == "" {
		t.Skip("DATABASE_URL not set (use a throwaway PostgreSQL)")
	}
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)

	b := make([]byte, 6)
	_, _ = rand.Read(b)
	name := "area_trigger_test_" + hex.EncodeToString(b)
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		conn, err := pgx.Connect(context.Background(), base)
		if err == nil {
			_, _ = conn.Exec(context.Background(), "DROP DATABASE "+name+" WITH (FORCE)")
			_ = conn.Close(context.Background())
		}
	})

	u, _ := url.Parse(base)
	u.Path = "/" + name
	if err := db.Migrate(u.String()); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func newWorkflow(t *testing.T, pool *pgxpool.Pool) (workflowID, userID uuid.UUID) {
	t.Helper()
	ctx := t.Context()
	if err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, 'x') RETURNING id`, uuid.NewString()+"@example.com",
	).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO workflows (user_id, name) VALUES ($1, 'demo') RETURNING id`, userID,
	).Scan(&workflowID); err != nil {
		t.Fatal(err)
	}
	return workflowID, userID
}

func serveTrigger(t *testing.T, pool *pgxpool.Pool, q api.RunQueue, id string, user uuid.UUID) *httptest.ResponseRecorder {
	t.Helper()
	tokens, err := auth.NewTokens(strings.Repeat("k", auth.MinSecretLen), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("POST /workflows/{workflowId}/trigger", tokens.RequireAuth(api.TriggerWorkflow(pool, q)))
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/workflows/"+id+"/trigger", strings.NewReader(""))
	if user != uuid.Nil {
		token, _, err := tokens.Issue(user)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func runStatus(t *testing.T, pool *pgxpool.Pool, workflowID uuid.UUID) []string {
	t.Helper()
	rows, err := pool.Query(t.Context(), `SELECT status FROM runs WHERE workflow_id = $1`, workflowID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var statuses []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		statuses = append(statuses, s)
	}
	return statuses
}

func TestTriggerWorkflowStoresPendingRunAndQueuesIt(t *testing.T) {
	pool := triggerPool(t)
	wf, user := newWorkflow(t, pool)
	q := &fakeQueue{}

	rec := serveTrigger(t, pool, q, wf.String(), user)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d (%s), want 202", rec.Code, rec.Body)
	}
	var body struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	runID, err := uuid.Parse(body.ID)
	if err != nil {
		t.Fatalf("id %q is not a uuid: %v", body.ID, err)
	}
	if len(q.runs) != 1 || q.runs[0] != runID {
		t.Fatalf("queued %v, want [%s]", q.runs, runID)
	}
	if got := runStatus(t, pool, wf); len(got) != 1 || got[0] != "PENDING" {
		t.Fatalf("runs %v, want one PENDING run", got)
	}
}

func TestTriggerWorkflowUnknownWorkflowIs404(t *testing.T) {
	pool := triggerPool(t)
	_, user := newWorkflow(t, pool)
	q := &fakeQueue{}

	rec := serveTrigger(t, pool, q, uuid.NewString(), user)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d (%s), want 404", rec.Code, rec.Body)
	}
	if len(q.runs) != 0 {
		t.Fatalf("queued %v for an unknown workflow", q.runs)
	}
}

func TestTriggerWorkflowMalformedIDIs404(t *testing.T) {
	pool := triggerPool(t)
	_, user := newWorkflow(t, pool)
	rec := serveTrigger(t, pool, &fakeQueue{}, "not-a-uuid", user)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404", rec.Code)
	}
}

func TestTriggerWorkflowKeepsRunWhenQueueFails(t *testing.T) {
	pool := triggerPool(t)
	wf, user := newWorkflow(t, pool)
	q := &fakeQueue{err: errors.New("redis down")}

	rec := serveTrigger(t, pool, q, wf.String(), user)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d, want 202 (the run is stored)", rec.Code)
	}
	if got := runStatus(t, pool, wf); len(got) != 1 || got[0] != "PENDING" {
		t.Fatalf("runs %v, want one PENDING run for the recovery sweep", got)
	}
}

func TestTriggerWorkflowRequiresAToken(t *testing.T) {
	pool := triggerPool(t)
	wf, _ := newWorkflow(t, pool)
	q := &fakeQueue{}

	rec := serveTrigger(t, pool, q, wf.String(), uuid.Nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401", rec.Code)
	}
	if len(q.runs) != 0 || len(runStatus(t, pool, wf)) != 0 {
		t.Fatal("a run was created without a token")
	}
}

func TestTriggerWorkflowOfAnotherUserIs404(t *testing.T) {
	pool := triggerPool(t)
	wf, _ := newWorkflow(t, pool)
	_, other := newWorkflow(t, pool)
	q := &fakeQueue{}

	rec := serveTrigger(t, pool, q, wf.String(), other)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404", rec.Code)
	}
	if len(q.runs) != 0 || len(runStatus(t, pool, wf)) != 0 {
		t.Fatal("a run was created for another user's workflow")
	}
}
