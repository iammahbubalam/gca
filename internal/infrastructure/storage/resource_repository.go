package storage

import (
	"context"
	"runtime"
	"sync"

	"github.com/iammahbubalam/ghost-agent/internal/domain/entity"
)

// InMemoryResourceRepository implements ResourceRepository using in-memory storage
type InMemoryResourceRepository struct {
	resource *entity.Resource
	mu       sync.RWMutex
}

// NewInMemoryResourceRepository creates a new in-memory resource repository
func NewInMemoryResourceRepository(reservedCPU, reservedRAMGB, reservedDiskGB int) *InMemoryResourceRepository {
	totalCPU := runtime.NumCPU()
	// TODO: Get actual total RAM and disk from system
	totalRAMGB := 32   // Placeholder
	totalDiskGB := 500 // Placeholder

	return &InMemoryResourceRepository{
		resource: &entity.Resource{
			TotalCPU:       totalCPU,
			AvailableCPU:   totalCPU - reservedCPU,
			ReservedCPU:    reservedCPU,
			TotalRAMGB:     totalRAMGB,
			AvailableRAMGB: totalRAMGB - reservedRAMGB,
			ReservedRAMGB:  reservedRAMGB,
			TotalDiskGB:    totalDiskGB,
			AvailableDiskGB: totalDiskGB - reservedDiskGB,
			ReservedDiskGB: reservedDiskGB,
		},
	}
}

// GetAvailable returns current available resources
func (r *InMemoryResourceRepository) GetAvailable(ctx context.Context) (*entity.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	resourceCopy := *r.resource
	return &resourceCopy, nil
}

// Update updates resource availability
func (r *InMemoryResourceRepository) Update(ctx context.Context, resource *entity.Resource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.resource = resource
	return nil
}
