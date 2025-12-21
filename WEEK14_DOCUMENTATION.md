# Week 14 - Production Documentation

## Production Deployment

### Staging vs Production Differences

Our application is deployed to two environments with different configurations:

#### Staging Environment (`emzops-staging`)
- **Purpose**: Testing and development
- **Replicas**: 1 (single pod)
- **Resources**: Lower limits (100m CPU, 128Mi memory requests)
- **Autoscaling**: Disabled
- **Affinity**: Not configured
- **Storage**: 1Gi PostgreSQL storage

#### Production Environment (`emzops-prod`)
- **Purpose**: Production workload
- **Replicas**: 2 (minimum for high availability)
- **Resources**: Higher limits (200m CPU, 256Mi memory requests; 500m CPU, 512Mi memory limits)
- **Autoscaling**: Enabled (2-5 replicas based on CPU/memory)
- **Affinity**: Pod anti-affinity configured (pods spread across nodes)
- **Storage**: 2Gi PostgreSQL storage

### Deployment Configuration

**Staging Deployment:**
```bash
helm upgrade --install devops-backend ./chart \
  --namespace emzops-staging \
  --set image.tag=$IMAGE_TAG \
  --set secrets.backend.databaseUrl=$DATABASE_URL \
  # ... other secrets
```

**Production Deployment:**
```bash
helm upgrade --install devops-backend ./chart \
  --namespace emzops-prod \
  --values ./chart/values-prod.yaml \
  --set image.tag=$IMAGE_TAG \
  --set secrets.backend.databaseUrl=$DATABASE_URL_PROD \
  # ... production-specific secrets
```

**Key Differences:**
- Production uses `--values ./chart/values-prod.yaml` to override default values
- Production uses different secret variables (suffixed with `_PROD`)
- Production deployment is **manual** (`when: manual`) for safety

### Pipeline Configuration

The CI/CD pipeline includes two deployment jobs:

1. **deploy-to-staging**: Automatic deployment on every push to `main`
2. **deploy-to-production**: Manual deployment (requires manual trigger)

This ensures:
- Staging is always up-to-date for testing
- Production deployments are controlled and intentional
- Different configurations are applied via Helm values files

## Automatic Scaling

### Implementation: Kubernetes HPA (Horizontal Pod Autoscaler)

We use **Kubernetes native HPA** (Horizontal Pod Autoscaler) for automatic scaling.

**Why HPA over KEDA?**
- HPA is built into Kubernetes (no additional dependencies)
- Simpler to configure and maintain
- Sufficient for our CPU/memory-based scaling needs
- Well-documented and widely used

### HPA Configuration

**Location**: `chart/templates/hpa.yaml`

**Configuration:**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
spec:
  minReplicas: 2
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 30  # Low threshold for easy demo
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 50  # Low threshold for easy demo
```

**Scaling Behavior:**
- **Scale Up**: Aggressive (100% increase or +2 pods every 30 seconds)
- **Scale Down**: Conservative (50% decrease every 60 seconds, 5-minute stabilization)

**Metrics Used:**
- **CPU Utilization**: Target 30% (low threshold for easy demonstration)
- **Memory Utilization**: Target 50% (low threshold for easy demonstration)

### Which Deployments are Scaled?

**Backend Deployment (`devops-backend`):**
- ✅ **Scaled** - Handles API requests and business logic
- Most likely to experience variable load
- Benefits from horizontal scaling

**Frontend Deployment (`devops-frontend`):**
- ✅ **Scaled** - Serves static files and handles user traffic
- Can benefit from scaling during traffic spikes
- Stateless, perfect for horizontal scaling

**PostgreSQL StatefulSet:**
- ❌ **Not Scaled** - StatefulSet with persistent storage
- Scaling requires careful coordination
- Single replica is sufficient for current workload

### Demonstrating Autoscaling

**Method 1: CPU Stress Test**
```bash
# Get a pod name
POD=$(kubectl get pods -n emzops-prod -l app=devops-backend -o jsonpath='{.items[0].metadata.name}')

# Install stress-ng in the pod and generate CPU load
kubectl exec -it $POD -n emzops-prod -- sh -c "apk add stress-ng && stress-ng --cpu 2 --timeout 300s"

# Watch HPA scale up
watch kubectl get hpa -n emzops-prod
watch kubectl get pods -n emzops-prod -l app=devops-backend
```

**Method 2: Generate Traffic**
```bash
# Generate load using curl in a loop
for i in {1..1000}; do
  curl -s https://devops-mezdit.kubernetes.devops.cs.ut.ee/api/posts > /dev/null
done

# Watch HPA metrics
kubectl get hpa devops-backend -n emzops-prod -w
```

**Expected Behavior:**
1. CPU/Memory usage increases
2. HPA detects threshold exceeded
3. Deployment scales up (2 → 3 → 4 → 5 pods)
4. When load decreases, HPA scales down (5 → 4 → 3 → 2 pods)

## Affinity and Topology Constraints

### Pod Anti-Affinity Configuration

To prevent a single node failure from taking the application offline, we use **Pod Anti-Affinity**.

**Location**: `chart/templates/deployment.yaml` and `chart/values-prod.yaml`

**Configuration:**
```yaml
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app
            operator: In
            values:
            - devops-backend
        topologyKey: kubernetes.io/hostname
```

**How it Works:**
- **Preferred** (soft constraint): Kubernetes tries to schedule pods on different nodes
- **Topology Key**: `kubernetes.io/hostname` ensures pods are on different nodes
- **Weight**: 100 (high priority for anti-affinity)

**Benefits:**
- If one node fails, pods on other nodes continue serving traffic
- Zero downtime during node failures
- Better resource distribution across cluster

### Verification

**Check Pod Distribution:**
```bash
# See which nodes pods are running on
kubectl get pods -n emzops-prod -l app=devops-backend -o wide

# Should show pods on different nodes:
# NAME                          NODE
# devops-backend-xxx-abc        do-worker8
# devops-backend-xxx-def        do-worker9
```

**Test Node Failure:**
1. Identify a node running a pod: `kubectl get pods -n emzops-prod -o wide`
2. Simulate node failure (or drain the node)
3. Verify:
   - Remaining pods continue serving traffic
   - New pods are scheduled on different nodes
   - Application remains available

### Multiple Deployments

**Backend Deployment:**
- ✅ Pod anti-affinity configured
- Minimum 2 replicas in production
- Spreads across nodes

**Frontend Deployment:**
- ✅ Pod anti-affinity configured
- Minimum 2 replicas in production
- Spreads across nodes

**PostgreSQL StatefulSet:**
- Uses persistent storage (Longhorn)
- Single replica (sufficient for current workload)
- Can be configured with node affinity if needed

## Chaos Engineering Readiness

### Zero Downtime Requirements

Our production deployment is configured to handle chaos engineering experiments:

✅ **Pod Faults:**
- Multiple replicas (min 2)
- Pod anti-affinity ensures pods on different nodes
- Rolling updates with `maxUnavailable: 0`
- Readiness probes ensure only healthy pods receive traffic

✅ **Network Faults:**
- Service load balancing distributes traffic
- Multiple pods provide redundancy
- Health checks detect and route around failed pods

✅ **Stress Scenarios:**
- Resource limits prevent resource exhaustion
- HPA scales up to handle increased load
- Pod anti-affinity prevents all pods from being affected

### Testing Chaos Readiness

**Pod Kill Test:**
```bash
# Kill a random pod
kubectl delete pod -n emzops-prod -l app=devops-backend --field-selector status.phase=Running --dry-run=client -o jsonpath='{.items[0].metadata.name}' | xargs kubectl delete pod -n emzops-prod

# Verify:
# - Application remains available
# - New pod starts automatically
# - No downtime observed
```

**Node Drain Test:**
```bash
# Drain a node (simulates node failure)
kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data

# Verify:
# - Pods reschedule on other nodes
# - Application remains available
# - No data loss
```

## Summary

✅ **Production Deployment**: Configured with `values-prod.yaml`  
✅ **Automatic Scaling**: HPA enabled (2-5 replicas, CPU/memory based)  
✅ **Affinity Constraints**: Pod anti-affinity ensures pods on different nodes  
✅ **Chaos Ready**: Zero downtime configuration for pod/node failures  
✅ **Pipeline Integration**: Manual production deployment for safety

