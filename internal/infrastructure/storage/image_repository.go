package storage

import (
	"context"
	"sync"

	"github.com/iammahbubalam/ghost-agent/internal/domain/entity"
	"github.com/iammahbubalam/ghost-agent/internal/domain/errors"
)

// InMemoryImageRepository implements ImageRepository using in-memory storage
type InMemoryImageRepository struct {
	images map[string]*entity.Image
	mu     sync.RWMutex
}

// NewInMemoryImageRepository creates a new in-memory image repository
func NewInMemoryImageRepository() *InMemoryImageRepository {
	return &InMemoryImageRepository{
		images: make(map[string]*entity.Image),
	}
}

// Get retrieves an image by name
func (r *InMemoryImageRepository) Get(ctx context.Context, name string) (*entity.Image, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	image, ok := r.images[name]
	if !ok {
		return nil, errors.New(errors.ErrCodeNotFound, "image not found", nil).
			WithContext("image_name", name)
	}

	return image, nil
}

// Save persists an image
func (r *InMemoryImageRepository) Save(ctx context.Context, image *entity.Image) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.images[image.Name] = image
	return nil
}

// Exists checks if an image exists
func (r *InMemoryImageRepository) Exists(ctx context.Context, name string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.images[name]
	return ok, nil
}

// Delete removes an image
func (r *InMemoryImageRepository) Delete(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.images, name)
	return nil
}
