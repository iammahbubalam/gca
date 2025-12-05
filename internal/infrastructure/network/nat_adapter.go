package network

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"libvirt.org/go/libvirt"

	"github.com/iammahbubalam/ghost-agent/internal/domain/errors"
)

// NATAdapter implements NetworkService using simple NAT networking
// This is the v1.0 implementation - designed to be replaced with more advanced networking later
type NATAdapter struct {
	conn   *libvirt.Connect
	logger *zap.Logger
}

// NewNATAdapter creates a new NAT network adapter
func NewNATAdapter(conn *libvirt.Connect, logger *zap.Logger) *NATAdapter {
	return &NATAdapter{
		conn:   conn,
		logger: logger,
	}
}

// AssignIP assigns an IP address to a VM
// In NAT mode, this is handled by DHCP automatically
func (n *NATAdapter) AssignIP(ctx context.Context, vmID string) (string, error) {
	n.logger.Debug("Assigning IP via NAT DHCP", zap.String("vm_id", vmID))
	
	// In NAT mode, IP is assigned by libvirt's DHCP server automatically
	// We just need to wait for it and return it
	ip, err := n.waitForDHCPLease(vmID, 2*time.Minute)
	if err != nil {
		return "", errors.New(errors.ErrCodeNetwork, "failed to get DHCP lease", err).
			WithContext("vm_id", vmID)
	}
	
	return ip, nil
}

// ReleaseIP releases an IP address from a VM
// In NAT mode, this is handled automatically when VM is deleted
func (n *NATAdapter) ReleaseIP(ctx context.Context, vmID string) error {
	n.logger.Debug("Releasing IP (NAT mode - automatic)", zap.String("vm_id", vmID))
	// In NAT mode, IP is released automatically when VM is destroyed
	return nil
}

// GetVMIP retrieves the current IP of a VM
func (n *NATAdapter) GetVMIP(ctx context.Context, vmID string) (string, error) {
	domain, err := n.conn.LookupDomainByName(vmID)
	if err != nil {
		return "", errors.New(errors.ErrCodeNotFound, "VM not found", err).
			WithContext("vm_id", vmID)
	}
	defer domain.Free()
	
	ifaces, err := domain.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
	if err != nil {
		return "", errors.New(errors.ErrCodeNetwork, "failed to get interfaces", err).
			WithContext("vm_id", vmID)
	}
	
	if len(ifaces) == 0 || len(ifaces[0].Addrs) == 0 {
		return "", errors.New(errors.ErrCodeNetwork, "no IP address found", nil).
			WithContext("vm_id", vmID)
	}
	
	return ifaces[0].Addrs[0].Addr, nil
}

// waitForDHCPLease waits for a VM to get an IP from DHCP
func (n *NATAdapter) waitForDHCPLease(vmID string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		ip, err := n.GetVMIP(context.Background(), vmID)
		if err == nil && ip != "" {
			return ip, nil
		}
		time.Sleep(2 * time.Second)
	}
	
	return "", fmt.Errorf("timeout waiting for DHCP lease")
}
