# Proxmox Auto-Restart System

Automatically restart whitelisted Proxmox nodes every 6 hours with SQLite3 tracking, manual controls, and a Next.js dashboard.

## Features

- **Auto-Restart Schedule**: Restart whitelisted nodes every 6 hours
- **Node Sync**: Update node status in SQLite every 1 minute
- **Manual Controls**: REST/Start/Stop nodes via API or UI
- **Whitelist Management**: CRUD operations for nodes to auto-restart
- **Audit Logging**: Complete restart history in SQLite3
- **Web Dashboard**: Next.js + shadcn UI for monitoring and control

## Architecture

```
┌─────────────────────────────────────────────────┐
│  Proxmox Server                                 │
│  ┌───────────────────────────────────────────┐  │
│  │  Go Binary                                │  │
│  │  ├─ HTTP API Server (:8080)               │  │
│  │  ├─ Node Sync Job (every 1 min)           │  │
│  │  ├─ Auto-Restart Job (every 6 hours)      │  │
│  │  └─ SQLite3 Database                      │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
                    ▲
                    │ HTTP API
                    │
        ┌───────────┴───────────┐
        │  Next.js Frontend     │
        │  + shadcn UI          │
        └───────────────────────┘
```

## Quick Start

### Build

```bash
# For Linux (Proxmox server)
make build

# For local development
make build-local
```

### Run Locally

```bash
make run
```

### Deploy to Proxmox Server

```bash
make deploy PROXMOX_HOST=root@your-proxmox-server
```

## Authentication

**Default Credentials:**
- Username: `admin`
- Password: `proxmox2024`

**⚠️ IMPORTANT**: Change these in production!

### Change Credentials

**Backend (systemd service):**
```bash
nano /etc/systemd/system/proxmox-auto-restart.service
# Update AUTH_USERNAME and AUTH_PASSWORD
systemctl daemon-reload && systemctl restart proxmox-auto-restart
```

**Frontend (.env.local):**
```bash
nano web/.env.local
# Update AUTH_USERNAME and AUTH_PASSWORD
pm2 restart proxmox-ui
```

### Testing API with Auth

```bash
# Health check (no auth required)
curl http://localhost:8080/health

# Get nodes (requires auth)
curl -u admin:proxmox2024 http://localhost:8080/api/nodes
```

## Pre-Deployment Testing

Before deploying, test Proxmox shell commands manually:

```bash
ssh root@your-proxmox-server
bash <(curl -s https://raw.githubusercontent.com/your-repo/main/test_proxmox.sh)
```

See [TESTING.md](TESTING.md) for complete testing guide.

## API Endpoints

### Nodes
- `GET /api/nodes` - List all nodes with status
- `GET /api/nodes/:node` - Get specific node details
- `POST /api/nodes/:node/restart` - Manual restart
- `POST /api/nodes/:node/stop` - Manual stop
- `POST /api/nodes/:node/start` - Manual start

### Whitelist
- `GET /api/whitelist` - Get whitelisted nodes
- `POST /api/whitelist` - Add node to whitelist
- `PUT /api/whitelist/:id` - Update whitelist entry
- `DELETE /api/whitelist/:id` - Remove from whitelist

### Logs
- `GET /api/logs` - Get restart logs (with pagination)
- `GET /api/logs/:id` - Get specific log entry

### System
- `GET /api/status` - System status
- `GET /health` - Health check

## Database Schema

### nodes
- Stores current status of all Proxmox nodes
- Updated every 1 minute

### whitelist
- Nodes configured for auto-restart every 6 hours
- Supports enable/disable and notes

### restart_logs
- Audit trail of all restart operations
- Tracks auto and manual restarts

## Configuration

Environment variables (optional):
- `PORT` - HTTP server port (default: 8080)
- `DB_PATH` - SQLite database path (default: ./proxmox.db)
- `SYNC_INTERVAL` - Node sync interval (default: 1m)
- `RESTART_INTERVAL` - Auto-restart interval (default: 6h)

## Development

```bash
# Install dependencies
make deps

# Run tests
make test

# Clean build artifacts
make clean
```

## License

MIT
