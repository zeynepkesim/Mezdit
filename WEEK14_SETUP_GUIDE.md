# Week 14 Setup Guide

## Quick Start Checklist

### 1. Production Namespace Setup

```bash
# Check if production namespace exists
kubectl get namespace emzops-prod

# ✅ Namespace already exists!
# If it doesn't exist, Helm will create it automatically with --create-namespace flag
```

### 2. Production Secrets Setup

**In GitLab CI/CD:**
✅ **No additional variables needed!** 

The pipeline uses the existing CI/CD variables:
- `DATABASE_URL`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `MINIO_ENDPOINT`
- `MINIO_ACCESS_KEY`
- `MINIO_SECRET_KEY`
- `MINIO_BUCKET_NAME`
- `MINIO_USE_SSL`

**Note:** Both staging and production use the same secrets. If you need production-specific secrets later, you can add `_PROD` suffixed variables and update the pipeline.

### 3. ImagePullSecret Setup

```bash
# Copy gitlab-registry-secret from staging to production
kubectl get secret gitlab-registry-secret -n emzops-staging -o yaml | \
  sed 's/namespace: emzops-staging/namespace: emzops-prod/' | \
  kubectl apply -f -

# Verify
kubectl get secret gitlab-registry-secret -n emzops-prod
```

### 4. Deploy to Production

**Option 1: Via GitLab CI/CD (Recommended)**
1. Push code to `main` branch
2. Go to GitLab CI/CD → Pipelines
3. Find the latest pipeline
4. Click "Play" button on `deploy-to-production` job (manual trigger)

**Option 2: Manual Helm Deploy**
```bash
# Backend
helm upgrade --install devops-backend ./devops-be/chart \
  --namespace emzops-prod \
  --values ./devops-be/chart/values-prod.yaml \
  --set image.tag=<your-image-tag> \
  --set secrets.backend.databaseUrl=$DATABASE_URL_PROD \
  # ... other secrets

# Frontend
helm upgrade --install devops-frontend ./devops-fe/chart \
  --namespace emzops-prod \
  --values ./devops-fe/chart/values-prod.yaml \
  --set image.tag=<your-image-tag>
```

### 5. Verify Deployment

```bash
# Check pods
kubectl get pods -n emzops-prod

# Check HPA
kubectl get hpa -n emzops-prod

# Check pod distribution
kubectl get pods -n emzops-prod -o wide
```

## What Was Implemented

### ✅ Production Deployment
- **values-prod.yaml** files created for both backend and frontend
- Production-specific configurations (higher replicas, resources, autoscaling)
- CI/CD pipeline updated with production deployment job (manual trigger)

### ✅ Automatic Scaling (HPA)
- HPA templates created (`chart/templates/hpa.yaml`)
- Configured for easy demonstration (low CPU/memory thresholds: 30%/50%)
- Scaling range: 2-5 replicas
- Enabled only in production

### ✅ Affinity/Topology Constraints
- Pod anti-affinity configured in production values
- Ensures pods are distributed across different nodes
- Prevents single node failure from taking application offline

### ✅ Documentation
- Backend documentation: `devops-be/WEEK14_DOCUMENTATION.md`
- Frontend documentation: `devops-fe/WEEK14_DOCUMENTATION.md`
- Testing guide: `WEEK14_TESTING_GUIDE.md`

## Key Files Created/Modified

### Backend
- `devops-be/chart/values-prod.yaml` - Production values
- `devops-be/chart/templates/hpa.yaml` - HPA template
- `devops-be/chart/templates/deployment.yaml` - Added affinity support
- `devops-be/.gitlab-ci.yml` - Added production deployment job
- `devops-be/WEEK14_DOCUMENTATION.md` - Documentation

### Frontend
- `devops-fe/chart/values-prod.yaml` - Production values
- `devops-fe/chart/templates/hpa.yaml` - HPA template
- `devops-fe/chart/templates/deployment.yaml` - Added affinity support
- `devops-fe/.gitlab-ci.yml` - Added production deployment job
- `devops-fe/WEEK14_DOCUMENTATION.md` - Documentation

## Testing Autoscaling

### Quick Demo (CPU Stress)

```bash
# 1. Watch HPA
watch kubectl get hpa -n emzops-prod

# 2. In another terminal, generate CPU load
POD=$(kubectl get pods -n emzops-prod -l app=devops-backend -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it $POD -n emzops-prod -- sh -c "apk add stress-ng && stress-ng --cpu 2 --timeout 300s"

# 3. Watch pods scale up (2 → 3 → 4 → 5)
watch kubectl get pods -n emzops-prod -l app=devops-backend

# 4. After stress stops, watch pods scale down (5 → 4 → 3 → 2)
```

## Next Steps

1. ✅ Create production namespace
2. ✅ Configure production secrets in GitLab CI/CD
3. ✅ Copy ImagePullSecret to production namespace
4. ✅ Deploy to production via pipeline
5. ✅ Test autoscaling
6. ✅ Test affinity constraints
7. ✅ Verify chaos engineering readiness

## Troubleshooting

**HPA not scaling:**
- Check metrics-server is installed: `kubectl get apiservice v1beta1.metrics.k8s.io`
- Check HPA status: `kubectl describe hpa -n emzops-prod`
- Verify resource requests/limits are set in deployment

**Pods not distributed across nodes:**
- Check node count: `kubectl get nodes`
- Verify affinity configuration: `kubectl get deployment -n emzops-prod -o yaml | grep -A 10 affinity`
- Check if nodes have required labels: `kubectl get nodes --show-labels`

**Production deployment fails:**
- Check namespace exists: `kubectl get namespace emzops-prod`
- Check secrets exist: `kubectl get secrets -n emzops-prod`
- Check ImagePullSecret: `kubectl get secret gitlab-registry-secret -n emzops-prod`

