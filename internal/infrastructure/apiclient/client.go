package apiclient

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/iammahbubalam/ghost-agent/internal/domain/entity"
	"github.com/iammahbubalam/ghost-agent/internal/infrastructure/observability"
	"github.com/iammahbubalam/ghost-agent/pkg/ghostapi"
)

// Client handles communication with Ghost Core API
type Client struct {
	apiURL      string
	conn        *grpc.ClientConn
	client      ghostapi.GhostCoreServiceClient
	logger      *zap.Logger
	metrics     *observability.Metrics
	agentID     string
	agentName   string
	tailscaleIP string
}

// NewClient creates a new Ghost Core API client
func NewClient(apiURL, agentName, tailscaleIP string, logger *zap.Logger, metrics *observability.Metrics) (*Client, error) {
	logger.Info("Connecting to Ghost Core API", zap.String("url", apiURL))

	// TODO: Add mTLS credentials when Ghost Core is ready
	conn, err := grpc.Dial(apiURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ghost Core API: %w", err)
	}

	client := ghostapi.NewGhostCoreServiceClient(conn)

	return &Client{
		apiURL:      apiURL,
		conn:        conn,
		client:      client,
		logger:      logger,
		metrics:     metrics,
		agentName:   agentName,
		tailscaleIP: tailscaleIP,
	}, nil
}

// RegisterAgent registers this agent with Ghost Core
func (c *Client) RegisterAgent(ctx context.Context, resources *entity.Resource, version string) (string, error) {
	c.logger.Info("Registering agent with Ghost Core",
		zap.String("agent_name", c.agentName),
		zap.String("tailscale_ip", c.tailscaleIP),
		zap.Int("total_cpu", resources.TotalCPU),
		zap.Int("total_ram_gb", resources.TotalRAMGB),
	)

	req := &ghostapi.RegisterAgentRequest{
		AgentName:   c.agentName,
		TailscaleIp: c.tailscaleIP,
		Version:     version,
		Resources: &ghostapi.ResourceInfo{
			TotalCpu:       int32(resources.TotalCPU),
			AvailableCpu:   int32(resources.AvailableCPU),
			TotalRamGb:     int32(resources.TotalRAMGB),
			AvailableRamGb: int32(resources.AvailableRAMGB),
			TotalDiskGb:    int32(resources.TotalDiskGB),
			AvailableDiskGb: int32(resources.AvailableDiskGB),
		},
	}

	resp, err := c.client.RegisterAgent(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to register agent: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("registration failed: %s", resp.Message)
	}

	c.agentID = resp.AgentId
	c.logger.Info("Agent registered successfully",
		zap.String("agent_id", c.agentID),
		zap.String("message", resp.Message),
	)

	return c.agentID, nil
}

// SendHeartbeat sends periodic heartbeat to Ghost Core
func (c *Client) SendHeartbeat(ctx context.Context, resources *entity.Resource, vms []*entity.VM) error {
	return retry.Do(
		func() error {
			return c.sendHeartbeatInternal(ctx, resources, vms)
		},
		retry.Attempts(3),
		retry.Delay(1*time.Second),
		retry.MaxDelay(10*time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			c.logger.Warn("Heartbeat retry",
				zap.Uint("attempt", n),
				zap.Error(err),
			)
		}),
		retry.Context(ctx),
	)
}

func (c *Client) sendHeartbeatInternal(ctx context.Context, resources *entity.Resource, vms []*entity.VM) error {
	start := time.Now()
	defer func() {
		c.metrics.APICallLatency.WithLabelValues("heartbeat").Observe(time.Since(start).Seconds())
	}()

	// Convert VMs to protobuf
	vmInfos := make([]*ghostapi.VMInfo, len(vms))
	for i, vm := range vms {
		vmInfos[i] = &ghostapi.VMInfo{
			VmId:      vm.ID,
			VmName:    vm.Name,
			Status:    string(vm.Status),
			IpAddress: vm.IP,
			Vcpu:      int32(vm.VCPU),
			RamGb:     int32(vm.RAMGB),
		}
	}

	req := &ghostapi.HeartbeatRequest{
		AgentId:   c.agentID,
		Timestamp: time.Now().Unix(),
		Resources: &ghostapi.ResourceInfo{
			TotalCpu:       int32(resources.TotalCPU),
			AvailableCpu:   int32(resources.AvailableCPU),
			TotalRamGb:     int32(resources.TotalRAMGB),
			AvailableRamGb: int32(resources.AvailableRAMGB),
			TotalDiskGb:    int32(resources.TotalDiskGB),
			AvailableDiskGb: int32(resources.AvailableDiskGB),
		},
		Vms: vmInfos,
	}

	c.logger.Debug("Sending heartbeat",
		zap.String("agent_id", c.agentID),
		zap.Int("vms_count", len(vms)),
		zap.Int("available_cpu", resources.AvailableCPU),
	)

	resp, err := c.client.Heartbeat(ctx, req)
	if err != nil {
		c.metrics.HeartbeatSuccess.Set(0)
		return fmt.Errorf("heartbeat failed: %w", err)
	}

	if !resp.Success {
		c.metrics.HeartbeatSuccess.Set(0)
		return fmt.Errorf("heartbeat rejected: %s", resp.Message)
	}

	c.metrics.HeartbeatSuccess.Set(1)

	// Process commands from Ghost Core (if any)
	if len(resp.Commands) > 0 {
		c.logger.Info("Received commands from Ghost Core", zap.Int("count", len(resp.Commands)))
		// TODO: Implement command processing
	}

	return nil
}

// ReportVMCreated reports a newly created VM to Ghost Core
func (c *Client) ReportVMCreated(ctx context.Context, vm *entity.VM) error {
	c.logger.Info("Reporting VM creation to Ghost Core",
		zap.String("vm_id", vm.ID),
		zap.String("vm_name", vm.Name),
	)

	req := &ghostapi.ReportVMCreatedRequest{
		AgentId:   c.agentID,
		VmId:      vm.ID,
		VmName:    vm.Name,
		Vcpu:      int32(vm.VCPU),
		RamGb:     int32(vm.RAMGB),
		DiskGb:    int32(vm.DiskGB),
		IpAddress: vm.IP,
		Template:  vm.Template,
	}

	resp, err := c.client.ReportVMCreated(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to report VM creation: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("VM creation report rejected: %s", resp.Message)
	}

	return nil
}

// ReportVMDeleted reports a deleted VM to Ghost Core
func (c *Client) ReportVMDeleted(ctx context.Context, vmID string) error {
	c.logger.Info("Reporting VM deletion to Ghost Core", zap.String("vm_id", vmID))

	req := &ghostapi.ReportVMDeletedRequest{
		AgentId: c.agentID,
		VmId:    vmID,
	}

	resp, err := c.client.ReportVMDeleted(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to report VM deletion: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("VM deletion report rejected")
	}

	return nil
}

// ReportVMStatusChange reports a VM status change to Ghost Core
func (c *Client) ReportVMStatusChange(ctx context.Context, vmID string, status entity.VMStatus) error {
	c.logger.Debug("Reporting VM status change to Ghost Core",
		zap.String("vm_id", vmID),
		zap.String("status", string(status)),
	)

	req := &ghostapi.ReportVMStatusChangeRequest{
		AgentId: c.agentID,
		VmId:    vmID,
		Status:  string(status),
	}

	resp, err := c.client.ReportVMStatusChange(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to report VM status change: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("VM status change report rejected")
	}

	return nil
}

// StartHeartbeat starts the heartbeat goroutine
func (c *Client) StartHeartbeat(ctx context.Context, interval time.Duration, getResources func() *entity.Resource, getVMs func() []*entity.VM) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	c.logger.Info("Starting heartbeat", zap.Duration("interval", interval))

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping heartbeat")
			return
		case <-ticker.C:
			resources := getResources()
			vms := getVMs()

			if err := c.SendHeartbeat(ctx, resources, vms); err != nil {
				c.logger.Error("Heartbeat failed", zap.Error(err))
				c.metrics.HeartbeatSuccess.Set(0)
			} else {
				c.logger.Debug("Heartbeat sent successfully")
			}
		}
	}
}

// UnregisterAgent unregisters the agent from Ghost Core
func (c *Client) UnregisterAgent(ctx context.Context) error {
	c.logger.Info("Unregistering agent from Ghost Core", zap.String("agent_id", c.agentID))

	req := &ghostapi.UnregisterAgentRequest{
		AgentId: c.agentID,
	}

	resp, err := c.client.UnregisterAgent(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to unregister agent: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("unregister rejected")
	}

	c.logger.Info("Agent unregistered successfully")
	return nil
}

// Close closes the API client connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Ping checks if the API connection is alive
func (c *Client) Ping() error {
	if c.conn == nil {
		return fmt.Errorf("not connected to Ghost Core API")
	}
	// TODO: Add actual ping RPC when Ghost Core implements it
	return nil
}

// GetAgentID returns the agent ID
func (c *Client) GetAgentID() string {
	return c.agentID
}
