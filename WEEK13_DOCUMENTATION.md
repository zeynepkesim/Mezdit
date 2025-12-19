# Week 13 - Templating & Infrastructure as Code Documentation

**Team:** emzops
**Templating Tool:** Helm

## Templating Choice

We chose **Helm** for the following reasons:
1.  **Packaging:** Helm Charts allow us to package our application as a single unit (Deployment, Service, ConfigMap, etc.).
2.  **Atomic Deployments:** Helm provides built-in support for atomic rollbacks using the `--atomic` flag, which automatically reverts changes if the upgrade fails or times out.
3.  **Community Standard:** Helm is the de-facto standard package manager for Kubernetes.
4.  **Value Substitution:** Easy to inject environment-specific values (like image tags and secrets) via `values.yaml` and CLI flags.

### Structure

We created two charts:
1.  `devops-be/chart/`: Manages the Backend, Database (Postgres StatefulSet), and Storage.
2.  `devops-fe/chart/`: Manages the Frontend and Ingress.

## Secret Management

Our strategy ensures **no secrets are stored in Git**:
1.  **Templated Secrets:** We typically use K8s `Secret` objects defined in `templates/secret.yaml`.
2.  **Values placeholders:** In `values.yaml`, sensitive fields are set to empty strings (`""`).
3.  **CI/CD Injection:** Secrets are injected strictly during the CI/CD pipeline execution using environment variables stored in GitLab CI/CD Settings (masked and protected).

Example command in `.gitlab-ci.yml`:
```bash
helm upgrade --install devops-backend ./chart \
  --set secrets.backend.databaseUrl=$DATABASE_URL \
  --set secrets.postgres.password=$POSTGRES_PASSWORD
```

## Statelessness and Persistence

To survive Chaos Engineering (random pod failures), we made our application stateless:

1.  **Application State:** The Backend application code is stateless. It relies on the database for data persistence.
2.  **Database Persistence:** PostgreSQL runs as a `StatefulSet` with a `VolumeClaimTemplate` requesting `ReadWriteOnce` storage from **Longhorn**. This ensures data survives pod restarts and moves with the pod.
3.  **File Storage:** We use a separate PVC (`backend-images-storage`) also backed by Longhorn to store uploaded images. This is mounted to the backend pod.

### Readiness Probes

We use `readinessProbe` to ensure traffic is only sent to healthy pods.
- **Backend:** Checks `/readyz` (verifies DB connection).
- **Frontend:** Checks `/` (verifies web server is up).
During deployments (`--atomic`), Helm waits for these probes to succeed before marking the release as successful.
