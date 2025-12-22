# DevOps Project Master Documentation

**Team:** emzops  
**Production URL:** https://devops-mezdit.kubernetes.devops.cs.ut.ee  
**Monitoring URL:** https://monitoring.kubernetes.devops.cs.ut.ee  
**Namespaces:** `emzops-staging` (Development), `emzops-prod` (Production)

---

## Table of Contents

1. [Project Overview & Architecture](#1-project-overview--architecture)
2. [Evolution Timeline (Weeks 9-16)](#2-evolution-timeline-weeks-9-16)
3. [Infrastructure & Networking](#3-infrastructure--networking)
4. [Templating & State Management (Helm)](#4-templating--state-management-helm)
5. [CI/CD & Container Security](#5-cicd--container-security)
6. [Observability, Monitoring & Alerts](#6-observability-monitoring--alerts)
7. [Production Readiness & Autoscaling](#7-production-readiness--autoscaling)
8. [Comprehensive Testing & Troubleshooting](#8-comprehensive-testing--troubleshooting)

---

## 1. Project Overview & Architecture

### 1.1 Executive Summary

This project implements a production-grade, containerized web application deployed on Kubernetes. It features a microservices architecture (Frontend/Backend), stateful persistence (PostgreSQL), automated CI/CD pipelines, full-stack observability (Prometheus/Grafana), and robust security practices.

The project has evolved through multiple phases (Weeks 9-16), starting with basic Kubernetes deployments and progressing to a fully automated, monitored, and scalable production system.

### 1.2 High-Level Architecture Diagram

```mermaid
graph TD
    User((User)) -->|HTTPS/443| Ingress[Nginx Ingress Controller]
    Ingress -->|/api/*| BackendService[Backend Service]
    Ingress -->|/*| FrontendService[Frontend Service]
    
    subgraph "Kubernetes Cluster (emzops-prod)"
        FrontendService -->|ClusterIP| FrontendPod[Frontend Pod (Vue.js + Nginx)]
        BackendService -->|ClusterIP| BackendPod[Backend Pod (Go)]
        BackendPod -->|TCP/5432| DBService[Postgres Service (Headless)]
        DBService --> DBPod[Postgres StatefulSet]
        
        BackendPod -.->|Metrics/9090| Prometheus[Prometheus]
        FrontendPod -.->|Metrics| Prometheus
    end
    
    subgraph "External Persistence"
        DBPod -->|PVC| Longhorn[Longhorn Storage]
    end
```

### 1.3 Traffic Flow & Routing Logic

1. **Ingress Layer**: 
   - Terminates TLS using Let's Encrypt certificates (`cert-manager`).
   - Routes `/api/*` and `/ws` (WebSockets) to the **Backend**.
   - Routes all other traffic (`/*`) to the **Frontend**.
   - Implements path rewriting (e.g., `/api/posts` -> `/posts`) using `nginx.ingress.kubernetes.io/rewrite-target`.
2. **Internal Communication**:
   - Frontend talks to Backend via the public Ingress URL (client-side routing).
   - Backend talks to Postgres via internal DNS (`postgres-service`).

### 1.4 Technology Stack

**Backend:**
- Language: Go 1.23
- Framework: Standard library HTTP server
- Database: PostgreSQL 15 (Alpine)
- Storage: MinIO (S3-compatible object storage)
- Container: Multi-stage Docker build (golang:1.23-alpine → alpine:latest)

**Frontend:**
- Framework: Vue.js 3
- Web Server: Nginx (Alpine)
- Container: Multi-stage Docker build (node:22-alpine → nginx:alpine)

**Infrastructure:**
- Container Orchestration: Kubernetes
- Package Manager: Helm 3
- CI/CD: GitLab CI/CD
- Container Registry: GitLab Container Registry
- Image Builder: Kaniko
- Security Scanner: Trivy
- Monitoring: Prometheus + Grafana
- Storage: Longhorn (distributed block storage)

---

## 2. Evolution Timeline (Weeks 9-16)

### Week 9: Initial Kubernetes Deployment

**Objective:** Deploy containerized applications to Kubernetes with basic networking.

**Key Achievements:**
- ✅ Container image building with Kaniko in CI/CD
- ✅ Multi-stage Docker builds for both backend and frontend
- ✅ GitLab Container Registry integration
- ✅ Private registry authentication using Kubernetes secrets
- ✅ Basic Kubernetes deployments (Deployment, Service)
- ✅ Port-forwarding for testing and verification
- ✅ Secret management (no secrets in Git)

**Architecture Changes:**
- Migrated from VM-based deployment to Kubernetes
- Introduced namespace isolation (`emzops-staging`)
- Implemented image pull secrets for private registry access

**Key Files:**
- `devops-be/Dockerfile` - Multi-stage Go build
- `devops-fe/Dockerfile` - Multi-stage Vue.js build
- `devops-be/k8s/deployment.yaml` - Backend deployment
- `devops-fe/k8s/deployment.yaml` - Frontend deployment

**Image Tagging Strategy:**
- `latest`: Most recent build from `main` branch
- `<commit-sha>`: Immutable version for rollback capability

### Week 10: Production-Ready Networking & Configuration

**Objective:** Implement production-grade networking, TLS, and externalized configuration.

**Key Achievements:**
- ✅ Kubernetes Services (ClusterIP for apps, Headless for database)
- ✅ Ingress with automatic TLS (cert-manager + Let's Encrypt)
- ✅ Path-based routing (`/api/*` → backend, `/*` → frontend)
- ✅ WebSocket support with proper annotations
- ✅ Separate PostgreSQL StatefulSet deployment
- ✅ ConfigMaps for non-sensitive configuration
- ✅ Secrets for sensitive data (database credentials, registry tokens)
- ✅ RBAC setup for GitLab CI/CD automation
- ✅ Automated deployment pipeline

**Architecture Changes:**
- Added Ingress controller for external access
- Implemented TLS termination with automatic certificate management
- Separated database into dedicated StatefulSet
- Externalized all configuration (no hardcoded values)

**Key Files:**
- `devops-be/k8s/service.yaml` - Backend service with Prometheus annotations
- `devops-be/k8s/ingress.yaml` - Ingress with TLS and path routing
- `devops-be/k8s/postgres-statefulset.yaml` - PostgreSQL StatefulSet with PVC
- `devops-be/k8s/configmap.yaml` - Non-sensitive configuration
- `devops-be/k8s/rbac.yaml` - ServiceAccount, Role, RoleBinding for CI/CD

**Configuration Management:**
- **ConfigMaps**: PORT, LOG_LEVEL, ENVIRONMENT, API URLs
- **Secrets**: DATABASE_URL, POSTGRES credentials, registry tokens
- **Strategy**: Template files in Git, actual secrets created via CI/CD

### Week 11: Observability & Persistent Storage

**Objective:** Implement comprehensive monitoring, health checks, and persistent storage.

**Key Achievements:**
- ✅ Health endpoints (`/healthz` for liveness, `/readyz` for readiness)
- ✅ Kubernetes probes (liveness and readiness)
- ✅ Prometheus metrics endpoint (`/metrics` on port 9090)
- ✅ Custom application metrics (requests, errors, latency, posts, users)
- ✅ Prometheus integration with service discovery
- ✅ NetworkPolicy for secure metrics scraping
- ✅ Grafana dashboards for visualization
- ✅ Persistent storage with StatefulSet and PVCs (Longhorn)
- ✅ Storage persistence verification (data survives pod restarts)

**Architecture Changes:**
- Added separate metrics server on port 9090
- Implemented metrics middleware for request tracking
- Created NetworkPolicy to allow only Prometheus scraping
- Upgraded database storage from emptyDir to PVC (1Gi)

**Key Files:**
- `devops-be/internal/handlers/health.go` - Health and readiness endpoints
- `devops-be/internal/handlers/metrics.go` - Prometheus metrics exposition
- `devops-be/k8s/networkpolicy.yaml` - NetworkPolicy for metrics access
- `devops-be/k8s/postgres-statefulset.yaml` - Updated with PVC template

**Metrics Implemented:**
- `app_requests_total` (Counter) - Total HTTP requests
- `app_errors_total` (Counter) - Total 4xx/5xx responses
- `app_request_latency_microseconds` (Counter) - Cumulative latency
- `app_posts_total` (Gauge) - Current number of posts
- `app_users_total` (Gauge) - Current number of users

**Probe Configuration:**
- **Liveness Probe**: `/healthz` endpoint, checks if process is running
- **Readiness Probe**: `/readyz` endpoint, checks database connectivity
- **Timing**: `initialDelaySeconds: 5`, `periodSeconds: 10`, `failureThreshold: 3`

### Week 13: Templating with Helm

**Objective:** Migrate from static YAML manifests to Helm charts for better configuration management.

**Key Achievements:**
- ✅ Helm chart structure for both backend and frontend
- ✅ Template-based resource generation
- ✅ Environment-specific values files (`values.yaml` for staging, `values-prod.yaml` for production)
- ✅ Atomic deployments with `--rollback-on-failure`
- ✅ Dynamic secret generation in templates
- ✅ Stateless application design (MinIO for file storage, PostgreSQL for data)
- ✅ Readiness probes for zero-downtime deployments

**Architecture Changes:**
- Migrated all Kubernetes manifests to Helm templates
- Introduced Helm release management
- Implemented dynamic DATABASE_URL generation based on PostgreSQL service name
- Added support for environment-specific overrides

**Key Files:**
- `devops-be/chart/Chart.yaml` - Chart metadata
- `devops-be/chart/values.yaml` - Staging default values
- `devops-be/chart/values-prod.yaml` - Production overrides
- `devops-be/chart/templates/*.yaml` - All Kubernetes resource templates
- `devops-be/.gitlab-ci.yml` - Updated with Helm deployment commands

**Helm Deployment Commands:**
```bash
# Staging
helm upgrade --install devops-backend ./chart \
  --namespace emzops-staging \
  --set image.tag=$IMAGE_TAG \
  --timeout 2m \
  --rollback-on-failure

# Production
helm upgrade --install devops-backend ./chart \
  --namespace emzops-prod \
  --values ./chart/values-prod.yaml \
  --set image.tag=$IMAGE_TAG \
  --timeout 2m \
  --rollback-on-failure
```

**Statelessness Implementation:**
- Application code is stateless (no session storage)
- File uploads stored in MinIO (external object storage)
- Database state persisted in PostgreSQL StatefulSet with PVC
- JWT tokens for stateless authentication

### Week 14: Production Deployment & Autoscaling

**Objective:** Implement production environment with automatic scaling and high availability.

**Key Achievements:**
- ✅ Separate production namespace (`emzops-prod`)
- ✅ Production-specific Helm values (higher replicas, resources)
- ✅ Horizontal Pod Autoscaler (HPA) for automatic scaling
- ✅ Pod anti-affinity for high availability
- ✅ Topology spread constraints for node distribution
- ✅ Chaos engineering readiness (zero-downtime pod kills)
- ✅ Production vs staging configuration comparison

**Architecture Changes:**
- Created production namespace with separate configuration
- Implemented HPA with CPU and memory metrics
- Added pod anti-affinity to prevent single-node failures
- Increased production resource limits and replicas

**Key Files:**
- `devops-be/chart/values-prod.yaml` - Production configuration
- `devops-be/chart/templates/hpa.yaml` - Horizontal Pod Autoscaler template
- `devops-be/chart/templates/deployment.yaml` - Updated with affinity rules
- `devops-be/.gitlab-ci.yml` - Added `deploy-to-production` job (manual trigger)

**HPA Configuration:**
- **Metric**: CPU utilization
- **Target**: 30% average CPU (low for demonstration)
- **Min Replicas**: 2
- **Max Replicas**: 5
- **Scaling Behavior**: Scale up when CPU > 30%, scale down when CPU < 30%

**Affinity Rules:**
- **Type**: `requiredDuringSchedulingIgnoredDuringExecution` (hard requirement)
- **Rule**: Pod anti-affinity - do not schedule backend pods on same node
- **Result**: Pods distributed across available nodes (e.g., `do-worker8`, `do-worker9`)

**Production vs Staging Differences:**

| Aspect | Staging | Production |
|--------|---------|------------|
| Replicas | 1 | 2 (min) - 5 (max with HPA) |
| Memory Limit | 256Mi | 512Mi |
| CPU Limit | 200m | 500m |
| PostgreSQL Storage | 1Gi | 2Gi |
| Autoscaling | Disabled | Enabled (HPA) |
| Affinity | None | Pod anti-affinity |
| Deployment | Automatic | Manual trigger |

### Week 15: Security Scanning & Production Monitoring

**Objective:** Implement container security scanning and comprehensive production monitoring dashboard.

**Key Achievements:**
- ✅ Container security scanning with Trivy
- ✅ Automated vulnerability scanning in CI/CD pipeline
- ✅ GitLab Security Dashboard integration
- ✅ Production Grafana dashboard with 12+ panels
- ✅ Visual thresholds for resource usage
- ✅ HPA visibility and monitoring
- ✅ Pod lifecycle tracking
- ✅ Service availability monitoring

**Architecture Changes:**
- Added `security-scan` stage to CI/CD pipeline
- Integrated Trivy scanner for container images
- Created comprehensive Grafana dashboard for production
- Implemented visual thresholds for resource monitoring

**Key Files:**
- `devops-be/.gitlab-ci.yml` - Added `container-security-scan` job
- `devops-fe/.gitlab-ci.yml` - Added `container-security-scan` job
- Grafana Dashboard: `emzops-prod` namespace monitoring

**Security Scanning Configuration:**
- **Tool**: Trivy (Aqua Security)
- **Severity**: CRITICAL and HIGH vulnerabilities
- **Format**: GitLab-compatible JSON report
- **Integration**: Results visible in GitLab Security tab
- **Execution**: Automatic after image build on `main` branch

**Production Dashboard Panels:**
1. **Application Health Overview**: Pods ready count (Backend, Frontend, PostgreSQL)
2. **Pod Status by Phase**: Distribution of pod phases (Running, Pending, Failed)
3. **Pod Restarts Over Time**: Time series of pod restart events
4. **HPA Current Replicas**: Current number of replicas per deployment
5. **HPA Desired Replicas**: Target replicas set by HPA
6. **Backend CPU Usage**: CPU utilization with thresholds (70% yellow, 90% red)
7. **Backend Memory Usage**: Memory usage with thresholds (400MB yellow, 480MB red)
8. **Frontend CPU Usage**: CPU utilization
9. **Frontend Memory Usage**: Memory usage with thresholds
10. **PostgreSQL CPU Usage**: Database CPU utilization
11. **PostgreSQL Memory Usage**: Database memory usage
12. **Service Availability**: Uptime percentage over last 6 hours

**Dashboard Best Practices:**
- All queries filter by `namespace="emzops-prod"`
- Visual thresholds show colored zones (green/yellow/red)
- Time series show historical data (6 hours default)
- Pod matching uses regex patterns (e.g., `pod=~"devops-backend-.*"`)

### Week 16: Checkpoint 4 - Final Integration

**Objective:** Consolidate all Phase 4 requirements and prepare for checkpoint demonstration.

**Key Achievements:**
- ✅ Complete Helm templating implementation
- ✅ Comprehensive secret management via CI/CD
- ✅ Stateless application architecture
- ✅ Production deployment with staging/production separation
- ✅ Automatic scaling (HPA) fully operational
- ✅ Affinity/topology constraints for high availability
- ✅ Security scanning integrated
- ✅ Production dashboard operational
- ✅ Complete documentation consolidation

**Final Architecture:**
- **Templating**: Helm charts for both backend and frontend
- **Secrets**: Managed via GitLab CI/CD variables, injected during deployment
- **Statelessness**: MinIO for files, PostgreSQL for data, JWT for sessions
- **Production**: Separate namespace with HPA, affinity, and monitoring
- **Security**: Automated container scanning with Trivy
- **Monitoring**: Comprehensive Grafana dashboard

---

## 3. Infrastructure & Networking

This section details the Kubernetes resources implemented across Weeks 9, 10, and 11.

### 3.1 Services Implementation

We rely on **Kubernetes Services** for stable networking.

*   **ClusterIP (Frontend/Backend)**: Default internal service type.
    *   Backend exposes port `8080` (App) and `9090` (Metrics).
    *   Frontend exposes port `80`.
*   **Headless Service (Database)**:
    *   Used for the StatefulSet to provide stable DNS entries (`postgres-0.postgres-service`).
    *   `clusterIP: None` allows direct pod addressability.

**Backend Service Configuration:**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: devops-backend-service
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9090"
    prometheus.io/path: "/metrics"
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
```

### 3.2 Ingress & TLS

We utilize the **Nginx Ingress Controller** for traffic routing and **cert-manager** for TLS automation.

*   **Host**: `devops-mezdit.kubernetes.devops.cs.ut.ee`
*   **TLS**: Automated via `letsencrypt-prod` cluster issuer.
*   **WebSocket Support**: Essential for the app's real-time features. Configured via `nginx.ingress.kubernetes.io/websocket-services`.

**Key Ingress Configuration:**
```yaml
metadata:
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/websocket-services: "devops-backend-service"
    nginx.ingress.kubernetes.io/rewrite-target: /$2
spec:
  rules:
  - host: devops-mezdit.kubernetes.devops.cs.ut.ee
    http:
      paths:
      - path: /api(/|$)(.*)
        pathType: ImplementationSpecific
        backend:
          service: { name: devops-backend-service, port: { number: 80 } }
      - path: /()(ws.*)  # WebSocket path preservation
        pathType: ImplementationSpecific
        backend:
          service: { name: devops-backend-service, port: { number: 80 } }
```

### 3.3 Secret Management & Security

Secure management of sensitive data (DB passwords, Registry credentials) is critical.

*   **Strategy**: Secrets are never committed to Git. We use template files (`secret.yaml.template`) and inject values via GitLab CI/CD masked variables.
*   **Image Pull Secrets**: `gitlab-registry-secret` is created using a GitLab Deploy Token (`read_registry` scope) to allow the cluster to pull images from our private registry.
*   **Database Secrets**: Helm templates dynamically generate `DATABASE_URL` using PostgreSQL service name if enabled.

**Secret Creation Process:**
1. Create GitLab Deploy Token with `read_registry` and `write_registry` scopes
2. Create Kubernetes secret in target namespace:
   ```bash
   kubectl create secret docker-registry gitlab-registry-secret \
     --docker-server=registry.gitlab.cs.ut.ee \
     --docker-username=gitlab+deploy-token-XXX \
     --docker-password=YOUR_TOKEN_HERE \
     --namespace=emzops-prod
   ```
3. Reference in Helm chart via `imagePullSecrets`

### 3.4 Network Policies

To secure the cluster, we implement **NetworkPolicies** to restrict traffic flow.

**Example Policy: `allow-prometheus-scraping`**
*   **Goal**: Allow Prometheus to scrape metrics, but block other unauthorized access.
*   **Rules**:
    *   Allow Ingress (port 8080) -> App.
    *   Allow Prometheus namespace (port 9090) -> App.
    *   Deny all other external traffic.

---

## 4. Templating & State Management (Helm)

To manage complexity and environment differences (Staging vs. Production), we use **Helm Charts**.

### 4.1 Helm Structure

We maintain two separate charts:
1.  `devops-be/chart/`: Manages Backend, Database (StatefulSet), and PVCs.
2.  `devops-fe/chart/`: Manages Frontend and Ingress.

**Why Helm?**
*   **Atomic Deployments**: Uses `--rollback-on-failure` to ensure rollbacks happen automatically if probes fail.
*   **Encapsulation**: Packages all related K8s manifests (Deployment, Service, ConfigMap) into a single unit.
*   **Value Substitution**: Allows easy injection of environment-specific values (replicas, resources, secrets).

**Chart Structure:**
```
chart/
├── Chart.yaml          # Chart metadata
├── values.yaml         # Default values (staging)
├── values-prod.yaml    # Production overrides
└── templates/
    ├── deployment.yaml
    ├── service.yaml
    ├── ingress.yaml
    ├── secret.yaml
    ├── configmap.yaml
    ├── hpa.yaml
    ├── postgres-statefulset.yaml
    └── postgres-service.yaml
```

### 4.2 State Persistence (StatefulSet)

The database component requires special handling for persistence.

*   Resource: **StatefulSet** (not Deployment)
*   Storage: **PersistentVolumeClaim (PVC)** requesting `1Gi` (staging) or `2Gi` (production) from `longhorn`.
*   Access Mode: `ReadWriteOnce`.
*   **Rationale**: `emptyDir` (suggested in earlier weeks) is ephemeral. We use PVCs to ensure data survives pod restarts and node rescheduling, critical for a production database.

**PostgreSQL StatefulSet Configuration:**
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: {{ .Release.Name }}-postgres
spec:
  serviceName: {{ .Release.Name }}-postgres-service
  replicas: 1
  template:
    spec:
      containers:
      - name: postgres
        image: postgres:15-alpine
        env:
        - name: POSTGRES_DB
          value: "{{ .Values.postgres.dbName }}"
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: {{ .Release.Name }}-secret
              key: POSTGRES_USER
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: longhorn
      resources:
        requests:
          storage: {{ .Values.postgres.storage.size }}
```

### 4.3 Dynamic Secret Generation

Helm templates dynamically generate `DATABASE_URL` based on configuration:

```yaml
# templates/secret.yaml
{{- if and .Values.postgres.enabled (not .Values.secrets.backend.databaseUrl) }}
# Auto-generate DATABASE_URL if PostgreSQL is enabled
DATABASE_URL: {{ printf "postgres://%s:%s@%s-postgres-service:5432/%s?sslmode=disable" 
  .Values.secrets.postgres.username 
  .Values.secrets.postgres.password 
  .Release.Name 
  .Values.postgres.dbName | b64enc | quote }}
{{- else if .Values.secrets.backend.databaseUrl }}
# Use explicit DATABASE_URL from CI/CD
DATABASE_URL: {{ .Values.secrets.backend.databaseUrl | b64enc | quote }}
{{- end }}
```

---

## 5. CI/CD & Container Security

Our fully automated pipeline handles building, testing, securing, and deploying the application.

### 5.1 GitLab CI/CD Pipeline Stages

Defined in `.gitlab-ci.yml`:

1.  **Security Scan (`security-scan`)**:
    *   **Tool**: Trivy (Aqua Security).
    *   **Action**: Scans the built container image for vulnerabilities (OS packages & dependencies).
    *   **Severity**: Reports CRITICAL and HIGH issues.
    *   **Integration**: Results are uploaded to GitLab's Security Dashboard.
    *   **Optimization**: Uses `--scanners vuln` to disable secret scanning (handled separately), `TRIVY_CACHE_DIR` for faster scans, pinned image version (0.51.2) for consistency.

2.  **Test (`test`)**:
    *   Runs Unit Tests and Integration Tests.
    *   Ensures code quality before building.

3.  **Build (`build-image`)**:
    *   **Tool**: Kaniko.
    *   **Context**: Builds Docker images inside the Kubernetes cluster securely (no Docker daemon required).
    *   **Caching**: Uses `--cache=true` with a 24h TTL to speed up builds.
    *   **Registry**: Pushes to GitLab Container Registry.
    *   **Tagging**: Creates both `latest` and commit SHA tags.

4.  **Deploy (`deploy`)**:
    *   **Staging**: Automated deployment to `emzops-staging`.
    *   **Production**: Manual trigger deployment to `emzops-prod`.

**Pipeline Flow:**
```
Code Push → Security Scan → Test → Build → Build Image → Deploy (Staging) → [Manual] Deploy (Production)
```

### 5.2 Container Security Scanning

**Trivy Configuration:**
```yaml
container-security-scan:
  stage: security-scan
  image:
    name: aquasec/trivy:0.51.2  # Pinned version
    entrypoint: [""]
  variables:
    TRIVY_NO_PROGRESS: "true"
    TRIVY_SKIP_DB_UPDATE: "false"
    TRIVY_FORMAT: "json"
    TRIVY_OUTPUT: "gl-container-scanning-report.json"
    TRIVY_EXIT_CODE: "0"
    TRIVY_IGNORE_UNFIXED: "false"
    TRIVY_SEVERITY: "CRITICAL,HIGH"
    TRIVY_CACHE_DIR: "/tmp/trivy-cache"
  cache:
    key: trivy-cache
    paths:
      - /tmp/trivy-cache
  script:
    - trivy image \
        --username "$CI_REGISTRY_USER" \
        --password "$CI_REGISTRY_PASSWORD" \
        --exit-code $TRIVY_EXIT_CODE \
        --severity $TRIVY_SEVERITY \
        --format $TRIVY_FORMAT \
        --scanners vuln \
        --cache-dir "$TRIVY_CACHE_DIR" \
        -o $TRIVY_OUTPUT \
        "$CI_REGISTRY/$CI_PROJECT_PATH:$IMAGE_TAG"
  artifacts:
    reports:
      container_scanning: gl-container-scanning-report.json
  dependencies:
    - build-docker-image
  timeout: 15 minutes
```

**Vulnerability Handling:**
- Current approach: `TRIVY_EXIT_CODE: "0"` - scans don't fail pipeline (allows review)
- Results visible in GitLab Security tab
- Team reviews and prioritizes vulnerabilities
- Future: Can be configured to fail on CRITICAL vulnerabilities

### 5.3 Deployment Strategy

**Staging Deployment:**
- Automatic on `main` branch push
- Uses default `values.yaml`
- Single replica, lower resources
- No autoscaling

**Production Deployment:**
- Manual trigger (requires approval)
- Uses `values-prod.yaml` overrides
- Multiple replicas (2-5 with HPA)
- Higher resources, autoscaling enabled
- Pod anti-affinity for high availability

**Helm Deployment Commands:**
```bash
# Staging
helm upgrade --install devops-backend ./chart \
  --namespace emzops-staging \
  --set image.tag=$IMAGE_TAG \
  --set secrets.backend.databaseUrl=$DATABASE_URL \
  --timeout 20m \
  --rollback-on-failure

# Production
helm upgrade --install devops-backend ./chart \
  --namespace emzops-prod \
  --values ./chart/values-prod.yaml \
  --set image.tag=$IMAGE_TAG \
  --timeout 20m \
  --rollback-on-failure
```

**Atomic Rollback:**
- `--rollback-on-failure` flag ensures automatic rollback if deployment fails
- Waits for pods to become ready before marking success
- If timeout occurs (e.g., ImagePullBackOff), rollback happens automatically
- Retry mechanism uses `--force-replace` without rollback for conflict resolution

---

## 6. Observability, Monitoring & Alerts

We implemented a full-stack observability solution using **Prometheus** and **Grafana**.

### 6.1 Architecture

*   **Metrics Exposition**: The Golang backend exposes a `/metrics` endpoint on port `9090` using the official Prometheus client library.
*   **Service Discovery**: Prometheus automatically discovers targets via Service annotations (`prometheus.io/scrape: "true"`).
*   **Scraping**: Metrics are scraped every 30 seconds.
*   **NetworkPolicy**: Only Prometheus namespace can access metrics endpoint (port 9090).

### 6.2 Key Metrics Implementation

Detailed breakdown of custom metrics implemented in `internal/handlers/metrics.go`:

| Metric Name | Type | Description | PromQL Example |
|-------------|------|-------------|----------------|
| `app_requests_total` | Counter | Total HTTP requests served. | `rate(app_requests_total[5m])` (RPS) |
| `app_errors_total` | Counter | Total 4xx/5xx responses. | `rate(app_errors_total[5m])` (Error Rate) |
| `app_request_latency_microseconds` | Counter | Cumulative latency. | `rate(latency)[5m] / rate(requests)[5m]` (Avg Latency) |
| `app_posts_total` | Gauge | Current number of posts in DB. | `app_posts_total` |
| `app_users_total` | Gauge | Current number of users in DB. | `app_users_total` |

### 6.3 Health Probes (Liveness vs. Readiness)

Properly configured probes are essential for Kubernetes reliability.

1.  **Liveness Probe (`/healthz`)**:
    *   **Check**: "Is the process running?"
    *   **Action**: If fails 3 times, **Restart Pod**.
    *   *Code Logic*: Always returns 200 OK if the HTTP server is up.

2.  **Readiness Probe (`/readyz`)**:
    *   **Check**: "Can I serve traffic?" (e.g., Is DB connected?)
    *   **Action**: If fails, **Remove Pod from Service Endpoints** (stop sending traffic).
    *   *Code Logic*: Pings the PostgreSQL database. Returns 503 if DB is unreachable.

**Probe Configuration:**
```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /readyz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
  failureThreshold: 3
```

### 6.4 Production Dashboard (Grafana)

We created a comprehensive dashboard (`emzops-prod` namespace) monitoring:

1.  **Golden Signals**: Latency, Traffic, Errors, Saturation.
2.  **Resource Usage**: CPU/Memory usage with visual thresholds (Red/Yellow zones) to detect leaks or limits.
3.  **HPA Status**: Visualizes current vs. desired replicas to verify autoscaling behavior.
4.  **Pod Lifecycle**: Tracks restarts and pod phases.
5.  **Service Availability**: Monitors uptime over time windows.

**Dashboard Panels:**

**Application Health:**
- Backend Pods Ready (Stat)
- Frontend Pods Ready (Stat)
- PostgreSQL Pods Ready (Stat)

**Pod Status:**
- Pod Status by Phase (Pie Chart)
- Pod Restarts Over Time (Time Series)

**Autoscaling:**
- HPA Current Replicas (Time Series)
- HPA Desired Replicas (Time Series)

**Resource Usage:**
- Backend CPU Usage (Time Series with thresholds: 70% yellow, 90% red)
- Backend Memory Usage (Time Series with thresholds: 400MB yellow, 480MB red)
- Frontend CPU/Memory Usage
- PostgreSQL CPU/Memory Usage

**Availability:**
- Service Availability (Last 6 Hours) - Uptime percentage

**Access:** https://monitoring.kubernetes.devops.cs.ut.ee → Team folder: `emzops`

---

## 7. Production Readiness & Autoscaling

Production (`emzops-prod`) differs from Staging to ensure high availability and resilience.

### 7.1 Horizontal Pod Autoscaling (HPA)

We implemented automated scaling to handle traffic spikes.

*   **Config**: Defined in `chart/templates/hpa.yaml`.
*   **Metric**: `cpu` utilization.
*   **Target**: 30% average utilization (set strictly low for demonstration purposes).
*   **Behavior**:
    1.  If CPU spikes > 30%, HPA adds replicas (up to 5).
    2.  If CPU drops, HPA removes replicas (down to 2).

**HPA Configuration:**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ .Release.Name }}-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ .Release.Name }}
  minReplicas: {{ .Values.autoscaling.minReplicas }}
  maxReplicas: {{ .Values.autoscaling.maxReplicas }}
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: {{ .Values.autoscaling.targetCPU }}
```

**Verification Command:**
```bash
watch kubectl get hpa -n emzops-prod
# Triggers: kubectl run stress --image=containerstack/cpustress -- -cpu 2
```

### 7.2 High Availability (Affinity Rules)

To prevent a single node failure from taking down the entire application, we use **Pod Anti-Affinity**.

*   **Rule**: `requiredDuringSchedulingIgnoredDuringExecution`.
*   **Logic**: "Do not schedule a Backend pod on a node that already has a Backend pod."
*   **Result**: Pods are spread across available nodes (e.g., `do-worker8`, `do-worker9`).

**Affinity Configuration:**
```yaml
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchExpressions:
        - key: app
          operator: In
          values:
          - {{ .Release.Name }}
      topologyKey: kubernetes.io/hostname
```

### 7.3 Chaos Engineering Resilience

The system is architected to withstand failures:

1.  **Pod Failure**: If a pod crashes, the ReplicaSet immediately schedules a replacement.
2.  **Node Failure**: If a node goes down (Simulated via `kubectl drain`), Kubernetes reschedules pods to healthy nodes.
3.  **Zero Downtime**: Rolling updates ensure that old pods are only terminated after new pods are `Running` and `Readiness Probe` passes.

**Chaos Testing:**
```bash
# Pod kill test
kubectl delete pod -n emzops-prod -l app=devops-backend

# Verify: New pod starts automatically, no downtime
kubectl get pods -n emzops-prod -l app=devops-backend -w

# Node failure simulation
kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data

# Verify: Pods rescheduled to other nodes
kubectl get pods -n emzops-prod -o wide
```

### 7.4 Statelessness

The application is designed to be stateless:

1. **Application State**: No session storage in application code
2. **File Storage**: MinIO (S3-compatible) for uploaded files
3. **Database**: PostgreSQL for persistent data
4. **Authentication**: JWT tokens (stateless)
5. **Session Management**: No server-side sessions

**Benefits:**
- Pods can be killed and restarted without data loss
- Horizontal scaling works seamlessly
- No sticky sessions required
- Easy to scale horizontally

---

## 8. Comprehensive Testing & Troubleshooting

### 8.1 Deployment Checklist

Use this checklist when deploying to Production.

- [ ] **Secrets**: Ensure `gitlab-registry-secret` exists in both `emzops-staging` and `emzops-prod`.
- [ ] **Namespace**: Ensure `emzops-prod` exists.
- [ ] **SSL**: Check `kubectl get certificates` to ensure Let's Encrypt cert is `Ready`.
- [ ] **Database**: Verify PostgreSQL StatefulSet is running and PVC is bound.
- [ ] **HPA**: Check HPA is created and showing metrics.
- [ ] **Monitoring**: Verify Prometheus is scraping metrics and Grafana dashboard shows data.

### 8.2 Testing Guide

#### 8.2.1 Connectivity Test
```bash
curl -I https://devops-mezdit.kubernetes.devops.cs.ut.ee
# Expected: HTTP 200 OK
```

#### 8.2.2 WebSocket Test
Connect a WebSocket client to `wss://devops-mezdit.kubernetes.devops.cs.ut.ee/ws`.
*   **Success**: Connection established, pings received.
*   **Failure**: Check Ingress WebSocket annotations.

#### 8.2.3 Load Scaling Test
1.  Watch HPA: `watch kubectl get hpa -n emzops-prod`
2.  Run Load: `kubectl run load-test --image=busybox /bin/sh -c "while true; do wget -q -O- https://...; done"`
3.  Observation: Replicas should increase from 2 -> 3 -> 4.

#### 8.2.4 Health Endpoints Test
```bash
# Port forward to backend
kubectl port-forward service/devops-backend-service -n emzops-prod 8080:80

# Test healthz
curl http://localhost:8080/healthz
# Expected: ok (200 OK)

# Test readyz
curl http://localhost:8080/readyz
# Expected: ok (200 OK) if DB is connected
```

#### 8.2.5 Metrics Endpoint Test
```bash
# Port forward to metrics port
kubectl port-forward service/devops-backend-service -n emzops-prod 9090:9090

# Fetch metrics
curl http://localhost:9090/metrics
# Expected: Prometheus format metrics
```

### 8.3 Troubleshooting Scenarios

| Issue | Probable Cause | Fix |
|-------|----------------|-----|
| **ImagePullBackOff** | Registry credentials expired or missing secret. | Re-create `gitlab-registry-secret` with valid Deploy Token. |
| **503 Service Unavailable** | Readiness probes failing (DB down?). | Check `kubectl logs -l app=postgres` and ensure DB is running. |
| **Certificate Error** | Let's Encrypt quota or DNS issue. | Check `kubectl describe challenge` and `cert-manager` logs. |
| **Pods Pending** | Insufficient CPU/RAM on nodes or Affinity conflict. | Check `kubectl describe pod <pod>` for events. |
| **HPA Not Scaling** | Metrics-server not installed or resource requests missing. | Check `kubectl get apiservice v1beta1.metrics.k8s.io` and verify resource requests in deployment. |
| **Readiness Probe Failing** | Database connection issue. | Check DATABASE_URL in secret, verify postgres-service is accessible. |
| **Prometheus Not Scraping** | NetworkPolicy blocking or missing annotations. | Check NetworkPolicy allows `prometheus-stack` namespace, verify service annotations. |

### 8.4 Common Commands Reference

```bash
# Pods
kubectl get pods -n emzops-prod
kubectl describe pod <pod-name> -n emzops-prod
kubectl logs -f <pod-name> -n emzops-prod

# Deployments
kubectl get deployments -n emzops-prod
kubectl rollout status deployment/devops-backend -n emzops-prod
kubectl rollout restart deployment/devops-backend -n emzops-prod

# Services
kubectl get svc -n emzops-prod
kubectl describe svc devops-backend-service -n emzops-prod

# HPA
kubectl get hpa -n emzops-prod
kubectl describe hpa devops-backend-hpa -n emzops-prod

# Secrets
kubectl get secrets -n emzops-prod
kubectl get secret gitlab-registry-secret -n emzops-prod -o yaml

# Storage
kubectl get pvc -n emzops-prod
kubectl describe pvc postgres-storage-postgres-0 -n emzops-prod

# Events
kubectl get events -n emzops-prod --sort-by='.lastTimestamp'

# Port Forwarding
kubectl port-forward service/devops-backend-service -n emzops-prod 8080:80
kubectl port-forward service/devops-backend-service -n emzops-prod 9090:9090
```

### 8.5 Production Secret Recreation

If `ImagePullBackOff` occurs due to expired registry credentials:

1. **Create new Deploy Token in GitLab:**
   - Go to Project → Settings → Repository → Deploy Tokens
   - Create token with `read_registry` and `write_registry` scopes
   - Save username and token

2. **Recreate Kubernetes Secret:**
   ```bash
   kubectl delete secret gitlab-registry-secret -n emzops-prod
   
   kubectl create secret docker-registry gitlab-registry-secret \
     --docker-server=registry.gitlab.cs.ut.ee \
     --docker-username=gitlab+deploy-token-XXX \
     --docker-password=YOUR_TOKEN_HERE \
     --namespace=emzops-prod
   ```

3. **Restart Pods:**
   ```bash
   kubectl delete pods -n emzops-prod -l app=devops-backend
   kubectl delete pods -n emzops-prod -l app=devops-frontend
   ```

4. **Verify:**
   ```bash
   kubectl get pods -n emzops-prod
   # Pods should be Running and Ready
   ```

---

## Summary

### Project Evolution Summary

| Week | Focus | Key Achievements |
|------|-------|------------------|
| **Week 9** | Initial Kubernetes Deployment | Container images, private registry, basic deployments |
| **Week 10** | Production Networking | Services, Ingress, TLS, ConfigMaps, Secrets, RBAC |
| **Week 11** | Observability | Health probes, Prometheus metrics, Grafana, NetworkPolicy, PVCs |
| **Week 13** | Templating | Helm charts, atomic deployments, statelessness |
| **Week 14** | Production & Autoscaling | Production namespace, HPA, affinity rules, chaos engineering |
| **Week 15** | Security & Monitoring | Container scanning, production dashboard |
| **Week 16** | Checkpoint 4 | Complete integration, comprehensive documentation |

### Key Achievements

- ✅ **Production-grade Kubernetes deployment** with Helm templating
- ✅ **Automated CI/CD pipeline** with security scanning
- ✅ **Full-stack observability** with Prometheus and Grafana
- ✅ **Automatic scaling** with HPA (2-5 replicas)
- ✅ **High availability** with pod anti-affinity
- ✅ **Container security** with automated Trivy scanning
- ✅ **Zero-downtime deployments** with atomic rollback
- ✅ **Stateless architecture** with external storage (MinIO, PostgreSQL)
- ✅ **Comprehensive monitoring** with 12+ Grafana panels
- ✅ **Complete documentation** covering all weeks

### Production Readiness Checklist

- ✅ Separate staging and production environments
- ✅ Automated deployments with rollback capability
- ✅ Health checks (liveness and readiness probes)
- ✅ Resource limits and requests configured
- ✅ Horizontal Pod Autoscaling enabled
- ✅ Pod anti-affinity for high availability
- ✅ Persistent storage for database
- ✅ TLS/HTTPS with automatic certificate management
- ✅ Container security scanning
- ✅ Comprehensive monitoring and alerting
- ✅ Stateless application design
- ✅ Complete documentation

---

> **End of Master Documentation**  
> *Consolidated from Weeks 9-16 for the DevOps Project Team emzops.*  
> *Last Updated: December 2025*

---

## AI Usage Declaration

This project was developed by human developers with AI tools used as coding assistants to improve productivity:

### How AI Was Used
- **Code Completion**: For boilerplate code and repetitive patterns
- **Problem Solving**: For debugging complex issues and exploring solutions
- **Documentation**: Assistance with writing clear README sections and code comments

### What Humans Did
- 🧑‍💻 Designed the entire system architecture and database schema
- 🧑‍💻 Implemented core business logic, authentication, and WebSocket features
- 🧑‍💻 Made all technical decisions regarding frameworks, libraries, and patterns
- 🧑‍💻 Wrote and debugged the majority of the codebase
- 🧑‍💻 Set up CI/CD pipeline, deployment infrastructure, and monitoring
- 🧑‍💻 Performed code reviews, testing, and quality assurance
- 🧑‍💻 Troubleshot production issues and optimized performance

**In Summary**: AI tools served as helpful assistants for routine tasks and suggestions, but all critical thinking, design decisions, and implementation work were done by the development team.
