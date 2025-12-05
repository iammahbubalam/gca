# Ghost Agent Bootstrap Installation Guide

> **Complete guide for installing Ghost Agent prerequisites and the agent itself**

---

## 📋 Overview

The bootstrap scripts install everything needed to run Ghost Agent on a user's PC:

1. **KVM/QEMU/Libvirt** - Hypervisor for creating VMs
2. **Tailscale** - VPN client for Headscale network
3. **Ghost Agent** - The agent software
4. **Configuration** - Setup and configuration

---

## 🚀 Quick Start

### One-Command Installation

```bash
# Download and run installer
curl -fsSL https://your-server.com/install.sh | bash
```

### Manual Installation

```bash
# Clone repository
git clone https://github.com/you/ghost-cloud.git
cd ghost-cloud/bootstrap

# Run installer
./install.sh
```

---

## 📦 What Gets Installed

### System Packages

| Package | Size | Purpose | Required |
|---------|------|---------|----------|
| **qemu-kvm** | ~15MB | KVM hypervisor | ✅ Yes |
| **libvirt-daemon-system** | ~20MB | VM management daemon | ✅ Yes |
| **libvirt-clients** | ~5MB | Libvirt CLI tools | ✅ Yes |
| **bridge-utils** | ~1MB | Network bridging | ✅ Yes |
| **cpu-checker** | ~1MB | Check CPU virtualization | ✅ Yes |
| **cloud-image-utils** | ~5MB | Cloud-init tools | ✅ Yes |
| **qemu-utils** | ~10MB | QEMU utilities | ✅ Yes |
| **genisoimage** | ~5MB | Create ISO files | ✅ Yes |
| **tailscale** | ~30MB | VPN client | ✅ Yes |
| **ghost-agent** | ~10MB | Ghost Agent binary | ✅ Yes |

**Total:** ~100MB download, ~200MB installed

---

## 📜 Bootstrap Scripts

### 1. install-kvm.sh

**Purpose:** Install KVM, QEMU, and Libvirt packages

**What it does:**
- Updates package list
- Installs virtualization packages
- Checks CPU virtualization support
- Adds user to libvirt groups

**Packages installed:**
```bash
qemu-kvm                # KVM hypervisor
libvirt-daemon-system   # Libvirt daemon
libvirt-clients         # virsh, virt-install
bridge-utils            # Network bridging
cpu-checker             # kvm-ok command
cloud-image-utils       # cloud-localds
qemu-utils              # qemu-img
genisoimage             # mkisofs
```

**User groups added:**
- `libvirt` - Access to Libvirt
- `kvm` - Access to KVM
- `libvirt-qemu` - Access to QEMU

**Usage:**
```bash
./install-kvm.sh
```

**Output:**
```
Installing virtualization packages...
✅ KVM virtualization supported
✅ KVM/QEMU/Libvirt installed
```

---

### 2. configure-libvirt.sh

**Purpose:** Configure Libvirt networks and storage pools

**What it does:**
- Starts libvirtd service
- Creates default network (192.168.122.0/24)
- Creates default storage pool (/var/lib/libvirt/images)
- Enables autostart for both

**Network configuration:**
```
Name: default
Type: NAT
IP Range: 192.168.122.0/24
Gateway: 192.168.122.1
DHCP: 192.168.122.2 - 192.168.122.254
```

**Storage configuration:**
```
Name: default
Type: Directory
Path: /var/lib/libvirt/images
```

**Usage:**
```bash
./configure-libvirt.sh
```

**Output:**
```
Configuring Libvirt...
✅ Libvirtd is running
✅ Libvirt configured successfully

Verification:
  Libvirtd: active
  Default network: yes
  Default pool: running
```

---

### 3. install-tailscale.sh

**Purpose:** Install Tailscale VPN client

**What it does:**
- Downloads Tailscale installer
- Installs Tailscale
- Does NOT connect (user must do this)

**Usage:**
```bash
./install-tailscale.sh
```

**After installation, user must connect:**
```bash
sudo tailscale up --login-server=https://your-headscale-server.com
```

---

### 4. install-ghost-agent.sh

**Purpose:** Download and install Ghost Agent binary

**What it does:**
- Detects system architecture (amd64/arm64)
- Downloads Ghost Agent binary from GitHub releases
- Installs to /usr/local/bin/ghost-agent
- Makes executable

**Usage:**
```bash
./install-ghost-agent.sh
```

**Output:**
```
Downloading Ghost Agent...
✅ Ghost Agent installed to /usr/local/bin/ghost-agent
```

---

### 5. configure-ghost-agent.sh

**Purpose:** Configure Ghost Agent

**What it does:**
- Prompts user for configuration
- Creates /etc/ghost/agent.yaml
- Creates systemd service
- Enables and starts service

**User prompts:**
```
Enter PC name: johns-laptop
Enter Headscale URL: https://headscale.example.com
Enter API URL: https://100.64.0.1:8080
Reserved CPU cores: 2
Reserved RAM (GB): 4
```

**Creates:**
- `/etc/ghost/agent.yaml` - Configuration file
- `/etc/systemd/system/ghost-agent.service` - Systemd service
- `/var/log/ghost/` - Log directory
- `/var/lib/ghost/images/` - Image cache directory

**Usage:**
```bash
./configure-ghost-agent.sh
```

---

### 6. fix-permissions.sh (Optional)

**Purpose:** Fix file permissions for Libvirt

**When to use:** Only if you encounter permission errors

**What it does:**
- Sets ownership on /var/lib/libvirt/images to libvirt-qemu:kvm
- Sets directory permissions to 771
- Sets file permissions to 644
- Restarts libvirtd

**Usage:**
```bash
./fix-permissions.sh
```

---

### 7. fix-apparmor.sh (Optional)

**Purpose:** Disable AppArmor for Libvirt

**When to use:** Only if you encounter AppArmor-related errors

**What it does:**
- Disables AppArmor profile for libvirtd
- Restarts libvirtd

**Usage:**
```bash
./fix-apparmor.sh
```

---

### 8. install.sh (Main Installer)

**Purpose:** Run all installation steps in order

**What it does:**
```
Step 1: Install KVM/QEMU/Libvirt
Step 2: Configure Libvirt
Step 3: Install Tailscale
Step 4: Install Ghost Agent
Step 5: Configure Ghost Agent
```

**Usage:**
```bash
./install.sh
```

**Full output:**
```
=========================================
  👻 Ghost Cloud Installation
=========================================

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 1/5: Installing KVM/QEMU/Libvirt...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ KVM/QEMU/Libvirt installed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 2/5: Configuring Libvirt...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Libvirt configured successfully

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 3/5: Installing Tailscale...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Tailscale installed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 4/5: Installing Ghost Agent...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Ghost Agent installed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 5/5: Configuring Ghost Agent...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Enter PC name: johns-laptop
Enter Headscale URL: https://headscale.example.com
Enter API URL: https://100.64.0.1:8080
Reserved CPU cores [2]: 2
Reserved RAM (GB) [4]: 4
✅ Ghost Agent configured

=========================================
  ✅ Installation Complete!
=========================================

Next steps:
  1. Connect to Headscale VPN:
     sudo tailscale up --login-server=https://headscale.example.com
  
  2. Check agent status:
     sudo systemctl status ghost-agent
  
  3. View logs:
     sudo journalctl -u ghost-agent -f
```

---

## 🔍 Prerequisites Check

Before running installer, verify:

### 1. CPU Virtualization Support

```bash
# Check if CPU supports virtualization
egrep -c '(vmx|svm)' /proc/cpuinfo

# Should return > 0
# If 0, enable VT-x/AMD-V in BIOS
```

### 2. Not Running in a VM

```bash
# Check if running in a VM
systemd-detect-virt

# Should return "none"
# If returns "kvm", "vmware", etc., you're in a VM
# Nested virtualization may not work well
```

### 3. Sufficient Resources

```bash
# Check CPU cores
nproc
# Recommended: 4+

# Check RAM
free -g | awk '/^Mem:/{print $2}'
# Recommended: 8+ GB

# Check disk space
df -h / | awk 'NR==2 {print $4}'
# Recommended: 50+ GB free
```

---

## 📂 Directory Structure After Installation

```
/usr/local/bin/
└── ghost-agent                    # Agent binary

/etc/ghost/
├── agent.yaml                     # Configuration
└── certs/                         # TLS certificates (if using mTLS)
    ├── agent.crt
    ├── agent.key
    └── ca.crt

/var/lib/ghost/
└── images/                        # Cached OS images
    ├── base-ubuntu-22.04.qcow2
    └── base-debian-12.qcow2

/var/log/ghost/
└── agent.log                      # Agent logs

/var/lib/libvirt/
├── images/                        # VM disk images
│   ├── web-1.qcow2
│   └── db-1.qcow2
└── cloud-init/                    # Cloud-init ISOs
    ├── web-1-cloud-init.iso
    └── db-1-cloud-init.iso

/etc/systemd/system/
└── ghost-agent.service            # Systemd service
```

---

## 🛠️ Troubleshooting

### Issue: "KVM not supported"

**Solution:**
```bash
# Check CPU support
kvm-ok

# If not supported, enable VT-x/AMD-V in BIOS
```

### Issue: "Permission denied" errors

**Solution:**
```bash
# Run permission fix script
./fix-permissions.sh

# Add user to groups
sudo usermod -aG libvirt,kvm $USER

# Log out and log back in
```

### Issue: "Cannot connect to libvirtd"

**Solution:**
```bash
# Check libvirtd status
sudo systemctl status libvirtd

# Restart libvirtd
sudo systemctl restart libvirtd

# Check socket
ls -l /var/run/libvirt/libvirt-sock
```

### Issue: AppArmor blocking Libvirt

**Solution:**
```bash
# Run AppArmor fix script
./fix-apparmor.sh
```

### Issue: "Network 'default' not found"

**Solution:**
```bash
# Re-run configure script
./configure-libvirt.sh
```

---

## 🔄 Updating Ghost Agent

```bash
# Download new version
wget https://github.com/you/ghost-cloud/releases/latest/download/ghost-agent-linux-amd64

# Stop agent
sudo systemctl stop ghost-agent

# Replace binary
sudo install -m 755 ghost-agent-linux-amd64 /usr/local/bin/ghost-agent

# Start agent
sudo systemctl start ghost-agent

# Verify
sudo systemctl status ghost-agent
```

---

## 🗑️ Uninstallation

```bash
# Stop and disable agent
sudo systemctl stop ghost-agent
sudo systemctl disable ghost-agent

# Remove agent
sudo rm /usr/local/bin/ghost-agent
sudo rm /etc/systemd/system/ghost-agent.service
sudo rm -rf /etc/ghost
sudo rm -rf /var/lib/ghost
sudo rm -rf /var/log/ghost

# Optionally remove KVM/Libvirt
sudo apt-get remove --purge qemu-kvm libvirt-daemon-system libvirt-clients

# Optionally remove Tailscale
sudo apt-get remove --purge tailscale
```

---

## 📊 Verification Checklist

After installation, verify everything works:

- [ ] KVM installed: `kvm-ok`
- [ ] Libvirtd running: `systemctl is-active libvirtd`
- [ ] Default network active: `virsh net-list`
- [ ] Default pool active: `virsh pool-list`
- [ ] Tailscale installed: `tailscale version`
- [ ] Tailscale connected: `tailscale status`
- [ ] Ghost Agent installed: `ghost-agent --version`
- [ ] Ghost Agent running: `systemctl is-active ghost-agent`
- [ ] Agent logs clean: `journalctl -u ghost-agent -n 50`

---

## 🎯 Next Steps After Installation

1. **Connect to Headscale:**
   ```bash
   sudo tailscale up --login-server=https://your-headscale-server.com
   ```

2. **Verify agent connection:**
   ```bash
   sudo journalctl -u ghost-agent -f
   # Should see: "Registered with API"
   # Should see: "Heartbeat sent"
   ```

3. **Check on Ghost API:**
   ```bash
   # On control PC
   ghost list-hosts
   # Should see your PC listed
   ```

4. **Test VM creation:**
   ```bash
   # On control PC
   ghost create test-vm --flavor small
   # Should create VM on your PC
   ```

---

## 📝 Notes

- **Reboot required:** No! Everything works immediately
- **Log out required:** Yes, for group permissions to take effect
- **Internet required:** Yes, for downloading packages and images
- **Sudo required:** Yes, for installation only
- **Time required:** ~5-10 minutes

---

## 🔗 Related Documentation

- [Ghost Agent Requirements](GHOST_AGENT_REQUIREMENTS.md)
- [Platform Architecture](PLATFORM_ARCHITECTURE.md)
- [VM Creation Technical Guide](VM_CREATION_TECHNICAL_GUIDE.md)
- [Complete Lifecycle](COMPLETE_LIFECYCLE.md)

---

**Document Version:** 1.0  
**Last Updated:** 2024-12-05
