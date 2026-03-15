.PHONY: up down build dev dev-api dev-web api start lint lint-go lint-ts setup

# Go variables
GOLANGCI_LINT_VERSION := v1.56.2

# ── Infrastructure ─────────────────────────────────────────────────────────────
up:
	docker-compose up -d

down:
	docker-compose down

# ── Go API ─────────────────────────────────────────────────────────────────────
api:
	@echo "Starting Go API on :8080..."
	cd services/api && go run .

# ── Frontend ───────────────────────────────────────────────────────────────────
dev-web:
	@echo "Starting Next.js on :3000..."
	cd apps/web && npm run dev

# ── Combined dev (infrastructure + API + frontend) ────────────────────────────
start: up
	@echo "Waiting for Postgres to be ready..."
	sleep 3
	@echo "Starting Go API in background..."
	cd services/api && go run . &
	@echo "Starting Next.js in foreground..."
	cd apps/web && npm run dev

# ── Legacy aliases ─────────────────────────────────────────────────────────────
build:
	npm run build

dev: dev-web

# ── Lint ───────────────────────────────────────────────────────────────────────
lint: lint-go lint-ts

lint-go:
	golangci-lint run ./...

lint-ts:
	npm run lint

# ── Setup ──────────────────────────────────────────────────────────────────────
setup:
	npm install
	@echo "Installing/updating golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
