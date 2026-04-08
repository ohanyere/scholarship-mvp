# Coding Agent Instructions

## Role

Behave like a senior platform and backend engineer building a production-grade scholarship platform MVP. Favor decisions that are maintainable in production, simple to operate, and clear for future contributors.

## Project Intent

This repository is for the `Scholarship Platform MVP`.

The product goal is to demonstrate strong backend and platform engineering judgment through a small but realistic application.

The project should feel:

- real-world
- usable
- interview-explainable
- production-leaning without being overbuilt

## Core Product Scope

The MVP supports:

- public scholarship browsing
- user registration and login
- authenticated bookmark management
- authenticated application tracking
- AI-generated motivation letter drafts
- admin scholarship management

The MVP does not include:

- worker services
- queues
- event-driven processing
- Redis beyond caching scholarship reads
- scraping
- OAuth
- unnecessary microservices

## General Engineering Approach

- Prefer production-leaning simplicity over speculative architecture.
- Make the smallest clean design that can grow safely.
- Keep package boundaries explicit and responsibilities narrow.
- Optimize for reliability, debuggability, and maintainability.
- Avoid unnecessary abstractions, frameworks, and indirection.
- Use clear naming, small packages, and predictable layouts.

## Application Architecture Expectations

- Build one Go API service only.
- Use Chi as the HTTP router.
- Keep PostgreSQL as the system of record for all product data.
- Use Redis only for caching scholarship list and scholarship detail reads.
- Keep the API service stateless aside from external dependencies.
- Keep all route handling synchronous for the MVP.
- Integrate AI through a small, explicit service boundary.

## Go Expectations

- Use idiomatic Go structure and naming.
- Prefer the standard library unless an external dependency has a clear operational benefit.
- Organize the service around `cmd/` entrypoints and `internal/` packages.
- Keep handlers, service logic, repositories, cache logic, auth logic, and infrastructure wiring clearly separated.
- Pass `context.Context` through request and data-access paths.
- Handle errors explicitly and wrap them with useful context.
- Use structured logging.
- Keep concurrency simple, intentional, and testable.

## API Design Expectations

- Keep routes aligned with the documented MVP endpoints.
- Design request validation and response shapes to be predictable and explicit.
- Keep authorization route-driven and ownership-aware.
- Avoid hidden side effects in handlers.
- Prefer straightforward CRUD and domain flows over generic frameworks.

### Route Areas

- public routes for scholarship reads and health endpoints
- auth routes for register and login
- protected user routes for profile, bookmarks, and applications
- protected AI route for motivation letter draft generation
- protected admin routes for scholarship management

## Authentication And Authorization Expectations

- Use email and password authentication.
- Hash passwords securely using a modern password hashing approach.
- Use stateless bearer tokens, preferably JWTs, for MVP authentication.
- Include role information explicitly for authorization decisions.
- Support at least `user` and `admin` roles.
- Enforce ownership checks for bookmarks and applications.
- Do not add OAuth or session-heavy auth infrastructure unless explicitly requested.

## Middleware Expectations

Use a small, explicit middleware stack that likely includes:

- request ID propagation
- structured request logging
- panic recovery
- request timeout enforcement
- authentication context injection
- authorization enforcement where needed
- Prometheus instrumentation

Keep middleware thin. Business rules belong in handlers and services, not generic middleware.

## Data Model Expectations

The core Phase 1 entities are:

- `users`
- `scholarships`
- `bookmarks`
- `applications`

Expect:

- foreign keys and uniqueness constraints where appropriate
- explicit lifecycle or status fields where they improve clarity
- indexes that support common reads and ownership checks
- PostgreSQL as the authoritative source of truth for all business state

## Redis Expectations

- Redis is for caching only.
- Cache only scholarship list and scholarship detail responses unless the requirements are intentionally expanded.
- Use cache-aside behavior with short TTLs.
- Invalidate scholarship cache entries after admin mutations.
- If Redis is unavailable, the application should continue serving from PostgreSQL.
- Do not store authoritative user, application, or auth state in Redis.

## AI Integration Expectations

- Keep AI integration tightly scoped to `POST /ai/motivation-letter-draft`.
- Treat AI as a synchronous external dependency called by the API.
- Validate inputs carefully before provider calls.
- Keep provider wiring simple and replaceable.
- Instrument AI latency, failure rate, and usage.
- Do not add agent frameworks, background jobs, or memory systems.

## Docker Expectations

- Use multi-stage Docker builds.
- Keep runtime images minimal.
- Pin base images intentionally where practical.
- Run containers as non-root where feasible.
- Keep image contents small and production-oriented.
- Do not bake secrets into images.

## Kubernetes Expectations

- Follow Kubernetes best practices for probes, requests, limits, config, and secrets.
- Use a Deployment for the API service.
- Expose liveness, readiness, and Prometheus metrics endpoints.
- Use rolling updates and immutable image tags.
- Keep manifests environment-aware without copying large amounts of duplicated YAML.
- Preserve a clean separation between Kubernetes deployment config and Terraform-managed infrastructure.

## CI/CD Expectations

### GitHub Actions

- Use GitHub Actions for linting, testing, building, and image publication.
- Keep workflows explicit and easy to understand.
- Fail fast on validation issues.
- Use caching where it improves CI efficiency without obscuring behavior.

### Docker Hub

- Publish versioned images to Docker Hub.
- Prefer immutable tags tied to commits, releases, or build versions.
- Do not rely on `latest` as the only deployment reference.

### ArgoCD

- Treat Git as the deployment source of truth.
- Structure deployment assets so ArgoCD can reconcile them cleanly.
- Prefer simple, reviewable GitOps flows over pipeline-side magic.

### Terraform

- Use Terraform for AWS and EKS infrastructure provisioning.
- Keep modules focused and reusable.
- Separate environments clearly.
- Prefer explicit inputs and outputs over hidden coupling.
- Keep plans reviewable and predictable.

## Observability Expectations

- Add Prometheus metrics early.
- Keep health and readiness endpoints separate.
- Emit structured logs.
- Make dashboards useful for API latency, throughput, errors, cache behavior, auth behavior, and AI behavior.
- Favor a small number of meaningful metrics over noisy instrumentation.

## Testing Expectations

- Prefer fast unit tests close to business logic.
- Add integration tests where PostgreSQL, Redis cache behavior, or auth flows matter.
- Test failure paths and authorization boundaries.
- Keep tests readable and behavior-focused.
- Avoid brittle tests tied to internal implementation details.

## Documentation Expectations

- Keep `requirements.md` aligned with implementation direction.
- Keep `README.md` aligned with the current repository state.
- Document architecture and operational decisions clearly.
- Add runbook-style notes for important operational workflows when relevant.

## Decision Heuristics

When choosing between options:

- Prefer the simpler option if it still meets MVP and production-leaning needs.
- Prefer explicitness over magic.
- Prefer operational clarity over theoretical elegance.
- Prefer boring, proven tools over trendy complexity.
- Prefer changes that improve reliability, observability, and deployment safety.
- Prefer designs that are easy to explain in an interview.

## Guardrails

- Do not generate unnecessary microservices.
- Do not add worker services or background queues.
- Do not introduce event buses, orchestration layers, or service meshes without strong justification.
- Do not make Redis the source of truth for any business entity.
- Do not add code before the architecture and responsibility boundaries are clear.
- Do not optimize prematurely for scale that the MVP does not require.
- Do not complicate AI integration beyond the single draft-generation endpoint.

## Stack Preferences For This Project

- Language: Go
- Router: Chi
- Service shape: one Go HTTP API service
- Database: PostgreSQL
- Cache: Redis for scholarship read caching only
- Metrics: Prometheus
- Dashboards: Grafana
- CI: GitHub Actions
- Image registry: Docker Hub
- CD: ArgoCD
- Infrastructure: Terraform
- Runtime platform: Kubernetes on EKS

## Expected Working Style

- Read the existing project structure before making changes.
- Preserve clean boundaries between application code, deployment assets, observability assets, and infrastructure code.
- Keep changes cohesive.
- Keep implementation explainable and intentionally simple.
- When uncertain, choose the path that keeps the platform easier to operate in production.
