package client

import (
	"net/http"
	"time"

	"data-plane/internal/transport/interfaces"
	"data-plane/internal/transport/resiliency"
)

// ClientFactory creates HTTP clients and resiliency components.
// This follows the Factory pattern to enable dependency injection and testability.
type ClientFactory interface {
	// Client creation
	CreateHTTPClient(httpClient *http.Client, timeout time.Duration) interfaces.IHTTPClient

	// Resiliency component creation
	CreateRetryPolicy(maxAttempts int) interfaces.IRetryPolicy
	CreateCircuitBreaker(failureThreshold int, timeout time.Duration) interfaces.ICircuitBreaker
	CreateRateLimiter(rps float64, burst int) interfaces.IRateLimiter
	CreateBulkhead(maxConcurrency int) interfaces.IBulkhead
}

// DefaultClientFactory is the default implementation of ClientFactory.
type DefaultClientFactory struct{}

// NewClientFactory creates a new default client factory.
func NewClientFactory() ClientFactory {
	return &DefaultClientFactory{}
}

// CreateHTTPClient creates a basic HTTP client.
func (f *DefaultClientFactory) CreateHTTPClient(httpClient *http.Client, timeout time.Duration) interfaces.IHTTPClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: timeout,
		}
	}
	return &HTTPClient{
		httpClient: httpClient,
		timeout:    timeout,
	}
}

// CreateRetryPolicy creates a retry policy with exponential backoff.
func (f *DefaultClientFactory) CreateRetryPolicy(maxAttempts int) interfaces.IRetryPolicy {
	return resiliency.NewRetryPolicy(maxAttempts)
}

// CreateCircuitBreaker creates a circuit breaker.
func (f *DefaultClientFactory) CreateCircuitBreaker(failureThreshold int, timeout time.Duration) interfaces.ICircuitBreaker {
	return resiliency.NewCircuitBreaker(failureThreshold, timeout)
}

// CreateRateLimiter creates a rate limiter.
func (f *DefaultClientFactory) CreateRateLimiter(rps float64, burst int) interfaces.IRateLimiter {
	return resiliency.NewRateLimiter(rps, burst)
}

// CreateBulkhead creates a bulkhead.
func (f *DefaultClientFactory) CreateBulkhead(maxConcurrency int) interfaces.IBulkhead {
	return resiliency.NewBulkhead(maxConcurrency)
}

// Global default factory instance
var defaultFactory ClientFactory = NewClientFactory()

// SetDefaultFactory sets the global default factory.
// This allows customization of component creation for testing or custom implementations.
func SetDefaultFactory(factory ClientFactory) {
	defaultFactory = factory
}

// GetDefaultFactory returns the global default factory.
func GetDefaultFactory() ClientFactory {
	return defaultFactory
}
