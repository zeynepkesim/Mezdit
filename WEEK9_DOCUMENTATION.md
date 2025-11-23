# Week 9 - Kubernetes Deployment Documentation

**Team:** emzops  
**Namespace:** emzops-staging  
**Application URL:** https://devops-mezdit.kubernetes.devops.cs.ut.ee

---

## Table of Contents

1. [Overview](#overview)
2. [Container Image Building](#container-image-building)
3. [Image Tagging and Versioning Strategy](#image-tagging-and-versioning-strategy)
4. [GitLab Container Registry Push Process](#gitlab-container-registry-push-process)
5. [Kubernetes Deployment Structure](#kubernetes-deployment-structure)
6. [Private Registry Authentication](#private-registry-authentication)
7. [Secret Management](#secret-management)
8. [Testing and Verification](#testing-and-verification)

---

## Overview

This document describes the Week 9 Kubernetes deployment implementation for our DevOps project. We have successfully migrated our application from VM-based deployment to Kubernetes cluster deployment.

### Architecture Summary

```
GitLab Pipeline → Build Container Images → GitLab Registry → Kubernetes Cluster
                                                              ↓
                                          namespace: emzops-staging
                                          ├── Backend (Go application)
                                          ├── Frontend (Vue.js + Nginx)
                                          └── PostgreSQL (StatefulSet)
```

### Key Components

- **Backend**: Go 1.23 application running on port 8080
- **Frontend**: Vue.js application served by Nginx on port 80
- **Database**: PostgreSQL 15 (StatefulSet with 1Gi persistent volume)
- **Ingress**: Automatic TLS termination with cert-manager
- **Replicas**: 1 replica for each deployment (Phase 3 requirement)

---

## Container Image Building

### Backend Image Build

**Technology:** Kaniko (rootless container builder)  
**Base Image:** `golang:1.23-alpine`  
**Build Location:** GitLab CI/CD Pipeline

#### Dockerfile Structure (Backend)

```dockerfile
# Single stage build
FROM golang:1.23-alpine
WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy source code
COPY . .

# Ensure Go 1.23 compatibility
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

EXPOSE 8080
CMD ["./main"]
```

**Key Features:**
- ✅ Alpine Linux for minimal image size
- ✅ Static binary compilation (`CGO_ENABLED=0`)
- ✅ CA certificates included for secure connections
- ✅ Go mod tidy ensures dependency compatibility

### Frontend Image Build

**Technology:** Kaniko (rootless container builder)  
**Base Image:** `node:22-alpine` (build), `nginx:alpine` (production)  
**Build Location:** GitLab CI/CD Pipeline

#### Dockerfile Structure (Frontend)

```dockerfile
# Multi-stage build
# Stage 1: Build Vue.js application
FROM node:22-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Stage 2: Production nginx server
FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

**Key Features:**
- ✅ Multi-stage build (reduced final image size)
- ✅ `npm ci` for reproducible builds
- ✅ Custom nginx configuration for SPA routing
- ✅ Alpine Linux for minimal footprint

### GitLab CI/CD Pipeline Configuration

Both backend and frontend use similar CI/CD pipeline structure:

#### Pipeline Stages

```yaml
stages:
  - security      # Secret scanning with gitleaks
  - test          # Unit tests (backend only)
  - build         # Compile application
  - build-image   # Build container image
  - deploy        # Deploy to Kubernetes
```

#### Build Image Stage (Example - Backend)

```yaml
build-docker-image:
  stage: build-image
  image:
    name: gcr.io/kaniko-project/executor:v1.23.0-debug
    entrypoint: [""]
  script:
    - echo "Building backend container image with Kaniko..."
    - mkdir -p /kaniko/.docker
    # Create Docker config with GitLab credentials
    - echo "{\"auths\":{\"$CI_REGISTRY\":{\"auth\":\"$(printf "%s:%s" "${CI_REGISTRY_USER}" "${CI_REGISTRY_PASSWORD}" | base64 | tr -d '\n')\"}}}" > /kaniko/.docker/config.json
    # Build and push with Kaniko
    - >-
      /kaniko/executor
      --context "${CI_PROJECT_DIR}"
      --dockerfile "${CI_PROJECT_DIR}/Dockerfile"
      --destination "${IMAGE_NAME}:${IMAGE_TAG}"
      --destination "${IMAGE_NAME}:latest"
      --cache=true
      --cache-ttl=24h
    - echo "✅ Backend image pushed - $IMAGE_NAME:latest"
  only:
    - main
```

**Why Kaniko?**
- ✅ Builds containers without requiring Docker daemon
- ✅ Works in unprivileged CI/CD environments
- ✅ Supports layer caching for faster builds
- ✅ Directly pushes to container registry

---

## Image Tagging and Versioning Strategy

### Tagging Strategy

We use a **dual-tag approach** for container images:

1. **`latest` tag**: Always points to the most recent build from `main` branch
2. **Git commit SHA tag**: Specific version for rollback capability

#### Implementation

```yaml
variables:
  IMAGE_NAME: $CI_REGISTRY_IMAGE
  IMAGE_TAG: $CI_COMMIT_SHORT_SHA
```

**Example image tags:**
- `registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-be:latest`
- `registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-be:a1b2c3d`

### Benefits

| Tag | Purpose | Use Case |
|-----|---------|----------|
| `latest` | Always current | Kubernetes deployments, testing |
| `<commit-sha>` | Specific version | Rollback, debugging, audit trail |

### Version Control

- ✅ Every `main` branch commit creates a new tagged image
- ✅ Commit SHA provides immutable reference
- ✅ GitLab Container Registry stores all historical versions
- ✅ Easy rollback using `kubectl set image` with specific tag

---

## GitLab Container Registry Push Process

### Registry Details

**Backend Registry:**  
`registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-be`

**Frontend Registry:**  
`registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-fe`

### Authentication Flow

```
GitLab CI/CD Pipeline
  ↓
Uses built-in variables:
  - CI_REGISTRY (registry.gitlab.cs.ut.ee)
  - CI_REGISTRY_USER (gitlab-ci-token)
  - CI_REGISTRY_PASSWORD (automatically provided)
  ↓
Kaniko authenticates to registry
  ↓
Builds image
  ↓
Pushes to GitLab Container Registry
```

### Push Process Steps

1. **Authentication**
   ```bash
   # Kaniko creates Docker config with credentials
   echo "{\"auths\":{\"$CI_REGISTRY\":{...}}}" > /kaniko/.docker/config.json
   ```

2. **Build**
   ```bash
   /kaniko/executor --context . --dockerfile ./Dockerfile
   ```

3. **Push** (automatic with Kaniko)
   ```bash
   --destination "${IMAGE_NAME}:${IMAGE_TAG}"
   --destination "${IMAGE_NAME}:latest"
   ```

4. **Verification**
   ```bash
   # Check in GitLab UI: Project > Deploy > Container Registry
   ```

### Caching Strategy

```yaml
--cache=true
--cache-ttl=24h
```

- ✅ Reuses layers from previous builds
- ✅ Reduces build time significantly
- ✅ 24-hour cache TTL for fresh daily builds

---

## Kubernetes Deployment Structure

### Namespace Resources

Our deployment in `emzops-staging` namespace consists of:

```
emzops-staging/
├── ConfigMaps
│   ├── backend-config (PORT=8080)
│   └── frontend-config (VUE_APP_API_URL=/api)
├── Secrets
│   ├── gitlab-registry-secret (backend image pull)
│   ├── frontend-registry-secret (frontend image pull)
│   ├── backend-secret (DATABASE_URL)
│   └── postgres-superuser-secret (DB credentials)
├── Deployments
│   ├── devops-backend (1 replica)
│   └── devops-frontend (1 replica)
├── StatefulSets
│   └── postgres (1 replica, 1Gi PVC)
├── Services
│   ├── devops-backend-service (ClusterIP:80)
│   ├── devops-frontend-service (ClusterIP:80)
│   └── postgres-service (Headless)
└── Ingress
    └── devops-backend-ingress (TLS enabled)
```

### Backend Deployment Manifest

**File:** `devops-be/k8s/deployment.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: devops-backend
  namespace: emzops-staging
spec:
  replicas: 1  # Phase 3 requirement
  selector:
    matchLabels:
      app: devops-backend
  template:
    metadata:
      labels:
        app: devops-backend
    spec:
      # Secret for pulling images from private registry
      imagePullSecrets:
        - name: gitlab-registry-secret
      containers:
      - name: backend
        # Image from GitLab Container Registry
        image: registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-be:latest
        ports:
        - containerPort: 8080
        # Environment variables from ConfigMap
        env:
        - name: PORT
          valueFrom:
            configMapKeyRef:
              name: backend-config
              key: PORT
        # Sensitive data from Secret
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: backend-secret
              key: DATABASE_URL
```

### Frontend Deployment Manifest

**File:** `devops-fe/k8s/deployment.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: devops-frontend
  namespace: emzops-staging
spec:
  replicas: 1
  selector:
    matchLabels:
      app: devops-frontend
  template:
    metadata:
      labels:
        app: devops-frontend
    spec:
      imagePullSecrets:
        - name: frontend-registry-secret
      containers:
      - name: frontend
        image: registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-fe:latest
        ports:
        - containerPort: 80
        # All environment variables from ConfigMap
        envFrom:
        - configMapRef:
            name: frontend-config
        # Resource limits
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        # Health checks
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Service Manifests

**Backend Service:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: devops-backend-service
  namespace: emzops-staging
spec:
  type: ClusterIP
  selector:
    app: devops-backend
  ports:
  - protocol: TCP
    port: 80        # Service port
    targetPort: 8080 # Container port
```

**Frontend Service:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: devops-frontend-service
  namespace: emzops-staging
spec:
  type: ClusterIP
  selector:
    app: devops-frontend
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
```

---

## Private Registry Authentication

### Challenge

Kubernetes needs credentials to pull images from our **private** GitLab Container Registry.

### Solution: Docker Registry Secret

We use Kubernetes `docker-registry` type secrets to provide authentication.

### Implementation Steps

#### 1. Create GitLab Deploy Token

**Location:** GitLab Project → Settings → Repository → Deploy Tokens

**Configuration:**
- Name: `kubernetes-deploy`
- Scopes: `read_registry` ✅
- Expiration: (optional)

**Generated credentials:**
- Username: (auto-generated)
- Token: (auto-generated, save this!)

#### 2. Create Kubernetes Secret

**Backend secret creation:**

```bash
kubectl create secret docker-registry gitlab-registry-secret \
  --docker-server=registry.gitlab.cs.ut.ee \
  --docker-username=<deploy-token-username> \
  --docker-password=<deploy-token-password> \
  --namespace=emzops-staging
```

**Frontend secret creation:**

```bash
kubectl create secret docker-registry frontend-registry-secret \
  --docker-server=registry.gitlab.cs.ut.ee \
  --docker-username=<deploy-token-username> \
  --docker-password=<deploy-token-password> \
  --namespace=emzops-staging
```

#### 3. Reference Secret in Deployment

```yaml
spec:
  template:
    spec:
      imagePullSecrets:
        - name: gitlab-registry-secret  # Reference to secret
      containers:
      - name: backend
        image: registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-be:latest
```

### Verification

```bash
# Check if secret exists
kubectl get secret gitlab-registry-secret -n emzops-staging

# Describe secret (credentials are base64 encoded)
kubectl describe secret gitlab-registry-secret -n emzops-staging
```

### How It Works

```
Kubernetes Deployment
  ↓
Reads imagePullSecrets
  ↓
Retrieves credentials from Secret
  ↓
Authenticates to registry.gitlab.cs.ut.ee
  ↓
Pulls container image
  ↓
Starts pod
```

---

## Secret Management

### Types of Secrets Used

| Secret Name | Type | Purpose | Contains |
|-------------|------|---------|----------|
| `gitlab-registry-secret` | `docker-registry` | Image pull auth | Deploy token credentials |
| `frontend-registry-secret` | `docker-registry` | Image pull auth | Deploy token credentials |
| `backend-secret` | `Opaque` | App config | `DATABASE_URL` |
| `postgres-superuser-secret` | `basic-auth` | DB auth | PostgreSQL credentials |

### Backend Secret Example

**Template:** `devops-be/k8s/secret.yaml.template`

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: backend-secret
  namespace: emzops-staging
type: Opaque
stringData:
  DATABASE_URL: "postgresql://devops:PASSWORD@postgres-service:5432/devops"
```

**Note:** Actual `secret.yaml` files are **NOT committed to Git** (in `.gitignore`)

### PostgreSQL Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: postgres-superuser-secret
  namespace: emzops-staging
type: kubernetes.io/basic-auth
stringData:
  username: devops
  password: <secure-password>
```

### Secret Creation Workflow

1. **Local creation:**
   ```bash
   # Copy template
   cp secret.yaml.template secret.yaml
   
   # Edit with real values
   vim secret.yaml
   
   # Apply to cluster
   kubectl apply -f secret.yaml -n emzops-staging
   ```

2. **GitLab CI/CD variables:**
   - Secrets are stored as GitLab CI/CD variables
   - Referenced in pipeline: `$KUBECONFIG_CONTENT`, `$DATABASE_PASSWORD`

3. **Security best practices:**
   - ✅ Never commit secrets to Git
   - ✅ Use `.gitignore` for `secret.yaml` files
   - ✅ Use GitLab protected variables
   - ✅ Regularly rotate credentials

### Using Secrets in Pods

**As environment variable:**

```yaml
env:
- name: DATABASE_URL
  valueFrom:
    secretKeyRef:
      name: backend-secret
      key: DATABASE_URL
```

**As volume mount (not used in our setup):**

```yaml
volumes:
- name: secret-volume
  secret:
    secretName: backend-secret
volumeMounts:
- name: secret-volume
  mountPath: /etc/secrets
  readOnly: true
```

---

## Testing and Verification

### Port Forwarding Tests (Week 9 Requirement)

Week 9 specifically requires testing pods using **port forwarding**.

#### Backend Test

```bash
# Start port forwarding
kubectl port-forward deployment/devops-backend -n emzops-staging 8080:8080

# Test in another terminal
curl -I http://localhost:8080

# Expected: Connection successful (may get 404 if no root endpoint)
```

**Result:** ✅ Backend successfully responds via port forwarding

#### Frontend Test

```bash
# Start port forwarding
kubectl port-forward deployment/devops-frontend -n emzops-staging 8081:80

# Test in browser or curl
curl -I http://localhost:8081

# Expected: HTTP 200 OK with nginx headers
```

**Result:** ✅ Frontend serves HTML successfully

#### PostgreSQL Test

```bash
# Start port forwarding
kubectl port-forward statefulset/postgres -n emzops-staging 5433:5432

# Test connection
nc -zv localhost 5433

# Expected: Connection succeeded
```

**Result:** ✅ PostgreSQL accepts connections

### Deployment Verification Commands

```bash
# Check all resources
kubectl get all -n emzops-staging

# Check pod status
kubectl get pods -n emzops-staging

# View pod logs
kubectl logs -f deployment/devops-backend -n emzops-staging

# Describe deployment
kubectl describe deployment devops-backend -n emzops-staging

# Check events
kubectl get events -n emzops-staging --sort-by='.lastTimestamp'
```

### Application Health Checks

**Current status (November 23, 2025):**

```bash
$ kubectl get pods -n emzops-staging
NAME                                READY   STATUS    RESTARTS   AGE
devops-backend-5ddbf4c5c7-x8q9b     1/1     Running   0          10d
devops-frontend-757689c47c-ggbbs    1/1     Running   0          31h
postgres-0                          1/1     Running   0          10d
```

✅ All pods running successfully  
✅ No restarts (indicates stability)  
✅ Backend running for 10 days  
✅ Frontend updated 31 hours ago

### Public Access Test

**Application URL:** https://devops-mezdit.kubernetes.devops.cs.ut.ee

**Ingress configuration:**
- ✅ Automatic TLS with cert-manager
- ✅ HTTP to HTTPS redirect
- ✅ Routes to backend and frontend services

**Test:**
```bash
curl -I https://devops-mezdit.kubernetes.devops.cs.ut.ee

# Expected: HTTP 200 with valid TLS certificate
```

---

## Summary

### Week 9 Requirements Completion

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Build container images in CI/CD | ✅ Complete | Kaniko builds in `.gitlab-ci.yml` |
| Push to GitLab Container Registry | ✅ Complete | Both repos push successfully |
| Access Kubernetes cluster | ✅ Complete | Using ETAIS auth with kubelogin |
| Deploy to namespace | ✅ Complete | `emzops-staging` namespace active |
| Use secrets for private registry | ✅ Complete | Deploy tokens and image pull secrets |
| Test with port forwarding | ✅ Complete | All services tested successfully |
| Documentation | ✅ Complete | This document |
| Phase out VM deployment | ⚠️ In Progress | Kubernetes primary, VM still running |

### Next Steps (Week 10)

Based on the course structure, Week 10 will likely cover:
- 🔄 Services and Ingress configuration (partially done)
- 🔄 Persistent volumes (already using for PostgreSQL)
- 🔄 ConfigMaps and Secrets management (already implemented)
- 🔄 Proper networking setup

### Key Achievements

- ✅ **Zero-downtime deployment** through Kubernetes
- ✅ **Automated CI/CD pipeline** with security scanning
- ✅ **Secure credential management** with Kubernetes secrets
- ✅ **Persistent data** with StatefulSet for PostgreSQL
- ✅ **Production-ready setup** with TLS and ingress
- ✅ **Reproducible builds** with version tagging

---

## Appendix

### Useful Commands Reference

```bash
# Context management (less typing)
kubectl config set-context --current --namespace=emzops-staging

# Watch pod status
kubectl get pods -w -n emzops-staging

# Get pod logs
kubectl logs -f <pod-name> -n emzops-staging

# Execute command in pod
kubectl exec -it <pod-name> -n emzops-staging -- /bin/sh

# Port forward (background)
kubectl port-forward deployment/devops-backend -n emzops-staging 8080:8080 &

# Stop all port forwards
pkill -f "kubectl port-forward"

# Delete and recreate deployment
kubectl delete deployment devops-backend -n emzops-staging
kubectl apply -f k8s/deployment.yaml

# Force update with new image
kubectl rollout restart deployment/devops-backend -n emzops-staging

# Check rollout status
kubectl rollout status deployment/devops-backend -n emzops-staging
```

### Troubleshooting Guide

**Problem:** ImagePullBackOff error

**Solution:**
```bash
# Check if secret exists
kubectl get secret gitlab-registry-secret -n emzops-staging

# Verify deploy token is still valid in GitLab

# Recreate secret if needed
kubectl delete secret gitlab-registry-secret -n emzops-staging
kubectl create secret docker-registry gitlab-registry-secret \
  --docker-server=registry.gitlab.cs.ut.ee \
  --docker-username=<username> \
  --docker-password=<token> \
  --namespace=emzops-staging
```

**Problem:** CrashLoopBackOff

**Solution:**
```bash
# Check logs
kubectl logs <pod-name> -n emzops-staging

# Common issues:
# - Missing environment variables (check ConfigMap/Secret)
# - Database connection failure (check postgres-service)
# - Application error (check application code)
```

---

> **Note:** This document was prepared with the assistance of an AI (artificial intelligence) tool.