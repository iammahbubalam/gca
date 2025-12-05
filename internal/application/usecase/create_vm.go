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

// CreateVMUseCase handles VM creation business logic
type CreateVMUseCase struct {
	hypervisor   service.HypervisorService
	network      service.NetworkService
	storage      service.StorageService
	vmRepo       repository.VMRepository
	resourceRepo repository.ResourceRepository
	validator    *validator.Validate
	logger       *zap.Logger
}

// NewCreateVMUseCase creates a new CreateVM use case
func NewCreateVMUseCase(
	hypervisor service.HypervisorService,
	network service.NetworkService,
	storage service.StorageService,
	vmRepo repository.VMRepository,
	resourceRepo repository.ResourceRepository,
	logger *zap.Logger,
) *CreateVMUseCase {
	return &CreateVMUseCase{
		hypervisor:   hypervisor,
		network:      network,
		storage:      storage,
		vmRepo:       vmRepo,
		resourceRepo: resourceRepo,
		validator:    validator.New(),
		logger:       logger,
	}
}

// Execute creates a new VM
func (uc *CreateVMUseCase) Execute(ctx context.Context, req *dto.CreateVMRequest) (*dto.CreateVMResponse, error) {
	uc.logger.Info("Creating VM",
		zap.String("name", req.Name),
		zap.Int("vcpu", req.VCPU),
		zap.Int("ram_gb", req.RAMGB),
	)

	// 1. Validate input
	if err := uc.validator.Struct(req); err != nil {
		return nil, errors.New(errors.ErrCodeValidation, "invalid request", err).
			WithContext("vm_name", req.Name)
	}

	// 2. Check if VM already exists
	exists, err := uc.vmRepo.Exists(ctx, req.Name)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "failed to check VM existence", err).
			WithContext("vm_name", req.Name)
	}
	if exists {
		return nil, errors.New(errors.ErrCodeConflict, "VM already exists", nil).
			WithContext("vm_name", req.Name)
	}

	// 3. Check available resources
	resources, err := uc.resourceRepo.GetAvailable(ctx)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "failed to check resources", err)
	}

	if !resources.CanAllocate(req.VCPU, req.RAMGB, req.DiskGB) {
		return nil, errors.New(errors.ErrCodeResourceLimit, "insufficient resources", nil).
			WithContext("requested_vcpu", req.VCPU).
			WithContext("requested_ram_gb", req.RAMGB).
			WithContext("requested_disk_gb", req.DiskGB).
			WithContext("available_vcpu", resources.AvailableCPU).
			WithContext("available_ram_gb", resources.AvailableRAMGB).
			WithContext("available_disk_gb", resources.AvailableDiskGB)
	}

	// 4. Get base image
	baseImage, err := uc.storage.GetImage(ctx, req.Template)
	if err != nil {
		return nil, errors.New(errors.ErrCodeStorage, "failed to get base image", err).
			WithContext("template", req.Template)
	}

	// 5. Create disk
	diskPath, err := uc.storage.CreateDisk(ctx, req.Name, baseImage, req.DiskGB)
	if err != nil {
		return nil, errors.New(errors.ErrCodeStorage, "failed to create disk", err).
			WithContext("vm_name", req.Name)
	}

	// 6. Create VM in hypervisor
	vmSpec := &service.VMSpec{
		Name:     req.Name,
		VCPU:     req.VCPU,
		RAMGB:    req.RAMGB,
		DiskGB:   req.DiskGB,
		Template: req.Template,
		DiskPath: diskPath,
	}

	vm, err := uc.hypervisor.CreateVM(ctx, vmSpec)
	if err != nil {
		// Cleanup disk on failure
		_ = uc.storage.DeleteDisk(ctx, req.Name)
		return nil, errors.New(errors.ErrCodeHypervisor, "failed to create VM", err).
			WithContext("vm_name", req.Name)
	}

	// 7. Get IP address
	ip, err := uc.network.GetVMIP(ctx, vm.ID)
	if err != nil {
		uc.logger.Warn("Failed to get VM IP", zap.Error(err))
		ip = "" // Continue without IP
	}
	vm.IP = ip

	// 8. Save VM to repository
	if err := uc.vmRepo.Save(ctx, vm); err != nil {
		uc.logger.Error("Failed to save VM to repository", zap.Error(err))
		// Don't fail the operation, VM is already created
	}

	// 9. Update resource allocation
	resources.Allocate(req.VCPU, req.RAMGB, req.DiskGB)
	if err := uc.resourceRepo.Update(ctx, resources); err != nil {
		uc.logger.Error("Failed to update resources", zap.Error(err))
	}

	uc.logger.Info("VM created successfully",
		zap.String("vm_id", vm.ID),
		zap.String("ip", vm.IP),
	)

	return &dto.CreateVMResponse{
		VMID:      vm.ID,
		IPAddress: vm.IP,
		Status:    string(vm.Status),
	}, nil
}
