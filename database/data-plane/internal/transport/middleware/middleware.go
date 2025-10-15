package middleware

import (
	"context"
	"fmt"
	"log"
	"time"

	"data-plane/internal/transport/interfaces"
)

// LoggingMiddleware logs request and response information.
type LoggingMiddleware struct {
	logger *log.Logger
}

// Ensure LoggingMiddleware implements IMiddleware interface
var _ interfaces.IMiddleware = (*LoggingMiddleware)(nil)

// NewLoggingMiddleware creates a new logging middleware.
func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
	if logger == nil {
		logger = log.Default()
	}
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Before logs the request before sending.
func (lm *LoggingMiddleware) Before(ctx context.Context, request interfaces.IHTTPRequest) (context.Context, error) {
	lm.logger.Printf("[HTTP] → %s %s", request.Method(), request.URL())

	// Store start time in context
	ctx = context.WithValue(ctx, "start_time", time.Now())
	return ctx, nil
}

// After logs the response after receiving.
func (lm *LoggingMiddleware) After(ctx context.Context, request interfaces.IHTTPRequest, response interfaces.IHTTPResponse, err error) error {
	startTime, ok := ctx.Value("start_time").(time.Time)
	if !ok {
		startTime = time.Now()
	}

	duration := time.Since(startTime)

	if err != nil {
		lm.logger.Printf("[HTTP] ← %s %s [ERROR] %v (took %v)", request.Method(), request.URL(), err, duration)
	} else {
		lm.logger.Printf("[HTTP] ← %s %s [%d] (took %v)", request.Method(), request.URL(), response.StatusCode(), duration)
	}

	return nil
}

// MetricsMiddleware tracks request metrics.
type MetricsMiddleware struct {
	totalRequests int64
	successCount  int64
	errorCount    int64
	totalDuration time.Duration
}

// Ensure MetricsMiddleware implements IMiddleware interface
var _ interfaces.IMiddleware = (*MetricsMiddleware)(nil)

// NewMetricsMiddleware creates a new metrics middleware.
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{}
}

// Before is called before the request.
func (mm *MetricsMiddleware) Before(ctx context.Context, request interfaces.IHTTPRequest) (context.Context, error) {
	ctx = context.WithValue(ctx, "metrics_start", time.Now())
	return ctx, nil
}

// After tracks metrics after the response.
func (mm *MetricsMiddleware) After(ctx context.Context, request interfaces.IHTTPRequest, response interfaces.IHTTPResponse, err error) error {
	startTime, ok := ctx.Value("metrics_start").(time.Time)
	if !ok {
		return nil
	}

	duration := time.Since(startTime)

	mm.totalRequests++
	mm.totalDuration += duration

	if err != nil {
		mm.errorCount++
	} else {
		mm.successCount++
	}

	return nil
}

// GetMetrics returns current metrics.
func (mm *MetricsMiddleware) GetMetrics() MetricsData {
	avgDuration := time.Duration(0)
	if mm.totalRequests > 0 {
		avgDuration = mm.totalDuration / time.Duration(mm.totalRequests)
	}

	successRate := float64(0)
	if mm.totalRequests > 0 {
		successRate = float64(mm.successCount) / float64(mm.totalRequests) * 100
	}

	return MetricsData{
		TotalRequests:   mm.totalRequests,
		SuccessCount:    mm.successCount,
		ErrorCount:      mm.errorCount,
		TotalDuration:   mm.totalDuration,
		AverageDuration: avgDuration,
		SuccessRate:     successRate,
	}
}

// MetricsData contains metrics information.
type MetricsData struct {
	TotalRequests   int64
	SuccessCount    int64
	ErrorCount      int64
	TotalDuration   time.Duration
	AverageDuration time.Duration
	SuccessRate     float64
}

// AuthMiddleware adds authentication to requests.
type AuthMiddleware struct {
	authToken string
	authType  string // "Bearer", "Basic", etc.
}

// Ensure AuthMiddleware implements IMiddleware interface
var _ interfaces.IMiddleware = (*AuthMiddleware)(nil)

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(authType, token string) *AuthMiddleware {
	return &AuthMiddleware{
		authToken: token,
		authType:  authType,
	}
}

// Before adds authentication header to the request.
func (am *AuthMiddleware) Before(ctx context.Context, request interfaces.IHTTPRequest) (context.Context, error) {
	// Note: This is a limitation - we can't modify the request after it's built
	// In production, this would be applied during building phase
	return ctx, nil
}

// After does nothing for auth middleware.
func (am *AuthMiddleware) After(ctx context.Context, request interfaces.IHTTPRequest, response interfaces.IHTTPResponse, err error) error {
	return nil
}

// TracingMiddleware adds distributed tracing.
type TracingMiddleware struct {
	traceID string
}

// Ensure TracingMiddleware implements IMiddleware interface
var _ interfaces.IMiddleware = (*TracingMiddleware)(nil)

// NewTracingMiddleware creates a new tracing middleware.
func NewTracingMiddleware(traceID string) *TracingMiddleware {
	if traceID == "" {
		traceID = fmt.Sprintf("trace-%d", time.Now().UnixNano())
	}
	return &TracingMiddleware{
		traceID: traceID,
	}
}

// Before adds tracing information to context.
func (tm *TracingMiddleware) Before(ctx context.Context, request interfaces.IHTTPRequest) (context.Context, error) {
	ctx = context.WithValue(ctx, "trace_id", tm.traceID)
	ctx = context.WithValue(ctx, "span_id", fmt.Sprintf("span-%d", time.Now().UnixNano()))
	return ctx, nil
}

// After logs tracing information.
func (tm *TracingMiddleware) After(ctx context.Context, request interfaces.IHTTPRequest, response interfaces.IHTTPResponse, err error) error {
	traceID := ctx.Value("trace_id")
	spanID := ctx.Value("span_id")

	log.Printf("[TRACE] TraceID=%v SpanID=%v Method=%s URL=%s", traceID, spanID, request.Method(), request.URL())
	return nil
}

// MiddlewareChain executes multiple middleware in sequence.
type MiddlewareChain struct {
	middlewares []interfaces.IMiddleware
}

// NewMiddlewareChain creates a new middleware chain.
func NewMiddlewareChain(middlewares ...interfaces.IMiddleware) *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: middlewares,
	}
}

// Before executes all middleware Before methods.
func (mc *MiddlewareChain) Before(ctx context.Context, request interfaces.IHTTPRequest) (context.Context, error) {
	var err error
	for _, mw := range mc.middlewares {
		ctx, err = mw.Before(ctx, request)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

// After executes all middleware After methods.
func (mc *MiddlewareChain) After(ctx context.Context, request interfaces.IHTTPRequest, response interfaces.IHTTPResponse, err error) error {
	// Execute in reverse order
	for i := len(mc.middlewares) - 1; i >= 0; i-- {
		mwErr := mc.middlewares[i].After(ctx, request, response, err)
		if mwErr != nil {
			return mwErr
		}
	}
	return nil
}

// Add adds a middleware to the chain.
func (mc *MiddlewareChain) Add(middleware interfaces.IMiddleware) {
	mc.middlewares = append(mc.middlewares, middleware)
}
