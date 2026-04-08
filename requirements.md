# Scholarship Platform MVP - Phase 1 Requirements

## 1. Project Overview

Scholarship Platform MVP is a small but real-world application designed to demonstrate backend and platform engineering skills while still feeling like a usable product. The platform allows users to browse scholarships, authenticate, bookmark scholarships, manage scholarship applications, and generate a motivation letter draft with AI assistance.

Phase 1 is intentionally documentation-first. The goal is to define a production-leaning but simple architecture before writing application code.

This MVP will use:

- one Go API service
- Chi router
- PostgreSQL as the system of record
- Redis only for caching scholarship read responses
- JWT-based authentication
- role-based authorization
- Docker multi-stage builds
- GitHub Actions CI
- Kubernetes manifests
- ArgoCD GitOps deployment
- Terraform for AWS and EKS
- Prometheus and Grafana for observability

## 2. Phase 1 Scope

Phase 1 includes:

- architecture definition
- API responsibility definition
- authentication strategy
- middleware strategy
- data model definition
- caching strategy
- AI integration scope
- request and data flow definition
- deployment flow definition
- repository structure proposal
- creation of `requirements.md`
- creation of `agents.md`

Phase 1 excludes:

- Go service implementation
- database migrations
- frontend application code
- background workers
- queues
- scraping
- OAuth
- detailed Terraform module implementation
- detailed Kubernetes manifest implementation

## 3. Product Goals

The MVP should demonstrate:

- a clean and interview-explainable backend architecture
- practical API design and access control
- correct use of PostgreSQL and Redis with clear boundaries
- production-oriented containerization and deployment thinking
- early observability and operational clarity

The MVP should not demonstrate:

- distributed systems complexity for its own sake
- event-driven architecture
- multi-service orchestration
- speculative scale-driven abstractions

## 4. Functional Scope

### 4.1 Public Routes

- `GET /scholarships`
- `GET /scholarships/{id}`
- `GET /healthz`
- `GET /readyz`
- `GET /metrics`

### 4.2 Authentication Routes

- `POST /auth/register`
- `POST /auth/login`

### 4.3 Protected User Routes

- `GET /me`
- `POST /me/bookmarks/{scholarshipId}`
- `GET /me/bookmarks`
- `DELETE /me/bookmarks/{scholarshipId}`
- `POST /me/applications`
- `GET /me/applications`
- `PATCH /me/applications/{id}`

### 4.4 AI Route

- `POST /ai/motivation-letter-draft`

### 4.5 Admin Routes

- `POST /admin/scholarships`
- `PATCH /admin/scholarships/{id}`
- `DELETE /admin/scholarships/{id}`

## 5. High-Level Architecture

### 5.1 Main Components

- `api-service`
  - the only application service in the MVP
  - serves all HTTP routes
  - handles authentication, authorization, validation, business logic, and metrics
  - reads and writes PostgreSQL
  - reads and writes Redis cache entries for scholarship read endpoints
  - calls an external AI provider for draft generation

- `postgres`
  - primary database
  - source of truth for users, scholarships, bookmarks, and applications

- `redis`
  - cache only
  - stores short-lived cached scholarship list and scholarship detail responses

- `prometheus`
  - scrapes API metrics

- `grafana`
  - dashboards for API health, latency, errors, auth activity, AI usage, and database/cache behavior

- `kubernetes on EKS`
  - runs the API and supporting platform components

- `terraform`
  - provisions AWS infrastructure including EKS and related resources

- `argocd`
  - reconciles Kubernetes manifests from Git

### 5.2 Architecture Principles

- keep the application as one deployable Go service
- keep PostgreSQL authoritative for all business data
- use Redis only to accelerate scholarship reads
- keep auth explicit and stateless at the API layer
- keep route responsibilities and package boundaries clear
- prefer synchronous request-response flows over background processing
- make every major design choice easy to explain in an interview

## 6. API Responsibilities

### 6.1 Public Scholarship Reads

The public scholarship endpoints are responsible for:

- listing scholarships with stable filtering and pagination later if needed
- returning scholarship detail records
- serving anonymous traffic safely
- using Redis cache where appropriate

The public scholarship endpoints are not responsible for:

- mutating scholarship state
- exposing private admin-only fields

### 6.2 Authentication

The auth endpoints are responsible for:

- registering users
- validating login credentials
- issuing signed access tokens

The auth layer should remain intentionally simple:

- email and password registration only
- no OAuth
- no password reset flow in Phase 1
- no refresh-token family management in Phase 1 unless added later

### 6.3 User Account Features

The user routes are responsible for:

- returning the authenticated user's profile
- managing bookmark relationships
- creating scholarship applications
- listing a user's own applications
- updating application state within allowed rules

### 6.4 AI Feature

The AI route is responsible for:

- generating a motivation letter draft from user-provided context
- enforcing authentication
- validating prompt inputs before calling the AI provider
- returning generated text without persisting sensitive prompt history by default in Phase 1

The AI route is not responsible for:

- autonomous application submission
- scraping scholarships
- long-running agent workflows

### 6.5 Admin Scholarship Management

The admin routes are responsible for:

- creating scholarship records
- updating scholarship records
- deleting or soft-deleting scholarship records depending on final implementation choice

These routes must require an admin role and must invalidate affected scholarship cache entries.

## 7. Authentication Strategy

### 7.1 Identity Model

Users authenticate with:

- unique email
- password hashed with a modern password hashing algorithm such as bcrypt or argon2id

Each user record should include a role:

- `user`
- `admin`

### 7.2 Token Strategy

The MVP should use stateless bearer tokens, preferably JWTs signed by the API service.

Recommended claims:

- subject as user ID
- email
- role
- issued at
- expiration

Recommended MVP behavior:

- issue short-lived access tokens
- validate signature and expiration in middleware
- do not introduce server-side session storage unless a later phase truly needs revocation semantics

This approach fits the MVP because:

- it keeps the API stateless
- it is easy to explain
- it avoids unnecessary infrastructure

### 7.3 Authorization Strategy

Authorization should be route-driven and explicit:

- public routes require no auth
- user routes require any authenticated user
- admin routes require authenticated users with role `admin`

Ownership checks are required for user-scoped data:

- users can only read and modify their own bookmarks
- users can only read and modify their own applications

## 8. Middleware Strategy

The API should use a small, explicit middleware stack.

Recommended middleware responsibilities:

- request ID injection
- structured request logging
- panic recovery
- timeout handling
- real IP or forwarded header handling when deployed behind ingress
- authentication parsing for protected routes
- authorization enforcement for admin routes
- content type and JSON handling where helpful
- Prometheus instrumentation

Middleware should stay thin:

- authentication middleware should validate and attach identity context
- authorization middleware should only enforce route access rules
- business validation should remain in handlers or service layer

## 9. Data Model

The initial data model contains:

- `users`
- `scholarships`
- `bookmarks`
- `applications`

### 9.1 Users

Purpose:

- stores account identity and role information

Suggested fields:

- `id`
- `email`
- `password_hash`
- `full_name`
- `role`
- `created_at`
- `updated_at`

Constraints:

- unique email
- non-null password hash
- constrained role values

### 9.2 Scholarships

Purpose:

- stores scholarship records shown publicly and managed by admins

Suggested fields:

- `id`
- `title`
- `provider`
- `description`
- `amount`
- `currency`
- `deadline_at`
- `eligibility`
- `application_url`
- `status`
- `created_by`
- `created_at`
- `updated_at`

Suggested status values:

- `draft`
- `published`
- `archived`

Public reads should only expose scholarships that are safe to show publicly, typically `published`.

### 9.3 Bookmarks

Purpose:

- stores which scholarships a user has bookmarked

Suggested fields:

- `user_id`
- `scholarship_id`
- `created_at`

Constraints:

- composite uniqueness on `user_id` and `scholarship_id`
- foreign keys to users and scholarships

### 9.4 Applications

Purpose:

- stores user-submitted scholarship applications or application tracking records

Suggested fields:

- `id`
- `user_id`
- `scholarship_id`
- `status`
- `motivation_letter`
- `notes`
- `submitted_at`
- `created_at`
- `updated_at`

Suggested application status values:

- `draft`
- `submitted`
- `withdrawn`

MVP rule:

- users should only manage their own applications
- updates should enforce a limited, explicit state model

### 9.5 Indexing Guidance

At minimum, the schema should support indexes for:

- `users.email`
- `scholarships.status`
- `scholarships.deadline_at`
- `bookmarks.user_id`
- `applications.user_id`
- `applications.scholarship_id`
- a unique composite index on `bookmarks(user_id, scholarship_id)`

## 10. Caching Strategy

Redis should be used only for scholarship reads.

### 10.1 What To Cache

- `GET /scholarships`
- `GET /scholarships/{id}`

### 10.2 What Not To Cache

- authenticated user responses
- bookmarks
- applications
- auth responses
- AI responses

### 10.3 Cache Behavior

Recommended cache pattern:

- cache-aside

Flow:

1. request arrives for scholarship list or detail
2. API checks Redis for a cache key
3. on hit, API returns cached response
4. on miss, API reads PostgreSQL
5. API serializes and stores response in Redis with a short TTL
6. API returns the fresh response

### 10.4 Cache Invalidation

Admin scholarship mutations must invalidate relevant cache keys:

- on create, invalidate scholarship list cache
- on update, invalidate scholarship detail cache for that ID and invalidate affected list cache
- on delete or archive, invalidate scholarship detail cache for that ID and invalidate affected list cache

Recommended TTL:

- short-lived, such as 1 to 5 minutes

This keeps the implementation simple and operationally safe. If Redis is unavailable, the API should still function by falling back to PostgreSQL reads.

## 11. AI Integration Scope

The AI feature should stay tightly scoped.

### 11.1 Feature Goal

Generate a first draft of a motivation letter using:

- user profile inputs
- scholarship context
- user-provided achievements, goals, or background

### 11.2 MVP Behavior

- require authentication
- accept structured input in the request body
- optionally load scholarship context from PostgreSQL if a scholarship ID is provided
- call an external AI provider synchronously
- return generated text directly to the client

### 11.3 Constraints

- no autonomous workflows
- no background jobs
- no vector databases
- no long-term conversation state
- no provider abstraction layers beyond what is needed to keep the integration clean

### 11.4 Safety And Product Boundaries

- validate input size
- avoid storing generated drafts automatically unless the user chooses to use them in an application flow later
- return provider failures as clear application errors
- instrument request latency and failure rate

## 12. Data Flow

### 12.1 Public Scholarship Read Flow

1. client requests scholarship list or detail
2. API checks Redis cache
3. on cache hit, API returns cached response
4. on cache miss, API queries PostgreSQL
5. API stores the response in Redis with TTL
6. API returns the response

### 12.2 Registration Flow

1. client submits registration data
2. API validates payload
3. API hashes password
4. API inserts user into PostgreSQL
5. API returns a success response, optionally with an access token

### 12.3 Login Flow

1. client submits email and password
2. API loads user by email from PostgreSQL
3. API verifies password hash
4. API signs and returns an access token

### 12.4 Bookmark Flow

1. authenticated user requests bookmark create or delete
2. auth middleware resolves identity
3. API writes bookmark change in PostgreSQL
4. API returns updated result

### 12.5 Application Flow

1. authenticated user creates an application
2. API validates ownership and scholarship eligibility constraints that exist in the MVP
3. API inserts the application in PostgreSQL
4. API returns the created application

For updates:

1. authenticated user requests application update
2. API loads the target application from PostgreSQL
3. API verifies the application belongs to the caller
4. API enforces allowed status transitions or editable fields
5. API updates PostgreSQL
6. API returns the updated application

### 12.6 AI Draft Flow

1. authenticated user sends motivation letter draft request
2. API validates payload and loads scholarship context if needed
3. API assembles a prompt from trusted application data and user inputs
4. API sends the request to the AI provider
5. API returns generated draft text
6. Prometheus captures latency and failure metrics

## 13. Deployment Flow

### 13.1 Source Of Truth

GitHub is the source of truth for:

- application code
- Dockerfiles
- Kubernetes manifests
- ArgoCD application definitions
- Terraform code
- observability assets

### 13.2 CI With GitHub Actions

On pull request:

- run formatting and linting
- run unit tests
- build the Go service
- optionally validate Docker build
- validate Terraform formatting and static checks

On merge to the main branch:

- build the Go binary
- build and tag the Docker image
- push the image to Docker Hub with immutable tags
- update the deployment manifest or values with the new image tag

### 13.3 CD With ArgoCD

Recommended flow:

1. GitHub Actions publishes a new image to Docker Hub
2. GitHub Actions commits the new image tag into Git-managed deployment config
3. ArgoCD detects the Git change
4. ArgoCD syncs the API deployment into EKS
5. Kubernetes rolls out the new version
6. Prometheus and Grafana are used to confirm application health

### 13.4 Infrastructure With Terraform

Terraform should provision:

- VPC and networking
- EKS cluster
- node groups
- IAM roles and policies
- supporting resources for application deployment

Phase 1 expectation:

- Terraform structure should be defined clearly
- actual infrastructure code can come in later phases

## 14. Observability Expectations

The API should expose:

- `GET /healthz` for liveness
- `GET /readyz` for readiness
- `GET /metrics` for Prometheus

Recommended first metrics:

- HTTP request count
- HTTP request duration
- HTTP response status counts
- database query latency where feasible
- Redis cache hit count
- Redis cache miss count
- auth login success and failure counts
- AI request count
- AI request latency
- AI error count

Grafana dashboards should emphasize:

- API latency
- API throughput
- error rate
- cache effectiveness
- database connectivity health
- deployment health

## 15. Repository Structure Proposal

The project should start as a monorepo with one application service and clear platform folders.

```text
.
|-- agents.md
|-- requirements.md
|-- README.md
|-- .env.example
|-- Makefile
|-- services/
|   `-- api/
|       |-- cmd/
|       |   `-- api/
|       |-- internal/
|       |   |-- app/
|       |   |-- auth/
|       |   |-- config/
|       |   |-- http/
|       |   |   |-- handlers/
|       |   |   |-- middleware/
|       |   |   `-- response/
|       |   |-- repository/
|       |   |-- service/
|       |   |-- cache/
|       |   |-- ai/
|       |   |-- telemetry/
|       |   `-- platform/
|       |-- migrations/
|       |-- go.mod
|       |-- go.sum
|       `-- Dockerfile
|-- deploy/
|   |-- kubernetes/
|   |   |-- base/
|   |   `-- overlays/
|   |       |-- dev/
|   |       |-- staging/
|   |       `-- prod/
|   `-- argocd/
|-- infra/
|   `-- terraform/
|       |-- envs/
|       |   |-- dev/
|       |   |-- staging/
|       |   `-- prod/
|       `-- modules/
|-- observability/
|   |-- prometheus/
|   `-- grafana/
|-- .github/
|   `-- workflows/
`-- docs/
    |-- architecture/
    `-- runbooks/
```

### Repository Structure Rationale

- `services/api` contains the only application service for the MVP
- `internal/auth` keeps token and password logic explicit
- `internal/http/middleware` isolates cross-cutting concerns cleanly
- `internal/repository` contains PostgreSQL data access
- `internal/cache` contains Redis scholarship caching logic only
- `internal/ai` contains the motivation letter provider integration
- `deploy/` stays separate from Terraform to keep GitOps clean
- `observability/` stores dashboards and Prometheus config alongside the repo

## 16. Delivery Strategy For Later Phases

Recommended implementation order after Phase 1:

1. scaffold the Go API service with health, readiness, and metrics
2. add PostgreSQL wiring and migrations
3. implement auth and protected user routes
4. implement scholarship admin and public read routes
5. add Redis caching for scholarship reads
6. add AI motivation letter draft integration
7. add Docker, CI, Kubernetes, ArgoCD, and Terraform assets

This order keeps the application usable early while preserving clean boundaries.

## 17. Key Decisions Summary

- one Go API service only
- no workers
- no queues
- PostgreSQL as the system of record
- Redis only for scholarship read caching
- JWT auth with explicit roles
- synchronous AI draft generation through the API
- GitHub Actions for CI
- Docker Hub for images
- ArgoCD for GitOps deployment
- Terraform for AWS and EKS provisioning

## 18. Phase 1 Deliverables

- architecture definition
- API responsibility definition
- auth strategy
- middleware strategy
- data model definition
- caching strategy
- AI integration scope
- data flow definition
- deployment flow definition
- repository structure proposal
- `requirements.md`
- `agents.md`
