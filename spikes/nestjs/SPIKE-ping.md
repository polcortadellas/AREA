# Spike: NestJS ping service

## Goal

Build a minimal NestJS service that answers on `/ping`, to check that the environment works and that the approach is viable.

## Result

Works. Tested in this environment (Node 22.14.0, NestJS 11, macOS arm64):

- `npm run build` and `npm run lint` pass with no errors.
- `GET /ping` returns `200 OK` with body `{"status":"ok"}` (`Content-Type: application/json`).
- `POST /ping` returns `404 Not Found` (Nest has no route for that method; Go's mux answers `405`).
- Unit test (`npm test`) and e2e test (`npm run test:e2e`, real HTTP request through supertest) pass.
- `npm audit --omit=dev`: 0 vulnerabilities.

## Approach

- Generated with the Nest CLI (`nest new --strict`), TypeScript in strict mode, Express under the hood.
- Standard Nest structure: `AppModule` → `AppController` (`@Get('ping')`) → `AppService`.
- Port read from the `PORT` environment variable, default `8080`.

## Difficulties

### Encountered during the spike

1. **Nest CLI 12 crashes on Node 22.14.** `npx @nestjs/cli@latest new` failed with `ERR_REQUIRE_CYCLE_MODULE` and generated nothing. Pinning `@nestjs/cli@11.0.24` worked, so the project is on Nest 11. Nest 12 came out recently and part of the ecosystem still targets 11 (e.g. latest `@nestjs/swagger` refuses to install on Nest 11), so the team would need to pin Node + Nest versions from day 1.

### Compared to the Go spike

- **More boilerplate.** The CLI generates ~15 files (package.json, tsconfigs, eslint, prettier, jest configs, module/controller/service) for one route, vs 3 files in Go. In exchange, tests, linting and the module structure are there from the start.
- **No method-mismatch status.** Wrong method gives `404`, not `405`.

## Open items

- Swagger, OAuth2, security middleware, PostgreSQL, scheduled hooks, Docker and a React client were also explored. They informed the scores in `planning/stack_decission.md` but are kept out of this spike so all stacks are compared on the same `/ping` snippet.

## How to reproduce

```bash
cd spikes/nestjs/server
npm install
npm run start          # or: PORT=9090 npm run start
curl -i localhost:8080/ping
curl -i -X POST localhost:8080/ping   # 404
npm test && npm run test:e2e
```
