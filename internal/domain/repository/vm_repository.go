package repository

import (
	"context"
	
	"github.com/iammahbubalam/ghost-agent/internal/domain/entity"
)

// VMRepository defines the interface for VM persistence
type VMRepository interface {
	// Save persists a VM
	Save(ctx context.Context, vm *entity.VM) error
	
	// FindByID retrieves a VM by ID
	FindByID(ctx context.Context, id string) (*entity.VM, error)
	
	// FindByName retrieves a VM by name
	FindByName(ctx context.Context, name string) (*entity.VM, error)
	
	// FindAll retrieves all VMs
	FindAll(ctx context.Context) ([]*entity.VM, error)
	
	// Delete removes a VM
	Delete(ctx context.Context, id string) error
	
	// Exists checks if a VM exists
	Exists(ctx context.Context, id string) (bool, error)
}
