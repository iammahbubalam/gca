package repository

import (
	"context"
	
	"github.com/iammahbubalam/ghost-agent/internal/domain/entity"
)

// ImageRepository defines the interface for image management
type ImageRepository interface {
	// Get retrieves an image by name
	Get(ctx context.Context, name string) (*entity.Image, error)
	
	// Save persists an image
	Save(ctx context.Context, image *entity.Image) error
	
	// Exists checks if an image exists
	Exists(ctx context.Context, name string) (bool, error)
	
	// Delete removes an image
	Delete(ctx context.Context, name string) error
}
