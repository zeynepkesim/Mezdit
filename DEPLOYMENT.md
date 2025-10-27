# Deployment Guide - VM Setup

This guide explains how to deploy the DevOps Backend application to your VM.

## 📋 Prerequisites

- GitLab repository access
- VM (Ubuntu/Debian) with SSH access
- PostgreSQL installed on VM

---

## 🔑 1. Create SSH Key Pair

Create an SSH key pair for GitLab CI/CD to access your VM:

```bash
# On your local machine
ssh-keygen -t rsa -b 4096 -C "gitlab-ci@devops-be" -f ~/.ssh/gitlab-ci-devops-be

# Display the public key
cat ~/.ssh/gitlab-ci-devops-be.pub
```

---

## 🖥️ 2. VM Configuration

### 2.1. Add SSH Public Key to VM

```bash
# SSH into your VM
ssh your-user@your-vm-hostname

# Add to authorized keys
mkdir -p ~/.ssh
echo "YOUR_PUBLIC_KEY_HERE" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
chmod 700 ~/.ssh
```

### 2.2. PostgreSQL Setup (if not installed)

```bash
# Install PostgreSQL
sudo apt update
sudo apt install postgresql postgresql-contrib -y

# Start PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Configure passwordless connection for postgres user
sudo -u postgres psql -c "ALTER USER postgres WITH PASSWORD '';"

# Edit pg_hba.conf
sudo nano /etc/postgresql/*/main/pg_hba.conf

# Find this line:
# local   all             postgres                                peer
# Change to:
# local   all             postgres                                trust

# Restart PostgreSQL
sudo systemctl restart postgresql
```

### 2.3. Create Required Directories

```bash
# Directory for binary
sudo mkdir -p /usr/local/bin

# Images directory
mkdir -p ~/images
```

### 2.4. Systemd Service Setup

```bash
# Create service file
sudo nano /etc/systemd/system/devops-be.service
```

Paste the following content (edit values as needed):

```ini
[Unit]
Description=DevOps Backend API Service
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=ubuntu
Group=ubuntu
WorkingDirectory=/home/ubuntu

# Environment variables
Environment="DATABASE_URL=postgres://postgres@localhost:5432/postgres?sslmode=disable"
Environment="JWT_SECRET=YOUR_SECRET_KEY_HERE"
Environment="PORT=8080"

# Service executable
ExecStart=/usr/local/bin/devops-be

# Restart policy
Restart=on-failure
RestartSec=5s

# Security settings
NoNewPrivileges=true
PrivateTmp=true

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=devops-be

[Install]
WantedBy=multi-user.target
```

Enable the service:

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service
sudo systemctl enable devops-be

# Check service status
sudo systemctl status devops-be
```

### 2.5. Firewall Configuration (if applicable)

```bash
# Allow port 8080
sudo ufw allow 8080/tcp

# Check firewall status
sudo ufw status
```

### 2.6. Sudo Permissions (for CI/CD)

Allow GitLab CI/CD to run systemctl commands:

```bash
# Edit sudoers file
sudo visudo

# Add these lines (replace 'ubuntu' with your username):
ubuntu ALL=(ALL) NOPASSWD: /bin/systemctl start devops-be
ubuntu ALL=(ALL) NOPASSWD: /bin/systemctl stop devops-be
ubuntu ALL=(ALL) NOPASSWD: /bin/systemctl restart devops-be
ubuntu ALL=(ALL) NOPASSWD: /bin/systemctl status devops-be
ubuntu ALL=(ALL) NOPASSWD: /bin/mv /home/ubuntu/devops-be-new /usr/local/bin/devops-be
ubuntu ALL=(ALL) NOPASSWD: /bin/chmod +x /usr/local/bin/devops-be
```

---

## 🦊 3. GitLab CI/CD Configuration

### 3.1. Add GitLab Variables

In GitLab: **Settings → CI/CD → Variables**

Add these variables:

| Key | Value | Type | Masked | Description |
|-----|-------|------|--------|-------------|
| `SSH_PRIVATE_KEY` | Private key content | Variable | ✅ Yes | For GitLab to connect to VM |
| `VM_HOST` | Your VM hostname/IP | Variable | ❌ No | Example: `192.168.1.100` |
| `VM_USER` | VM username | Variable | ❌ No | Example: `ubuntu` |

#### How to Add SSH_PRIVATE_KEY?

**Option 1: Direct (Unmasked)**
```bash
# Copy private key
cat ~/.ssh/gitlab-ci-devops-be

# Paste in GitLab (including BEGIN and END lines)
```

**Option 2: Base64 Encoded (Masked - Recommended)**
```bash
# Encode private key with base64
cat ~/.ssh/gitlab-ci-devops-be | base64 -w 0

# Add encoded string to GitLab
# NOTE: Don't forget to decode in .gitlab-ci.yml:
# echo "$SSH_PRIVATE_KEY" | base64 -d > ~/.ssh/id_rsa
```

### 3.2. Coverage Badge Settings

In GitLab: **Settings → CI/CD → General pipelines → Test coverage parsing**

Add regex:
```
/total:\s+\(statements\)\s+(\d+\.\d+)%/
```

---

## 🚀 4. Initial Deployment

### 4.1. Manual First Deployment

Upload binary manually for the first time:

```bash
# Build on local machine
go build -o devops-be .

# Copy to VM
scp devops-be your-user@your-vm:/home/your-user/

# On VM, move binary
ssh your-user@your-vm
sudo mv ~/devops-be /usr/local/bin/devops-be
sudo chmod +x /usr/local/bin/devops-be

# Start service
sudo systemctl start devops-be
sudo systemctl status devops-be
```

### 4.2. Deployment via GitLab Pipeline

```bash
# Push to main branch
git add .
git commit -m "Add VM deployment configuration"
git push origin main
```

Watch the pipeline in GitLab:
1. **Test Stage** ✅ - Tests run, coverage calculated
2. **Build Stage** ✅ - Binary created
3. **Deploy Stage** ✅ - Deployed to VM

---

## 📊 5. Post-Deployment Checks

### 5.1. Check Service Status

```bash
# Service status
sudo systemctl status devops-be

# View logs
sudo journalctl -u devops-be -f

# Last 100 lines
sudo journalctl -u devops-be -n 100
```

### 5.2. Test API

```bash
# Health check
curl http://your-vm-hostname:8080/posts

# Login test
curl -X POST http://your-vm-hostname:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","password":"password123"}'
```

### 5.3. Check Database Connection

```bash
# Connect to PostgreSQL
sudo -u postgres psql

# Check database
\c postgres
\dt

# List users
SELECT * FROM users;
```

---

## 🔧 6. Troubleshooting

### Problem: Service won't start

```bash
# View detailed error logs
sudo journalctl -u devops-be -xe

# Check if binary is executable
ls -la /usr/local/bin/devops-be

# Run binary manually to see error
/usr/local/bin/devops-be
```

### Problem: Database connection error

```bash
# Check PostgreSQL is running
sudo systemctl status postgresql

# Test connection
psql -U postgres -d postgres -c "SELECT 1;"

# If you get "peer authentication failed":
sudo nano /etc/postgresql/*/main/pg_hba.conf
# Change "peer" lines to "trust"
sudo systemctl restart postgresql
```

### Problem: Port 8080 not accessible

```bash
# Check app is listening
sudo netstat -tlnp | grep 8080

# Check firewall rules
sudo ufw status

# Open port
sudo ufw allow 8080/tcp
```

### Problem: SSH connection error (GitLab)

```bash
# Check SSH key is added on VM
cat ~/.ssh/authorized_keys

# Check SSH_PRIVATE_KEY variable in GitLab
# Look for SSH errors in pipeline logs
```

### Problem: Permission denied errors

```bash
# Check user's sudo permissions
sudo -l

# Edit sudoers file
sudo visudo

# Add service restart permissions
ubuntu ALL=(ALL) NOPASSWD: /bin/systemctl * devops-be
```

---

## 📝 7. Maintenance

### Log Rotation

```bash
# Clean systemd logs
sudo journalctl --vacuum-time=7d

# Or set size limit
sudo journalctl --vacuum-size=100M
```

### Update Deployment

Trigger pipeline by pushing to main branch:

```bash
git push origin main
```

### Rollback

To rollback to an old version:

```bash
# Restore old binary from backup
sudo systemctl stop devops-be
sudo cp /backup/devops-be-old /usr/local/bin/devops-be
sudo systemctl start devops-be
```

---

## ✅ Checklist

VM side:
- [ ] SSH public key added
- [ ] PostgreSQL installed and configured
- [ ] Systemd service created and enabled
- [ ] Firewall port opened (8080)
- [ ] Sudo permissions granted
- [ ] Images directory created

GitLab side:
- [ ] `SSH_PRIVATE_KEY` variable added
- [ ] `VM_HOST` variable added
- [ ] `VM_USER` variable added
- [ ] Coverage regex configured
- [ ] Pipeline is running

Test:
- [ ] Service starts successfully
- [ ] API endpoints accessible
- [ ] Database connection working
- [ ] Pipeline deploy stage successful

---

## 🎯 Summary

Now every time you push to main branch:
1. ✅ Tests run automatically
2. ✅ Coverage is calculated
3. ✅ Binary is built
4. ✅ Auto-deployed to VM
5. ✅ Service is restarted

Congratulations! 🎉

---

## 📚 Additional Resources

- [GitLab CI/CD Documentation](https://docs.gitlab.com/ee/ci/)
- [Systemd Service Documentation](https://www.freedesktop.org/software/systemd/man/systemd.service.html)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

---

## 🔐 Security Notes

### Change JWT_SECRET

In `devops-be.service` file:
```ini
Environment="JWT_SECRET=YOUR_SECRET_KEY_HERE"
```

Generate a secure secret:
```bash
openssl rand -base64 32
```

### Database URL

If PostgreSQL is on a different server:
```ini
Environment="DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=disable"
```

### Change Port

If using a port other than 8080:
```ini
Environment="PORT=3000"
```

And open in firewall:
```bash
sudo ufw allow 3000/tcp
```

---

## 📞 Support

For issues or questions:
1. Check logs: `sudo journalctl -u devops-be -xe`
2. Review GitLab pipeline logs
3. Verify all checklist items are completed
4. Check course documentation: https://devops.cs.ut.ee/phase2/week5/
