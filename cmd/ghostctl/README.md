# ghostctl - Ghost Agent CLI

**ghostctl** is a command-line tool to manage Ghost Agent and VMs directly from your terminal.

## Installation

```bash
# Build
make build-cli

# Or manually
go build -o build/ghostctl cmd/ghostctl/main.go

# Install to system
sudo cp build/ghostctl /usr/local/bin/
```

## Usage

### VM Management

```bash
# List all VMs
ghostctl vm list

# Create a VM
ghostctl vm create --name my-vm --vcpu 2 --ram 4 --disk 50 --template ubuntu-22.04

# Get VM status
ghostctl vm status vm-123

# Start a VM
ghostctl vm start vm-123

# Stop a VM
ghostctl vm stop vm-123
ghostctl vm stop vm-123 --force  # Force stop (power off)

# Delete a VM
ghostctl vm delete vm-123
```

### Agent Status

```bash
# Check if Ghost Agent is running
ghostctl status

# Show version
ghostctl version
```

## Available Templates

- `ubuntu-22.04` - Ubuntu 22.04 LTS
- `ubuntu-20.04` - Ubuntu 20.04 LTS
- `debian-12` - Debian 12 (Bookworm)
- `debian-11` - Debian 11 (Bullseye)

## Examples

### Create and manage a VM

```bash
# Create a VM
ghostctl vm create \
  --name web-server \
  --vcpu 4 \
  --ram 8 \
  --disk 100 \
  --template ubuntu-22.04

# Check status
ghostctl vm status web-server

# List all VMs
ghostctl vm list

# Stop the VM
ghostctl vm stop web-server

# Delete the VM
ghostctl vm delete web-server
```

### Quick VM for testing

```bash
# Create with defaults (2 vCPU, 4GB RAM, 50GB disk, ubuntu-22.04)
ghostctl vm create --name test-vm

# Use it...

# Clean up
ghostctl vm delete test-vm
```

## Global Flags

```bash
--agent string    Ghost Agent gRPC address (default "localhost:9090")
--timeout duration Request timeout (default 30s)
```

### Examples with custom agent address

```bash
# Connect to remote agent
ghostctl --agent 192.168.1.100:9090 vm list

# Increase timeout for slow operations
ghostctl --timeout 5m vm create --name big-vm --disk 500
```

## Output

### VM List
```
VM ID                Name                           Status          IP Address
--------------------------------------------------------------------------------
vm-abc123           web-server                     running         192.168.122.10
vm-def456           database                       stopped         192.168.122.11
```

### VM Status
```
VM Status:
  ID: vm-abc123
  Name: web-server
  Status: running
  vCPU: 4
  RAM: 8 GB
  Disk: 100 GB
  IP Address: 192.168.122.10
  Uptime: 3600 seconds
  CPU Usage: 25.50%
  RAM Usage: 60.20%
```

## Troubleshooting

### "failed to connect to Ghost Agent"

Make sure Ghost Agent is running:
```bash
sudo systemctl status ghost-agent
```

If not running:
```bash
sudo systemctl start ghost-agent
```

### Permission denied

Some operations require the agent to be running as root. Make sure:
```bash
sudo systemctl start ghost-agent
```

## Integration with Scripts

```bash
#!/bin/bash

# Create multiple VMs
for i in {1..5}; do
  ghostctl vm create --name "worker-$i" --vcpu 2 --ram 4 --disk 50
done

# List all VMs
ghostctl vm list

# Get VM IDs
VM_IDS=$(ghostctl vm list | tail -n +3 | awk '{print $1}')

# Stop all VMs
for vm_id in $VM_IDS; do
  ghostctl vm stop $vm_id
done
```

## See Also

- `ghost-agent` - The main daemon
- `systemctl` - Control the Ghost Agent service
- `journalctl -u ghost-agent` - View agent logs
