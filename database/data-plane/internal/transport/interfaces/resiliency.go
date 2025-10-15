package interfaces

import (
	"context"
	"time"
)

// IRetryPolicy defines the interface for retry policies.
type IRetryPolicy interface {
	// ShouldRetry determines if a request should be retried based on the error.
	ShouldRetry(err error, attempt int) bool

	// GetDelay returns the delay duration before the next retry attempt.
	GetDelay(attempt int) time.Duration

	// MaxAttempts returns the maximum number of retry attempts.
	MaxAttempts() int
}

// ICircuitBreaker defines the interface for circuit breaker pattern.
type ICircuitBreaker interface {
	// Execute wraps the request execution with circuit breaker logic.
	Execute(ctx context.Context, fn func() (IHTTPResponse, error)) (IHTTPResponse, error)

	// State returns the current state of the circuit breaker.
	State() CircuitState

	// Reset manually resets the circuit breaker to closed state.
	Reset()

	// Trip manually trips the circuit breaker to open state.
	Trip()
}

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	// StateClosed means requests are flowing normally.
	StateClosed CircuitState = iota

	// StateOpen means the circuit is broken and requests fail fast.
	StateOpen

	// StateHalfOpen means the circuit is testing if the service recovered.
	StateHalfOpen
)

// IMiddleware defines the interface for request/response middleware.
type IMiddleware interface {
	// Before is called before the request is sent.
	Before(ctx context.Context, request IHTTPRequest) (context.Context, error)

	// After is called after the response is received.
	After(ctx context.Context, request IHTTPRequest, response IHTTPResponse, err error) error
}

// IRateLimiter defines the interface for rate limiting.
type IRateLimiter interface {
	// Allow checks if a request is allowed under the rate limit.
	Allow() bool

	// Wait blocks until a request is allowed or context is canceled.
	Wait(ctx context.Context) error
}

// IBulkhead defines the interface for bulkhead pattern (concurrency limiting).
type IBulkhead interface {
	// Execute runs the function with bulkhead protection.
	Execute(ctx context.Context, fn func() (IHTTPResponse, error)) (IHTTPResponse, error)

	// ActiveRequests returns the current number of active requests.
	ActiveRequests() int

	// MaxConcurrency returns the maximum allowed concurrent requests.
	MaxConcurrency() int
}

// IAsyncRequest defines the interface for asynchronous request execution.
type IAsyncRequest interface {
	// Execute sends the request asynchronously and returns a channel for the response.
	Execute() <-chan AsyncResult

	// ExecuteBatch sends multiple requests concurrently.
	ExecuteBatch(requests []IHTTPRequest) <-chan AsyncResult

	// ExecuteWithCallback sends the request and calls the callback when done.
	ExecuteWithCallback(callback func(IHTTPResponse, error))
}

// AsyncResult represents the result of an async request.
type AsyncResult struct {
	Request  IHTTPRequest
	Response IHTTPResponse
	Error    error
	Duration time.Duration
}

// IHealthChecker defines the interface for health checking.
type IHealthChecker interface {
	// Check performs a health check and returns an error if unhealthy.
	Check(ctx context.Context) error

	// IsHealthy returns true if the service is healthy.
	IsHealthy() bool
}
