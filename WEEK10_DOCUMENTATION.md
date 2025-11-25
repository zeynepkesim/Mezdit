# Week 10 - More Kubernetes - Complete Documentation

**Team:** emzops  
**Namespace:** emzops-staging  
**Production URL:** https://devops-mezdit.kubernetes.devops.cs.ut.ee

---

## Table of Contents

1. [Overview](#overview)
2. [Kubernetes Architecture](#kubernetes-architecture)
3. [Services Implementation](#services-implementation)
4. [Ingress and Traffic Routing](#ingress-and-traffic-routing)
5. [TLS Certificates (cert-manager)](#tls-certificates-cert-manager)
6. [Database Deployment](#database-deployment)
7. [Configuration Management](#configuration-management)
8. [GitLab-Kubernetes Integration (RBAC)](#gitlab-kubernetes-integration-rbac)
9. [CI/CD Pipeline Deployment](#cicd-pipeline-deployment)
10. [Testing and Verification](#testing-and-verification)

---

## Overview

Week 10 focuses on production-ready Kubernetes deployments with:
- ✅ Proper Service configuration for internal communication
- ✅ Ingress for external access with path-based routing
- ✅ Automatic TLS certificate management
- ✅ Separate database deployment
- ✅ Externalized configuration (ConfigMaps + Secrets)
- ✅ RBAC for CI/CD automation
- ✅ Fully automated deployment pipeline

### What Changed from Week 9

| Aspect | Week 9 | Week 10 |
|--------|--------|---------|
| Access | Port forwarding | Ingress (public HTTPS) |
| Database | (could be embedded) | Separate StatefulSet |
| Configuration | (could be hardcoded) | ConfigMaps + Secrets |
| Deployment | Manual `kubectl apply` | Automated via CI/CD |
| TLS | N/A | Automatic with cert-manager |

---

## Kubernetes Architecture

### High-Level Architecture Diagram

```
                                 INTERNET
                                    ↓
                    [Let's Encrypt] (TLS Cert)
                                    ↓
                          ┌─────────────────┐
                          │  Ingress (nginx)│
                          │  TLS Termination│
                          └────────┬────────┘
                                   │
                    ┌──────────────┼──────────────┐
                    │              │              │
              Path: /api/*    Path: /ws     Path: /*
                    │              │              │
                    ↓              ↓              ↓
         ┌──────────────────┐           ┌──────────────────┐
         │ Backend Service  │           │Frontend Service  │
         │   (ClusterIP)    │           │   (ClusterIP)    │
         └────────┬─────────┘           └────────┬─────────┘
                  │                              │
                  ↓                              ↓
      ┌──────────────────────┐       ┌──────────────────────┐
      │ Backend Deployment   │       │Frontend Deployment   │
      │   (1 replica)        │       │   (1 replica)        │
      │                      │       │                      │
      │ ┌──────────────────┐ │       │ ┌──────────────────┐ │
      │ │  Go App Pod      │ │       │ │ Nginx+Vue Pod    │ │
      │ │  Port: 8080      │ │       │ │ Port: 80         │ │
      │ └────────┬─────────┘ │       │ └──────────────────┘ │
      └──────────┼───────────┘       └──────────────────────┘
                 │
                 │ Database connection
                 ↓
      ┌────────────────────┐
      │ Postgres Service   │
      │ (Headless ClusterIP)│
      └──────────┬─────────┘
                 │
                 ↓
      ┌──────────────────────┐
      │ PostgreSQL StatefulSet│
      │   (1 replica)         │
      │                       │
      │ ┌───────────────────┐ │
      │ │ postgres-0        │ │
      │ │ Port: 5432        │ │
      │ │ PVC: 1Gi          │ │
      │ └───────────────────┘ │
      └───────────────────────┘

              ┌────────────────┐
              │  ConfigMaps    │
              │  - backend     │
              │  - frontend    │
              └────────────────┘

              ┌────────────────┐
              │    Secrets     │
              │  - database    │
              │  - registry    │
              │  - TLS cert    │
              └────────────────┘
```

### Traffic Flow

**External User → Application:**

```
1. User → https://devops-mezdit.kubernetes.devops.cs.ut.ee/
   ↓
2. Ingress Controller (nginx)
   - TLS termination (cert from Let's Encrypt)
   - Path routing based on rules
   ↓
3a. Path /api/* → backend-service:80 → backend pod:8080
3b. Path /ws    → backend-service:80 → backend pod:8080
3c. Path /*     → frontend-service:80 → frontend pod:80
   ↓
4. Response back through same path
```

**Backend → Database:**

```
Backend Pod
  ↓
DATABASE_URL: postgresql://user:pass@postgres-service:5432/devops
  ↓
postgres-service (Headless ClusterIP)
  ↓
postgres-0 StatefulSet pod
  ↓
PostgreSQL database on PVC
```

### Resource Summary

**Namespace:** `emzops-staging`

| Resource Type | Name | Purpose |
|---------------|------|---------|
| **Deployments** |
| | devops-backend | Go application (1 replica) |
| | devops-frontend | Vue.js + Nginx (1 replica) |
| **StatefulSets** |
| | postgres | PostgreSQL database (1 replica) |
| **Services** |
| | devops-backend-service | ClusterIP for backend |
| | devops-frontend-service | ClusterIP for frontend |
| | postgres-service | Headless ClusterIP for database |
| **Ingress** |
| | devops-backend-ingress | Path-based routing + TLS |
| **ConfigMaps** |
| | backend-config | PORT=8080 |
| | frontend-config | API URLs |
| **Secrets** |
| | backend-secret | DATABASE_URL |
| | postgres-superuser-secret | DB credentials |
| | gitlab-registry-secret | Image pull auth (backend) |
| | frontend-registry-secret | Image pull auth (frontend) |
| | backend-tls-cert | TLS certificate (auto-managed) |
| **RBAC** |
| | gitlab-deploy (ServiceAccount) | CI/CD identity |
| | gitlab-deploy-role (Role) | Permissions |
| | gitlab-deploy-binding (RoleBinding) | Binding |

---

## Services Implementation

### What is a Kubernetes Service?

A Service provides **stable networking** for Pods:
- Pods have ephemeral IP addresses (change on restart)
- Services have stable ClusterIP addresses
- Services use label selectors to find Pods

### Service Types Used

#### 1. ClusterIP (Backend & Frontend)

**Why ClusterIP?**
- ✅ Internal cluster communication
- ✅ Ingress can route to it
- ✅ Secure (not exposed externally)
- ✅ Simple and recommended for most apps

**Backend Service (`k8s/service.yaml`):**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: devops-backend-service
  namespace: emzops-staging
  annotations:
    # Prometheus annotations (Week 11)
    prometheus.io/scrape: "true"
    prometheus.io/port: "9090"
    prometheus.io/path: "/metrics"
spec:
  type: ClusterIP  # Internal only
  selector:
    app: devops-backend  # Matches pod labels
  ports:
  - name: http
    protocol: TCP
    port: 80           # Service port (what others connect to)
    targetPort: 8080   # Container port (where app listens)
  - name: metrics
    protocol: TCP
    port: 9090         # Metrics port for Prometheus scraping
    targetPort: 9090
```

**How it works:**
```
# API requests:
Other pods → devops-backend-service:80 → backend pod:8080

# Prometheus scraping:
Prometheus → devops-backend-service:9090 → backend pod:9090
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

#### 2. Headless ClusterIP (Database)

**Why Headless?**
- ✅ StatefulSets need stable network identities
- ✅ Direct pod-to-pod communication
- ✅ Each pod gets its own DNS: `postgres-0.postgres-service`

**PostgreSQL Service:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: emzops-staging
spec:
  clusterIP: None  # Headless (no ClusterIP assigned)
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
```

**DNS Resolution:**
```bash
# Direct pod access:
postgres-0.postgres-service.emzops-staging.svc.cluster.local

# Service name also works:
postgres-service.emzops-staging.svc.cluster.local
```

### Service Discovery

**Within same namespace:**
```
http://devops-backend-service:80
postgresql://postgres-service:5432
```

**From another namespace:**
```
http://devops-backend-service.emzops-staging.svc.cluster.local:80
```

### Testing Services

```bash
# Port forward to service (not pod):
kubectl port-forward service/devops-backend-service -n emzops-staging 8080:80

# Test:
curl http://localhost:8080
```

---

## Ingress and Traffic Routing

### What is Ingress?

Ingress is an **intelligent HTTP(S) Load Balancer** that:
- Exposes HTTP/HTTPS routes from outside → Services inside cluster
- Provides URL-based routing
- Handles TLS termination
- Maps domains to Services

**NOT an Ingress Controller** (already installed by cluster admins)

### Our Ingress Configuration

**File:** `k8s/ingress.yaml`

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: devops-backend-ingress
  namespace: emzops-staging
  annotations:
    # TLS certificate automation
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    # Force HTTPS
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    # WebSocket support - critical for WebSocket upgrade
    nginx.ingress.kubernetes.io/websocket-services: "devops-backend-service"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
    # Rewrite paths for backend API
    nginx.ingress.kubernetes.io/rewrite-target: /$2
    nginx.ingress.kubernetes.io/use-regex: "true"
spec:
  ingressClassName: nginx  # Use centrally-installed nginx controller
  tls:
  - hosts:
    - devops-mezdit.kubernetes.devops.cs.ut.ee
    secretName: backend-tls-cert  # cert-manager creates this
  rules:
  - host: devops-mezdit.kubernetes.devops.cs.ut.ee
    http:
      paths:
      # Backend API (MUST be first - most specific)
      # Regex: /api/posts -> /posts, /api/login -> /login
      - path: /api(/|$)(.*)
        pathType: ImplementationSpecific
        backend:
          service:
            name: devops-backend-service
            port:
              number: 80
      
      # WebSocket - preserve /ws path
      # Pattern captures "ws" in $2 so rewrite keeps it: /ws -> /ws
      - path: /()(ws.*)
        pathType: ImplementationSpecific
        backend:
          service:
            name: devops-backend-service
            port:
              number: 80
      
      # Frontend (MUST be last - catch-all)
      # Regex captures full path so rewrite-target /$2 works correctly
      - path: /()(.*)
        pathType: ImplementationSpecific
        backend:
          service:
            name: devops-frontend-service
            port:
              number: 80
```

### Path Routing Logic

**Order matters!** Most specific paths first:

| Request | Matches | Routed To | Rewritten To |
|---------|---------|-----------|--------------|
| `/api/posts` | `/api(/\|$)(.*)` | backend-service | `/posts` |
| `/api/login` | `/api(/\|$)(.*)` | backend-service | `/login` |
| `/ws` | `/()(ws.*)` | backend-service | `/ws` (preserved) |
| `/` | `/()(.*)` | frontend-service | `/` |
| `/about` | `/()(.*)` | frontend-service | `/about` |

### WebSocket Support

**Challenge:** WebSocket connections require special handling:
- HTTP Upgrade headers must be preserved
- Long-lived connections need extended timeouts
- Path rewriting must not break WebSocket handshake

**Annotations for WebSocket:**

```yaml
annotations:
  # Tell nginx which services handle WebSocket
  nginx.ingress.kubernetes.io/websocket-services: "devops-backend-service"
  # Extend timeouts for long-lived WebSocket connections
  nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"  # 1 hour
  nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"  # 1 hour
```

**WebSocket Path Pattern:**

```yaml
# Pattern: /()(ws.*)
# This captures "ws" in $2, so rewrite-target: /$2 preserves the path
- path: /()(ws.*)
  pathType: ImplementationSpecific
```

**Why this pattern?**

```
Request:  /ws
Matches:  /  (empty)  ws
          ─  ───────  ──
          literal $1  $2
Rewrite:  /$2 = /ws  ✅ Path preserved!

Without the pattern (old approach):
Path: /ws
Rewrite would conflict with global rewrite-target
Result: WebSocket connection fails ❌
```

### API Path Rewriting Explained

**Problem:** Backend app expects `/posts`, but external URL is `/api/posts`

**Solution:** `rewrite-target: /$2`

**Regex breakdown:**
```
Path pattern: /api(/|$)(.*)
              ─┬─ ───┬── ─┬─
               │     │    └─ $2 = everything after /api/
               │     └────── $1 = slash or end of string  
               └──────────── literal "/api"

Example:
Request:  /api/posts?id=123
Matches:  /api  /  posts?id=123
          ───── ─  ──────────
                $1     $2
Rewrite:  /$2 = /posts?id=123
```

### Hostname Requirement

**MUST use format:** `<your-name>.kubernetes.devops.cs.ut.ee`

**Why?**
- Central Ingress controller only handles this domain
- Wildcard DNS already configured: `*.kubernetes.devops.cs.ut.ee`
- cert-manager configured for this issuer

**Our hostname:** `devops-mezdit.kubernetes.devops.cs.ut.ee`

### Ingress vs Service vs Pod

```
External Client
  ↓
[Ingress] → Routes based on hostname + path
  ↓
[Service] → Routes to healthy pods (load balancing)
  ↓
[Pod] → Runs the actual container
```

---

## TLS Certificates (cert-manager)

### What is cert-manager?

**cert-manager** is a Kubernetes Operator that:
- ✅ Automatically provisions TLS certificates
- ✅ Renews certificates before expiration
- ✅ Integrates with Let's Encrypt (free certificates)
- ✅ Stores certificates in Kubernetes Secrets

**Already installed by cluster admins** - we just use it!

### How It Works

```
1. You create Ingress with annotation:
   cert-manager.io/cluster-issuer: "letsencrypt-prod"

2. cert-manager detects new Ingress
   ↓
3. Requests certificate from Let's Encrypt
   ↓
4. Let's Encrypt validates domain ownership (HTTP-01 challenge)
   ↓
5. Let's Encrypt issues certificate
   ↓
6. cert-manager creates Secret with certificate
   ↓
7. Ingress uses Secret for TLS termination
   ↓
8. HTTPS works! 🎉

9. Before expiry (90 days), cert-manager auto-renews
```

### Our Configuration

**Ingress annotation:**
```yaml
metadata:
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
```

**TLS configuration:**
```yaml
spec:
  tls:
  - hosts:
    - devops-mezdit.kubernetes.devops.cs.ut.ee
    secretName: backend-tls-cert  # cert-manager creates this
```

### Verifying TLS Certificate

**Check secret exists:**
```bash
kubectl get secret backend-tls-cert -n emzops-staging

# Output:
NAME               TYPE                DATA   AGE
backend-tls-cert   kubernetes.io/tls   2      10d
```

**Check certificate details:**
```bash
kubectl describe secret backend-tls-cert -n emzops-staging
```

**Test HTTPS:**
```bash
curl -I https://devops-mezdit.kubernetes.devops.cs.ut.ee

# Look for:
HTTP/2 200
strict-transport-security: max-age=31536000
```

### Force HTTPS Redirect

**Annotation:**
```yaml
nginx.ingress.kubernetes.io/ssl-redirect: "true"
```

**What it does:**
```
HTTP request  → 301 Redirect → HTTPS request
http://...                     https://...
```

**Test:**
```bash
curl -I http://devops-mezdit.kubernetes.devops.cs.ut.ee
# Should return: 308 Permanent Redirect
# Location: https://devops-mezdit.kubernetes.devops.cs.ut.ee
```

---

## Database Deployment

### Requirement

Week 10 requires:
- ✅ Database in separate pod (not in app container)
- ✅ One replica minimum
- ✅ Use emptyDir (we use PVC - better!)

### Implementation: StatefulSet with PVC

**Why StatefulSet (not Deployment)?**
- ✅ Stable network identity: `postgres-0`
- ✅ Persistent storage with PVCs
- ✅ Ordered pod creation/deletion
- ✅ Database-appropriate workload type

**File:** `k8s/postgres-statefulset.yaml`

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: emzops-staging
spec:
  serviceName: postgres-service  # Headless service
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15-alpine
        ports:
        - containerPort: 5432
        env:
        # Database name (non-sensitive)
        - name: POSTGRES_DB
          value: "devops"
        # Credentials from Secret
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: postgres-superuser-secret
              key: username
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-superuser-secret
              key: password
        # Data directory
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
  
  # PersistentVolumeClaim template
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 1Gi
```

### Storage: PVC vs emptyDir

**Week 10 suggests emptyDir:**
```yaml
volumes:
- name: db-storage
  emptyDir:
    sizeLimit: 400Mi
```

**We use PVC (better):**
```yaml
volumeClaimTemplates:
- metadata:
    name: postgres-storage
  spec:
    resources:
      requests:
        storage: 1Gi
```

**Comparison:**

| Aspect | emptyDir | PVC |
|--------|----------|-----|
| Persistence | Lost on pod delete | Survives pod deletion |
| Use case | Temporary data | Production databases |
| Week 10 requirement | ✅ Meets | ✅ Exceeds |

### Alternative: CloudNativePG

We also have `postgres-cluster.yaml` for CNPG operator:

```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: postgres-cluster
  namespace: emzops-staging
spec:
  instances: 1  # Can scale to 2-3 for HA
  storage:
    size: 400Mi
  bootstrap:
    initdb:
      database: devops
      owner: devops
      secret:
        name: postgres-superuser-secret
```

**Benefits of CNPG:**
- ✅ Automated backups
- ✅ High availability (multi-instance)
- ✅ Automated failover
- ✅ Connection pooling

**Current setup:** Using StatefulSet (simpler, works great for Week 10)

### Connecting to Database

**From backend pod:**

```go
// Environment variable from Secret:
DATABASE_URL=postgresql://devops:PASSWORD@postgres-service:5432/devops

// Connection in Go:
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
```

**Service DNS resolution:**
```
postgres-service → postgres-service.emzops-staging.svc.cluster.local → postgres-0
```

### Database Secret

**File:** `k8s/postgres-secret.yaml.template` (actual secret NOT in Git)

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: postgres-superuser-secret
  namespace: emzops-staging
type: kubernetes.io/basic-auth
stringData:
  username: devops
  password: <strong-password-here>
```

**Created once, used by:**
- PostgreSQL StatefulSet (sets `POSTGRES_USER`, `POSTGRES_PASSWORD`)
- Backend Secret (constructs `DATABASE_URL`)

---

## Configuration Management

### Principle: 12-Factor App

**Configuration should be:**
- ✅ Externalized (not in code or image)
- ✅ Environment-specific
- ✅ Easy to change without rebuilding

### ConfigMaps: Non-Sensitive Configuration

**Backend ConfigMap (`k8s/configmap.yaml`):**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: backend-config
  namespace: emzops-staging
data:
  PORT: "8080"
  # Add more non-sensitive config here
  LOG_LEVEL: "info"
  ENVIRONMENT: "staging"
```

**Frontend ConfigMap:**

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

**Using in Deployment:**

```yaml
spec:
  containers:
  - name: backend
    env:
    # Single value
    - name: PORT
      valueFrom:
        configMapKeyRef:
          name: backend-config
          key: PORT
    
    # All values from ConfigMap
    envFrom:
    - configMapRef:
        name: backend-config
```

### Secrets: Sensitive Data

**Backend Secret (`k8s/secret.yaml.template`):**

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

**⚠️ Important:**
- `.gitignore` includes `secret.yaml`
- Only `secret.yaml.template` is in Git
- Actual secrets created manually or via CI/CD

**Using in Deployment:**

```yaml
spec:
  containers:
  - name: backend
    env:
    - name: DATABASE_URL
      valueFrom:
        secretKeyRef:
          name: backend-secret
          key: DATABASE_URL
```

### Registry Secrets (Image Pull)

**For private GitLab registry:**

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
          "username": "<deploy-token>",
          "password": "<token>"
        }
      }
    }
```

**Referenced in Deployment:**

```yaml
spec:
  template:
    spec:
      imagePullSecrets:
        - name: gitlab-registry-secret
```

### Configuration Workflow

**1. Development:**
```bash
# Create from template
cp k8s/secret.yaml.template k8s/secret.yaml

# Edit with real values
vim k8s/secret.yaml
```

**2. Apply to cluster:**
```bash
# ConfigMap (safe to commit)
kubectl apply -f k8s/configmap.yaml

# Secret (never commit!)
kubectl apply -f k8s/secret.yaml
```

**3. CI/CD Pipeline:**
```yaml
# ConfigMap from file
kubectl apply -f k8s/configmap.yaml

# Secret from GitLab variable
echo "$DATABASE_URL" | kubectl create secret generic backend-secret \
  --from-literal=DATABASE_URL="$DATABASE_URL" \
  --dry-run=client -o yaml | kubectl apply -f -
```

### Updating Configuration

**ConfigMap changes:**
```bash
# Edit ConfigMap
kubectl edit configmap backend-config -n emzops-staging

# Restart pods to pick up changes
kubectl rollout restart deployment/devops-backend -n emzops-staging
```

**Secret changes:**
```bash
# Update secret
kubectl create secret generic backend-secret \
  --from-literal=DATABASE_URL="new-value" \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart pods
kubectl rollout restart deployment/devops-backend -n emzops-staging
```

### What Goes Where?

| Data Type | ConfigMap | Secret |
|-----------|-----------|--------|
| Port numbers | ✅ | |
| API URLs | ✅ | |
| Log levels | ✅ | |
| Environment names | ✅ | |
| Database passwords | | ✅ |
| API keys | | ✅ |
| Connection strings | | ✅ |
| TLS certificates | | ✅ |
| Registry credentials | | ✅ |

**Rule of thumb:** If you wouldn't commit it to Git, it goes in a Secret!

---

## NetworkPolicy for Observability

### Purpose

NetworkPolicy is a Kubernetes resource that controls network traffic between pods using label selectors. For Week 11's observability requirements, we need to allow Prometheus to scrape metrics from our backend while maintaining security.

### Why NetworkPolicy?

**Security Benefits:**
- ✅ Restrict which namespaces can access your pods
- ✅ Limit exposure of metrics endpoints
- ✅ Prevent unauthorized pod-to-pod communication
- ✅ Implement zero-trust networking

### Implementation

**File:** `devops-be/k8s/networkpolicy.yaml`

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-prometheus-scraping
  namespace: emzops-staging
spec:
  podSelector:
    matchLabels:
      app: devops-backend  # Apply to backend pods only
  policyTypes:
  - Ingress  # Control incoming traffic
  ingress:
  # Rule 1: Allow Prometheus to scrape metrics (port 9090)
  - from:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: prometheus-stack
    ports:
    - protocol: TCP
      port: 9090
  
  # Rule 2: Allow ingress controller traffic (API requests on port 8080)
  - from:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  
  # Rule 3: Allow traffic from same namespace (frontend, postgres, etc.)
  - from:
    - podSelector: {}  # Empty selector = all pods in namespace
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 9090
```

### Traffic Flow with NetworkPolicy

```
┌─────────────────────────────────────────────────────┐
│  Prometheus (prometheus-stack namespace)            │
│    ↓                                                 │
│  NetworkPolicy Check                                │
│    ├─ Source namespace: prometheus-stack ✅         │
│    ├─ Target port: 9090 ✅                          │
│    └─ ALLOW                                         │
│    ↓                                                 │
│  Backend Pod (emzops-staging namespace)             │
│    Receives metrics scrape on port 9090             │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│  External User → Ingress Controller                 │
│    ↓                                                 │
│  NetworkPolicy Check                                │
│    ├─ Source namespace: ingress-nginx ✅            │
│    ├─ Target port: 8080 ✅                          │
│    └─ ALLOW                                         │
│    ↓                                                 │
│  Backend Pod receives API request                   │
└─────────────────────────────────────────────────────┘
```

### Key Rules Explained

| Rule | Source | Port | Purpose |
|------|--------|------|---------|
| 1 | `prometheus-stack` namespace | 9090 | Prometheus metrics scraping |
| 2 | `ingress-nginx` namespace | 8080 | Public API access via Ingress |
| 3 | Same namespace (any pod) | 8080, 9090 | Internal communication |

### Deployment

```bash
# Apply NetworkPolicy
kubectl apply -f k8s/networkpolicy.yaml

# Verify it exists
kubectl get networkpolicy -n emzops-staging

# Describe the policy
kubectl describe networkpolicy allow-prometheus-scraping -n emzops-staging
```

### Verification

```bash
# Check if Prometheus can reach metrics endpoint
# Visit: https://monitoring.kubernetes.devops.cs.ut.ee
# Navigate to: Status → Targets
# Look for: emzops-staging/devops-backend-service endpoints

# Test metrics endpoint via port-forward (bypasses NetworkPolicy)
kubectl port-forward service/devops-backend-service -n emzops-staging 9090:9090
curl http://localhost:9090/metrics
```

### What Happens Without NetworkPolicy?

**Without NetworkPolicy:**
- ❌ Any pod from any namespace can access your backend
- ❌ Metrics endpoint exposed to entire cluster
- ❌ No network segmentation

**With NetworkPolicy:**
- ✅ Only specific namespaces can access specific ports
- ✅ Metrics only accessible to Prometheus
- ✅ Defense-in-depth security

---

## GitLab-Kubernetes Integration (RBAC)

### The Problem

GitLab CI/CD needs to deploy resources to Kubernetes, but:
- ❌ Can't use your personal credentials (security risk)
- ❌ Can't use cluster admin (too many permissions)

### The Solution: RBAC

**RBAC = Role-Based Access Control**

Creates a dedicated identity for CI/CD with minimal required permissions.

### Components

**File:** `k8s/rbac.yaml`

#### 1. ServiceAccount

**What:** Identity for CI/CD pipeline

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gitlab-deploy
  namespace: emzops-staging
```

**Think of it as:** A "robot user" for GitLab

#### 2. Secret (Service Account Token)

**What:** Authentication token for ServiceAccount

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gitlab-deploy-token
  namespace: emzops-staging
  annotations:
    kubernetes.io/service-account.name: gitlab-deploy
type: kubernetes.io/service-account-token
```

**Contains:**
- `token` - Bearer token for authentication
- `ca.crt` - Cluster CA certificate

#### 3. Role

**What:** Defines what actions are allowed

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: gitlab-deploy-role
  namespace: emzops-staging
rules:
# Allow managing core resources
- apiGroups: [""]
  resources: ["services", "configmaps", "secrets", "pods", "pods/log"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

# Allow managing apps
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

# Allow managing ingress
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

**Permissions explained:**
- `get, list, watch` - Read permissions
- `create` - Create new resources
- `update, patch` - Modify existing resources
- `delete` - Remove resources

**Principle of Least Privilege:**
- ✅ Only resources needed for deployment
- ✅ Only namespace `emzops-staging`
- ❌ No cluster-wide permissions
- ❌ No RBAC modification permissions

#### 4. RoleBinding

**What:** Connects ServiceAccount to Role

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: gitlab-deploy-binding
  namespace: emzops-staging
subjects:
- kind: ServiceAccount
  name: gitlab-deploy
  namespace: emzops-staging
roleRef:
  kind: Role
  name: gitlab-deploy-role
  apiGroup: rbac.authorization.k8s.io
```

**Think of it as:** "gitlab-deploy ServiceAccount has gitlab-deploy-role permissions"

### RBAC Flow

```
GitLab CI/CD Pipeline
  ↓
Authenticates with token from gitlab-deploy-token Secret
  ↓
Acts as gitlab-deploy ServiceAccount
  ↓
RoleBinding grants gitlab-deploy-role permissions
  ↓
Role allows: create/update/delete deployments, services, etc.
  ↓
Pipeline can deploy resources to emzops-staging namespace
```

### Creating RBAC Resources

**1. Apply RBAC manifests:**
```bash
kubectl apply -f k8s/rbac.yaml
```

**2. Verify creation:**
```bash
kubectl get serviceaccount gitlab-deploy -n emzops-staging
kubectl get role gitlab-deploy-role -n emzops-staging
kubectl get rolebinding gitlab-deploy-binding -n emzops-staging
kubectl get secret gitlab-deploy-token -n emzops-staging
```

**3. Extract token and CA cert:**
```bash
NAMESPACE=emzops-staging
SERVICE_ACCOUNT_NAME=gitlab-deploy
SECRET_NAME=gitlab-deploy-token

TOKEN=$(kubectl get secret $SECRET_NAME -n $NAMESPACE -o jsonpath='{.data.token}' | base64 --decode)
CA_DATA=$(kubectl get secret $SECRET_NAME -n $NAMESPACE -o jsonpath='{.data.ca\.crt}')
API_SERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
```

**4. Generate kubeconfig:**
```bash
cat <<EOF > gitlab-deploy-kubeconfig
apiVersion: v1
kind: Config
clusters:
- name: kubernetes-cluster
  cluster:
    certificate-authority-data: $CA_DATA
    server: $API_SERVER
contexts:
- name: gitlab-deploy-context
  context:
    cluster: kubernetes-cluster
    namespace: $NAMESPACE
    user: $SERVICE_ACCOUNT_NAME
current-context: gitlab-deploy-context
users:
- name: $SERVICE_ACCOUNT_NAME
  user:
    token: $TOKEN
EOF
```

**5. Encode kubeconfig for GitLab:**
```bash
cat gitlab-deploy-kubeconfig | base64 | tr -d '\n' > kubeconfig-base64.txt
```

**6. Add to GitLab CI/CD variables:**
```
Settings → CI/CD → Variables → Add Variable

Key: KUBECONFIG_CONTENT
Value: <paste from kubeconfig-base64.txt>
Protected: Yes
Masked: No (too large)
```

### Using in Pipeline

**`.gitlab-ci.yml`:**
```yaml
deploy-to-k8s:
  stage: deploy
  image: bitnami/kubectl:latest
  before_script:
    - mkdir -p ~/.kube
    - echo "$KUBECONFIG_CONTENT" | base64 -d > ~/.kube/config
    - chmod 600 ~/.kube/config
  script:
    - kubectl apply -f k8s/deployment.yaml
```

---

## CI/CD Pipeline Deployment

### Pipeline Architecture

```
Code Push to main branch
  ↓
┌─────────────────────────────────────┐
│  GitLab CI/CD Pipeline              │
│                                     │
│  Stage 1: Security                  │
│  - gitleaks (secret scan)           │
│                                     │
│  Stage 2: Test                      │
│  - go test (backend)                │
│  - npm ci + build (frontend)        │
│                                     │
│  Stage 3: Build                     │
│  - go build (backend)               │
│                                     │
│  Stage 4: Build Image               │
│  - Kaniko: build container          │
│  - Push to GitLab Registry          │
│                                     │
│  Stage 5: Deploy                    │
│  - kubectl apply manifests          │
│  - Wait for rollout                 │
└─────────────┬───────────────────────┘
              ↓
      Kubernetes Cluster
      (emzops-staging namespace)
```

### Backend Pipeline

**File:** `devops-be/.gitlab-ci.yml`

**Deploy stage (Week 10 focus):**

```yaml
deploy-to-k8s:
  stage: deploy
  image: bitnami/kubectl:latest
  
  before_script:
    # Setup kubectl with service account credentials
    - mkdir -p ~/.kube
    - echo "$KUBECONFIG_CONTENT" | base64 -d > ~/.kube/config
    - chmod 600 ~/.kube/config
  
  script:
    - echo "Deploying to Kubernetes namespace emzops-staging..."
    
    # Apply ConfigMap (non-sensitive config)
    - kubectl apply -f k8s/configmap.yaml
    
    # Apply PostgreSQL StatefulSet
    - kubectl apply -f k8s/postgres-statefulset.yaml
    
    # Apply backend deployment
    - kubectl apply -f k8s/deployment.yaml
    
    # Apply backend service
    - kubectl apply -f k8s/service.yaml
    
    # Apply ingress (NEW in this update)
    - kubectl apply -f k8s/ingress.yaml
    
    # Wait for deployment to complete
    - echo "Waiting for deployment to be ready..."
    - kubectl rollout status deployment/devops-backend -n emzops-staging --timeout=5m
    
    # Show pod status
    - echo "✅ Deployment successful!"
    - kubectl get pods -n emzops-staging
  
  environment:
    name: staging
    url: https://devops-mezdit.kubernetes.devops.cs.ut.ee
  
  only:
    - main  # Only deploy from main branch
```

### Frontend Pipeline

**File:** `devops-fe/.gitlab-ci.yml`

**Deploy stage:**

```yaml
deploy-to-k8s:
  stage: deploy
  image: bitnami/kubectl:latest
  
  before_script:
    - mkdir -p ~/.kube
    - echo "$KUBECONFIG_CONTENT" | base64 -d > ~/.kube/config
    - chmod 600 ~/.kube/config
    - export KUBECONFIG=~/.kube/config
  
  script:
    - echo "Deploying frontend to Kubernetes namespace emzops-staging..."
    - kubectl apply -f k8s/configmap.yaml
    - kubectl apply -f k8s/deployment.yaml
    - kubectl apply -f k8s/service.yaml
    - kubectl apply -f k8s/ingress.yaml
    - echo "Waiting for deployment to be ready..."
    - kubectl rollout status deployment/devops-frontend -n emzops-staging --timeout=5m
    - echo "✅ Deployment successful!"
    - echo "Frontend URL - https://devops-mezdit.kubernetes.devops.cs.ut.ee"
    - kubectl get pods -n emzops-staging -l app=devops-frontend
  
  environment:
    name: staging
    url: https://devops-mezdit.kubernetes.devops.cs.ut.ee
  
  only:
    - main
```

### Deployment Workflow

**1. Developer pushes code:**
```bash
git add .
git commit -m "feat: new feature"
git push origin main
```

**2. GitLab triggers pipeline:**
```
✅ Security scan
✅ Tests pass
✅ Build successful
✅ Container image built
✅ Image pushed to registry
→ Deploy stage starts
```

**3. Deploy stage:**
```bash
# Authenticate to cluster
KUBECONFIG decoded → ~/.kube/config

# Apply manifests
kubectl apply -f k8s/configmap.yaml     # ✅ Updated
kubectl apply -f k8s/deployment.yaml    # ✅ New pods created
kubectl apply -f k8s/service.yaml       # ✅ No change (stable)
kubectl apply -f k8s/ingress.yaml       # ✅ Routes updated

# Wait for rollout
kubectl rollout status deployment/...
```

**4. Kubernetes rolling update:**
```
Old pods running (v1) → New pods starting (v2)
  ↓
New pods ready → Old pods terminating
  ↓
All traffic to new pods → Deployment complete
```

**5. Verification:**
```bash
# Pipeline shows:
✅ Deployment successful!
pod/devops-backend-xxx    1/1   Running
pod/devops-frontend-xxx   1/1   Running
```

### GitLab CI/CD Variables Required

| Variable | Type | Purpose |
|----------|------|---------|
| `KUBECONFIG_CONTENT` | Variable | base64-encoded kubeconfig for gitlab-deploy SA |
| `CI_REGISTRY` | Built-in | GitLab registry URL |
| `CI_REGISTRY_USER` | Built-in | Registry username (gitlab-ci-token) |
| `CI_REGISTRY_PASSWORD` | Built-in | Registry password (auto-generated) |
| `CI_COMMIT_SHORT_SHA` | Built-in | Git commit SHA for image tagging |

### Rollout Strategy

**Kubernetes uses Rolling Update by default:**

```yaml
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1        # Max 1 extra pod during update
      maxUnavailable: 0  # Always keep at least 1 pod running
```

**What happens:**
```
1. New pod created (v2)
2. Wait for new pod Ready
3. Route traffic to new pod
4. Terminate old pod (v1)
5. Deployment complete - zero downtime!
```

### Monitoring Deployment

**In GitLab CI/CD pipeline log:**
```bash
Waiting for deployment to be ready...
Waiting for deployment "devops-backend" rollout to finish: 0 of 1 updated replicas are available...
deployment "devops-backend" successfully rolled out
✅ Deployment successful!
```

**Manually check:**
```bash
# Watch deployment status
kubectl rollout status deployment/devops-backend -n emzops-staging

# Check pod status
kubectl get pods -n emzops-staging

# View deployment history
kubectl rollout history deployment/devops-backend -n emzops-staging
```

---

## Testing and Verification

### 1. Service Connectivity Tests

**Backend service:**
```bash
kubectl port-forward service/devops-backend-service -n emzops-staging 8080:80
curl http://localhost:8080
```

**Frontend service:**
```bash
kubectl port-forward service/devops-frontend-service -n emzops-staging 8081:80
curl http://localhost:8081
```

**PostgreSQL service:**
```bash
kubectl port-forward service/postgres-service -n emzops-staging 5432:5432
psql postgresql://devops:PASSWORD@localhost:5432/devops
```

### 2. Ingress Tests

**HTTPS access:**
```bash
curl -I https://devops-mezdit.kubernetes.devops.cs.ut.ee

# Expected:
HTTP/2 200
strict-transport-security: max-age=31536000
```

**HTTP redirect to HTTPS:**
```bash
curl -I http://devops-mezdit.kubernetes.devops.cs.ut.ee

# Expected:
HTTP/1.1 308 Permanent Redirect
Location: https://devops-mezdit.kubernetes.devops.cs.ut.ee
```

**API path routing:**
```bash
curl https://devops-mezdit.kubernetes.devops.cs.ut.ee/api/posts
# Should hit backend service
```

**Frontend path routing:**
```bash
curl https://devops-mezdit.kubernetes.devops.cs.ut.ee/
# Should hit frontend service (HTML)
```

### 3. TLS Certificate Verification

**Check certificate:**
```bash
echo | openssl s_client -showcerts -servername devops-mezdit.kubernetes.devops.cs.ut.ee \
  -connect devops-mezdit.kubernetes.devops.cs.ut.ee:443 2>/dev/null | openssl x509 -inform pem -noout -text

# Look for:
Issuer: C = US, O = Let's Encrypt
Subject: CN = devops-mezdit.kubernetes.devops.cs.ut.ee
```

**Check in browser:**
```
https://devops-mezdit.kubernetes.devops.cs.ut.ee
→ Click lock icon
→ Should show "Let's Encrypt" certificate
```

### 4. Database Connectivity

**From backend pod:**
```bash
# Get pod name
BACKEND_POD=$(kubectl get pod -n emzops-staging -l app=devops-backend -o jsonpath='{.items[0].metadata.name}')

# Exec into pod
kubectl exec -it $BACKEND_POD -n emzops-staging -- sh

# Inside pod, test connection
apk add postgresql-client
psql $DATABASE_URL -c "SELECT version();"
```

### 5. Configuration Verification

**Check ConfigMap:**
```bash
kubectl get configmap backend-config -n emzops-staging -o yaml

# Verify values:
data:
  PORT: "8080"
```

**Check Secret (values are base64 encoded):**
```bash
kubectl get secret backend-secret -n emzops-staging -o jsonpath='{.data.DATABASE_URL}' | base64 -d
# Output: postgresql://devops:***@postgres-service:5432/devops
```

### 6. Pod Health

**Check all pods running:**
```bash
kubectl get pods -n emzops-staging

# Expected:
NAME                               READY   STATUS    RESTARTS   AGE
devops-backend-xxx                 1/1     Running   0          10d
devops-frontend-xxx                1/1     Running   0          1d
postgres-0                         1/1     Running   0          10d
```

**Check pod logs:**
```bash
kubectl logs -f deployment/devops-backend -n emzops-staging
kubectl logs -f deployment/devops-frontend -n emzops-staging
kubectl logs -f postgres-0 -n emzops-staging
```

### 7. RBAC Verification

**Test kubectl as service account:**
```bash
# Use service account kubeconfig
export KUBECONFIG=./gitlab-deploy-kubeconfig

# Should work (allowed):
kubectl get pods -n emzops-staging
kubectl get deployments -n emzops-staging

# Should fail (not allowed):
kubectl get nodes
kubectl get pods -n other-namespace
```

### 8. Resource Status Check

**One command to see everything:**
```bash
kubectl get all,ingress,configmap,secret -n emzops-staging
```

**Check deployment rollout history:**
```bash
kubectl rollout history deployment/devops-backend -n emzops-staging
```

**Check ingress details:**
```bash
kubectl describe ingress devops-backend-ingress -n emzops-staging
```

---

## Week 10 Checklist

### Requirements Completion

| Requirement | Status | Evidence |
|-------------|--------|----------|
| **Services** |
| Create Service for app | ✅ | devops-backend-service, devops-frontend-service |
| Use appropriate type (ClusterIP) | ✅ | All services use ClusterIP |
| **Ingress** |
| Create Ingress resource | ✅ | devops-backend-ingress |
| Route external traffic | ✅ | Path-based routing working |
| Use correct hostname format | ✅ | devops-mezdit.kubernetes.devops.cs.ut.ee |
| **TLS** |
| Automatic cert provisioning | ✅ | cert-manager annotation present |
| HTTPS accessible | ✅ | https://devops-mezdit... works |
| **Database** |
| Separate deployment | ✅ | PostgreSQL StatefulSet |
| One replica | ✅ | spec.replicas: 1 |
| Storage (asked emptyDir, we use PVC) | ✅ | 1Gi PVC (better than emptyDir) |
| **ConfigMaps** |
| All env vars in ConfigMap | ✅ | No hardcoded values |
| **Secrets** |
| All sensitive data in Secrets | ✅ | DATABASE_URL, passwords, tokens |
| Not in Git | ✅ | Only templates in Git |
| **RBAC** |
| ServiceAccount created | ✅ | gitlab-deploy |
| Role created | ✅ | gitlab-deploy-role |
| RoleBinding created | ✅ | gitlab-deploy-binding |
| **CI/CD** |
| Automated deployment | ✅ | deploy-to-k8s stage |
| Uses service account | ✅ | KUBECONFIG_CONTENT |
| **Documentation** |
| Architecture documented | ✅ | This document |
| Traffic flow explained | ✅ | Diagrams included |
| Configuration explained | ✅ | ConfigMaps + Secrets |

**Overall: 100% COMPLETE** ✅

---

## Summary

### What We Built

1. **Kubernetes Services** - Internal networking for pods
2. **Ingress + TLS** - Public HTTPS access with automatic certificates
3. **Separate Database** - PostgreSQL StatefulSet with persistent storage
4. **External Configuration** - ConfigMaps and Secrets (no hardcoded values)
5. **RBAC** - Secure GitLab → Kubernetes integration
6. **Automated CI/CD** - Push to main → Deploy to Kubernetes

### Architecture Highlights

- ✅ **High availability ready:** Can scale to multiple replicas
- ✅ **Secure:** TLS, RBAC, secrets management
- ✅ **Maintainable:** External configuration, automated deployment
- ✅ **Production-ready:** Persistent database, proper networking

### Next Steps (Week 11)

Week 11 will likely cover:
- Persistent volumes (Longhorn)
- Horizontal Pod Autoscaling
- Resource limits and requests
- Monitoring and logging

---

> **Note:** This document was prepared with the assistance of an AI (artificial intelligence) tool.