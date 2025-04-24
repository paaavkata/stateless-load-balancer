package balancer

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

type LoadBalancer struct {
	mu                sync.RWMutex
	healthyNodes      []string
	healthCheckCh     chan struct{}
	nodeFilter        string
	tagName           string
	tagValue          string
	asgName           string
	nodeCheckPeriod   time.Duration
	healthCheckPeriod time.Duration
	healthCheckPath   string
	targetPort        int
	region            string
}

func NewLoadBalancer(
	nodeFilter,
	tagName,
	tagValue,
	asgName,
	healthCheckPath,
	region string,
	targetPort,
	nodeCheckPeriod,
	healthCheckPeriod int) *LoadBalancer {
	return &LoadBalancer{
		healthCheckCh:     make(chan struct{}),
		asgName:           asgName,
		nodeFilter:        nodeFilter,
		tagName:           tagName,
		tagValue:          tagValue,
		nodeCheckPeriod:   time.Duration(nodeCheckPeriod) * time.Second,
		healthCheckPeriod: time.Duration(healthCheckPeriod) * time.Second,
		healthCheckPath:   healthCheckPath,
		targetPort:        targetPort,
		region:            region,
	}
}

func (lb *LoadBalancer) StartHealthChecks() {
	if lb.nodeFilter == "asg" {
		lb.healthyNodes = GetASGInstances(lb.asgName, lb.region)
	} else {
		lb.healthyNodes = GetInstancesByTag(lb.tagName, lb.tagValue, lb.region)
	}
	lb.healthyNodes = GetASGInstances(lb.asgName, lb.region)
	nodeCheckTicker := time.NewTicker(lb.nodeCheckPeriod)
	healthCheckTicker := time.NewTicker(lb.healthCheckPeriod)
	for {
		select {
		case <-nodeCheckTicker.C:
			nodes := GetASGInstances(lb.asgName, lb.region)
			lb.mu.Lock()
			lb.healthyNodes = nodes
			lb.mu.Unlock()
		case <-healthCheckTicker.C:
			lb.HealthCheck()
		case <-lb.healthCheckCh:
			return
		}
	}
}

func (lb *LoadBalancer) StartTCPProxy(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to start TCP listener: %v", err)
	}
	defer listener.Close()

	log.Printf("TCP proxy listening on %s", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go lb.handleConnection(conn)
	}
}

// func (lb *LoadBalancer) handleConnection(conn net.Conn) {
// 	defer conn.Close()

// 	lb.mu.RLock()
// 	nodes := lb.healthyNodes
// 	lb.mu.RUnlock()

// 	if len(nodes) == 0 {
// 		log.Println("No healthy nodes available")
// 		return
// 	}

// 	// Forward the request to one of the healthy nodes (simple round-robin approach)
// 	target := fmt.Sprintf("%s:%d", nodes[rand.Intn(len(nodes))], lb.targetPort)

// 	targetConn, err := net.Dial("tcp", target)
// 	if err != nil {
// 		log.Printf("Failed to connect to target %s: %v", target, err)
// 		return
// 	}
// 	defer targetConn.Close()

// 	// Bidirectional copy with error handling
// 	go func() {
// 		_, err := io.Copy(targetConn, conn)
// 		if err != nil {
// 			log.Printf("Error copying from client to target %s: %v", target, err)
// 		}
// 	}()

// 	_, err = io.Copy(conn, targetConn)
// 	if err != nil {
// 		log.Printf("Error copying from target %s to client: %v", target, err)
// 	}
// }

func (lb *LoadBalancer) handleConnection(conn net.Conn) {
	defer conn.Close()

	lb.mu.RLock()
	nodes := lb.healthyNodes
	lb.mu.RUnlock()

	if len(nodes) == 0 {
		log.Println("No healthy nodes available")
		return
	}

	target := fmt.Sprintf("%s:%d", nodes[rand.Intn(len(nodes))], lb.targetPort)
	targetConn, err := net.DialTimeout("tcp", target, 2*time.Second)
	if err != nil {
		log.Printf("Failed to connect to target %s: %v", target, err)
		return
	}
	defer targetConn.Close()

	// Use sync.WaitGroup to wait for both directions
	var wg sync.WaitGroup
	wg.Add(2)

	// Pipe client → backend
	go func() {
		defer wg.Done()
		_, err := io.Copy(targetConn, conn)
		if err != nil {
			log.Printf("[INFO] Client->Target closed %s: %v", target, err)
		}
		targetConn.Close()
	}()

	// Pipe backend → client
	go func() {
		defer wg.Done()
		_, err := io.Copy(conn, targetConn)
		if err != nil {
			log.Printf("[INFO] Target->Client closed %s: %v", target, err)
		}
		conn.Close()
	}()

	wg.Wait()
}

func (lb *LoadBalancer) StopHealthChecks() {
	close(lb.healthCheckCh)
}
