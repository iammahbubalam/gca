package libvirt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
	"libvirt.org/go/libvirt"

	"github.com/iammahbubalam/ghost-agent/internal/domain/entity"
	"github.com/iammahbubalam/ghost-agent/internal/domain/errors"
	"github.com/iammahbubalam/ghost-agent/internal/domain/service"
)

// Adapter implements HypervisorService using Libvirt
type Adapter struct {
	conn           *libvirt.Connect
	circuitBreaker *gobreaker.CircuitBreaker
	logger         *zap.Logger
	mu             sync.RWMutex
}

// NewAdapter creates a new Libvirt adapter with circuit breaker
func NewAdapter(uri string, logger *zap.Logger) (*Adapter, error) {
	conn, err := libvirt.NewConnect(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}

	// Configure circuit breaker
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "libvirt",
		MaxRequests: 3,
		Interval:    time.Minute,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn("Circuit breaker state changed",
				zap.String("name", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	})

	adapter := &Adapter{
		conn:           conn,
		circuitBreaker: cb,
		logger:         logger,
	}

	return adapter, nil
}

// CreateVM creates a new virtual machine
func (a *Adapter) CreateVM(ctx context.Context, spec *service.VMSpec) (*entity.VM, error) {
	a.logger.Info("Creating VM",
		zap.String("name", spec.Name),
		zap.Int("vcpu", spec.VCPU),
		zap.Int("ram_gb", spec.RAMGB),
	)

	// Execute with circuit breaker
	result, err := a.circuitBreaker.Execute(func() (interface{}, error) {
		return a.createVMInternal(ctx, spec)
	})

	if err != nil {
		return nil, errors.New(errors.ErrCodeHypervisor, "failed to create VM", err).
			WithContext("vm_name", spec.Name)
	}

	return result.(*entity.VM), nil
}

func (a *Adapter) createVMInternal(ctx context.Context, spec *service.VMSpec) (*entity.VM, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Generate VM XML
	xml := a.generateVMXML(spec)

	// Define domain
	domain, err := a.conn.DomainDefineXML(xml)
	if err != nil {
		return nil, fmt.Errorf("failed to define domain: %w", err)
	}
	defer domain.Free()

	// Start domain
	if err := domain.Create(); err != nil {
		return nil, fmt.Errorf("failed to start domain: %w", err)
	}

	// Wait for IP address (with timeout)
	ip, err := a.waitForIP(domain, 2*time.Minute)
	if err != nil {
		a.logger.Warn("Failed to get VM IP", zap.Error(err))
		ip = "" // Continue without IP
	}

	vm := &entity.VM{
		ID:        spec.Name,
		Name:      spec.Name,
		VCPU:      spec.VCPU,
		RAMGB:     spec.RAMGB,
		DiskGB:    spec.DiskGB,
		Status:    entity.VMStatusRunning,
		IP:        ip,
		Template:  spec.Template,
		DiskPath:  spec.DiskPath,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return vm, nil
}

// DeleteVM deletes a virtual machine
func (a *Adapter) DeleteVM(ctx context.Context, id string) error {
	a.logger.Info("Deleting VM", zap.String("id", id))

	_, err := a.circuitBreaker.Execute(func() (interface{}, error) {
		return nil, a.deleteVMInternal(ctx, id)
	})

	if err != nil {
		return errors.New(errors.ErrCodeHypervisor, "failed to delete VM", err).
			WithContext("vm_id", id)
	}

	return nil
}

func (a *Adapter) deleteVMInternal(ctx context.Context, id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	domain, err := a.conn.LookupDomainByName(id)
	if err != nil {
		return fmt.Errorf("failed to lookup domain: %w", err)
	}
	defer domain.Free()

	// Stop if running
	state, _, err := domain.GetState()
	if err != nil {
		return fmt.Errorf("failed to get domain state: %w", err)
	}

	if state == libvirt.DOMAIN_RUNNING {
		if err := domain.Destroy(); err != nil {
			return fmt.Errorf("failed to stop domain: %w", err)
		}
	}

	// Undefine domain
	if err := domain.Undefine(); err != nil {
		return fmt.Errorf("failed to undefine domain: %w", err)
	}

	return nil
}

// StartVM starts a stopped virtual machine
func (a *Adapter) StartVM(ctx context.Context, id string) error {
	a.logger.Info("Starting VM", zap.String("id", id))

	_, err := a.circuitBreaker.Execute(func() (interface{}, error) {
		return nil, a.startVMInternal(ctx, id)
	})

	if err != nil {
		return errors.New(errors.ErrCodeHypervisor, "failed to start VM", err).
			WithContext("vm_id", id)
	}

	return nil
}

func (a *Adapter) startVMInternal(ctx context.Context, id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	domain, err := a.conn.LookupDomainByName(id)
	if err != nil {
		return fmt.Errorf("failed to lookup domain: %w", err)
	}
	defer domain.Free()

	if err := domain.Create(); err != nil {
		return fmt.Errorf("failed to start domain: %w", err)
	}

	return nil
}

// StopVM stops a running virtual machine
func (a *Adapter) StopVM(ctx context.Context, id string, force bool) error {
	a.logger.Info("Stopping VM", zap.String("id", id), zap.Bool("force", force))

	_, err := a.circuitBreaker.Execute(func() (interface{}, error) {
		return nil, a.stopVMInternal(ctx, id, force)
	})

	if err != nil {
		return errors.New(errors.ErrCodeHypervisor, "failed to stop VM", err).
			WithContext("vm_id", id)
	}

	return nil
}

func (a *Adapter) stopVMInternal(ctx context.Context, id string, force bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	domain, err := a.conn.LookupDomainByName(id)
	if err != nil {
		return fmt.Errorf("failed to lookup domain: %w", err)
	}
	defer domain.Free()

	if force {
		if err := domain.Destroy(); err != nil {
			return fmt.Errorf("failed to force stop domain: %w", err)
		}
	} else {
		if err := domain.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown domain: %w", err)
		}
	}

	return nil
}

// GetVMStatus retrieves detailed status of a VM
func (a *Adapter) GetVMStatus(ctx context.Context, id string) (*service.VMStatusInfo, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	domain, err := a.conn.LookupDomainByName(id)
	if err != nil {
		return nil, errors.New(errors.ErrCodeNotFound, "VM not found", err).
			WithContext("vm_id", id)
	}
	defer domain.Free()

	state, _, err := domain.GetState()
	if err != nil {
		return nil, fmt.Errorf("failed to get domain state: %w", err)
	}

	status := &service.VMStatusInfo{
		Status:          a.mapLibvirtState(state),
		UptimeSeconds:   0, // TODO: Calculate uptime
		CPUUsagePercent: 0, // TODO: Get CPU usage
		RAMUsagePercent: 0, // TODO: Get RAM usage
	}

	return status, nil
}

// ListVMs lists all VMs
func (a *Adapter) ListVMs(ctx context.Context) ([]*entity.VM, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	domains, err := a.conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
	if err != nil {
		return nil, errors.New(errors.ErrCodeHypervisor, "failed to list domains", err)
	}

	vms := make([]*entity.VM, 0, len(domains))
	for _, domain := range domains {
		name, _ := domain.GetName()
		state, _, _ := domain.GetState()

		vm := &entity.VM{
			ID:     name,
			Name:   name,
			Status: a.mapLibvirtState(state),
		}
		vms = append(vms, vm)
		domain.Free()
	}

	return vms, nil
}

// Ping checks if hypervisor connection is alive
func (a *Adapter) Ping(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if _, err := a.conn.GetVersion(); err != nil {
		return fmt.Errorf("libvirt connection is down: %w", err)
	}

	return nil
}

// Close closes the Libvirt connection
func (a *Adapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.conn != nil {
		if _, err := a.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Helper methods

func (a *Adapter) mapLibvirtState(state libvirt.DomainState) entity.VMStatus {
	switch state {
	case libvirt.DOMAIN_RUNNING:
		return entity.VMStatusRunning
	case libvirt.DOMAIN_SHUTOFF:
		return entity.VMStatusStopped
	case libvirt.DOMAIN_PAUSED:
		return entity.VMStatusPaused
	default:
		return entity.VMStatusError
	}
}

func (a *Adapter) waitForIP(domain *libvirt.Domain, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		ifaces, err := domain.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
		if err == nil && len(ifaces) > 0 {
			for _, iface := range ifaces {
				if len(iface.Addrs) > 0 {
					return iface.Addrs[0].Addr, nil
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
	
	return "", fmt.Errorf("timeout waiting for IP address")
}

func (a *Adapter) generateVMXML(spec *service.VMSpec) string {
	// Simplified XML generation - in production, use proper XML templating
	return fmt.Sprintf(`
<domain type='kvm'>
  <name>%s</name>
  <memory unit='GiB'>%d</memory>
  <vcpu>%d</vcpu>
  <os>
    <type arch='x86_64'>hvm</type>
    <boot dev='hd'/>
  </os>
  <devices>
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='%s'/>
      <target dev='vda' bus='virtio'/>
    </disk>
    <interface type='network'>
      <source network='default'/>
      <model type='virtio'/>
    </interface>
    <console type='pty'/>
  </devices>
</domain>
`, spec.Name, spec.RAMGB, spec.VCPU, spec.DiskPath)
}
