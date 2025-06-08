package balancer

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"
)

func checkNodeHealth(targetPort int, node, path string, timeout time.Duration) bool {
	endpoint := fmt.Sprintf("http://%s:%d%s", node, targetPort, path)

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(endpoint)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Node %s unhealthy: %v", endpoint, err)
		return false
	}
	defer resp.Body.Close()
	return true
}

func (lb *LoadBalancer) HealthCheck() {
	lb.mu.RLock()
	nodes := lb.healthyNodes
	lb.mu.RUnlock()

	var healthyNodes []string
	for _, node := range nodes {
		start := time.Now()
		healthy := checkNodeHealth(lb.targetPort, node, lb.healthCheckPath, lb.healthCheckTimeout)
		HealthCheckLatency.Observe(time.Since(start).Seconds())

		if healthy {
			healthyNodes = append(healthyNodes, node)
			NodeHealthStatus.WithLabelValues(node).Set(1)
		} else {
			NodeHealthStatus.WithLabelValues(node).Set(0)
		}
	}

	if !reflect.DeepEqual(lb.healthyNodes, healthyNodes) && len(healthyNodes) > 0 {
		lb.mu.Lock()
		lb.healthyNodes = healthyNodes
		lb.mu.Unlock()
	}
}
