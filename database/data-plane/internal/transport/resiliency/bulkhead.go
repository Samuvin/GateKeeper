package resiliency

import (
	"context"
	"errors"
	"sync/atomic"

	"data-plane/internal/transport/interfaces"
)

// Bulkhead implements the bulkhead pattern to limit concurrent requests.
// This prevents resource exhaustion and provides fault isolation.
type Bulkhead struct {
	semaphore      chan struct{} // Channel-based semaphore
	maxConcurrency int
	activeCount    int64 // Atomic counter for active requests
}

// Ensure Bulkhead implements IBulkhead interface
var _ interfaces.IBulkhead = (*Bulkhead)(nil)

// NewBulkhead creates a new bulkhead with specified max concurrency.
func NewBulkhead(maxConcurrency int) *Bulkhead {
	if maxConcurrency <= 0 {
		maxConcurrency = 10 // Default
	}

	return &Bulkhead{
		semaphore:      make(chan struct{}, maxConcurrency),
		maxConcurrency: maxConcurrency,
		activeCount:    0,
	}
}

// Execute runs the function with bulkhead protection.
func (b *Bulkhead) Execute(ctx context.Context, fn func() (interfaces.IHTTPResponse, error)) (interfaces.IHTTPResponse, error) {
	// Try to acquire a slot
	select {
	case b.semaphore <- struct{}{}:
		// Slot acquired
		atomic.AddInt64(&b.activeCount, 1)
		defer func() {
			<-b.semaphore // Release slot
			atomic.AddInt64(&b.activeCount, -1)
		}()

		// Execute the function
		return fn()

	case <-ctx.Done():
		// Context canceled while waiting
		return nil, ctx.Err()

	default:
		// No slots available
		return nil, errors.New("bulkhead: maximum concurrency reached, request rejected")
	}
}

// ActiveRequests returns the current number of active requests.
func (b *Bulkhead) ActiveRequests() int {
	return int(atomic.LoadInt64(&b.activeCount))
}

// MaxConcurrency returns the maximum allowed concurrent requests.
func (b *Bulkhead) MaxConcurrency() int {
	return b.maxConcurrency
}

// AvailableSlots returns the number of available slots.
func (b *Bulkhead) AvailableSlots() int {
	return b.maxConcurrency - b.ActiveRequests()
}

// GetMetrics returns current bulkhead metrics.
func (b *Bulkhead) GetMetrics() BulkheadMetrics {
	active := b.ActiveRequests()
	return BulkheadMetrics{
		MaxConcurrency:     b.maxConcurrency,
		ActiveRequests:     active,
		AvailableSlots:     b.maxConcurrency - active,
		UtilizationPercent: float64(active) / float64(b.maxConcurrency) * 100,
	}
}

// BulkheadMetrics contains bulkhead statistics.
type BulkheadMetrics struct {
	MaxConcurrency     int
	ActiveRequests     int
	AvailableSlots     int
	UtilizationPercent float64
}
