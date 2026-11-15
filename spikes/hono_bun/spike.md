# Spike: Hono (Bun) ping service

## Goal

Build a minimal Hono + Bun service that answers on `/ping`, to check that the environment works and that the approach is viable.

## Result

Works. Tested in this environment (Bun 1.3.13, Hono 4.13.17, linux/amd64):

- `bunx tsc --noEmit` and `bun test` pass with no errors.
- `GET /ping` returns `200 OK` with body `{"status":"ok"}` (`Content-Type: application/json`).
- `POST /ping` returns `404 Not Found`, because only `app.get("/ping", ...)` is registered.

## Approach

- [Hono](https://hono.dev/) web framework running on the [Bun](https://bun.sh/) runtime (`export default { port, fetch: app.fetch }`).
- Built-in `hono/cors` middleware (`app.use("*", cors())`).
- Route registered with `app.get("/ping", (c) => c.json({ status: "ok" }))`.
- Port read and validated from the `PORT` environment variable (`Bun.env.PORT` / `process.env.PORT`).

## Difficulties

### Encountered during the spike

- **Strict `PORT` check at import time.** `src/index.ts` throws immediately if `PORT` is not set, so a `.env` file (or `PORT=8080` in the environment) must be present even when importing the module for `bun test`.

### Encountered in the research before the spike (documentation)

1. **Multi-runtime documentation.** Hono targets many runtimes (Cloudflare Workers, Deno, Node.js, Bun). Care is needed to follow the Bun-specific entry point pattern (`export default { port, fetch: app.fetch }`) rather than the Node adapter (`@hono/node-server`).
2. **Environment variable handling.** Bun automatically loads `.env` files from the working directory and exposes variables on both `Bun.env` and `process.env`.

## Open items

- Add a test for unsupported HTTP methods (e.g. `POST /ping` returning `404`) in `src/index.test.ts`.
- Verify the PostgreSQL connection in `/ping` (or a dedicated healthcheck route) using `src/db.ts`.

## How to reproduce

```bash
cd spikes/hono_bun/backend
cp -n .env.example .env
bun install
bun test
bun run src/index.ts
curl -i localhost:8080/ping
curl -i -X POST localhost:8080/ping   # 404
```
