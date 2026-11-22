// Package queue publishes runs to Redis Streams for the workers (ADR 0003).
package queue

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const Stream = "runs"

type Streams struct {
	client *redis.Client
}

func NewStreams(client *redis.Client) *Streams {
	return &Streams{client: client}
}

func (s *Streams) Enqueue(ctx context.Context, runID uuid.UUID) error {
	return s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: Stream,
		Values: map[string]any{"run_id": runID.String()},
	}).Err()
}
