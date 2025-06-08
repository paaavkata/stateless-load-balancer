package balancer

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// ActiveConnections tracks the number of currently active connections
	ActiveConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "load_balancer_active_connections",
		Help: "Number of currently active connections",
	})

	// HealthCheckLatency tracks the latency of health checks
	HealthCheckLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "load_balancer_health_check_latency_seconds",
		Help:    "Latency of health checks in seconds",
		Buckets: prometheus.DefBuckets,
	})

	// FailedConnections tracks the number of failed connection attempts
	FailedConnections = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "load_balancer_failed_connections_total",
		Help: "Total number of failed connection attempts",
	})

	// SuccessfulConnections tracks the number of successful connection attempts
	SuccessfulConnections = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "load_balancer_successful_connections_total",
		Help: "Total number of successful connection attempts",
	})

	// NodeHealthStatus tracks the health status of each node
	NodeHealthStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "load_balancer_node_health_status",
		Help: "Health status of each node (1 for healthy, 0 for unhealthy)",
	}, []string{"node"})
)

func init() {
	prometheus.MustRegister(
		ActiveConnections,
		HealthCheckLatency,
		FailedConnections,
		SuccessfulConnections,
		NodeHealthStatus,
	)
}
