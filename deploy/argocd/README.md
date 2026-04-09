# ArgoCD Manifests

This directory contains the ArgoCD manifests for the Scholarship Platform MVP.

Included manifests:

- `project.yaml` for the ArgoCD `AppProject`
- `application.yaml` for the ArgoCD `Application`

The Application points directly at:

- `deploy/k8s/base`

This keeps the GitOps setup intentionally simple:

- one project
- one application
- one Kubernetes base
- no app-of-apps
- no sync hooks or sync waves
