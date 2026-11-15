# ADR 0003: Workflow engine and broker (Transactional Outbox + Redis Streams)

**Status:** Accepted
**Date:** 2026-11-15

## Context

Hooks (polling and webhooks) must detect external events and run multi-step workflows without blocking HTTP traffic and without losing queued runs when a worker stops (ENG-ASYNC). Three architectures were compared in [`docs/ARCHITECTURE.md`](../ARCHITECTURE.md).

## Decision

A modular monolith in Go, deployed as one binary in three roles: `api`, `scheduler` (single leader, PostgreSQL advisory lock) and `worker` (horizontally scalable). Runs are stored in PostgreSQL as `PENDING` (transactional outbox) and published to Redis Streams (AOF on); workers consume through Consumer Groups and take a run with an atomic SQL claim. A recovery sweep re-enqueues stale `PENDING` runs and `XAUTOCLAIM` reclaims messages from crashed workers.

## Alternatives considered

- **Modular monolith + PostgreSQL `SKIP LOCKED` queue:** rejected for this project, not as a worse design in general. Consumer groups, pending-entry reclaim and acknowledgement are native in Redis Streams, while on PostgreSQL we would adopt a queue library or write the visibility timeout and redelivery logic ourselves. The team also wants experience operating a dedicated broker. It stays a valid fallback if Redis proves too costly to operate.
- **Microservices:** rejected because tracing, inter-service authentication, several Dockerfiles and distributed sagas are too much overhead for 4 people under fixed milestones.

## Consequences

- **Positive:** workers can stop, crash or scale without losing queued runs; shared Go types between modules; native stream semantics.
- **Negative (accepted):**
  - A dual-write window between PostgreSQL and Redis if the API crashes before publishing, covered by the recovery sweep.
  - One more stateful container to run and to check in `/healthz`.
  - Delivery is at-least-once: for providers without idempotency keys (Discord webhooks, Gmail send), a crash between the external call and the `run_steps` commit can repeat one side effect on replay.
  - Module boundaries are enforced only at compile time (`internal/` packages) and in reviews.

## Evidence

**Redis Streams has not been validated in a spike yet.** The validation (stop and restart a consumer with `XAUTOCLAIM`) is part of task M-11. This decision rests on the documentation and on the comparison above, not on measurements.
