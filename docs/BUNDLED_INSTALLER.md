# Ghost Agent Bundled Installer Guide

> **Create and distribute a single-file installer with all dependencies bundled**

---

## 🎁 What is the Bundled Installer?

A **self-extracting shell script** that contains:
- ✅ Ghost Agent binary
- ✅ Tailscale binaries (tailscale + tailscaled)
- ✅ All installation scripts
- ✅ Configuration templates
- ✅ Everything needed for installation

**One file, zero external dependencies!**

---

## 🚀 Quick Start

### For Developers: Create the Bundle

```bash
cd ghost-cloud/bootstrap
./create-bundle.sh
```

**Output:** `ghost-agent-installer.sh` (~40MB)

### For Users: Install from Bundle

```bash
# Download and run
curl -fsSL https://your-server.com/ghost-agent-installer.sh | bash

# Or download first, then run
wget https://your-server.com/ghost-agent-installer.sh
bash ghost-agent-installer.sh
```

---

## 📦 What's Inside the Bundle?

### Bundle Structure

```
ghost-agent-installer.sh (self-extracting)
│
├── Ghost Agent binary (~10MB)
├── Tailscale binaries (~30MB)
├── Installation scripts
│   ├── install-kvm.sh
│   ├── configure-libvirt.sh
│   ├── fix-permissions.sh
│   └── fix-apparmor.sh
└── Configuration template
    └── agent.yaml.template
```

### How It Works

```
┌─────────────────────────────────────────┐
│  ghost-agent-installer.sh               │
├─────────────────────────────────────────┤
│  ┌───────────────────────────────────┐  │
│  │  Shell Script (installer logic)   │  │
│  └───────────────────────────────────┘  │
│                                         │
│  ┌───────────────────────────────────┐  │
│  │  Base64-encoded tarball           │  │
│  │  (all files compressed)           │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘

When executed:
1. Extracts tarball to /tmp
2. Runs installation steps
3. Cleans up temporary files
```

---

## 🔨 Creating the Bundle

### Step 1: Build Ghost Agent

```bash
# Build Ghost Agent binary
cd ghost-cloud/cmd/ghost-agent
go build -ldflags="-s -w" -o ghost-agent

# Move to bootstrap directory
mv ghost-agent ../../bootstrap/
```

### Step 2: Run Bundle Creator

```bash
cd ghost-cloud/bootstrap
./create-bundle.sh
```

**What it does:**
1. Downloads Tailscale binaries
2. Copies Ghost Agent binary
3. Copies installation scripts
4. Creates configuration template
5. Creates self-extracting installer
6. Bundles everything into one file

**Output:**
```
🎁 Creating Ghost Agent Installation Bundle
===========================================

📦 Step 1/5: Downloading Ghost Agent binary...
✅ Ghost Agent binary added

📦 Step 2/5: Downloading Tailscale...
✅ Tailscale binaries added

📦 Step 3/5: Adding installation scripts...
✅ Installation scripts added

📦 Step 4/5: Creating configuration template...
✅ Configuration template added

📦 Step 5/5: Creating self-extracting installer...
✅ Self-extracting installer created

===========================================
  ✅ Bundle Created Successfully!
===========================================

Output file: ghost-agent-installer.sh
Size: 38M

To install on a user's PC:
  1. Copy ghost-agent-installer.sh to user's PC
  2. Run: bash ghost-agent-installer.sh
```

---

## 📤 Distributing the Bundle

### Option 1: Web Server

```bash
# Host on your web server
scp ghost-agent-installer.sh user@your-server.com:/var/www/html/

# Users install with:
curl -fsSL https://your-server.com/ghost-agent-installer.sh | bash
```

### Option 2: GitHub Releases

```bash
# Upload to GitHub releases
gh release create v1.0.0 ghost-agent-installer.sh

# Users install with:
curl -fsSL https://github.com/you/ghost-cloud/releases/latest/download/ghost-agent-installer.sh | bash
```

### Option 3: Direct Transfer

```bash
# Copy to USB drive
cp ghost-agent-installer.sh /media/usb/

# On user's PC:
bash /media/usb/ghost-agent-installer.sh
```

---

## 🎯 User Installation Experience

### What User Sees

```bash
$ bash ghost-agent-installer.sh

=========================================
  👻 Ghost Agent Installer v1.0.0
=========================================

📦 Extracting installer files...
✅ Files extracted to /tmp/ghost-agent-install-12345

🔍 Checking prerequisites...
✅ CPU virtualization supported
✅ Sufficient disk space (120 GB available)
✅ RAM check passed (16GB total)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 1/6: Installing KVM/Libvirt...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Installing virtualization packages...
✅ KVM/QEMU/Libvirt installed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 2/6: Configuring Libvirt...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Libvirt configured successfully

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 3/6: Installing Tailscale...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Tailscale installed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 4/6: Installing Ghost Agent...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Ghost Agent installed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 5/6: Configuring Ghost Agent...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Enter PC name (e.g., johns-laptop): johns-laptop
Enter Headscale server URL: https://headscale.example.com
Enter Ghost API URL: https://100.64.0.1:8080

Your PC has: 8 CPU cores, 16GB RAM
How many CPU cores to RESERVE for yourself? [2]: 2
How much RAM (GB) to RESERVE for yourself? [4]: 4

✅ Ghost Agent configured

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 6/6: Connecting to Headscale...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Please run this command to connect to Headscale:

  sudo tailscale up --login-server=https://headscale.example.com

Press Enter after you've connected to Tailscale...

🚀 Starting Ghost Agent...
✅ Ghost Agent is running!

🧹 Cleaning up temporary files...

=========================================
  ✅ Installation Complete!
=========================================

Your PC 'johns-laptop' is now part of Ghost Cloud!

Next steps:
  1. Ask admin to approve your PC on Headscale
  2. Check agent status: sudo systemctl status ghost-agent
  3. View logs: sudo journalctl -u ghost-agent -f
```

**Total time:** ~5-10 minutes

---

## 🔧 Advanced: Customizing the Bundle

### Add Custom Scripts

```bash
# Edit create-bundle.sh
# Add your custom script to the bundle

cp my-custom-script.sh $BUNDLE_DIR/scripts/

# It will be included in the bundle
```

### Change Default Configuration

```bash
# Edit the template in create-bundle.sh
# Modify the agent.yaml.template section

cat > $BUNDLE_DIR/config/agent.yaml.template <<'EOF'
agent:
  name: "{{PC_NAME}}"
  # Add your custom defaults here
  custom_setting: "value"
EOF
```

### Add Pre-installation Checks

```bash
# Edit the installer section in create-bundle.sh
# Add custom checks before installation

# Check for specific software
if ! command -v docker &> /dev/null; then
    echo "❌ Docker required but not installed"
    exit 1
fi
```

---

## 📊 Bundle Size Breakdown

| Component | Size | Compressed |
|-----------|------|------------|
| Ghost Agent binary | 10 MB | 3 MB |
| Tailscale binaries | 30 MB | 10 MB |
| Installation scripts | 10 KB | 3 KB |
| Config templates | 2 KB | 1 KB |
| Installer script | 5 KB | 2 KB |
| **Total** | **~40 MB** | **~13 MB** |

**After base64 encoding:** ~18 MB (base64 adds ~37% overhead)

**Final installer size:** ~18-20 MB

---

## 🔒 Security Considerations

### Verify Installer Integrity

```bash
# Generate checksum when creating bundle
sha256sum ghost-agent-installer.sh > ghost-agent-installer.sh.sha256

# Users verify before running
sha256sum -c ghost-agent-installer.sh.sha256
```

### Sign the Installer

```bash
# Sign with GPG
gpg --detach-sign --armor ghost-agent-installer.sh

# Users verify signature
gpg --verify ghost-agent-installer.sh.asc ghost-agent-installer.sh
```

### HTTPS Distribution

Always distribute over HTTPS:
```bash
# ✅ Good
curl -fsSL https://your-server.com/ghost-agent-installer.sh | bash

# ❌ Bad (insecure)
curl -fsSL http://your-server.com/ghost-agent-installer.sh | bash
```

---

## 🐛 Troubleshooting

### Bundle Creation Fails

**Problem:** `create-bundle.sh` fails to download Tailscale

**Solution:**
```bash
# Check internet connection
ping pkgs.tailscale.com

# Manually download Tailscale
wget https://pkgs.tailscale.com/stable/tailscale_1.56.1_amd64.tgz

# Place in bootstrap directory
```

### Installer Extraction Fails

**Problem:** "Cannot extract archive"

**Solution:**
```bash
# Check if base64 is installed
which base64

# Check if tar is installed
which tar

# Try manual extraction
ARCHIVE_LINE=$(awk '/^__ARCHIVE_BELOW__/ {print NR + 1; exit 0; }' ghost-agent-installer.sh)
tail -n +$ARCHIVE_LINE ghost-agent-installer.sh | base64 -d | tar -xz
```

### Installation Fails Midway

**Problem:** Installation stops at Step 3

**Solution:**
```bash
# Check logs
sudo journalctl -xe

# Check what was installed
which tailscale
which ghost-agent

# Manually complete installation
cd /tmp/ghost-agent-install-*/scripts
sudo bash install-kvm.sh
```

---

## 🎯 Benefits of Bundled Installer

### For Users

✅ **One command** - No multiple downloads  
✅ **Offline capable** - Works without internet (after download)  
✅ **No dependencies** - Everything included  
✅ **Fast** - No waiting for downloads during install  
✅ **Reliable** - No broken external links

### For Developers

✅ **Version control** - Bundle specific versions together  
✅ **Easy distribution** - One file to manage  
✅ **Consistent** - Everyone gets same versions  
✅ **Portable** - Works on USB drives, air-gapped systems  
✅ **Simple** - No package repository needed

---

## 📋 Checklist: Creating Production Bundle

- [ ] Build Ghost Agent with optimizations (`-ldflags="-s -w"`)
- [ ] Test Ghost Agent binary works
- [ ] Download correct Tailscale version
- [ ] Update version number in create-bundle.sh
- [ ] Run create-bundle.sh
- [ ] Test installer on clean Ubuntu VM
- [ ] Generate SHA256 checksum
- [ ] Sign with GPG (optional)
- [ ] Upload to distribution server
- [ ] Update documentation with download URL
- [ ] Test download URL works
- [ ] Announce to users

---

## 🔗 Related Documentation

- [Ghost Agent Requirements](GHOST_AGENT_REQUIREMENTS.md)
- [Bootstrap Installation](BOOTSTRAP_INSTALLATION.md)
- [Platform Architecture](PLATFORM_ARCHITECTURE.md)

---

**Document Version:** 1.0  
**Last Updated:** 2024-12-05
