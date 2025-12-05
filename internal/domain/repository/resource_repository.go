package repository

import (
	"context"
	
	"github.com/iammahbubalam/ghost-agent/internal/domain/entity"
)

// ResourceRepository defines the interface for resource management
type ResourceRepository interface {
	// GetAvailable returns current available resources
	GetAvailable(ctx context.Context) (*entity.Resource, error)
	
	// Update updates resource availability
	Update(ctx context.Context, resource *entity.Resource) error
}
