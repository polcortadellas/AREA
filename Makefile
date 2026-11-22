# AREA monorepo: server/ (Go), web/ (React + Vite), mobile/ (React Native, Android).
# `make lint` must exit 0 before a PR is opened (M-01 "done when").

.PHONY: install lint lint-server lint-web lint-mobile lint-api format format-server format-web format-mobile env up up-back up-front down logs generate

# Docker Compose
env:
	@test ! -f .env || { echo ".env already exists, remove it first"; exit 1; }
	@while IFS= read -r line; do \
		case "$$line" in \
			*=change-me) echo "$${line%change-me}$$(openssl rand -hex 24)" ;; \
			*) echo "$$line" ;; \
		esac; \
	done < .env.example > .env
	@echo ".env created with random secrets"

BACK = postgres redis mailpit api worker scheduler

up:
	docker compose up --build -d

# Skip the Android APK build, which takes minutes
up-back:
	docker compose up --build -d $(BACK)

up-front:
	docker compose up --build -d $(BACK) web

down:
	docker compose down

logs:
	docker compose logs -f

# Regenerates server/internal/db/sqlc from internal/db/queries and the migrations
generate:
	cd server && sqlc generate

install:
	cd server && go mod download
	cd web && npm install
	cd mobile && npm install

lint: lint-server lint-web lint-mobile lint-api

# golangci-lint also runs go vet; `fmt --diff` fails if gofmt/goimports would change a file.
lint-server:
	cd server && golangci-lint run ./... && golangci-lint fmt --diff ./...

# ESLint (0 warnings allowed), tsc type-check, Prettier check.
lint-web:
	cd web && npm run lint && npm run typecheck && npm run format:check

lint-mobile:
	cd mobile && npm run lint && npm run typecheck && npm run format:check

# OpenAPI contract; rules in server/api/redocly.yaml.
lint-api:
	cd server/api && REDOCLY_TELEMETRY=off npx --yes @redocly/cli@2.73.0 lint openapi.yaml

format: format-server format-web format-mobile

format-server:
	cd server && golangci-lint fmt ./...

format-web:
	cd web && npm run format

format-mobile:
	cd mobile && npm run format
