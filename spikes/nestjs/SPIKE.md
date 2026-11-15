# Spike: Stack A (React + NestJS + PostgreSQL)

- **Member:** Achille Bertani
- **Issues:** #2 (back-end, `/ping` part in [`SPIKE-ping.md`](SPIKE-ping.md)), #11 (front-end + database)
- **Date:** 2026-11-11, front-end and database re-added 2026-11-14
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

## 2. Back-end (NestJS): strengths and weaknesses

Feeds [`docs/tech-selection/backend-selection.md`](../../docs/tech-selection/backend-selection.md). Score from 1 (very poor) to 5 (excellent).

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

## 3. Front-end (React + Vite): strengths and weaknesses

Feeds [`docs/tech-selection/frontend-selection.md`](../../docs/tech-selection/frontend-selection.md). "Tested" = seen in this spike; "estimate" = from documentation only.

| Criterion (weight) | Strength | Weakness | Tested? | Score (1-5) |
|--------------------|----------|----------|:-------:|:-----------:|
| Security: XSS, CSRF, token storage (3) | JSX escapes everything by default; raw HTML needs the explicit `dangerouslySetInnerHTML`. API calls limited by the server's CORS origin. | Nothing built in for CSRF or token storage: cookie/token handling is up to us. | Partly (CORS, escaping) | 4 |
| Dynamic forms from `/about.json` (3) | Forms are plain JS: mapping a JSON schema to inputs is a `.map()`; the page already renders a list from the API and shows server validation errors. | No two-way binding: every dynamic input needs controlled-state boilerplate (or a lib like react-jsonschema-form). | Partly (list + errors, not a schema form) | 4 |
| Accessibility (WCAG AA) & i18n (3) | Semantic HTML in JSX; `eslint-plugin-jsx-a11y` and `react-i18next` are mature. | Nothing built in: i18n and a11y are third-party libs + discipline. | Estimate | 3 |
| Team learning curve (2) | Most widely known front-end library; prior Next.js experience applied directly; the page took minutes. | `useEffect` / re-render pitfalls; many ways to do the same thing (state, routing, forms). | Tested | 5 |
| OpenAPI / TypeScript integration (2) | Vite template is strict TS; types can be generated from the server's `/docs-json` (`openapi-typescript`). Same language as the NestJS server. | Client generation not wired up in the spike. | Partly (TS only) | 4 |
| Docker build & bundle footprint (1) | `vite build` in ~0.4 s; output is static files. | 70 KB gzipped JS for a one-page app (heavier baseline than Vue/Svelte); needs its own static server (e.g. nginx). | Tested | 3 |

**Weighted total: 54 / 70**
(4×3 + 4×3 + 3×3 + 5×2 + 4×2 + 3×1)

## 4. Persistence (PostgreSQL + TypeORM): strengths and weaknesses

Feeds [`docs/tech-selection/persistence-selection.md`](../../docs/tech-selection/persistence-selection.md).

| Criterion (weight) | Strength | Weakness | Tested? | Score (1-5) |
|--------------------|----------|----------|:-------:|:-----------:|
| Security & encryption at rest (3) | `bytea` for AES-GCM ciphertexts, roles, TLS; long CVE track record. | No transparent encryption at rest: SEC-CRYPTO must be done in the app (pgcrypto would send keys through SQL). | Estimate | 4 |
| Hybrid data model: relational + JSON (3) | Foreign keys for users/areas plus `jsonb` with GIN indexes for per-service configs. | — | Partly (relational table + `timestamptz`; `jsonb` not used yet) | 5 |
| ACID transactions & concurrency (3) | Full ACID; `SELECT … FOR UPDATE SKIP LOCKED` for job queues; unique constraints for idempotency keys. | — | Partly (hook engine updates rows concurrently) | 5 |
| Backup & restore (2) | `pg_dump` / `pg_restore` are single commands with a consistent snapshot. | Restore script still to write. | Estimate | 5 |
| Driver & ecosystem (2) | `pg` driver + `@nestjs/typeorm` worked first time; table created from the entity class. | TypeORM's `synchronize` is unsafe for production and its migrations are clunky (Prisma is the alternative). | Tested | 4 |
| Docker footprint (1) | Official `postgres:17-alpine` image, healthy in seconds, `pg_isready` healthcheck; data persisted across restarts and days. | Needs a named volume and a healthcheck so the API waits for it. | Tested | 4 |

**Weighted total: 64 / 70**
(4×3 + 5×3 + 5×3 + 5×2 + 4×2 + 4×1)
