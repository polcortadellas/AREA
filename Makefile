# AREA monorepo: server/ (Go), web/ (React + Vite), mobile/ (React Native, Android).
# `make lint` must exit 0 before a PR is opened (M-01 "done when").

.PHONY: install lint lint-server lint-web lint-mobile lint-api format format-server format-web format-mobile env up down logs

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

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

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
