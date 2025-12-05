package usecase

import (
	"context"

	"go.uber.org/zap"

	"github.com/iammahbubalam/ghost-agent/internal/application/dto"
	"github.com/iammahbubalam/ghost-agent/internal/domain/errors"
	"github.com/iammahbubalam/ghost-agent/internal/domain/service"
)

// ListVMsUseCase handles listing all VMs
type ListVMsUseCase struct {
	hypervisor service.HypervisorService
	network    service.NetworkService
	logger     *zap.Logger
}

// NewListVMsUseCase creates a new ListVMs use case
func NewListVMsUseCase(
	hypervisor service.HypervisorService,
	network service.NetworkService,
	logger *zap.Logger,
) *ListVMsUseCase {
	return &ListVMsUseCase{
		hypervisor: hypervisor,
		network:    network,
		logger:     logger,
	}
}

// Execute lists all VMs
func (uc *ListVMsUseCase) Execute(ctx context.Context, req *dto.ListVMsRequest) (*dto.ListVMsResponse, error) {
	uc.logger.Debug("Listing VMs")

	// 1. Get all VMs from hypervisor
	vms, err := uc.hypervisor.ListVMs(ctx)
	if err != nil {
		return nil, errors.New(errors.ErrCodeHypervisor, "failed to list VMs", err)
	}

	// 2. Build response
	vmInfos := make([]dto.VMInfo, 0, len(vms))
	for _, vm := range vms {
		// Try to get IP
		ip, err := uc.network.GetVMIP(ctx, vm.ID)
		if err != nil {
			ip = vm.IP // Use cached IP if available
		}

		vmInfos = append(vmInfos, dto.VMInfo{
			VMID:      vm.ID,
			Name:      vm.Name,
			Status:    string(vm.Status),
			IPAddress: ip,
		})
	}

	return &dto.ListVMsResponse{
		VMs: vmInfos,
	}, nil
}
