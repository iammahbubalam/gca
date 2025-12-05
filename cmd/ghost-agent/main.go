package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/iammahbubalam/ghost-agent/internal/application/usecase"
	"github.com/iammahbubalam/ghost-agent/internal/domain/entity"
	"github.com/iammahbubalam/ghost-agent/internal/infrastructure/apiclient"
	"github.com/iammahbubalam/ghost-agent/internal/infrastructure/config"
	"github.com/iammahbubalam/ghost-agent/internal/infrastructure/libvirt"
	"github.com/iammahbubalam/ghost-agent/internal/infrastructure/network"
	"github.com/iammahbubalam/ghost-agent/internal/infrastructure/observability"
	"github.com/iammahbubalam/ghost-agent/internal/infrastructure/storage"
	"github.com/iammahbubalam/ghost-agent/internal/presentation/grpc/server"
	httpserver "github.com/iammahbubalam/ghost-agent/internal/presentation/http"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Parse flags
	configPath := "/etc/ghost/agent.yaml"
	if len(os.Args) > 2 && os.Args[1] == "--config" {
		configPath = os.Args[2]
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	logger, err := observability.NewLogger(
		cfg.Logging.Level,
		cfg.Logging.Format,
		cfg.Logging.Output,
		cfg.Logging.File,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Ghost Agent",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
	)

	// Setup metrics
	metrics := observability.NewMetrics()

	// Get Tailscale IP
	tailscaleIP := getTailscaleIP(logger)
	logger.Info("Tailscale IP detected", zap.String("ip", tailscaleIP))

	// Create infrastructure adapters
	logger.Info("Connecting to Libvirt", zap.String("uri", cfg.Libvirt.URI))
	hypervisor, err := libvirt.NewAdapter(cfg.Libvirt.URI, logger)
	if err != nil {
		logger.Fatal("Failed to create Libvirt adapter", zap.Error(err))
	}
	defer hypervisor.Close()

	// Create repositories
	// Use persistent VM repository so state survives PC restarts
	vmRepo, err := storage.NewPersistentVMRepository("/var/lib/ghost/data")
	if err != nil {
		logger.Fatal("Failed to create VM repository", zap.Error(err))
	}
	resourceRepo := storage.NewInMemoryResourceRepository(
		cfg.Resources.ReservedCPU,
		cfg.Resources.ReservedRAMGB,
		cfg.Resources.ReservedDiskGB,
	)
	imageRepo := storage.NewInMemoryImageRepository()

	// Create network adapter
	conn, err := libvirt.NewConnect(cfg.Libvirt.URI)
	if err != nil {
		logger.Fatal("Failed to connect to Libvirt for network", zap.Error(err))
	}
	networkAdapter := network.NewNATAdapter(conn, logger)

	// Create storage adapter
	storageAdapter := storage.NewAdapter(cfg.Libvirt.ImageCache, imageRepo, logger)

	// Create use cases
	createVMUC := usecase.NewCreateVMUseCase(
		hypervisor, networkAdapter, storageAdapter,
		vmRepo, resourceRepo, logger,
	)
	deleteVMUC := usecase.NewDeleteVMUseCase(
		hypervisor, storageAdapter,
		vmRepo, resourceRepo, logger,
	)
	startVMUC := usecase.NewStartVMUseCase(hypervisor, logger)
	stopVMUC := usecase.NewStopVMUseCase(hypervisor, logger)
	getVMStatusUC := usecase.NewGetVMStatusUseCase(
		hypervisor, networkAdapter, vmRepo, logger,
	)
	listVMsUC := usecase.NewListVMsUseCase(hypervisor, networkAdapter, logger)

	// Create Ghost Core API client
	var apiClient *apiclient.Client
	var heartbeatCancel context.CancelFunc

	logger.Info("Connecting to Ghost Core API", zap.String("url", cfg.Agent.APIURL))
	apiClient, err = apiclient.NewClient(cfg.Agent.APIURL, cfg.Agent.Name, tailscaleIP, logger, metrics)
	if err != nil {
		logger.Warn("Failed to connect to Ghost Core API, will retry in heartbeat",
			zap.Error(err),
		)
		// Don't fail startup if API is unavailable
	}

	// Register agent with Ghost Core
	if apiClient != nil {
		resources, _ := resourceRepo.GetAvailable(context.Background())
		agentID, err := apiClient.RegisterAgent(context.Background(), resources, Version)
		if err != nil {
			logger.Warn("Failed to register agent", zap.Error(err))
		} else {
			logger.Info("Agent registered with Ghost Core", zap.String("agent_id", agentID))

			// Start heartbeat in background
			heartbeatCtx, cancel := context.WithCancel(context.Background())
			heartbeatCancel = cancel

			go apiClient.StartHeartbeat(
				heartbeatCtx,
				cfg.Agent.HeartbeatInterval,
				func() *entity.Resource {
					res, _ := resourceRepo.GetAvailable(context.Background())
					return res
				},
				func() []*entity.VM {
					vms, _ := vmRepo.FindAll(context.Background())
					return vms
				},
			)
		}
	}

	// Create gRPC server
	logger.Info("Starting gRPC server", zap.String("addr", cfg.GRPC.ListenAddr))
	grpcServer := server.NewServer(
		createVMUC, deleteVMUC, startVMUC, stopVMUC,
		getVMStatusUC, listVMsUC,
		metrics, logger,
	)

	lis, err := net.Listen("tcp", cfg.GRPC.ListenAddr)
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	grpcSrv := grpc.NewServer()
	server.RegisterAgentService(grpcSrv, grpcServer)

	// Start gRPC server in goroutine
	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	// Start metrics server
	if cfg.Metrics.Enabled {
		logger.Info("Starting metrics server", zap.String("addr", cfg.Metrics.ListenAddr))
		go func() {
			http.Handle(cfg.Metrics.Path, promhttp.Handler())
			if err := http.ListenAndServe(cfg.Metrics.ListenAddr, nil); err != nil {
				logger.Error("Metrics server failed", zap.Error(err))
			}
		}()
	}

	// Start health check server
	logger.Info("Starting health check server", zap.String("addr", cfg.Health.ListenAddr))
	healthServer := httpserver.NewHealthServer(hypervisor, Version, logger)
	go func() {
		http.HandleFunc(cfg.Health.Path, healthServer.HealthCheck)
		http.HandleFunc("/ready", healthServer.ReadinessCheck)
		http.HandleFunc("/live", healthServer.LivenessCheck)
		if err := http.ListenAndServe(cfg.Health.ListenAddr, nil); err != nil {
			logger.Error("Health server failed", zap.Error(err))
		}
	}()

	logger.Info("Ghost Agent started successfully")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigChan

	logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop heartbeat
	if heartbeatCancel != nil {
		heartbeatCancel()
	}

	// Unregister from Ghost Core
	if apiClient != nil {
		logger.Info("Unregistering from Ghost Core")
		if err := apiClient.UnregisterAgent(shutdownCtx); err != nil {
			logger.Warn("Failed to unregister agent", zap.Error(err))
		}
	}

	logger.Info("Stopping gRPC server")
	grpcSrv.GracefulStop()

	logger.Info("Closing Libvirt connection")
	if err := hypervisor.Close(); err != nil {
		logger.Error("Failed to close Libvirt connection", zap.Error(err))
	}

	if apiClient != nil {
		logger.Info("Closing API client connection")
		if err := apiClient.Close(); err != nil {
			logger.Error("Failed to close API client", zap.Error(err))
		}
	}

	logger.Info("Ghost Agent shutdown complete")
}

// getTailscaleIP retrieves the Tailscale IP address
func getTailscaleIP(logger *zap.Logger) string {
	cmd := exec.Command("tailscale", "ip", "-4")
	output, err := cmd.Output()
	if err != nil {
		logger.Warn("Failed to get Tailscale IP, using fallback", zap.Error(err))
		return "unknown"
	}

	ip := strings.TrimSpace(string(output))
	return ip
}
