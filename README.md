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
- production-leaning API Docker build
- simple GitHub Actions CI

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
- Docker Desktop or Docker Engine with the Docker Compose plugin
- GNU Make
- Bash, which is available by default in WSL Ubuntu and GitHub Actions Linux runners

## Local Startup Flow

1. Copy the example environment file:

```bash
cp .env.example .env
```

2. Start PostgreSQL and Redis:

```bash
make db-up
```

3. Apply migrations:

```bash
make migrate-up
```

4. Start the API:

```bash
make run-api
```

5. In another terminal, verify the public scholarship endpoints:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/scholarships
```

6. Register a user and log in:

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","full_name":"Scholarship User"}'

curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
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

## CI Pipeline

The GitHub Actions pipeline is intentionally simple but DevSecOps-oriented:

- run `go test ./...`
- run `go vet ./...`
- lint the API Dockerfile with Hadolint
- scan Terraform and Kubernetes manifests with Checkov
- build the API Docker image
- scan the built image with Trivy and fail on `HIGH` and `CRITICAL`
- on pushes to `main`, push the image to a registry
- update `deploy/k8s/base/api-deployment.yaml` with the new image tag and commit that change back to Git

Trigger rules:

- pull requests run lint, tests, scans, and image build only
- pushes to `main` run the full pipeline, including image push and GitOps manifest update

Image tagging strategy:

- every published image gets a commit-based `sha-<full-sha>` tag
- pushes from `main` also get a stable `main` tag

This keeps deployments reviewable and reproducible without relying on `latest` alone.

## Required GitHub Secrets

To enable registry push support, configure these repository secrets:

- `CONTAINER_REGISTRY`
  Example: `docker.io` or `ghcr.io`
- `CONTAINER_REGISTRY_USERNAME`
  Example: Docker Hub username or GitHub username
- `CONTAINER_REGISTRY_TOKEN`
  Example: Docker Hub access token or GHCR token
- `IMAGE_REPOSITORY`
  Example: `your-user/scholarship-platform-api` or `your-org/scholarship-platform-api`

This keeps the workflow usable with Docker Hub or GHCR without maintaining two separate pipelines.

## Kubernetes Base Manifests

Base manifests live in `deploy/k8s/base` and are designed to stay simple and readable:

- one API Deployment and Service
- one PostgreSQL Deployment and Service
- one Redis Deployment and Service
- one ConfigMap
- one example Secret manifest
- one Ingress

Apply the base with:

```bash
kubectl apply -k deploy/k8s/base
```

Before applying:

1. create a real secret from `deploy/k8s/base/secret.example.yaml`
2. set the API image in `deploy/k8s/base/api-deployment.yaml`

Image reference guidance:

- for shared environments, use a Docker Hub image such as `docker.io/<your-user>/scholarship-platform-api:main` or a commit-based tag
- for local clusters, point the manifest at an image already loaded into the cluster runtime or your local registry
- prefer immutable commit-based tags for repeatable deployments, and use `main` only as a convenience tag

## ArgoCD Manifests

ArgoCD manifests live in `deploy/argocd` and stay intentionally simple:

- one `AppProject`
- one `Application`
- the Application points directly at `deploy/k8s/base`

The ArgoCD Application targets:

- cluster: `https://kubernetes.default.svc`
- namespace: `scholarship-platform`
- revision: `main`

Before applying the ArgoCD manifests:

1. update the GitHub repository URL in `deploy/argocd/project.yaml`
2. update the same repository URL in `deploy/argocd/application.yaml`

## GitOps Image Updates

The intended GitOps flow is:

1. GitHub Actions builds and pushes a new API image on `main`
2. the workflow updates the image tag in `deploy/k8s/base/api-deployment.yaml`
3. ArgoCD detects the Git change
4. ArgoCD syncs the updated manifests into the cluster

That means ArgoCD still treats Git as the source of truth. The image tag should be changed in Git, not manually in the cluster.

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

The runner is a bash script at `platform/scripts/migrate-up.sh`. It loads `.env`
with standard shell syntax, so keep environment values in `KEY=value` form.

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

- `platform/scripts/migrate-up.sh`
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

- `services/api/Dockerfile`
  Defines the production-leaning multi-stage API image build.

- `.github/workflows/api-ci.yml`
  Defines the DevSecOps workflow for linting, testing, IaC scanning, image scanning, image publishing, and GitOps manifest updates.

- `.env.example`
  Shows the environment variables needed by both the API and the local PostgreSQL container.

- `Makefile`
  Provides the local developer entrypoints for database startup, migrations, API startup, and tests.
