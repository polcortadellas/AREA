// Package scheduler runs the polling loop on one replica at a time: the one
// holding a PostgreSQL session advisory lock (M-13).
package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultLockKey is "AREA" in ASCII, every replica must use the same key.
const DefaultLockKey int64 = 0x41524541

// queryTimeout bounds each lock query so a hung network cannot freeze the loop.
const queryTimeout = 5 * time.Second

type Scheduler struct {
	RetryEvery time.Duration
	PollEvery  time.Duration
	LockKey    int64

	pool *pgxpool.Pool
	// conn is a session taken out of the pool: the lock belongs to it, so it
	// must not be shared with other queries.
	conn   *pgx.Conn
	leader bool
}

func New(pool *pgxpool.Pool) *Scheduler {
	return &Scheduler{
		RetryEvery: 500 * time.Millisecond,
		PollEvery:  30 * time.Second,
		LockKey:    DefaultLockKey,
		pool:       pool,
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

func (s *Scheduler) pollAll(_ context.Context) {
	if !s.leader {
		return
	}
	// shortcut: no triggers are polled yet, the loop with event-ID deduplication is the next part of M-13.
}
