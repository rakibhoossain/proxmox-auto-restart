# Proxmox Auto-Restart System - Complete Deployment Guide

Step-by-step guide to deploy the Proxmox auto-restart system with authentication.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Pre-Deployment Testing](#pre-deployment-testing)
3. [Backend Deployment](#backend-deployment)
4. [Frontend Deployment](#frontend-deployment)
5. [Configuration](#configuration)
6. [Testing Deployment](#testing-deployment)
7. [Usage](#usage)
8. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### On Your Development Machine
- Go 1.21+ installed (for building)
- Node.js 20+ and npm (for frontend)
- SSH access to Proxmox server
- Git (optional, for version control)

### On Proxmox Server
- Proxmox VE 7.x or 8.x installed
- Root SSH access
- `pvesh` command available (comes with Proxmox)
- Ports 8080 (backend) and 3000 (frontend) available

---

## Pre-Deployment Testing

**⚠️ IMPORTANT**: Test Proxmox commands manually before deploying!

See [TESTING.md](TESTING.md) for complete test guide.

### Quick Test

```bash
# SSH into your Proxmox server
ssh root@your-proxmox-ip

# Run the test script
bash /tmp/test_proxmox.sh
```

**Expected Output:**
```
✅ pvesh found
✅ Nodes retrieved successfully
✅ Current node found in Proxmox
```

If all tests pass, proceed to deployment!

---

## Backend Deployment

### Step 1: Build the Binary

On your **development machine**:

```bash
cd /path/to/proxmox-auto-restart

# Build Linux binary
make build

# Verify binary created
ls -lh proxmox-auto-restart
```

**Expected:** `proxmox-auto-restart` file (~15-20MB)

### Step 2: Copy Binary to Proxmox Server

```bash
# Create directory on server
ssh root@your-proxmox-server 'mkdir -p /opt/proxmox-auto-restart'

# Copy binary
scp proxmox-auto-restart root@your-proxmox-server:/opt/proxmox-auto-restart/

# Make it executable
ssh root@your-proxmox-server 'chmod +x /opt/proxmox-auto-restart/proxmox-auto-restart'
```

### Step 3: Configure Authentication

Set environment variables for API authentication:

```bash
ssh root@your-proxmox-server

# Edit systemd service to set credentials
nano /etc/systemd/system/proxmox-auto-restart.service
```

**Service File Content:**

```ini
[Unit]
Description=Proxmox Auto-Restart Service
After=network.target pve-cluster.service
Wants=pve-cluster.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/proxmox-auto-restart
ExecStart=/opt/proxmox-auto-restart/proxmox-auto-restart
Restart=always
RestartSec=10

# API Server Port
Environment="PORT=8080"

# Database Path
Environment="DB_PATH=/opt/proxmox-auto-restart/proxmox.db"

# Authentication Credentials (CHANGE THESE!)
Environment="AUTH_USERNAME=admin"
Environment="AUTH_PASSWORD=your-secure-password-here"

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=proxmox-auto-restart

[Install]
WantedBy=multi-user.target
```

### Step 4: Install and Start Service

```bash
# Copy service file
scp deployments/proxmox-auto-restart.service root@your-proxmox-server:/etc/systemd/system/

# Or create it manually (shown above)

# Reload systemd
systemctl daemon-reload

# Enable auto-start
systemctl enable proxmox-auto-restart

# Start serviceservice
systemctl start proxmox-auto-restart

# Check status
systemctl status proxmox-auto-restart
```

**Expected Output:**
```
● proxmox-auto-restart.service - Proxmox Auto-Restart Service
   Loaded: loaded (/etc/systemd/system/proxmox-auto-restart.service)
   Active: active (running) since...
```

### Step 5: Verify Backend is Running

```bash
# Test health endpoint (no auth required)
curl http://localhost:8080/health

# Expected response:
# {"status":"ok","proxmox_available":true,"timestamp":"..."}

# Test authenticated API (use your credentials)curl -u admin:your-secure-password-here http://localhost:8080/api/nodes

# Expected: JSON array of nodes
```

### View Logs

```bash
# Real-time logs
journalctl -u proxmox-auto-restart -f

# Last 50 lines
journalctl -u proxmox-auto-restart -n 50
```

---

## Frontend Deployment

### Option 1: Deploy Frontend on Proxmox Server (Recommended)

#### Build on Development Machine

```bash
cd web

# Set API URL (same server)
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env.local
echo "AUTH_USERNAME=admin" >> .env.local
echo "AUTH_PASSWORD=your-secure-password-here" >> .env.local

# Build Next.js
npm run build
```

#### Copy to Server

```bash
# Create directory
ssh root@your-proxmox-server 'mkdir -p /opt/proxmox-auto-restart/web'

# Copy build files
scp -r .next package.json package-lock.json root@your-proxmox-server:/opt/proxmox-auto-restart/web/
```

#### Install Node.js on Proxmox (if not installed)

```bash
ssh root@your-proxmox-server

# Add NodeSource repository
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -

# Install Node.js
apt-get install -y nodejs

# Verify installation
node --version
npm --version
```

#### Install Dependencies and Start

```bash
cd /opt/proxmox-auto-restart/web
npm ci --production

# Install PM2 for process management
npm install -g pm2

# Start Next.js
pm2 start npm --name "proxmox-ui" -- start

# Save PM2 configuration
pm2 save

# Enable PM2 startup
pm2 startup
```

**Access Dashboard:** `http://your-proxmox-server:3000`

### Option 2: Deploy Frontend on Separate Server

#### Configure .env.local

```bash
# Point to Proxmox server
NEXT_PUBLIC_API_URL=http://your-proxmox-server:8080
AUTH_USERNAME=admin
AUTH_PASSWORD=your-secure-password-here
```

#### Build and Run

```bash
npm run build
npm start

# Or use PM2
pm2 start npm --name "proxmox-ui" -- start
```

---

## Configuration

### Default Credentials

**Username:** `admin`  
**Password:** `proxmox2024`

**⚠️ IMPORTANT**: Change these immediately in production!

### Change Credentials

#### Backend (Proxmox Server)

```bash
# Edit systemd service
nano /etc/systemd/system/proxmox-auto-restart.service

# Update these lines:
Environment="AUTH_USERNAME=your-username"
Environment="AUTH_PASSWORD=your-password"

# Restart service
systemctl daemon-reload
systemctl restart proxmox-auto-restart
```

#### Frontend

```bash
# Edit .env.local
nano /opt/proxmox-auto-restart/web/.env.local

# Update:
AUTH_USERNAME=your-username
AUTH_PASSWORD=your-password

# Restart Next.js
pm2 restart proxmox-ui
```

### Environment Variables

**Backend (`/etc/systemd/system/proxmox-auto-restart.service`):**

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP API port | `8080` |
| `DB_PATH` | SQLite database path | `./proxmox.db` |
| `AUTH_USERNAME` | API username | `admin` |
| `AUTH_PASSWORD` | API password | `proxmox2024` |

**Frontend (`.env.local`):**

| Variable | Description | Example |
|----------|-------------|---------|
| `NEXT_PUBLIC_API_URL` | Backend API URL | `http://localhost:8080` |
| `AUTH_USERNAME` | API username | `admin` |
| `AUTH_PASSWORD` | API password | `proxmox2024` |

---

## Testing Deployment

### 1. Test Backend API

```bash
# Health check (no auth)
curl http://your-proxmox-server:8080/health

# Get nodes (with auth)
curl -u admin:your-password http://your-proxmox-server:8080/api/nodes

# Get system status
curl -u admin:your-password http://your-proxmox-server:8080/api/status
```

### 2. Test Frontend

1. Open browser: `http://your-proxmox-server:3000`
2. You should see the dashboard (no login page needed - auth is in server actions)
3. Verify nodes are displayed
4. Check system stats at top

### 3. Test Whitelist Management

1. Go to "Whitelist" tab
2. Click "Add Node to Whitelist"
3. Select a node
4. Add optional notes
5. Click save
6. Verify node appears in whitelist

### 4. Test Manual Controls

1. Go to "Nodes" tab
2. Click "Manage" on a node
3. Click "Restart" button
4. Confirm the action
5. Go to "Logs" tab
6. Verify restart log entry appears

### 5. Verify Auto-Restart

Check that scheduler is running:

```bash
journalctl -u proxmox-auto-restart -f | grep "Auto-restart"
```

Expected every 6 hours:
```
Auto-restart scheduler started (every 6 hours, next restart: ...)
Starting auto-restart of whitelisted nodes...
```

---

## Usage

### Daily Operations

#### Add Node to Auto-Restart

1. Dashboard → "Whitelist" tab
2. "Add Node to Whitelist"
3. Select node from dropdown
4. Add notes (optional)
5. Save

→ Node will now restart every 6 hours

#### Disable Auto-Restart (Without Removing)

1. "Whitelist" tab
2. Click toggle button on node entry
3. Node stays in list but won't auto-restart

#### Manual Restart

1. "Nodes" tab → Click "Manage"
2. Click "Restart" button
3. Confirm

### View History

1. "Logs" tab
2. See all restart operations with:
   - Timestamp
   - Action (restart/stop/start)
   - Trigger (auto/manual)
   - Status (success/failed)
   - Duration

---

## Troubleshooting

### Backend Issues

#### Service Won't Start

**Check logs:**
```bash
journalctl -u proxmox-auto-restart -n 50
```

**Common issues:**
- Port 8080 in use: Change PORT environment variable
- Database permissions: Check `/opt/proxmox-auto restart/proxmox.db` ownership
- pvesh not found: Not running on Proxmox server

#### Can't Access API

**Test locally:**
```bash
ssh root@proxmox-server
curl http://localhost:8080/health
```

**If fails:**
- Service not running: `systemctl status proxmox-auto-restart`
- Firewall blocking: `iptables -L | grep 8080`

#### Authentication Failing

**Check credentials:**
```bash
# View service file
cat /etc/systemd/system/proxmox-auto-restart.service | grep AUTH

# Test with curl
curl -u admin:your-password http://localhost:8080/api/nodes
```

### Frontend Issues

#### Can't Connect to Backend

**Check API URL:**
```bash
cat /opt/proxmox-auto-restart/web/.env.local
```

**Test connection:**
```bash
curl -u admin:your-password http://your-api-url/health
```

#### No Nodes Showing

**Check node sync:**
```bash
journalctl -u proxmox-auto-restart | grep "Node sync"
```

Should see:
```
Node sync completed, updated X nodes
```

#### PM2 Process Stopped

**Check PM2:**
```bash
pm2 list
pm2 logs proxmox-ui
pm2 restart proxmox-ui
```

### Network Issues

#### Open Firewall Ports

If accessing from external network:

```bash
# Backend API
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT

# Frontend
iptables -A INPUT -p tcp --dport 3000 -j ACCEPT

# Save rules
iptables-save > /etc/iptables/rules.v4
```

---

## Security Recommendations

### 1. Change Default Credentials

**Immediately after deployment**, change from defaults:

```bash
# Backend
nano /etc/systemd/system/proxmox-auto-restart.service
# Update AUTH_USERNAME and AUTH_PASSWORD

# Frontend
nano /opt/proxmox-auto-restart/web/.env.local
# Update AUTH_USERNAME and AUTH_PASSWORD
```

### 2. Use Strong Passwords

Generate secure password:

```bash
openssl rand -base64 32
```

### 3. Restrict Network Access

**Option A: Firewall Rules**

Only allow from your management network:

```bash
iptables -A INPUT -p tcp -s 192.168.1.0/24 --dport 8080 -j ACCEPT
iptables -A INPUT -p tcp --dport 8080 -j DROP
```

**Option B: Reverse Proxy with SSL**

Use nginx as reverse proxy:

```nginx
server {
    listen 443 ssl;
    server_name proxmox-restart.example.com;

    ssl_certificate /etc/ssl/certs/cert.pem;
    ssl_certificate_key /etc/ssl/private/key.pem;

    location / {
        proxy_pass http://localhost:3000;
    }

    location /api {
        proxy_pass http://localhost:8080;
    }
}
```

### 4. Regular Backups

Backup SQLite database:

```bash
# Create backup script
cat > /opt/proxmox-auto-restart/backup.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
cp /opt/proxmox-auto-restart/proxmox.db /opt/proxmox-auto-restart/backups/proxmox_$DATE.db
find /opt/proxmox-auto-restart/backups -name "proxmox_*.db" -mtime +7 -delete
EOF

chmod +x /opt/proxmox-auto-restart/backup.sh

# Add to crontab (daily at 2 AM)
echo "0 2 * * * /opt/proxmox-auto-restart/backup.sh" | crontab -
```

---

## Quick Reference

### Start/Stop/Restart

```bash
# Backend
systemctl start proxmox-auto-restart
systemctl stop proxmox-auto-restart
systemctl restart proxmox-auto-restart
systemctl status proxmox-auto-restart

# Frontend
pm2 start proxmox-ui
pm2 stop proxmox-ui
pm2 restart proxmox-ui
pm2 logs proxmox-ui
```

### View Logs

```bash
# Backend (real-time)
journalctl -u proxmox-auto-restart -f

# Frontend
pm2 logs proxmox-ui

# SQLite database
sqlite3 /opt/proxmox-auto-restart/proxmox.db "SELECT * FROM restart_logs ORDER BY started_at DESC LIMIT 10"
```

### Update  System

```bash
# On development machine
make build
scp proxmox-auto-restart root@proxmox-server:/opt/proxmox-auto-restart/

# On Proxmox server
systemctl restart proxmox-auto-restart
```

### Uninstall

```bash
# Stop services
systemctl stop proxmox-auto-restart
pm2 stop proxmox-ui
pm2 delete proxmox-ui

# Remove files
rm -rf /opt/proxmox-auto-restart
rm /etc/systemd/system/proxmox-auto-restart.service

# Reload systemd
systemctl daemon-reload
```

---

## Support

### Check System Health

```bash
# Backend health
curl http://localhost:8080/health

# Database size
ls -lh /opt/proxmox-auto-restart/proxmox.db

# Service uptime
systemctl status proxmox-auto-restart | grep Active

# PM2 status
pm2 list
```

### Logs Location

- **Backend**: `journalctl -u proxmox-auto-restart`
- **Frontend**: `pm2 logs proxmox-ui`
- **Database**: `/opt/proxmox-auto-restart/proxmox.db`

### Default Login Credentials

- **Username**: `admin`
- **Password**: `proxmox2024`

**⚠️ CHANGE THESE IN PRODUCTION!**

---

## Next Steps

1. ✅ Test SSH commands → [TESTING.md](TESTING.md)
2. ✅ Deploy backend → Follow steps above
3. ✅ Deploy frontend → Follow steps above
4. ✅ Change default passwords
5. ✅ Add nodes to whitelist
6. ✅ Monitor first auto-restart cycle

**You're all set! 🎉**
