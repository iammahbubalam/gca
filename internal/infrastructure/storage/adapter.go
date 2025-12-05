package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/iammahbubalam/ghost-agent/internal/domain/entity"
	"github.com/iammahbubalam/ghost-agent/internal/domain/errors"
	"github.com/iammahbubalam/ghost-agent/internal/domain/repository"
)

// Adapter implements StorageService
type Adapter struct {
	imageCache     string
	imageRepo      repository.ImageRepository
	logger         *zap.Logger
	imageTemplates map[string]string // template name -> download URL
}

// NewAdapter creates a new storage adapter
func NewAdapter(imageCache string, imageRepo repository.ImageRepository, logger *zap.Logger) *Adapter {
	// Define OS image templates
	templates := map[string]string{
		"ubuntu-22.04": "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img",
		"ubuntu-20.04": "https://cloud-images.ubuntu.com/releases/20.04/release/ubuntu-20.04-server-cloudimg-amd64.img",
		"debian-12":    "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-generic-amd64.qcow2",
		"debian-11":    "https://cloud.debian.org/images/cloud/bullseye/latest/debian-11-generic-amd64.qcow2",
	}

	return &Adapter{
		imageCache:     imageCache,
		imageRepo:      imageRepo,
		logger:         logger,
		imageTemplates: templates,
	}
}

// GetImage retrieves or downloads an OS image
func (a *Adapter) GetImage(ctx context.Context, template string) (string, error) {
	a.logger.Info("Getting image", zap.String("template", template))

	// Check if image exists in cache
	exists, err := a.imageRepo.Exists(ctx, template)
	if err != nil {
		return "", errors.New(errors.ErrCodeStorage, "failed to check image cache", err).
			WithContext("template", template)
	}

	if exists {
		image, err := a.imageRepo.Get(ctx, template)
		if err != nil {
			return "", errors.New(errors.ErrCodeStorage, "failed to get cached image", err).
				WithContext("template", template)
		}
		
		// Verify checksum
		actualChecksum, err := a.calculateChecksum(image.Path)
		if err == nil && image.IsValid(actualChecksum) {
			a.logger.Info("Using cached image", zap.String("template", template))
			return image.Path, nil
		}
		
		a.logger.Warn("Cached image checksum mismatch, re-downloading", zap.String("template", template))
	}

	// Download image
	url, ok := a.imageTemplates[template]
	if !ok {
		return "", errors.New(errors.ErrCodeValidation, "unknown template", nil).
			WithContext("template", template)
	}

	imagePath := filepath.Join(a.imageCache, fmt.Sprintf("%s.qcow2", template))
	if err := a.downloadImage(url, imagePath); err != nil {
		return "", errors.New(errors.ErrCodeStorage, "failed to download image", err).
			WithContext("template", template).
			WithContext("url", url)
	}

	// Calculate checksum
	checksum, err := a.calculateChecksum(imagePath)
	if err != nil {
		return "", errors.New(errors.ErrCodeStorage, "failed to calculate checksum", err).
			WithContext("path", imagePath)
	}

	// Save to repository
	image := &entity.Image{
		Name:     template,
		Path:     imagePath,
		URL:      url,
		Checksum: checksum,
	}
	
	if err := a.imageRepo.Save(ctx, image); err != nil {
		a.logger.Warn("Failed to save image to repository", zap.Error(err))
	}

	return imagePath, nil
}

// CreateDisk creates a new disk for a VM from a base image
func (a *Adapter) CreateDisk(ctx context.Context, vmID string, baseImage string, sizeGB int) (string, error) {
	a.logger.Info("Creating disk",
		zap.String("vm_id", vmID),
		zap.String("base_image", baseImage),
		zap.Int("size_gb", sizeGB),
	)

	diskPath := filepath.Join(a.imageCache, "disks", fmt.Sprintf("%s.qcow2", vmID))
	
	// Ensure disks directory exists
	if err := os.MkdirAll(filepath.Dir(diskPath), 0755); err != nil {
		return "", errors.New(errors.ErrCodeStorage, "failed to create disks directory", err)
	}

	// Create copy-on-write disk from base image
	cmd := exec.Command("qemu-img", "create",
		"-f", "qcow2",
		"-F", "qcow2",
		"-b", baseImage,
		diskPath,
		fmt.Sprintf("%dG", sizeGB),
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", errors.New(errors.ErrCodeStorage, "failed to create disk", err).
			WithContext("output", string(output))
	}

	return diskPath, nil
}

// DeleteDisk deletes a VM's disk
func (a *Adapter) DeleteDisk(ctx context.Context, vmID string) error {
	a.logger.Info("Deleting disk", zap.String("vm_id", vmID))

	diskPath := filepath.Join(a.imageCache, "disks", fmt.Sprintf("%s.qcow2", vmID))
	
	if err := os.Remove(diskPath); err != nil && !os.IsNotExist(err) {
		return errors.New(errors.ErrCodeStorage, "failed to delete disk", err).
			WithContext("path", diskPath)
	}

	return nil
}

// GetDiskPath returns the path to a VM's disk
func (a *Adapter) GetDiskPath(ctx context.Context, vmID string) (string, error) {
	diskPath := filepath.Join(a.imageCache, "disks", fmt.Sprintf("%s.qcow2", vmID))
	
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		return "", errors.New(errors.ErrCodeNotFound, "disk not found", err).
			WithContext("vm_id", vmID)
	}

	return diskPath, nil
}

// Helper methods

func (a *Adapter) downloadImage(url, destPath string) error {
	a.logger.Info("Downloading image", zap.String("url", url))

	// Ensure cache directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy data
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	a.logger.Info("Image downloaded successfully", zap.String("path", destPath))
	return nil
}

func (a *Adapter) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
