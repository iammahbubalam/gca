# Ghost Agent - Complete Requirements Specification

> **Purpose:** This document provides complete functional and non-functional requirements for building the Ghost Agent, a distributed VM management agent that runs on user PCs and communicates with the Ghost Cloud API.

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Functional Requirements](#functional-requirements)
3. [Non-Functional Requirements](#non-functional-requirements)
4. [Technical Architecture](#technical-architecture)
5. [Dependencies](#dependencies)
6. [Build Instructions](#build-instructions)
7. [API Specifications](#api-specifications)
8. [Configuration](#configuration)
9. [Deployment](#deployment)
10. [Testing Requirements](#testing-requirements)

---

## 🎯 Overview

### What is Ghost Agent?

Ghost Agent is a **production-grade daemon service** that runs on user PCs to:
- Create and manage virtual machines using KVM/Libvirt
- Communicate with Ghost Cloud API via gRPC over Headscale VPN
- Report system resources and health status
- Execute VM lifecycle operations (create, start, stop, delete)
- Provide observability through structured logging and metrics

### Design Principles

- **Production-Ready:** Industry-standard patterns, comprehensive error handling, graceful degradation
- **Secure by Default:** mTLS authentication, input validation, principle of least privilege
- **Observable:** Structured logging, metrics, distributed tracing, health checks
- **Extensible:** Clean architecture, dependency injection, plugin-ready design
- **Reliable:** Retry logic, circuit breakers, graceful shutdown, state recovery
- **Testable:** High test coverage, integration tests, contract testing

### Key Characteristics

- **Language:** Go (Golang) 1.21+
- **Platform:** Linux (Ubuntu, Debian, Fedora, RHEL)
- **Architecture:** x86_64 (amd64), ARM64
- **Deployment:** Systemd service with auto-restart
- **Communication:** gRPC with mTLS over Headscale VPN
- **Hypervisor:** KVM via Libvirt
- **Architecture Pattern:** Clean Architecture / Hexagonal Architecture
- **Networking (v1.0):** Simple NAT (designed for future extensibility)

---

## ✅ Functional Requirements

### FR1: VM Lifecycle Management

#### FR1.1: Create VM
**Description:** Create a new virtual machine with specified resources.

**Inputs:**
- VM name (string, unique)
- vCPU count (integer, 1-32)
- RAM size (integer, GB, 1-128)
- Disk size (integer, GB, 10-1000)
- OS template (string: ubuntu-22.04, debian-12, etc.)

**Process:**
1. Download OS cloud image (if not cached)
2. Create copy-on-write disk from base image
3. Generate cloud-init configuration
4. Define VM in Libvirt
5. Start VM
6. Wait for IP address assignment (timeout: 2 minutes)
7. Return VM details

**Outputs:**
- VM ID (string)
- IP address (string)
- Status (string: running, error)
- Error message (if failed)

**Success Criteria:**
- VM boots successfully
- VM gets IP address within 2 minutes
- VM is accessible via network

---

#### FR1.2: Delete VM
**Description:** Permanently delete a virtual machine.

**Inputs:**
- VM name or ID (string)

**Process:**
1. Stop VM if running
2. Undefine VM from Libvirt
3. Delete VM disk files
4. Delete cloud-init ISO
5. Return success status

**Outputs:**
- Success (boolean)
- Error message (if failed)

**Success Criteria:**
- VM is stopped
- All VM files are deleted
- Resources are freed

---

#### FR1.3: Start VM
**Description:** Start a stopped virtual machine.

**Inputs:**
- VM name or ID (string)

**Process:**
1. Verify VM exists
2. Start VM via Libvirt
3. Wait for VM to boot
4. Return status

**Outputs:**
- Status (string: running, error)
- Error message (if failed)

---

#### FR1.4: Stop VM
**Description:** Gracefully stop a running virtual machine.

**Inputs:**
- VM name or ID (string)
- Force flag (boolean, optional)

**Process:**
1. Send ACPI shutdown signal (if not forced)
2. Wait for graceful shutdown (timeout: 30 seconds)
3. Force destroy if timeout or force flag set
4. Return status

**Outputs:**
- Status (string: stopped, error)
- Error message (if failed)

---

#### FR1.5: Get VM Status
**Description:** Retrieve current status and details of a VM.

**Inputs:**
- VM name or ID (string)

**Outputs:**
- VM name
- Status (running, stopped, paused, error)
- vCPU count
- RAM size
- Disk size
- IP address
- Uptime
- Resource usage (CPU %, RAM %)

---

### FR2: Resource Management

#### FR2.1: Resource Discovery
**Description:** Detect and report available system resources.

**Process:**
1. Query total CPU cores
2. Query total RAM
3. Query total disk space
4. Calculate available resources (total - reserved - used)
5. Return resource information

**Outputs:**
- Total CPU cores
- Available CPU cores
- Total RAM (GB)
- Available RAM (GB)
- Total disk (GB)
- Available disk (GB)

---

#### FR2.2: Resource Reservation
**Description:** Reserve resources for the PC owner.

**Configuration:**
- `reserved_cpu`: Number of CPU cores to reserve
- `reserved_ram_gb`: Amount of RAM to reserve (GB)
- `reserved_disk_gb`: Amount of disk to reserve (GB)

**Behavior:**
- Never allocate reserved resources to VMs
- Ensure owner's applications always have resources

---

#### FR2.3: Heartbeat Reporting
**Description:** Periodically report agent health and resources to API.

**Frequency:** Every 30 seconds

**Data Sent:**
- Agent ID
- Timestamp
- Available CPU
- Available RAM
- Available disk
- Running VMs count
- Agent version
- Status (online, degraded, offline)

---

### FR3: Communication

#### FR3.1: gRPC Server
**Description:** Expose gRPC API for receiving commands from Ghost API.

**Port:** 9090 (configurable)

**Endpoints:**
- `CreateVM(CreateVMRequest) → CreateVMResponse`
- `DeleteVM(DeleteVMRequest) → DeleteVMResponse`
- `StartVM(StartVMRequest) → StartVMResponse`
- `StopVM(StopVMRequest) → StopVMResponse`
- `GetVMStatus(GetVMStatusRequest) → GetVMStatusResponse`
- `ListVMs(ListVMsRequest) → ListVMsResponse`

---

#### FR3.2: API Client
**Description:** Connect to Ghost API for registration and heartbeats.

**Operations:**
- Register agent on startup
- Send heartbeats every 30 seconds
- Report VM status changes
- Handle API connection failures (retry with backoff)

---

### FR4: Image Management

#### FR4.1: Image Caching
**Description:** Cache OS cloud images locally to speed up VM creation.

**Cache Location:** `/var/lib/ghost/images/`

**Behavior:**
- Download image on first use
- Reuse cached image for subsequent VMs
- Verify image checksum before use

**Supported Images:**
- Ubuntu 22.04 LTS
- Ubuntu 20.04 LTS
- Debian 12
- Debian 11
- CentOS Stream 9

---

#### FR4.2: Image Updates
**Description:** Update cached images when new versions are available.

**Trigger:** Manual or scheduled (weekly)

**Process:**
1. Check for new image versions
2. Download new image
3. Verify checksum
4. Replace old image
5. Clean up old image

---

### FR5: Networking

#### FR5.1: VM Network Configuration
**Description:** Configure network for VMs using Libvirt default network.

**Network Type:** NAT (Network Address Translation)

**IP Range:** 192.168.122.0/24 (Libvirt default)

**Features:**
- DHCP for automatic IP assignment
- DNS forwarding
- Outbound internet access via NAT

---

#### FR5.2: Headscale VPN Integration
**Description:** Communicate with Ghost API over Headscale VPN.

**Requirements:**
- Tailscale client installed
- Connected to Headscale server
- Agent uses Headscale IP for API communication

---

### FR6: Logging and Monitoring

#### FR6.1: Logging
**Description:** Log all agent activities for debugging and auditing.

**Log Levels:**
- DEBUG: Detailed information
- INFO: General information
- WARN: Warning messages
- ERROR: Error messages

**Log Destinations:**
- File: `/var/log/ghost/agent.log`
- Systemd journal: `journalctl -u ghost-agent`

**Log Rotation:**
- Max size: 100MB
- Keep last 7 days

---

#### FR6.2: Metrics Collection
**Description:** Collect and expose metrics for monitoring.

**Metrics:**
- VMs created (counter)
- VMs deleted (counter)
- VMs running (gauge)
- CPU usage per VM (gauge)
- RAM usage per VM (gauge)
- API call latency (histogram)
- Heartbeat success rate (gauge)

**Exposure:** Prometheus metrics endpoint (optional)

---

## 🏗️ Non-Functional Requirements

### NFR1: Performance

#### NFR1.1: VM Creation Time
- **Requirement:** Create VM in ≤ 60 seconds (first time)
- **Requirement:** Create VM in ≤ 40 seconds (cached image)

#### NFR1.2: API Response Time
- **Requirement:** gRPC calls respond in ≤ 100ms (excluding VM operations)
- **Requirement:** Heartbeat completes in ≤ 500ms

#### NFR1.3: Resource Overhead
- **Requirement:** Agent uses ≤ 50MB RAM when idle
- **Requirement:** Agent uses ≤ 5% CPU when idle

---

### NFR2: Reliability

#### NFR2.1: Uptime
- **Requirement:** Agent uptime ≥ 99.9%
- **Requirement:** Auto-restart on crash (systemd)

#### NFR2.2: Error Handling
- **Requirement:** Gracefully handle all errors
- **Requirement:** Retry failed operations with exponential backoff
- **Requirement:** Never crash on invalid input

#### NFR2.3: Data Integrity
- **Requirement:** Never corrupt VM disks
- **Requirement:** Atomic operations (all or nothing)

---

### NFR3: Security

#### NFR3.1: Authentication
- **Requirement:** mTLS for gRPC communication
- **Requirement:** Verify API server certificate
- **Requirement:** Present agent certificate

#### NFR3.2: Authorization
- **Requirement:** Only accept commands from authenticated API
- **Requirement:** Validate all input parameters

#### NFR3.3: Isolation
- **Requirement:** VMs cannot access host filesystem
- **Requirement:** VMs cannot access other VMs' data
- **Requirement:** Network isolation between VMs (optional)

---

### NFR4: Scalability

#### NFR4.1: VM Capacity
- **Requirement:** Support up to 20 VMs per agent (depending on hardware)
- **Requirement:** Handle concurrent VM operations

#### NFR4.2: Network Capacity
- **Requirement:** Handle 100 gRPC requests/second

---

### NFR5: Maintainability

#### NFR5.1: Code Quality
- **Requirement:** Go code follows standard conventions
- **Requirement:** 80% test coverage
- **Requirement:** All public functions documented

#### NFR5.2: Configuration
- **Requirement:** All settings configurable via YAML file
- **Requirement:** Validate configuration on startup
- **Requirement:** Provide sensible defaults

#### NFR5.3: Upgrades
- **Requirement:** Support in-place upgrades
- **Requirement:** Preserve running VMs during upgrade
- **Requirement:** Rollback capability

---

### NFR6: Observability

#### NFR6.1: Logging
- **Requirement:** Structured logging (JSON format)
- **Requirement:** Correlation IDs for request tracking
- **Requirement:** Log all errors with stack traces

#### NFR6.2: Monitoring
- **Requirement:** Health check endpoint
- **Requirement:** Metrics endpoint (Prometheus format)
- **Requirement:** Status reporting to API

---

## 🏛️ Technical Architecture

### Architecture Pattern: Clean Architecture

Ghost Agent follows **Clean Architecture** (Hexagonal/Ports & Adapters) principles for maintainability and extensibility.

```
┌─────────────────────────────────────────────────────────┐
│                    Ghost Agent Architecture                    │
├─────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────┐  │
│  │  Layer 1: Presentation (Adapters/Interfaces)      │  │
│  ├─────────────────────────────────────────────────┤  │
│  │                                                  │  │
│  │  ┌────────────────┐  ┌────────────────┐  │  │
│  │  │  gRPC Server   │  │  HTTP Server  │  │  │
│  │  │  (Port 9090)   │  │  (Metrics/    │  │  │
│  │  │  - VM Ops      │  │   Health)     │  │  │
│  │  │  - mTLS Auth   │  │  - Prometheus │  │  │
│  │  └────────────────┘  └────────────────┘  │  │
│  │                                                  │  │
│  │  ┌────────────────┐  ┌────────────────┐  │  │
│  │  │  API Client    │  │  CLI Handler  │  │  │
│  │  │  (Ghost API)   │  │  (agentctl)   │  │  │
│  │  │  - Heartbeat   │  │  - Local mgmt │  │  │
│  │  │  - Register    │  │              │  │  │
│  │  └────────────────┘  └────────────────┘  │  │
│  └─────────────────────────────────────────────────┘  │
│                         │                                    │
│                         │ (Dependency Injection)             │
│                         │                                    │
│  ┌─────────────────────────────────────────────────┐  │
│  │  Layer 2: Application (Use Cases)                 │  │
│  ├─────────────────────────────────────────────────┤  │
│  │                                                  │  │
│  │  ┌──────────────────────────────────────────┐  │  │
│  │  │  VM Use Cases                           │  │  │
│  │  │  - CreateVMUseCase                      │  │  │
│  │  │  - DeleteVMUseCase                      │  │  │
│  │  │  - StartVMUseCase                       │  │  │
│  │  │  - StopVMUseCase                        │  │  │
│  │  │  - GetVMStatusUseCase                   │  │  │
│  │  │                                          │  │  │
│  │  │  Business Logic:                        │  │  │
│  │  │  - Input validation                     │  │  │
│  │  │  - Resource allocation checks           │  │  │
│  │  │  - Error handling & retries             │  │  │
│  │  │  - Event emission                       │  │  │
│  │  └──────────────────────────────────────────┘  │  │
│  └─────────────────────────────────────────────────┘  │
│                         │                                    │
│                         │ (Interfaces/Ports)                 │
│                         │                                    │
│  ┌─────────────────────────────────────────────────┐  │
│  │  Layer 3: Domain (Business Logic)                 │  │
│  ├─────────────────────────────────────────────────┤  │
│  │                                                  │  │
│  │  Entities:                                        │  │
│  │  - VM (id, name, vcpu, ram, disk, status)         │  │
│  │  - Resource (cpu, ram, disk, reserved)            │  │
│  │  - Image (name, path, checksum, size)             │  │
│  │                                                  │  │
│  │  Repository Interfaces (Ports):                   │  │
│  │  - VMRepository                                   │  │
│  │  - ResourceRepository                             │  │
│  │  - ImageRepository                                │  │
│  │                                                  │  │
│  │  Service Interfaces (Ports):                      │  │
│  │  - HypervisorService                              │  │
│  │  - NetworkService (extensibility point)           │  │
│  │  - StorageService                                 │  │
│  │  - MetricsService                                 │  │
│  │                                                  │  │
│  └─────────────────────────────────────────────────┘  │
│                         │                                    │
│                         │ (Implementations)                  │
│                         │                                    │
│  ┌─────────────────────────────────────────────────┐  │
│  │  Layer 4: Infrastructure (Adapters)               │  │
│  ├─────────────────────────────────────────────────┤  │
│  │                                                  │  │
│  │  ┌──────────────────────────────────────────┐  │  │
│  │  │  Libvirt Adapter                        │  │  │
│  │  │  - Implements HypervisorService         │  │  │
│  │  │  - Connection pool management           │  │  │
│  │  │  - Circuit breaker for resilience       │  │  │
│  │  └──────────────────────────────────────────┘  │  │
│  │                                                  │  │
│  │  ┌──────────────────────────────────────────┐  │  │
│  │  │  NAT Network Adapter (v1.0)             │  │  │
│  │  │  - Implements NetworkService            │  │  │
│  │  │  - Simple NAT (192.168.122.0/24)        │  │  │
│  │  │  - Pluggable design for future upgrade  │  │  │
│  │  └──────────────────────────────────────────┘  │  │
│  │                                                  │  │
│  │  ┌──────────────────────────────────────────┐  │  │
│  │  │  Storage Adapter                        │  │  │
│  │  │  - Implements StorageService            │  │  │
│  │  │  - Image caching & management           │  │  │
│  │  │  - Disk provisioning (qcow2)            │  │  │
│  │  └──────────────────────────────────────────┘  │  │
│  │                                                  │  │
│  │  ┌──────────────────────────────────────────┐  │  │
│  │  │  Observability Adapters                 │  │  │
│  │  │  - Zap Logger (structured JSON)         │  │  │
│  │  │  - Prometheus Metrics                   │  │  │
│  │  │  - OpenTelemetry Tracing (optional)     │  │  │
│  │  └──────────────────────────────────────────┘  │  │
│  └─────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Key Architectural Decisions

#### 1. **Dependency Inversion**
- Domain layer defines interfaces (ports)
- Infrastructure layer implements interfaces (adapters)
- Application layer orchestrates via interfaces
- Easy to swap implementations (e.g., replace Libvirt with another hypervisor)

#### 2. **Separation of Concerns**
- **Presentation:** Protocol-specific logic (gRPC, HTTP)
- **Application:** Use case orchestration, no business rules
- **Domain:** Pure business logic, framework-agnostic
- **Infrastructure:** External dependencies, I/O operations

#### 3. **Extensibility Points**

**NetworkService Interface** (Future-proof for advanced networking):
```go
type NetworkService interface {
    // v1.0: Simple NAT implementation
    AssignIP(vmID string) (string, error)
    ReleaseIP(vmID string) error
    
    // Future: Can add methods for advanced networking
    // CreateNetwork(cidr string) error
    // ConnectVMToNetwork(vmID, networkID string) error
    // SetupVMRouting(vmID string, routes []Route) error
}
```

**Multiple implementations possible**:
- `NATNetworkAdapter` (v1.0 - simple NAT)
- `TailscaleSubnetAdapter` (v2.0 - subnet router)
- `VXLANNetworkAdapter` (v3.0 - overlay network)
- `WireGuardMeshAdapter` (v3.0 - mesh networking)

Swap via configuration:
```yaml
network:
  provider: "nat"  # or "tailscale-subnet", "vxlan", etc.
```

#### 4. **Resilience Patterns**

**Circuit Breaker** (for Libvirt connection):
```go
type LibvirtAdapter struct {
    conn          *libvirt.Connect
    circuitBreaker *CircuitBreaker
}

func (l *LibvirtAdapter) CreateVM(vm *domain.VM) error {
    return l.circuitBreaker.Execute(func() error {
        // Libvirt operations
        return l.conn.DomainDefineXML(xml)
    })
}
```

**Retry with Exponential Backoff**:
```go
func (a *APIClient) SendHeartbeat(ctx context.Context) error {
    return retry.Do(
        func() error {
            return a.client.Heartbeat(ctx, req)
        },
        retry.Attempts(3),
        retry.Delay(1*time.Second),
        retry.MaxDelay(10*time.Second),
        retry.DelayType(retry.BackOffDelay),
    )
}
```

---

## 📦 Dependencies

### System Dependencies

#### Required Packages (Linux)

**Ubuntu/Debian:**
```bash
apt-get install -y \
    qemu-kvm \
    libvirt-daemon-system \
    libvirt-clients \
    libvirt-dev \
    bridge-utils \
    cpu-checker \
    cloud-image-utils
```

**Fedora/RHEL:**
```bash
dnf install -y \
    qemu-kvm \
    libvirt \
    libvirt-devel \
    virt-install \
    bridge-utils
```

---

### Go Dependencies

**go.mod** (Latest Stable Versions - December 2024):
```go
module github.com/yourusername/ghost-cloud

go 1.22  // Latest stable Go version

require (
    // === Core Hypervisor ===
    // Libvirt Go bindings
    libvirt.org/go/libvirt v1.10002.0
    
    // === Communication ===
    // gRPC and Protocol Buffers
    google.golang.org/grpc v1.60.1
    google.golang.org/protobuf v1.32.0
    
    // === Configuration Management ===
    github.com/spf13/viper v1.18.2        // Config loading (YAML, ENV, etc.)
    github.com/go-playground/validator/v10 v10.16.0  // Struct validation
    
    // === Logging (Structured) ===
    go.uber.org/zap v1.26.0               // High-performance structured logging
    
    // === Metrics & Observability ===
    github.com/prometheus/client_golang v1.18.0  // Prometheus metrics
    go.opentelemetry.io/otel v1.21.0      // OpenTelemetry tracing (optional)
    go.opentelemetry.io/otel/trace v1.21.0
    go.opentelemetry.io/otel/sdk v1.21.0
    
    // === Resilience Patterns ===
    github.com/sony/gobreaker v0.5.0      // Circuit breaker
    github.com/avast/retry-go/v4 v4.5.1   // Retry with exponential backoff
    golang.org/x/time v0.5.0              // Rate limiting
    
    // === Utilities ===
    github.com/google/uuid v1.5.0         // UUID generation
    golang.org/x/sync v0.6.0              // Advanced concurrency primitives
    github.com/hashicorp/go-multierror v1.1.1  // Multiple error handling
    
    // === Storage & Disk ===
    github.com/diskfs/go-diskfs v1.4.0    // Disk image manipulation
    
    // === Testing ===
    github.com/stretchr/testify v1.8.4    // Testing assertions
    github.com/golang/mock v1.6.0         // Mocking framework
    github.com/testcontainers/testcontainers-go v0.27.0  // Integration testing
    
    // === Security ===
    golang.org/x/crypto v0.18.0           // Cryptographic functions
    
    // === CLI (for agentctl) ===
    github.com/spf13/cobra v1.8.0         // CLI framework
    github.com/spf13/pflag v1.0.5         // POSIX/GNU-style flags
)
```

**Key Library Choices:**

1. **Resilience:**
   - `gobreaker`: Circuit breaker pattern for Libvirt connections
   - `retry-go`: Exponential backoff for API calls
   - `golang.org/x/time/rate`: Rate limiting for resource-intensive operations

2. **Observability:**
   - `zap`: Zero-allocation JSON logger (production-grade)
   - `prometheus/client_golang`: Industry-standard metrics
   - `opentelemetry`: Distributed tracing (optional, for advanced debugging)

3. **Validation:**
   - `validator/v10`: Struct tag-based validation for all inputs

4. **Testing:**
   - `testify`: Assertions and mocking
   - `testcontainers`: Spin up real dependencies for integration tests
   - `golang/mock`: Generate mocks from interfaces

---

### External Services

1. **Libvirt Daemon (libvirtd)**
   - Purpose: VM management
   - Port: Unix socket `/var/run/libvirt/libvirt-sock`
   - Required: Yes

2. **Tailscale/Headscale**
   - Purpose: VPN connectivity
   - Port: N/A (managed by Tailscale)
   - Required: Yes

3. **Ghost API**
   - Purpose: Control plane communication
   - Port: 8080 (configurable)
   - Protocol: gRPC over Headscale VPN
   - Required: Yes

---

## 🔨 Build Instructions

### Prerequisites

1. **Go 1.21 or later**
```bash
# Install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

2. **Libvirt development headers**
```bash
# Ubuntu/Debian
sudo apt-get install libvirt-dev

# Fedora/RHEL
sudo dnf install libvirt-devel
```

3. **Protocol Buffer Compiler**
```bash
# Install protoc
sudo apt-get install protobuf-compiler

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

---

### Build Steps

#### 1. Clone Repository
```bash
git clone https://github.com/you/ghost-cloud.git
cd ghost-cloud
```

#### 2. Generate Protocol Buffers
```bash
# Generate gRPC code from .proto files
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/agentpb/agent.proto
```

#### 3. Download Dependencies
```bash
go mod download
go mod tidy
```

#### 4. Build Agent
```bash
# Build for current platform
cd cmd/ghost-agent
go build -o ghost-agent

# Build with optimizations
go build -ldflags="-s -w" -o ghost-agent

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o ghost-agent-linux-amd64
GOOS=linux GOARCH=arm64 go build -o ghost-agent-linux-arm64
```

#### 5. Run Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### 6. Install
```bash
# Install binary
sudo install -m 755 ghost-agent /usr/local/bin/ghost-agent

# Verify installation
ghost-agent --version
```

---

### Build Automation (Makefile)

```makefile
# Makefile
.PHONY: all build test clean install

# Variables
BINARY_NAME=ghost-agent
VERSION=$(shell git describe --tags --always --dirty)
BUILD_DIR=build
LDFLAGS=-ldflags="-s -w -X main.Version=$(VERSION)"

all: clean build test

# Generate protobuf code
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/agentpb/agent.proto

# Build binary
build: proto
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/ghost-agent/main.go

# Build for multiple platforms
build-all: proto
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 cmd/ghost-agent/main.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 cmd/ghost-agent/main.go

# Run tests
test:
	go test -v -cover ./...

# Run tests with race detection
test-race:
	go test -v -race ./...

# Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Install binary
install: build
	sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	goimports -w .
```

**Usage:**
```bash
make build        # Build binary
make test         # Run tests
make install      # Install to /usr/local/bin
make build-all    # Build for all platforms
```

---

## 📡 API Specifications

### gRPC Protocol Definition

```protobuf
// pkg/agentpb/agent.proto
syntax = "proto3";

package agentpb;

option go_package = "github.com/you/ghost-cloud/pkg/agentpb";

// Agent Service - VM management operations
service AgentService {
  rpc CreateVM(CreateVMRequest) returns (CreateVMResponse);
  rpc DeleteVM(DeleteVMRequest) returns (DeleteVMResponse);
  rpc StartVM(StartVMRequest) returns (StartVMResponse);
  rpc StopVM(StopVMRequest) returns (StopVMResponse);
  rpc GetVMStatus(GetVMStatusRequest) returns (GetVMStatusResponse);
  rpc ListVMs(ListVMsRequest) returns (ListVMsResponse);
}

// CreateVM Request
message CreateVMRequest {
  string name = 1;
  int32 vcpu = 2;
  int32 ram_gb = 3;
  int32 disk_gb = 4;
  string template = 5;  // e.g., "ubuntu-22.04"
  map<string, string> metadata = 6;  // Optional metadata
}

// CreateVM Response
message CreateVMResponse {
  string vm_id = 1;
  string ip_address = 2;
  string status = 3;  // "running", "error"
  string error = 4;   // Error message if failed
}

// DeleteVM Request
message DeleteVMRequest {
  string vm_id = 1;
}

// DeleteVM Response
message DeleteVMResponse {
  bool success = 1;
  string error = 2;
}

// StartVM Request
message StartVMRequest {
  string vm_id = 1;
}

// StartVM Response
message StartVMResponse {
  string status = 1;
  string error = 2;
}

// StopVM Request
message StopVMRequest {
  string vm_id = 1;
  bool force = 2;  // Force shutdown
}

// StopVM Response
message StopVMResponse {
  string status = 1;
  string error = 2;
}

// GetVMStatus Request
message GetVMStatusRequest {
  string vm_id = 1;
}

// GetVMStatus Response
message GetVMStatusResponse {
  string vm_id = 1;
  string name = 2;
  string status = 3;
  int32 vcpu = 4;
  int32 ram_gb = 5;
  int32 disk_gb = 6;
  string ip_address = 7;
  int64 uptime_seconds = 8;
  float cpu_usage_percent = 9;
  float ram_usage_percent = 10;
}

// ListVMs Request
message ListVMsRequest {
  // Empty - list all VMs
}

// ListVMs Response
message ListVMsResponse {
  repeated VMInfo vms = 1;
}

message VMInfo {
  string vm_id = 1;
  string name = 2;
  string status = 3;
  string ip_address = 4;
}
```

---

## ⚙️ Configuration

### Configuration File Format

**Location:** `/etc/ghost/agent.yaml`

```yaml
# Ghost Agent Configuration

agent:
  # Unique name for this agent
  name: "johns-laptop"
  
  # Ghost API URL (Headscale IP)
  api_url: "https://100.64.0.1:8080"
  
  # Heartbeat interval
  heartbeat_interval: 30s
  
  # Agent version (auto-populated)
  version: "1.0.0"

# Libvirt configuration
libvirt:
  # Libvirt connection URI
  uri: "qemu:///system"
  
  # Storage pool for VM disks
  storage_pool: "default"
  
  # Network for VMs
  network: "default"
  
  # Image cache directory
  image_cache: "/var/lib/ghost/images"

# Resource configuration
resources:
  # CPU cores to reserve for PC owner
  reserved_cpu: 2
  
  # RAM to reserve for PC owner (GB)
  reserved_ram_gb: 4
  
  # Disk space to reserve for PC owner (GB)
  reserved_disk_gb: 50

# gRPC server configuration
grpc:
  # Listen address
  listen_addr: "0.0.0.0:9090"
  
  # Enable TLS
  tls_enabled: true
  
  # TLS certificate file
  tls_cert: "/etc/ghost/certs/agent.crt"
  
  # TLS key file
  tls_key: "/etc/ghost/certs/agent.key"
  
  # CA certificate file
  tls_ca: "/etc/ghost/certs/ca.crt"

# Logging configuration
logging:
  # Log level: debug, info, warn, error
  level: "info"
  
  # Log output: stdout, file, both
  output: "both"
  
  # Log file path
  file: "/var/log/ghost/agent.log"
  
  # Log format: json, text
  format: "json"
  
  # Max log file size (MB)
  max_size: 100
  
  # Max log file age (days)
  max_age: 7

# Metrics configuration (optional)
metrics:
  # Enable Prometheus metrics
  enabled: true
  
  # Metrics endpoint
  listen_addr: "0.0.0.0:9091"
  
  # Metrics path
  path: "/metrics"
```

---

## 🚀 Deployment

### Systemd Service File

**Location:** `/etc/systemd/system/ghost-agent.service`

```ini
[Unit]
Description=Ghost Cloud Agent
Documentation=https://github.com/you/ghost-cloud
After=network.target libvirtd.service
Requires=libvirtd.service

[Service]
Type=simple
User=root
Group=root

# Main process
ExecStart=/usr/local/bin/ghost-agent --config /etc/ghost/agent.yaml

# Restart policy
Restart=always
RestartSec=10

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=ghost-agent

# Security
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

### Deployment Steps

```bash
# 1. Install binary
sudo install -m 755 ghost-agent /usr/local/bin/ghost-agent

# 2. Create directories
sudo mkdir -p /etc/ghost
sudo mkdir -p /var/log/ghost
sudo mkdir -p /var/lib/ghost/images

# 3. Copy configuration
sudo cp agent.yaml /etc/ghost/agent.yaml

# 4. Copy systemd service
sudo cp ghost-agent.service /etc/systemd/system/

# 5. Reload systemd
sudo systemctl daemon-reload

# 6. Enable service
sudo systemctl enable ghost-agent

# 7. Start service
sudo systemctl start ghost-agent

# 8. Check status
sudo systemctl status ghost-agent
```

---

## 🧪 Testing Requirements

### Unit Tests

**Coverage:** Minimum 80%

**Test Files:**
```
cmd/ghost-agent/
├── main_test.go
pkg/
├── vmmanager/
│   ├── manager_test.go
│   └── libvirt_test.go
├── resource/
│   └── monitor_test.go
└── image/
    └── cache_test.go
```

**Example Test:**
```go
// pkg/vmmanager/manager_test.go
package vmmanager

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCreateVM(t *testing.T) {
    // Setup
    manager := NewVMManager()
    
    // Test
    vm, err := manager.CreateVM(&CreateVMRequest{
        Name: "test-vm",
        VCPU: 2,
        RAMGB: 4,
        DiskGB: 20,
        Template: "ubuntu-22.04",
    })
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, vm)
    assert.Equal(t, "test-vm", vm.Name)
    assert.Equal(t, "running", vm.Status)
}
```

---

### Integration Tests

**Test Scenarios:**
1. Create VM end-to-end
2. Delete VM end-to-end
3. Start/Stop VM
4. Heartbeat communication
5. Resource monitoring
6. Image caching

---

### Performance Tests

**Benchmarks:**
```go
// pkg/vmmanager/manager_bench_test.go
func BenchmarkCreateVM(b *testing.B) {
    manager := NewVMManager()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        manager.CreateVM(&CreateVMRequest{
            Name: fmt.Sprintf("bench-vm-%d", i),
            VCPU: 2,
            RAMGB: 4,
        })
    }
}
```

---

## 📝 Additional Requirements

### Error Handling (Production-Grade)

#### 1. **Error Types & Classification**

Define custom error types for better handling:
```go
// domain/errors.go
package domain

type ErrorCode string

const (
    ErrCodeValidation    ErrorCode = "VALIDATION_ERROR"
    ErrCodeResourceLimit ErrorCode = "RESOURCE_LIMIT_EXCEEDED"
    ErrCodeHypervisor    ErrorCode = "HYPERVISOR_ERROR"
    ErrCodeNetwork       ErrorCode = "NETWORK_ERROR"
    ErrCodeNotFound      ErrorCode = "NOT_FOUND"
    ErrCodeConflict      ErrorCode = "CONFLICT"
    ErrCodeInternal      ErrorCode = "INTERNAL_ERROR"
)

type AppError struct {
    Code    ErrorCode
    Message string
    Err     error
    Context map[string]interface{}
}

func (e *AppError) Error() string {
    return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
}

func (e *AppError) Unwrap() error {
    return e.Err
}
```

#### 2. **Error Wrapping with Context**

Always wrap errors with context:
```go
func (uc *CreateVMUseCase) Execute(ctx context.Context, req *CreateVMRequest) (*VM, error) {
    // Validate input
    if err := uc.validator.Struct(req); err != nil {
        return nil, &AppError{
            Code:    ErrCodeValidation,
            Message: "invalid VM creation request",
            Err:     err,
            Context: map[string]interface{}{
                "vm_name": req.Name,
                "vcpu":    req.VCPU,
            },
        }
    }
    
    // Check resources
    available, err := uc.resourceRepo.GetAvailable(ctx)
    if err != nil {
        return nil, &AppError{
            Code:    ErrCodeInternal,
            Message: "failed to check available resources",
            Err:     err,
            Context: map[string]interface{}{
                "operation": "resource_check",
            },
        }
    }
    
    // Create VM
    vm, err := uc.hypervisor.CreateVM(ctx, req)
    if err != nil {
        return nil, &AppError{
            Code:    ErrCodeHypervisor,
            Message: "failed to create VM in hypervisor",
            Err:     err,
            Context: map[string]interface{}{
                "vm_name": req.Name,
                "template": req.Template,
            },
        }
    }
    
    return vm, nil
}
```

#### 3. **Structured Error Logging**

Log all errors with full context:
```go
func (s *GRPCServer) CreateVM(ctx context.Context, req *pb.CreateVMRequest) (*pb.CreateVMResponse, error) {
    vm, err := s.createVMUseCase.Execute(ctx, req)
    if err != nil {
        // Extract context from error
        var appErr *AppError
        if errors.As(err, &appErr) {
            s.logger.Error("VM creation failed",
                zap.String("error_code", string(appErr.Code)),
                zap.String("message", appErr.Message),
                zap.Error(appErr.Err),
                zap.Any("context", appErr.Context),
                zap.String("trace_id", getTraceID(ctx)),
                zap.Stack("stack"),
            )
            
            // Map to gRPC error
            return nil, toGRPCError(appErr)
        }
        
        // Unknown error
        s.logger.Error("Unexpected error",
            zap.Error(err),
            zap.Stack("stack"),
        )
        return nil, status.Error(codes.Internal, "internal server error")
    }
    
    return toProtoVM(vm), nil
}
```

#### 4. **Error Recovery & Retry**

Implement retry logic for transient failures:
```go
func (a *APIClient) SendHeartbeat(ctx context.Context, req *HeartbeatRequest) error {
    return retry.Do(
        func() error {
            return a.client.Heartbeat(ctx, req)
        },
        retry.Attempts(3),
        retry.Delay(1*time.Second),
        retry.MaxDelay(10*time.Second),
        retry.DelayType(retry.BackOffDelay),
        retry.OnRetry(func(n uint, err error) {
            a.logger.Warn("Heartbeat retry",
                zap.Uint("attempt", n),
                zap.Error(err),
            )
        }),
        retry.RetryIf(func(err error) bool {
            // Only retry on network errors
            return isNetworkError(err)
        }),
    )
}
```

---

### Security (Production-Grade)

#### 1. **Input Validation**

Validate ALL inputs using struct tags:
```go
type CreateVMRequest struct {
    Name     string `validate:"required,min=3,max=63,hostname"`
    VCPU     int    `validate:"required,min=1,max=32"`
    RAMGB    int    `validate:"required,min=1,max=128"`
    DiskGB   int    `validate:"required,min=10,max=1000"`
    Template string `validate:"required,oneof=ubuntu-22.04 ubuntu-20.04 debian-12 debian-11"`
}

// In use case
func (uc *CreateVMUseCase) Execute(ctx context.Context, req *CreateVMRequest) (*VM, error) {
    if err := uc.validator.Struct(req); err != nil {
        return nil, &AppError{
            Code:    ErrCodeValidation,
            Message: "validation failed",
            Err:     err,
        }
    }
    // ... proceed with creation
}
```

#### 2. **mTLS Configuration**

Enforce mutual TLS for all gRPC connections:
```go
func NewGRPCServer(cfg *Config) (*grpc.Server, error) {
    // Load server certificate
    cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
    if err != nil {
        return nil, fmt.Errorf("failed to load server cert: %w", err)
    }
    
    // Load CA certificate
    caCert, err := os.ReadFile(cfg.TLSCA)
    if err != nil {
        return nil, fmt.Errorf("failed to load CA cert: %w", err)
    }
    
    caCertPool := x509.NewCertPool()
    if !caCertPool.AppendCertsFromPEM(caCert) {
        return nil, errors.New("failed to add CA cert to pool")
    }
    
    // Create TLS config
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        ClientCAs:    caCertPool,
        ClientAuth:   tls.RequireAndVerifyClientCert,  // Enforce mTLS
        MinVersion:   tls.VersionTLS13,                // TLS 1.3 only
    }
    
    // Create gRPC server with TLS
    return grpc.NewServer(
        grpc.Creds(credentials.NewTLS(tlsConfig)),
        grpc.UnaryInterceptor(authInterceptor),
    ), nil
}
```

#### 3. **Secrets Management**

Never hardcode secrets, use environment variables or secret managers:
```go
// config/config.go
type Config struct {
    APIURL    string `mapstructure:"api_url"`
    APIKey    string `mapstructure:"api_key" validate:"required"`  // From env
    TLSCert   string `mapstructure:"tls_cert"`
    TLSKey    string `mapstructure:"tls_key"`
}

func LoadConfig() (*Config, error) {
    viper.SetConfigName("agent")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("/etc/ghost")
    
    // Override with environment variables
    viper.SetEnvPrefix("GHOST")
    viper.AutomaticEnv()
    
    // API key MUST come from environment
    viper.BindEnv("api_key", "GHOST_API_KEY")
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }
    
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    // Validate config
    validate := validator.New()
    if err := validate.Struct(&cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

#### 4. **Principle of Least Privilege**

Run with minimal permissions:
```ini
# systemd service
[Service]
# Run as dedicated user (not root if possible)
User=ghost-agent
Group=ghost-agent

# Restrict capabilities
CapabilityBoundingSet=CAP_NET_ADMIN CAP_SYS_ADMIN
AmbientCapabilities=CAP_NET_ADMIN CAP_SYS_ADMIN

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/ghost /var/log/ghost
```

#### 5. **Rate Limiting**

Protect against abuse:
```go
type RateLimitedServer struct {
    limiter *rate.Limiter
    server  pb.AgentServiceServer
}

func (s *RateLimitedServer) CreateVM(ctx context.Context, req *pb.CreateVMRequest) (*pb.CreateVMResponse, error) {
    if !s.limiter.Allow() {
        return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
    }
    return s.server.CreateVM(ctx, req)
}
```

---

### Graceful Shutdown (Production-Grade)

#### 1. **Signal Handling**

```go
func main() {
    // Create context that cancels on signal
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
    
    // Start agent
    agent := NewAgent(cfg)
    
    // Start in goroutine
    go func() {
        if err := agent.Start(ctx); err != nil {
            logger.Fatal("Agent failed", zap.Error(err))
        }
    }()
    
    // Wait for signal
    sig := <-sigChan
    logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
    
    // Graceful shutdown
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()
    
    if err := agent.Shutdown(shutdownCtx); err != nil {
        logger.Error("Shutdown error", zap.Error(err))
        os.Exit(1)
    }
    
    logger.Info("Agent shutdown complete")
}
```

#### 2. **Shutdown Procedure**

```go
func (a *Agent) Shutdown(ctx context.Context) error {
    a.logger.Info("Starting graceful shutdown")
    
    // 1. Stop accepting new requests
    a.grpcServer.GracefulStop()
    a.logger.Info("Stopped accepting new requests")
    
    // 2. Stop heartbeat
    a.heartbeatCancel()
    a.logger.Info("Stopped heartbeat")
    
    // 3. Wait for in-flight operations
    done := make(chan struct{})
    go func() {
        a.wg.Wait()  // Wait for all goroutines
        close(done)
    }()
    
    select {
    case <-done:
        a.logger.Info("All operations completed")
    case <-ctx.Done():
        a.logger.Warn("Shutdown timeout, forcing exit")
        return ctx.Err()
    }
    
    // 4. Save state
    if err := a.saveState(); err != nil {
        a.logger.Error("Failed to save state", zap.Error(err))
        return err
    }
    
    // 5. Close connections
    if err := a.libvirtConn.Close(); err != nil {
        a.logger.Error("Failed to close libvirt connection", zap.Error(err))
    }
    
    a.logger.Info("Graceful shutdown complete")
    return nil
}
```

#### 3. **State Persistence**

```go
type AgentState struct {
    VMs           []string          `json:"vms"`
    LastHeartbeat time.Time         `json:"last_heartbeat"`
    Metadata      map[string]string `json:"metadata"`
}

func (a *Agent) saveState() error {
    state := &AgentState{
        VMs:           a.vmManager.ListVMIDs(),
        LastHeartbeat: time.Now(),
        Metadata: map[string]string{
            "version": a.version,
        },
    }
    
    data, err := json.Marshal(state)
    if err != nil {
        return fmt.Errorf("failed to marshal state: %w", err)
    }
    
    if err := os.WriteFile("/var/lib/ghost/agent-state.json", data, 0600); err != nil {
        return fmt.Errorf("failed to write state: %w", err)
    }
    
    return nil
}

func (a *Agent) loadState() (*AgentState, error) {
    data, err := os.ReadFile("/var/lib/ghost/agent-state.json")
    if err != nil {
        if os.IsNotExist(err) {
            return &AgentState{}, nil  // First run
        }
        return nil, err
    }
    
    var state AgentState
    if err := json.Unmarshal(data, &state); err != nil {
        return nil, err
    }
    
    return &state, nil
}
```

---

### Health Checks (Production-Grade)

#### 1. **HTTP Health Endpoint**

```go
type HealthHandler struct {
    agent   *Agent
    version string
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    health := h.checkHealth()
    
    status := http.StatusOK
    if health.Status != "healthy" {
        status = http.StatusServiceUnavailable
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(health)
}

func (h *HealthHandler) checkHealth() *HealthResponse {
    resp := &HealthResponse{
        Status:    "healthy",
        Version:   h.version,
        Timestamp: time.Now(),
        Checks:    make(map[string]CheckResult),
    }
    
    // Check Libvirt connection
    if err := h.agent.libvirtConn.Ping(); err != nil {
        resp.Status = "unhealthy"
        resp.Checks["libvirt"] = CheckResult{
            Status:  "down",
            Message: err.Error(),
        }
    } else {
        resp.Checks["libvirt"] = CheckResult{Status: "up"}
    }
    
    // Check API connection
    if err := h.agent.apiClient.Ping(); err != nil {
        resp.Status = "degraded"
        resp.Checks["api"] = CheckResult{
            Status:  "down",
            Message: err.Error(),
        }
    } else {
        resp.Checks["api"] = CheckResult{Status: "up"}
    }
    
    // Add metrics
    resp.Metrics = map[string]interface{}{
        "uptime_seconds": time.Since(h.agent.startTime).Seconds(),
        "vms_running":    h.agent.vmManager.CountRunning(),
        "goroutines":     runtime.NumGoroutine(),
        "memory_mb":      getMemoryUsageMB(),
    }
    
    return resp
}

type HealthResponse struct {
    Status    string                    `json:"status"`
    Version   string                    `json:"version"`
    Timestamp time.Time                 `json:"timestamp"`
    Checks    map[string]CheckResult    `json:"checks"`
    Metrics   map[string]interface{}    `json:"metrics"`
}

type CheckResult struct {
    Status  string `json:"status"`
    Message string `json:"message,omitempty"`
}
```

#### 2. **Readiness vs Liveness**

```go
// Liveness: Is the agent running?
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

// Readiness: Can the agent handle requests?
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
    if !h.agent.IsReady() {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("NOT READY"))
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("READY"))
}
```

---

## 🎯 Success Criteria

The Ghost Agent is considered complete when:

1. ✅ All functional requirements are implemented
2. ✅ All non-functional requirements are met
3. ✅ Unit test coverage ≥ 80%
4. ✅ Integration tests pass
5. ✅ Performance benchmarks meet targets
6. ✅ Documentation is complete
7. ✅ Code passes linting (golangci-lint)
8. ✅ Successfully deploys on Ubuntu 22.04
9. ✅ Successfully creates and manages VMs
10. ✅ Successfully communicates with Ghost API

---

## 📚 References

- **Libvirt Go Bindings:** https://libvirt.org/go/libvirt.html
- **gRPC Go Tutorial:** https://grpc.io/docs/languages/go/
- **Protocol Buffers:** https://protobuf.dev/
- **KVM Documentation:** https://www.linux-kvm.org/
- **Systemd Service:** https://www.freedesktop.org/software/systemd/man/systemd.service.html

---

## 📞 Support

For questions or issues:
- GitHub Issues: https://github.com/you/ghost-cloud/issues
- Documentation: https://github.com/you/ghost-cloud/docs

---

**Document Version:** 1.0  
**Last Updated:** 2024-12-05  
**Status:** Final
