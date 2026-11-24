// Package scheduler runs the polling loop on one replica at a time: the one
// holding a PostgreSQL session advisory lock (M-13).
package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polcortadellas/AREA/server/internal/node"
)

// DefaultLockKey is "AREA" in ASCII, every replica must use the same key.
const DefaultLockKey int64 = 0x41524541

// queryTimeout bounds each lock query so a hung network cannot freeze the loop.
const queryTimeout = 5 * time.Second

// pollTimeout stops one slow service from holding up every other workflow.
const pollTimeout = 30 * time.Second

type RunQueue interface {
	Enqueue(ctx context.Context, runID uuid.UUID) error
}

type Scheduler struct {
	RetryEvery time.Duration
	PollEvery  time.Duration
	LockKey    int64

	pool     *pgxpool.Pool
	registry *node.Registry
	queue    RunQueue
	// conn is a session taken out of the pool: the lock belongs to it, so it
	// must not be shared with other queries.
	conn   *pgx.Conn
	leader bool
}

func New(pool *pgxpool.Pool, registry *node.Registry, queue RunQueue) *Scheduler {
	return &Scheduler{
		RetryEvery: 500 * time.Millisecond,
		PollEvery:  30 * time.Second,
		LockKey:    DefaultLockKey,
		pool:       pool,
		registry:   registry,
		queue:      queue,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	defer s.resign()
	for {
		wait := s.RetryEvery
		if s.tryLead(ctx) {
			s.pollAll(ctx)
			wait = s.PollEvery
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(wait):
		}
	}
}

func (s *Scheduler) tryLead(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	if s.conn == nil {
		c, err := s.pool.Acquire(ctx)
		if err != nil {
			slog.Error("scheduler: connect", "err", err)
			return false
		}
		s.conn = c.Hijack()
	}

	if s.leader {
		// The lock lives as long as the session, so a live session still holds it.
		if err := s.conn.Ping(ctx); err != nil {
			slog.Warn("scheduler: lost the lock session", "err", err)
			s.resign()
			return false
		}
		return true
	}

	if err := s.conn.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, s.LockKey).Scan(&s.leader); err != nil {
		slog.Error("scheduler: try lock", "err", err)
		s.resign()
		return false
	}
	if s.leader {
		slog.Info("scheduler: became leader")
	}
	return s.leader
}

// resign closes the session, PostgreSQL releases the lock with it.
func (s *Scheduler) resign() {
	if s.conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		_ = s.conn.Close(ctx)
		cancel()
		s.conn = nil
	}
	s.leader = false
}

type watched struct {
	workflowID uuid.UUID
	trigger    string
	config     node.Values
	cursor     string
}

func (s *Scheduler) pollAll(ctx context.Context) {
	if !s.leader {
		return
	}
	rows, err := s.pool.Query(ctx, `
		SELECT w.id, n.type, n.config, COALESCE(w.poll_cursor #>> '{}', '')
		FROM workflows w JOIN nodes n ON n.workflow_id = w.id AND n.kind = 'trigger'
		WHERE w.enabled`)
	if err != nil {
		slog.Error("scheduler: list workflows", "err", err)
		return
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (watched, error) {
		var w watched
		err := row.Scan(&w.workflowID, &w.trigger, &w.config, &w.cursor)
		return w, err
	})
	if err != nil {
		slog.Error("scheduler: list workflows", "err", err)
		return
	}
	for _, w := range list {
		if ctx.Err() != nil {
			return
		}
		s.pollOne(ctx, w)
	}
}

// pollOne stores the new runs and the cursor in one transaction, so a crash
// in between can only make the next poll read the same events again, and the
// UNIQUE (workflow_id, external_event_id) on runs drops those.
func (s *Scheduler) pollOne(ctx context.Context, w watched) {
	if !s.leader {
		return
	}
	trigger, ok := s.registry.Trigger(w.trigger)
	if !ok {
		slog.Error("scheduler: unknown trigger", "workflow", w.workflowID, "type", w.trigger)
		return
	}

	pollCtx, cancel := context.WithTimeout(ctx, pollTimeout)
	defer cancel()
	// shortcut: credentials stay empty until M-14 stores the users' tokens, then load them from user_services.
	events, next, err := trigger.Poll(pollCtx, w.config, w.cursor, node.Credentials{})
	if err != nil {
		slog.Warn("scheduler: poll", "workflow", w.workflowID, "type", w.trigger, "err", err)
		return
	}

	var created []uuid.UUID
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		for _, e := range events {
			if e.ID == "" {
				slog.Error("scheduler: event without ID dropped", "workflow", w.workflowID, "type", w.trigger)
				continue
			}
			var runID uuid.UUID
			err := tx.QueryRow(ctx, `
				INSERT INTO runs (workflow_id, external_event_id, trigger_payload)
				VALUES ($1, $2, COALESCE($3::jsonb, '{}'))
				ON CONFLICT (workflow_id, external_event_id) DO NOTHING
				RETURNING id`, w.workflowID, e.ID, e.Outputs).Scan(&runID)
			if errors.Is(err, pgx.ErrNoRows) {
				continue // already seen
			}
			if err != nil {
				return err
			}
			created = append(created, runID)
		}
		_, err := tx.Exec(ctx, `UPDATE workflows SET poll_cursor = to_jsonb($2::text) WHERE id = $1`, w.workflowID, next)
		return err
	})
	if err != nil {
		slog.Error("scheduler: store runs", "workflow", w.workflowID, "err", err)
		return
	}

	// Published after the commit: a failed publish leaves the run PENDING for the recovery sweep.
	for _, runID := range created {
		if err := s.queue.Enqueue(ctx, runID); err != nil {
			slog.Warn("scheduler: enqueue failed, the recovery sweep will retry", "run", runID, "err", err)
		}
	}
}
