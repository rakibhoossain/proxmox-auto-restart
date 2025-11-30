# API Usage Guide

Complete API reference for the Proxmox VM/Container Auto-Restart System.

## Base URL

```
http://localhost:8080
```

## Authentication

All API endpoints (except `/health`) require Basic HTTP Authentication.

**Default Credentials**:
- Username: `admin`
- Password: `proxmox2024`

**Change credentials** via environment variables:
```bash
export AUTH_USERNAME="your_username"
export AUTH_PASSWORD="your_password"
```

### Example with curl

```bash
curl -u admin:proxmox2024 http://localhost:8080/api/resources
```

## Endpoints

### 1. Health Check

**No authentication required**

```http
GET /health
```

**Response**:
```json
{
  "status": "ok",
  "proxmox_available": true,
  "timestamp": "2024-11-30T12:00:00Z"
}
```

---

### 2. List All VMs/Containers

Fetches all VMs and containers **in real-time** from Proxmox.

```http
GET /api/resources
```

**Response**:
```json
[
  {
    "vmid": 103,
    "name": "Final-Issabel-4",
    "type": "lxc",
    "node": "www",
    "status": "running",
    "uptime": 1086,
    "cpu_usage": 0.0814,
    "memory_used": 437579776,
    "memory_total": 4297064448,
    "disk_used": 3732348928,
    "disk_total": 41956900864
  },
  {
    "vmid": 105,
    "name": "Fresh-Issabel-4",
    "type": "lxc",
    "node": "www",
    "status": "running",
    "uptime": 1053,
    "cpu_usage": 0.0106,
    "memory_used": 311554048,
    "memory_total": 4297064448,
    "disk_used": 3473055744,
    "disk_total": 41956900864
  }
]
```

**curl example**:
```bash
curl -u admin:proxmox2024 http://localhost:8080/api/resources | jq
```

---

### 3. Get Specific VM/Container

Fetches details for a specific VM/Container **in real-time** from Proxmox.

```http
GET /api/resources/{vmid}?node={node}
```

**Parameters**:
- `vmid` (path) - VM/Container ID (e.g., 103)
- `node` (query) - Proxmox node name (e.g., www)

**Example**:
```http
GET /api/resources/103?node=www
```

**Response**:
```json
{
  "vmid": 103,
  "name": "Final-Issabel-4",
  "type": "lxc",
  "node": "www",
  "status": "running",
  "uptime": 1086,
  "cpu_usage": 0.0814,
  "memory_used": 437579776,
  "memory_total": 4297064448,
  "disk_used": 3732348928,
  "disk_total": 41956900864
}
```

**curl example**:
```bash
curl -u admin:proxmox2024 "http://localhost:8080/api/resources/103?node=www" | jq
```

---

### 4. Restart VM/Container

Triggers a restart operation (asynchronous).

```http
POST /api/resources/{vmid}/restart?node={node}
```

**Parameters**:
- `vmid` (path) - VM/Container ID
- `node` (query) - Proxmox node name

**Request Body** (optional):
```json
{
  "triggered_by": "user@dashboard"
}
```

**Response**:
```json
{
  "message": "Restart triggered",
  "vmid": 103,
  "node": "www"
}
```

**curl example**:
```bash
curl -u admin:proxmox2024 \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"triggered_by":"admin"}' \
  "http://localhost:8080/api/resources/103/restart?node=www"
```

---

### 5. Stop VM/Container

Stops a running VM/Container.

```http
POST /api/resources/{vmid}/stop?node={node}
```

**Parameters**:
- `vmid` (path) - VM/Container ID
- `node` (query) - Proxmox node name

**Request Body** (optional):
```json
{
  "triggered_by": "user@dashboard"
}
```

**Response**:
```json
{
  "message": "Stop triggered",
  "vmid": 103,
  "node": "www"
}
```

**curl example**:
```bash
curl -u admin:proxmox2024 \
  -X POST \
  "http://localhost:8080/api/resources/103/stop?node=www"
```

---

### 6. Start VM/Container

Starts a stopped VM/Container.

```http
POST /api/resources/{vmid}/start?node={node}
```

**Parameters**:
- `vmid` (path) - VM/Container ID
- `node` (query) - Proxmox node name

**Request Body** (optional):
```json
{
  "triggered_by": "user@dashboard"
}
```

**Response**:
```json
{
  "message": "Start triggered",
  "vmid": 103,
  "node": "www"
}
```

**curl example**:
```bash
curl -u admin:proxmox2024 \
  -X POST \
  "http://localhost:8080/api/resources/103/start?node=www"
```

---

### 7. List Whitelist

Get all VMs/Containers in the auto-restart whitelist.

```http
GET /api/whitelist
```

**Response**:
```json
[
  {
    "id": 1,
    "vmid": 103,
    "resource_name": "Final-Issabel-4",
    "node": "www",
    "enabled": true,
    "created_at": "2024-11-30T10:00:00Z",
    "created_by": "admin",
    "notes": "Production PBX"
  }
]
```

**curl example**:
```bash
curl -u admin:proxmox2024 http://localhost:8080/api/whitelist | jq
```

---

### 8. Add to Whitelist

Add a VM/Container to the auto-restart whitelist.

```http
POST /api/whitelist
```

**Request Body**:
```json
{
  "vmid": 103,
  "resource_name": "Final-Issabel-4",
  "node": "www",
  "created_by": "admin",
  "notes": "Production PBX server"
}
```

**Response**:
```json
{
  "message": "Added to whitelist successfully",
  "vmid": 103,
  "resource_name": "Final-Issabel-4",
  "node": "www"
}
```

**curl example**:
```bash
curl -u admin:proxmox2024 \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "vmid": 103,
    "resource_name": "Final-Issabel-4",
    "node": "www",
    "created_by": "admin",
    "notes": "Production PBX"
  }' \
  http://localhost:8080/api/whitelist
```

---

### 9. Update Whitelist Entry

Update an existing whitelist entry (enable/disable or change notes).

```http
PUT /api/whitelist/{id}
```

**Parameters**:
- `id` (path) - Whitelist entry ID

**Request Body**:
```json
{
  "enabled": false,
  "notes": "Disabled for maintenance"
}
```

**Response**:
```json
{
  "message": "Updated successfully"
}
```

**curl example**:
```bash
curl -u admin:proxmox2024 \
  -X PUT \
  -H "Content-Type: application/json" \
  -d '{"enabled": false, "notes": "Maintenance"}' \
  http://localhost:8080/api/whitelist/1
```

---

### 10. Delete from Whitelist

Remove a VM/Container from the auto-restart whitelist.

```http
DELETE /api/whitelist/{id}
```

**Parameters**:
- `id` (path) - Whitelist entry ID

**Response**:
```json
{
  "message": "Deleted successfully"
}
```

**curl example**:
```bash
curl -u admin:proxmox2024 \
  -X DELETE \
  http://localhost:8080/api/whitelist/1
```

---

### 11. Get Restart Logs

Retrieve restart operation logs with optional filtering.

```http
GET /api/logs?vmid=103&status=success&limit=50
```

**Query Parameters** (all optional):
- `vmid` - Filter by VM/Container ID
- `resource_name` - Filter by resource name (partial match)
- `node` - Filter by Proxmox node
- `action` - Filter by action (`restart`, `stop`, `start`)
- `trigger_type` - Filter by trigger (`auto`, `manual`)
- `status` - Filter by status (`success`, `failed`, `pending`)
- `start_date` - Filter by start date (RFC3339 format)
- `end_date` - Filter by end date (RFC3339 format)
- `limit` - Maximum number of results (default: 100)
- `offset` - Pagination offset (default: 0)

**Response**:
```json
[
  {
    "id": 1,
    "vmid": 103,
    "resource_name": "Final-Issabel-4",
    "node": "www",
    "action": "restart",
    "trigger_type": "auto",
    "triggered_by": "system",
    "status": "success",
    "started_at": "2024-11-30T06:00:00Z",
    "completed_at": "2024-11-30T06:00:15Z",
    "duration_seconds": 15
  }
]
```

**curl examples**:
```bash
# Get all logs
curl -u admin:proxmox2024 http://localhost:8080/api/logs | jq

# Get logs for specific VM
curl -u admin:proxmox2024 "http://localhost:8080/api/logs?vmid=103" | jq

# Get failed restarts
curl -u admin:proxmox2024 "http://localhost:8080/api/logs?status=failed" | jq

# Get auto-restarts only
curl -u admin:proxmox2024 "http://localhost:8080/api/logs?trigger_type=auto&limit=20" | jq
```

---

### 12. Get System Status

Retrieve overall system status and statistics.

```http
GET /api/status
```

**Response**:
```json
{
  "total_resources": 9,
  "running_resources": 3,
  "whitelisted_count": 2,
  "next_restart_time": "2024-11-30T18:00:00Z",
  "total_restarts": 45,
  "failed_restarts": 2
}
```

**curl example**:
```bash
curl -u admin:proxmox2024 http://localhost:8080/api/status | jq
```

---

## Response Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `202 Accepted` - Operation triggered (async)
- `400 Bad Request` - Invalid parameters
- `401 Unauthorized` - Authentication required/failed
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Proxmox not available

---

## Error Response Format

All errors return JSON:

```json
{
  "error": "Error message description"
}
```

---

## Full Workflow Example

### 1. Check system health
```bash
curl http://localhost:8080/health
```

### 2. List all VMs/containers
```bash
curl -u admin:proxmox2024 http://localhost:8080/api/resources | jq
```

### 3. Add VM 103 to whitelist
```bash
curl -u admin:proxmox2024 \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "vmid": 103,
    "resource_name": "Final-Issabel-4",
    "node": "www",
    "created_by": "admin",
    "notes": "Production PBX"
  }' \
  http://localhost:8080/api/whitelist
```

### 4. Manually restart VM 103
```bash
curl -u admin:proxmox2024 \
  -X POST \
  "http://localhost:8080/api/resources/103/restart?node=www"
```

### 5. Check restart logs
```bash
curl -u admin:proxmox2024 "http://localhost:8080/api/logs?vmid=103&limit=10" | jq
```

### 6. View system status
```bash
curl -u admin:proxmox2024 http://localhost:8080/api/status | jq
```

---

## Notes

- **Real-time Data**: VM/Container status is fetched directly from Proxmox in real-time (not cached)
- **Auto-restart**: Whitelisted VMs/containers are automatically restarted every 6 hours
- **Async Operations**: Restart/stop/start operations are asynchronous - check logs for results
- **Database**: Only stores whitelist entries and operation logs (no VM status caching)
