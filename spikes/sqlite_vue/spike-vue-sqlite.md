# Spike: SQLite + Go + Vue

## Goal

Check that SQLite, Go and Vue 3 work together end to end: a Vue page talks to a small Go API that stores data in a SQLite file.

## Result

Works. Tested in this environment (Go 1.22.2, Node 20.19.5, Vue 3.5, Vite 6, linux/amd64):

- `go vet` and `go test ./...` (backend): pass (`GET /ping`, a note written and read back from SQLite, empty note rejected with `400`).
- `bun run build` (web): `vue-tsc` and `vite build` pass.
- Manual end-to-end through the Vite dev proxy: `GET /api/ping` → `{"status":"ok"}`, `POST /api/notes` stores a row, `GET /api/notes` returns it; the `.db` file is created on disk.
- Docker (`backend/Dockerfile`, multi-stage, non-root, `CGO_ENABLED=0`): image is **21 MB**; run with a volume on `/data`, `/ping` and `POST`/`GET /notes` work.
- Not tested: the page in a real browser (the Vue rendering was not looked at).
- Estimates only (not tested here): encryption at rest (plain SQLite has none; it would need SQLCipher or app-level AES), input validation beyond the empty-note check, JSON columns (`json1`), concurrent writes, and the front-end security score (XSS/CSRF/token storage).

## Approach

- Backend: Go standard library (`net/http`, `database/sql`) plus the pure-Go SQLite driver `modernc.org/sqlite`, so no CGO and no C compiler are needed. The DB is a single file (`DB_PATH`, default `spike.db`).
- Frontend: Vue 3 (`<script setup>`, TypeScript) built with Vite; `/api` is proxied to the backend on `:8080`.
- Same `/ping` snippet as the other spikes, plus `GET`/`POST /notes` to prove data persists in SQLite.

## Difficulties

- **Driver version pinned.** The latest `modernc.org/sqlite` needs a newer Go than 1.22, so it is pinned to `v1.29.10`. The team would need to fix the Go version and driver version from day 1.
- **`modernc` pulls many transitive modules** (a Go transpilation of SQLite), a heavier dependency graph than the stdlib-only Go spike. The alternative, `mattn/go-sqlite3`, needs CGO.
- **One connection only.** `db.SetMaxOpenConns(1)`: `:memory:` is per connection and SQLite allows a single writer. Fine for a spike; limits write throughput.
- SQLite is a file on one machine: no separate DB container, but scaling to several instances needs another database.
- Lockfile: Bun 1.1.34 writes the binary `bun.lockb` for `web/`, which is git-ignored here. Newer Bun writes `bun.lock` (text) and that one should be committed.

## Open items

- Look at the page in a browser, and add a front-end test.
- Compare with PostgreSQL (the `hono_bun` spike) on migrations, types and concurrency.

## How to reproduce

```bash
cd spikes/sqlite_vue/backend && go test ./... && go run .
# in another terminal
cd spikes/sqlite_vue/web && bun install && bun run dev   # http://localhost:5173
```
