package middleware

import (
	"sync"
	"time"

	"data-plane/internal/transport/http/models"
	"data-plane/internal/transport/interfaces"
)

// AsyncRequest handles asynchronous request execution using goroutines.
type AsyncRequest struct {
	client interfaces.IHTTPClient
}

// Ensure AsyncRequest implements IAsyncRequest interface
var _ interfaces.IAsyncRequest = (*AsyncRequest)(nil)

// NewAsyncRequest creates a new async request handler.
func NewAsyncRequest(client interfaces.IHTTPClient) *AsyncRequest {
	return &AsyncRequest{
		client: client,
	}
}

// Execute sends a single request asynchronously.
func (ar *AsyncRequest) Execute() <-chan interfaces.AsyncResult {
	// This method needs a request, so it's not fully implemented without context.
	// The actual implementation will be in ResilientClient
	resultChan := make(chan interfaces.AsyncResult, 1)
	return resultChan
}

// ExecuteBatch sends multiple requests concurrently using goroutines.
func (ar *AsyncRequest) ExecuteBatch(requests []interfaces.IHTTPRequest) <-chan interfaces.AsyncResult {
	resultChan := make(chan interfaces.AsyncResult, len(requests))

	// Use WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Launch a goroutine for each request
	for _, req := range requests {
		wg.Add(1)
		go func(request interfaces.IHTTPRequest) {
			defer wg.Done()

			start := time.Now()
			resp, err := ar.client.Send(request)
			duration := time.Since(start)

			resultChan <- interfaces.AsyncResult{
				Request:  request,
				Response: resp,
				Error:    err,
				Duration: duration,
			}
		}(req)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan
}

// ExecuteWithCallback sends a request and calls the callback when done.
func (ar *AsyncRequest) ExecuteWithCallback(callback func(interfaces.IHTTPResponse, error)) {
	// This method needs a request, implemented in ResilientClient
	// Placeholder implementation
	go func() {
		callback(nil, nil)
	}()
}

// ExecuteConcurrent executes requests with controlled concurrency.
func ExecuteConcurrent(client interfaces.IHTTPClient, requests []interfaces.IHTTPRequest, maxConcurrency int) <-chan interfaces.AsyncResult {
	resultChan := make(chan interfaces.AsyncResult, len(requests))

	// Create a semaphore channel to limit concurrency
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup

	for _, req := range requests {
		wg.Add(1)

		go func(request interfaces.IHTTPRequest) {
			defer wg.Done()

			// Acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Release slot

			start := time.Now()
			resp, err := client.Send(request)
			duration := time.Since(start)

			resultChan <- interfaces.AsyncResult{
				Request:  request,
				Response: resp,
				Error:    err,
				Duration: duration,
			}
		}(req)
	}

	// Close channel when all requests complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan
}

// FanOut distributes a single request to multiple endpoints concurrently.
func FanOut(client interfaces.IHTTPClient, requests []interfaces.IHTTPRequest) interfaces.AsyncResult {
	results := make(chan interfaces.AsyncResult, len(requests))

	// Launch all requests concurrently
	for _, req := range requests {
		go func(request interfaces.IHTTPRequest) {
			start := time.Now()
			resp, err := client.Send(request)
			duration := time.Since(start)

			results <- interfaces.AsyncResult{
				Request:  request,
				Response: resp,
				Error:    err,
				Duration: duration,
			}
		}(req)
	}

	// Return the first successful result (fan-in pattern)
	for i := 0; i < len(requests); i++ {
		result := <-results
		if result.Error == nil {
			return result
		}
	}

	// All failed, return last result
	return <-results
}

// Pipeline creates a pipeline of async requests where each depends on the previous.
func Pipeline(client interfaces.IHTTPClient, requestBuilders []func(prev interfaces.IHTTPResponse) interfaces.IHTTPRequest) <-chan interfaces.AsyncResult {
	resultChan := make(chan interfaces.AsyncResult, 1)

	go func() {
		defer close(resultChan)

		var prevResp interfaces.IHTTPResponse

		for _, builder := range requestBuilders {
			request := builder(prevResp)
			if request == nil {
				resultChan <- interfaces.AsyncResult{
					Error: &models.HTTPError{Message: "pipeline: request builder returned nil"},
				}
				return
			}

			start := time.Now()
			resp, err := client.Send(request)
			duration := time.Since(start)

			if err != nil {
				resultChan <- interfaces.AsyncResult{
					Request:  request,
					Response: resp,
					Error:    err,
					Duration: duration,
				}
				return
			}

			prevResp = resp
		}

		// Return final result
		resultChan <- interfaces.AsyncResult{
			Response: prevResp,
			Duration: 0, // Total duration would need aggregation
		}
	}()

	return resultChan
}
