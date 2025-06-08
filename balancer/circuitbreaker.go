package balancer

import (
	"sync"
	"time"
)

// CircuitBreaker implements the circuit breaker pattern for handling failing nodes
type CircuitBreaker struct {
	failures    int
	threshold   int
	lastFailure time.Time
	mu          sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker with the specified failure threshold
func NewCircuitBreaker(threshold int) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
	}
}

// IsOpen returns true if the circuit breaker is open (too many failures)
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures >= cb.threshold && time.Since(cb.lastFailure) < time.Minute*5
}

// RecordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
}

// RecordSuccess resets the failure count
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
}

// GetFailureCount returns the current number of failures
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}
