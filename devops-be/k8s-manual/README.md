# Manual Kubernetes Resources

This directory contains Kubernetes resources that need to be deployed manually (not via Helm/GitLab CI/CD).

## MinIO Tenant

The MinIO tenant must be deployed manually because the GitLab service account doesn't have permissions to manage MinIO CRDs.

### Prerequisites

- MinIO Operator must be installed in the cluster
- `kubectl` access to the cluster

### Deployment

```bash
# 1. Create MinIO secret (contains root credentials)
kubectl apply -f minio-secret.yaml

# 2. Create MinIO tenant
kubectl apply -f minio-tenant.yaml

# 3. Verify deployment
kubectl get tenant -n emzops-staging
kubectl get pods -n emzops-staging | grep minio
```

### MinIO Connection Details

**For GitLab CI/CD Variables:**

```
MINIO_ENDPOINT=devops-backend-minio-hl.emzops-staging.svc.cluster.local:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_BUCKET_NAME=images
MINIO_USE_SSL=false
```

**Services:**
- S3 API: `devops-backend-minio-hl.emzops-staging.svc.cluster.local:9000`
- Console: `devops-backend-minio-console.emzops-staging.svc.cluster.local:9443`

### Storage

- **StorageClass:** `longhorn-single` (1 replica to avoid double replication)
- **Size:** 5Gi
- **Access Mode:** ReadWriteOnce

### Resources

- **CPU Request:** 100m
- **Memory Request:** 256Mi
- **CPU Limit:** 500m
- **Memory Limit:** 512Mi

## Notes

- The MinIO tenant is deployed once and persists across backend deployments
- Credentials are stored in `devops-backend-minio-env-config` secret
- Backend application connects via S3-compatible API
- Bucket is created automatically on first connection
