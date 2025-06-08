package balancer

import (
	"net"
	"sync"
	"time"
)

// ConnectionPool manages a pool of TCP connections
type ConnectionPool struct {
	pool    map[string][]net.Conn
	mu      sync.RWMutex
	maxSize int
	timeout time.Duration
}

// NewConnectionPool creates a new connection pool with the specified maximum size and timeout
func NewConnectionPool(maxSize int, timeout time.Duration) *ConnectionPool {
	return &ConnectionPool{
		pool:    make(map[string][]net.Conn),
		maxSize: maxSize,
		timeout: timeout,
	}
}

// Get returns a connection from the pool or creates a new one
func (p *ConnectionPool) Get(target string) (net.Conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Try to get a connection from the pool
	if conns, ok := p.pool[target]; ok && len(conns) > 0 {
		conn := conns[len(conns)-1]
		p.pool[target] = conns[:len(conns)-1]
		return conn, nil
	}

	// Create a new connection
	return net.DialTimeout("tcp", target, p.timeout)
}

// Put returns a connection to the pool
func (p *ConnectionPool) Put(target string, conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if the connection is still valid
	if err := conn.SetReadDeadline(time.Now().Add(time.Millisecond)); err != nil {
		conn.Close()
		return
	}
	conn.SetReadDeadline(time.Time{})

	// Add the connection to the pool if there's space
	if conns, ok := p.pool[target]; ok {
		if len(conns) < p.maxSize {
			p.pool[target] = append(conns, conn)
			return
		}
	} else {
		p.pool[target] = []net.Conn{conn}
		return
	}

	// If we get here, the pool is full
	conn.Close()
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, conns := range p.pool {
		for _, conn := range conns {
			conn.Close()
		}
	}
	p.pool = make(map[string][]net.Conn)
}
