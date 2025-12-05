# Ghost Agent Architecture

## System Overview

Ghost Agent is a production-grade VM management daemon that runs on user PCs, managing virtual machines via KVM/Libvirt and communicating with Ghost Cloud Core.

```mermaid
graph TB
    subgraph "User PC"
        subgraph "Ghost Agent"
            CLI[ghostctl CLI]
            GRPC[gRPC Server :9090]
            HTTP[HTTP Server :9092]
            METRICS[Metrics :9091]
            
            subgraph "Application Layer"
                UC1[CreateVM UseCase]
                UC2[DeleteVM UseCase]
                UC3[StartVM UseCase]
                UC4[StopVM UseCase]
                UC5[GetStatus UseCase]
                UC6[ListVMs UseCase]
            end
            
            subgraph "Domain Layer"
                ENT[Entities: VM, Resource, Image]
                REPO[Repository Interfaces]
                SVC[Service Interfaces]
            end
            
            subgraph "Infrastructure Layer"
                LIB[Libvirt Adapter]
                NET[Network Adapter]
                STOR[Storage Adapter]
                API[Ghost Core Client]
                OBS[Observability]
                PERSIST[Persistent Storage]
            end
        end
        
        KVM[KVM/QEMU]
        TS[Tailscale]
        DISK[/var/lib/ghost/]
    end
    
    CLOUD[Ghost Cloud Core API]
    
    CLI --> GRPC
    GRPC --> UC1 & UC2 & UC3 & UC4 & UC5 & UC6
    UC1 & UC2 & UC3 & UC4 & UC5 & UC6 --> REPO & SVC
    REPO & SVC --> LIB & NET & STOR & PERSIST
    LIB --> KVM
    NET --> KVM
    STOR --> DISK
    PERSIST --> DISK
    API --> CLOUD
    API -.Heartbeat.-> CLOUD
    TS -.VPN.-> CLOUD
    
    style CLI fill:#e1f5ff
    style GRPC fill:#fff4e1
    style API fill:#ffe1e1
    style CLOUD fill:#e1ffe1
```

---

## Clean Architecture Layers

Ghost Agent follows **Clean Architecture** principles with 4 distinct layers:

```mermaid
graph LR
    subgraph "Presentation Layer"
        GRPC[gRPC Server]
        HTTP[HTTP Server]
        CLI[CLI Tool]
    end
    
    subgraph "Application Layer"
        UC[Use Cases]
        DTO[DTOs]
    end
    
    subgraph "Domain Layer"
        E[Entities]
        R[Repositories]
        S[Services]
        ERR[Errors]
    end
    
    subgraph "Infrastructure Layer"
        LIBVIRT[Libvirt Adapter]
        NETWORK[Network Adapter]
        STORAGE[Storage Adapter]
        APICLIENT[API Client]
        CONFIG[Config]
        METRICS[Metrics]
    end
    
    GRPC --> UC
    HTTP --> UC
    CLI --> GRPC
    UC --> E & R & S
    R --> LIBVIRT & NETWORK & STORAGE
    S --> LIBVIRT & NETWORK & STORAGE
    
    style E fill:#e1f5ff
    style UC fill:#fff4e1
    style LIBVIRT fill:#ffe1e1
    style GRPC fill:#e1ffe1
```

### Layer Responsibilities

**1. Domain Layer** (Core Business Logic)
- Entities: VM, Resource, Image
- Repository interfaces
- Service interfaces
- Domain errors
- **No dependencies on other layers**

**2. Application Layer** (Use Cases)
- CreateVM, DeleteVM, StartVM, StopVM, GetVMStatus, ListVMs
- DTOs for requests/responses
- Input validation
- Orchestrates domain + infrastructure

**3. Infrastructure Layer** (Adapters)
- Libvirt adapter (implements HypervisorService)
- Network adapter (implements NetworkService)
- Storage adapter (implements StorageService)
- Ghost Core API client
- Configuration, logging, metrics

**4. Presentation Layer** (Interfaces)
- gRPC server (receives commands)
- HTTP server (health, metrics)
- CLI tool (ghostctl)

---

## Data Flow

### VM Creation Flow

```mermaid
sequenceDiagram
    participant User
    participant ghostctl
    participant gRPC
    participant CreateVM
    participant Storage
    participant Libvirt
    participant Network
    participant APIClient
    participant GhostCore
    
    User->>ghostctl: ghostctl vm create --name my-vm
    ghostctl->>gRPC: CreateVMRequest
    gRPC->>CreateVM: Execute(request)
    
    CreateVM->>CreateVM: Validate input
    CreateVM->>CreateVM: Check resources
    
    CreateVM->>Storage: GetImage(ubuntu-22.04)
    Storage-->>CreateVM: /path/to/image.qcow2
    
    CreateVM->>Storage: CreateDisk(my-vm, base_image, 50GB)
    Storage-->>CreateVM: /var/lib/ghost/disks/my-vm.qcow2
    
    CreateVM->>Libvirt: CreateVM(spec)
    Libvirt->>Libvirt: Generate XML
    Libvirt->>Libvirt: Define domain
    Libvirt->>Libvirt: Start domain
    Libvirt-->>CreateVM: VM created
    
    CreateVM->>Network: GetVMIP(my-vm)
    Network-->>CreateVM: 192.168.122.10
    
    CreateVM->>CreateVM: Save to repository
    CreateVM->>CreateVM: Update resources
    
    CreateVM->>APIClient: ReportVMCreated(vm)
    APIClient->>GhostCore: ReportVMCreatedRequest
    GhostCore-->>APIClient: Success
    
    CreateVM-->>gRPC: CreateVMResponse
    gRPC-->>ghostctl: Response
    ghostctl-->>User: ✅ VM created!
```

### Heartbeat Flow

```mermaid
sequenceDiagram
    participant Agent
    participant APIClient
    participant VMRepo
    participant ResourceRepo
    participant GhostCore
    
    loop Every 30 seconds
        Agent->>APIClient: StartHeartbeat()
        APIClient->>ResourceRepo: GetAvailable()
        ResourceRepo-->>APIClient: Resources
        APIClient->>VMRepo: FindAll()
        VMRepo-->>APIClient: VMs list
        
        APIClient->>GhostCore: HeartbeatRequest
        Note over APIClient,GhostCore: Includes: agent_id, resources, VMs
        
        alt Success
            GhostCore-->>APIClient: HeartbeatResponse
            APIClient->>APIClient: Set heartbeat_success=1
        else Failure
            GhostCore-->>APIClient: Error
            APIClient->>APIClient: Retry (3 attempts)
            APIClient->>APIClient: Set heartbeat_success=0
        end
    end
```

---

## Deployment Architecture

### Single PC Deployment

```mermaid
graph TB
    subgraph "User PC"
        subgraph "systemd"
            SERVICE[ghost-agent.service]
        end
        
        subgraph "Ghost Agent Process"
            MAIN[main.go]
            GRPC[gRPC :9090]
            HTTP[Health :9092]
            METRICS[Metrics :9091]
        end
        
        subgraph "System Services"
            LIBVIRTD[libvirtd]
            TAILSCALED[tailscaled]
        end
        
        subgraph "Storage"
            CONFIG[/etc/ghost/agent.yaml]
            DATA[/var/lib/ghost/data/vms.json]
            IMAGES[/var/lib/ghost/images/]
            LOGS[/var/log/ghost/agent.log]
        end
        
        KVM[KVM/QEMU VMs]
    end
    
    INTERNET[Internet]
    GHOSTCORE[Ghost Cloud Core<br/>100.64.0.1:8080]
    
    SERVICE -->|starts| MAIN
    MAIN --> GRPC & HTTP & METRICS
    MAIN --> LIBVIRTD & TAILSCALED
    MAIN --> CONFIG & DATA & IMAGES & LOGS
    LIBVIRTD --> KVM
    TAILSCALED --> INTERNET
    MAIN -.Heartbeat.-> GHOSTCORE
    TAILSCALED -.VPN.-> GHOSTCORE
    
    style SERVICE fill:#e1f5ff
    style MAIN fill:#fff4e1
    style GHOSTCORE fill:#e1ffe1
```

### Multi-Agent Deployment

```mermaid
graph TB
    subgraph "Ghost Cloud Core"
        CORE[Ghost Core API<br/>100.64.0.1:8080]
        DB[(Database)]
    end
    
    subgraph "Tailscale VPN Network"
        TS[Tailscale Mesh]
    end
    
    subgraph "User PC 1 - 100.64.0.5"
        AGENT1[Ghost Agent]
        VM1[VMs: vm-1, vm-2]
    end
    
    subgraph "User PC 2 - 100.64.0.6"
        AGENT2[Ghost Agent]
        VM2[VMs: vm-3, vm-4, vm-5]
    end
    
    subgraph "User PC 3 - 100.64.0.7"
        AGENT3[Ghost Agent]
        VM3[VMs: vm-6]
    end
    
    AGENT1 & AGENT2 & AGENT3 --> TS
    TS --> CORE
    AGENT1 -.Heartbeat.-> CORE
    AGENT2 -.Heartbeat.-> CORE
    AGENT3 -.Heartbeat.-> CORE
    CORE --> DB
    
    AGENT1 --> VM1
    AGENT2 --> VM2
    AGENT3 --> VM3
    
    style CORE fill:#e1ffe1
    style TS fill:#ffe1e1
```

---

## State Persistence

### PC Restart Flow

```mermaid
sequenceDiagram
    participant PC
    participant systemd
    participant Agent
    participant Disk
    participant GhostCore
    
    Note over PC: PC Running
    PC->>Agent: VMs: vm-1, vm-2, vm-3
    Agent->>Disk: Save state to vms.json
    
    Note over PC: User shuts down PC
    PC->>systemd: SIGTERM
    systemd->>Agent: Graceful shutdown
    Agent->>GhostCore: UnregisterAgent()
    Agent->>Disk: Final state save
    Agent->>Agent: Exit
    
    Note over PC: PC Restarted
    PC->>systemd: Boot complete
    systemd->>Agent: Start ghost-agent
    Agent->>Disk: Load vms.json
    Disk-->>Agent: VMs: vm-1, vm-2, vm-3
    
    Agent->>GhostCore: RegisterAgent()
    GhostCore-->>Agent: agent_id
    
    Agent->>GhostCore: Heartbeat(VMs: vm-1, vm-2, vm-3)
    GhostCore-->>Agent: Success
    
    Note over GhostCore: Ghost Core knows<br/>all VMs survived restart
```

---

## Component Diagram

```mermaid
graph TB
    subgraph "Ghost Agent Components"
        subgraph "Servers"
            GRPC[gRPC Server<br/>Port 9090]
            HTTP[HTTP Server<br/>Port 9092]
            PROM[Prometheus<br/>Port 9091]
        end
        
        subgraph "Core Logic"
            UC[Use Cases]
            DOM[Domain Entities]
        end
        
        subgraph "Adapters"
            LIB[Libvirt Adapter<br/>+ Circuit Breaker]
            NET[Network Adapter<br/>NAT/DHCP]
            STOR[Storage Adapter<br/>Image Cache]
            API[API Client<br/>+ Retry Logic]
        end
        
        subgraph "Infrastructure"
            LOG[Logger<br/>Zap/JSON]
            MET[Metrics<br/>Prometheus]
            CFG[Config<br/>Viper]
            PERSIST[Persistent Repo<br/>JSON File]
        end
    end
    
    EXT1[Libvirt/KVM]
    EXT2[Ghost Core API]
    EXT3[File System]
    
    GRPC --> UC
    HTTP --> DOM
    UC --> DOM
    DOM --> LIB & NET & STOR & PERSIST
    LIB --> EXT1
    NET --> EXT1
    STOR --> EXT3
    PERSIST --> EXT3
    API --> EXT2
    
    LOG & MET & CFG -.supports.-> UC & LIB & NET & STOR & API
```

---

## Technology Stack

```mermaid
mindmap
  root((Ghost Agent))
    Language
      Go 1.25.5
    Architecture
      Clean Architecture
      4 Layers
      Interface-based
    Infrastructure
      Libvirt/KVM
      Tailscale VPN
      systemd
    Protocols
      gRPC
      HTTP
      Protobuf
    Observability
      Zap Logger
      Prometheus
      Health Checks
    Resilience
      Circuit Breaker
      Retry Logic
      State Persistence
    Storage
      JSON Files
      qcow2 Disks
      Image Cache
```

---

## Security Architecture

```mermaid
graph TB
    subgraph "Security Layers"
        subgraph "Network Security"
            FW[Firewall Rules]
            TS[Tailscale VPN]
            MTLS[mTLS Planned]
        end
        
        subgraph "Application Security"
            VAL[Input Validation]
            ERR[Error Context<br/>No Sensitive Data]
            PRIV[Least Privilege]
        end
        
        subgraph "System Security"
            SYSTEMD[systemd Hardening]
            PERMS[File Permissions]
            ROOT[Root Required]
        end
    end
    
    USER[User/ghostctl] --> FW
    FW --> GRPC[gRPC Server]
    GRPC --> VAL
    VAL --> APP[Application]
    
    AGENT[Ghost Agent] --> TS
    TS --> MTLS
    MTLS --> CORE[Ghost Core]
    
    APP --> SYSTEMD
    SYSTEMD --> PERMS
    
    style FW fill:#ffe1e1
    style VAL fill:#fff4e1
    style TS fill:#e1ffe1
```

---

## Performance Characteristics

- **VM Creation Time:** 30-60 seconds (depends on image download)
- **VM Start Time:** 5-10 seconds
- **VM Stop Time:** 5-15 seconds (graceful) / 1-2 seconds (force)
- **Heartbeat Interval:** 30 seconds
- **API Response Time:** < 100ms (except VM creation)
- **Memory Usage:** ~50MB (idle) / ~100MB (active)
- **CPU Usage:** < 1% (idle) / 5-10% (during VM operations)

---

## Scalability

- **VMs per Agent:** Limited by PC resources (typically 5-20 VMs)
- **Agents per Core:** Unlimited (tested up to 100)
- **Concurrent Operations:** Thread-safe, supports parallel VM operations
- **State Size:** ~1KB per VM (JSON storage)

---

## Future Architecture Enhancements

1. **Advanced Networking**
   - Replace NAT with Tailscale subnet router
   - VXLAN for VM-to-VM communication
   - WireGuard mesh networking

2. **High Availability**
   - VM migration between agents
   - Automatic failover
   - Distributed state

3. **Enhanced Security**
   - mTLS for all communications
   - API key authentication
   - Secrets management (Vault)

4. **Monitoring**
   - Grafana dashboards
   - Alert manager integration
   - Distributed tracing (OpenTelemetry)
