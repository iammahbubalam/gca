package http

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"go.uber.org/zap"

	"github.com/iammahbubalam/ghost-agent/internal/domain/service"
)

// HealthServer handles health check endpoints
type HealthServer struct {
	hypervisor service.HypervisorService
	version    string
	startTime  time.Time
	logger     *zap.Logger
}

// NewHealthServer creates a new health server
func NewHealthServer(hypervisor service.HypervisorService, version string, logger *zap.Logger) *HealthServer {
	return &HealthServer{
		hypervisor: hypervisor,
		version:    version,
		startTime:  time.Now(),
		logger:     logger,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Version   string                 `json:"version"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// CheckResult represents a single health check result
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthCheck handles /health endpoint
func (h *HealthServer) HealthCheck(w http.ResponseWriter, r *http.Request) {
	resp := &HealthResponse{
		Status:    "healthy",
		Version:   h.version,
		Timestamp: time.Now(),
		Checks:    make(map[string]CheckResult),
		Metrics:   make(map[string]interface{}),
	}

	// Check Libvirt connection
	if err := h.hypervisor.Ping(r.Context()); err != nil {
		resp.Status = "unhealthy"
		resp.Checks["libvirt"] = CheckResult{
			Status:  "down",
			Message: err.Error(),
		}
	} else {
		resp.Checks["libvirt"] = CheckResult{Status: "up"}
	}

	// Add metrics
	resp.Metrics["uptime_seconds"] = time.Since(h.startTime).Seconds()
	resp.Metrics["goroutines"] = runtime.NumGoroutine()

	// Set HTTP status
	statusCode := http.StatusOK
	if resp.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// LivenessCheck handles /live endpoint
func (h *HealthServer) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ReadinessCheck handles /ready endpoint
func (h *HealthServer) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// Check if Libvirt is ready
	if err := h.hypervisor.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("NOT READY"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("READY"))
}
