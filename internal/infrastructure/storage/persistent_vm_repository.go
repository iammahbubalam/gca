package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/iammahbubalam/ghost-agent/internal/domain/entity"
	"github.com/iammahbubalam/ghost-agent/internal/domain/errors"
)

// PersistentVMRepository implements VMRepository with file-based persistence
// This ensures VM state survives PC restarts
type PersistentVMRepository struct {
	vms      map[string]*entity.VM
	mu       sync.RWMutex
	filePath string
}

// NewPersistentVMRepository creates a new persistent VM repository
func NewPersistentVMRepository(dataDir string) (*PersistentVMRepository, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(dataDir, "vms.json")
	repo := &PersistentVMRepository{
		vms:      make(map[string]*entity.VM),
		filePath: filePath,
	}

	// Load existing state from disk
	if err := repo.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return repo, nil
}

// Save persists a VM
func (r *PersistentVMRepository) Save(ctx context.Context, vm *entity.VM) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.vms[vm.ID] = vm
	return r.persist()
}

// FindByID retrieves a VM by ID
func (r *PersistentVMRepository) FindByID(ctx context.Context, id string) (*entity.VM, error) {
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
func (r *PersistentVMRepository) FindByName(ctx context.Context, name string) (*entity.VM, error) {
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
func (r *PersistentVMRepository) FindAll(ctx context.Context) ([]*entity.VM, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	vms := make([]*entity.VM, 0, len(r.vms))
	for _, vm := range r.vms {
		vms = append(vms, vm)
	}

	return vms, nil
}

// Delete removes a VM
func (r *PersistentVMRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.vms, id)
	return r.persist()
}

// Exists checks if a VM exists
func (r *PersistentVMRepository) Exists(ctx context.Context, id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.vms[id]
	return ok, nil
}

// persist saves the current state to disk
func (r *PersistentVMRepository) persist() error {
	data, err := json.MarshalIndent(r.vms, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first, then rename (atomic operation)
	tempFile := r.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tempFile, r.filePath)
}

// load reads the state from disk
func (r *PersistentVMRepository) load() error {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &r.vms)
}
