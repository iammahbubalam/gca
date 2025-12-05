package service

import (
	"context"
	
	"github.com/iammahbubalam/ghost-agent/internal/domain/entity"
)

// VMSpec defines the specification for creating a VM
type VMSpec struct {
	Name     string
	VCPU     int
	RAMGB    int
	DiskGB   int
	Template string
	DiskPath string
	IP       string
}

// VMStatusInfo contains detailed VM status information
type VMStatusInfo struct {
	Status           entity.VMStatus
	UptimeSeconds    int64
	CPUUsagePercent  float32
	RAMUsagePercent  float32
}

// HypervisorService defines the interface for hypervisor operations
type HypervisorService interface {
	// CreateVM creates a new virtual machine
	CreateVM(ctx context.Context, spec *VMSpec) (*entity.VM, error)
	
	// DeleteVM permanently deletes a virtual machine
	DeleteVM(ctx context.Context, id string) error
	
	// StartVM starts a stopped virtual machine
	StartVM(ctx context.Context, id string) error
	
	// StopVM stops a running virtual machine
	StopVM(ctx context.Context, id string, force bool) error
	
	// GetVMStatus retrieves detailed status of a VM
	GetVMStatus(ctx context.Context, id string) (*VMStatusInfo, error)
	
	// ListVMs lists all VMs managed by the hypervisor
	ListVMs(ctx context.Context) ([]*entity.VM, error)
	
	// Ping checks if hypervisor connection is alive
	Ping(ctx context.Context) error
}
