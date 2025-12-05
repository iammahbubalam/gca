#!/bin/bash
set -e

# Ghost Agent - Automated Installation Script
# This script installs and configures everything needed for Ghost Agent

echo "========================================="
echo "Ghost Agent - Automated Installation"
echo "========================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Please run as root (use sudo)"
    exit 1
fi

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    VERSION=$VERSION_ID
else
    echo "❌ Cannot detect OS"
    exit 1
fi

echo "✅ Detected OS: $OS $VERSION"
echo ""

# ============================================
# 1. Install System Dependencies
# ============================================
echo "📦 Installing system dependencies..."

if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
    apt-get update
    apt-get install -y \
        qemu-kvm \
        libvirt-daemon-system \
        libvirt-clients \
        bridge-utils \
        virt-manager \
        cpu-checker \
        curl \
        wget \
        gnupg \
        ca-certificates
    
    echo "✅ System dependencies installed"
else
    echo "❌ Unsupported OS: $OS"
    exit 1
fi

# ============================================
# 2. Verify KVM Support
# ============================================
echo ""
echo "🔍 Checking KVM support..."

if kvm-ok &>/dev/null; then
    echo "✅ KVM acceleration is supported"
else
    echo "⚠️  KVM acceleration may not be available"
    echo "   The agent will still work but VMs will be slower"
fi

# ============================================
# 3. Configure Libvirt
# ============================================
echo ""
echo "⚙️  Configuring Libvirt..."

# Start and enable libvirtd
systemctl enable libvirtd
systemctl start libvirtd

# Add current user to libvirt group (if not root)
if [ -n "$SUDO_USER" ]; then
    usermod -aG libvirt $SUDO_USER
    usermod -aG kvm $SUDO_USER
    echo "✅ Added $SUDO_USER to libvirt and kvm groups"
fi

# Configure default network
virsh net-autostart default || true
virsh net-start default || true

echo "✅ Libvirt configured"

# ============================================
# 4. Install Tailscale
# ============================================
echo ""
echo "🔗 Installing Tailscale..."

if ! command -v tailscale &> /dev/null; then
    curl -fsSL https://tailscale.com/install.sh | sh
    echo "✅ Tailscale installed"
else
    echo "✅ Tailscale already installed"
fi

# ============================================
# 5. Configure Tailscale (Manual step required)
# ============================================
echo ""
echo "⚠️  Tailscale Configuration Required:"
echo "   Run: sudo tailscale up --accept-routes"
echo "   Then authenticate in your browser"
echo ""
read -p "Press Enter after Tailscale is connected..."

# Verify Tailscale connection
if tailscale status &>/dev/null; then
    TAILSCALE_IP=$(tailscale ip -4)
    echo "✅ Tailscale connected: $TAILSCALE_IP"
else
    echo "❌ Tailscale not connected. Please run: sudo tailscale up"
    exit 1
fi

# ============================================
# 6. Create Ghost Agent Directories
# ============================================
echo ""
echo "📁 Creating Ghost Agent directories..."

mkdir -p /etc/ghost
mkdir -p /var/log/ghost
mkdir -p /var/lib/ghost/images
mkdir -p /var/lib/ghost/disks

chmod 755 /etc/ghost
chmod 755 /var/log/ghost
chmod 755 /var/lib/ghost

echo "✅ Directories created"

# ============================================
# 7. Install Ghost Agent Binary
# ============================================
echo ""
echo "📥 Installing Ghost Agent binary..."

INSTALL_DIR="/usr/local/bin"
BINARY_PATH="./build/ghost-agent"

if [ -f "$BINARY_PATH" ]; then
    install -m 755 $BINARY_PATH $INSTALL_DIR/ghost-agent
    echo "✅ Ghost Agent binary installed to $INSTALL_DIR/ghost-agent"
else
    echo "❌ Binary not found at $BINARY_PATH"
    echo "   Please build the agent first: make build"
    exit 1
fi

# ============================================
# 8. Configure Ghost Agent
# ============================================
echo ""
echo "⚙️  Configuring Ghost Agent..."

# Get configuration from user
read -p "Enter agent name (default: $(hostname)): " AGENT_NAME
AGENT_NAME=${AGENT_NAME:-$(hostname)}

read -p "Enter Ghost API URL (e.g., https://100.64.0.1:8080): " API_URL
if [ -z "$API_URL" ]; then
    echo "❌ API URL is required"
    exit 1
fi

# Create configuration file
cat > /etc/ghost/agent.yaml <<EOF
# Ghost Agent Configuration
agent:
  name: "$AGENT_NAME"
  api_url: "$API_URL"
  heartbeat_interval: 30s
  version: "1.0.0"

libvirt:
  uri: "qemu:///system"
  storage_pool: "default"
  network: "default"
  image_cache: "/var/lib/ghost/images"

resources:
  reserved_cpu: 2
  reserved_ram_gb: 4
  reserved_disk_gb: 50

grpc:
  listen_addr: "0.0.0.0:9090"
  tls_enabled: false
  tls_cert: "/etc/ghost/certs/agent.crt"
  tls_key: "/etc/ghost/certs/agent.key"
  tls_ca: "/etc/ghost/certs/ca.crt"

logging:
  level: "info"
  output: "both"
  file: "/var/log/ghost/agent.log"
  format: "json"
  max_size: 100
  max_age: 7

metrics:
  enabled: true
  listen_addr: "0.0.0.0:9091"
  path: "/metrics"

health:
  listen_addr: "0.0.0.0:9092"
  path: "/health"
EOF

chmod 644 /etc/ghost/agent.yaml
echo "✅ Configuration created at /etc/ghost/agent.yaml"

# ============================================
# 9. Install Systemd Service
# ============================================
echo ""
echo "🔧 Installing systemd service..."

cat > /etc/systemd/system/ghost-agent.service <<EOF
[Unit]
Description=Ghost Cloud Agent
Documentation=https://github.com/iammahbubalam/ghost-agent
After=network.target libvirtd.service tailscaled.service
Requires=libvirtd.service
Wants=tailscaled.service

[Service]
Type=simple
User=root
Group=root

ExecStart=/usr/local/bin/ghost-agent --config /etc/ghost/agent.yaml

Restart=always
RestartSec=10

LimitNOFILE=65536
LimitNPROC=4096

StandardOutput=journal
StandardError=journal
SyslogIdentifier=ghost-agent

NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable ghost-agent

echo "✅ Systemd service installed"

# ============================================
# 10. Configure Firewall (if UFW is active)
# ============================================
echo ""
echo "🔥 Configuring firewall..."

if command -v ufw &> /dev/null && ufw status | grep -q "Status: active"; then
    ufw allow 9090/tcp comment "Ghost Agent gRPC"
    ufw allow 9091/tcp comment "Ghost Agent Metrics"
    ufw allow 9092/tcp comment "Ghost Agent Health"
    echo "✅ Firewall rules added"
else
    echo "ℹ️  UFW not active, skipping firewall configuration"
fi

# ============================================
# 11. Final Summary
# ============================================
echo ""
echo "========================================="
echo "✅ Installation Complete!"
echo "========================================="
echo ""
echo "📋 Summary:"
echo "  • Libvirt/KVM: Installed and configured"
echo "  • Tailscale: Connected ($TAILSCALE_IP)"
echo "  • Ghost Agent: Installed at /usr/local/bin/ghost-agent"
echo "  • Configuration: /etc/ghost/agent.yaml"
echo "  • Service: ghost-agent.service"
echo ""
echo "🚀 Next Steps:"
echo "  1. Start the agent:"
echo "     sudo systemctl start ghost-agent"
echo ""
echo "  2. Check status:"
echo "     sudo systemctl status ghost-agent"
echo ""
echo "  3. View logs:"
echo "     sudo journalctl -u ghost-agent -f"
echo ""
echo "  4. Check health:"
echo "     curl http://localhost:9092/health"
echo ""
echo "  5. View metrics:"
echo "     curl http://localhost:9091/metrics"
echo ""
echo "========================================="
echo ""

# Ask if user wants to start the service now
read -p "Start Ghost Agent now? (y/n): " START_NOW
if [ "$START_NOW" = "y" ] || [ "$START_NOW" = "Y" ]; then
    systemctl start ghost-agent
    sleep 2
    systemctl status ghost-agent --no-pager
    echo ""
    echo "✅ Ghost Agent is running!"
else
    echo "ℹ️  You can start it later with: sudo systemctl start ghost-agent"
fi

echo ""
echo "🎉 Installation complete!"
