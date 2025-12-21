# Week 14 Testing Guide

## Prerequisites

1. **Production namespace exists:**
```bash
kubectl get namespace emzops-prod
# ✅ Namespace already exists (created 24 days ago)
# If it doesn't exist, Helm will create it with --create-namespace flag
```

2. **Production secrets configured in GitLab CI/CD:**
   - ✅ Production uses the same CI/CD variables as staging (no `_PROD` suffix needed)
   - Go to GitLab project → Settings → CI/CD → Variables
   - Verify these variables exist (already configured):
     - `DATABASE_URL`
     - `POSTGRES_USER`
     - `POSTGRES_PASSWORD`
     - `MINIO_ENDPOINT`
     - `MINIO_ACCESS_KEY`
     - `MINIO_SECRET_KEY`
     - `MINIO_BUCKET_NAME`
     - `MINIO_USE_SSL`

3. **ImagePullSecret exists in production namespace:**
```bash
# Copy secret from staging to production
kubectl get secret gitlab-registry-secret -n emzops-staging -o yaml | \
  sed 's/namespace: emzops-staging/namespace: emzops-prod/' | \
  kubectl apply -f -
```

## 1. Production Deployment Verification

### ✅ Deploy to Production

1. **Trigger production deployment:**
   - Go to GitLab CI/CD → Pipelines
   - Find the latest pipeline
   - Click "Play" button on `deploy-to-production` job

2. **Verify deployment:**
```bash
# Check pods are running
kubectl get pods -n emzops-prod

# Check deployments
kubectl get deployments -n emzops-prod

# Check services
kubectl get services -n emzops-prod
```

3. **Verify different configuration:**
```bash
# Compare staging vs production replicas
kubectl get deployment devops-backend -n emzops-staging -o jsonpath='{.spec.replicas}'
kubectl get deployment devops-backend -n emzops-prod -o jsonpath='{.spec.replicas}'
# Production should show 2, staging should show 1

# Compare resources
kubectl get deployment devops-backend -n emzops-prod -o yaml | grep -A 5 "resources:"
# Production should have higher limits
```

## 2. Automatic Scaling Verification

### ✅ Check HPA is Created

```bash
# Check HPA exists
kubectl get hpa -n emzops-prod

# Should show:
# NAME              REFERENCE                    TARGETS         MINPODS   MAXPODS   REPLICAS
# devops-backend    Deployment/devops-backend    <unknown>/30%  2         5         2
```

### ✅ Test Autoscaling (CPU Stress)

**Method 1: Using stress-ng (if available)**
```bash
# Get a pod name
POD=$(kubectl get pods -n emzops-prod -l app=devops-backend -o jsonpath='{.items[0].metadata.name}')

# Install stress-ng and generate CPU load
kubectl exec -it $POD -n emzops-prod -- sh -c "apk add stress-ng && stress-ng --cpu 2 --timeout 300s" &

# Watch HPA scale up (in another terminal)
watch kubectl get hpa -n emzops-prod
watch kubectl get pods -n emzops-prod -l app=devops-backend
```

**Method 2: Using kubectl run (simpler)**
```bash
# Run a CPU-intensive job
kubectl run cpu-stress --image=containerstack/cpustress --restart=Never -n emzops-prod -- --cpu 2

# Watch scaling
watch kubectl get pods -n emzops-prod -l app=devops-backend
```

**Expected Result:**
- HPA detects high CPU usage
- Deployment scales from 2 → 3 → 4 → 5 pods
- After load stops, scales down to 2 pods

### ✅ Test Autoscaling (Memory Stress)

```bash
# Get a pod name
POD=$(kubectl get pods -n emzops-prod -l app=devops-backend -o jsonpath='{.items[0].metadata.name}')

# Generate memory load
kubectl exec -it $POD -n emzops-prod -- sh -c "dd if=/dev/zero of=/tmp/memory bs=1M count=200" &

# Watch HPA
watch kubectl get hpa -n emzops-prod
```

### ✅ Verify HPA Metrics

```bash
# Check HPA status
kubectl describe hpa devops-backend -n emzops-prod

# Should show:
# - Current metrics (CPU, memory)
# - Target metrics
# - Scaling events
```

## 3. Affinity and Topology Constraints Verification

### ✅ Check Pod Distribution

```bash
# Check which nodes pods are running on
kubectl get pods -n emzops-prod -l app=devops-backend -o wide

# Should show pods on different nodes:
# NAME                          NODE          STATUS
# devops-backend-xxx-abc        do-worker8    Running
# devops-backend-xxx-def        do-worker9    Running
```

### ✅ Test Node Failure

1. **Identify a node with a pod:**
```bash
kubectl get pods -n emzops-prod -l app=devops-backend -o wide
# Note the node name
```

2. **Drain the node (simulates failure):**
```bash
NODE_NAME="do-worker8"  # Replace with actual node
kubectl drain $NODE_NAME --ignore-daemonsets --delete-emptydir-data
```

3. **Verify:**
```bash
# Check pods rescheduled
kubectl get pods -n emzops-prod -l app=devops-backend -o wide

# Should show:
# - Pods moved to different nodes
# - All pods Running
# - Application still accessible
```

4. **Uncordon the node:**
```bash
kubectl uncordon $NODE_NAME
```

### ✅ Verify Anti-Affinity Configuration

```bash
# Check deployment has affinity configured
kubectl get deployment devops-backend -n emzops-prod -o yaml | grep -A 10 "affinity:"

# Should show podAntiAffinity configuration
```

## 4. Chaos Engineering Readiness

### ✅ Pod Kill Test

```bash
# Kill a random pod
POD=$(kubectl get pods -n emzops-prod -l app=devops-backend -o jsonpath='{.items[0].metadata.name}')
kubectl delete pod $POD -n emzops-prod

# Verify:
# - New pod starts automatically
# - Application remains available
# - No downtime
```

### ✅ Multiple Pod Kill Test

```bash
# Kill multiple pods simultaneously
kubectl delete pods -n emzops-prod -l app=devops-backend --field-selector status.phase=Running

# Verify:
# - New pods start
# - Application remains available
# - HPA maintains minimum replicas
```

### ✅ Network Partition Test

```bash
# Simulate network issues (if tools available)
# Or use network policies to isolate pods temporarily

# Verify:
# - Remaining pods continue serving
# - Health checks detect failures
# - Traffic routes to healthy pods
```

## 5. Production vs Staging Comparison

### ✅ Configuration Differences

```bash
# Compare replica counts
echo "Staging replicas:"
kubectl get deployment devops-backend -n emzops-staging -o jsonpath='{.spec.replicas}'
echo ""
echo "Production replicas:"
kubectl get deployment devops-backend -n emzops-prod -o jsonpath='{.spec.replicas}'

# Compare resources
echo "Staging resources:"
kubectl get deployment devops-backend -n emzops-staging -o jsonpath='{.spec.template.spec.containers[0].resources}'
echo ""
echo "Production resources:"
kubectl get deployment devops-backend -n emzops-prod -o jsonpath='{.spec.template.spec.containers[0].resources}'

# Check HPA
echo "Staging HPA:"
kubectl get hpa -n emzops-staging
echo ""
echo "Production HPA:"
kubectl get hpa -n emzops-prod
```

## Checklist Summary

### Production Deployment
- [ ] Production namespace exists (`emzops-prod`)
- [ ] Production secrets configured in GitLab CI/CD
- [ ] Production deployment job runs successfully
- [ ] Production uses `values-prod.yaml`
- [ ] Production has different configuration than staging

### Automatic Scaling
- [ ] HPA created in production namespace
- [ ] HPA shows correct min/max replicas
- [ ] HPA metrics are being collected
- [ ] Autoscaling can be demonstrated (CPU/memory stress)
- [ ] Scaling up works correctly
- [ ] Scaling down works correctly

### Affinity/Topology Constraints
- [ ] Pod anti-affinity configured
- [ ] Pods are distributed across nodes
- [ ] Node failure doesn't take application offline
- [ ] Pods reschedule correctly after node failure

### Chaos Engineering Readiness
- [ ] Pod kill test passes (zero downtime)
- [ ] Multiple pod kill test passes
- [ ] Node failure test passes
- [ ] Application remains available during chaos

## Final Verification

If all checks pass, you can confidently say:

**✅ "Yes, Week 14 is complete!"**

All requirements have been met:
- ✅ Production deployment configured
- ✅ Automatic scaling (HPA) implemented
- ✅ Affinity/topology constraints configured
- ✅ Chaos engineering ready (zero downtime)

