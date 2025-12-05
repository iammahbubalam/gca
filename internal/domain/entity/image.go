package entity

import "time"

// Image represents a cached OS image
type Image struct {
	Name      string    // e.g., "ubuntu-22.04"
	Path      string    // Local file path
	URL       string    // Download URL
	Checksum  string    // SHA256 checksum
	SizeBytes int64     // File size in bytes
	CachedAt  time.Time // When it was cached
}

// IsValid checks if image checksum matches
func (i *Image) IsValid(actualChecksum string) bool {
	return i.Checksum == actualChecksum
}
