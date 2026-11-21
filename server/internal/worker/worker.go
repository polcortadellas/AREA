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
	"github.com/jackc/pgx/v5"
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
	// SweepEvery is how often stale PENDING runs are re-enqueued and idle messages reclaimed.
	SweepEvery time.Duration
	// StaleAfter is how long a PENDING run or an unacknowledged message waits before recovery.
	StaleAfter time.Duration

	pool  *pgxpool.Pool
	redis *redis.Client
	queue *queue.Streams
	exec  Executor
	name  string
}

func New(pool *pgxpool.Pool, rdb *redis.Client, exec Executor, name string) *Worker {
	return &Worker{
		SweepEvery: 30 * time.Second,
		StaleAfter: 30 * time.Second,
		pool:       pool,
		redis:      rdb,
		queue:      queue.NewStreams(rdb),
		exec:       exec,
		name:       name,
	}
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
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.recover(ctx)
	}()
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
			Block:    time.Second,
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

func (w *Worker) recover(ctx context.Context) {
	ticker := time.NewTicker(w.SweepEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.sweep(ctx)
			w.reclaim(ctx)
		}
	}
}

// sweep publishes again the runs stored as PENDING that never reached a worker,
// e.g. because the api crashed between the commit and the XADD.
// shortcut: it re-enqueues every stale PENDING run on each pass, so a long outage
// leaves duplicate messages that the claim drops; capped at 100 per pass.
func (w *Worker) sweep(ctx context.Context) {
	rows, err := w.pool.Query(ctx,
		`SELECT id FROM runs WHERE status = 'PENDING' AND created_at < now() - make_interval(secs => $1) LIMIT 100`,
		w.StaleAfter.Seconds())
	if err != nil {
		slog.Error("worker: sweep", "err", err)
		return
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		slog.Error("worker: sweep", "err", err)
		return
	}
	for _, id := range ids {
		if err := w.queue.Enqueue(ctx, id); err != nil {
			slog.Error("worker: sweep enqueue", "run", id, "err", err)
		}
	}
}

// reclaim takes over messages whose consumer stopped before acknowledging them.
func (w *Worker) reclaim(ctx context.Context) {
	start := "0-0"
	for ctx.Err() == nil {
		msgs, next, err := w.redis.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   queue.Stream,
			Group:    Group,
			Consumer: w.name + "-reclaim",
			MinIdle:  w.StaleAfter,
			Start:    start,
			Count:    100,
		}).Result()
		if err != nil {
			slog.Error("worker: reclaim", "err", err)
			return
		}
		for _, msg := range msgs {
			w.handle(ctx, msg)
		}
		if next == "0-0" || len(msgs) == 0 {
			return
		}
		start = next
	}
}
