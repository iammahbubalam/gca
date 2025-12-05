package service

import "context"

// StorageService defines the interface for storage operations
type StorageService interface {
	// GetImage retrieves or downloads an OS image
	// Returns the local path to the image
	GetImage(ctx context.Context, template string) (string, error)
	
	// CreateDisk creates a new disk for a VM from a base image
	// Returns the path to the created disk
	CreateDisk(ctx context.Context, vmID string, baseImage string, sizeGB int) (string, error)
	
	// DeleteDisk deletes a VM's disk
	DeleteDisk(ctx context.Context, vmID string) error
	
	// GetDiskPath returns the path to a VM's disk
	GetDiskPath(ctx context.Context, vmID string) (string, error)
}
