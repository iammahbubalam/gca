# Ghost Agent Installation Guide

This document explains how to install, run, and uninstall Ghost Agent.

## Prerequisites

- Go 1.23+
- Libvirt/KVM installed
- Linux system with systemd (optional, for service mode)

---

## Installation

### 1. Build the binaries

```bash
cd /home/morphosis/projects/ghost-agent
make build
```

This creates:
- `build/ghost-agent` - The daemon
- `build/ghostctl` - The CLI tool

### 2. Create required directories

```bash
sudo mkdir -p /etc/ghost
sudo mkdir -p /var/log/ghost
sudo mkdir -p /var/lib/ghost/images
```

### 3. Copy configuration

```bash
sudo cp configs/agent.yaml /etc/ghost/agent.yaml
```

### 4. Install binaries

```bash
sudo install -m 755 build/ghost-agent /usr/local/bin/ghost-agent
sudo install -m 755 build/ghostctl /usr/local/bin/ghostctl
```

### 5. Set permissions (for non-root usage)

```bash
sudo chown -R $USER:$USER /var/log/ghost /var/lib/ghost
```

---

## What Gets Installed

| Component | Location | Description |
|-----------|----------|-------------|
| `ghost-agent` | `/usr/local/bin/ghost-agent` | Main daemon binary |
| `ghostctl` | `/usr/local/bin/ghostctl` | CLI management tool |
| Config | `/etc/ghost/agent.yaml` | Configuration file |
| Logs | `/var/log/ghost/` | Log files directory |
| Data | `/var/lib/ghost/images/` | VM images cache |

---

## Running Ghost Agent

### Option 1: Run directly (foreground)

```bash
ghost-agent
```

Or with sudo for full libvirt access:

```bash
sudo ghost-agent
```

### Option 2: Run as systemd service

```bash
sudo cp deployments/systemd/ghost-agent.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable ghost-agent
sudo systemctl start ghost-agent
sudo systemctl status ghost-agent
```

---

## Using ghostctl CLI

```bash
# Check agent status
ghostctl status

# Show version
ghostctl version

# List VMs
ghostctl vm list

# Start a VM
ghostctl vm start <vm-id>

# Stop a VM
ghostctl vm stop <vm-id>

# Delete a VM
ghostctl vm delete <vm-id>

# Create a VM (see help for options)
ghostctl vm create --help
```

### Connection Options

```bash
# Connect to a different agent
ghostctl --agent localhost:9090 vm list

# Set custom timeout
ghostctl --timeout 60s vm list
```

---

## Uninstallation

### 1. Stop the service (if running as systemd)

```bash
sudo systemctl stop ghost-agent
sudo systemctl disable ghost-agent
sudo rm /etc/systemd/system/ghost-agent.service
sudo systemctl daemon-reload
```

### 2. Remove binaries

```bash
sudo rm -f /usr/local/bin/ghost-agent
sudo rm -f /usr/local/bin/ghostctl
```

### 3. Remove configuration and data

```bash
sudo rm -rf /etc/ghost
sudo rm -rf /var/log/ghost
sudo rm -rf /var/lib/ghost
```

### Quick Uninstall (one-liner)

```bash
sudo systemctl stop ghost-agent 2>/dev/null; \
sudo systemctl disable ghost-agent 2>/dev/null; \
sudo rm -f /etc/systemd/system/ghost-agent.service \
           /usr/local/bin/ghost-agent \
           /usr/local/bin/ghostctl; \
sudo rm -rf /etc/ghost /var/log/ghost /var/lib/ghost; \
sudo systemctl daemon-reload
```

---

## Troubleshooting

### "permission denied" for log file

```bash
sudo chown -R $USER:$USER /var/log/ghost /var/lib/ghost
```

### "address already in use" for health server

Another process is using port 9092. Check with:

```bash
sudo lsof -i :9092
```

### ghostctl can't connect

Make sure ghost-agent is running and listening on port 9090:

```bash
ghostctl --agent localhost:9090 status
```

---

## Configuration

Edit `/etc/ghost/agent.yaml` to configure:

- Agent name and API URL
- Libvirt connection settings
- gRPC server address and TLS
- Logging options
- Resource reservations

See `configs/agent.yaml` for all available options.
