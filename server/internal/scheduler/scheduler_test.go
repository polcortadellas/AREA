package scheduler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	mathrand "math/rand/v2"
	"net/url"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polcortadellas/AREA/server/internal/connectors"
	"github.com/polcortadellas/AREA/server/internal/db"
	"github.com/polcortadellas/AREA/server/internal/node"
)

// testPool returns a pool on a fresh, migrated database that is dropped after the test.
func testPool(t *testing.T) *pgxpool.Pool {
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
	name := "area_scheduler_test_" + hex.EncodeToString(b)
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

type fakeQueue struct{ runs []uuid.UUID }

func (q *fakeQueue) Enqueue(_ context.Context, runID uuid.UUID) error {
	q.runs = append(q.runs, runID)
	return nil
}

func newScheduler(pool *pgxpool.Pool) (*Scheduler, *fakeQueue) {
	registry := node.NewRegistry()
	connectors.Register(registry)
	q := &fakeQueue{}
	s := New(pool, registry, q)
	// A random key so the test never fights a real scheduler on the same database.
	s.LockKey = mathrand.Int64()
	return s, q
}

func TestOneLeaderAndTakeover(t *testing.T) {
	ctx := t.Context()
	pool := testPool(t)
	a, _ := newScheduler(pool)
	b, _ := newScheduler(pool)
	b.LockKey = a.LockKey
	t.Cleanup(a.resign)
	t.Cleanup(b.resign)

	if !a.tryLead(ctx) {
		t.Fatal("a should take the free lock")
	}
	if b.tryLead(ctx) {
		t.Fatal("b must not lead while a holds the lock")
	}
	if !a.tryLead(ctx) {
		t.Fatal("a should still lead")
	}

	a.resign() // same as the leader dying: its session closes
	if !b.tryLead(ctx) {
		t.Fatal("b should take over once a is gone")
	}
	if a.tryLead(ctx) {
		t.Fatal("a must not lead again while b holds the lock")
	}
}

// M-13 "done when": reading the same external event twice enqueues 1 run.
func TestSameEventTwiceMakesOneRun(t *testing.T) {
	ctx := t.Context()
	pool := testPool(t)
	s, q := newScheduler(pool)
	t.Cleanup(s.resign)

	var userID, workflowID uuid.UUID
	if err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ('poll@example.com', 'x') RETURNING id`,
	).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO workflows (user_id, name) VALUES ($1, 'poll') RETURNING id`, userID,
	).Scan(&workflowID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO nodes (workflow_id, kind, type) VALUES ($1, 'trigger', 'dummy.counter')`, workflowID,
	); err != nil {
		t.Fatal(err)
	}

	runs := func() int {
		var n int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM runs WHERE workflow_id = $1`, workflowID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	if !s.tryLead(ctx) {
		t.Fatal("scheduler should take the free lock")
	}
	s.pollAll(ctx) // reads event "1"
	// A cursor that was never saved, e.g. a crash: the next poll reads event "1" again.
	if _, err := pool.Exec(ctx, `UPDATE workflows SET poll_cursor = NULL WHERE id = $1`, workflowID); err != nil {
		t.Fatal(err)
	}
	s.pollAll(ctx)
	if n := runs(); n != 1 || len(q.runs) != 1 {
		t.Fatalf("same event read twice: %d runs and %d enqueued, want 1 and 1", n, len(q.runs))
	}

	s.pollAll(ctx) // the saved cursor moves on to event "2"
	if n := runs(); n != 2 || len(q.runs) != 2 {
		t.Fatalf("new event: %d runs and %d enqueued, want 2 and 2", n, len(q.runs))
	}

	s.resign()
	s.pollAll(ctx) // without the lock nothing is polled
	if n := runs(); n != 2 {
		t.Fatalf("poll without the lock: %d runs, want 2", n)
	}
}
