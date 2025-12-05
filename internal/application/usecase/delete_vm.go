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

// DeleteVMUseCase handles VM deletion business logic
type DeleteVMUseCase struct {
	hypervisor   service.HypervisorService
	storage      service.StorageService
	vmRepo       repository.VMRepository
	resourceRepo repository.ResourceRepository
	validator    *validator.Validate
	logger       *zap.Logger
}

// NewDeleteVMUseCase creates a new DeleteVM use case
func NewDeleteVMUseCase(
	hypervisor service.HypervisorService,
	storage service.StorageService,
	vmRepo repository.VMRepository,
	resourceRepo repository.ResourceRepository,
	logger *zap.Logger,
) *DeleteVMUseCase {
	return &DeleteVMUseCase{
		hypervisor:   hypervisor,
		storage:      storage,
		vmRepo:       vmRepo,
		resourceRepo: resourceRepo,
		validator:    validator.New(),
		logger:       logger,
	}
}

// Execute deletes a VM
func (uc *DeleteVMUseCase) Execute(ctx context.Context, req *dto.DeleteVMRequest) (*dto.DeleteVMResponse, error) {
	uc.logger.Info("Deleting VM", zap.String("vm_id", req.VMID))

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

	// 3. Delete VM from hypervisor
	if err := uc.hypervisor.DeleteVM(ctx, req.VMID); err != nil {
		return nil, errors.New(errors.ErrCodeHypervisor, "failed to delete VM", err).
			WithContext("vm_id", req.VMID)
	}

	// 4. Delete disk
	if err := uc.storage.DeleteDisk(ctx, req.VMID); err != nil {
		uc.logger.Warn("Failed to delete disk", zap.Error(err))
		// Continue even if disk deletion fails
	}

	// 5. Remove from repository
	if err := uc.vmRepo.Delete(ctx, req.VMID); err != nil {
		uc.logger.Error("Failed to delete VM from repository", zap.Error(err))
	}

	// 6. Release resources
	resources, err := uc.resourceRepo.GetAvailable(ctx)
	if err == nil {
		resources.Release(vm.VCPU, vm.RAMGB, vm.DiskGB)
		if err := uc.resourceRepo.Update(ctx, resources); err != nil {
			uc.logger.Error("Failed to update resources", zap.Error(err))
		}
	}

	uc.logger.Info("VM deleted successfully", zap.String("vm_id", req.VMID))

	return &dto.DeleteVMResponse{
		Success: true,
	}, nil
}
