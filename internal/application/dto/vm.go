package dto

// CreateVMRequest represents a request to create a VM
type CreateVMRequest struct {
	Name     string            `json:"name" validate:"required,min=3,max=63,hostname"`
	VCPU     int               `json:"vcpu" validate:"required,min=1,max=32"`
	RAMGB    int               `json:"ram_gb" validate:"required,min=1,max=128"`
	DiskGB   int               `json:"disk_gb" validate:"required,min=10,max=1000"`
	Template string            `json:"template" validate:"required,oneof=ubuntu-22.04 ubuntu-20.04 debian-12 debian-11"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CreateVMResponse represents the response after creating a VM
type CreateVMResponse struct {
	VMID      string `json:"vm_id"`
	IPAddress string `json:"ip_address"`
	Status    string `json:"status"`
}

// DeleteVMRequest represents a request to delete a VM
type DeleteVMRequest struct {
	VMID string `json:"vm_id" validate:"required"`
}

// DeleteVMResponse represents the response after deleting a VM
type DeleteVMResponse struct {
	Success bool `json:"success"`
}

// StartVMRequest represents a request to start a VM
type StartVMRequest struct {
	VMID string `json:"vm_id" validate:"required"`
}

// StartVMResponse represents the response after starting a VM
type StartVMResponse struct {
	Status string `json:"status"`
}

// StopVMRequest represents a request to stop a VM
type StopVMRequest struct {
	VMID  string `json:"vm_id" validate:"required"`
	Force bool   `json:"force"`
}

// StopVMResponse represents the response after stopping a VM
type StopVMResponse struct {
	Status string `json:"status"`
}

// GetVMStatusRequest represents a request to get VM status
type GetVMStatusRequest struct {
	VMID string `json:"vm_id" validate:"required"`
}

// GetVMStatusResponse represents detailed VM status
type GetVMStatusResponse struct {
	VMID            string  `json:"vm_id"`
	Name            string  `json:"name"`
	Status          string  `json:"status"`
	VCPU            int     `json:"vcpu"`
	RAMGB           int     `json:"ram_gb"`
	DiskGB          int     `json:"disk_gb"`
	IPAddress       string  `json:"ip_address"`
	UptimeSeconds   int64   `json:"uptime_seconds"`
	CPUUsagePercent float32 `json:"cpu_usage_percent"`
	RAMUsagePercent float32 `json:"ram_usage_percent"`
}

// ListVMsRequest represents a request to list all VMs
type ListVMsRequest struct {
	// Empty for now - could add filters later
}

// ListVMsResponse represents the response with all VMs
type ListVMsResponse struct {
	VMs []VMInfo `json:"vms"`
}

// VMInfo represents basic VM information
type VMInfo struct {
	VMID      string `json:"vm_id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	IPAddress string `json:"ip_address"`
}
