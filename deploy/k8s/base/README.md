# Kubernetes Base

This directory contains the base Kubernetes manifests for the Scholarship Platform MVP.

The base includes:

- namespace
- API deployment and service
- PostgreSQL deployment and service
- Redis deployment and service
- ConfigMap
- example Secret manifest
- Ingress

Apply the base with:

```bash
kubectl apply -k deploy/k8s/base
```

Before applying:

- copy `secret.example.yaml` to a real secret manifest or create the secret separately
- set the API image reference in `api-deployment.yaml` to either a local registry image or a Docker Hub image
