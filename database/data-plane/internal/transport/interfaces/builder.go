package interfaces

import (
	"context"
	"io"
	"time"
)

// IRequestBuilder provides a fluent interface for building HTTP requests.
// This follows the builder pattern to construct requests without executing them.
type IRequestBuilder interface {
	// Host sets the host for the request (e.g., "api.example.com").
	Host(host string) IRequestBuilder

	// Scheme sets the URL scheme (http or https).
	Scheme(scheme string) IRequestBuilder

	// AddPath appends a path segment to the URL.
	AddPath(path string) IRequestBuilder

	// Path sets the complete path, replacing previous paths.
	Path(path string) IRequestBuilder

	// QueryParam adds a single query parameter.
	QueryParam(key, value string) IRequestBuilder

	// QueryParams sets multiple query parameters.
	QueryParams(params map[string]string) IRequestBuilder

	// Header adds a header to the request.
	Header(key, value string) IRequestBuilder

	// Headers sets multiple headers.
	Headers(headers map[string]string) IRequestBuilder

	// ContentType sets the Content-Type header.
	ContentType(contentType string) IRequestBuilder

	// Accept sets the Accept header.
	Accept(accept string) IRequestBuilder

	// Authorization sets the Authorization header.
	Authorization(token string) IRequestBuilder

	// BearerToken sets the Authorization header with Bearer token.
	BearerToken(token string) IRequestBuilder

	// Body sets the request body from an io.Reader.
	Body(body io.Reader) IRequestBuilder

	// BodyBytes sets the request body from bytes.
	BodyBytes(data []byte) IRequestBuilder

	// BodyString sets the request body from string.
	BodyString(data string) IRequestBuilder

	// JSON sets the request body from a JSON-encodable object.
	JSON(v interface{}) IRequestBuilder

	// Timeout sets the request timeout.
	Timeout(timeout time.Duration) IRequestBuilder

	// WithContext sets the context for the request.
	WithContext(ctx context.Context) IRequestBuilder

	// ============= RESILIENCY CONFIGURATION =============

	// WithRetry configures retry behavior with exponential backoff.
	WithRetry(maxAttempts int) IRequestBuilder

	// WithRetryPolicy sets a custom retry policy.
	WithRetryPolicy(policy IRetryPolicy) IRequestBuilder

	// WithCircuitBreaker configures circuit breaker pattern.
	WithCircuitBreaker(failureThreshold int, timeout time.Duration) IRequestBuilder

	// WithRateLimiter configures rate limiting.
	WithRateLimiter(rps float64, burst int) IRequestBuilder

	// WithBulkhead configures bulkhead pattern (concurrency limiting).
	WithBulkhead(maxConcurrency int) IRequestBuilder

	// WithLogging enables request/response logging.
	WithLogging() IRequestBuilder

	// WithMetrics enables metrics collection.
	WithMetrics() IRequestBuilder

	// WithMiddleware adds custom middleware to the request.
	WithMiddleware(middleware IMiddleware) IRequestBuilder

	// ============= HTTP METHODS =============

	// GET sets the HTTP method to GET and builds the request.
	GET() IRequestBuilder

	// POST sets the HTTP method to POST and builds the request.
	POST() IRequestBuilder

	// PUT sets the HTTP method to PUT and builds the request.
	PUT() IRequestBuilder

	// PATCH sets the HTTP method to PATCH and builds the request.
	PATCH() IRequestBuilder

	// DELETE sets the HTTP method to DELETE and builds the request.
	DELETE() IRequestBuilder

	// Method sets a custom HTTP method.
	Method(method string) IRequestBuilder

	// Build constructs and returns the IHTTPRequest without executing it.
	// This allows separation between request construction and execution.
	Build() (IHTTPRequest, error)

	// Sync executes the request synchronously and returns the response.
	// This is a convenience method that builds and executes in one call.
	Sync() (IHTTPResponse, error)

	// Async executes the request asynchronously and returns a channel.
	// The response will be sent to the channel when available.
	Async() <-chan AsyncResult
}
