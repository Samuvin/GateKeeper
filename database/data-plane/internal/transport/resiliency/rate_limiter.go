package resiliency

import (
	"context"
	"sync"
	"time"

	"data-plane/internal/transport/interfaces"
)

// RateLimiter implements token bucket rate limiting.
type RateLimiter struct {
	mu             sync.Mutex
	rate           float64 // Tokens per second
	burst          int     // Maximum burst size
	tokens         float64 // Current tokens
	lastRefillTime time.Time
}

// Ensure RateLimiter implements IRateLimiter interface
var _ interfaces.IRateLimiter = (*RateLimiter)(nil)

// NewRateLimiter creates a new rate limiter.
// rate: requests per second, burst: maximum burst capacity
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	return &RateLimiter{
		rate:           rate,
		burst:          burst,
		tokens:         float64(burst),
		lastRefillTime: time.Now(),
	}
}

// Allow checks if a request is allowed under the rate limit.
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}

// Wait blocks until a request is allowed or context is canceled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.Allow() {
			return nil
		}

		// Calculate wait time
		waitTime := rl.calculateWaitTime()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Continue to next iteration
		}
	}
}

// refill adds tokens based on elapsed time since last refill.
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefillTime).Seconds()

	// Add tokens based on rate and elapsed time
	tokensToAdd := elapsed * rl.rate
	rl.tokens += tokensToAdd

	// Cap at burst size
	if rl.tokens > float64(rl.burst) {
		rl.tokens = float64(rl.burst)
	}

	rl.lastRefillTime = now
}

// calculateWaitTime returns the duration to wait before next token is available.
func (rl *RateLimiter) calculateWaitTime() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens >= 1.0 {
		return 0
	}

	// Calculate time until next token
	tokensNeeded := 1.0 - rl.tokens
	secondsToWait := tokensNeeded / rl.rate

	return time.Duration(secondsToWait * float64(time.Second))
}

// GetMetrics returns current rate limiter metrics.
func (rl *RateLimiter) GetMetrics() RateLimiterMetrics {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	return RateLimiterMetrics{
		Rate:            rl.rate,
		Burst:           rl.burst,
		AvailableTokens: rl.tokens,
	}
}

// RateLimiterMetrics contains rate limiter statistics.
type RateLimiterMetrics struct {
	Rate            float64
	Burst           int
	AvailableTokens float64
}
