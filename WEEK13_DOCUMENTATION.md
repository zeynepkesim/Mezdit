# Week 13 - Templating & Infrastructure as Code Documentation (Frontend)

## Templating Choice

### Tool Selected: Helm

**Why Helm?**
- Consistent with backend deployment strategy
- Industry-standard Kubernetes package manager
- Excellent templating capabilities
- Built-in rollback support
- Easy configuration management with values files

### Helm Chart Structure

Our frontend application uses a Helm chart located in `devops-fe/chart/`:

```
chart/
├── Chart.yaml          # Chart metadata
├── values.yaml         # Default configuration values
└── templates/          # Kubernetes manifest templates
    ├── deployment.yaml
    ├── service.yaml
    ├── ingress.yaml
    └── configmap.yaml
```

### Template Usage

**Chart.yaml** defines the chart metadata:
```yaml
apiVersion: v2
name: devops-frontend
description: A Helm chart for the DevOps Frontend Application
type: application
version: 0.1.0
appVersion: "1.0.0"
```

**values.yaml** contains default configuration:
- Image repository and tag
- Resource limits and requests
- Replica count
- Service configuration
- Ingress settings
- API configuration (API URL, WebSocket URL)

**Templates** use Go template syntax:
- `{{ .Release.Name }}` - Release name (e.g., "devops-frontend")
- `{{ .Values.image.repository }}` - Image repository
- `{{ .Values.replicaCount }}` - Number of replicas
- `{{ .Values.ingress.host }}` - Ingress hostname

**Example from deployment.yaml:**
```yaml
spec:
  replicas: {{ .Values.replicaCount | default 1 }}
  template:
    spec:
      containers:
      - name: frontend
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
```

### Deployment Process

The CI/CD pipeline uses Helm to deploy:

```bash
helm upgrade --install devops-frontend ./chart \
  --namespace emzops-staging \
  --set image.tag=$IMAGE_TAG \
  --wait \
  --timeout 10m \
  --rollback-on-failure
```

**Key Features:**
- `--upgrade --install`: Creates or upgrades the release
- `--set image.tag=$IMAGE_TAG`: Sets image tag from CI/CD variable
- `--wait`: Waits for pods to be ready
- `--timeout 10m`: Maximum wait time
- `--rollback-on-failure`: Automatically rolls back on failure

## Secret Management

### Strategy

**Security Rule:** Frontend does not require secrets as it's a static application served by NGINX.

**Configuration Management:**
- API URLs are configured via ConfigMap (not secrets)
- ConfigMap is created from `values.yaml`:
```yaml
config:
  apiUrl: "/api"
  wsUrl: "/ws"
```

- ConfigMap template:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-config
data:
  VUE_APP_API_URL: {{ .Values.config.apiUrl | quote }}
  VUE_APP_WS_URL: {{ .Values.config.wsUrl | quote }}
```

- No secrets are needed for frontend deployment
- All configuration is environment-agnostic and safe to commit

## Statelessness

### Application State Management

The frontend application is **fully stateless**:

#### 1. Static Files

**NGINX serves static files:**
- Built Vue.js application is compiled to static HTML/CSS/JS
- Files are served from `/usr/share/nginx/html` in the container
- No server-side state or session storage
- All files are part of the Docker image

**Build Process:**
- `npm run build` creates static files in `dist/` directory
- Files are copied into NGINX image during Docker build
- No runtime file generation or modification

#### 2. Client-Side State

**Vue.js application state:**
- All state is managed client-side (browser)
- JWT tokens stored in localStorage
- No server-side session management
- Application state is lost on page refresh (intentional for security)

#### 3. No Backend State

**Frontend is a pure client application:**
- No database connections
- No file uploads (handled by backend)
- No server-side processing
- Only serves static files via NGINX

### Verification

**Pod Failure Test:**
1. Delete a frontend pod: `kubectl delete pod <pod-name> -n emzops-staging`
2. New pod starts automatically
3. Application continues working:
   - Static files are served from the new pod
   - No data loss (all state is client-side)
   - Users can refresh and continue using the app

**Multi-Replica Test:**
1. Scale deployment: `kubectl scale deployment devops-frontend --replicas=3 -n emzops-staging`
2. All pods serve identical static files
3. Load balancer distributes traffic
4. No shared state between pods needed

## Summary

✅ **Templating**: Helm charts with proper separation of logic and configuration  
✅ **Atomic Rollback**: `--rollback-on-failure` flag ensures automatic rollback on failure  
✅ **Secret Management**: No secrets needed (static application)  
✅ **Statelessness**: 
   - Static files served by NGINX
   - Client-side state only
   - No server-side state
   - Ready for chaos engineering

