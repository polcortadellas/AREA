package queue_test

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/polcortadellas/AREA/server/internal/queue"
)

func TestEnqueueAddsRunToStream(t *testing.T) {
	addr := os.Getenv("REDIS_URL")
	if addr == "" {
		t.Skip("REDIS_URL not set (use a throwaway Redis)")
	}
	opt, err := redis.ParseURL(addr)
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(opt)
	defer client.Close()
	ctx := t.Context()

	// Other packages' tests share the stream, so look for this run's entry only.
	runID := uuid.New()
	if err := queue.NewStreams(client).Enqueue(ctx, runID); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	entries, err := client.XRange(ctx, queue.Stream, "-", "+").Result()
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Values["run_id"] == runID.String() {
			return
		}
	}
	t.Fatalf("stream holds no entry for run %s", runID)
}
