SHELL := /bin/bash

COMPOSE_FILE := platform/docker-compose/docker-compose.yml
ENV_FILE ?= .env
ENV_FILE_ABS := $(abspath $(ENV_FILE))
API_DIR := services/api
GO_ENV := GOTELEMETRY=off GOCACHE=$(CURDIR)/$(API_DIR)/.gocache GOMODCACHE=$(CURDIR)/$(API_DIR)/.gomodcache GOTMPDIR=$(CURDIR)/$(API_DIR)/.tmp

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
	./platform/scripts/migrate-up.sh --compose-file "$(COMPOSE_FILE)" --env-file "$(ENV_FILE)"

run-api:
	mkdir -p "$(API_DIR)/.gocache" "$(API_DIR)/.gomodcache" "$(API_DIR)/.tmp"
	set -a; source "$(ENV_FILE_ABS)"; set +a; cd "$(API_DIR)" && $(GO_ENV) go run ./cmd/api

test:
	mkdir -p "$(API_DIR)/.gocache" "$(API_DIR)/.gomodcache" "$(API_DIR)/.tmp"
	cd "$(API_DIR)" && $(GO_ENV) go test ./...
