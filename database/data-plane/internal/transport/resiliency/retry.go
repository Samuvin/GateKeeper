package resiliency

import (
	"math"
	"time"

	"data-plane/internal/transport/http/models"
	"data-plane/internal/transport/interfaces"
)

// RetryPolicy implements exponential backoff retry logic.
type RetryPolicy struct {
	maxAttempts     int
	initialDelay    time.Duration
	maxDelay        time.Duration
	multiplier      float64
	retryableErrors []int // HTTP status codes to retry
}

// Ensure RetryPolicy implements IRetryPolicy interface
var _ interfaces.IRetryPolicy = (*RetryPolicy)(nil)

// NewRetryPolicy creates a new retry policy with exponential backoff.
func NewRetryPolicy(maxAttempts int) *RetryPolicy {
	return &RetryPolicy{
		maxAttempts:     maxAttempts,
		initialDelay:    100 * time.Millisecond,
		maxDelay:        30 * time.Second,
		multiplier:      2.0,
		retryableErrors: []int{408, 429, 500, 502, 503, 504}, // Timeout, rate limit, server errors
	}
}

// NewRetryPolicyWithConfig creates a retry policy with custom configuration.
func NewRetryPolicyWithConfig(maxAttempts int, initialDelay, maxDelay time.Duration, multiplier float64) *RetryPolicy {
	return &RetryPolicy{
		maxAttempts:     maxAttempts,
		initialDelay:    initialDelay,
		maxDelay:        maxDelay,
		multiplier:      multiplier,
		retryableErrors: []int{408, 429, 500, 502, 503, 504},
	}
}

// ShouldRetry determines if a request should be retried.
func (rp *RetryPolicy) ShouldRetry(err error, attempt int) bool {
	if attempt >= rp.maxAttempts {
		return false
	}

	// Check if it's a retryable HTTP error
	if httpErr, ok := err.(*models.HTTPError); ok {
		// Retry on timeout or temporary errors
		if httpErr.IsTimeout() || httpErr.IsTemporary() {
			return true
		}

		// Retry on specific status codes
		for _, code := range rp.retryableErrors {
			if httpErr.StatusCode == code {
				return true
			}
		}

		// Don't retry client errors (4xx) except specific ones
		if httpErr.IsClientError() {
			return false
		}

		// Retry server errors (5xx)
		if httpErr.IsServerError() {
			return true
		}
	}

	return false
}

// GetDelay calculates the delay for the next retry using exponential backoff.
func (rp *RetryPolicy) GetDelay(attempt int) time.Duration {
	if attempt == 0 {
		return 0
	}

	// Exponential backoff: delay = initialDelay * (multiplier ^ attempt)
	delay := float64(rp.initialDelay) * math.Pow(rp.multiplier, float64(attempt-1))
	delayDuration := time.Duration(delay)

	// Cap at max delay
	if delayDuration > rp.maxDelay {
		delayDuration = rp.maxDelay
	}

	return delayDuration
}

// MaxAttempts returns the maximum number of retry attempts.
func (rp *RetryPolicy) MaxAttempts() int {
	return rp.maxAttempts
}

// WithRetryableStatusCodes sets custom retryable status codes.
func (rp *RetryPolicy) WithRetryableStatusCodes(codes ...int) *RetryPolicy {
	rp.retryableErrors = codes
	return rp
}

// AddRetryableStatusCode adds a status code to the retryable list.
func (rp *RetryPolicy) AddRetryableStatusCode(code int) *RetryPolicy {
	rp.retryableErrors = append(rp.retryableErrors, code)
	return rp
}
