# Week 13 - Templating & Infrastructure as Code Documentation

## Templating Choice

### Tool Selected: Helm

**Why Helm?**
- Helm is the most widely adopted package manager for Kubernetes
- Excellent community support and extensive documentation
- Native Kubernetes integration with `helm` CLI tool
- Supports templating with Go template syntax
- Built-in dependency management and versioning
- Easy rollback capabilities with `--rollback-on-failure` flag
- Values files allow clean separation of configuration and logic

### Helm Chart Structure

Our backend application uses a Helm chart located in `devops-be/chart/`:

```
chart/
├── Chart.yaml          # Chart metadata
├── values.yaml         # Default configuration values
└── templates/          # Kubernetes manifest templates
    ├── deployment.yaml
    ├── service.yaml
    ├── configmap.yaml
    ├── secret.yaml
    └── networkpolicy.yaml
```

### Template Usage

**Chart.yaml** defines the chart metadata:
```yaml
apiVersion: v2
name: devops-backend
description: A Helm chart for the DevOps Backend Application
type: application
version: 0.1.0
appVersion: "1.0.0"
```

**values.yaml** contains default configuration:
- Image repository and tag
- Resource limits and requests
- Replica count
- Service ports
- Secret placeholders (empty by default)

**Templates** use Go template syntax to inject values:
- `{{ .Release.Name }}` - Release name (e.g., "devops-backend")
- `{{ .Values.image.repository }}` - Image repository from values.yaml
- `{{ .Values.replicaCount }}` - Replica count from values.yaml
- `{{ .Values.secrets.backend.databaseUrl | b64enc }}` - Base64 encoded secret

**Example from deployment.yaml:**
```yaml
spec:
  replicas: {{ .Values.replicaCount }}
  template:
    spec:
      containers:
      - name: backend
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
```

### Deployment Process

The CI/CD pipeline uses Helm to deploy:

```bash
helm upgrade --install devops-backend ./chart \
  --namespace emzops-staging \
  --set image.tag=$IMAGE_TAG \
  --set secrets.backend.databaseUrl=$DATABASE_URL \
  --set secrets.postgres.username=$POSTGRES_USER \
  --set secrets.postgres.password=$POSTGRES_PASSWORD \
  --set secrets.backend.minioEndpoint=$MINIO_ENDPOINT \
  --set secrets.backend.minioAccessKey=$MINIO_ACCESS_KEY \
  --set secrets.backend.minioSecretKey=$MINIO_SECRET_KEY \
  --set secrets.backend.minioBucketName=$MINIO_BUCKET_NAME \
  --set secrets.backend.minioUseSSL=$MINIO_USE_SSL \
  --wait \
  --timeout 15m \
  --rollback-on-failure
```

**Key Features:**
- `--upgrade --install`: Creates release if it doesn't exist, upgrades if it does
- `--set`: Overrides values.yaml with pipeline environment variables
- `--wait`: Waits for pods to be ready before completing
- `--timeout 15m`: Maximum wait time
- `--rollback-on-failure`: Automatically rolls back if deployment fails

## Secret Management

### Strategy

**Security Rule:** NEVER commit secrets to Git. All secrets are injected via CI/CD pipeline.

### Implementation

1. **values.yaml** contains empty placeholders:
```yaml
secrets:
  backend:
    databaseUrl: ""
    minioEndpoint: ""
    minioAccessKey: ""
    minioSecretKey: ""
    minioBucketName: ""
    minioUseSSL: ""
  postgres:
    username: ""
    password: ""
```

2. **Secret Template** (templates/secret.yaml) creates Kubernetes Secret:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {{ .Release.Name }}-secret
type: Opaque
data:
  DATABASE_URL: {{ .Values.secrets.backend.databaseUrl | b64enc | quote }}
  MINIO_ENDPOINT: {{ .Values.secrets.backend.minioEndpoint | b64enc | quote }}
  # ... other secrets
```

3. **CI/CD Pipeline** injects secrets via `--set` flags:
   - Secrets are stored as GitLab CI/CD variables (Settings → CI/CD → Variables)
   - Variables are marked as "Protected" and "Masked"
   - Pipeline uses `--set secrets.backend.databaseUrl=$DATABASE_URL` to inject values
   - Secrets are never visible in pipeline logs (masked)

4. **Deployment** uses secrets from Kubernetes Secret:
```yaml
env:
- name: DATABASE_URL
  valueFrom:
    secretKeyRef:
      name: devops-backend-secret
      key: DATABASE_URL
```

### Secret Variables in GitLab CI/CD

The following variables are configured in GitLab CI/CD settings:
- `DATABASE_URL` - PostgreSQL connection string
- `POSTGRES_USER` - PostgreSQL username
- `POSTGRES_PASSWORD` - PostgreSQL password
- `MINIO_ENDPOINT` - MinIO endpoint URL
- `MINIO_ACCESS_KEY` - MinIO access key
- `MINIO_SECRET_KEY` - MinIO secret key
- `MINIO_BUCKET_NAME` - MinIO bucket name
- `MINIO_USE_SSL` - MinIO SSL flag

All variables are:
- **Protected**: Only available in protected branches
- **Masked**: Hidden in pipeline logs
- **Environment-specific**: Can be scoped to specific environments

## Statelessness

### Application State Management

Our application is **fully stateless** and ready for chaos engineering:

#### 1. Files/Images Storage

**MinIO Object Storage** is used for all uploaded images:
- Images are uploaded to MinIO (S3-compatible object storage)
- MinIO is deployed as a separate tenant in the cluster
- Images are stored with unique UUIDs: `{uuid}_{filename}`
- Presigned URLs are generated for image access
- No local filesystem storage is used

**Implementation:**
- `internal/storage/minio.go` handles all MinIO operations
- `UploadImage()` uploads to MinIO bucket
- `DeleteImage()` removes from MinIO
- `GetPresignedURL()` generates temporary access URLs

**Benefits:**
- Images persist across pod restarts
- Multiple replicas can access the same images
- No data loss when pods are killed
- Scalable and distributed storage

#### 2. Data Storage

**PostgreSQL Database** stores all application data:
- User accounts and authentication
- Posts and content
- Votes and interactions
- Image metadata (paths stored in database, actual files in MinIO)

**Database Deployment:**
- PostgreSQL StatefulSet with persistent storage
- Longhorn PVC for data persistence
- Database survives pod restarts and failures
- All application state is in the database

#### 3. Sessions

**JWT (JSON Web Tokens)** are used for authentication:
- Stateless authentication - no server-side session storage
- Tokens are signed and contain user information
- Tokens are stored client-side (browser localStorage)
- No session state on the server

**Implementation:**
- `internal/auth/jwt.go` handles token generation and validation
- Tokens expire after 24 hours
- No Redis or session store needed
- Fully stateless authentication

### Verification

**Pod Failure Test:**
1. Delete a backend pod: `kubectl delete pod <pod-name> -n emzops-staging`
2. New pod starts automatically
3. Application continues working:
   - Images are still accessible (stored in MinIO)
   - User data is intact (stored in PostgreSQL)
   - Authentication works (JWT tokens are stateless)
   - No data loss occurs

**Multi-Replica Test:**
1. Scale deployment: `kubectl scale deployment devops-backend --replicas=2 -n emzops-staging`
2. Both pods can:
   - Access the same database
   - Serve the same images from MinIO
   - Handle requests independently
   - No shared state between pods

## Summary

✅ **Templating**: Helm charts with proper separation of logic and configuration  
✅ **Atomic Rollback**: `--rollback-on-failure` flag ensures automatic rollback on failure  
✅ **Secret Management**: All secrets injected via CI/CD, never committed to Git  
✅ **Statelessness**: 
   - Images in MinIO (object storage)
   - Data in PostgreSQL (database)
   - Sessions via JWT (stateless tokens)
   - Ready for chaos engineering

