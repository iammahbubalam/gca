# Ghost Agent API Documentation

## Overview

Ghost Agent exposes two APIs:
1. **Agent gRPC API** - For managing VMs (used by ghostctl and Ghost Core)
2. **Ghost Core API Client** - For communicating with Ghost Cloud Core

---

## 1. Agent gRPC API

**Address:** `localhost:9090` (configurable)  
**Protocol:** gRPC  
**Authentication:** None (mTLS planned for production)

### Service Definition

```protobuf
service AgentService {
  rpc CreateVM(CreateVMRequest) returns (CreateVMResponse);
  rpc DeleteVM(DeleteVMRequest) returns (DeleteVMResponse);
  rpc StartVM(StartVMRequest) returns (StartVMResponse);
  rpc StopVM(StopVMRequest) returns (StopVMResponse);
  rpc GetVMStatus(GetVMStatusRequest) returns (GetVMStatusResponse);
  rpc ListVMs(ListVMsRequest) returns (ListVMsResponse);
}
```

### Methods

#### CreateVM

Creates a new virtual machine.

**Request:**
```json
{
  "name": "my-vm",
  "vcpu": 2,
  "ram_gb": 4,
  "disk_gb": 50,
  "template": "ubuntu-22.04",
  "metadata": {
    "key": "value"
  }
}
```

**Response:**
```json
{
  "vm_id": "vm-abc123",
  "ip_address": "192.168.122.10",
  "status": "running"
}
```

**Errors:**
- `INVALID_ARGUMENT` - Invalid parameters
- `ALREADY_EXISTS` - VM with same name exists
- `RESOURCE_EXHAUSTED` - Insufficient resources
- `INTERNAL` - Hypervisor error

**Example (ghostctl):**
```bash
ghostctl vm create --name my-vm --vcpu 2 --ram 4 --disk 50 --template ubuntu-22.04
```

---

#### DeleteVM

Deletes a virtual machine and its disk.

**Request:**
```json
{
  "vm_id": "vm-abc123"
}
```

**Response:**
```json
{
  "success": true
}
```

**Errors:**
- `NOT_FOUND` - VM doesn't exist
- `INTERNAL` - Failed to delete

**Example:**
```bash
ghostctl vm delete vm-abc123
```

---

#### StartVM

Starts a stopped virtual machine.

**Request:**
```json
{
  "vm_id": "vm-abc123"
}
```

**Response:**
```json
{
  "status": "running"
}
```

**Errors:**
- `NOT_FOUND` - VM doesn't exist
- `FAILED_PRECONDITION` - VM already running
- `INTERNAL` - Failed to start

**Example:**
```bash
ghostctl vm start vm-abc123
```

---

#### StopVM

Stops a running virtual machine.

**Request:**
```json
{
  "vm_id": "vm-abc123",
  "force": false
}
```

**Parameters:**
- `force`: If true, power off immediately. If false, graceful shutdown.

**Response:**
```json
{
  "status": "stopped"
}
```

**Errors:**
- `NOT_FOUND` - VM doesn't exist
- `FAILED_PRECONDITION` - VM already stopped
- `INTERNAL` - Failed to stop

**Example:**
```bash
ghostctl vm stop vm-abc123
ghostctl vm stop vm-abc123 --force  # Force stop
```

---

#### GetVMStatus

Gets detailed status of a virtual machine.

**Request:**
```json
{
  "vm_id": "vm-abc123"
}
```

**Response:**
```json
{
  "vm_id": "vm-abc123",
  "name": "my-vm",
  "status": "running",
  "vcpu": 2,
  "ram_gb": 4,
  "disk_gb": 50,
  "ip_address": "192.168.122.10",
  "uptime_seconds": 3600,
  "cpu_usage_percent": 25.5,
  "ram_usage_percent": 60.2
}
```

**Errors:**
- `NOT_FOUND` - VM doesn't exist
- `INTERNAL` - Failed to get status

**Example:**
```bash
ghostctl vm status vm-abc123
```

---

#### ListVMs

Lists all virtual machines on this agent.

**Request:**
```json
{}
```

**Response:**
```json
{
  "vms": [
    {
      "vm_id": "vm-abc123",
      "name": "my-vm",
      "status": "running",
      "ip_address": "192.168.122.10"
    },
    {
      "vm_id": "vm-def456",
      "name": "test-vm",
      "status": "stopped",
      "ip_address": "192.168.122.11"
    }
  ]
}
```

**Example:**
```bash
ghostctl vm list
```

---

## 2. Ghost Core API (Client)

**Address:** Configured in `agent.yaml` (e.g., `100.64.0.1:8080`)  
**Protocol:** gRPC  
**Authentication:** mTLS (planned)

### Service Definition

```protobuf
service GhostCoreService {
  rpc RegisterAgent(RegisterAgentRequest) returns (RegisterAgentResponse);
  rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);
  rpc ReportVMCreated(ReportVMCreatedRequest) returns (ReportVMCreatedResponse);
  rpc ReportVMDeleted(ReportVMDeletedRequest) returns (ReportVMDeletedResponse);
  rpc ReportVMStatusChange(ReportVMStatusChangeRequest) returns (ReportVMStatusChangeResponse);
  rpc UnregisterAgent(UnregisterAgentRequest) returns (UnregisterAgentResponse);
}
```

### Methods

#### RegisterAgent

Registers the agent with Ghost Core on startup.

**Request:**
```json
{
  "agent_name": "my-pc",
  "tailscale_ip": "100.64.0.5",
  "version": "1.0.0",
  "resources": {
    "total_cpu": 8,
    "available_cpu": 6,
    "total_ram_gb": 32,
    "available_ram_gb": 28,
    "total_disk_gb": 500,
    "available_disk_gb": 450
  }
}
```

**Response:**
```json
{
  "agent_id": "agent-abc123",
  "success": true,
  "message": "Agent registered successfully"
}
```

**When:** Called once on agent startup

---

#### Heartbeat

Sends periodic status updates to Ghost Core.

**Request:**
```json
{
  "agent_id": "agent-abc123",
  "timestamp": 1701234567,
  "resources": {
    "total_cpu": 8,
    "available_cpu": 4,
    "total_ram_gb": 32,
    "available_ram_gb": 20,
    "total_disk_gb": 500,
    "available_disk_gb": 400
  },
  "vms": [
    {
      "vm_id": "vm-abc123",
      "vm_name": "my-vm",
      "status": "running",
      "ip_address": "192.168.122.10",
      "vcpu": 2,
      "ram_gb": 4
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Heartbeat received",
  "commands": []
}
```

**When:** Every 30 seconds (configurable)  
**Retry:** 3 attempts with exponential backoff

---

#### ReportVMCreated

Reports a newly created VM to Ghost Core.

**Request:**
```json
{
  "agent_id": "agent-abc123",
  "vm_id": "vm-abc123",
  "vm_name": "my-vm",
  "vcpu": 2,
  "ram_gb": 4,
  "disk_gb": 50,
  "ip_address": "192.168.122.10",
  "template": "ubuntu-22.04"
}
```

**Response:**
```json
{
  "success": true,
  "message": "VM registered"
}
```

**When:** Immediately after VM creation

---

#### ReportVMDeleted

Reports a deleted VM to Ghost Core.

**Request:**
```json
{
  "agent_id": "agent-abc123",
  "vm_id": "vm-abc123"
}
```

**Response:**
```json
{
  "success": true
}
```

**When:** Immediately after VM deletion

---

#### ReportVMStatusChange

Reports VM status changes to Ghost Core.

**Request:**
```json
{
  "agent_id": "agent-abc123",
  "vm_id": "vm-abc123",
  "status": "stopped"
}
```

**Response:**
```json
{
  "success": true
}
```

**When:** When VM status changes (start/stop)

---

#### UnregisterAgent

Unregisters the agent from Ghost Core.

**Request:**
```json
{
  "agent_id": "agent-abc123"
}
```

**Response:**
```json
{
  "success": true
}
```

**When:** On agent shutdown (graceful)

---

## 3. HTTP Endpoints

### Health Check

**Endpoint:** `GET /health`  
**Port:** 9092 (configurable)

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2024-12-05T06:00:00Z",
  "checks": {
    "libvirt": {
      "status": "up"
    }
  },
  "metrics": {
    "uptime_seconds": 3600,
    "goroutines": 25
  }
}
```

**Status Codes:**
- `200 OK` - Agent is healthy
- `503 Service Unavailable` - Agent is unhealthy

---

### Readiness Check

**Endpoint:** `GET /ready`  
**Port:** 9092

**Response:** `READY` or `NOT READY`

**Status Codes:**
- `200 OK` - Agent is ready to accept requests
- `503 Service Unavailable` - Agent is not ready

---

### Liveness Check

**Endpoint:** `GET /live`  
**Port:** 9092

**Response:** `OK`

**Status Codes:**
- `200 OK` - Agent process is alive

---

### Prometheus Metrics

**Endpoint:** `GET /metrics`  
**Port:** 9091 (configurable)

**Metrics:**
```
# VM Operations
ghost_agent_vms_created_total
ghost_agent_vms_deleted_total
ghost_agent_vms_running
ghost_agent_vm_operations_total{operation="create",status="success"}

# API Calls
ghost_agent_api_call_duration_seconds{endpoint="heartbeat"}

# Heartbeat
ghost_agent_heartbeat_success
```

---

## 4. Supported OS Templates

- `ubuntu-22.04` - Ubuntu 22.04 LTS
- `ubuntu-20.04` - Ubuntu 20.04 LTS
- `debian-12` - Debian 12 (Bookworm)
- `debian-11` - Debian 11 (Bullseye)

---

## 5. Error Codes

| gRPC Code | Description | Example |
|-----------|-------------|---------|
| `OK` | Success | Request completed |
| `INVALID_ARGUMENT` | Invalid parameters | Missing required field |
| `NOT_FOUND` | Resource not found | VM doesn't exist |
| `ALREADY_EXISTS` | Resource exists | VM name conflict |
| `RESOURCE_EXHAUSTED` | Insufficient resources | Not enough RAM |
| `FAILED_PRECONDITION` | Invalid state | VM already running |
| `INTERNAL` | Internal error | Libvirt failure |

---

## 6. Rate Limits

Currently no rate limits. Will be added in future versions.

---

## 7. Examples

### Using gRPC (Go)

```go
import "github.com/iammahbubalam/ghost-agent/pkg/agentpb"

conn, _ := grpc.Dial("localhost:9090", grpc.WithInsecure())
client := agentpb.NewAgentServiceClient(conn)

// Create VM
resp, err := client.CreateVM(ctx, &agentpb.CreateVMRequest{
    Name: "my-vm",
    Vcpu: 2,
    RamGb: 4,
    DiskGb: 50,
    Template: "ubuntu-22.04",
})
```

### Using ghostctl

```bash
# All operations via CLI
ghostctl vm create --name my-vm --vcpu 2 --ram 4 --disk 50
ghostctl vm list
ghostctl vm status my-vm
ghostctl vm stop my-vm
ghostctl vm delete my-vm
```

---

## 8. Security

**Current:**
- No authentication (localhost only)
- No encryption

**Planned:**
- mTLS for Ghost Core communication
- API keys for Agent API
- Network policies (firewall rules)

---

## 9. Versioning

**Current Version:** 1.0.0  
**API Version:** v1  
**Compatibility:** Backward compatible within major version
