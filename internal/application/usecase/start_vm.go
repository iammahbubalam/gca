package usecase

import (
	"context"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/iammahbubalam/ghost-agent/internal/application/dto"
	"github.com/iammahbubalam/ghost-agent/internal/domain/errors"
	"github.com/iammahbubalam/ghost-agent/internal/domain/service"
)

// StartVMUseCase handles VM start business logic
type StartVMUseCase struct {
	hypervisor service.HypervisorService
	validator  *validator.Validate
	logger     *zap.Logger
}

// NewStartVMUseCase creates a new StartVM use case
func NewStartVMUseCase(
	hypervisor service.HypervisorService,
	logger *zap.Logger,
) *StartVMUseCase {
	return &StartVMUseCase{
		hypervisor: hypervisor,
		validator:  validator.New(),
		logger:     logger,
	}
}

// Execute starts a VM
func (uc *StartVMUseCase) Execute(ctx context.Context, req *dto.StartVMRequest) (*dto.StartVMResponse, error) {
	uc.logger.Info("Starting VM", zap.String("vm_id", req.VMID))

	// 1. Validate input
	if err := uc.validator.Struct(req); err != nil {
		return nil, errors.New(errors.ErrCodeValidation, "invalid request", err)
	}

	// 2. Start VM
	if err := uc.hypervisor.StartVM(ctx, req.VMID); err != nil {
		return nil, errors.New(errors.ErrCodeHypervisor, "failed to start VM", err).
			WithContext("vm_id", req.VMID)
	}

	uc.logger.Info("VM started successfully", zap.String("vm_id", req.VMID))

	return &dto.StartVMResponse{
		Status: "running",
	}, nil
}
