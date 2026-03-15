.PHONY: up down build dev lint lint-go lint-ts setup

# Go variables
GOLANGCI_LINT_VERSION := v1.56.2

up:
	docker-compose up -d

down:
	docker-compose down

build:
	npm run build

dev:
	npm run dev

lint: lint-go lint-ts

lint-go:
	golangci-lint run ./...

lint-ts:
	npm run lint

setup:
	npm install
	@echo "Installing/updating golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
