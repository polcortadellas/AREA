package scheduler

import (
	"math/rand/v2"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestOneLeaderAndTakeover(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set (use a throwaway PostgreSQL)")
	}
	ctx := t.Context()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	// A random key so the test never fights a real scheduler on the same database.
	key := rand.Int64()
	a, b := New(pool), New(pool)
	a.LockKey, b.LockKey = key, key
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
