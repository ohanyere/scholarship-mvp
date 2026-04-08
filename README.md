# Scholarship Platform MVP

This repository contains a production-leaning scholarship platform MVP built to demonstrate backend and platform engineering skills with a simple, explainable architecture.

Current implementation focus:

- one Go API service using Chi
- PostgreSQL-backed scholarship endpoints
- Redis-backed caching for scholarship read endpoints
- PostgreSQL-backed user registration and login
- PostgreSQL-backed bookmarks and applications
- local PostgreSQL via Docker Compose
- SQL-file migration workflow

Not implemented yet:

- AI provider behavior
- full auth behavior

## Repository Overview

- `services/api` contains the Go API service
- `platform/docker-compose` contains the local PostgreSQL setup
- `platform/db/migrations` contains SQL migrations
- `platform/scripts` contains developer helper scripts
- `deploy/`, `infra/`, and `observability/` contain deployment and platform scaffolding for later phases

## Local Prerequisites

- Go 1.24+
- Docker Desktop or Docker Engine with Compose support
- GNU Make

## Local Startup Flow

1. Copy the example environment file:

```powershell
Copy-Item .env.example .env
```

2. Start PostgreSQL and Redis:

```powershell
make db-up
```

3. Apply migrations:

```powershell
make migrate-up
```

4. Start the API:

```powershell
make run-api
```

5. In another terminal, verify the public scholarship endpoints:

```powershell
Invoke-WebRequest http://localhost:8080/healthz
Invoke-WebRequest http://localhost:8080/scholarships
```

6. Register a user and log in:

```powershell
$registerBody = @{
  email = "user@example.com"
  password = "password123"
  full_name = "Scholarship User"
} | ConvertTo-Json

Invoke-WebRequest http://localhost:8080/auth/register -Method POST -ContentType "application/json" -Body $registerBody

$loginBody = @{
  email = "user@example.com"
  password = "password123"
} | ConvertTo-Json

$loginResponse = Invoke-WebRequest http://localhost:8080/auth/login -Method POST -ContentType "application/json" -Body $loginBody
```

## Local Development Commands

- `make db-up`
  Starts the local PostgreSQL and Redis containers defined in `platform/docker-compose/docker-compose.yml`.

- `make db-down`
  Stops the local PostgreSQL and Redis containers.

- `make migrate-up`
  Applies any unapplied SQL files from `platform/db/migrations` into the local database.

- `make run-api`
  Runs the Go API locally from `services/api`.

- `make test`
  Runs `go test ./...` for the API module.

## Migration Workflow

The migration path is intentionally simple:

- write migrations as ordered `.sql` files in `platform/db/migrations`
- start PostgreSQL with `make db-up`
- apply unapplied migrations with `make migrate-up`

The migration runner:

- waits for PostgreSQL to become ready
- creates a `schema_migrations` table if needed
- applies SQL files in filename order
- records each applied filename so it is not re-applied

## Currently Working Endpoints

- `GET /healthz`
- `GET /readyz`
- `GET /metrics`
- `GET /scholarships`
- `GET /scholarships/{id}`
- `POST /auth/register`
- `POST /auth/login`
- `GET /me`
- `POST /me/bookmarks/{scholarshipId}`
- `GET /me/bookmarks`
- `DELETE /me/bookmarks/{scholarshipId}`
- `POST /me/applications`
- `GET /me/applications`
- `PATCH /me/applications/{id}`
- `POST /admin/scholarships`

`POST /admin/scholarships` now requires a JWT with the `admin` role. The admin role is intentionally simple for now and can be assigned directly in the database for local development.

## Important Files

- `platform/docker-compose/docker-compose.yml`
  Defines the local PostgreSQL and Redis containers for development.

- `platform/db/migrations/0001_create_scholarships.sql`
  Creates the initial `scholarships` table and its supporting index.

- `platform/scripts/migrate-up.ps1`
  Runs ordered SQL migrations against the local PostgreSQL container and records applied versions.

- `services/api/internal/app/app.go`
  Initializes config, logging, PostgreSQL, repositories, services, and the HTTP server.

- `services/api/internal/repository/postgres.go`
  Contains the concrete PostgreSQL scholarship repository implementation.

- `services/api/internal/cache/cache.go`
  Contains the Redis-backed scholarship cache implementation used only for `GET /scholarships` and `GET /scholarships/{id}`.

- `services/api/internal/repository/postgres_users.go`
  Contains the concrete PostgreSQL user repository implementation for registration, login lookup, and current-user lookup.

- `services/api/internal/repository/postgres_bookmarks.go`
  Contains the concrete PostgreSQL bookmark repository implementation for create, list, and delete.

- `services/api/internal/repository/postgres_applications.go`
  Contains the concrete PostgreSQL application repository implementation for create, list, and patch-style update.

- `services/api/internal/auth/auth.go`
  Contains password hashing, JWT generation, JWT validation, and auth claims handling.

- `.env.example`
  Shows the environment variables needed by both the API and the local PostgreSQL container.

- `Makefile`
  Provides the local developer entrypoints for database startup, migrations, API startup, and tests.
