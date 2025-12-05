package storage

import (
	"context"
	"sync"

	"github.com/iammahbubalam/ghost-agent/internal/domain/entity"
	"github.com/iammahbubalam/ghost-agent/internal/domain/errors"
)

// InMemoryVMRepository implements VMRepository using in-memory storage
type InMemoryVMRepository struct {
	vms map[string]*entity.VM
	mu  sync.RWMutex
}

// NewInMemoryVMRepository creates a new in-memory VM repository
func NewInMemoryVMRepository() *InMemoryVMRepository {
	return &InMemoryVMRepository{
		vms: make(map[string]*entity.VM),
	}
}

// Save persists a VM
func (r *InMemoryVMRepository) Save(ctx context.Context, vm *entity.VM) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.vms[vm.ID] = vm
	return nil
}

// FindByID retrieves a VM by ID
func (r *InMemoryVMRepository) FindByID(ctx context.Context, id string) (*entity.VM, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	vm, ok := r.vms[id]
	if !ok {
		return nil, errors.New(errors.ErrCodeNotFound, "VM not found", nil).
			WithContext("vm_id", id)
	}

	return vm, nil
}

// FindByName retrieves a VM by name
func (r *InMemoryVMRepository) FindByName(ctx context.Context, name string) (*entity.VM, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, vm := range r.vms {
		if vm.Name == name {
			return vm, nil
		}
	}

	return nil, errors.New(errors.ErrCodeNotFound, "VM not found", nil).
		WithContext("vm_name", name)
}

// FindAll retrieves all VMs
func (r *InMemoryVMRepository) FindAll(ctx context.Context) ([]*entity.VM, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	vms := make([]*entity.VM, 0, len(r.vms))
	for _, vm := range r.vms {
		vms = append(vms, vm)
	}

	return vms, nil
}

// Delete removes a VM
func (r *InMemoryVMRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.vms, id)
	return nil
}

// Exists checks if a VM exists
func (r *InMemoryVMRepository) Exists(ctx context.Context, id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.vms[id]
	return ok, nil
}
