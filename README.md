# AREA

AREA: an IFTTT/Zapier-style automation platform (Go server, React web client, React Native Android client).

## Milestone 1 deliverables

- Technology selection: [back-end](docs/tech-selection/backend-selection.md), [front-end](docs/tech-selection/frontend-selection.md), [persistence](docs/tech-selection/persistence-selection.md)
- Spikes: [`spikes/`](spikes/) and the comparison in [`spikes/DISTANCE.md`](spikes/DISTANCE.md)
- [Recommendation](decision/RECOMMENDATION.md)
- [Features and additional requirements](docs/FEATURES.md)
- [Technical questions](docs/TECHNICAL_QUESTIONS.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Project plan and WBS](planning/project-wbs.md) (management method in section 2) and the [board](planning/board-project.md)
- Architecture decision records: [`docs/adr/`](docs/adr/)
- Exported versions (Word): [architecture](planning/AREA_System_Architecture_and_Options_Evaluation.docx), [features](planning/AREA_Feature_Prioritisation_and_Catalogue_Selection.docx), [plan and WBS](planning/AREA_Milestone_1_Project_Plan_and_WBS.docx)

## Repository layout

| Folder | What | Lint / type-check |
|---|---|---|
| [`server/`](server/) | Go application server, one binary with three roles (`ROLE=api\|scheduler\|worker`), entry point `cmd/area` | golangci-lint (config in `server/.golangci.yml`), gofmt + goimports |
| [`web/`](web/) | React + Vite + TypeScript web client | ESLint, `tsc`, Prettier |
| [`mobile/`](mobile/) | React Native + TypeScript Android client | ESLint, `tsc`, Prettier |

Prettier settings are shared from the root [`.prettierrc.json`](.prettierrc.json).

## Development

Requirements: Go 1.26+, [golangci-lint](https://golangci-lint.run/welcome/install/) v2, [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) (only to run `make generate`), Node.js 22.11+.

```bash
make install   # download Go modules and npm packages
make lint      # every linter and type-check; must exit 0 before opening a PR
make format    # auto-fix formatting in all three modules
```

Each module can also be checked on its own: `make lint-server`, `make lint-web`, `make lint-mobile`, and `make lint-api` for the OpenAPI contract ([`server/api/openapi.yaml`](server/api/openapi.yaml)).

Every pull request runs [`ci.yml`](.github/workflows/ci.yml): the same linters, type-checks and tests (the server tests use a PostgreSQL service), plus the server and web Docker builds. Changes under `mobile/` also build the APK ([`android.yml`](.github/workflows/android.yml)).

## Installation Process

Requirements: Docker with Docker Compose v2.

```bash
make env                  # creates .env from .env.example with random secrets
docker compose up --build
```

The first build takes a while: the Android APK is compiled inside Docker (on Apple Silicon it runs under x86_64 emulation).

## Usage

| What | Where |
|---|---|
| API | http://localhost:8080 (`GET /ping` answers `pong`) |
| API documentation (Swagger UI) | http://localhost:8080/docs (raw contract: `/openapi.yaml`) |
| Web client | http://localhost:8081 |
| Mailpit (inbox for confirmation e-mails) | http://localhost:8025 |
| Android APK | `./dist/client.apk` (written by the `android-builder` container, which then exits) |

PostgreSQL and Redis are only reachable from inside the Compose network. Host ports can be changed in `.env`.

```bash
make up        # docker compose up --build -d
make up-back   # server roles and infrastructure only, skips the Android APK build
make up-front  # up-back plus the web client
make logs      # follow the logs of every container
make down      # stop the stack (data volumes are kept)
```

## Testing

```bash
# Server tests need a throwaway PostgreSQL and Redis; without them those tests are skipped.
cd server && DATABASE_URL=postgres://user:pass@localhost:5432/postgres?sslmode=disable \
  REDIS_URL=redis://localhost:6379/0 go test -race ./...
make lint   # lint of the server, web, mobile and the OpenAPI contract
```

## License

Add the license here.