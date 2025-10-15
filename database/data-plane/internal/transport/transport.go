package transport

// Unified Transport API - Protocol Agnostic
// This package provides a consistent interface for HTTP, gRPC, HTTPS, and other protocols.

import (
	"time"

	"data-plane/internal/transport/http/builder"
	"data-plane/internal/transport/http/client"
	"data-plane/internal/transport/http/handler"
	"data-plane/internal/transport/http/models"
	"data-plane/internal/transport/interfaces"
	"data-plane/internal/transport/middleware"
	"data-plane/internal/transport/resiliency"
)

// ============= HTTP PROTOCOL =============

// HTTP provides convenient access to HTTP client components
type HTTP struct{}

// NewHTTPClient creates a new HTTP client with default configuration
func (HTTP) NewClient() interfaces.IHTTPClient {
	return client.NewHTTPClient()
}

// NewHTTPClientWithTimeout creates a new HTTP client with the specified timeout
func (HTTP) NewClientWithTimeout(timeout time.Duration) interfaces.IHTTPClient {
	return client.NewHTTPClientWithTimeout(timeout)
}

// NewBuilder creates a new HTTP request builder
func (HTTP) NewBuilder() interfaces.IRequestBuilder {
	return builder.NewBuilder()
}

// NewResponseHandler creates a new HTTP response handler builder
func (HTTP) NewResponseHandler() *handler.ResponseHandlerBuilder {
	return handler.NewResponseHandler()
}

// ============= RESILIENCY (Protocol-Agnostic) =============

// Resiliency provides resiliency patterns that work with any protocol
type Resiliency struct{}

// NewRetryPolicy creates a retry policy with exponential backoff
func (Resiliency) NewRetryPolicy(maxAttempts int) *resiliency.RetryPolicy {
	return resiliency.NewRetryPolicy(maxAttempts)
}

// NewCircuitBreaker creates a circuit breaker
func (Resiliency) NewCircuitBreaker(failureThreshold int, timeout time.Duration) *resiliency.CircuitBreaker {
	return resiliency.NewCircuitBreaker(failureThreshold, timeout)
}

// NewRateLimiter creates a rate limiter
func (Resiliency) NewRateLimiter(rate float64, burst int) *resiliency.RateLimiter {
	return resiliency.NewRateLimiter(rate, burst)
}

// NewBulkhead creates a bulkhead with the specified max concurrency
func (Resiliency) NewBulkhead(maxConcurrency int) *resiliency.Bulkhead {
	return resiliency.NewBulkhead(maxConcurrency)
}

// ============= MIDDLEWARE (Protocol-Agnostic) =============

// Middleware provides middleware components that work with any protocol
type Middleware struct{}

// NewLoggingMiddleware creates a logging middleware
func (Middleware) NewLoggingMiddleware() *middleware.LoggingMiddleware {
	return middleware.NewLoggingMiddleware(nil)
}

// NewMetricsMiddleware creates a metrics middleware
func (Middleware) NewMetricsMiddleware() *middleware.MetricsMiddleware {
	return middleware.NewMetricsMiddleware()
}

// NewAsyncRequest creates an async request handler
func (Middleware) NewAsyncRequest(client interfaces.IHTTPClient) *middleware.AsyncRequest {
	return middleware.NewAsyncRequest(client)
}

// ============= TYPE ALIASES FOR CONVENIENCE =============

// HTTP Models
type (
	HTTPRequest  = models.Request
	HTTPResponse = models.Response
	HTTPError    = models.HTTPError
)

// HTTP Client types
type (
	HTTPClient    = client.HTTPClient
	ClientFactory = client.ClientFactory
)

// Builder types
type (
	RequestBuilder = builder.RequestBuilder
)

// Handler types
type (
	ResponseHandler = handler.ResponseHandler
	JSONMarshaller  = handler.JSONMarshaller
)

// Resiliency types (Protocol-agnostic)
type (
	RetryPolicy    = resiliency.RetryPolicy
	CircuitBreaker = resiliency.CircuitBreaker
	RateLimiter    = resiliency.RateLimiter
	Bulkhead       = resiliency.Bulkhead
)

// Middleware types (Protocol-agnostic)
type (
	LoggingMiddleware = middleware.LoggingMiddleware
	MetricsMiddleware = middleware.MetricsMiddleware
	AuthMiddleware    = middleware.AuthMiddleware
	TracingMiddleware = middleware.TracingMiddleware
	AsyncRequest      = middleware.AsyncRequest
)

// ============= CONVENIENT GLOBALS =============

var (
	// HTTPTransport provides HTTP-specific functions
	HTTPTransport = HTTP{}

	// ResiliencyFeatures provides protocol-agnostic resiliency
	ResiliencyFeatures = Resiliency{}

	// MiddlewareFeatures provides protocol-agnostic middleware
	MiddlewareFeatures = Middleware{}
)

// ============= CONVENIENCE FUNCTIONS (Backward Compatible) =============

// NewHTTPClient creates a new HTTP client
func NewHTTPClient() interfaces.IHTTPClient {
	return HTTPTransport.NewClient()
}

// NewHTTPClientWithTimeout creates an HTTP client with timeout
func NewHTTPClientWithTimeout(timeout time.Duration) interfaces.IHTTPClient {
	return HTTPTransport.NewClientWithTimeout(timeout)
}

// NewHTTPBuilder creates a new HTTP request builder
func NewHTTPBuilder() interfaces.IRequestBuilder {
	return HTTPTransport.NewBuilder()
}

// NewHTTPResponseHandler creates a new HTTP response handler
func NewHTTPResponseHandler() *handler.ResponseHandlerBuilder {
	return HTTPTransport.NewResponseHandler()
}

// NewRetryPolicy creates a retry policy
func NewRetryPolicy(maxAttempts int) *resiliency.RetryPolicy {
	return ResiliencyFeatures.NewRetryPolicy(maxAttempts)
}

// NewCircuitBreaker creates a circuit breaker
func NewCircuitBreaker(failureThreshold int, timeout time.Duration) *resiliency.CircuitBreaker {
	return ResiliencyFeatures.NewCircuitBreaker(failureThreshold, timeout)
}

// NewRateLimiter creates a rate limiter
func NewRateLimiter(rate float64, burst int) *resiliency.RateLimiter {
	return ResiliencyFeatures.NewRateLimiter(rate, burst)
}

// NewBulkhead creates a bulkhead
func NewBulkhead(maxConcurrency int) *resiliency.Bulkhead {
	return ResiliencyFeatures.NewBulkhead(maxConcurrency)
}

// GetDefaultFactory returns the global default client factory
func GetDefaultFactory() client.ClientFactory {
	return client.GetDefaultFactory()
}

// SetDefaultFactory sets the global default client factory
func SetDefaultFactory(factory client.ClientFactory) {
	client.SetDefaultFactory(factory)
}
