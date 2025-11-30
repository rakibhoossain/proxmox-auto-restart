# Proxmox Auto-Restart System - Testing Guide

Complete guide for manually testing Proxmox shell commands for VM/Container management before deploying the system.

## Prerequisites

- SSH access to Proxmox server: `ssh root@your-proxmox-ip`
- Root privileges on Proxmox server
- Proxmox VE installed and running

## Step 1: Connect to Proxmox Server

```bash
# Replace with your Proxmox server IP
ssh root@192.168.1.100

# Or using hostname
ssh root@pve.local
```

## Step 2: Verify Proxmox Tools

### Check if pvesh is available

```bash
which pvesh
# Expected output: /usr/bin/pvesh
```

```bash
pvesh --version
# Expected output: pve-manager/X.X.X/...
```

### Test pvesh help

```bash
pvesh help
# Should show available commands
```

## Step 3: List All VMs and Containers

### Get all VMs and containers (JSON format)

```bash
pvesh get /cluster/resources --type vm --output-format json
```

**Expected Output:**
```json
[
  {
    "cpu": 0.0106,
    "disk": 3736354816,
    "diskread": 472944640,
    "diskwrite": 22896640,
    "id": "lxc/103",
    "maxcpu": 4,
    "maxdisk": 41940746240,
    "maxmem": 4294967296,
    "mem": 437170176,
    "name": "Final-Issabel-4",
    "node": "www",
    "status": "running",
    "type": "lxc",
    "uptime": 12345,
    "vmid": "103"
  },
  {
    "cpu": 0.0909,
    "disk": 3466027008,
    "id": "lxc/105",
    "maxcpu": 2,
    "maxmem": 4294967296,
    "mem": 302915584,
    "name": "Fresh-Issabel-4",
    "node": "www",
    "status": "running",
    "type": "lxc",
    "vmid": "105"
  },
  {
    "cpu": 0,
    "disk": 0,
    "id": "lxc/100",
    "maxcpu": 4,
    "maxmem": 536870912,
    "mem": 0,
    "name": "CT100",
    "node": "www",
    "status": "stopped",
    "type": "lxc",
    "vmid": "100"
  }
]
```

### Get all resources (human-readable)

```bash
pvesh get /cluster/resources --type vm
```

**Expected Output:**
```
┌─────────┬──────┬────────┬─────────────────────┬─────────┬───────┬
│ id      │ type │  vmid  │ name                │ status  │ node  │
├─────────┼──────┼────────┼─────────────────────┼─────────┼───────┤
│ lxc/100 │ lxc  │    100 │ CT100               │ stopped │ www   │
│ lxc/103 │ lxc  │    103 │ Final-Issabel-4     │ running │ www   │
│ lxc/105 │ lxc  │    105 │ Fresh-Issabel-4     │ running │ www   │
└─────────┴──────┴────────┴─────────────────────┴─────────┴───────┘
```

## Step 4: Get VM/Container Details

### Get container status

Replace `www` with your node name and `103` with your VMID:

```bash
pvesh get /nodes/www/lxc/103/status/current --output-format json
```

**Expected Output:**
```json
{
  "cpu": 0.0106,
  "cpus": 4,
  "disk": 3736354816,
  "diskread": 472944640,
  "diskwrite": 22896640,
  "mem": 437170176,
  "maxdisk": 41940746240,
  "maxmem": 4294967296,
  "name": "Final-Issabel-4",
  "status": "running",
  "type": "lxc",
  "uptime": 12345,
  "vmid": 103
}
```

### Get VM status (if you have qemu VMs)

```bash
pvesh get /nodes/www/qemu/103/status/current --output-format json
```

## Step 5: Test Different VMs and Containers

### List all VMIDs

```bash
pvesh get /cluster/resources --type vm --output-format json | grep -o '"vmid":"[^"]*"' | cut -d'"' -f4
```

**Expected Output:**
```
100
101
102
103
104
105
106
107
200
```

### Test each container

```bash
# Test getting status for each
for vmid in 100 103 105; do
  echo "=== Container $vmid ==="
  pvesh get /nodes/www/lxc/$vmid/status/current --output-format json | jq '.status, .name'
done
```

## Step 6: Parse VM/Container Information

### Extract specific fields using jq

First, install jq if not available:

```bash
apt-get update && apt-get install -y jq
```

### Get VMIDs only

```bash
pvesh get /cluster/resources --type vm --output-format json | jq -r '.[].vmid'
```

**Expected Output:**
```
100
101
102
103
105
```

### Get VM/CT names and status

```bash
pvesh get /cluster/resources --type vm --output-format json | jq -r '.[] | "\(.vmid): \(.name) - \(.status)"'
```

**Expected Output:**
```
100: CT100 - stopped
103: Final-Issabel-4 - running
105: Fresh-Issabel-4 - running
```

### Get only running VMs/CTs

```bash
pvesh get /cluster/resources --type vm --output-format json | jq -r '.[] | select(.status == "running") | "\(.vmid): \(.name)"'
```

### Get CPU usage

```bash
pvesh get /cluster/resources --type vm --output-format json | jq -r '.[] | "\(.vmid) \(.name): CPU \((.cpu * 100) | floor)%"'
```

### Get memory usage

```bash
pvesh get /cluster/resources --type vm --output-format json | jq -r '.[] | select(.mem and .maxmem) | "\(.vmid) \(.name): \(.mem / 1024 / 1024 / 1024 | floor)GB / \(.maxmem / 1024 / 1024 / 1024 | floor)GB"'
```

### Filter by type (lxc or qemu)

```bash
# Only containers
pvesh get /cluster/resources --type vm --output-format json | jq -r '.[] | select(.type == "lxc") | "\(.vmid): \(.name)"'

# Only VMs
pvesh get /cluster/resources --type vm --output-format json | jq -r '.[] | select(.type == "qemu") | "\(.vmid): \(.name)"'
```

## Step 7: Test VM/Container Control Operations (CAUTION!)

> **⚠️ WARNING**: These commands will actually control your VMs/containers. Only test on non-production systems!

### Determine VM/CT Type

Before controlling a VM/CT, check its type:

```bash
TYPE=$(pvesh get /cluster/resources --type vm --output-format json | jq -r '.[] | select(.vmid == "103") | .type')
NODE=$(pvesh get /cluster/resources --type vm --output-format json | jq -r '.[] | select(.vmid == "103") | .node')
echo "VMID 103 is type: $TYPE on node: $NODE"
```

### Restart a Container

```bash
# Restart container 103 (replace with your test container)
pvesh create /nodes/www/lxc/103/status/reboot
```

**What happens:**
- Container immediately begins shutdown
- System reboots
- Container status becomes "running" again in ~10-30 seconds

### Restart a VM (if you have qemu VMs)

```bash
# Restart VM 200
pvesh create /nodes/www/qemu/200/status/reboot
```

### Stop a Container

```bash
pvesh create /nodes/www/lxc/103/status/stop
```

**What happens:**
- Container shuts down gracefully
- Status becomes "stopped"

### Start a Container

```bash
pvesh create /nodes/www/lxc/103/status/start
```

**What happens:**
- Container boots up
- Status becomes "running"

### Monitor Status Changes

From another terminal, watch status:

```bash
# Watch specific VM/CT
watch -n 1 'pvesh get /nodes/www/lxc/103/status/current --output-format json | jq -r ".status"'

# Watch all VMs/CTs
watch -n 1 'pvesh get /cluster/resources --type vm --output-format json | jq -r ".[] | \"\(.vmid): \(.status)\""'
```

## Step 9: Verify Hostname

This is important for the Go app to know which node it's running on:

```bash
hostname
# Output: pve (or your node name)
```

Check if hostname matches node name in Proxmox:

```bash
HOSTNAME=$(hostname)
pvesh get /nodes/$HOSTNAME/status --output-format json
```

If this works, the Go app will work correctly!

## Step 10: Test Error Handling

### Try to get a non-existent node

```bash
pvesh get /nodes/notarealnode/status --output-format json
```

**Expected:** Error message
```
<h1>Error 500: Parameter verification failed.</h1>
```

The Go app handles this gracefully.

## Step 11: Full Integration Test Script

Create a test script:

```bash
cat > /tmp/test_proxmox_vms.sh << 'EOF'
#!/bin/bash

echo "=== Proxmox VM/Container Management Test ==="
echo ""

# Test 1: Check pvesh
echo "1. Checking pvesh availability..."
if ! which pvesh > /dev/null; then
    echo "   ❌ FAILED: pvesh not found"
    exit 1
fi
echo "   ✅ pvesh found at $(which pvesh)"
echo ""

# Test 2: Get VMs/Containers
echo "2. Getting all VMs and Containers..."
RESOURCES=$(pvesh get /cluster/resources --type vm --output-format json 2>/dev/null)
if [ $? -ne 0 ]; then
    echo "   ❌ FAILED: Could not get VMs/Containers"
    exit 1
fi
echo "   ✅ VMs/Containers retrieved successfully"
echo ""

# Test 3: Count resources
RESOURCE_COUNT=$(echo "$RESOURCES" | jq length)
echo "3. Found $RESOURCE_COUNT VM(s)/Container(s)"
echo ""

# Test 4: List VMIDs and names
echo "4. VM/Container list:"
echo "$RESOURCES" | jq -r '.[] | "   - VMID \(.vmid): \(.name) (\(.type)) - \(.status)"'
echo ""

# Test 5: Count by type
LXC_COUNT=$(echo "$RESOURCES" | jq '[.[] | select(.type == "lxc")] | length')
QEMU_COUNT=$(echo "$RESOURCES" | jq '[.[] | select(.type == "qemu")] | length')
echo "5. Resource types:"
echo "   - Containers (LXC): $LXC_COUNT"
echo "   - VMs (QEMU): $QEMU_COUNT"
echo ""

# Test 6: Check node names
echo "6. Nodes hosting VMs/CTs:"
echo "$RESOURCES" | jq -r '.[].node' | sort | uniq | while read node; do
    echo "   - $node"
done
echo ""

# Test 7: Test getting a specific VM/CT status
echo "7. Testing individual VM/CT status..."
FIRST_VMID=$(echo "$RESOURCES" | jq -r '.[0].vmid')
FIRST_TYPE=$(echo "$RESOURCES" |jq -r '.[0].type')
FIRST_NODE=$(echo "$RESOURCES" | jq -r '.[0].node')
FIRST_NAME=$(echo "$RESOURCES" | jq -r '.[0].name')

if [ -n "$FIRST_VMID" ]; then
    echo "   Testing: $FIRST_NAME (VMID: $FIRST_VMID, Type: $FIRST_TYPE, Node: $FIRST_NODE)"
    if [ "$FIRST_TYPE" = "lxc" ]; then
        STATUS=$(pvesh get /nodes/$FIRST_NODE/lxc/$FIRST_VMID/status/current --output-format json 2>/dev/null | jq -r '.status // "error"')
    else
        STATUS=$(pvesh get /nodes/$FIRST_NODE/qemu/$FIRST_VMID/status/current --output-format json 2>/dev/null | jq -r '.status // "error"')
    fi
    
    if [ "$STATUS" != "error" ]; then
        echo "   ✅ Status check successful: $STATUS"
    else
        echo "   ⚠️  WARNING: Could not get status for $FIRST_NAME"
    fi
else
    echo "   ⚠️  No VMs/Containers found to test"
fi
echo ""

echo "=== Test Complete ==="
echo ""
echo "Summary:"
echo "  - pvesh: ✅ Working"
echo "  - VM/Container listing: ✅ Working"
echo "  - Total resources: $RESOURCE_COUNT"
echo "  - Containers: $LXC_COUNT"
echo "  - VMs: $QEMU_COUNT"
echo ""
echo "You can now proceed with deploying the Proxmox VM/Container Auto-Restart System!"
EOF

chmod +x /tmp/test_proxmox_vms.sh
```

### Run the test script

```bash
bash /tmp/test_proxmox_vms.sh
```

**Expected Output:**
```
=== Proxmox VM/Container Management Test ===

1. Checking pvesh availability...
   ✅ pvesh found at /usr/bin/pvesh

2. Getting all VMs and Containers...
   ✅ VMs/Containers retrieved successfully

3. Found 9 VM(s)/Container(s)

4. VM/Container list:
   - VMID 100: CT100 (lxc) - stopped
   - VMID 101: CT101 (lxc) - stopped
   - VMID 103: Final-Issabel-4 (lxc) - running
   - VMID 105: Fresh-Issabel-4 (lxc) - running

5. Resource types:
   - Containers (LXC): 9
   - VMs (QEMU): 0

6. Nodes hosting VMs/CTs:
   - www

7. Testing individual VM/CT status...
   Testing: CT100 (VMID: 100, Type: lxc, Node: www)
   ✅ Status check successful: stopped

=== Test Complete ===

Summary:
  - pvesh: ✅ Working
  - VM/Container listing: ✅ Working
  - Total resources: 9
  - Containers: 9  
  - VMs: 0

You can now proceed with deploying the Proxmox VM/Container Auto-Restart System!
```

## Step 12: Test Permissions

Verify root has necessary permissions:

```bash
# Check user
whoami
# Should output: root

# Check pvesh permissions
pvesh get /nodes --output-format json > /dev/null && echo "✅ Read access OK"

# Check if we can create (reboot) - this is read-only test
pvesh help nodes/{node}/status create > /dev/null && echo "✅ Write command exists"
```

## Common Issues and Solutions

### Issue 1: pvesh not found

**Error:** `bash: pvesh: command not found`

**Solution:**
```bash
# Verify Proxmox is installed
dpkg -l | grep pve-manager

# Reinstall if needed (Debian-based)
apt-get update
apt-get install --reinstall pve-manager
```

### Issue 2: Permission denied

**Error:** `permission denied`

**Solution:**
```bash
# Must be root
sudo su -

# Or use sudo with commands
sudo pvesh get /nodes --output-format json
```

### Issue 3: JSON parsing errors

**Error:** `parse error: Invalid numeric literal`

**Solution:**
```bash
# Install jq
apt-get install -y jq

# Or parse manually
pvesh get /nodes --output-format json | python3 -m json.tool
```

### Issue 4: Node name mismatch

**Error:** Node name doesn't match hostname

**Solution:**
```bash
# Check cluster configuration
cat /etc/pve/corosync.conf

# Update hostname if needed
hostnamectl set-hostname pve

# Reboot to apply
reboot
```

## Next Steps

Once all tests pass successfully:

1. ✅ pvesh commands work
2. ✅ Can list nodes
3. ✅ Can get node status
4. ✅ Hostname matches a Proxmox node

**You're ready to deploy the Proxmox Auto-Restart System!**

See [DEPLOYMENT.md](DEPLOYMENT.md) for deployment instructions.

## Quick Reference Commands

```bash
# List all VMs and containers
pvesh get /cluster/resources --type vm --output-format json

# Get specific resource status (container)
pvesh get /nodes/www/lxc/103/status/current --output-format json

# Get specific resource status (VM)
pvesh get /nodes/www/qemu/200/status/current --output-format json

# Restart container (CAREFUL!)
pvesh create /nodes/www/lxc/103/status/reboot

# Restart VM (CAREFUL!)
pvesh create /nodes/www/qemu/200/status/reboot

# Stop container
pvesh create /nodes/www/lxc/103/status/stop

# Start container  
pvesh create /nodes/www/lxc/103/status/start

# Get cluster status
pvesh get /cluster/status --output-format json

# Get current hostname
hostname

# Test full script
bash /tmp/test_proxmox_vms.sh
```
