package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the agent
type Metrics struct {
	VMsCreated    prometheus.Counter
	VMsDeleted    prometheus.Counter
	VMsRunning    prometheus.Gauge
	VMOperations  *prometheus.CounterVec
	APICallLatency *prometheus.HistogramVec
	HeartbeatSuccess prometheus.Gauge
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		VMsCreated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "ghost_agent_vms_created_total",
			Help: "Total number of VMs created",
		}),
		VMsDeleted: promauto.NewCounter(prometheus.CounterOpts{
			Name: "ghost_agent_vms_deleted_total",
			Help: "Total number of VMs deleted",
		}),
		VMsRunning: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "ghost_agent_vms_running",
			Help: "Current number of running VMs",
		}),
		VMOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ghost_agent_vm_operations_total",
				Help: "Total VM operations by type and status",
			},
			[]string{"operation", "status"},
		),
		APICallLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ghost_agent_api_call_duration_seconds",
				Help:    "API call latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint"},
		),
		HeartbeatSuccess: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "ghost_agent_heartbeat_success",
			Help: "1 if last heartbeat was successful, 0 otherwise",
		}),
	}
}
