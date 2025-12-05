package usecase

import (
	"context"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/iammahbubalam/ghost-agent/internal/application/dto"
	"github.com/iammahbubalam/ghost-agent/internal/domain/errors"
	"github.com/iammahbubalam/ghost-agent/internal/domain/repository"
	"github.com/iammahbubalam/ghost-agent/internal/domain/service"
)

// GetVMStatusUseCase handles getting VM status
type GetVMStatusUseCase struct {
	hypervisor service.HypervisorService
	network    service.NetworkService
	vmRepo     repository.VMRepository
	validator  *validator.Validate
	logger     *zap.Logger
}

// NewGetVMStatusUseCase creates a new GetVMStatus use case
func NewGetVMStatusUseCase(
	hypervisor service.HypervisorService,
	network service.NetworkService,
	vmRepo repository.VMRepository,
	logger *zap.Logger,
) *GetVMStatusUseCase {
	return &GetVMStatusUseCase{
		hypervisor: hypervisor,
		network:    network,
		vmRepo:     vmRepo,
		validator:  validator.New(),
		logger:     logger,
	}
}

// Execute gets VM status
func (uc *GetVMStatusUseCase) Execute(ctx context.Context, req *dto.GetVMStatusRequest) (*dto.GetVMStatusResponse, error) {
	uc.logger.Debug("Getting VM status", zap.String("vm_id", req.VMID))

	// 1. Validate input
	if err := uc.validator.Struct(req); err != nil {
		return nil, errors.New(errors.ErrCodeValidation, "invalid request", err)
	}

	// 2. Get VM from repository
	vm, err := uc.vmRepo.FindByID(ctx, req.VMID)
	if err != nil {
		return nil, errors.New(errors.ErrCodeNotFound, "VM not found", err).
			WithContext("vm_id", req.VMID)
	}

	// 3. Get detailed status from hypervisor
	status, err := uc.hypervisor.GetVMStatus(ctx, req.VMID)
	if err != nil {
		return nil, errors.New(errors.ErrCodeHypervisor, "failed to get VM status", err).
			WithContext("vm_id", req.VMID)
	}

	// 4. Get current IP
	ip, err := uc.network.GetVMIP(ctx, req.VMID)
	if err != nil {
		uc.logger.Warn("Failed to get VM IP", zap.Error(err))
		ip = vm.IP // Use cached IP
	}

	return &dto.GetVMStatusResponse{
		VMID:            vm.ID,
		Name:            vm.Name,
		Status:          string(status.Status),
		VCPU:            vm.VCPU,
		RAMGB:           vm.RAMGB,
		DiskGB:          vm.DiskGB,
		IPAddress:       ip,
		UptimeSeconds:   status.UptimeSeconds,
		CPUUsagePercent: status.CPUUsagePercent,
		RAMUsagePercent: status.RAMUsagePercent,
	}, nil
}
