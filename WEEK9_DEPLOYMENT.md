# Week 9 - Kubernetes Deployment Guide (Backend)

**Project:** DevOps Backend  
**Team:** emzops  
**Namespace:** emzops-staging  
**Registry:** `registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-be`

---

## Quick Reference

### Container Image

**Build:** Kaniko in GitLab CI/CD pipeline  
**Base Image:** `golang:1.23-alpine`  
**Tags:**
- `latest` - Most recent build
- `<commit-sha>` - Specific version

### Deployment

**Replicas:** 1 (Phase 3 requirement)  
**Port:** 8080  
**Service:** `devops-backend-service` (ClusterIP)

### Authentication

**Image Pull Secret:** `gitlab-registry-secret`  
**Type:** GitLab Deploy Token with `read_registry` scope

---

## How Container Images Are Built

### 1. GitLab CI/CD Pipeline

The `.gitlab-ci.yml` file defines our build process:

```yaml
build-docker-image:
  stage: build-image
  image:
    name: gcr.io/kaniko-project/executor:v1.23.0-debug
    entrypoint: [""]
  script:
    - echo "Building backend container image with Kaniko..."
    - mkdir -p /kaniko/.docker
    - echo "{\"auths\":{\"$CI_REGISTRY\":{\"auth\":\"$(printf "%s:%s" "${CI_REGISTRY_USER}" "${CI_REGISTRY_PASSWORD}" | base64 | tr -d '\n')\"}}}" > /kaniko/.docker/config.json
    - >-
      /kaniko/executor
      --context "${CI_PROJECT_DIR}"
      --dockerfile "${CI_PROJECT_DIR}/Dockerfile"
      --destination "${IMAGE_NAME}:${IMAGE_TAG}"
      --destination "${IMAGE_NAME}:latest"
      --cache=true
      --cache-ttl=24h
```

### 2. Build Process Flow

```
Code Push to main
  ↓
GitLab CI/CD Triggered
  ↓
Kaniko Executor Starts
  ↓
Reads Dockerfile
  ↓
Builds Image Layers
  ↓
Tags Image (latest + commit SHA)
  ↓
Pushes to GitLab Container Registry
  ↓
Deployment Stage Updates Kubernetes
```

### 3. Why Kaniko?

- ✅ No Docker daemon required
- ✅ Works in unprivileged CI/CD runners
- ✅ Layer caching for faster builds
- ✅ Secure, rootless container builds

---

## Image Tagging and Versioning Strategy

### Dual-Tag Approach

**1. Latest Tag:**
```
registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-be:latest
```
- Always points to most recent build
- Used by Kubernetes deployments
- Automatically updated on every `main` branch build

**2. Commit SHA Tag:**
```
registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-be:a1b2c3d
```
- Immutable reference to specific build
- Enables rollback to exact version
- Audit trail for deployed versions

### Implementation

```yaml
variables:
  IMAGE_NAME: $CI_REGISTRY_IMAGE
  IMAGE_TAG: $CI_COMMIT_SHORT_SHA
```

---

## How Images Are Pushed to GitLab Container Registry

### Authentication

GitLab CI/CD provides built-in authentication:

```bash
# Automatic variables available in pipeline:
$CI_REGISTRY          # registry.gitlab.cs.ut.ee
$CI_REGISTRY_USER     # gitlab-ci-token
$CI_REGISTRY_PASSWORD # Auto-generated token
$CI_REGISTRY_IMAGE    # Full image path
```

### Push Process

**Step 1: Authenticate**
```bash
# Kaniko creates Docker config
echo "{\"auths\":{\"$CI_REGISTRY\":{...}}}" > /kaniko/.docker/config.json
```

**Step 2: Build and Push**
```bash
/kaniko/executor \
  --destination "${IMAGE_NAME}:${IMAGE_TAG}" \
  --destination "${IMAGE_NAME}:latest"
```

**Step 3: Verify**
```bash
# In GitLab UI: Project → Deploy → Container Registry
# Or via curl:
curl --header "PRIVATE-TOKEN: <token>" \
  "https://registry.gitlab.cs.ut.ee/v2/devops/teams/emzops/devops-be/tags/list"
```

---

## Kubernetes Manifest Files Structure

### Deployment (`k8s/deployment.yaml`)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: devops-backend
  namespace: emzops-staging
spec:
  replicas: 1
  selector:
    matchLabels:
      app: devops-backend
  template:
    metadata:
      labels:
        app: devops-backend
    spec:
      imagePullSecrets:
        - name: gitlab-registry-secret  # For private registry
      containers:
      - name: backend
        image: registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-be:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          valueFrom:
            configMapKeyRef:
              name: backend-config
              key: PORT
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: backend-secret
              key: DATABASE_URL
```

**Key Components:**
- `replicas: 1` - Phase 3 requirement
- `imagePullSecrets` - Authentication for private registry
- `configMapKeyRef` - Non-sensitive configuration
- `secretKeyRef` - Sensitive data (DATABASE_URL)

### Service (`k8s/service.yaml`)

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
    port: 80
    targetPort: 8080
```

**Purpose:** Internal cluster communication

### ConfigMap (`k8s/configmap.yaml`)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: backend-config
  namespace: emzops-staging
data:
  PORT: "8080"
```

**Purpose:** Non-sensitive environment variables

### Secret (NOT committed - `k8s/secret.yaml.template`)

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

**⚠️ Important:** Actual `secret.yaml` is in `.gitignore`

---

## How Pods Access Images from Private GitLab Registry

### The Problem

Kubernetes cannot pull images from private registries without authentication.

### The Solution: Docker Registry Secret

#### Step 1: Create GitLab Deploy Token

**Location:** GitLab → Settings → Repository → Deploy Tokens

**Configuration:**
- Name: `kubernetes-deploy`
- Username: (auto-generated)
- Scopes: ✅ `read_registry`
- Token: (save this!)

#### Step 2: Create Kubernetes Secret

```bash
kubectl create secret docker-registry gitlab-registry-secret \
  --docker-server=registry.gitlab.cs.ut.ee \
  --docker-username=<deploy-token-username> \
  --docker-password=<deploy-token-token> \
  --namespace=emzops-staging
```

#### Step 3: Reference in Deployment

```yaml
spec:
  template:
    spec:
      imagePullSecrets:
        - name: gitlab-registry-secret
```

### How It Works

```
Pod creation requested
  ↓
Kubernetes reads imagePullSecrets
  ↓
Retrieves credentials from secret
  ↓
Authenticates to registry.gitlab.cs.ut.ee
  ↓
Pulls image
  ↓
Starts container
```

### Verification

```bash
# Check secret exists
kubectl get secret gitlab-registry-secret -n emzops-staging

# View secret details (credentials are base64 encoded)
kubectl describe secret gitlab-registry-secret -n emzops-staging
```

---

## Secret Management for Registry Authentication

### Secrets in Use

| Secret Name | Type | Purpose |
|-------------|------|---------|
| `gitlab-registry-secret` | `docker-registry` | Pull images from GitLab |
| `backend-secret` | `Opaque` | Database connection string |
| `postgres-superuser-secret` | `basic-auth` | PostgreSQL credentials |

### Creating Docker Registry Secret

**Method 1: kubectl (recommended)**

```bash
kubectl create secret docker-registry gitlab-registry-secret \
  --docker-server=registry.gitlab.cs.ut.ee \
  --docker-username=<username> \
  --docker-password=<token> \
  --namespace=emzops-staging
```

**Method 2: YAML file**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gitlab-registry-secret
  namespace: emzops-staging
type: kubernetes.io/dockerconfigjson
stringData:
  .dockerconfigjson: |
    {
      "auths": {
        "registry.gitlab.cs.ut.ee": {
          "username": "<username>",
          "password": "<token>",
          "auth": "<base64(username:password)>"
        }
      }
    }
```

### Security Best Practices

- ✅ Never commit secrets to Git
- ✅ Use `.gitignore` for `secret.yaml`
- ✅ Store credentials in GitLab CI/CD variables
- ✅ Use deploy tokens (not personal access tokens)
- ✅ Set deploy token expiration dates
- ✅ Regularly rotate credentials

### GitLab CI/CD Integration

In the deployment stage, secrets are managed via GitLab variables:

```yaml
deploy-to-k8s:
  stage: deploy
  before_script:
    - echo "$KUBECONFIG_CONTENT" | base64 -d > ~/.kube/config
  script:
    - kubectl apply -f k8s/deployment.yaml
```

**Required GitLab CI/CD Variables:**
- `KUBECONFIG_CONTENT` - base64 encoded kubeconfig file
- Deploy token credentials (used for creating registry secret)

---

## Testing and Validation

### Port Forwarding Test (Week 9 Requirement)

```bash
# Start port forwarding
kubectl port-forward deployment/devops-backend -n emzops-staging 8080:8080

# In another terminal, test backend
curl -I http://localhost:8080

# Expected: Connection successful
```

**Test Result:** ✅ Backend responds successfully

### Deployment Verification

```bash
# Check deployment status
kubectl get deployment devops-backend -n emzops-staging

# Expected output:
# NAME             READY   UP-TO-DATE   AVAILABLE   AGE
# devops-backend   1/1     1            1           10d

# Check pod status
kubectl get pods -n emzops-staging -l app=devops-backend

# View logs
kubectl logs -f deployment/devops-backend -n emzops-staging

# Check events
kubectl describe deployment devops-backend -n emzops-staging
```

### Image Pull Verification

```bash
# Check pod events for image pull
kubectl describe pod <pod-name> -n emzops-staging | grep -A5 "Events"

# Successful output should show:
# - Pulling image
# - Successfully pulled image
# - Created container
# - Started container
```

---

## Common Issues and Solutions

### ImagePullBackOff Error

**Symptom:** Pod stuck in `ImagePullBackOff` state

**Diagnosis:**
```bash
kubectl describe pod <pod-name> -n emzops-staging
# Look for: "Failed to pull image" or "unauthorized"
```

**Solutions:**

1. **Check secret exists:**
   ```bash
   kubectl get secret gitlab-registry-secret -n emzops-staging
   ```

2. **Verify deploy token is active in GitLab:**
   - Settings → Repository → Deploy Tokens
   - Check expiration date

3. **Recreate secret:**
   ```bash
   kubectl delete secret gitlab-registry-secret -n emzops-staging
   kubectl create secret docker-registry gitlab-registry-secret ...
   ```

4. **Restart deployment:**
   ```bash
   kubectl rollout restart deployment/devops-backend -n emzops-staging
   ```

### CrashLoopBackOff Error

**Symptom:** Pod keeps restarting

**Diagnosis:**
```bash
kubectl logs <pod-name> -n emzops-staging
```

**Common causes:**
- Missing DATABASE_URL (check `backend-secret`)
- PostgreSQL not ready (check `postgres-service`)
- Application code error

---

## Deployment Workflow

### Manual Deployment

```bash
# Apply ConfigMap
kubectl apply -f k8s/configmap.yaml

# Apply Secret (created once)
kubectl apply -f k8s/secret.yaml

# Apply Deployment
kubectl apply -f k8s/deployment.yaml

# Apply Service
kubectl apply -f k8s/service.yaml

# Check rollout status
kubectl rollout status deployment/devops-backend -n emzops-staging
```

### Automatic Deployment (GitLab CI/CD)

Every push to `main` branch triggers:

```
1. Security scan (gitleaks)
2. Tests (go test)
3. Build (go build)
4. Build image (Kaniko)
5. Deploy to Kubernetes (kubectl apply)
```

### Rollback

```bash
# Rollback to previous version
kubectl rollout undo deployment/devops-backend -n emzops-staging

# Rollback to specific revision
kubectl rollout undo deployment/devops-backend --to-revision=2 -n emzops-staging

# Check rollout history
kubectl rollout history deployment/devops-backend -n emzops-staging
```

---

## Week 9 Checklist

- ✅ Dockerfile created
- ✅ GitLab CI/CD pipeline builds container images
- ✅ Images pushed to GitLab Container Registry
- ✅ Kubernetes manifests created
- ✅ Deploy token created in GitLab
- ✅ Docker registry secret created in Kubernetes
- ✅ imagePullSecrets referenced in deployment
- ✅ Application deployed with 1 replica
- ✅ Port forwarding test completed
- ✅ Documentation completed

---

## Useful Commands

```bash
# Set default namespace (less typing)
kubectl config set-context --current --namespace=emzops-staging

# Watch deployment
kubectl get deployment devops-backend -w

# Restart deployment
kubectl rollout restart deployment/devops-backend

# Scale deployment (testing only - Phase 3 uses 1 replica)
kubectl scale deployment devops-backend --replicas=2

# Update image
kubectl set image deployment/devops-backend backend=registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-be:abc123

# Port forward (background)
kubectl port-forward deployment/devops-backend 8080:8080 &

# Stop port forward
pkill -f "kubectl port-forward"
```

---

**For complete Week 9 documentation, see:** `/WEEK9_DOCUMENTATION.md` in project root

---

> **Note:** This document was prepared with the assistance of an AI (artificial intelligence) tool.