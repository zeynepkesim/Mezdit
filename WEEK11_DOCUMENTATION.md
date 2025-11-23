# Week 11 - Kubernetes Observability & Storage Documentation

## 1. Health & Readiness Endpoints

- `/healthz`: Checks if the application is running (always returns 200 OK)
- `/readyz`: Returns 200 OK if the database connection is healthy, 503 otherwise
- Code: `internal/handlers/health.go`, `main.go`
- Test: After `kubectl port-forward svc/devops-backend-service -n emzops-staging 8080:80`, run `curl http://localhost:8080/healthz` and `curl http://localhost:8080/readyz`

## 2. Kubernetes Probes

- `livenessProbe`: `/healthz` endpoint, port 8080
- `readinessProbe`: `/readyz` endpoint, port 8080
- Location: `k8s/deployment.yaml`

## 3. Metrics Endpoint

- `/metrics`: Exposes metrics in Prometheus format
- Metrics:
  - `app_requests_total`: Total HTTP requests
  - `app_errors_total`: Total errors
  - `app_request_latency_microseconds`: Total latency
  - `app_posts_total`: Total post count
  - `app_users_total`: Total user count
- Code: `internal/handlers/metrics.go`, `main.go`

## 4. Prometheus Scraping

- Service annotation:
  - `prometheus.io/scrape: "true"`
  - `prometheus.io/port: "8080"`
  - `prometheus.io/path: "/metrics"`
- Service ports: both 80 and 8080 (for metrics)
- Location: `k8s/service.yaml`

## 5. Storage

- PostgreSQL uses StatefulSet and PersistentVolumeClaim (PVC)
- By default, Longhorn storage class is used (since storageClassName is not specified)
- Data is preserved even if the pod restarts
- Location: `k8s/postgres-statefulset.yaml`

## 6. Grafana Dashboard

- Accessible at: https://monitoring.kubernetes.devops.cs.ut.ee/
- Panel suggestions:
  - 4 custom metrics (app_requests_total, app_errors_total, app_request_latency_microseconds, app_posts_total/app_users_total)
  - 2 status panels (application and database up)
  - 1 storage panel (PVC/S3 usage)
  - 2 resource usage panels (CPU/Memory)
- If your team folder does not appear, wait 10 minutes; if it still does not appear, notify the assistants

## 7. References
- Code: `main.go`, `internal/handlers/health.go`, `internal/handlers/metrics.go`, `internal/database/user.go`, `k8s/deployment.yaml`, `k8s/service.yaml`, `k8s/postgres-statefulset.yaml`
- Test: `curl`, `kubectl port-forward`, Prometheus and Grafana interfaces

---
> **Note:** This document was prepared with the assistance of an AI (artificial intelligence) tool.
