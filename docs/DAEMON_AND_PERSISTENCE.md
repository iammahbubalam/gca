# Ghost Agent - Daemon & State Persistence

## What is systemd?

**systemd** is the service manager for Linux. It's what makes Ghost Agent run as a **daemon** (background service) that:
- ✅ Starts automatically when PC boots
- ✅ Runs in the background 24/7
- ✅ Restarts automatically if it crashes
- ✅ Can be controlled with `systemctl` commands

## How Ghost Agent Acts as a Daemon

### 1. **Systemd Service File**
Location: `/etc/systemd/system/ghost-agent.service`

```ini
[Unit]
Description=Ghost Cloud Agent - VM Management Daemon
After=network-online.target libvirtd.service tailscaled.service
Requires=libvirtd.service

[Service]
Type=simple
ExecStart=/usr/local/bin/ghost-agent --config /etc/ghost/agent.yaml
Restart=always          # ← Always restart if crashes
RestartSec=10

[Install]
WantedBy=multi-user.target  # ← Start on boot
```

### 2. **State Persistence**
**Problem:** When PC shuts down and restarts, the agent needs to remember:
- Which VMs existed
- Their status before shutdown
- Report them to Ghost Core

**Solution:** Persistent VM Repository

**File:** `internal/infrastructure/storage/persistent_vm_repository.go`

```go
// Saves VM state to: /var/lib/ghost/data/vms.json
type PersistentVMRepository struct {
    vms      map[string]*entity.VM
    filePath string  // /var/lib/ghost/data/vms.json
}

// Every time a VM is created/deleted, state is saved to disk
func (r *PersistentVMRepository) Save(ctx context.Context, vm *entity.VM) error {
    r.vms[vm.ID] = vm
    return r.persist()  // ← Writes to disk
}
```

### 3. **Startup Behavior**
When PC restarts:

1. **systemd** starts `ghost-agent.service`
2. Agent loads VM state from `/var/lib/ghost/data/vms.json`
3. Agent connects to Ghost Core API
4. Agent registers and sends heartbeat with **all VMs** (including ones from before shutdown)
5. Ghost Core knows which VMs exist on this agent

## Daemon Control Commands

```bash
# Start the daemon
sudo systemctl start ghost-agent

# Stop the daemon
sudo systemctl stop ghost-agent

# Restart the daemon
sudo systemctl restart ghost-agent

# Enable auto-start on boot
sudo systemctl enable ghost-agent

# Disable auto-start
sudo systemctl disable ghost-agent

# Check status
sudo systemctl status ghost-agent

# View logs
sudo journalctl -u ghost-agent -f
```

## How It Works

### Normal Operation
```
1. PC boots
2. systemd starts ghost-agent
3. Agent loads VMs from /var/lib/ghost/data/vms.json
4. Agent connects to Ghost Core
5. Agent sends heartbeat with VM list
6. Ghost Core knows: "This agent has VMs X, Y, Z"
```

### After PC Shutdown & Restart
```
1. PC was running VMs: vm-1, vm-2, vm-3
2. User shuts down PC
3. VM state saved to: /var/lib/ghost/data/vms.json
4. PC restarts
5. systemd auto-starts ghost-agent
6. Agent loads: vm-1, vm-2, vm-3 from disk
7. Agent sends heartbeat to Ghost Core
8. Ghost Core receives: "Agent has vm-1, vm-2, vm-3"
9. ✅ No data lost!
```

## File Locations

```
/usr/local/bin/ghost-agent              # Binary
/etc/ghost/agent.yaml                   # Configuration
/etc/systemd/system/ghost-agent.service # Systemd service
/var/lib/ghost/data/vms.json           # VM state (persisted)
/var/lib/ghost/images/                  # OS images cache
/var/log/ghost/agent.log                # Logs
```

## Why No Docker?

Ghost Agent runs **directly on the PC** (not in Docker) because:

1. **Needs KVM/Libvirt access** - Must access hardware virtualization
2. **Needs Tailscale** - Must be on the VPN network
3. **Daemon behavior** - systemd manages it like any other system service
4. **Performance** - No container overhead

## Installation

The automated installer (`scripts/install.sh`) does everything:

```bash
sudo ./scripts/install.sh
```

This:
1. Installs Libvirt, KVM, Tailscale
2. Copies binary to `/usr/local/bin/`
3. Creates systemd service
4. Enables auto-start on boot
5. Creates all directories
6. Starts the daemon

## Summary

- ✅ **systemd** = Makes it a daemon (background service)
- ✅ **Persistent storage** = Remembers VMs after restart
- ✅ **Auto-start** = Runs on PC boot
- ✅ **Auto-restart** = Recovers from crashes
- ✅ **No Docker** = Runs directly on PC
- ✅ **State sync** = Reports all VMs to Ghost Core on startup

**Ghost Agent is a true system daemon!** 🚀
