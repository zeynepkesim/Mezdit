# Kubernetes Setup Guide for Team Members

This guide explains how to set up Kubernetes deployment for the backend application (Week 10).

## 🎯 What We've Already Done

The following resources have been created and configured:
- ✅ All Kubernetes manifest files in `k8s/` directory
- ✅ GitLab CI/CD pipeline for automated deployment
- ✅ RBAC resources (ServiceAccount, Role, RoleBinding)
- ✅ GitLab Deploy Token for container registry
- ✅ Secrets and ConfigMaps in the cluster
- ✅ PostgreSQL database deployed
- ✅ Backend application deployed
- ✅ Ingress with TLS certificate

**Application URL:** https://devops-mezdit.kubernetes.devops.cs.ut.ee

## 🚀 What You Need to Do (Team Members)

### 1. Install Tools

```bash
# Install kubectl (if not already installed)
brew install kubectl

# Install kubelogin plugin
brew install int128/kubelogin/kubelogin
```

### 2. Configure kubectl Access

```bash
# Create kubectl directory
mkdir -p ~/.kube

# Copy the KUBECONFIG file provided by instructors to ~/.kube/config-ut

# Set environment variable
export KUBECONFIG=~/.kube/config-ut

# Make it permanent (add to ~/.zshrc or ~/.bashrc)
echo 'export KUBECONFIG=~/.kube/config-ut' >> ~/.zshrc

# Verify access
kubectl get namespaces
kubectl get pods -n emzops-staging
```

### 3. View Deployed Resources

```bash
# See all pods
kubectl get pods -n emzops-staging

# See deployments
kubectl get deployments -n emzops-staging

# See services
kubectl get services -n emzops-staging

# See ingress
kubectl get ingress -n emzops-staging

# View all resources at once
kubectl get all -n emzops-staging
```

### 4. Check Application Logs

```bash
# Backend application logs
kubectl logs -f deployment/devops-backend -n emzops-staging

# PostgreSQL logs
kubectl logs -f postgres-0 -n emzops-staging

# Get pod name and view logs
kubectl get pods -n emzops-staging
kubectl logs -f <pod-name> -n emzops-staging
```

### 5. Local Testing (Port Forwarding)

```bash
# Forward backend service to localhost
kubectl port-forward service/devops-backend-service 8080:80 -n emzops-staging

# In another terminal, test the application
curl http://localhost:8080/health
```

## 🔐 Secret Files (Not in Git)

Two secret files are NOT committed to git (they're in `.gitignore`):
- `k8s/secret.yaml` - Backend secrets (DATABASE_URL)
- `k8s/postgres-secret.yaml` - PostgreSQL password

**Templates are provided:**
- `k8s/secret.yaml.template`
- `k8s/postgres-secret.yaml.template`

If you need to recreate them locally:

```bash
cd k8s
cp secret.yaml.template secret.yaml
cp postgres-secret.yaml.template postgres-secret.yaml

# Edit the files with actual passwords
# (Ask team for the password we're using)
```

Then apply to cluster:

```bash
kubectl apply -f k8s/postgres-secret.yaml
kubectl apply -f k8s/secret.yaml
```

## 📦 CI/CD Pipeline

The GitLab CI/CD pipeline automatically deploys to Kubernetes when you push to `main` branch.

**Pipeline stages:**
1. Security scan (gitleaks)
2. Tests (go test)
3. Build (compile)
4. Build Docker image
5. **Deploy to Kubernetes** (new!)

**GitLab Variable configured:**
- `KUBECONFIG_CONTENT` - Base64 encoded kubeconfig for cluster access

## 🛠️ Useful Commands

### Debugging

```bash
# Describe a pod (shows events and issues)
kubectl describe pod <pod-name> -n emzops-staging

# Get events (troubleshooting)
kubectl get events -n emzops-staging --sort-by='.lastTimestamp'

# Execute command in pod
kubectl exec -it <pod-name> -n emzops-staging -- /bin/sh

# Connect to PostgreSQL
kubectl exec -it postgres-0 -n emzops-staging -- psql -U devops -d devops
```

### Restarting Resources

```bash
# Restart deployment (creates new pods)
kubectl rollout restart deployment/devops-backend -n emzops-staging

# Delete pod (will be recreated automatically)
kubectl delete pod <pod-name> -n emzops-staging
```

### Viewing Secrets and ConfigMaps

```bash
# List secrets
kubectl get secrets -n emzops-staging

# View secret content (base64 encoded)
kubectl get secret backend-secret -n emzops-staging -o yaml

# Decode a secret value
kubectl get secret backend-secret -n emzops-staging -o jsonpath='{.data.DATABASE_URL}' | base64 -d

# View ConfigMap
kubectl get configmap backend-config -n emzops-staging -o yaml
```

## 📁 Project Structure

```
k8s/
├── configmap.yaml               # Environment variables (non-sensitive)
├── deployment.yaml              # Backend deployment
├── ingress.yaml                 # Ingress with TLS
├── postgres-statefulset.yaml    # PostgreSQL database
├── postgres-secret.yaml.template # Template for DB password
├── rbac.yaml                    # ServiceAccount, Role, RoleBinding
├── secret.yaml.template         # Template for app secrets
└── service.yaml                 # Backend service
```

## ⚠️ Important Notes

1. **Never commit secret files** - they're in `.gitignore`
2. **Use the same password** in both `postgres-secret.yaml` and `secret.yaml`
3. **GitLab CI/CD** handles deployment automatically
4. **Namespace:** All resources are in `emzops-staging`
5. **TLS certificate** is auto-provisioned by cert-manager (1-2 minutes)

## 🆘 Troubleshooting

### "No resources found" error

```bash
# Make sure you're using the right namespace
kubectl get pods -n emzops-staging

# Check if namespace exists
kubectl get namespaces | grep emzops
```

### "Forbidden" errors

```bash
# Check your KUBECONFIG
echo $KUBECONFIG

# Verify it's set correctly
kubectl config current-context
```

### Pod not starting

```bash
# Check pod status
kubectl get pods -n emzops-staging

# Describe the pod for details
kubectl describe pod <pod-name> -n emzops-staging

# Check logs
kubectl logs <pod-name> -n emzops-staging
```

### Database connection issues

```bash
# Check if PostgreSQL is running
kubectl get pods -n emzops-staging | grep postgres

# Test PostgreSQL connection
kubectl exec -it postgres-0 -n emzops-staging -- psql -U devops -d devops -c "SELECT 1;"
```

## 📚 Additional Resources

- [Kubernetes Docs](https://kubernetes.io/docs/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- Course Week 10 Instructions: https://devops.cs.ut.ee/phase3/week10/

## 💡 Tips

1. Use `-w` flag to watch resources: `kubectl get pods -n emzops-staging -w`
2. Use `-f` flag to follow logs: `kubectl logs -f <pod-name> -n emzops-staging`
3. Use `describe` to see detailed info: `kubectl describe pod <pod-name> -n emzops-staging`
4. Use tab completion for kubectl (add to your shell config)

---

**Questions?** Ask in the team chat or check the course instructions.
