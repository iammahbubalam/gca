package service

import "context"

// NetworkService defines the interface for network operations
// This is designed to be extensible for future networking implementations
type NetworkService interface {
	// AssignIP assigns an IP address to a VM
	// v1.0: Returns IP from NAT DHCP (192.168.122.x)
	// Future: Could support custom networks, VLANs, etc.
	AssignIP(ctx context.Context, vmID string) (string, error)
	
	// ReleaseIP releases an IP address from a VM
	ReleaseIP(ctx context.Context, vmID string) error
	
	// GetVMIP retrieves the current IP of a VM
	GetVMIP(ctx context.Context, vmID string) (string, error)
}
