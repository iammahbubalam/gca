package server

import (
	"context"
	stderrors "errors"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/iammahbubalam/ghost-agent/internal/application/dto"
	"github.com/iammahbubalam/ghost-agent/internal/application/usecase"
	"github.com/iammahbubalam/ghost-agent/internal/domain/errors"
	"github.com/iammahbubalam/ghost-agent/internal/infrastructure/observability"
	"github.com/iammahbubalam/ghost-agent/pkg/agentpb"
)

// Server implements the gRPC AgentService
type Server struct {
	agentpb.UnimplementedAgentServiceServer
	
	createVMUC    *usecase.CreateVMUseCase
	deleteVMUC    *usecase.DeleteVMUseCase
	startVMUC     *usecase.StartVMUseCase
	stopVMUC      *usecase.StopVMUseCase
	getVMStatusUC *usecase.GetVMStatusUseCase
	listVMsUC     *usecase.ListVMsUseCase
	
	metrics *observability.Metrics
	logger  *zap.Logger
}

// NewServer creates a new gRPC server
func NewServer(
	createVMUC *usecase.CreateVMUseCase,
	deleteVMUC *usecase.DeleteVMUseCase,
	startVMUC *usecase.StartVMUseCase,
	stopVMUC *usecase.StopVMUseCase,
	getVMStatusUC *usecase.GetVMStatusUseCase,
	listVMsUC *usecase.ListVMsUseCase,
	metrics *observability.Metrics,
	logger *zap.Logger,
) *Server {
	return &Server{
		createVMUC:    createVMUC,
		deleteVMUC:    deleteVMUC,
		startVMUC:     startVMUC,
		stopVMUC:      stopVMUC,
		getVMStatusUC: getVMStatusUC,
		listVMsUC:     listVMsUC,
		metrics:       metrics,
		logger:        logger,
	}
}

// CreateVM creates a new VM
func (s *Server) CreateVM(ctx context.Context, req *agentpb.CreateVMRequest) (*agentpb.CreateVMResponse, error) {
	s.logger.Info("gRPC CreateVM request", zap.String("name", req.Name))
	
	// Convert protobuf to DTO
	dtoReq := &dto.CreateVMRequest{
		Name:     req.Name,
		VCPU:     int(req.Vcpu),
		RAMGB:    int(req.RamGb),
		DiskGB:   int(req.DiskGb),
		Template: req.Template,
		Metadata: req.Metadata,
	}
	
	// Execute use case
	resp, err := s.createVMUC.Execute(ctx, dtoReq)
	if err != nil {
		s.metrics.VMOperations.WithLabelValues("create", "error").Inc()
		s.logger.Error("CreateVM failed", zap.Error(err))
		return nil, toGRPCError(err)
	}
	
	s.metrics.VMsCreated.Inc()
	s.metrics.VMsRunning.Inc()
	s.metrics.VMOperations.WithLabelValues("create", "success").Inc()
	
	return &agentpb.CreateVMResponse{
		VmId:      resp.VMID,
		IpAddress: resp.IPAddress,
		Status:    resp.Status,
	}, nil
}

// DeleteVM deletes a VM
func (s *Server) DeleteVM(ctx context.Context, req *agentpb.DeleteVMRequest) (*agentpb.DeleteVMResponse, error) {
	s.logger.Info("gRPC DeleteVM request", zap.String("vm_id", req.VmId))
	
	dtoReq := &dto.DeleteVMRequest{
		VMID: req.VmId,
	}
	
	resp, err := s.deleteVMUC.Execute(ctx, dtoReq)
	if err != nil {
		s.metrics.VMOperations.WithLabelValues("delete", "error").Inc()
		s.logger.Error("DeleteVM failed", zap.Error(err))
		return nil, toGRPCError(err)
	}
	
	s.metrics.VMsDeleted.Inc()
	s.metrics.VMsRunning.Dec()
	s.metrics.VMOperations.WithLabelValues("delete", "success").Inc()
	
	return &agentpb.DeleteVMResponse{
		Success: resp.Success,
	}, nil
}

// StartVM starts a VM
func (s *Server) StartVM(ctx context.Context, req *agentpb.StartVMRequest) (*agentpb.StartVMResponse, error) {
	s.logger.Info("gRPC StartVM request", zap.String("vm_id", req.VmId))
	
	dtoReq := &dto.StartVMRequest{
		VMID: req.VmId,
	}
	
	resp, err := s.startVMUC.Execute(ctx, dtoReq)
	if err != nil {
		s.metrics.VMOperations.WithLabelValues("start", "error").Inc()
		s.logger.Error("StartVM failed", zap.Error(err))
		return nil, toGRPCError(err)
	}
	
	s.metrics.VMsRunning.Inc()
	s.metrics.VMOperations.WithLabelValues("start", "success").Inc()
	
	return &agentpb.StartVMResponse{
		Status: resp.Status,
	}, nil
}

// StopVM stops a VM
func (s *Server) StopVM(ctx context.Context, req *agentpb.StopVMRequest) (*agentpb.StopVMResponse, error) {
	s.logger.Info("gRPC StopVM request", zap.String("vm_id", req.VmId))
	
	dtoReq := &dto.StopVMRequest{
		VMID:  req.VmId,
		Force: req.Force,
	}
	
	resp, err := s.stopVMUC.Execute(ctx, dtoReq)
	if err != nil {
		s.metrics.VMOperations.WithLabelValues("stop", "error").Inc()
		s.logger.Error("StopVM failed", zap.Error(err))
		return nil, toGRPCError(err)
	}
	
	s.metrics.VMsRunning.Dec()
	s.metrics.VMOperations.WithLabelValues("stop", "success").Inc()
	
	return &agentpb.StopVMResponse{
		Status: resp.Status,
	}, nil
}

// GetVMStatus gets VM status
func (s *Server) GetVMStatus(ctx context.Context, req *agentpb.GetVMStatusRequest) (*agentpb.GetVMStatusResponse, error) {
	s.logger.Debug("gRPC GetVMStatus request", zap.String("vm_id", req.VmId))
	
	dtoReq := &dto.GetVMStatusRequest{
		VMID: req.VmId,
	}
	
	resp, err := s.getVMStatusUC.Execute(ctx, dtoReq)
	if err != nil {
		s.logger.Error("GetVMStatus failed", zap.Error(err))
		return nil, toGRPCError(err)
	}
	
	return &agentpb.GetVMStatusResponse{
		VmId:            resp.VMID,
		Name:            resp.Name,
		Status:          resp.Status,
		Vcpu:            int32(resp.VCPU),
		RamGb:           int32(resp.RAMGB),
		DiskGb:          int32(resp.DiskGB),
		IpAddress:       resp.IPAddress,
		UptimeSeconds:   resp.UptimeSeconds,
		CpuUsagePercent: resp.CPUUsagePercent,
		RamUsagePercent: resp.RAMUsagePercent,
	}, nil
}

// ListVMs lists all VMs
func (s *Server) ListVMs(ctx context.Context, req *agentpb.ListVMsRequest) (*agentpb.ListVMsResponse, error) {
	s.logger.Debug("gRPC ListVMs request")
	
	dtoReq := &dto.ListVMsRequest{}
	
	resp, err := s.listVMsUC.Execute(ctx, dtoReq)
	if err != nil {
		s.logger.Error("ListVMs failed", zap.Error(err))
		return nil, toGRPCError(err)
	}
	
	vms := make([]*agentpb.VMInfo, len(resp.VMs))
	for i, vm := range resp.VMs {
		vms[i] = &agentpb.VMInfo{
			VmId:      vm.VMID,
			Name:      vm.Name,
			Status:    vm.Status,
			IpAddress: vm.IPAddress,
		}
	}
	
	return &agentpb.ListVMsResponse{
		Vms: vms,
	}, nil
}

// Helper function to convert domain errors to gRPC errors
func toGRPCError(err error) error {
	var appErr *errors.AppError
	if !stderrors.As(err, &appErr) {
		return status.Error(codes.Internal, "internal server error")
	}
	
	switch appErr.Code {
	case errors.ErrCodeValidation:
		return status.Error(codes.InvalidArgument, appErr.Message)
	case errors.ErrCodeNotFound:
		return status.Error(codes.NotFound, appErr.Message)
	case errors.ErrCodeConflict:
		return status.Error(codes.AlreadyExists, appErr.Message)
	case errors.ErrCodeResourceLimit:
		return status.Error(codes.ResourceExhausted, appErr.Message)
	default:
		return status.Error(codes.Internal, appErr.Message)
	}
}

// RegisterAgentService registers the agent service with gRPC server
func RegisterAgentService(grpcServer *grpc.Server, server *Server) {
	agentpb.RegisterAgentServiceServer(grpcServer, server)
}
