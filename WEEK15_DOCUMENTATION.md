# Week 15 - Final Requirements Documentation (Frontend)

## Security Scanning

### Tool Choice: Trivy

We chose **Trivy** as our container image security scanner. See `devops-be/WEEK15_DOCUMENTATION.md` for detailed explanation.

### Pipeline Integration

**Location**: `devops-fe/.gitlab-ci.yml`

**Stage**: `security-scan` (runs after `build-image`)

**Configuration**: Same as backend - scans frontend container images for vulnerabilities.

## Production Dashboard

Frontend metrics are included in the production Grafana dashboard. See `devops-be/WEEK15_DOCUMENTATION.md` for complete dashboard documentation.

### Frontend-Specific Metrics

- CPU usage
- Memory usage
- Request rate
- Pod status and distribution

