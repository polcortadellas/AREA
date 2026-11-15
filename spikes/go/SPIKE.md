# Spike: Go ping service

The full report is in [`SPIKE-ping.md`](SPIKE-ping.md). This file follows the bootstrap's required name and points to it.

## Where I was when time ran out

The `/ping` service works: `GET /ping` answers `{"status":"ok"}`, `POST /ping` answers `405`, and the port comes from `PORT`. About 20 minutes; no timebox was recorded.

## What consumed the most minutes

Finding usable examples: they live in the reference documentation, not in the tutorials (see "Difficulties" in the report).

## What I would need to go further

Tests with `go test`, a Dockerfile, an OpenAPI description, OAuth2 / OIDC, and a PostgreSQL driver. A Go API with tests, SQLite and a 21 MB Docker image exists in [`../sqlite_vue/`](../sqlite_vue/spike-vue-sqlite.md).
