# Scholarship Platform API

This service is the single HTTP API for the Scholarship Platform MVP.

The scaffold is intentionally light:

- Chi router
- explicit package boundaries
- PostgreSQL-backed scholarship repository
- minimal scholarship handlers and services
- no real Redis cache logic yet
- no real AI provider calls yet

Key packages:

- `cmd/api` for process startup
- `internal/app` for dependency wiring and server lifecycle
- `internal/config` for environment-driven configuration
- `internal/auth` for auth primitives and token parsing scaffolding
- `internal/cache` for scholarship cache boundaries
- `internal/ai` for motivation letter provider boundaries
- `internal/repository` for persistence interfaces and stubs
- `internal/service` for application service boundaries
- `internal/health` for liveness and readiness helpers
- `internal/metrics` for metrics endpoint scaffolding
- `internal/http` for routing
- `internal/http/handlers` for route handlers
- `internal/http/middleware` for auth and request middleware
- `internal/http/response` for JSON response helpers

Currently implemented:

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
