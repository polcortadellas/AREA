package worker_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/polcortadellas/AREA/server/internal/db"
	"github.com/polcortadellas/AREA/server/internal/queue"
	"github.com/polcortadellas/AREA/server/internal/worker"
)

func environment(t *testing.T) (*pgxpool.Pool, *redis.Client) {
	t.Helper()
	base, redisURL := os.Getenv("DATABASE_URL"), os.Getenv("REDIS_URL")
	if base == "" || redisURL == "" {
		t.Skip("DATABASE_URL and REDIS_URL must be set (throwaway PostgreSQL and Redis)")
	}
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)

	b := make([]byte, 6)
	_, _ = rand.Read(b)
	name := "area_worker_test_" + hex.EncodeToString(b)
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

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(opt)
	t.Cleanup(func() { _ = rdb.Close() })
	return pool, rdb
}

func pendingRuns(t *testing.T, pool *pgxpool.Pool, n int) []uuid.UUID {
	t.Helper()
	ctx := t.Context()
	var userID, workflowID uuid.UUID
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
	ids := make([]uuid.UUID, n)
	for i := range ids {
		if err := pool.QueryRow(ctx,
			`INSERT INTO runs (workflow_id, external_event_id) VALUES ($1, $2) RETURNING id`,
			workflowID, uuid.NewString(),
		).Scan(&ids[i]); err != nil {
			t.Fatal(err)
		}
	}
	return ids
}

type counter struct {
	mu   sync.Mutex
	runs map[uuid.UUID]int
}

func (c *counter) Execute(_ context.Context, id uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.runs[id]++
	return nil
}

func (c *counter) count(id uuid.UUID) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.runs[id]
}

func statusOf(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) string {
	t.Helper()
	var s string
	if err := pool.QueryRow(t.Context(), `SELECT status FROM runs WHERE id = $1`, id).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

// Runs queued while no worker is up, and queued twice, run exactly once: the
// second delivery finds the run no longer PENDING and is acknowledged away.
func TestWorkersRunEachQueuedRunExactlyOnce(t *testing.T) {
	pool, rdb := environment(t)
	ids := pendingRuns(t, pool, 5)

	q := queue.NewStreams(rdb)
	for _, id := range ids {
		if err := q.Enqueue(t.Context(), id); err != nil {
			t.Fatal(err)
		}
		if err := q.Enqueue(t.Context(), id); err != nil {
			t.Fatal(err)
		}
	}

	exec := &counter{runs: map[uuid.UUID]int{}}
	w := worker.New(pool, rdb, exec, "test")
	ctx, stop := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx, 3) }()

	deadline := time.Now().Add(15 * time.Second)
	for _, id := range ids {
		for statusOf(t, pool, id) != "SUCCEEDED" {
			if time.Now().After(deadline) {
				t.Fatalf("run %s never finished (status %s)", id, statusOf(t, pool, id))
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
	time.Sleep(500 * time.Millisecond) // give the duplicates time to be read and dropped
	stop()
	if err := <-done; err != nil {
		t.Fatalf("run: %v", err)
	}

	for _, id := range ids {
		if n := exec.count(id); n != 1 {
			t.Fatalf("run %s executed %d times, want 1", id, n)
		}
	}
}

func TestRunStopsWhenContextIsCancelled(t *testing.T) {
	pool, rdb := environment(t)
	ctx, stop := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- worker.New(pool, rdb, &counter{runs: map[uuid.UUID]int{}}, "test").Run(ctx, 2) }()

	time.Sleep(200 * time.Millisecond)
	stop()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("worker did not stop after cancellation")
	}
}

func TestRunRejectsZeroConcurrency(t *testing.T) {
	pool, rdb := environment(t)
	if err := worker.New(pool, rdb, &counter{runs: map[uuid.UUID]int{}}, "test").Run(t.Context(), 0); err == nil {
		t.Fatal("expected an error for concurrency 0")
	}
}
