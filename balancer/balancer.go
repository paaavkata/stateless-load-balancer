package balancer

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

type LoadBalancer struct {
	mu                    sync.RWMutex
	healthyNodes          []string
	healthCheckCh         chan struct{}
	nodeFilter            string
	tagName               string
	tagValue              string
	asgName               string
	nodeCheckPeriod       time.Duration
	healthCheckPeriod     time.Duration
	healthCheckPath       string
	healthCheckTimeout    time.Duration
	targetPort            int
	region                string
	circuitBreakers       map[string]*CircuitBreaker
	circuitBreakerTimeout time.Duration
	connectionPool        *ConnectionPool
	listener              net.Listener
	wg                    sync.WaitGroup
	ctx                   context.Context
	cancel                context.CancelFunc
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
	healthCheckPeriod,
	healthCheckTimeout,
	poolSize,
	poolTimeout,
	circuitBreakerThreshold,
	circuitBreakerTimeout int) *LoadBalancer {
	ctx, cancel := context.WithCancel(context.Background())
	return &LoadBalancer{
		healthCheckCh:         make(chan struct{}),
		asgName:               asgName,
		nodeFilter:            nodeFilter,
		tagName:               tagName,
		tagValue:              tagValue,
		nodeCheckPeriod:       time.Duration(nodeCheckPeriod) * time.Second,
		healthCheckPeriod:     time.Duration(healthCheckPeriod) * time.Second,
		healthCheckTimeout:    time.Duration(healthCheckTimeout) * time.Second,
		healthCheckPath:       healthCheckPath,
		targetPort:            targetPort,
		region:                region,
		circuitBreakers:       make(map[string]*CircuitBreaker),
		circuitBreakerTimeout: time.Duration(circuitBreakerTimeout) * time.Minute,
		connectionPool:        NewConnectionPool(poolSize, time.Duration(poolTimeout)*time.Second),
		ctx:                   ctx,
		cancel:                cancel,
	}
}

func (lb *LoadBalancer) StartHealthChecks() {
	var err error
	if lb.nodeFilter == "asg" {
		lb.healthyNodes, err = GetASGInstances(lb.asgName, lb.region)
	} else {
		lb.healthyNodes, err = GetInstancesByTag(lb.tagName, lb.tagValue, lb.region)
	}
	if err != nil {
		log.Printf("Failed to get initial nodes: %v", err)
	}

	nodeCheckTicker := time.NewTicker(lb.nodeCheckPeriod)
	healthCheckTicker := time.NewTicker(lb.healthCheckPeriod)
	defer nodeCheckTicker.Stop()
	defer healthCheckTicker.Stop()

	for {
		select {
		case <-nodeCheckTicker.C:
			var nodes []string
			var err error
			if lb.nodeFilter == "asg" {
				nodes, err = GetASGInstancesWithRetry(lb.asgName, lb.region)
			} else {
				nodes, err = GetInstancesByTagWithRetry(lb.tagName, lb.tagValue, lb.region)
			}
			if err != nil {
				log.Printf("Failed to get nodes: %v", err)
				continue
			}
			lb.mu.Lock()
			lb.healthyNodes = nodes
			lb.mu.Unlock()
		case <-healthCheckTicker.C:
			lb.HealthCheck()
		case <-lb.healthCheckCh:
			return
		case <-lb.ctx.Done():
			return
		}
	}
}

func (lb *LoadBalancer) StartTCPProxy(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start TCP listener: %v", err)
	}
	lb.listener = listener
	defer listener.Close()

	log.Printf("TCP proxy listening on %s", address)

	for {
		select {
		case <-lb.ctx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept connection: %v", err)
				continue
			}

			lb.wg.Add(1)
			go func() {
				defer lb.wg.Done()
				lb.handleConnection(conn)
			}()
		}
	}
}

func (lb *LoadBalancer) handleConnection(conn net.Conn) {
	defer conn.Close()
	ActiveConnections.Inc()
	defer ActiveConnections.Dec()

	lb.mu.RLock()
	nodes := lb.healthyNodes
	lb.mu.RUnlock()

	if len(nodes) == 0 {
		log.Println("No healthy nodes available")
		FailedConnections.Inc()
		return
	}

	// Select a node using round-robin
	target := fmt.Sprintf("%s:%d", nodes[rand.Intn(len(nodes))], lb.targetPort)

	// Check circuit breaker
	lb.mu.RLock()
	cb, exists := lb.circuitBreakers[target]
	if !exists {
		cb = NewCircuitBreaker(5) // 5 failures threshold
		lb.circuitBreakers[target] = cb
	}
	lb.mu.RUnlock()

	if cb.IsOpen() {
		log.Printf("Circuit breaker open for %s", target)
		FailedConnections.Inc()
		return
	}

	// Get connection from pool
	targetConn, err := lb.connectionPool.Get(target)
	if err != nil {
		log.Printf("Failed to connect to target %s: %v", target, err)
		cb.RecordFailure()
		FailedConnections.Inc()
		return
	}

	// Return connection to pool when done
	defer lb.connectionPool.Put(target, targetConn)

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
	SuccessfulConnections.Inc()
	cb.RecordSuccess()
}

func (lb *LoadBalancer) Shutdown(ctx context.Context) error {
	// Stop accepting new connections
	if lb.listener != nil {
		lb.listener.Close()
	}

	// Cancel the context to stop health checks
	lb.cancel()

	// Wait for existing connections to finish
	done := make(chan struct{})
	go func() {
		lb.wg.Wait()
		close(done)
	}()

	// Close the connection pool
	lb.connectionPool.Close()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
