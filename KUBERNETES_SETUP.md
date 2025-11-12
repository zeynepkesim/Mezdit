# Complete Kubernetes Deployment Guide - Week 10

This comprehensive guide documents the full Kubernetes setup process for our team's backend and frontend applications.

**Team:** emzops
**Namespace:** emzops-staging
**Backend URL:** https://devops-mezdit.kubernetes.devops.cs.ut.ee
**Git Repository:** https://gitlab.cs.ut.ee/devops/teams/emzops/devops-be

---

## 📋 Table of Contents

1. [Current Status](#current-status)
2. [Architecture Overview](#architecture-overview)
3. [Complete Setup Process (Already Done)](#complete-setup-process-already-done)
4. [What Team Members Need to Do](#what-team-members-need-to-do)
5. [Backend Deployment Details](#backend-deployment-details)
6. [Frontend Deployment Guide](#frontend-deployment-guide)
7. [Common Operations](#common-operations)
8. [Troubleshooting](#troubleshooting)

---

## 🎯 Current Status

### ✅ What's Already Running

**Backend Application:**
- ✅ Kubernetes manifests created in `k8s/` directory
- ✅ PostgreSQL database (StatefulSet, 1Gi storage)
- ✅ Backend deployment (Go application)
- ✅ Service (ClusterIP)
- ✅ Ingress with automatic TLS (cert-manager)
- ✅ ConfigMap for non-sensitive config
- ✅ Secrets for sensitive data
- ✅ RBAC resources (ServiceAccount, Role, RoleBinding)
- ✅ GitLab CI/CD pipeline with automatic deployment
- ✅ GitLab Deploy Token for registry access
- ✅ Application running at: https://devops-mezdit.kubernetes.devops.cs.ut.ee

**CI/CD Pipeline:**
- ✅ Security scan (gitleaks)
- ✅ Tests (go test with PostgreSQL)
- ✅ Build (compile Go binary)
- ✅ Build Docker image (Kaniko)
- ✅ Deploy to Kubernetes (automatic)

### ⚠️ What Team Members Don't Need to Repeat

**The following were ONE-TIME setup tasks (already completed):**
- ❌ Creating Kubernetes manifest files
- ❌ Creating RBAC resources in cluster
- ❌ Creating GitLab Deploy Token
- ❌ Creating GitLab CI/CD variables
- ❌ Creating secrets in Kubernetes cluster
- ❌ Initial deployment to cluster

---

## 🏗️ Architecture Overview

### Before (VM Deployment)
```
GitLab Pipeline → VM (devops-mezdit-vm13.cloud.ut.ee:8080)
                  ↳ Single binary running
                  ↳ Manual systemd service
```

### After (Kubernetes Deployment)
```
GitLab Pipeline → Container Registry → Kubernetes Cluster
                                        ↳ namespace: emzops-staging
                                        ↳ PostgreSQL (StatefulSet)
                                        ↳ Backend (Deployment)
                                        ↳ Ingress (TLS) → External Access
```

### Resources in Kubernetes

```
emzops-staging namespace:
├── ConfigMap: backend-config
├── Secrets:
│   ├── backend-secret (DATABASE_URL)
│   ├── postgres-superuser-secret (DB credentials)
│   └── gitlab-registry-secret (image pull)
├── PostgreSQL:
│   ├── StatefulSet: postgres (1 replica)
│   ├── Service: postgres-service (ClusterIP None)
│   └── PVC: postgres-storage-postgres-0 (1Gi)
├── Backend:
│   ├── Deployment: devops-backend (1 replica)
│   ├── Service: devops-backend-service (ClusterIP)
│   └── Ingress: devops-backend-ingress (TLS)
└── RBAC:
    ├── ServiceAccount: gitlab-deploy
    ├── Role: gitlab-deploy-role
    └── RoleBinding: gitlab-deploy-binding
```

---

## 🔨 Complete Setup Process (Already Done)

This section documents what was done to set up Kubernetes deployment. **Team members don't need to repeat these steps.**

### Phase 1: Local Development Setup

**1. Install Required Tools**
```bash
brew install kubectl
brew install int128/kubelogin/kubelogin
```

**2. Configure kubectl Access**
```bash
mkdir -p ~/.kube
# Copy KUBECONFIG file from instructors to ~/.kube/config-ut
export KUBECONFIG=~/.kube/config-ut
echo 'export KUBECONFIG=~/.kube/config-ut' >> ~/.zshrc
kubectl get namespaces  # Test access
```

### Phase 2: Create Kubernetes Manifests

**Created the following files in `k8s/` directory:**

**1. ConfigMap (`k8s/configmap.yaml`)**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: backend-config
  namespace: emzops-staging
data:
  PORT: "8080"
```

**2. Secret Templates**
- `k8s/secret.yaml.template` - Backend secrets
- `k8s/postgres-secret.yaml.template` - PostgreSQL password

**3. PostgreSQL StatefulSet (`k8s/postgres-statefulset.yaml`)**
- 1 replica
- 1Gi persistent storage
- PostgreSQL 15-alpine image
- Uses secrets for credentials

**4. Backend Deployment (`k8s/deployment.yaml`)**
- 1 replica
- Image from GitLab registry
- Environment variables from ConfigMap and Secret
- Image pull secret configured

**5. Service (`k8s/service.yaml`)**
- Type: ClusterIP
- Port 80 → Container port 8080

**6. Ingress (`k8s/ingress.yaml`)**
- Host: `devops-mezdit.kubernetes.devops.cs.ut.ee`
- TLS with cert-manager annotation
- NGINX ingress class

**7. RBAC (`k8s/rbac.yaml`)**
- ServiceAccount: gitlab-deploy
- Role: permissions for deployments, services, configmaps, secrets, ingresses, statefulsets
- RoleBinding: binds role to service account

### Phase 3: Create Secrets in Kubernetes

**1. PostgreSQL Secret**
```bash
# Created k8s/postgres-secret.yaml (not in git)
kubectl apply -f k8s/postgres-secret.yaml
```

**2. Backend Secret**
```bash
# Created k8s/secret.yaml (not in git)
# DATABASE_URL: postgresql://devops:PASSWORD@postgres-service:5432/devops
kubectl apply -f k8s/secret.yaml
```

**3. GitLab Registry Secret**

First, created Deploy Token in GitLab:
- Settings → Repository → Deploy Tokens
- Name: `k8s-registry-access`
- Scopes: `read_registry` ✅

Then:
```bash
kubectl create secret docker-registry gitlab-registry-secret \
  --docker-server=registry.gitlab.cs.ut.ee \
  --docker-username=<deploy-token-username> \
  --docker-password=<deploy-token> \
  --docker-email=<email@ut.ee> \
  -n emzops-staging
```

### Phase 4: Create RBAC Resources

```bash
kubectl apply -f k8s/rbac.yaml
```

This created:
- ServiceAccount: `gitlab-deploy`
- Role: `gitlab-deploy-role`
- RoleBinding: `gitlab-deploy-binding`

### Phase 5: Create ServiceAccount Token for CI/CD

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: gitlab-deploy-token
  namespace: emzops-staging
  annotations:
    kubernetes.io/service-account.name: gitlab-deploy
type: kubernetes.io/service-account-token
EOF
```

### Phase 6: Generate Kubeconfig for GitLab CI/CD

```bash
# Extract token
TOKEN=$(kubectl get secret gitlab-deploy-token -n emzops-staging -o jsonpath='{.data.token}' | base64 -d)

# Get cluster info
CLUSTER_SERVER=$(kubectl config view -o jsonpath="{.clusters[0].cluster.server}")
CLUSTER_CA=$(kubectl config view --raw -o jsonpath="{.clusters[0].cluster.certificate-authority-data}")

# Create kubeconfig file
cat > gitlab-kubeconfig.yaml <<EOF
apiVersion: v1
kind: Config
clusters:
- name: k8s-cluster
  cluster:
    certificate-authority-data: $CLUSTER_CA
    server: $CLUSTER_SERVER
contexts:
- name: gitlab-context
  context:
    cluster: k8s-cluster
    namespace: emzops-staging
    user: gitlab-deploy
current-context: gitlab-context
users:
- name: gitlab-deploy
  user:
    token: $TOKEN
EOF

# Encode to base64 (single line, no line breaks)
base64 -w 0 < gitlab-kubeconfig.yaml
```

### Phase 7: Configure GitLab CI/CD

**1. Added GitLab CI/CD Variable**

Settings → CI/CD → Variables → Add variable:
- **Key**: `KUBECONFIG_CONTENT`
- **Value**: [base64 encoded kubeconfig from above]
- **Type**: Variable
- **Protected**: ✅ Yes
- **Masked**: ❌ No

**2. Updated `.gitlab-ci.yml`**

Key changes:
- Replaced Buildah with Kaniko (better Kubernetes compatibility)
- Added `deploy-to-k8s` stage
- Removed VM deployment stage

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

deploy-to-k8s:
  stage: deploy
  image: bitnami/kubectl:latest
  before_script:
    - mkdir -p ~/.kube
    - echo "$KUBECONFIG_CONTENT" | base64 -d > ~/.kube/config
    - chmod 600 ~/.kube/config
  script:
    - kubectl apply -f k8s/configmap.yaml
    - kubectl apply -f k8s/postgres-statefulset.yaml
    - kubectl apply -f k8s/deployment.yaml
    - kubectl apply -f k8s/service.yaml
    - kubectl apply -f k8s/ingress.yaml
    - kubectl rollout status deployment/devops-backend -n emzops-staging --timeout=5m
    - kubectl get pods -n emzops-staging
  environment:
    name: staging
    url: https://devops-mezdit.kubernetes.devops.cs.ut.ee
  only:
    - main
```

### Phase 8: Deploy Resources

**1. Apply ConfigMap**
```bash
kubectl apply -f k8s/configmap.yaml
```

**2. Deploy PostgreSQL**
```bash
kubectl apply -f k8s/postgres-statefulset.yaml
kubectl get pods -n emzops-staging -w  # Wait for postgres-0 to be Running
```

**3. Deploy Backend**
```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml
```

**4. Verify Deployment**
```bash
kubectl get all -n emzops-staging
kubectl get ingress -n emzops-staging
kubectl get certificate -n emzops-staging
kubectl logs -f deployment/devops-backend -n emzops-staging
```

### Phase 9: Test and Verify

```bash
# Test via port forwarding
kubectl port-forward service/devops-backend-service 8080:80 -n emzops-staging
curl http://localhost:8080/posts

# Test via HTTPS
curl https://devops-mezdit.kubernetes.devops.cs.ut.ee/posts
```

---

## 👥 What Team Members Need to Do

### ✅ Required Setup (All Team Members)

**1. Install Tools**
```bash
brew install kubectl
brew install int128/kubelogin/kubelogin
```

**2. Configure kubectl Access**
```bash
mkdir -p ~/.kube
# Copy KUBECONFIG file from course instructors to ~/.kube/config-ut

export KUBECONFIG=~/.kube/config-ut
echo 'export KUBECONFIG=~/.kube/config-ut' >> ~/.zshrc
source ~/.zshrc

# Test access
kubectl get namespaces
kubectl get pods -n emzops-staging
```

**3. Verify You Can See Resources**
```bash
kubectl get all -n emzops-staging
kubectl get ingress -n emzops-staging
kubectl logs -f deployment/devops-backend -n emzops-staging
```

### ❌ What You DON'T Need to Do

You **DON'T** need to:
- ❌ Create any Kubernetes manifests (already in git)
- ❌ Create secrets in Kubernetes (already done)
- ❌ Create RBAC resources (already done)
- ❌ Create GitLab Deploy Token (already done)
- ❌ Configure GitLab CI/CD variables (already done)
- ❌ Deploy anything manually (CI/CD does it automatically)

### 🔄 Daily Workflow

**Normal development workflow:**

1. Make code changes
2. Commit and push to `main` branch:
   ```bash
   git add .
   git commit -m "Your changes"
   git push origin main
   ```
3. GitLab pipeline automatically:
   - Runs security scan
   - Runs tests
   - Builds binary
   - Builds Docker image
   - **Deploys to Kubernetes automatically**
4. Check deployment:
   ```bash
   kubectl get pods -n emzops-staging -w
   kubectl logs -f deployment/devops-backend -n emzops-staging
   ```

---

## 🔧 Backend Deployment Details

### File Structure

```
devops-be/
├── .gitlab-ci.yml              # CI/CD pipeline configuration
├── Dockerfile                  # Container image definition
├── k8s/                        # Kubernetes manifests
│   ├── configmap.yaml          # Non-sensitive config (in git)
│   ├── deployment.yaml         # Backend deployment (in git)
│   ├── ingress.yaml            # Ingress + TLS (in git)
│   ├── postgres-cluster.yaml  # CloudNativePG (not used)
│   ├── postgres-statefulset.yaml  # PostgreSQL (in git)
│   ├── postgres-secret.yaml.template  # DB password template
│   ├── postgres-secret.yaml   # Actual secret (NOT in git)
│   ├── rbac.yaml               # RBAC resources (in git)
│   ├── secret.yaml.template   # Backend secret template
│   ├── secret.yaml             # Actual secret (NOT in git)
│   └── service.yaml            # Backend service (in git)
├── .gitignore                  # Includes k8s/secret.yaml, k8s/postgres-secret.yaml
└── KUBERNETES_SETUP.md         # This file
```

### Important: Secret Files

**These files are NOT in git** (listed in `.gitignore`):
- `k8s/secret.yaml` - Backend secrets (DATABASE_URL)
- `k8s/postgres-secret.yaml` - PostgreSQL password

**Templates are provided:**
- `k8s/secret.yaml.template`
- `k8s/postgres-secret.yaml.template`

**If you need the actual passwords**, ask a team member who has them.

### Database Connection

**Connection string format:**
```
postgresql://devops:PASSWORD@postgres-service:5432/devops
```

- **Host**: `postgres-service` (Kubernetes service name)
- **Port**: `5432`
- **Database**: `devops`
- **User**: `devops`
- **Password**: (ask team)

### Environment Variables

**ConfigMap (non-sensitive):**
- `PORT`: "8080"

**Secret (sensitive):**
- `DATABASE_URL`: PostgreSQL connection string

---

## 🎨 Frontend Deployment Guide

### Current State

Frontend is currently deployed to VM at: `http://devops-mezdit-vm13.cloud.ut.ee`

### Migration to Kubernetes

**Step-by-step guide to deploy frontend to Kubernetes:**

#### 1. Create Frontend Dockerfile

If your frontend repository doesn't have a Dockerfile, create one:

**For React/Vue/Angular (Static Build):**
```dockerfile
# Build stage
FROM node:18 AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
# or for Vue: /app/dist
# or for Angular: /app/dist/your-app-name

# Copy custom nginx config if needed
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

**nginx.conf example:**
```nginx
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # API proxy (if backend is separate)
    location /api {
        proxy_pass https://devops-mezdit.kubernetes.devops.cs.ut.ee;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

#### 2. Create Frontend Kubernetes Manifests

Create a `k8s/` directory in your frontend repository:

**k8s/configmap.yaml:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: frontend-config
  namespace: emzops-staging
data:
  REACT_APP_API_URL: "https://devops-mezdit.kubernetes.devops.cs.ut.ee"
  # Add other frontend environment variables
```

**k8s/deployment.yaml:**
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
        - name: gitlab-registry-secret
      containers:
      - name: frontend
        image: registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-fe:latest
        ports:
        - containerPort: 80
        envFrom:
        - configMapRef:
            name: frontend-config
```

**k8s/service.yaml:**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: devops-frontend-service
  namespace: emzops-staging
spec:
  selector:
    app: devops-frontend
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: ClusterIP
```

**k8s/ingress.yaml:**
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: devops-frontend-ingress
  namespace: emzops-staging
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - devops-mezdit-frontend.kubernetes.devops.cs.ut.ee
    secretName: frontend-tls-cert
  rules:
  - host: devops-mezdit-frontend.kubernetes.devops.cs.ut.ee
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

#### 3. Update Frontend GitLab CI/CD

**Update `.gitlab-ci.yml` in frontend repository:**

```yaml
stages:
  - test
  - build-image
  - deploy

variables:
  IMAGE_NAME: $CI_REGISTRY_IMAGE
  IMAGE_TAG: $CI_COMMIT_SHORT_SHA

# Add tests if you have them
test:
  stage: test
  image: node:18
  script:
    - npm ci
    - npm run test
  only:
    - main

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
    - echo "✅ Frontend image pushed - $IMAGE_NAME:latest"
  only:
    - main

deploy-to-k8s:
  stage: deploy
  image: bitnami/kubectl:latest
  before_script:
    - mkdir -p ~/.kube
    - echo "$KUBECONFIG_CONTENT" | base64 -d > ~/.kube/config
    - chmod 600 ~/.kube/config
  script:
    - echo "Deploying frontend to Kubernetes namespace emzops-staging..."
    - kubectl apply -f k8s/configmap.yaml
    - kubectl apply -f k8s/deployment.yaml
    - kubectl apply -f k8s/service.yaml
    - kubectl apply -f k8s/ingress.yaml
    - echo "Waiting for deployment to be ready..."
    - kubectl rollout status deployment/devops-frontend -n emzops-staging --timeout=5m
    - echo "✅ Deployment successful!"
    - kubectl get pods -n emzops-staging
  environment:
    name: staging
    url: https://devops-mezdit-frontend.kubernetes.devops.cs.ut.ee
  only:
    - main
```

#### 4. Add GitLab CI/CD Variable (Frontend Repo)

**In frontend GitLab repository:**

Settings → CI/CD → Variables → Add variable:
- **Key**: `KUBECONFIG_CONTENT`
- **Value**: [same base64 encoded kubeconfig from backend - available in `KUBECONFIG_BASE64.txt`]
- **Type**: Variable
- **Protected**: ✅ Yes
- **Masked**: ❌ No

#### 5. Deploy Frontend

```bash
# Push changes to trigger pipeline
git add .
git commit -m "Add Kubernetes deployment for frontend"
git push origin main

# Monitor deployment
kubectl get pods -n emzops-staging -w
kubectl logs -f deployment/devops-frontend -n emzops-staging
```

#### 6. Verify Frontend

```bash
# Check resources
kubectl get deployment devops-frontend -n emzops-staging
kubectl get service devops-frontend-service -n emzops-staging
kubectl get ingress devops-frontend-ingress -n emzops-staging

# Check TLS certificate
kubectl get certificate frontend-tls-cert -n emzops-staging

# Access frontend
curl https://devops-mezdit-frontend.kubernetes.devops.cs.ut.ee
```

### Frontend + Backend Integration

**Important configuration:**

1. **Frontend environment variable** (`k8s/configmap.yaml`):
   ```yaml
   REACT_APP_API_URL: "https://devops-mezdit.kubernetes.devops.cs.ut.ee"
   ```

2. **CORS in Backend**: Ensure backend allows requests from frontend domain

3. **WebSocket connections**: If using WebSockets, configure proxy in nginx:
   ```nginx
   location /ws {
       proxy_pass https://devops-mezdit.kubernetes.devops.cs.ut.ee;
       proxy_http_version 1.1;
       proxy_set_header Upgrade $http_upgrade;
       proxy_set_header Connection "Upgrade";
   }
   ```

---

## 🛠️ Common Operations

### Viewing Logs

```bash
# Backend logs
kubectl logs -f deployment/devops-backend -n emzops-staging

# Frontend logs
kubectl logs -f deployment/devops-frontend -n emzops-staging

# PostgreSQL logs
kubectl logs -f postgres-0 -n emzops-staging

# Get logs from specific pod
kubectl logs <pod-name> -n emzops-staging

# Get previous pod logs (if crashed)
kubectl logs <pod-name> -n emzops-staging --previous
```

### Restarting Applications

```bash
# Restart backend (rolling restart)
kubectl rollout restart deployment/devops-backend -n emzops-staging

# Restart frontend
kubectl rollout restart deployment/devops-frontend -n emzops-staging

# Force delete and recreate pod
kubectl delete pod <pod-name> -n emzops-staging
```

### Viewing Resources

```bash
# All resources
kubectl get all -n emzops-staging

# Specific resources
kubectl get pods -n emzops-staging
kubectl get deployments -n emzops-staging
kubectl get services -n emzops-staging
kubectl get ingress -n emzops-staging
kubectl get configmaps -n emzops-staging
kubectl get secrets -n emzops-staging

# Watch resources (auto-refresh)
kubectl get pods -n emzops-staging -w

# Detailed information
kubectl describe pod <pod-name> -n emzops-staging
kubectl describe deployment devops-backend -n emzops-staging
```

### Port Forwarding (Local Testing)

```bash
# Forward backend
kubectl port-forward service/devops-backend-service 8080:80 -n emzops-staging
# Access: http://localhost:8080

# Forward frontend
kubectl port-forward service/devops-frontend-service 3000:80 -n emzops-staging
# Access: http://localhost:3000

# Forward PostgreSQL
kubectl port-forward postgres-0 5432:5432 -n emzops-staging
# Connect: psql -h localhost -U devops -d devops
```

### Executing Commands in Pods

```bash
# Get shell in backend pod
kubectl exec -it <backend-pod-name> -n emzops-staging -- /bin/sh

# Get shell in PostgreSQL
kubectl exec -it postgres-0 -n emzops-staging -- /bin/sh

# Run single command
kubectl exec postgres-0 -n emzops-staging -- psql -U devops -d devops -c "SELECT * FROM users;"
```

### Viewing and Editing ConfigMaps

```bash
# View ConfigMap
kubectl get configmap backend-config -n emzops-staging -o yaml

# Edit ConfigMap (opens in editor)
kubectl edit configmap backend-config -n emzops-staging

# After editing, restart deployment to pick up changes
kubectl rollout restart deployment/devops-backend -n emzops-staging
```

### Viewing Secrets

```bash
# List secrets
kubectl get secrets -n emzops-staging

# View secret (base64 encoded)
kubectl get secret backend-secret -n emzops-staging -o yaml

# Decode secret value
kubectl get secret backend-secret -n emzops-staging -o jsonpath='{.data.DATABASE_URL}' | base64 -d
echo ""
```

### Checking TLS Certificates

```bash
# List certificates
kubectl get certificate -n emzops-staging

# Describe certificate (shows status and errors)
kubectl describe certificate backend-tls-cert -n emzops-staging
kubectl describe certificate frontend-tls-cert -n emzops-staging

# Check certificate expiry
kubectl get certificate backend-tls-cert -n emzops-staging -o jsonpath='{.status.notAfter}'
```

### Viewing Events

```bash
# Recent events in namespace
kubectl get events -n emzops-staging --sort-by='.lastTimestamp'

# Watch events
kubectl get events -n emzops-staging -w

# Events for specific resource
kubectl describe pod <pod-name> -n emzops-staging | grep -A 20 Events:
```

---

## 🆘 Troubleshooting

### Pod Not Starting

**Symptoms:**
- Pod stuck in `Pending`, `ContainerCreating`, or `CrashLoopBackOff`

**Diagnosis:**
```bash
# Check pod status
kubectl get pods -n emzops-staging

# Describe pod (shows events and errors)
kubectl describe pod <pod-name> -n emzops-staging

# Check logs
kubectl logs <pod-name> -n emzops-staging

# Check previous logs if pod crashed
kubectl logs <pod-name> -n emzops-staging --previous
```

**Common Issues:**

1. **ImagePullBackOff**
   - Image doesn't exist in registry
   - Image pull secret is incorrect
   - Check: `kubectl describe pod <pod-name> -n emzops-staging | grep Image`

2. **CrashLoopBackOff**
   - Application is crashing
   - Check logs: `kubectl logs <pod-name> -n emzops-staging --previous`
   - Common causes: database connection issues, missing environment variables

3. **Pending**
   - Not enough resources
   - Storage issues
   - Check: `kubectl describe pod <pod-name> -n emzops-staging | grep -A 10 Events:`

### Database Connection Issues

**Symptoms:**
- Backend logs show database connection errors
- Application can't start

**Diagnosis:**
```bash
# Check if PostgreSQL is running
kubectl get pods -n emzops-staging | grep postgres

# Check PostgreSQL logs
kubectl logs postgres-0 -n emzops-staging

# Test PostgreSQL connection
kubectl exec -it postgres-0 -n emzops-staging -- psql -U devops -d devops -c "SELECT 1;"

# Check if service exists
kubectl get service postgres-service -n emzops-staging

# Verify DATABASE_URL secret
kubectl get secret backend-secret -n emzops-staging -o jsonpath='{.data.DATABASE_URL}' | base64 -d
echo ""
```

**Common Issues:**
1. Wrong password in DATABASE_URL
2. PostgreSQL pod not running
3. Wrong service name (should be `postgres-service`)
4. Database doesn't exist (create it in PostgreSQL)

### Ingress Not Working

**Symptoms:**
- Can't access application via HTTPS URL
- SSL/TLS certificate errors

**Diagnosis:**
```bash
# Check ingress
kubectl get ingress -n emzops-staging

# Describe ingress
kubectl describe ingress devops-backend-ingress -n emzops-staging

# Check certificate
kubectl get certificate -n emzops-staging
kubectl describe certificate backend-tls-cert -n emzops-staging

# Check if cert-manager is working
kubectl get certificaterequest -n emzops-staging
```

**Common Issues:**
1. Certificate not ready (wait 1-2 minutes)
2. DNS not configured correctly
3. Ingress controller issues

**Solution:**
```bash
# Delete and recreate certificate
kubectl delete certificate backend-tls-cert -n emzops-staging
kubectl delete secret backend-tls-cert -n emzops-staging
kubectl delete ingress devops-backend-ingress -n emzops-staging
kubectl apply -f k8s/ingress.yaml
```

### CI/CD Pipeline Failing

**Common Pipeline Issues:**

1. **"base64: invalid input"** in deploy stage
   - KUBECONFIG_CONTENT has line breaks
   - Solution: Regenerate base64 without line breaks: `base64 -w 0 < file`

2. **"Forbidden" errors** in deploy stage
   - ServiceAccount doesn't have permissions
   - Solution: Check RBAC role has necessary permissions

3. **Image build failing** with Kaniko
   - Dockerfile syntax errors
   - Solution: Test Dockerfile locally with Docker

4. **"ImagePullBackOff"** after deployment
   - Image wasn't pushed to registry
   - Registry secret is wrong
   - Solution: Check build-docker-image stage succeeded

### Permission Issues

**Symptoms:**
- "Forbidden" errors when running kubectl commands
- Can't access certain resources

**Diagnosis:**
```bash
# Check current context
kubectl config current-context

# Check which user you're using
kubectl config view

# Try to list resources
kubectl get pods -n emzops-staging
```

**Solution:**
```bash
# Make sure KUBECONFIG is set correctly
echo $KUBECONFIG

# Should be: ~/.kube/config-ut or similar
export KUBECONFIG=~/.kube/config-ut
```

### Storage Issues

**Symptoms:**
- PostgreSQL pod stuck in Pending
- PVC errors

**Diagnosis:**
```bash
# Check PVC status
kubectl get pvc -n emzops-staging

# Describe PVC
kubectl describe pvc postgres-storage-postgres-0 -n emzops-staging

# Check storage class
kubectl get storageclass
```

**Common Issues:**
1. Requested storage too small (minimum 1Gi)
2. Storage class doesn't exist
3. Not enough disk space in cluster

---

## 📚 Additional Resources

### Documentation
- [Kubernetes Official Docs](https://kubernetes.io/docs/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [GitLab Kubernetes Integration](https://docs.gitlab.com/ee/user/project/clusters/)
- [Kaniko Documentation](https://github.com/GoogleContainerTools/kaniko)
- [cert-manager Documentation](https://cert-manager.io/docs/)

### Course Resources
- Week 9: https://devops.cs.ut.ee/phase3/week9/
- Week 10: https://devops.cs.ut.ee/phase3/week10/

### Useful kubectl Commands

```bash
# Get help for any command
kubectl --help
kubectl get --help
kubectl describe --help

# Explain resource types
kubectl explain deployment
kubectl explain deployment.spec
kubectl explain pod.spec.containers

# Get resource YAML
kubectl get deployment devops-backend -n emzops-staging -o yaml

# Get resource JSON
kubectl get deployment devops-backend -n emzops-staging -o json

# Use JSONPath to extract specific fields
kubectl get pods -n emzops-staging -o jsonpath='{.items[*].metadata.name}'
```

---

## 💡 Pro Tips

1. **Use aliases** for faster kubectl commands:
   ```bash
   alias k='kubectl'
   alias kgp='kubectl get pods -n emzops-staging'
   alias kgs='kubectl get services -n emzops-staging'
   alias kl='kubectl logs -f -n emzops-staging'
   ```

2. **Use stern** for better log viewing (multiple pods):
   ```bash
   brew install stern
   stern devops-backend -n emzops-staging
   ```

3. **Use k9s** for terminal UI:
   ```bash
   brew install k9s
   k9s -n emzops-staging
   ```

4. **Tab completion** for kubectl:
   ```bash
   # Add to ~/.zshrc
   source <(kubectl completion zsh)
   ```

5. **Always specify namespace** or set default:
   ```bash
   kubectl config set-context --current --namespace=emzops-staging
   ```

6. **Use `--dry-run` to test changes**:
   ```bash
   kubectl apply -f k8s/deployment.yaml --dry-run=client
   ```

7. **Use `diff` before applying**:
   ```bash
   kubectl diff -f k8s/deployment.yaml
   ```

---

## 🎯 Summary

### Current Deployment

- **Backend**: https://devops-mezdit.kubernetes.devops.cs.ut.ee
- **Frontend**: (To be deployed) https://devops-mezdit-frontend.kubernetes.devops.cs.ut.ee
- **Namespace**: `emzops-staging`
- **Deployment Method**: GitLab CI/CD → Kubernetes (automatic)

### Key Points

1. ✅ All Kubernetes resources already created and deployed
2. ✅ CI/CD pipeline automatically deploys on push to `main`
3. ✅ Team members only need kubectl access to view/debug
4. ✅ Secrets are not in git - templates provided
5. ✅ Frontend can be deployed following the same pattern as backend

### Team Workflow

```
Developer → Code Changes → Git Push → GitLab Pipeline → Kubernetes Cluster
                                           ↓
                                      (automatic)
                                           ↓
                                    Application Updated
```

---

**Questions or Issues?** Ask in team chat or refer to course instructors.

**Document Version:** 1.0
**Last Updated:** 2025-01-12
**Author:** Team emzops
