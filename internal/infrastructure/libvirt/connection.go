package libvirt

import "libvirt.org/go/libvirt"

// NewConnect creates a new libvirt connection
// This is a wrapper to make it easier to use in other packages
func NewConnect(uri string) (*libvirt.Connect, error) {
	return libvirt.NewConnect(uri)
}
