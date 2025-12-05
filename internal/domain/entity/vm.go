package entity

import "time"

// VMStatus represents the current state of a VM
type VMStatus string

const (
	VMStatusRunning VMStatus = "running"
	VMStatusStopped VMStatus = "stopped"
	VMStatusPaused  VMStatus = "paused"
	VMStatusError   VMStatus = "error"
)

// VM represents a virtual machine entity
type VM struct {
	ID        string
	Name      string
	VCPU      int
	RAMGB     int
	DiskGB    int
	Status    VMStatus
	IP        string
	Template  string
	DiskPath  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsRunning returns true if VM is in running state
func (v *VM) IsRunning() bool {
	return v.Status == VMStatusRunning
}

// IsStopped returns true if VM is in stopped state
func (v *VM) IsStopped() bool {
	return v.Status == VMStatusStopped
}
