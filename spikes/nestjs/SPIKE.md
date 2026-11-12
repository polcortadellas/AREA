# Spike: Stack A (React + NestJS + PostgreSQL)

- **Member:** Achille Bertani
- **Issue:** #2
- **Date:** 2026-11-11
- **Versions:** Node 22.14, NestJS 11, TypeORM 1.x, PostgreSQL 17, React 19 + Vite 6
- **Total time:** ~15 min, AI-assisted

## How to run

```bash
cd spikes/nestjs
docker compose up -d --build        # Postgres + API on :8080
curl localhost:8080/ping            # {"status":"ok"}
open http://localhost:8080/docs     # Swagger UI (raw spec: /docs-json)
cd web && npm install && npm run dev  # React page on :8081
```

## 1. Spike log

All 9 steps working.

| # | Step | Worked? | Notes / what slowed me down |
|---|------|:-------:|-----------------------------|
| 0 | Project setup (`nest new`) | ✅ | `@nestjs/cli@latest` (v12) crashed on Node 22.14 (`ERR_REQUIRE_CYCLE_MODULE`); had to pin CLI 11.0.24 → project is on Nest 11. |
| 1 | `GET /ping` → `{"status":"ok"}` on port 8080, read from `PORT` | ✅ | Template already reads `PORT`; changed default to 8080 + updated unit/e2e tests. |
| 2 | OpenAPI: `@nestjs/swagger`, spec served at `/docs` | ✅ | Latest swagger (v12) needs Nest 12 → `ERESOLVE`; pinned `@nestjs/swagger@^11`. After that: ~10 lines, UI at `/docs`, OpenAPI 3.0 JSON at `/docs-json`, generated from DTO classes. |
| 3 | OAuth2 / OIDC: library setup (passport, openid-client) | ✅ | `@nestjs/passport` + `passport-google-oauth20`. `/auth/google` → 302 to Google consent with `openid email profile` scopes. Full round-trip not tested (needs real client ID/secret). |
| 4 | Security: `npm audit`, helmet, ValidationPipe | ✅ | Prod deps: **0 vulns**. Dev deps: 28 low/moderate (jest, eslint, CLI). helmet sets CSP/HSTS/nosniff; `ValidationPipe({whitelist, forbidNonWhitelisted})` rejects missing fields *and* injected fields (`isAdmin` → 400). CORS restricted to the web origin. |
| 5 | PostgreSQL connection (Prisma or TypeORM) | ✅ | Chose TypeORM (first-party `@nestjs/typeorm`). `areas` table auto-created from the entity, data survives restart. `synchronize: true` is spike-only; real project needs migrations. |
| 6 | Hooks: `@nestjs/schedule` cron job | ✅ | `@Interval` poller loads enabled AREAs from DB, runs them in parallel (`Promise.allSettled`), writes `lastTriggeredAt`. Had to add a manual guard to avoid overlapping ticks. |
| 7 | Docker: multi-stage Dockerfile + compose with Postgres | ✅ | Multi-stage, non-root user, healthcheck, `depends_on: service_healthy`. Bug: compose passes unset vars as `""`, `??` doesn't catch it → passport crashed at boot; fixed with `\|\|`. Image is **401 MB** (node_modules). |
| 8 | (optional) React page calling `/ping` | ✅ | Vite React-TS on :8081: shows `/ping` status, lists AREAs (live refresh shows hooks firing), creates AREAs and displays the server's validation errors. |

**Where I stopped:** everything in the plan works end to end (web → API → Postgres → hook engine), in Docker. Not done: a real Google login round-trip, user accounts/JWT, the Android client.

**What caused the most friction:** npm version churn rather than NestJS itself. Nest 12 was released recently and the ecosystem is split: CLI 12 crashes on our Node, swagger 12 needs Nest 12, schedule supports both. Then the empty-env-var bug in Docker.

**What I would need to go further:**
- Pin the Node and Nest major versions for the whole team (`.nvmrc` + exact deps) on day 1.
- `@nestjs/config` with schema validation so a missing env var fails clearly instead of crashing deep inside passport.
- TypeORM migrations instead of `synchronize`.
- A session or state store for OAuth `state` / PKCE, plus encrypted storage of per-service tokens (each AREA needs its own tokens).
- For multiple server instances: a queue (BullMQ + Redis) instead of the in-process scheduler.
- The e2e test now needs a DB (AppModule loads TypeORM): use a test DB in compose or testcontainers.

## 2. Strengths and weaknesses per criterion

These feed the individual scoring in `planning/stack_decission.md`. Score from 1 (very poor) to 5 (excellent).

| Criterion (weight) | Strength | Weakness | Score (1-5) |
|--------------------|----------|----------|:-----------:|
| Security (3) | helmet, a strict whitelist ValidationPipe and CORS each take one line; 0 prod vulns; non-root container. | Not secure by default (you must remember to add each piece). Large dependency tree (~740 packages) = bigger supply-chain surface; dev deps have audit findings. OAuth `state` had to be disabled (needs a session store). | 3 |
| OAuth2 / OIDC support (3) | Passport has strategies for nearly every service we'd plug in (Google, GitHub, Discord, Spotify…); the Nest wrapper makes each one a small class + a guard. | Passport is old callback-style code wrapped by Nest: `req.user` is untyped, `state`/PKCE needs a session store. | 4 |
| OpenAPI generation (3) | Code-first: the same DTO class drives validation, TS types and the OpenAPI schema, so docs can't drift. UI + JSON spec in ~10 lines. | Needs `@ApiProperty`/`@ApiOkResponse` decorators everywhere (or the CLI plugin); version pinning issue at install. | 5 |
| Concurrency, hooks (2) | Event loop fits I/O-bound polling of many APIs well; `@nestjs/schedule` + `Promise.allSettled` cover polling out of the box. | Single-threaded (CPU work blocks everything); in-process scheduler doesn't scale to several instances without Redis/BullMQ; manual overlap guard. | 3 |
| Team learning curve (2) | TypeScript on both web and server, one language; Nest's structure (modules/controllers/services) keeps 4 people's code consistent. | DI + decorators are "magic" for newcomers; you need to learn Nest's way on top of TS. | 4 |
| Ecosystem and libraries (2) | First-party modules for everything we needed (swagger, passport, typeorm, schedule, config); huge npm ecosystem for service SDKs. | Version churn: Nest 11 vs 12 split broke the CLI and swagger install. | 5 |
| Docker deployment (2) | Standard multi-stage Node pattern, well documented; healthchecks and non-root user are simple to add. | 401 MB image (vs ~20 MB for a Go binary); needs a build step; env-var gotcha. | 4 |
| Typing and maintainability (1) | Strict TS, typed DTOs/entities; types can be shared with the React client. | `process.env` and `req.user` untyped; `!` on every entity/DTO field. | 4 |

**Weighted total: 72 / 90**
(3×3 + 4×3 + 5×3 + 3×2 + 4×2 + 5×2 + 4×2 + 4×1)
