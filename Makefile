COMPOSE_FILE := platform/docker-compose/docker-compose.yml
ENV_FILE ?= .env
API_DIR := services/api

.PHONY: help db-up db-down migrate-up run-api test

help:
	@echo "Scholarship Platform MVP"
	@echo "make db-up       - start local PostgreSQL and Redis"
	@echo "make db-down     - stop local PostgreSQL and Redis"
	@echo "make migrate-up  - apply SQL migrations to local PostgreSQL"
	@echo "make run-api     - run the API locally"
	@echo "make test        - run API tests"

db-up:
	docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d postgres redis

db-down:
	docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) down

migrate-up:
	powershell -ExecutionPolicy Bypass -File platform/scripts/migrate-up.ps1 -ComposeFile "$(COMPOSE_FILE)" -EnvFile "$(ENV_FILE)"

run-api:
	powershell -ExecutionPolicy Bypass -Command "Set-Location '$(API_DIR)'; New-Item -ItemType Directory -Force .gocache, .gomodcache, .tmp | Out-Null; $$env:GOTELEMETRY='off'; $$env:GOCACHE=(Resolve-Path .gocache); $$env:GOMODCACHE=(Resolve-Path .gomodcache); $$env:GOTMPDIR=(Resolve-Path .tmp); go run ./cmd/api"

test:
	powershell -ExecutionPolicy Bypass -Command "Set-Location '$(API_DIR)'; New-Item -ItemType Directory -Force .gocache, .gomodcache, .tmp | Out-Null; $$env:GOTELEMETRY='off'; $$env:GOCACHE=(Resolve-Path .gocache); $$env:GOMODCACHE=(Resolve-Path .gomodcache); $$env:GOTMPDIR=(Resolve-Path .tmp); go test ./..."
