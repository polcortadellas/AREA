# AREA monorepo: server/ (Go), web/ (React + Vite), mobile/ (React Native, Android).
# `make lint` must exit 0 before a PR is opened (M-01 "done when").

.PHONY: install lint lint-server lint-web lint-mobile format format-server format-web format-mobile

install:
	cd server && go mod download
	cd web && npm install
	cd mobile && npm install

lint: lint-server lint-web lint-mobile

# golangci-lint also runs go vet; `fmt --diff` fails if gofmt/goimports would change a file.
lint-server:
	cd server && golangci-lint run ./... && golangci-lint fmt --diff ./...

# ESLint (0 warnings allowed), tsc type-check, Prettier check.
lint-web:
	cd web && npm run lint && npm run typecheck && npm run format:check

lint-mobile:
	cd mobile && npm run lint && npm run typecheck && npm run format:check

format: format-server format-web format-mobile

format-server:
	cd server && golangci-lint fmt ./...

format-web:
	cd web && npm run format

format-mobile:
	cd mobile && npm run format
