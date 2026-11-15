# Spike distance comparison

## What "distance travelled" means here

The number of checkpoints reached, out of six, inside one spike. Time alone cannot be compared: the timeboxes were not recorded the same way and some spikes were AI-assisted (NestJS, FastAPI). Checkpoints can be checked in the repository.

1. `GET /ping` returns exactly `{"status":"ok"}`.
2. The port comes from the `PORT` environment variable.
3. An automated test exists.
4. A Docker image builds and runs.
5. An OpenAPI description is served.
6. At least one AREA-specific piece works (OAuth2 library, database, or hook scheduler).

## Back-end spikes

| Spike | Time | 1 ping | 2 PORT | 3 test | 4 Docker | 5 OpenAPI | 6 AREA piece | Distance |
|---|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| [Go](go/SPIKE-ping.md) | ~20 min | yes | yes | no (open item) | no, in this spike | no | no | **2 / 6** |
| [NestJS](nestjs/SPIKE.md) | ~15 min, AI-assisted | yes | yes | yes | yes, 401 MB | yes, from DTOs | yes (Passport, hooks, PostgreSQL) | **6 / 6** |
| [Hono / Bun](hono_bun/SPIKE.md) | ~15 min | yes | yes | yes | files written, not measured | no | partial (`db.ts`, not verified) | **3 / 6** |
| [FastAPI](fastapi/SPIKE.md) | ~3 min, AI-assisted | yes | yes | yes | yes, 214 MB | yes, zero code | no | **5 / 6** |

## Why Go scores lower

The Go ping spike has no Dockerfile and no test (the test is listed as an open item), while NestJS, Hono/Bun and FastAPI have both. That is a difference in what was done inside the spike, not a measured difference in the language. The Go backend in the SQLite + Vue spike has tests and a 21 MB image, but it is a different spike and is not counted in the table.

## How to read it

- Go went the shortest distance **in its own ping spike**, but a Go backend was also built in the [SQLite + Vue spike](sqlite_vue/spike-vue-sqlite.md) with tests and a 21 MB image. The table undercounts Go.
- NestJS went the farthest, but it is also the one that got the most extra scope (authentication, hooks, database). Distance measures progress, not quality: it does not say which stack suits AREA. That is the job of the decision matrices.
- Timeboxes were not recorded for Go, NestJS and Hono/Bun, so no spike can be called "out of time". Only FastAPI recorded one (60 min, finished in 3).
