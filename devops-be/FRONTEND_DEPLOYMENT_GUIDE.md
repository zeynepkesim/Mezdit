# Vue.js Frontend - Kubernetes Deployment Guide

This document contains all the steps required to deploy your Vue.js frontend application to Kubernetes.

---

## 📋 Overview

**Current State:**
- Frontend is currently running on VM: `http://devops-mezdit-vm13.cloud.ut.ee`

**Target:**
- Frontend and Backend will run on a single domain: `https://devops-mezdit.kubernetes.devops.cs.ut.ee`
  - Frontend: `https://devops-mezdit.kubernetes.devops.cs.ut.ee/` (root path)
  - Backend API: `https://devops-mezdit.kubernetes.devops.cs.ut.ee/api/` (API path)
- Automatic deployment (GitLab CI/CD)
- Secure access via HTTPS
- No CORS issues (same-origin)

---

## 🚀 Step-by-Step Setup

### 1. Create Dockerfile

Create a **Dockerfile** in your frontend repository:

```dockerfile
# Build stage - Build Vue.js application
FROM node:18 AS builder

WORKDIR /app

# Copy package.json and package-lock.json
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy all source code
COPY . .

# Production build
RUN npm run build

# Production stage - Serve with NGINX
FROM nginx:alpine

# Copy built files to NGINX
COPY --from=builder /app/dist /usr/share/nginx/html

# Custom NGINX config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

### 2. Create NGINX Configuration

Create **nginx.conf** file:

```nginx
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    # For Vue.js SPA - redirect all routes to index.html
    # Except API and WS paths, all other requests go to index.html
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json;
}
```

**IMPORTANT NOTE:** This simple NGINX config is sufficient because routing will be done through Kubernetes Ingress. API requests will be routed to the backend at the Ingress level.

### 3. Create Kubernetes Manifest Directory

Create a `k8s/` directory in your frontend repository:

```bash
mkdir k8s
cd k8s
```

### 4. Create ConfigMap

**k8s/configmap.yaml:**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: frontend-config
  namespace: emzops-staging
data:
  # Backend API URL - Relative path can be used since it's same origin
  VUE_APP_API_URL: "/api"
  # WebSocket URL (if you're using it) - Same origin
  VUE_APP_WS_URL: "/ws"
  # Other environment variables can be added here
```

**Note:**
- Environment variables in Vue.js must start with `VUE_APP_` prefix
- Relative paths are sufficient since we're using a single domain (no CORS issues)

### 5. Create Deployment

**k8s/deployment.yaml:**

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
      # Secret for pulling images from GitLab Registry
      imagePullSecrets:
        - name: gitlab-registry-secret
      containers:
      - name: frontend
        # Change image name according to your frontend repository
        image: registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-fe:latest
        ports:
        - containerPort: 80
          name: http
        # Get environment variables from ConfigMap
        envFrom:
        - configMapRef:
            name: frontend-config
        # Resource limits (optional but recommended)
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        # Health check (optional)
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

### 6. Create Service

**k8s/service.yaml:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: devops-frontend-service
  namespace: emzops-staging
  labels:
    app: devops-frontend
spec:
  type: ClusterIP
  selector:
    app: devops-frontend
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
    name: http
```

### 7. Update Ingress (Update Existing Backend Ingress)

**WARNING:** Do not create a new Ingress! Update your existing backend Ingress.

Update the `k8s/ingress.yaml` file in your backend repository as follows:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: devops-ingress
  namespace: emzops-staging
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - devops-mezdit.kubernetes.devops.cs.ut.ee
    secretName: devops-tls-cert
  rules:
  - host: devops-mezdit.kubernetes.devops.cs.ut.ee
    http:
      paths:
      # Backend API - /api path (must come before frontend)
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: devops-backend-service
            port:
              number: 80
      # WebSocket (if you're using it)
      - path: /ws
        pathType: Prefix
        backend:
          service:
            name: devops-backend-service
            port:
              number: 80
      # Frontend - Root path (must be last)
      - path: /
        pathType: Prefix
        backend:
          service:
            name: devops-frontend-service
            port:
              number: 80
```

**IMPORTANT NOTES:**
1. Do **NOT** create a separate Ingress for the frontend
2. Path order is important: `/api` and `/ws` first, `/` last
3. The same TLS certificate will be used for both services
4. This configuration ensures same-origin (no CORS issues)

### 8. Update .gitignore

Add to `.gitignore` file in your frontend repository:

```
# Kubernetes secrets (never commit)
k8s/secret.yaml
k8s/*-secret.yaml
```

### 9. Create/Update GitLab CI/CD Pipeline

Create or update `.gitlab-ci.yml` in your frontend repository:

```yaml
stages:
  - test
  - build-image
  - deploy

variables:
  IMAGE_NAME: $CI_REGISTRY_IMAGE
  IMAGE_TAG: $CI_COMMIT_SHORT_SHA

# Test stage (if you have tests)
test:
  stage: test
  image: node:18
  script:
    - npm ci
    - npm run test:unit
  only:
    - main
    - merge_requests

# Lint check (optional)
lint:
  stage: test
  image: node:18
  script:
    - npm ci
    - npm run lint
  only:
    - main
    - merge_requests

# Docker image build and push
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

# Kubernetes deployment
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
    - echo "Waiting for deployment to be ready..."
    - kubectl rollout status deployment/devops-frontend -n emzops-staging --timeout=5m
    - echo "✅ Deployment successful!"
    - echo "Frontend URL: https://devops-mezdit.kubernetes.devops.cs.ut.ee"
    - kubectl get pods -n emzops-staging -l app=devops-frontend
  environment:
    name: staging
    url: https://devops-mezdit.kubernetes.devops.cs.ut.ee
  only:
    - main
```

### 10. Add GitLab CI/CD Variable

In your frontend GitLab repository:

1. **Settings → CI/CD → Variables**
2. Click **Add variable**
3. Enter the following values:

**Variable 1:**
- **Key**: `KUBECONFIG_CONTENT`
- **Value**: Content of `KUBECONFIG_BASE64.txt` file from backend repository
- **Type**: Variable
- **Protected**: ✅ Yes
- **Masked**: ❌ No
- **Expand variable reference**: ❌ No

### 11. Configure Vue.js Environment Variables

Get API URL from environment variable in your Vue.js application:

**Example usage (src/api/axios.js or similar):**

```javascript
import axios from 'axios';

// Use relative path since it's same origin
const baseURL = process.env.VUE_APP_API_URL || '/api';

const apiClient = axios.create({
  baseURL: baseURL,
  withCredentials: true,  // Required for cookies
  headers: {
    'Content-Type': 'application/json',
  }
});

export default apiClient;
```

**For WebSocket (if you're using it):**

```javascript
// Use relative path since it's same origin
const wsURL = process.env.VUE_APP_WS_URL || '/ws';

// Convert relative path to absolute URL
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const fullWsURL = `${protocol}//${window.location.host}${wsURL}`;

const socket = new WebSocket(fullWsURL);
```

### 12. Backend CORS Configuration

**IMPORTANT:** No CORS configuration needed since it's same-origin! However, you may need settings for local development:

```go
// Go backend example
router.Use(cors.New(cors.Config{
    AllowOrigins: []string{
        "https://devops-mezdit.kubernetes.devops.cs.ut.ee",  // Production (same-origin, no CORS issue)
        "http://localhost:8080",      // For local development
        "http://localhost:3000",      // For local Vue.js dev server
    },
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}))
```

**Note:** In production, frontend and backend are on the same domain, so CORS is unnecessary, only needed for local development.

### 13. Deployment

Everything is ready! Commit and push:

```bash
git add .
git commit -m "Add Kubernetes deployment for frontend"
git push origin main
```

GitLab pipeline will automatically:
1. ✅ Run tests
2. ✅ Build Docker image
3. ✅ Push to GitLab Registry
4. ✅ Deploy to Kubernetes

### 14. Verify Deployment

```bash
# Check pods
kubectl get pods -n emzops-staging -l app=devops-frontend

# Check logs
kubectl logs -f deployment/devops-frontend -n emzops-staging

# Check service
kubectl get service devops-frontend-service -n emzops-staging

# Check ingress
kubectl get ingress devops-ingress -n emzops-staging

# Check TLS certificate
kubectl get certificate devops-tls-cert -n emzops-staging

# Local test with port forward
kubectl port-forward service/devops-frontend-service 3000:80 -n emzops-staging
# In browser: http://localhost:3000
```

### 15. Test in Browser

1. Go to https://devops-mezdit.kubernetes.devops.cs.ut.ee (now a single URL!)
2. Check that TLS certificate is working (green lock icon)
3. Check that backend API calls are working (e.g., https://devops-mezdit.kubernetes.devops.cs.ut.ee/api/health)
4. Check that there are no CORS errors in browser console (there shouldn't be since it's same-origin)

---

## 🔧 Troubleshooting

### Problem 1: Image Pull Error

**Symptom:** `ImagePullBackOff` or `ErrImagePull`

**Solution:**
```bash
# Check registry secret
kubectl get secret gitlab-registry-secret -n emzops-staging

# If it doesn't exist, create it like in backend (should already exist)
```

### Problem 2: Pod Not Starting

```bash
# Check pod details
kubectl describe pod <pod-name> -n emzops-staging

# Check logs
kubectl logs <pod-name> -n emzops-staging
```

### Problem 3: CORS Error

**There should be NO CORS errors since we're using same-origin!** If you see them:

1. Check that frontend is running on the correct URL (https://devops-mezdit.kubernetes.devops.cs.ut.ee)
2. Check that API requests use relative paths (`/api/...`)
3. If using different ports in local development, check backend CORS settings

### Problem 4: TLS Certificate Not Creating

```bash
# Check certificate status (using the same certificate as backend)
kubectl describe certificate devops-tls-cert -n emzops-staging

# Wait 1-2 minutes, it will be created automatically
```

### Problem 5: API Requests Returning 404

1. Check that backend is running:
   ```bash
   kubectl get pods -n emzops-staging -l app=devops-backend
   ```

2. Check that API URL is correct:
   ```bash
   kubectl get configmap frontend-config -n emzops-staging -o yaml
   ```

3. Check that frontend code uses the correct API URL

---

## 📝 Important Notes

1. **Environment Variables:**
   - Environment variables in Vue.js must start with `VUE_APP_` prefix
   - Embedded at build time, cannot be changed at runtime
   - Rebuild required to change

2. **NGINX Configuration:**
   - `try_files` is important for Vue.js SPA routing
   - API proxy configuration required for backend communication

3. **CORS:**
   - No CORS issues since we're using same-origin
   - Backend CORS settings may be needed for local development
   - If using credentials, `withCredentials: true` must be set

4. **Secrets:**
   - Use Kubernetes Secrets for sensitive information (API keys, etc.)
   - Never commit to git

5. **Testing:**
   - Test locally before pushing to production:
     ```bash
     docker build -t frontend-test .
     docker run -p 3000:80 frontend-test
     ```

---

## 🎯 Checklist

Pre-deployment checklist:

- [ ] Dockerfile created
- [ ] nginx.conf created
- [ ] k8s/ directory and manifest files created
- [ ] .gitlab-ci.yml updated
- [ ] GitLab CI/CD variable (`KUBECONFIG_CONTENT`) added
- [ ] Vue.js code uses environment variables
- [ ] Backend CORS settings configured
- [ ] .gitignore updated
- [ ] Local test done (docker build)
- [ ] Committed and pushed to git
- [ ] Pipeline succeeded
- [ ] Pods are running
- [ ] TLS certificate created
- [ ] Browser access successful
- [ ] API calls working
- [ ] No CORS errors

---

**For questions:** Refer to backend KUBERNETES_SETUP.md file or ask your team members.

**Good luck!** 🚀
