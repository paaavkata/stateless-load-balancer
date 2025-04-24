package balancer

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
)

func (lb *LoadBalancer) HealthCheck() {
	lb.mu.RLock()
	nodes := lb.healthyNodes
	lb.mu.RUnlock()

	var healthyNodes []string
	for _, node := range nodes {
		if checkNodeHealth(lb.targetPort, node, lb.healthCheckPath) {
			healthyNodes = append(healthyNodes, node)
		}
	}

	if !reflect.DeepEqual(lb.healthyNodes, healthyNodes) && len(healthyNodes) > 0 {
		lb.mu.Lock()
		lb.healthyNodes = healthyNodes
		lb.mu.Unlock()
	}
}

func checkNodeHealth(targetPort int, node, path string) bool {
	endpoint := fmt.Sprintf("http://%s:%d%s", node, targetPort, path)

	resp, err := http.Get(endpoint)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Node %s unhealthy", endpoint)
		return false
	}
	return true
}
