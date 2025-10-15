package middleware

import (
	"fmt"
	"net/http"
	"time"

	"data-plane/internal/transport/http/models"
	"data-plane/internal/transport/interfaces"
)

// ============= RETRY DECORATOR =============

// RetryDecorator wraps an HTTP client with retry logic.
type RetryDecorator struct {
	wrapped interfaces.IHTTPClient
	policy  interfaces.IRetryPolicy
}

// NewRetryDecorator creates a new retry decorator.
func NewRetryDecorator(wrapped interfaces.IHTTPClient, policy interfaces.IRetryPolicy) interfaces.IHTTPClient {
	return &RetryDecorator{
		wrapped: wrapped,
		policy:  policy,
	}
}

// Send executes the request with retry logic.
func (d *RetryDecorator) Send(request interfaces.IHTTPRequest) (interfaces.IHTTPResponse, error) {
	ctx := request.HTTPRequest().Context()
	var lastErr error

	for attempt := 0; attempt < d.policy.MaxAttempts(); attempt++ {
		// Check context cancellation before each attempt
		select {
		case <-ctx.Done():
			return nil, &models.HTTPError{
				Request: request,
				Message: "request cancelled during retry",
				Err:     ctx.Err(),
			}
		default:
		}

		resp, err := d.wrapped.Send(request)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !d.policy.ShouldRetry(err, attempt) {
			break
		}

		// Context-aware sleep with exponential backoff
		delay := d.policy.GetDelay(attempt)
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return nil, &models.HTTPError{
				Request: request,
				Message: "request cancelled during retry backoff",
				Err:     ctx.Err(),
			}
		}
	}

	return nil, lastErr
}

// SendWithHandler delegates to wrapped client.
func (d *RetryDecorator) SendWithHandler(request interfaces.IHTTPRequest, handler interfaces.IResponseHandler) (interface{}, error) {
	resp, err := d.Send(request)
	if err != nil {
		return nil, err
	}
	return handler.Handle(resp)
}

// SetTimeout sets the timeout on the wrapped client.
func (d *RetryDecorator) SetTimeout(timeout time.Duration) {
	d.wrapped.SetTimeout(timeout)
}

// SetHTTPClient sets the HTTP client on the wrapped client.
func (d *RetryDecorator) SetHTTPClient(client *http.Client) {
	d.wrapped.SetHTTPClient(client)
}

// GetHTTPClient returns the HTTP client from the wrapped client.
func (d *RetryDecorator) GetHTTPClient() *http.Client {
	return d.wrapped.GetHTTPClient()
}

// ============= CIRCUIT BREAKER DECORATOR =============

// CircuitBreakerDecorator wraps an HTTP client with circuit breaker logic.
type CircuitBreakerDecorator struct {
	wrapped        interfaces.IHTTPClient
	circuitBreaker interfaces.ICircuitBreaker
}

// NewCircuitBreakerDecorator creates a new circuit breaker decorator.
func NewCircuitBreakerDecorator(wrapped interfaces.IHTTPClient, circuitBreaker interfaces.ICircuitBreaker) interfaces.IHTTPClient {
	return &CircuitBreakerDecorator{
		wrapped:        wrapped,
		circuitBreaker: circuitBreaker,
	}
}

// Send executes the request with circuit breaker protection.
func (d *CircuitBreakerDecorator) Send(request interfaces.IHTTPRequest) (interfaces.IHTTPResponse, error) {
	ctx := request.HTTPRequest().Context()

	return d.circuitBreaker.Execute(ctx, func() (interfaces.IHTTPResponse, error) {
		return d.wrapped.Send(request)
	})
}

// SendWithHandler delegates to wrapped client.
func (d *CircuitBreakerDecorator) SendWithHandler(request interfaces.IHTTPRequest, handler interfaces.IResponseHandler) (interface{}, error) {
	resp, err := d.Send(request)
	if err != nil {
		return nil, err
	}
	return handler.Handle(resp)
}

// SetTimeout sets the timeout on the wrapped client.
func (d *CircuitBreakerDecorator) SetTimeout(timeout time.Duration) {
	d.wrapped.SetTimeout(timeout)
}

// SetHTTPClient sets the HTTP client on the wrapped client.
func (d *CircuitBreakerDecorator) SetHTTPClient(client *http.Client) {
	d.wrapped.SetHTTPClient(client)
}

// GetHTTPClient returns the HTTP client from the wrapped client.
func (d *CircuitBreakerDecorator) GetHTTPClient() *http.Client {
	return d.wrapped.GetHTTPClient()
}

// ============= RATE LIMITER DECORATOR =============

// RateLimiterDecorator wraps an HTTP client with rate limiting.
type RateLimiterDecorator struct {
	wrapped     interfaces.IHTTPClient
	rateLimiter interfaces.IRateLimiter
}

// NewRateLimiterDecorator creates a new rate limiter decorator.
func NewRateLimiterDecorator(wrapped interfaces.IHTTPClient, rateLimiter interfaces.IRateLimiter) interfaces.IHTTPClient {
	return &RateLimiterDecorator{
		wrapped:     wrapped,
		rateLimiter: rateLimiter,
	}
}

// Send executes the request with rate limiting.
func (d *RateLimiterDecorator) Send(request interfaces.IHTTPRequest) (interfaces.IHTTPResponse, error) {
	ctx := request.HTTPRequest().Context()

	// Wait for rate limiter
	select {
	case <-ctx.Done():
		return nil, &models.HTTPError{
			Request: request,
			Message: "request cancelled before rate limiting",
			Err:     ctx.Err(),
		}
	default:
		if err := d.rateLimiter.Wait(ctx); err != nil {
			return nil, &models.HTTPError{
				Request: request,
				Message: "rate limit exceeded",
				Err:     err,
			}
		}
	}

	return d.wrapped.Send(request)
}

// SendWithHandler delegates to wrapped client.
func (d *RateLimiterDecorator) SendWithHandler(request interfaces.IHTTPRequest, handler interfaces.IResponseHandler) (interface{}, error) {
	resp, err := d.Send(request)
	if err != nil {
		return nil, err
	}
	return handler.Handle(resp)
}

// SetTimeout sets the timeout on the wrapped client.
func (d *RateLimiterDecorator) SetTimeout(timeout time.Duration) {
	d.wrapped.SetTimeout(timeout)
}

// SetHTTPClient sets the HTTP client on the wrapped client.
func (d *RateLimiterDecorator) SetHTTPClient(client *http.Client) {
	d.wrapped.SetHTTPClient(client)
}

// GetHTTPClient returns the HTTP client from the wrapped client.
func (d *RateLimiterDecorator) GetHTTPClient() *http.Client {
	return d.wrapped.GetHTTPClient()
}

// ============= BULKHEAD DECORATOR =============

// BulkheadDecorator wraps an HTTP client with bulkhead pattern.
type BulkheadDecorator struct {
	wrapped  interfaces.IHTTPClient
	bulkhead interfaces.IBulkhead
}

// NewBulkheadDecorator creates a new bulkhead decorator.
func NewBulkheadDecorator(wrapped interfaces.IHTTPClient, bulkhead interfaces.IBulkhead) interfaces.IHTTPClient {
	return &BulkheadDecorator{
		wrapped:  wrapped,
		bulkhead: bulkhead,
	}
}

// Send executes the request with bulkhead protection.
func (d *BulkheadDecorator) Send(request interfaces.IHTTPRequest) (interfaces.IHTTPResponse, error) {
	ctx := request.HTTPRequest().Context()

	return d.bulkhead.Execute(ctx, func() (interfaces.IHTTPResponse, error) {
		return d.wrapped.Send(request)
	})
}

// SendWithHandler delegates to wrapped client.
func (d *BulkheadDecorator) SendWithHandler(request interfaces.IHTTPRequest, handler interfaces.IResponseHandler) (interface{}, error) {
	resp, err := d.Send(request)
	if err != nil {
		return nil, err
	}
	return handler.Handle(resp)
}

// SetTimeout sets the timeout on the wrapped client.
func (d *BulkheadDecorator) SetTimeout(timeout time.Duration) {
	d.wrapped.SetTimeout(timeout)
}

// SetHTTPClient sets the HTTP client on the wrapped client.
func (d *BulkheadDecorator) SetHTTPClient(client *http.Client) {
	d.wrapped.SetHTTPClient(client)
}

// GetHTTPClient returns the HTTP client from the wrapped client.
func (d *BulkheadDecorator) GetHTTPClient() *http.Client {
	return d.wrapped.GetHTTPClient()
}

// ============= LOGGING DECORATOR =============

// LoggingDecorator wraps an HTTP client with logging.
type LoggingDecorator struct {
	wrapped interfaces.IHTTPClient
}

// NewLoggingDecorator creates a new logging decorator.
func NewLoggingDecorator(wrapped interfaces.IHTTPClient) interfaces.IHTTPClient {
	return &LoggingDecorator{
		wrapped: wrapped,
	}
}

// Send executes the request with logging.
func (d *LoggingDecorator) Send(request interfaces.IHTTPRequest) (interfaces.IHTTPResponse, error) {
	fmt.Printf("→ %s %s\n", request.Method(), request.URL())

	startTime := time.Now()
	resp, err := d.wrapped.Send(request)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("← %s %s failed in %v: %v\n", request.Method(), request.URL(), duration, err)
	} else {
		fmt.Printf("← %s %s returned %d in %v\n", request.Method(), request.URL(), resp.StatusCode(), duration)
	}

	return resp, err
}

// SendWithHandler delegates to wrapped client.
func (d *LoggingDecorator) SendWithHandler(request interfaces.IHTTPRequest, handler interfaces.IResponseHandler) (interface{}, error) {
	resp, err := d.Send(request)
	if err != nil {
		return nil, err
	}
	return handler.Handle(resp)
}

// SetTimeout sets the timeout on the wrapped client.
func (d *LoggingDecorator) SetTimeout(timeout time.Duration) {
	d.wrapped.SetTimeout(timeout)
}

// SetHTTPClient sets the HTTP client on the wrapped client.
func (d *LoggingDecorator) SetHTTPClient(client *http.Client) {
	d.wrapped.SetHTTPClient(client)
}

// GetHTTPClient returns the HTTP client from the wrapped client.
func (d *LoggingDecorator) GetHTTPClient() *http.Client {
	return d.wrapped.GetHTTPClient()
}

// ============= METRICS DECORATOR =============

// MetricsDecorator wraps an HTTP client with metrics collection.
type MetricsDecorator struct {
	wrapped interfaces.IHTTPClient
}

// NewMetricsDecorator creates a new metrics decorator.
func NewMetricsDecorator(wrapped interfaces.IHTTPClient) interfaces.IHTTPClient {
	return &MetricsDecorator{
		wrapped: wrapped,
	}
}

// Send executes the request with metrics collection.
func (d *MetricsDecorator) Send(request interfaces.IHTTPRequest) (interfaces.IHTTPResponse, error) {
	startTime := time.Now()
	resp, err := d.wrapped.Send(request)
	duration := time.Since(startTime)

	// Record metrics (placeholder for actual metrics implementation)
	fmt.Printf("[METRICS] method=%s, duration=%v, error=%v\n", request.Method(), duration, err != nil)

	return resp, err
}

// SendWithHandler delegates to wrapped client.
func (d *MetricsDecorator) SendWithHandler(request interfaces.IHTTPRequest, handler interfaces.IResponseHandler) (interface{}, error) {
	resp, err := d.Send(request)
	if err != nil {
		return nil, err
	}
	return handler.Handle(resp)
}

// SetTimeout sets the timeout on the wrapped client.
func (d *MetricsDecorator) SetTimeout(timeout time.Duration) {
	d.wrapped.SetTimeout(timeout)
}

// SetHTTPClient sets the HTTP client on the wrapped client.
func (d *MetricsDecorator) SetHTTPClient(client *http.Client) {
	d.wrapped.SetHTTPClient(client)
}

// GetHTTPClient returns the HTTP client from the wrapped client.
func (d *MetricsDecorator) GetHTTPClient() *http.Client {
	return d.wrapped.GetHTTPClient()
}

// ============= MIDDLEWARE DECORATOR =============

// MiddlewareDecorator wraps an HTTP client with middleware execution.
type MiddlewareDecorator struct {
	wrapped     interfaces.IHTTPClient
	middlewares []interfaces.IMiddleware
}

// NewMiddlewareDecorator creates a new middleware decorator.
func NewMiddlewareDecorator(wrapped interfaces.IHTTPClient, middlewares []interfaces.IMiddleware) interfaces.IHTTPClient {
	return &MiddlewareDecorator{
		wrapped:     wrapped,
		middlewares: middlewares,
	}
}

// Send executes the request with middleware.
func (d *MiddlewareDecorator) Send(request interfaces.IHTTPRequest) (interfaces.IHTTPResponse, error) {
	ctx := request.HTTPRequest().Context()

	// Apply middleware Before() hooks
	for _, mw := range d.middlewares {
		newCtx, err := mw.Before(ctx, request)
		if err != nil {
			return nil, err
		}
		ctx = newCtx
	}

	// Execute request
	resp, err := d.wrapped.Send(request)

	// Apply middleware After() hooks
	for _, mw := range d.middlewares {
		if afterErr := mw.After(ctx, request, resp, err); afterErr != nil {
			fmt.Printf("Middleware After() error: %v\n", afterErr)
		}
	}

	return resp, err
}

// SendWithHandler delegates to wrapped client.
func (d *MiddlewareDecorator) SendWithHandler(request interfaces.IHTTPRequest, handler interfaces.IResponseHandler) (interface{}, error) {
	resp, err := d.Send(request)
	if err != nil {
		return nil, err
	}
	return handler.Handle(resp)
}

// SetTimeout sets the timeout on the wrapped client.
func (d *MiddlewareDecorator) SetTimeout(timeout time.Duration) {
	d.wrapped.SetTimeout(timeout)
}

// SetHTTPClient sets the HTTP client on the wrapped client.
func (d *MiddlewareDecorator) SetHTTPClient(client *http.Client) {
	d.wrapped.SetHTTPClient(client)
}

// GetHTTPClient returns the HTTP client from the wrapped client.
func (d *MiddlewareDecorator) GetHTTPClient() *http.Client {
	return d.wrapped.GetHTTPClient()
}
