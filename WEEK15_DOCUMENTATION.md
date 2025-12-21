# Week 15 - Final Requirements Documentation

## Security Scanning

### Tool Choice: Trivy

We chose **Trivy** as our container image security scanner for the following reasons:

1. **Comprehensive Scanning**: Trivy scans for vulnerabilities in OS packages, application dependencies, and configuration files
2. **GitLab Integration**: Trivy has excellent GitLab CI/CD integration with native artifact reports
3. **Fast Performance**: Trivy is fast and doesn't require a database, making it ideal for CI/CD pipelines
4. **Multiple Formats**: Supports various output formats including GitLab's security scanning format
5. **Active Maintenance**: Actively maintained by Aqua Security with regular vulnerability database updates

### Pipeline Integration

#### Backend Pipeline

**Location**: `devops-be/.gitlab-ci.yml`

**Stage**: `security-scan` (runs after `build-image`)

**Job Configuration**:
```yaml
container-security-scan:
  stage: security-scan
  image:
    name: aquasec/trivy:latest
    entrypoint: [""]
  variables:
    TRIVY_NO_PROGRESS: "true"
    TRIVY_SKIP_UPDATE: "false"
    TRIVY_FORMAT: "template"
    TRIVY_TEMPLATE: "@contrib/gitlab.tpl"
    TRIVY_OUTPUT: "gl-container-scanning-report.json"
    TRIVY_EXIT_CODE: "0"
    TRIVY_IGNORE_UNFIXED: "false"
    TRIVY_SEVERITY: "CRITICAL,HIGH"
  script:
    - echo "🔒 Scanning backend container image for vulnerabilities..."
    - trivy image --exit-code $TRIVY_EXIT_CODE --severity $TRIVY_SEVERITY --format $TRIVY_FORMAT --template $TRIVY_TEMPLATE -o $TRIVY_OUTPUT $IMAGE_NAME:$IMAGE_TAG
    - echo "✅ Security scan completed for $IMAGE_NAME:$IMAGE_TAG"
  artifacts:
    reports:
      container_scanning: gl-container-scanning-report.json
    paths:
      - gl-container-scanning-report.json
    expire_in: 1 week
  dependencies:
    - build-docker-image
  only:
    - main
```

**How it Works**:
1. Runs automatically after `build-docker-image` job completes
2. Scans the newly built image (`$IMAGE_NAME:$IMAGE_TAG`)
3. Generates a GitLab-compatible security report
4. Reports are visible in GitLab's Security tab
5. Scans for CRITICAL and HIGH severity vulnerabilities

#### Frontend Pipeline

**Location**: `devops-fe/.gitlab-ci.yml`

**Configuration**: Same as backend, but scans frontend images

### Scan Execution

The security scan runs automatically:
- **When**: After every image build on `main` branch
- **What**: Scans the container image for known vulnerabilities
- **Severity**: CRITICAL and HIGH vulnerabilities
- **Output**: GitLab security scanning report

### Vulnerability Handling

**Current Approach**:
- Scans are configured with `TRIVY_EXIT_CODE: "0"` to not fail the pipeline
- This allows us to review vulnerabilities without blocking deployments
- Reports are available in GitLab's Security tab for review

**Future Improvements**:
1. **Critical Vulnerabilities**: Should block deployment if critical vulnerabilities are found
2. **Automated Patching**: Integrate with Dependabot or similar for automatic dependency updates
3. **Base Image Updates**: Regularly update base images (golang:1.23-alpine, node:22-alpine)
4. **Vulnerability Triage**: Create a process for reviewing and addressing vulnerabilities

**Example Handling Process**:
1. Review scan results in GitLab Security tab
2. Identify vulnerable dependencies
3. Update dependencies to patched versions
4. Rebuild and rescan
5. Verify vulnerabilities are resolved

## Production Dashboard

### Dashboard Overview

**Location**: Grafana at https://monitoring.kubernetes.devops.cs.ut.ee

**Namespace**: `emzops-prod`

**Purpose**: Monitor production application health, performance, and availability

### Dashboard Metrics

Our production dashboard includes the following key metrics:

#### 1. Application Health Overview
- **Backend Pods**: Current ready replicas vs desired
- **Frontend Pods**: Current ready replicas vs desired
- **PostgreSQL**: StatefulSet pod status
- **Overall Health**: Aggregated health status

#### 2. Pod Status & Chaos Engineering
- **Current Pod Status**: Running, Pending, CrashLoopBackOff, etc.
- **Pod Restarts**: Time series showing pod restarts
- **Pod Lifecycle Events**: Pod kills, restarts, scheduling events
- **Node Distribution**: Which nodes pods are running on

#### 3. Autoscaling Visibility
- **HPA Current Replicas**: Current number of replicas for each deployment
- **HPA Desired Replicas**: Target number of replicas
- **HPA Metrics**: CPU and Memory utilization vs targets
- **Scale Events**: Timeline showing when HPA scales up/down

#### 4. Critical Components

**Backend**:
- CPU usage (requests vs limits)
- Memory usage (requests vs limits)
- HTTP request rate
- Error rate (4xx, 5xx)
- Response time (p50, p95, p99)

**Frontend**:
- CPU usage
- Memory usage
- Request rate

**PostgreSQL**:
- CPU usage
- Memory usage
- Pod status
- Connection count (if available)

#### 5. Golden Signals with Thresholds

**Latency**:
- Request duration (p50, p95, p99)
- Thresholds: Warning > 500ms, Critical > 1s

**Traffic**:
- Request rate (requests/second)
- By HTTP method (GET, POST, etc.)

**Errors**:
- Error rate (4xx, 5xx)
- Thresholds: Warning > 1%, Critical > 5%

**Saturation**:
- CPU saturation (usage / limits)
- Memory saturation (usage / limits)
- Thresholds: Warning > 70%, Critical > 90%

#### 6. Resource Usage with Thresholds
- **CPU**: Usage vs requests vs limits (visual thresholds)
- **Memory**: Usage vs requests vs limits (visual thresholds)
- **Visual Representation**: Colored zones (green/yellow/red)

#### 7. Service State
- **Current Status**: Healthy/Unhealthy indicators
- **Time Series**: Service availability over last 6 hours
- **Uptime Percentage**: Overall service uptime

### Dashboard Usage for Issue Diagnosis

#### Scenario 1: High Error Rate
1. Check **Golden Signals - Errors** panel
2. Identify error type (4xx vs 5xx)
3. Check **Backend Metrics** for specific endpoints
4. Review **Pod Status** for pod restarts or crashes
5. Check **Resource Usage** for resource exhaustion

#### Scenario 2: Slow Response Times
1. Check **Golden Signals - Latency** panel
2. Identify which percentiles are affected (p95, p99)
3. Review **Resource Usage** for CPU/Memory saturation
4. Check **HPA Autoscaling** - may need to scale up
5. Review **Backend Metrics** for specific slow endpoints

#### Scenario 3: Pod Killed by Chaos Engineering
1. Check **Pod Status & Chaos Engineering** panel
2. See pod status change (Running → Terminating → Pending → Running)
3. Verify new pod starts automatically
4. Check **Service State** - should remain healthy
5. Review **Resource Usage** - new pod should handle traffic

#### Scenario 4: Autoscaling Not Working
1. Check **HPA Autoscaling** panel
2. Verify HPA metrics (CPU/Memory) vs targets
3. Check if metrics are being collected
4. Review **Resource Usage** - should show high utilization
5. Check HPA events: `kubectl describe hpa -n emzops-prod`

#### Scenario 5: Service Unavailable
1. Check **Application Health Overview** - pods ready?
2. Check **Service State** - current status and time series
3. Review **Pod Status** - any pods in CrashLoopBackOff?
4. Check **PostgreSQL** - database available?
5. Review **Golden Signals** - any errors or latency spikes?

### Dashboard Best Practices

1. **Concise Titles**: Panel titles are short and descriptive
2. **Readable Labels**: All labels are clear and understandable
3. **Visual Thresholds**: Resource usage shows colored zones
4. **Time Series**: Service state shows historical data (6 hours)
5. **Golden Signals**: All four signals (latency, traffic, errors, saturation) with thresholds

## Summary

✅ **Security Scanning**: Trivy integrated into both backend and frontend pipelines
✅ **Production Dashboard**: Comprehensive Grafana dashboard for `emzops-prod` namespace
✅ **Documentation**: Complete documentation of scanning and dashboard implementation

Both requirements are implemented and ready for production use.

