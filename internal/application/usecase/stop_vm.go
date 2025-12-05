package usecase

import (
	"context"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/iammahbubalam/ghost-agent/internal/application/dto"
	"github.com/iammahbubalam/ghost-agent/internal/domain/errors"
	"github.com/iammahbubalam/ghost-agent/internal/domain/service"
)

// StopVMUseCase handles VM stop business logic
type StopVMUseCase struct {
	hypervisor service.HypervisorService
	validator  *validator.Validate
	logger     *zap.Logger
}

// NewStopVMUseCase creates a new StopVM use case
func NewStopVMUseCase(
	hypervisor service.HypervisorService,
	logger *zap.Logger,
) *StopVMUseCase {
	return &StopVMUseCase{
		hypervisor: hypervisor,
		validator:  validator.New(),
		logger:     logger,
	}
}

// Execute stops a VM
func (uc *StopVMUseCase) Execute(ctx context.Context, req *dto.StopVMRequest) (*dto.StopVMResponse, error) {
	uc.logger.Info("Stopping VM",
		zap.String("vm_id", req.VMID),
		zap.Bool("force", req.Force),
	)

	// 1. Validate input
	if err := uc.validator.Struct(req); err != nil {
		return nil, errors.New(errors.ErrCodeValidation, "invalid request", err)
	}

	// 2. Stop VM
	if err := uc.hypervisor.StopVM(ctx, req.VMID, req.Force); err != nil {
		return nil, errors.New(errors.ErrCodeHypervisor, "failed to stop VM", err).
			WithContext("vm_id", req.VMID)
	}

	uc.logger.Info("VM stopped successfully", zap.String("vm_id", req.VMID))

	return &dto.StopVMResponse{
		Status: "stopped",
	}, nil
}
