# Week 9 - Kubernetes Deployment Guide (Frontend)

**Project:** DevOps Frontend  
**Team:** emzops  
**Namespace:** emzops-staging  
**Registry:** `registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-fe`

---

## Quick Reference

### Container Image

**Build:** Kaniko in GitLab CI/CD pipeline  
**Base Images:**
- Build: `node:22-alpine`
- Production: `nginx:alpine`

**Tags:**
- `latest` - Most recent build
- `<commit-sha>` - Specific version

### Deployment

**Replicas:** 1 (Phase 3 requirement)  
**Port:** 80  
**Service:** `devops-frontend-service` (ClusterIP)  
**Web Server:** Nginx with custom configuration

### Authentication

**Image Pull Secret:** `frontend-registry-secret`  
**Type:** GitLab Deploy Token with `read_registry` scope

---

## How Container Images Are Built

### 1. Multi-Stage Dockerfile

The frontend uses a **multi-stage build** to minimize production image size:

```dockerfile
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

**Benefits:**
- ✅ Build dependencies NOT included in final image
- ✅ Only `dist/` folder and nginx in production
- ✅ Significantly smaller image size
- ✅ Faster deployments and pulls

### 2. GitLab CI/CD Pipeline

The `.gitlab-ci.yml` file defines our build process:

```yaml
build-docker-image:
  stage: build-image
  image:
    name: gcr.io/kaniko-project/executor:v1.23.0-debug
    entrypoint: [""]
  script:
    - echo "Building frontend container image with Kaniko..."
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

### 3. Build Process Flow

```
Code Push to main
  ↓
GitLab CI/CD Triggered
  ↓
Build Stage: npm ci && npm run build
  ↓
Build-Image Stage: Kaniko Executor
  ↓
Multi-stage Dockerfile:
  - Build app with Node.js
  - Copy dist/ to nginx container
  ↓
Tags Image (latest + commit SHA)
  ↓
Pushes to GitLab Container Registry
  ↓
Deployment Stage Updates Kubernetes
```

### 4. Why Kaniko?

- ✅ No Docker daemon required
- ✅ Works in unprivileged CI/CD runners
- ✅ Layer caching for faster builds
- ✅ Supports multi-stage builds
- ✅ Secure, rootless container builds

---

## Image Tagging and Versioning Strategy

### Dual-Tag Approach

**1. Latest Tag:**
```
registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-fe:latest
```
- Always points to most recent build
- Used by Kubernetes deployments
- Automatically updated on every `main` branch build

**2. Commit SHA Tag:**
```
registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-fe:a1b2c3d
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
```

---

## Kubernetes Manifest Files Structure

### Deployment (`k8s/deployment.yaml`)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: devops-frontend
  namespace: emzops-staging
  labels:
    app: devops-frontend
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
        - name: frontend-registry-secret  # For private registry
      containers:
      - name: frontend
        image: registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-fe:latest
        ports:
        - containerPort: 80
          name: http
        envFrom:
        - configMapRef:
            name: frontend-config
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
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

**Key Components:**
- `replicas: 1` - Phase 3 requirement
- `imagePullSecrets` - Authentication for private registry
- `envFrom` - Load all ConfigMap values as environment variables
- `resources` - Memory and CPU limits
- `livenessProbe` - Restart pod if unhealthy
- `readinessProbe` - Don't send traffic until ready

### Service (`k8s/service.yaml`)

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

**Purpose:** Internal cluster communication to frontend pods

### ConfigMap (`k8s/configmap.yaml`)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: frontend-config
  namespace: emzops-staging
data:
  VUE_APP_API_URL: "/api"
  VUE_APP_WS_URL: "/ws"
```

**Purpose:** Frontend environment variables for API endpoints

### Ingress (`k8s/ingress.yaml`)

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: devops-frontend-ingress
  namespace: emzops-staging
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - devops-mezdit.kubernetes.devops.cs.ut.ee
    secretName: frontend-tls-cert
  rules:
  - host: devops-mezdit.kubernetes.devops.cs.ut.ee
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: devops-frontend-service
            port:
              number: 80
```

**Purpose:** External HTTPS access with automatic TLS

---

## How Pods Access Images from Private GitLab Registry

### The Problem

Kubernetes cannot pull images from private registries without authentication.

### The Solution: Docker Registry Secret

#### Step 1: Create GitLab Deploy Token

**Location:** GitLab → Settings → Repository → Deploy Tokens

**Configuration:**
- Name: `kubernetes-deploy-frontend`
- Username: (auto-generated)
- Scopes: ✅ `read_registry`
- Token: (save this!)

#### Step 2: Create Kubernetes Secret

```bash
kubectl create secret docker-registry frontend-registry-secret \
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
        - name: frontend-registry-secret
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
Starts nginx container
```

### Verification

```bash
# Check secret exists
kubectl get secret frontend-registry-secret -n emzops-staging

# View secret details
kubectl describe secret frontend-registry-secret -n emzops-staging
```

---

## Secret Management for Registry Authentication

### Secrets in Use

| Secret Name | Type | Purpose |
|-------------|------|---------|
| `frontend-registry-secret` | `docker-registry` | Pull images from GitLab |
| `frontend-tls-cert` | `kubernetes.io/tls` | HTTPS certificate (auto-managed) |

### Creating Docker Registry Secret

**Method 1: kubectl (recommended)**

```bash
kubectl create secret docker-registry frontend-registry-secret \
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
  name: frontend-registry-secret
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
- ✅ Use deploy tokens (not personal access tokens)
- ✅ Set deploy token expiration dates
- ✅ Regularly rotate credentials
- ✅ Store credentials in GitLab CI/CD variables

---

## Nginx Configuration

### Custom Nginx Config

The `nginx.conf` file handles:
- SPA routing (Vue Router)
- API proxying
- WebSocket upgrades
- Security headers

**File:** `nginx.conf`

```nginx
server {
    listen 80;
    server_name _;

    root /usr/share/nginx/html;
    index index.html;

    # SPA routing - serve index.html for all routes
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Proxy API requests to backend
    location /api {
        proxy_pass http://devops-backend-service:80;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # WebSocket support
    location /ws {
        proxy_pass http://devops-backend-service:80;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
    }

    # Gzip compression
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
}
```

---

## Testing and Validation

### Port Forwarding Test (Week 9 Requirement)

```bash
# Start port forwarding
kubectl port-forward deployment/devops-frontend -n emzops-staging 8081:80

# Test in browser
open http://localhost:8081

# Or test with curl
curl -I http://localhost:8081

# Expected: HTTP 200 OK
# Expected: Server: nginx
```

**Test Result:** ✅ Frontend serves HTML successfully

### Deployment Verification

```bash
# Check deployment status
kubectl get deployment devops-frontend -n emzops-staging

# Expected output:
# NAME              READY   UP-TO-DATE   AVAILABLE   AGE
# devops-frontend   1/1     1            1           31h

# Check pod status
kubectl get pods -n emzops-staging -l app=devops-frontend

# View nginx logs
kubectl logs -f deployment/devops-frontend -n emzops-staging

# Check events
kubectl describe deployment devops-frontend -n emzops-staging
```

### Health Check Verification

```bash
# Check readiness probe
kubectl get pods -n emzops-staging -l app=devops-frontend

# Pod should show 1/1 READY (both liveness and readiness passing)

# Describe pod to see probe results
kubectl describe pod <pod-name> -n emzops-staging | grep -A10 "Liveness\|Readiness"
```

### Public Access Test

**URL:** https://devops-mezdit.kubernetes.devops.cs.ut.ee

```bash
# Test HTTPS access
curl -I https://devops-mezdit.kubernetes.devops.cs.ut.ee

# Expected:
# HTTP/2 200
# server: nginx
# Valid TLS certificate
```

---

## Common Issues and Solutions

### ImagePullBackOff Error

**Symptom:** Pod stuck in `ImagePullBackOff` state

**Solutions:**

1. **Check secret exists:**
   ```bash
   kubectl get secret frontend-registry-secret -n emzops-staging
   ```

2. **Verify deploy token is active in GitLab**

3. **Recreate secret:**
   ```bash
   kubectl delete secret frontend-registry-secret -n emzops-staging
   kubectl create secret docker-registry frontend-registry-secret ...
   ```

4. **Restart deployment:**
   ```bash
   kubectl rollout restart deployment/devops-frontend -n emzops-staging
   ```

### Nginx 404 Errors (SPA Routing)

**Symptom:** Vue Router routes return 404

**Solution:** Verify `nginx.conf` has:
```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

### API Requests Failing

**Symptom:** Frontend can't connect to backend

**Check:**
1. Backend service exists:
   ```bash
   kubectl get service devops-backend-service -n emzops-staging
   ```

2. Nginx config proxies correctly:
   ```nginx
   location /api {
       proxy_pass http://devops-backend-service:80;
   }
   ```

3. ConfigMap has correct API URL:
   ```bash
   kubectl get configmap frontend-config -n emzops-staging -o yaml
   ```

---

## Deployment Workflow

### Manual Deployment

```bash
# Apply ConfigMap
kubectl apply -f k8s/configmap.yaml

# Apply Deployment
kubectl apply -f k8s/deployment.yaml

# Apply Service
kubectl apply -f k8s/service.yaml

# Apply Ingress
kubectl apply -f k8s/ingress.yaml

# Check rollout status
kubectl rollout status deployment/devops-frontend -n emzops-staging
```

### Automatic Deployment (GitLab CI/CD)

Every push to `main` branch triggers:

```
1. Security scan (gitleaks)
2. Build (npm ci && npm run build)
3. Build image (Kaniko)
4. Deploy to Kubernetes (kubectl apply)
```

### Rollback

```bash
# Rollback to previous version
kubectl rollout undo deployment/devops-frontend -n emzops-staging

# Rollback to specific revision
kubectl rollout undo deployment/devops-frontend --to-revision=2 -n emzops-staging

# Check rollout history
kubectl rollout history deployment/devops-frontend -n emzops-staging
```

---

## Week 9 Checklist

- ✅ Multi-stage Dockerfile created
- ✅ GitLab CI/CD pipeline builds container images
- ✅ Images pushed to GitLab Container Registry
- ✅ Kubernetes manifests created
- ✅ Deploy token created in GitLab
- ✅ Docker registry secret created in Kubernetes
- ✅ imagePullSecrets referenced in deployment
- ✅ Application deployed with 1 replica
- ✅ Nginx configuration for SPA routing
- ✅ Health probes configured
- ✅ Port forwarding test completed
- ✅ Public HTTPS access working
- ✅ Documentation completed

---

## Useful Commands

```bash
# Set default namespace
kubectl config set-context --current --namespace=emzops-staging

# Watch deployment
kubectl get deployment devops-frontend -w

# Restart deployment
kubectl rollout restart deployment/devops-frontend

# Update image
kubectl set image deployment/devops-frontend frontend=registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-fe:abc123

# Port forward
kubectl port-forward deployment/devops-frontend 8081:80 &

# Stop port forward
pkill -f "kubectl port-forward"

# View nginx access logs
kubectl logs -f deployment/devops-frontend -n emzops-staging

# Execute into pod (debugging)
kubectl exec -it <pod-name> -n emzops-staging -- /bin/sh

# Check nginx config inside pod
kubectl exec <pod-name> -n emzops-staging -- cat /etc/nginx/conf.d/default.conf
```

---

**For complete Week 9 documentation, see:** `/WEEK9_DOCUMENTATION.md` in project root

---

> **Note:** This document was prepared with the assistance of an AI (artificial intelligence) tool.