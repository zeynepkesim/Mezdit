# Week 11 - Kubernetes Observability & Storage Documentation

**Team:** emzops  
**Namespace:** emzops-staging  
**Production URL:** https://devops-mezdit.kubernetes.devops.cs.ut.ee  
**Monitoring URL:** https://monitoring.kubernetes.devops.cs.ut.ee

---

## Table of Contents

1. [Overview](#overview)
2. [Observability Architecture](#observability-architecture)
3. [Health and Readiness Endpoints](#health-and-readiness-endpoints)
4. [Kubernetes Probes](#kubernetes-probes)
5. [Metrics Implementation](#metrics-implementation)
6. [Prometheus Integration](#prometheus-integration)
7. [Grafana Dashboards](#grafana-dashboards)
8. [PostgreSQL Storage](#postgresql-storage)
9. [Testing and Verification](#testing-and-verification)
10. [Troubleshooting Guide](#troubleshooting-guide)

---

## Overview

Week 11 focuses on **observability** and **persistent storage** in Kubernetes:
- ✅ Health and readiness endpoints for application monitoring
- ✅ Liveness and readiness probes in Kubernetes
- ✅ Custom metrics in Prometheus format
- ✅ Prometheus integration for metrics collection
- ✅ Grafana dashboards for visualization
- ✅ Persistent storage with StatefulSet  and PVCs
- ✅ NetworkPolicy for secure metrics scraping

### What Changed from Week 10

| Aspect | Week 10 | Week 11 |
|--------|---------|---------|
| Health Checks | Basic liveness probe | Liveness + readiness probes |
| Metrics | None | Prometheus metrics endpoint |
| Monitoring | Not covered | Prometheus + Grafana |
| Storage | StatefulSet mentioned | Detailed PVC implementation |
| Network Security | Not covered | NetworkPolicy for metrics |

### Key Concepts

**Observability** = Logs + Metrics + Traces
- **Metrics** (Week 11 focus): Numerical data about system behavior
- **Health Probes**: Kubernetes mechanisms to manage pod lifecycle
- **Prometheus**: Time-series database for metrics
- **Grafana**: Visualization and alerting platform

---

## Observability Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Backend Application (Go)                           │   │
│  │  - /healthz endpoint (liveness)                     │   │
│  │  - /readyz endpoint (readiness)                     │   │
│  │  - /metrics endpoint (Prometheus format)            │   │
│  │  - MetricsMiddleware (tracks requests/latency)      │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                   Kubernetes Layer                          │
│  ┌──────────────────┐          ┌──────────────────┐        │
│  │ Liveness Probe   │          │ Readiness Probe  │        │
│  │ GET /healthz     │          │ GET /readyz      │        │
│  │ Port: 8080       │          │ Port: 8080       │        │
│  │                  │          │                  │        │
│  │ Failure→Restart  │          │ Failure→No Route │        │
│  └──────────────────┘          └──────────────────┘        │
│                                                              │
│  ┌──────────────────────────────────────────────┐          │
│  │ Service (devops-backend-service)             │          │
│  │ - Port 80  → Backend API (8080)              │          │
│  │ - Port 9090 → Metrics (9090)                 │          │
│  │ Annotations:                                 │          │
│  │   prometheus.io/scrape: "true"               │          │
│  │   prometheus.io/port: "9090"                 │          │
│  └──────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│               Monitoring Stack (Separate Namespace)         │
│  ┌──────────────────────────────────────────────┐          │
│  │ Prometheus (prometheus-stack namespace)      │          │
│  │ - Service discovery via annotations          │          │
│  │ - Scrapes /metrics every 30s                 │          │
│  │ - Stores time-series data                    │          │
│  │                                               │          │
│  │ NetworkPolicy allows:                        │          │
│  │   prometheus-stack → emzops-staging:9090     │          │
│  └──────────────────┬───────────────────────────┘          │
│                     ↓                                        │
│  ┌──────────────────────────────────────────────┐          │
│  │ Grafana (prometheus-stack namespace)         │          │
│  │ - Queries Prometheus for metrics             │          │
│  │ - Visualizes data in dashboards              │          │
│  │ - Team folder: emzops                        │          │
│  └──────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

```
1. Application generates metrics
   ↓
2. MetricsMiddleware collects request data
   ↓
3. /metrics endpoint exposes in Prometheus format
   ↓
4. Prometheus discovers service via annotations
   ↓
5. NetworkPolicy allows scraping from prometheus-stack namespace
   ↓
6. Prometheus scrapes metrics every 30 seconds
   ↓
7. Stores in time-series database
   ↓
8. Grafana queries Prometheus
   ↓
9. Displays in dashboards (emzops team folder)
```

---

## Health and Readiness Endpoints

### Concept: Liveness vs Readiness

| Aspect | Liveness | Readiness |
|--------|----------|-----------|
| **Question** | Is the app alive? | Is the app ready to serve traffic? |
| **Failure Action** | Restart the pod | Remove from service endpoints |
| **Use Case** | Detect deadlocks, crashes | Wait for dependencies (DB, cache) |
| **Example** | Always return 200 if running | Return 200 only if DB is reachable |

### Implementation

**File:** `internal/handlers/health.go`

```go
package handlers

import (
\t"net/http"
\t"devops-be/internal/database"
)

// HealthzHandler returns 200 OK if the app is running (no external deps)
// Used for liveness probe - if this fails, Kubernetes will restart the pod
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
\tw.WriteHeader(http.StatusOK)
\tw.Write([]byte("ok"))
}

// ReadyzHandler returns 200 OK if the app is ready (DB reachable)
// Used for readiness probe - if this fails, pod is removed from Service
func ReadyzHandler(w http.ResponseWriter, r *http.Request) {
\tif database.DB == nil {
\t\tw.WriteHeader(http.StatusServiceUnavailable)
\t\tw.Write([]byte("db not ready"))
\t\treturn
\t}
\tif err := database.DB.Ping(); err != nil {
\t\tw.WriteHeader(http.StatusServiceUnavailable)
\t\tw.Write([]byte("db not ready"))
\t\treturn
\t}
\tw.WriteHeader(http.StatusOK)
\tw.Write([]byte("ok"))
}
```

### Healthz Endpoint (/healthz)

**Purpose:** Liveness check - is the application process alive?

**Logic:**
- No dependencies checked
- Always returns 200 OK if code is executing
- Fast and simple

**When Kubernetes uses it:**
- Pod starts → wait `initialDelaySeconds` → start checking
- Check every `periodSeconds`
- If `failureThreshold` consecutive failures → **restart pod**

**Typical failure scenarios:**
- Application deadlocked
- Out of memory
- Unrecoverable error state

### Readyz Endpoint (/readyz)

**Purpose:** Readiness check - is the application ready to serve requests?

**Logic:**
1. Check if database connection exists (`database.DB != nil`)
2. Ping database to verify connectivity
3. Return 200 OK only if database is reachable
4. Return 503 Service Unavailable if database is down

**When Kubernetes uses it:**
- Pod starts → wait `initialDelaySeconds` → start checking
- Check every `periodSeconds`
- If probe fails → **remove pod from Service** (no traffic routed)
- If probe succeeds again → add pod back to Service

**Typical failure scenarios:**
- Database temporarily unavailable
- Database credentials incorrect
- Network issue to database

### Testing Locally

```bash
# Test healthz
curl http://localhost:8080/healthz
# Expected: ok (200 OK)

# Test readyz (requires DB running)
curl http://localhost:8080/readyz
# Expected: ok (200 OK) if DB is up
# Expected: db not ready (503) if DB is down
```

### Testing in Kubernetes

```bash
# Port forward to backend
kubectl port-forward service/devops-backend-service -n emzops-staging 8080:80

# Test healthz
curl http://localhost:8080/healthz

# Test readyz
curl http://localhost:8080/readyz

# Check pod events for probe failures
kubectl describe pod -l app=devops-backend -n emzops-staging | grep -A 10 Events
```

---

## Kubernetes Probes

### What are Kubernetes Probes?

Probes are **periodic diagnostic actions** performed by kubelet on containers:
- **Liveness Probe**: Determines if container should be restarted
- **Readiness Probe**: Determines if container should receive traffic
- **Startup Probe**: (not used) Determines if container has started

### Our Configuration

**File:** `devops-be/k8s/deployment.yaml`

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
        - name: gitlab-registry-secret
      containers:
      - name: backend
        image: registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-be:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        
        # Liveness Probe - restart if this fails
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5    # Wait 5s after container starts
          periodSeconds: 10          # Check every 10 seconds
          failureThreshold: 3        # Restart after 3 consecutive failures
        
        # Readiness Probe - remove from service if this fails
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
          failureThreshold: 3
        
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

### Probe Configuration Explained

| Parameter | Liveness | Readiness | Explanation |
|-----------|----------|-----------|-------------|
| `initialDelaySeconds` | 5 | 5 | Wait before first probe (allows app to start) |
| `periodSeconds` | 10 | 10 | How often to perform probe |
| `timeoutSeconds` | 1 (default) | 1 (default) | Probe timeout |
| `failureThreshold` | 3 | 3 | Failures needed before action |
| `successThreshold` | 1 (default) | 1 (default) | Successes needed to be "healthy" |

### Probe Lifecycle Example

```
Pod created (t=0s)
  ↓
Container starts (t=0s)
  ↓
Wait initialDelaySeconds (t=0-5s)
  ↓
First liveness probe (t=5s) → Success ✅
First readiness probe (t=5s) → Fail ❌ (DB not ready)
  ↓
Pod Status: Running but NOT Ready (no traffic)
  ↓
Probe every 10s (t=15s, 25s, 35s...)
  ↓
Readiness probe succeeds (t=35s) → Success ✅
  ↓
Pod added to Service endpoints → Receives traffic 🚀
  ↓
Database goes down (t=120s)
  ↓
Readiness probe fails (t=125s, 135s, 145s)
  ↓
After 3 failures → Pod removed from Service (no restart!)
  ↓
Database comes back (t=200s)
  ↓
Readiness probe succeeds → Pod re-added to Service ✅
```

### Best Practices

**DO:**
- ✅ Keep probes lightweight (< 1s execution time)
- ✅ Use `/healthz` for liveness (no dependencies)
- ✅ Use `/readyz` for readiness (check critical dependencies)
- ✅ Set appropriate `initialDelaySeconds` (longer for slow-starting apps)

**DON'T:**
- ❌ Check external APIs in liveness (may cause unnecessary restarts)
- ❌ Set `periodSeconds` too low (creates load on app)
- ❌ Use same endpoint for liveness and readiness

### Verification

```bash
# Check probe status
kubectl get pods -n emzops-staging
# Look for READY column: 1/1 means readiness probe passing

# Describe pod to see probe results
kubectl describe pod -l app=devops-backend -n emzops-staging

# Look for:
# - Liveness: http-get http://:8080/healthz
# - Readiness: http-get http://:8080/readyz

# Check probe failures in events
kubectl get events -n emzops-staging --field-selector involvedObject.kind=Pod
```

---

## Metrics Implementation

### What are Metrics?

**Metrics** are numerical measurements collected over time:
- **Counters**: Values that only increase (e.g., total requests)
- **Gauges**: Values that can go up or down (e.g., active connections)
- **Histograms**: Distribution of values (e.g., request latencies)

### Prometheus Exposition Format

Prometheus expects metrics in a specific text format:

```
# HELP metric_name Description of the metric
# TYPE metric_name counter|gauge|histogram
metric_name{label="value"} 123 timestamp
```

### Our Metrics

**File:** `internal/handlers/metrics.go`

```go
package handlers

import (
\t"bufio"
\t"fmt"
\t"net"
\t"net/http"
\t"sync/atomic"
\t"time"
\t"devops-be/internal/database"
)

var (
\ttotalRequests uint64  // Counter: total HTTP requests
\ttotalErrors uint64    // Counter: total HTTP errors (4xx, 5xx)
\ttotalLatency int64    // Counter: cumulative latency in microseconds
)

// MetricsHandler exposes Prometheus-style metrics
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
\tw.Header().Set("Content-Type", "text/plain; version=0.0.4")
\t
\t// Counter metrics
\tfmt.Fprintf(w, "# HELP app_requests_total Total HTTP requests\\n")
\tfmt.Fprintf(w, "# TYPE app_requests_total counter\\n")
\tfmt.Fprintf(w, "app_requests_total %d\\n", atomic.LoadUint64(&totalRequests))

\tfmt.Fprintf(w, "# HELP app_errors_total Total HTTP errors\\n")
\tfmt.Fprintf(w, "# TYPE app_errors_total counter\\n")
\tfmt.Fprintf(w, "app_errors_total %d\\n", atomic.LoadUint64(&totalErrors))

\tfmt.Fprintf(w, "# HELP app_request_latency_microseconds Total request latency in microseconds\\n")
\tfmt.Fprintf(w, "# TYPE app_request_latency_microseconds counter\\n")
\tfmt.Fprintf(w, "app_request_latency_microseconds %d\\n", atomic.LoadInt64(&totalLatency))

\t// Gauge metrics from database
\tposts, err := database.GetPosts()
\tif err == nil {
\t\tfmt.Fprintf(w, "# HELP app_posts_total Total number of posts\\n")
\t\tfmt.Fprintf(w, "# TYPE app_posts_total gauge\\n")
\t\tfmt.Fprintf(w, "app_posts_total %d\\n", len(posts))
\t}

\tusers, err := database.GetAllUsers()
\tif err == nil {
\t\tfmt.Fprintf(w, "# HELP app_users_total Total number of users\\n")
\t\tfmt.Fprintf(w, "# TYPE app_users_total gauge\\n")
\t\tfmt.Fprintf(w, "app_users_total %d\\n", len(users))
\t}
}

// MetricsMiddleware tracks requests, errors, and latency
func MetricsMiddleware(next http.Handler) http.Handler {
\treturn http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
\t\tstart := time.Now()
\t\tatomic.AddUint64(&totalRequests, 1)
\t\t
\t\t// Wrap ResponseWriter to capture status code
\t\tww := &responseWriter{ResponseWriter: w, status: 200}
\t\tnext.ServeHTTP(ww, r)
\t\t
\t\t// Track errors (4xx, 5xx)
\t\tif ww.status >= 400 {
\t\t\tatomic.AddUint64(&totalErrors, 1)
\t\t}
\t\t
\t\t// Track latency in microseconds
\t\tatomic.AddInt64(&totalLatency, int64(time.Since(start).Microseconds()))
\t})
}

type responseWriter struct {
\thttp.ResponseWriter
\tstatus int
}

func (rw *responseWriter) WriteHeader(code int) {
\trw.status = code
\trw.ResponseWriter.WriteHeader(code)
}

// Hijack implements http.Hijacker interface for WebSocket support
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
\thijacker, ok := rw.ResponseWriter.(http.Hijacker)
\tif !ok {
\t\treturn nil, nil, fmt.Errorf("underlying ResponseWriter does not support hijacking")
\t}
\treturn hijacker.Hijack()
}
```

### Metrics Explained

#### 1. app_requests_total (Counter)

**What it measures:** Total number of HTTP requests since application start

**Why it's useful:**
- Track traffic patterns
- Calculate requests per second (rate)
- Identify traffic spikes

**Prometheus query examples:**
```promql
# Current total
app_requests_total

# Requests per second (5-minute average)
rate(app_requests_total[5m])

# Total requests in last hour
increase(app_requests_total[1h])
```

#### 2. app_errors_total (Counter)

**What it measures:** Total number of HTTP errors (status >= 400)

**Why it's useful:**
- Monitor application health
- Calculate error rate
- Alert on abnormal error levels

**Prometheus query examples:**
```promql
# Error rate (errors per second)
rate(app_errors_total[5m])

# Error percentage
rate(app_errors_total[5m]) / rate(app_requests_total[5m]) * 100
```

#### 3. app_request_latency_microseconds (Counter)

**What it measures:** Cumulative request processing time in microseconds

**Why it's useful:**
- Calculate average latency
- Identify performance degradation
- Track response times over time

**Prometheus query examples:**
```promql
# Average latency per request (microseconds)
rate(app_request_latency_microseconds[5m]) / rate(app_requests_total[5m])

# Convert to milliseconds
rate(app_request_latency_microseconds[5m]) / rate(app_requests_total[5m]) / 1000
```

#### 4. app_posts_total (Gauge)

**What it measures:** Current number of posts in database

**Why it's useful:**
- Track content growth
- Monitor business metrics
- Detect data issues

**Prometheus query:**
```promql
# Current post count
app_posts_total
```

#### 5. app_users_total (Gauge)

**What it measures:** Current number of users in database

**Why it's useful:**
- Track user growth
- Monitor registration trends
- Business analytics

**Prometheus query: ```promql
# Current user count
app_users_total
```

### Metrics Server Configuration

**In `main.go`:**

```go
// Metrics server (separate port for Week 11 requirement)
metricsServeMux := http.NewServeMux()
metricsServeMux.HandleFunc("/metrics", handlers.MetricsHandler)
go func() {
\tfmt.Println("📊 Metrics server running on http://localhost:9090/metrics")
\tif err := http.ListenAndServe(":9090", metricsServeMux); err != nil {
\t\tlog.Fatalf("Metrics server failed: %v", err)
\t}
}()

// Main application with metrics middleware
handler := handlers.MetricsMiddleware(mux)
log.Fatal(http.ListenAndServe(":8080", handler))
```

**Why separate port?**
- ✅ Isolates metrics from application traffic
- ✅ Enables separate NetworkPolicy rules
- ✅ Standard practice in production systems
- ✅ Port 8080 = Application, Port 9090 = Metrics

### Testing Metrics Endpoint

```bash
# Port forward to metrics port
kubectl port-forward service/devops-backend-service -n emzops-staging 9090:9090

# Fetch metrics
curl http://localhost:9090/metrics

# Expected output:
# HELP app_requests_total Total HTTP requests
# TYPE app_requests_total counter
app_requests_total 1234
# HELP app_errors_total Total HTTP errors
# TYPE app_errors_total counter
app_errors_total 5
# HELP app_request_latency_microseconds Total request latency in microseconds
# TYPE app_request_latency_microseconds counter
app_request_latency_microseconds 567890
# HELP app_posts_total Total number of posts
# TYPE app_posts_total gauge
app_posts_total 42
# HELP app_users_total Total number of users
# TYPE app_users_total gauge
app_users_total 10
```

---

## Prometheus Integration

### What is Prometheus?

**Prometheus** is an open-source monitoring system that:
- Pulls metrics from targets via HTTP
- Stores data as time-series
- Provides powerful query language (PromQL)
- Supports alerting

### Service Discovery

Prometheus automatically discovers our backend service via **annotations**:

**File:** `devops-be/k8s/service.yaml`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: devops-backend-service
  namespace: emzops-staging
  annotations:
    # Tell Prometheus to scrape this service
    prometheus.io/scrape: "true"
    # Port where metrics are exposed
    prometheus.io/port: "9090"
    # Path to metrics endpoint
    prometheus.io/path: "/metrics"
spec:
  type: ClusterIP
  selector:
    app: devops-backend
  ports:
  - name: http
    protocol: TCP
    port: 80
    targetPort: 8080
  - name: metrics
    protocol: TCP
    port: 9090
    targetPort: 9090
```

### How Prometheus Discovers Services

```
1. Prometheus watches Kubernetes API for Services
   ↓
2. Finds Service with annotation: prometheus.io/scrape: "true"
   ↓
3. Reads prometheus.io/port and prometheus.io/path
   ↓
4. Constructs scrape target:
   http://devops-backend-service.emzops-staging.svc.cluster.local:9090/metrics
   ↓
5. Adds target to scrape configuration
   ↓
6. Scrapes metrics every 30 seconds (default interval)
   ↓
7. Stores in time-series database
```

### NetworkPolicy for Scraping

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
      app: devops-backend
  policyTypes:
  - Ingress
  ingress:
  # Allow Prometheus to scrape metrics (port 9090)
  - from:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: prometheus-stack
    ports:
    - protocol: TCP
      port: 9090
  
  # Allow ingress controller traffic (for normal API requests on port 8080)
  - from:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  
  # Allow traffic from same namespace (frontend, postgres, etc.)
  - from:
    - podSelector: {}
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 9090
```

**What this does:**
- ✅ Allows Prometheus (in `prometheus-stack` namespace) to access port 9090
- ✅ Allows Ingress controller to access port 8080
- ✅ Allows internal namespace traffic
- ❌ Blocks all other cross-namespace traffic

### Verification in Prometheus

**Access Prometheus UI:**
```
URL: https://monitoring.kubernetes.devops.cs.ut.ee
```

**Check if targets are discovered:**
1. Navigate to: **Status → Targets**
2. Look for: `emzops-staging/devops-backend-service/0`
3. Status should be: **UP** (green)

**Run queries:**
```promql
# Check if metrics are being collected
app_requests_total{namespace="emzops-staging"}

# Calculate request rate
rate(app_requests_total{namespace="emzops-staging"}[5m])

# Check error rate
rate(app_errors_total{namespace="emzops-staging"}[5m]) / rate(app_requests_total{namespace="emzops-staging"}[5m]) * 100
```

### Troubleshooting Prometheus

**Problem:** Target not discovered

**Solutions:**
```bash
# Check service annotations
kubectl get service devops-backend-service -n emzops-staging -o yaml | grep annotations -A 5

# Annotations should be present:
# prometheus.io/scrape: "true"
# prometheus.io/port: "9090"
# prometheus.io/path: "/metrics"
```

**Problem:** Target discovered but DOWN

**Solutions:**
```bash
# Check if metrics endpoint is accessible from within cluster
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://devops-backend-service.emzops-staging.svc.cluster.local:9090/metrics

# Check NetworkPolicy
kubectl get networkpolicy -n emzops-staging
kubectl describe networkpolicy allow-prometheus-scraping -n emzops-staging

# Check service and pod
kubectl get svc devops-backend-service -n emzops-staging
kubectl get pods -l app=devops-backend -n emzops-staging
```

---

## Grafana Dashboards

### Accessing Grafana

**URL:** https://monitoring.kubernetes.devops.cs.ut.ee

**Login:** Use your university credentials (same as Kubernetes cluster)

**Team Folder:** After 10 minutes of deployment, look for `emzops` folder

### Creating a Dashboard

**Step-by-step:**

1. **Access Grafana** → Log in with university credentials
2. **Navigate** → Dashboards → Browse → Find `emzops` folder
3. **Create** → New Dashboard → Add visualization
4. **Configure panel:**
   - Data source: Prometheus
   - Metric: `app_requests_total{namespace="emzops-staging"}`
   - Apply

### Recommended Panels

#### Panel 1: Request Rate (Graph)

**Query:**
```promql
rate(app_requests_total{namespace="emzops-staging"}[5m])
```

**Panel settings:**
- Visualization: Time series
- Title: "Requests per Second"
- Y-axis: requests/s
- Legend: Show

#### Panel 2: Error Rate Percentage (Stat)

**Query:**
```promql
(rate(app_errors_total{namespace="emzops-staging"}[5m]) / rate(app_requests_total{namespace="emzops-staging"}[5m])) * 100
```

**Panel settings:**
- Visualization: Stat
- Title: "Error Rate %"
- Unit: percent (0-100)
- Thresholds: Green < 1%, Yellow < 5%, Red >= 5%

#### Panel 3: Average Latency (Graph)

**Query:**
```promql
rate(app_request_latency_microseconds{namespace="emzops-staging"}[5m]) / rate(app_requests_total{namespace="emzops-staging"}[5m]) / 1000
```

**Panel settings:**
- Visualization: Time series
- Title: "Average Latency (ms)"
- Y-axis: milliseconds

#### Panel 4: Total Posts (Stat)

**Query:**
```promql
app_posts_total{namespace="emzops-staging"}
```

**Panel settings:**
- Visualization: Stat
- Title: "Total Posts"
- Unit: short

#### Panel 5: Total Users (Stat)

**Query:**
```promql
app_users_total{namespace="emzops-staging"}
```

**Panel settings:**
- Visualization: Stat
- Title: "Total Users"
- Unit: short

#### Panel 6: Application Up Status (Stat)

**Query:**
```promql
up{namespace="emzops-staging", service="devops-backend-service"}
```

**Panel settings:**
- Visualization: Stat
- Title: "Application Status"
- Value mappings: 1 = UP, 0 = DOWN
- Thresholds: 1 = Green, 0 = Red

#### Panel 7: Database Up Status (Stat)

**Query:**
```promql
# This requires custom metric or use readiness probe status
# Alternatively, monitor if backend can connect to DB:
probe_success{job="postgres"}
```

**Alternative (simpler):**
Check if backend pods are ready:
```promql
kube_pod_status_ready{namespace="emzops-staging", pod=~"postgres-.*"} == 1
```

#### Panel 8: Storage Usage (Gauge)

**Query:**
```promql
# PVC usage percentage
kubelet_volume_stats_used_bytes{namespace="emzops-staging"} / kubelet_volume_stats_capacity_bytes{namespace="emzops-staging"} * 100
```

**Panel settings:**
- Visualization: Gauge
- Title: "PostgreSQL Storage Usage %"
- Unit: percent (0-100)
- Thresholds: Green < 70%, Yellow < 85%, Red >= 85%

#### Panel 9: CPU Usage (Graph)

**Query:**
```promql
rate(container_cpu_usage_seconds_total{namespace="emzops-staging", pod=~"devops-backend-.*"}[5m])
```

**Panel settings:**
- Visualization: Time series
- Title: "Backend CPU Usage"
- Y-axis: cores

#### Panel 10: Memory Usage (Graph)

**Query:**
```promql
container_memory_usage_bytes{namespace="emzops-staging", pod=~"devops-backend-.*"} / 1024 / 1024
```

**Panel settings:**
- Visualization: Time series
- Title: "Backend Memory Usage"
- Y-axis: MB

### Dashboard Layout Example

```
┌─────────────────────────────────────────────────────────────┐
│  Dashboard: emzops Application Monitoring                   │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ App Status   │  │ DB Status    │  │ Error Rate % │     │
│  │   UP 🟢      │  │   UP 🟢      │  │    0.5%      │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Requests per Second                                  │   │
│  │ [graph showing request rate over time]              │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────┐  ┌────────────────────────┐ │
│  │ Average Latency (ms)     │  │ Storage Usage %        │ │
│  │ [graph]                  │  │ [gauge showing 45%]    │ │
│  └──────────────────────────┘  └────────────────────────┘ │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │ Total Posts  │  │ Total Users  │                        │
│  │     42       │  │     10       │                        │
│  └──────────────┘  └──────────────┘                        │
│                                                              │
│  ┌───────────────────────────┐  ┌──────────────────────┐  │
│  │ CPU Usage                 │  │ Memory Usage         │  │
│  │ [graph]                   │  │ [graph]              │  │
│  └───────────────────────────┘  └──────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Alerting (Optional - Advanced)

You can set up alerts in Grafana:

**Example alert: High error rate**
```
Condition: Error rate > 5% for 5 minutes
Action: Send notification
```

---

## PostgreSQL Storage

### StatefulSet vs Deployment for Databases

| Aspect | Deployment | StatefulSet |
|--------|------------|-------------|
| Pod naming | Random hash | Ordered (postgres-0, postgres-1, ...) |
| Network identity | Ephemeral | Stable (postgres-0.postgres-service) |
| Storage | Shared PVC or emptyDir | Dedicated PVC per pod |
| Use case | Stateless apps | Stateful apps (databases) |

### Our PostgreSQL Implementation

**File:** `k8s/postgres-statefulset.yaml`

```yaml
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: emzops-staging
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
  clusterIP: None  # Headless service for StatefulSet
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: emzops-staging
spec:
  serviceName: postgres-service
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
        - name: POSTGRES_DB
          value: "devops"
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
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
  
  # PVC template - creates one PVC per pod
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 1Gi
```

### PersistentVolumeClaim (PVC)

**What is created:**
```bash
$ kubectl get pvc -n emzops-staging
NAME                        STATUS   VOLUME     CAPACITY   ACCESS MODES   STORAGECLASS
postgres-storage-postgres-0 Bound    pvc-xyz    1Gi        RWO            longhorn
```

**Explanation:**
- Name: `postgres-storage-postgres-0` (template name + pod ordinal)
- Status: `Bound` (successfully attached to a PersistentVolume)
- Capacity: `1Gi` (as requested)
- Access Mode: `RWO` (ReadWriteOnce - one pod at a time)
- StorageClass: `longhorn` (default in your cluster)

### Storage Class: Longhorn

**What is Longhorn?**
- Distributed block storage for Kubernetes
- Replicated across multiple nodes
- Automatic backups
- Snapshots supported

**Why Longhorn?**
- ✅ Data survives pod deletion
- ✅ Data survives node failure
- ✅ Persistent across deployments
- ✅ Production-ready

### emptyDir vs PVC Comparison

**Week 11 suggests emptyDir:**
```yaml
volumes:
- name: db-storage
  emptyDir:
    sizeLimit: 400Mi
```

**We use PVC - comparison:**

| Feature | emptyDir | PVC (Longhorn) |
|---------|----------|----------------|
| **Persistence** | Lost when pod deletes | Persists across pod restarts |
| **Size** | 400Mi (course minimum) | 1Gi (better) |
| **Replication** | Single node | Replicated across nodes |
| **Backup** | Manual | Automated |
| **Use Case** | Development, cache | Production databases |
| **Week 11 Requirement** | ✅ Meets | ✅ Exceeds |

**Our choice: PVC** because:
- ✅ Production-grade
- ✅ Data safety
- ✅ Better than minimum requirement

### Testing Storage Persistence

```bash
# Step 1: Write data to database
kubectl port-forward service/postgres-service -n emzops-staging 5432:5432
psql -h localhost -U devops -d devops
# CREATE TABLE test (id SERIAL, data TEXT);
# INSERT INTO test (data) VALUES ('persistent data');
# SELECT * FROM test;

# Step 2: Delete the pod
kubectl delete pod postgres-0 -n emzops-staging

# Step 3: Wait for pod to recreate
kubectl get pods -n emzops-staging -w

# Step 4: Check data is still there
kubectl port-forward service/postgres-service -n emzops-staging 5432:5432
psql -h localhost -U devops -d devops
# SELECT * FROM test;
# Expected: Data is still present! ✅
```

### Monitoring Storage

```bash
# Check PVC
kubectl get pvc -n emzops-staging

# Describe PVC
kubectl describe pvc postgres-storage-postgres-0 -n emzops-staging

# Check storage usage (if metrics available)
kubectl exec -it postgres-0 -n emzops-staging -- df -h /var/lib/postgresql/data
```

---

## Testing and Verification

### Local Testing (Development)

**Prerequisites:**
- PostgreSQL running locally
- Go 1.23 installed

**Steps:**
```bash
# 1. Set environment variables
export DATABASE_URL="postgresql://devops:password@localhost:5432/devops"
export PORT="8080"

# 2. Run application
cd devops-be
go run main.go

# 3. Test health endpoints
curl http://localhost:8080/healthz  # Should return: ok
curl http://localhost:8080/readyz   # Should return: ok (if DB is up)

# 4. Test metrics endpoint
curl http://localhost:9090/metrics

# 5. Generate some traffic
for i in {1..100}; do curl http://localhost:8080/posts; done

# 6. Check if metrics increased
curl http://localhost:9090/metrics | grep app_requests_total
```

### Kubernetes Testing

#### Test 1: Health Endpoints via Port Forward

```bash
# Port forward to backend service
kubectl port-forward service/devops-backend-service -n emzops-staging 8080:80

# Test healthz
curl http://localhost:8080/healthz
# Expected: ok

# Test readyz
curl http://localhost:8080/readyz
# Expected: ok (if database is running)
```

#### Test 2: Metrics Endpoint

```bash
# Port forward to metrics port
kubectl port-forward service/devops-backend-service -n emzops-staging 9090:9090

# Fetch metrics
curl http://localhost:9090/metrics

# Verify metrics are present
curl http://localhost:9090/metrics | grep -E "app_(requests|errors|posts|users)_total"
```

#### Test 3: Probe Functionality

```bash
# Check pod is ready
kubectl get pods -n emzops-staging -l app=devops-backend
# READY column should show 1/1

# Describe pod to see probe results
kubectl describe pod -l app=devops-backend -n emzops-staging | grep -A 20 "Liveness\\|Readiness"

# Simulate failure by breaking database connection
kubectl scale statefulset postgres -n emzops-staging --replicas=0

# Watch pod status
kubectl get pods -n emzops-staging -l app=devops-backend -w
# After ~30 seconds, READY should change from 1/1 to 0/1

# Restore database
kubectl scale statefulset postgres -n emzops-staging --replicas=1

# Pod should become ready again
kubectl get pods -n emzops-staging -l app=devops-backend -w
```

#### Test 4: Prometheus Scraping

**Option A: Via Prometheus UI**
1. Visit: https://monitoring.kubernetes.devops.cs.ut.ee
2. Go to: Status → Targets
3. Search for: `emzops-staging`
4. Verify: Status is UP (green)

**Option B: Via kubectl**
```bash
# Check NetworkPolicy allows Prometheus
kubectl get networkpolicy -n emzops-staging

# Verify service annotations
kubectl get service devops-backend-service -n emzops-staging -o yaml | grep -A 5 annotations
```

**Option C: Run Test Query**
1. Go to Prometheus UI → Graph
2. Enter query: `app_requests_total{namespace="emzops-staging"}`
3. Execute
4. Should see data points

#### Test 5: Storage Persistence

```bash
# Create test data
kubectl exec -it postgres-0 -n emzops-staging -- psql -U devops -d devops -c \
  "CREATE TABLE IF NOT EXISTS test_persistence (id SERIAL, timestamp TIMESTAMP DEFAULT NOW());"

kubectl exec -it postgres-0 -n emzops-staging -- psql -U devops -d devops -c \
  "INSERT INTO test_persistence VALUES (DEFAULT, NOW());"

kubectl exec -it postgres-0 -n emzops-staging -- psql -U devops -d devops -c \
  "SELECT * FROM test_persistence;"

# Delete pod
kubectl delete pod postgres-0 -n emzops-staging

# Wait for recreation
kubectl wait --for=condition=Ready pod/postgres-0 -n emzops-staging --timeout=60s

# Verify data persists
kubectl exec -it postgres-0 -n emzops-staging -- psql -U devops -d devops -c \
  "SELECT * FROM test_persistence;"
# Data should still be there! ✅
```

### End-to-End Verification Checklist

- [ ] **Health endpoints work**
  ```bash
  curl via port-forward → /healthz returns ok
  curl via port-forward → /readyz returns ok
  ```

- [ ] **Kubernetes probes configured**
  ```bash
  kubectl describe pod shows livenessProbe and readinessProbe
  Pod is in Ready state (1/1)
  ```

- [ ] **Metrics endpoint exposed**
  ```bash
  curl http://localhost:9090/metrics shows all 5 metrics
  Metrics increase after generating traffic
  ```

- [ ] **Prometheus scraping works**
  ```bash
  Prometheus UI shows target as UP
  Queries return data: app_requests_total{namespace="emzops-staging"}
  ```

- [ ] **Grafana dashboard exists**
  ```bash
  emzops folder exists in Grafana
  Dashboard shows panels with data
  All panels are green/working
  ```

- [ ] **Storage is persistent**
  ```bash
  PVC exists and is Bound
  Data survives pod deletion
  Storage usage is monitored
  ```

- [ ] **NetworkPolicy allows scraping**
  ```bash
  NetworkPolicy exists
  Prometheus can reach metrics endpoint
  ```

---

## Troubleshooting Guide

### Problem 1: Readiness Probe Failing

**Symptoms:**
```bash
$ kubectl get pods -n emzops-staging
NAME                               READY   STATUS    RESTARTS   AGE
devops-backend-5ddbf4c5c7-x8q9b    0/1     Running   0          2m
```

Pod is Running but not Ready (0/1).

**Diagnosis:**
```bash
# Check probe failures in events
kubectl describe pod -l app=devops-backend -n emzops-staging | grep -A 10 Events

# Look for:
# Readiness probe failed: HTTP probe failed with statuscode: 503

# Check readyz endpoint manually
kubectl port-forward pod/devops-backend-xxx -n emzops-staging 8080:8080
curl -v http://localhost:8080/readyz
```

**Possible Causes:**
1. **Database not ready**
   ```bash
   # Check postgres pod
   kubectl get pods -n emzops-staging -l app=postgres
   
   # If postgres is not running:
   kubectl get statefulset postgres -n emzops-staging
   kubectl describe statefulset postgres -n emzops-staging
   ```

2. **Wrong DATABASE_URL**
   ```bash
   # Check secret
   kubectl get secret backend-secret -n emzops-staging -o yaml
   
   # Verify DATABASE_URL format:
   # postgresql://USER:PASSWORD@postgres-service:5432/DATABASE
   ```

3. **Network issue**
   ```bash
   # Test connectivity from backend to postgres
   kubectl exec -it deployment/devops-backend -n emzops-staging -- \
     nc -zv postgres-service 5432
   ```

**Solutions:**
```bash
# Fix 1: Ensure postgres is running
kubectl scale statefulset postgres -n emzops-staging --replicas=1

# Fix 2: Recreate secret with correct DATABASE_URL
kubectl delete secret backend-secret -n emzops-staging
kubectl create secret generic backend-secret -n emzops-staging \
  --from-literal=DATABASE_URL="postgresql://devops:PASSWORD@postgres-service:5432/devops"

# Fix 3: Restart backend deployment
kubectl rollout restart deployment/devops-backend -n emzops-staging
```

### Problem 2: Prometheus Not Scraping

**Symptoms:**
- Prometheus UI shows target as DOWN
- Or target not discovered at all

**Diagnosis:**
```bash
# Check if target is discovered
# Go to Prometheus UI → Status → Targets
# Search for emzops-staging

# Check service annotations
kubectl get service devops-backend-service -n emzops-staging -o yaml | grep -A 5 annotations
```

**Cause 1: Missing Annotations**

**Solution:**
```bash
# Add annotations to service
kubectl annotate service devops-backend-service -n emzops-staging \
  prometheus.io/scrape=true \
  prometheus.io/port=9090 \
  prometheus.io/path=/metrics
```

**Cause 2: NetworkPolicy Blocking**

**Diagnosis:**
```bash
# Check NetworkPolicy
kubectl get networkpolicy -n emzops-staging
kubectl describe networkpolicy allow-prometheus-scraping -n emzops-staging
```

**Solution:**
```bash
# Ensure NetworkPolicy allows prometheus-stack namespace
kubectl apply -f k8s/networkpolicy.yaml
```

**Cause 3: Metrics Endpoint Not Responding**

**Diagnosis:**
```bash
# Test from within cluster
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://devops-backend-service.emzops-staging.svc.cluster.local:9090/metrics
```

**Solution:**
```bash
# Check backend logs
kubectl logs -f deployment/devops-backend -n emzops-staging

# Look for: "📊 Metrics server running on http://localhost:9090/metrics"

# If not present, backend might not have started metrics server
# Check main.go has metrics server goroutine
```

### Problem 3: No Data in Grafana

**Symptoms:**
- Panels show "No data"
- Queries return empty results

**Diagnosis:**
```bash
# Test query in Prometheus first
# Go to Prometheus UI → Graph
# Run: app_requests_total{namespace="emzops-staging"}
```

**Cause 1: No Traffic to Application**

**Solution:**
```bash
# Generate some traffic
for i in {1..100}; do
  curl https://devops-mezdit.kubernetes.devops.cs.ut.ee/api/posts
done

# Wait 1 minute, then check metrics again
```

**Cause 2: Wrong Query in Grafana**

**Solution:**
- Ensure query includes namespace filter:
  ```promql
  app_requests_total{namespace="emzops-staging"}
  ```
- Check time range (top right) - set to "Last 1 hour"

**Cause 3: Grafana Not Connected to Prometheus**

**Solution:**
- Check data source: Configuration → Data Sources → Prometheus
- Should be pre-configured by cluster admins

### Problem 4: Storage Full

**Symptoms:**
```bash
$ kubectl get pods -n emzops-staging
NAME         READY   STATUS             RESTARTS   AGE
postgres-0   0/1     CrashLoopBackOff   5          10m
```

**Diagnosis:**
```bash
# Check PVC usage
kubectl exec -it postgres-0 -n emzops-staging -- df -h /var/lib/postgresql/data

# If near 100%, storage is full
```

**Solutions:**
```bash
# Option 1: Increase PVC size (requires Longhorn support for expansion)
kubectl edit pvc postgres-storage-postgres-0 -n emzops-staging
# Change storage: 1Gi to storage: 2Gi

# Option 2: Clean up old data
kubectl exec -it postgres-0 -n emzops-staging -- psql -U devops -d devops
# VACUUM FULL;  -- Reclaim space

# Option 3: Delete and recreate (⚠️ DATA LOSS)
kubectl delete statefulset postgres -n emzops-staging
kubectl delete pvc postgres-storage-postgres-0 -n emzops-staging
kubectl apply -f k8s/postgres-statefulset.yaml
```

### Problem 5: High Error Rate

**Symptoms:**
- Grafana shows error rate > 5%
- `app_errors_total` increasing rapidly

**Diagnosis:**
```bash
# Check backend logs
kubectl logs -f deployment/devops-backend -n emzops-staging

# Look for error patterns
kubectl logs deployment/devops-backend -n emzops-staging | grep -i error

# Check specific endpoints
curl -I https://devops-mezdit.kubernetes.devops.cs.ut.ee/api/posts
```

**Possible Causes:**
1. **Application bug** → Check logs, fix code
2. **Database connection issues** → Verify DATABASE_URL, check postgres
3. **External dependency failure** → Check if all services are up

### Useful Commands Reference

```bash
# Pods
kubectl get pods -n emzops-staging
kubectl describe pod POD_NAME -n emzops-staging
kubectl logs -f POD_NAME -n emzops-staging

# Services
kubectl get svc -n emzops-staging
kubectl describe svc SERVICE_NAME -n emzops-staging

# Port Forwarding
kubectl port-forward service/devops-backend-service -n emzops-staging 8080:80
kubectl port-forward service/devops-backend-service -n emzops-staging 9090:9090

# Probes
kubectl describe pod -l app=devops-backend -n emzops-staging | grep -A 10 "Liveness\\|Readiness"

# Metrics
curl http://localhost:9090/metrics  # after port-forward

# Storage
kubectl get pvc -n emzops-staging
kubectl describe pvc PVC_NAME -n emzops-staging

# NetworkPolicy
kubectl get networkpolicy -n emzops-staging
kubectl describe networkpolicy POLICY_NAME -n emzops-staging

# Events
kubectl get events -n emzops-staging --sort-by='.lastTimestamp'

# Prometheus queries
# Go to: https://monitoring.kubernetes.devops.cs.ut.ee
app_requests_total{namespace="emzops-staging"}
rate(app_requests_total{namespace="emzops-staging"}[5m])
```

---

## Summary

### Week 11 Requirements Completion

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Health endpoint (`/healthz`) | ✅ Complete | `internal/handlers/health.go` |
| Readiness endpoint (`/readyz`) | ✅ Complete | `internal/handlers/health.go` |
| Liveness probe | ✅ Complete | `k8s/deployment.yaml` livenessProbe |
| Readiness probe | ✅ Complete | `k8s/deployment.yaml` readinessProbe |
| Metrics endpoint (`/metrics`) | ✅ Complete | `internal/handlers/metrics.go` |
| 4+ custom metrics | ✅ Complete | requests, errors, latency, posts, users |
| Prometheus scraping | ✅ Complete | Service annotations + NetworkPolicy |
| Grafana dashboard | ✅ Complete | Team folder with panels |
| Persistent storage | ✅ Complete | StatefulSet with PVC (1Gi, Longhorn) |
| Documentation | ✅ Complete | This document |

### Key Achievements

- ✅ **Production-grade observability** with Prometheus and Grafana
- ✅ **Robust health checks** with separate liveness and readiness
- ✅ **Custom application metrics** tracking business and technical metrics
- ✅ **Secure network policies** allowing only Prometheus scraping
- ✅ **Persistent database storage** with PVC and Longhorn
- ✅ **Separated metrics endpoint** on port 9090 for security
- ✅ **WebSocket support** with Hijacker interface in metrics middleware

### Next Steps (Beyond Week 11)

Based on production best practices:
- 🔄 Set up alerting rules in Grafana
- 🔄 Implement distributed tracing (e.g., Jaeger)
- 🔄 Add structured logging (e.g., JSON logs)
- 🔄 Configure log aggregation (e.g., Loki)
- 🔄 Implement backup strategy for PostgreSQL
- 🔄 Scale to multiple replicas with proper load balancing

---

> **Note:** This document was prepared with the assistance of an AI (artificial intelligence) tool.
