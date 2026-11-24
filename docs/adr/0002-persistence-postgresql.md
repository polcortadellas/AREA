# ADR 0002: Persistence (PostgreSQL)

**Status:** Accepted
**Date:** 2026-11-15 (evidence added 2026-11-24)

## Why a record

Auth, workflows, runs and the outbox all rely on the database and its guarantees. Changing it later means rewriting the schema and every query.

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
- **Tested:** PostgreSQL with the Go driver (`pgx`) and `jsonb` columns, through the M-07 migrations (`server/internal/db`), including up, down and up again.

## Evidence

- Five migrations in [`server/internal/db/migrations`](../../server/internal/db/migrations), applied by the api at start-up and tested up, down and up again ([`migrate_test.go`](../../server/internal/db/migrate_test.go)).
- Foreign keys with `ON DELETE CASCADE`: deleting a user removes their workflows, nodes, edges and runs.
- `jsonb` holds `nodes.config`, `workflows.poll_cursor`, `runs.trigger_payload` and `run_steps.output`.
- `UNIQUE (workflow_id, external_event_id)` on `runs` makes the same external event start at most one run.
- Queries are written in SQL and turned into typed Go code by `sqlc` ([`server/internal/db/queries`](../../server/internal/db/queries)).
- PostgreSQL 17 in Docker Compose and in CI.
- **Not done yet:** `user_services.access_token_enc` is a `bytea` column ready for AES-GCM, but the encryption (SEC-CRYPTO) and the backups (SYS-BACKUP) are not written yet.

## Under the main alternative (MongoDB)

- A workflow would be one document with its nodes and edges, loaded without joins.
- Ownership and cascade deletes would have to be enforced in Go, since there are no foreign keys.
- Storing a run and its outbox message together needs a multi-document transaction, so even the development stack would need a replica set.
