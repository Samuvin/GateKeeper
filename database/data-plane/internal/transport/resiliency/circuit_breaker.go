package resiliency

import (
	"context"
	"errors"
	"sync"
	"time"

	"data-plane/internal/transport/interfaces"
)

// CircuitBreaker implements the circuit breaker pattern to prevent cascading failures.
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            interfaces.CircuitState
	failureCount     int
	successCount     int
	lastFailureTime  time.Time
	lastSuccessTime  time.Time
	failureThreshold int           // Number of failures before opening
	successThreshold int           // Number of successes to close from half-open
	timeout          time.Duration // Time to wait before trying half-open
}

// Ensure CircuitBreaker implements ICircuitBreaker interface
var _ interfaces.ICircuitBreaker = (*CircuitBreaker)(nil)

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(failureThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            interfaces.StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: 2, // Default: 2 successful requests to close
		timeout:          timeout,
	}
}

// Execute wraps request execution with circuit breaker logic.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() (interfaces.IHTTPResponse, error)) (interfaces.IHTTPResponse, error) {
	// Check if circuit allows execution
	if !cb.canExecute() {
		return nil, errors.New("circuit breaker is open: request rejected")
	}

	// Execute the request
	resp, err := fn()

	// Record the result
	cb.recordResult(err)

	return resp, err
}

// canExecute checks if the circuit breaker allows request execution.
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case interfaces.StateClosed:
		// Closed state: allow all requests
		return true

	case interfaces.StateOpen:
		// Check if timeout has passed to transition to half-open
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = interfaces.StateHalfOpen
			cb.successCount = 0
			return true
		}
		// Still in timeout period, reject request
		return false

	case interfaces.StateHalfOpen:
		// Half-open: allow limited requests to test
		return true

	default:
		return false
	}
}

// recordResult records the result of a request execution.
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onFailure handles a failed request.
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case interfaces.StateClosed:
		// Check if we've hit the failure threshold
		if cb.failureCount >= cb.failureThreshold {
			cb.state = interfaces.StateOpen
			cb.failureCount = 0
		}

	case interfaces.StateHalfOpen:
		// Any failure in half-open immediately opens the circuit
		cb.state = interfaces.StateOpen
		cb.failureCount = 0
		cb.successCount = 0
	}
}

// onSuccess handles a successful request.
func (cb *CircuitBreaker) onSuccess() {
	cb.lastSuccessTime = time.Now()

	switch cb.state {
	case interfaces.StateClosed:
		// Reset failure count on success
		cb.failureCount = 0

	case interfaces.StateHalfOpen:
		cb.successCount++
		// Check if we've hit the success threshold to close
		if cb.successCount >= cb.successThreshold {
			cb.state = interfaces.StateClosed
			cb.failureCount = 0
			cb.successCount = 0
		}
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() interfaces.CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset manually resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = interfaces.StateClosed
	cb.failureCount = 0
	cb.successCount = 0
}

// Trip manually trips the circuit breaker to open state.
func (cb *CircuitBreaker) Trip() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = interfaces.StateOpen
	cb.lastFailureTime = time.Now()
	cb.failureCount = 0
	cb.successCount = 0
}

// GetMetrics returns current circuit breaker metrics.
func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerMetrics{
		State:           cb.state,
		FailureCount:    cb.failureCount,
		SuccessCount:    cb.successCount,
		LastFailureTime: cb.lastFailureTime,
		LastSuccessTime: cb.lastSuccessTime,
	}
}

// CircuitBreakerMetrics contains circuit breaker statistics.
type CircuitBreakerMetrics struct {
	State           interfaces.CircuitState
	FailureCount    int
	SuccessCount    int
	LastFailureTime time.Time
	LastSuccessTime time.Time
}
