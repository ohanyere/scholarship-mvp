# Scholarship Platform MVP Architecture

This document summarizes the intended repository layout for the Scholarship Platform MVP.

## Application Layer

- `services/api` is reserved for the single Go API service.
- The API will handle public scholarship reads, authentication, user features, admin scholarship management, and AI draft generation.

## Platform Layer

- `platform/db/migrations` is reserved for PostgreSQL schema migrations.
- `platform/scripts` is reserved for local development and operational helper scripts.

## Deployment Layer

- `deploy/k8s/base` is reserved for shared Kubernetes manifests.
- `deploy/k8s/overlays/local` is reserved for local environment overlays.
- `deploy/k8s/overlays/prod` is reserved for production overlays.
- `deploy/argocd` is reserved for ArgoCD application definitions.

## Infrastructure Layer

- `infra/terraform/modules` is reserved for reusable Terraform modules.
- `infra/terraform/environments/dev` is reserved for the development environment entrypoint.
- `infra/terraform/environments/prod` is reserved for the production environment entrypoint.

## Observability Layer

- `observability/prometheus` is reserved for Prometheus configuration.
- `observability/grafana/dashboards` is reserved for Grafana dashboards.
- `observability/grafana/datasources` is reserved for Grafana datasource configuration.

## Documentation And Automation

- `docs/diagrams` is reserved for architecture and flow diagrams.
- `.github/workflows` is reserved for GitHub Actions workflows.

## Notes

- This phase creates structure and placeholders only.
- No application or business logic is implemented yet.
- Legacy directories that do not match the MVP may still exist until a cleanup phase removes them.
