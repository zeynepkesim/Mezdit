# Week 13 Testing Guide

This guide helps you verify that all Week 13 requirements are met.

## Prerequisites

1. Helm installed locally: `helm version`
2. kubectl configured: `kubectl get pods -n emzops-staging`
3. Access to GitLab CI/CD pipelines

## 1. Templating Verification

### ✅ Check Helm Charts Exist

```bash
# Backend chart
ls -la devops-be/chart/
# Should show: Chart.yaml, values.yaml, templates/

# Frontend chart
ls -la devops-fe/chart/
# Should show: Chart.yaml, values.yaml, templates/
```

### ✅ Verify Chart Structure

```bash
# Backend
helm template devops-backend devops-be/chart --dry-run
# Should render Kubernetes manifests without errors

# Frontend
helm template devops-frontend devops-fe/chart --dry-run
# Should render Kubernetes manifests without errors
```

### ✅ Check Values Separation

```bash
# Verify values.yaml contains configuration (not logic)
cat devops-be/chart/values.yaml
cat devops-fe/chart/values.yaml

# Verify templates contain logic (not hardcoded values)
grep -r "{{" devops-be/chart/templates/
grep -r "{{" devops-fe/chart/templates/
```

## 2. Atomic Rollback Verification

### ✅ Check CI/CD Pipeline Configuration

**Backend (.gitlab-ci.yml):**
```bash
grep -A 5 "helm upgrade" devops-be/.gitlab-ci.yml
# Should show: --wait --timeout 15m --rollback-on-failure
```

**Frontend (.gitlab-ci.yml):**
```bash
grep -A 5 "helm upgrade" devops-fe/.gitlab-ci.yml
# Should show: --wait --timeout 10m --rollback-on-failure
```

### ✅ Test Atomic Rollback

**Simulate a failed deployment:**

1. Deploy with a broken image tag:
```bash
helm upgrade devops-backend devops-be/chart \
  --namespace emzops-staging \
  --set image.tag=broken-tag \
  --wait \
  --timeout 5m \
  --rollback-on-failure
```

2. Verify rollback:
```bash
# Check release history
helm history devops-backend -n emzops-staging

# Should show:
# - Previous release (working)
# - Failed release (broken-tag)
# - Rollback to previous release
```

3. Verify pods are running:
```bash
kubectl get pods -n emzops-staging -l app=devops-backend
# Pods should be Running with the previous working image
```

## 3. Secret Management Verification

### ✅ Verify No Secrets in Git

```bash
# Check values.yaml for secrets
grep -i "password\|secret\|key" devops-be/chart/values.yaml
# Should show empty strings: ""

# Check templates for hardcoded secrets
grep -r "password\|secret\|key" devops-be/chart/templates/ --exclude="*.yaml"
# Should not show actual secret values
```

### ✅ Verify Secrets Injected via CI/CD

1. Check GitLab CI/CD variables:
   - Go to GitLab project → Settings → CI/CD → Variables
   - Verify these variables exist:
     - `DATABASE_URL`
     - `POSTGRES_USER`
     - `POSTGRES_PASSWORD`
     - `MINIO_ACCESS_KEY`
     - `MINIO_SECRET_KEY`
   - All should be marked as "Protected" and "Masked"

2. Check pipeline logs:
   - Run a pipeline
   - Verify secrets are not visible in logs (masked)
   - Verify `--set` flags are used to inject secrets

### ✅ Verify Kubernetes Secrets

```bash
# Check secret exists
kubectl get secret devops-backend-secret -n emzops-staging

# Verify secret contains data (not empty)
kubectl get secret devops-backend-secret -n emzops-staging -o yaml
# Should show base64 encoded values in data: section
```

## 4. Statelessness Verification

### ✅ Verify Images Stored in MinIO

1. Check MinIO is deployed:
```bash
kubectl get pods -n emzops-staging | grep minio
# Should show MinIO pods running
```

2. Check backend uses MinIO:
```bash
# Check environment variables
kubectl get deployment devops-backend -n emzops-staging -o yaml | grep MINIO
# Should show MINIO_ENDPOINT, MINIO_ACCESS_KEY, etc.

# Check code uses MinIO
grep -r "minio\|MinIO" devops-be/internal/storage/
# Should show MinIO client usage
```

3. Test image upload:
   - Upload an image via the application
   - Delete the backend pod
   - Verify image is still accessible after pod restart

### ✅ Verify Database Storage

1. Check PostgreSQL is deployed:
```bash
kubectl get statefulset postgres -n emzops-staging
kubectl get pods -n emzops-staging | grep postgres
```

2. Test data persistence:
   - Create a post via the application
   - Delete the backend pod
   - Verify post still exists after pod restart

### ✅ Verify Stateless Authentication

1. Check JWT usage:
```bash
grep -r "JWT\|jwt" devops-be/internal/auth/
# Should show JWT token generation/validation
```

2. Test authentication:
   - Login to the application
   - Delete the backend pod
   - Verify you're still logged in (JWT token is stateless)

### ✅ Test Pod Failure Recovery

1. Delete a pod:
```bash
# Get pod name
POD=$(kubectl get pods -n emzops-staging -l app=devops-backend -o jsonpath='{.items[0].metadata.name}')

# Delete pod
kubectl delete pod $POD -n emzops-staging
```

2. Verify recovery:
```bash
# Watch new pod start
kubectl get pods -n emzops-staging -l app=devops-backend -w

# Verify application still works:
# - Images are accessible
# - Posts are visible
# - Authentication works
```

### ✅ Test Multi-Replica Deployment

1. Scale deployment:
```bash
kubectl scale deployment devops-backend --replicas=2 -n emzops-staging
```

2. Verify both pods work:
```bash
kubectl get pods -n emzops-staging -l app=devops-backend
# Should show 2 pods Running

# Test application:
# - Upload image (should work from any pod)
# - Create post (should work from any pod)
# - Access images (should work from any pod)
```

## 5. Pipeline Verification

### ✅ Check Pipeline Success

1. Run backend pipeline:
   - Push code to main branch
   - Check pipeline stages:
     - ✅ Security (secrets detection)
     - ✅ Test
     - ✅ Build
     - ✅ Build Image
     - ✅ Deploy (should use Helm)

2. Run frontend pipeline:
   - Push code to main branch
   - Check pipeline stages:
     - ✅ Security (secrets detection)
     - ✅ Build
     - ✅ Build Image
     - ✅ Deploy (should use Helm)

### ✅ Verify Deployment Commands

Check pipeline logs for:
- `helm upgrade --install` command
- `--wait` flag present
- `--timeout` flag present
- `--rollback-on-failure` flag present
- `--set` flags for secrets

## 6. Documentation Verification

### ✅ Check Documentation Exists

```bash
# Backend documentation
ls devops-be/WEEK13_DOCUMENTATION.md

# Frontend documentation
ls devops-fe/WEEK13_DOCUMENTATION.md
```

### ✅ Verify Documentation Content

Documentation should include:
- ✅ Templating choice explanation (Helm)
- ✅ Template tool usage description
- ✅ Secret management strategy
- ✅ Statelessness confirmation
- ✅ Image/file storage location (MinIO)
- ✅ Database deployment details (PostgreSQL)

## Checklist Summary

### Operational Requirements
- [x] Helm installed locally
- [x] Charts created (backend and frontend)
- [x] Templates use Go template syntax
- [x] Values.yaml contains configuration
- [x] `--wait` flag in CI/CD
- [x] `--timeout` flag in CI/CD
- [x] `--rollback-on-failure` flag in CI/CD
- [x] Secrets injected via `--set` flags
- [x] No secrets in Git

### Development Requirements
- [x] Images stored in MinIO (object storage)
- [x] Data stored in PostgreSQL (database)
- [x] Sessions use JWT (stateless)
- [x] Pod failure recovery tested
- [x] Multi-replica deployment tested

### Documentation Requirements
- [x] Templating choice documented
- [x] Secret management strategy documented
- [x] Statelessness confirmed and documented
- [x] Storage locations documented

## Final Verification

If all checks pass, you can confidently say:

**✅ "Yes, Week 13 is complete!"**

All requirements have been met:
- ✅ Helm templating implemented
- ✅ Atomic rollback configured
- ✅ Secrets managed securely
- ✅ Application is stateless
- ✅ Documentation complete

