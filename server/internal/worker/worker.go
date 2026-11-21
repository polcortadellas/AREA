// Package worker executes queued runs (ADR 0003).
package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/polcortadellas/AREA/server/internal/queue"
)

const Group = "workers"

// Executor runs one claimed run; an error marks it FAILED.
type Executor interface {
	Execute(ctx context.Context, runID uuid.UUID) error
}

type ExecutorFunc func(ctx context.Context, runID uuid.UUID) error

func (f ExecutorFunc) Execute(ctx context.Context, runID uuid.UUID) error { return f(ctx, runID) }

type Worker struct {
	pool  *pgxpool.Pool
	redis *redis.Client
	exec  Executor
	name  string
}

func New(pool *pgxpool.Pool, rdb *redis.Client, exec Executor, name string) *Worker {
	return &Worker{pool: pool, redis: rdb, exec: exec, name: name}
}

// Run blocks until ctx is cancelled; runs in flight finish first.
func (w *Worker) Run(ctx context.Context, concurrency int) error {
	if concurrency < 1 {
		return fmt.Errorf("concurrency must be at least 1, got %d", concurrency)
	}
	err := w.redis.XGroupCreateMkStream(ctx, queue.Stream, Group, "0").Err()
	if err != nil && !strings.HasPrefix(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("create consumer group: %w", err)
	}

	var wg sync.WaitGroup
	for i := range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.consume(ctx, fmt.Sprintf("%s-%d", w.name, i))
		}()
	}
	wg.Wait()
	return nil
}

func (w *Worker) consume(ctx context.Context, consumer string) {
	for ctx.Err() == nil {
		streams, err := w.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    Group,
			Consumer: consumer,
			Streams:  []string{queue.Stream, ">"},
			Count:    1,
			Block:    5 * time.Second,
		}).Result()
		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			if ctx.Err() == nil {
				slog.Error("worker: read stream", "consumer", consumer, "err", err)
				time.Sleep(time.Second)
			}
			continue
		}
		for _, msg := range streams[0].Messages {
			w.handle(ctx, msg)
		}
	}
}

// Finishes the run even if ctx is cancelled, so a stop never leaves it half done.
func (w *Worker) handle(ctx context.Context, msg redis.XMessage) {
	ctx = context.WithoutCancel(ctx)

	runID, err := uuid.Parse(fmt.Sprint(msg.Values["run_id"]))
	if err != nil {
		slog.Error("worker: bad message, dropped", "message", msg.ID, "err", err)
		w.ack(ctx, msg.ID)
		return
	}

	claimed, err := w.claim(ctx, runID)
	if err != nil {
		// Not acknowledged: it stays pending in the group for XAUTOCLAIM to redeliver.
		slog.Error("worker: claim", "run", runID, "err", err)
		return
	}
	if !claimed {
		// Already claimed or finished: a duplicate delivery.
		w.ack(ctx, msg.ID)
		return
	}

	if err := w.exec.Execute(ctx, runID); err != nil {
		slog.Error("worker: run failed", "run", runID, "err", err)
		w.finish(ctx, runID, "FAILED", err.Error())
	} else {
		w.finish(ctx, runID, "SUCCEEDED", "")
	}
	w.ack(ctx, msg.ID)
}

func (w *Worker) claim(ctx context.Context, runID uuid.UUID) (bool, error) {
	tag, err := w.pool.Exec(ctx,
		`UPDATE runs SET status = 'RUNNING', started_at = now() WHERE id = $1 AND status = 'PENDING'`,
		runID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (w *Worker) finish(ctx context.Context, runID uuid.UUID, status, errText string) {
	var errPtr *string
	if errText != "" {
		errPtr = &errText
	}
	if _, err := w.pool.Exec(ctx,
		`UPDATE runs SET status = $2, error = $3, finished_at = now() WHERE id = $1 AND status = 'RUNNING'`,
		runID, status, errPtr); err != nil {
		slog.Error("worker: finish run", "run", runID, "status", status, "err", err)
	}
}

func (w *Worker) ack(ctx context.Context, id string) {
	if err := w.redis.XAck(ctx, queue.Stream, Group, id).Err(); err != nil {
		slog.Error("worker: ack", "message", id, "err", err)
	}
}
