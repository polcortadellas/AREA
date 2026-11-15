# ADR 0001: Backend stack (Go)

**Status:** Accepted
**Date:** 2026-11-15

## Context

The server runs hooks that poll many external services in parallel, exposes an OpenAPI-described HTTP API to the web and mobile clients, and ships in Docker. Four options were compared in the spikes and in [`docs/tech-selection/backend-selection.md`](../tech-selection/backend-selection.md): Go, NestJS, Hono/Bun and FastAPI. The matrix totals are within two points (71, 72, 73, 72), so they do not decide.

## Decision

Go with the standard library (`net/http`, `database/sql`).

## Alternatives considered

- **NestJS:** best ecosystem and code-first OpenAPI (scored 5), but a 401 MB image, an in-process scheduler that needed a manual guard, and a large dependency tree.
- **Hono / Bun:** easiest to learn (5), but the smallest ecosystem for our needs.
- **FastAPI:** OpenAPI for free (5), but no scheduler and a 214 MB image. Kept as the fallback.

## Consequences

- **Positive:** goroutines fit parallel hooks; 21 MB static image; few dependencies; strong typing.
- **Negative:** nobody in the team knew Go (learning curve scored 2); OpenAPI generation from Go code and OAuth2 / OIDC were not tested in the spikes; more code to write than with a framework.
- **Mitigation:** test OpenAPI and OAuth2 first. If either fails, switch to FastAPI.
