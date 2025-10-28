# Week 7 - Security, Secrets, and Deployment Automation

## Overview

This document describes the security improvements, secrets management, and deployment automation changes implemented for Week 7 requirements.

---

## 🔒 Security Improvements

### XSS (Cross-Site Scripting) Prevention

#### Backend Implementation

**New Security Package**: `internal/security/sanitization.go`

We implemented comprehensive input sanitization to prevent XSS attacks:

**Features:**
- **HTML Escaping**: All user inputs are escaped to prevent script injection
- **Pattern Filtering**: Removes dangerous patterns (script tags, event handlers, javascript: protocol)
- **Username Sanitization**: Only allows alphanumeric characters, underscores, and hyphens
- **Length Validation**: Prevents DoS attacks through oversized inputs

**Sanitization Functions:**
```go
// Sanitizes general input
SanitizeInput(input string) string

// Sanitizes text content (posts, comments)
SanitizeText(text string) string

// Sanitizes username (alphanumeric only)
SanitizeUsername(username string) string

// Validates length limits
ValidateLength(input string, maxLength int) bool
```

**Protected Endpoints:**
- `/posts` (POST) - Post creation
- `/posts/edit` (POST) - Post editing
- `/posts/upload` (POST) - Post with image
- `/posts/edit-with-image` (POST) - Edit with image
- `/login` (POST) - User login
- `/register` (POST) - User registration
- `/change-password` (POST) - Password change

**Input Validation Limits:**
- Username: max 100 characters
- Name: max 255 characters
- Post Header: max 500 characters
- Post Content: max 5000 characters
- Password: max 128 characters

#### Frontend Implementation

**New Validation Utilities**: `src/utils/validation.js`

**Features:**
- HTML escaping using browser's built-in mechanisms
- Pattern detection for dangerous content
- Client-side validation before sending to backend
- Length validation matching backend limits

**Validation Functions:**
```javascript
// Sanitize and escape HTML
sanitizeInput(input)

// Validate post content
validatePostContent(content, maxLength)

// Validate post header/title
validatePostHeader(header, maxLength)

// Validate name fields
validateName(name, maxLength)

// Detect dangerous patterns
isSafeInput(input)
```

**Protected Pages:**
- Create Post page (`CreatePost.vue`)
- Edit Post page (`EditPost.vue`)

**Note:** Vue.js automatically escapes HTML in templates using `{{ }}` syntax, providing additional built-in XSS protection.

---

## 🔐 Secrets Management

### Secrets Detection in CI/CD Pipeline

**Tool Used**: [Gitleaks](https://github.com/gitleaks/gitleaks) v8.x

**Implementation:**

Both backend and frontend pipelines now include a `secrets-detection` job in the `security` stage:

```yaml
secrets-detection:
  stage: security
  image: zricethezav/gitleaks:latest
  script:
    - echo "🔍 Scanning for secrets in codebase..."
    - gitleaks detect --source . --verbose --no-git --redact
  allow_failure: false
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
    - if: '$CI_COMMIT_BRANCH == "main"'
    - if: '$CI_COMMIT_BRANCH =~ /^(develop|staging)$/'
```

**What it Detects:**
- API keys
- Passwords
- JWT secrets
- Database credentials
- SSH private keys
- OAuth tokens
- AWS credentials
- Generic secrets

**Exclusions (`.gitleaksignore`):**
- Test files (`*_test.go`, `*.test.js`)
- Documentation files (`*.md`)
- Example environment files (`.env.example`)
- Generated files (`coverage.out`, `package-lock.json`)

### GitLab CI/CD Secret Variables

All sensitive information is stored as **protected** GitLab CI/CD variables:

| Variable | Type | Masked | Used For |
|----------|------|--------|----------|
| `SSH_PRIVATE_KEY` | Variable | ✅ Yes | SSH access to VM for deployment |
| `VM_HOST` | Variable | ❌ No | VM hostname/IP address |
| `VM_USER` | Variable | ❌ No | VM username (e.g., centos) |

**Secret Protection in Logs:**
- SSH private keys are never printed to stdout/stderr
- Deployment scripts include comments indicating secret handling
- GitLab automatically masks variables marked as "masked"

---

## 🚀 Automated Deployment

### Current Deployment Architecture

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│   GitLab    │─────▶│  CI/CD       │─────▶│   VM        │
│  Repository │      │  Pipeline    │      │  (CentOS)   │
└─────────────┘      └──────────────┘      └─────────────┘
                            │
                            ├─ Security Stage (Secrets Detection)
                            ├─ Test Stage (Unit Tests, Coverage)
                            ├─ Build Stage (Compile Binary/Build Assets)
                            └─ Deploy Stage (SSH → VM → Systemd Restart)
```

### Backend Deployment Process

**Trigger**: Push to `main` branch

**Steps:**
1. **Security Stage**
   - Secrets detection with Gitleaks
   - Blocks deployment if secrets found

2. **Test Stage**
   - Code formatting check (`go fmt`, `go vet`)
   - Unit tests with PostgreSQL service
   - Code coverage calculation (min 50% target)

3. **Build Stage**
   - Compile Go binary
   - Store as artifact

4. **Deploy Stage**
   - SSH into VM using secure key
   - Copy binary to VM
   - Stop systemd service
   - Replace binary
   - Fix SELinux context
   - Start systemd service
   - Verify service is running

**Service Management:**
```bash
# Systemd service: devops-be.service
sudo systemctl restart devops-be
```

### Frontend Deployment Process

**Trigger**: Push to `main` branch

**Steps:**
1. **Security Stage**
   - Secrets detection with Gitleaks

2. **Build Stage**
   - Install npm dependencies
   - Build production assets (`npm run build`)
   - Store dist/ as artifact

3. **Deploy Stage**
   - SSH into VM
   - Create web directory
   - Rsync built files to `/var/www/html/devops-fe/`
   - Configure Apache `.htaccess` for SPA routing
   - Set correct permissions and SELinux context
   - Restart Apache

**Web Server:**
- Apache (httpd)
- Serves static files from `/var/www/html/devops-fe/`

---

## 🖥️ Virtual Machine Configuration

### Operating System
- **OS**: CentOS Linux (or similar RHEL-based)
- **Architecture**: x86_64

### Installed Services

**Backend:**
- PostgreSQL 12+ (database)
- Go binary running as systemd service

**Frontend:**
- Apache HTTP Server (httpd)
- Static file serving with SPA routing

### Network Configuration

**Backend Access:**
- Port: 8080 (HTTP)
- URL: http://$VM_HOST:8080
- Endpoints: `/posts`, `/login`, `/register`, etc.

**Frontend Access:**
- Port: 80 (HTTP) / 443 (HTTPS)
- URL: https://$VM_HOST
- Serves: Vue.js SPA

### Security Configuration

**SELinux:**
- Enabled on VM
- Contexts configured for:
  - Binary execution (`bin_t`)
  - Web content (`httpd_sys_content_t`)
  - Network connections (`httpd_can_network_connect`)

**Firewall:**
- Ports 80, 443 (frontend)
- Port 8080 (backend)
- Port 22 (SSH for deployment)

**SSH Access:**
- Key-based authentication only
- Dedicated SSH key for GitLab CI/CD
- Passwordless sudo for deployment commands

---

## 📊 CI/CD Pipeline Stages

### Complete Pipeline Overview

```
┌──────────────┐
│   Security   │  ← Secrets Detection (Gitleaks)
└──────┬───────┘
       │
┌──────▼───────┐
│     Test     │  ← Code Quality, Unit Tests, Coverage
└──────┬───────┘
       │
┌──────▼───────┐
│    Build     │  ← Compile Binary / Build Assets
└──────┬───────┘
       │
┌──────▼───────┐
│    Deploy    │  ← SSH to VM, Update Service
└──────────────┘
```

### Pipeline Jobs

**Backend Pipeline:**
1. `secrets-detection` - Scan for leaked secrets
2. `format` - Go code formatting validation
3. `test` - Unit tests with coverage
4. `compile` - Build Go binary
5. `deploy` - Deploy to VM (main branch only)

**Frontend Pipeline:**
1. `secrets-detection` - Scan for leaked secrets
2. `build` - Build production assets
3. `deploy` - Deploy to VM (main branch only)

---

## 🔄 Summary of Changes

### New Files Created

**Backend:**
- `internal/security/sanitization.go` - XSS prevention utilities
- `internal/security/sanitization_test.go` - Security tests
- `.gitleaksignore` - Secrets scanning exclusions
- `WEEK7_CHANGES.md` - This documentation

**Frontend:**
- `src/utils/validation.js` - Client-side validation
- `.gitleaksignore` - Secrets scanning exclusions

### Modified Files

**Backend:**
- `.gitlab-ci.yml` - Added security stage, secrets detection
- `internal/handlers/post.go` - Added input sanitization
- `internal/handlers/auth.go` - Added input sanitization
- `internal/handlers/image.go` - Added input sanitization

**Frontend:**
- `.gitlab-ci.yml` - Added security stage, secrets detection
- `src/pages/CreatePost.vue` - Added validation
- `src/pages/EditPost.vue` - Added validation

### Security Enhancements

✅ **XSS Protection**
- Backend: HTML escaping and pattern filtering
- Frontend: Client-side validation and sanitization

✅ **Secrets Detection**
- Automated scanning in CI/CD
- Blocks deployment if secrets found

✅ **Secrets Management**
- All secrets in GitLab CI/CD variables
- SSH keys never logged
- Protected and masked variables

✅ **Input Validation**
- Length limits on all user inputs
- Pattern validation for usernames
- DoS prevention through size limits

---

## 📚 Testing Security Features

### Test XSS Prevention

**Try these inputs** (should be safely escaped):

```html
<script>alert('XSS')</script>
<img src=x onerror="alert('XSS')">
javascript:alert('XSS')
<iframe src="evil.com"></iframe>
```

**Expected Result**: Content is displayed as plain text, not executed

### Test Secrets Detection

**Trigger secrets detection:**
```bash
# This will fail the pipeline:
echo "AWS_SECRET_KEY=AKIAIOSFODNN7EXAMPLE" > test.go
git add test.go
git commit -m "Test secrets detection"
git push
```

**Expected Result**: Pipeline fails at `secrets-detection` stage

### Test Deployment

**Trigger deployment:**
```bash
git add .
git commit -m "Deploy changes"
git push origin main
```

**Expected Result**: 
1. Secrets detection passes ✅
2. Tests pass ✅
3. Build succeeds ✅
4. Deployment to VM succeeds ✅
5. Service restarts automatically ✅

---

## 🎯 Compliance with Week 7 Requirements

### Development Requirements

✅ **XSS Prevention**
- Application is not vulnerable to XSS attacks
- All user input validated on frontend and backend
- User-generated content properly escaped
- No untrusted HTML inserted into DOM

### Operational Requirements

✅ **Secrets Detection**
- Gitleaks integrated in CI/CD pipeline
- Secrets never appear in logs
- GitLab secret management used for all sensitive data
- Automated checks prevent accidental exposure

✅ **Automatic Deployment**
- Push to main triggers deployment
- Application transferred to VM
- Service restarted automatically
- Deployment verified in pipeline

### Documentation Requirements

✅ **VM Configuration**
- Operating system documented (CentOS)
- Dependencies listed (PostgreSQL, Apache)
- Service management explained (systemd)

✅ **Application Access**
- URLs documented
- Ports listed
- No HTTPS yet (can be added with Let's Encrypt)

✅ **CI/CD Pipeline**
- All stages documented
- Secrets storage explained
- Deployment process detailed

✅ **Change Summary**
- Security measures documented
- Deployment differences explained
- Infrastructure changes listed

---

## 📞 Troubleshooting

### Secrets Detection Fails

**Problem**: Pipeline fails at secrets-detection stage

**Solution**: 
1. Check which file triggered the alert
2. Remove the secret from the file
3. Add to `.gitleaksignore` if it's a false positive
4. Store real secrets in GitLab CI/CD variables

### XSS Validation Fails

**Problem**: Valid input rejected as "dangerous"

**Solution**:
1. Check validation rules in `security/sanitization.go` or `utils/validation.js`
2. Adjust pattern matching if needed
3. Ensure input doesn't contain HTML tags or event handlers

### Deployment Fails

**Problem**: Deploy stage fails

**Solution**:
1. Check SSH_PRIVATE_KEY variable is set correctly
2. Verify VM_HOST and VM_USER are correct
3. Ensure VM is accessible
4. Check VM logs: `sudo journalctl -u devops-be -xe`

---

## 🔗 Additional Resources

- [Gitleaks Documentation](https://github.com/gitleaks/gitleaks)
- [OWASP XSS Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html)
- [GitLab CI/CD Variables](https://docs.gitlab.com/ee/ci/variables/)
- [Week 7 Requirements](https://devops.cs.ut.ee/phase2/week7/)

---
