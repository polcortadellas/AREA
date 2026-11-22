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

Validated by the automated tests of `server/internal/worker` (M-11), run against Redis 7.4 and PostgreSQL 16:

- Runs queued while no worker is up are executed exactly once when workers start, even when the same run is queued twice (the atomic claim drops the duplicate).
- Runs stored as `PENDING` but never published (the api crashed between the commit and the `XADD`) are published again by the recovery sweep after `StaleAfter` (30 s).
- Messages read by a consumer that never acknowledged them are taken over with `XAUTOCLAIM` and executed once.
- Capacity is set with `WORKER_CONCURRENCY` (default 4).

Timeouts today: an e-mail send is bounded by 10 s, a queue read blocks for 1 s, and recovery acts after 30 s. Calls to third-party services must set their own timeout when M-12 adds them.

Not covered:

- A run whose worker dies after the claim stays `RUNNING`: nothing recovers it yet (retries and the dead-letter store belong to ENG-RETRY).
- The live demonstration with `docker compose stop worker` was not run; the tests stand in for it.
- The sweep re-enqueues every stale `PENDING` run on each pass, so a long outage leaves duplicate messages that the claim drops.
