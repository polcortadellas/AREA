# ADR 0002: Persistence (PostgreSQL)

**Status:** Accepted
**Date:** 2026-11-15

## Context

AREA stores users, linked identities, per-service tokens and workflow definitions, and needs deduplication and idempotency for hooks. Three options were compared in [`docs/tech-selection/persistence-selection.md`](../tech-selection/persistence-selection.md): PostgreSQL (64 / 70), SQLite (43) and MongoDB (42).

## Decision

PostgreSQL, with `jsonb` for per-service configuration.

## Alternatives considered

- **SQLite:** 21 MB image and no extra container, but a single writer (the spike limited the pool to one connection) and one file on one machine, which blocks several server instances.
- **MongoDB:** workflows fit one document and a unique index deduplicated events, but it accepted a workflow owned by a non-existent user (no foreign keys) and refused multi-document transactions without a replica set.

## Consequences

- **Positive:** foreign keys, full ACID, unique constraints for idempotency, `SKIP LOCKED` for job queues, `pg_dump` / `pg_restore` for backups.
- **Negative:** no encryption at rest by default, so SEC-CRYPTO is done in the application; a migration tool is needed (TypeORM's `synchronize` is for spikes only); an extra container with a volume and a healthcheck.
- **Job queue:** `SKIP LOCKED` is not used as the job queue; the broker decision is [ADR 0003](0003-workflow-engine-broker.md), where a PostgreSQL queue is kept as the fallback.
- **Not tested yet:** PostgreSQL with the Go driver, and `jsonb` columns.
