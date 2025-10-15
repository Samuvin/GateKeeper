package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"data-plane/internal/transport/http/models"
	"data-plane/internal/transport/interfaces"
)

// HTTPClient implements IHTTPClient and provides basic HTTP request execution.
// This follows the Single Responsibility Principle - it only performs HTTP calls.
// Resiliency features (retry, circuit breaker, etc.) are handled by decorators.
type HTTPClient struct {
	httpClient *http.Client
	timeout    time.Duration
}

// Ensure HTTPClient implements IHTTPClient interface
var _ interfaces.IHTTPClient = (*HTTPClient)(nil)

// NewHTTPClient creates a new HTTPClient with default configuration.
func NewHTTPClient() interfaces.IHTTPClient {
	return &HTTPClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30 * time.Second,
	}
}

// NewHTTPClientWithTimeout creates a new HTTPClient with the specified timeout.
func NewHTTPClientWithTimeout(timeout time.Duration) interfaces.IHTTPClient {
	return &HTTPClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Send executes the given request and returns a response.
// This method only performs the HTTP call - no resiliency logic.
// Resiliency is handled by decorators wrapping this client.
func (c *HTTPClient) Send(request interfaces.IHTTPRequest) (interfaces.IHTTPResponse, error) {
	if request == nil {
		return nil, &models.HTTPError{
			Message: "request cannot be nil",
		}
	}

	httpReq := request.HTTPRequest()
	if httpReq == nil {
		return nil, &models.HTTPError{
			Message: "invalid request: HTTPRequest is nil",
		}
	}

	// Create context with timeout if configured
	ctx := httpReq.Context()
	if c.timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()
		httpReq = httpReq.WithContext(timeoutCtx)
	}

	// Execute HTTP request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &models.HTTPError{
			Request: request,
			Message: fmt.Sprintf("%s request failed", request.Method()),
			Err:     err,
		}
	}

	resp := &models.Response{
		HttpResp:   httpResp,
		RequestRef: request,
	}

	// Check for HTTP errors (4xx, 5xx)
	if httpResp.StatusCode >= 400 {
		return resp, &models.HTTPError{
			Request:    request,
			Response:   resp,
			StatusCode: httpResp.StatusCode,
			Message:    fmt.Sprintf("%s request returned error status %d", request.Method(), httpResp.StatusCode),
		}
	}

	return resp, nil
}

// SendWithHandler executes the request and processes the response with a handler.
func (c *HTTPClient) SendWithHandler(request interfaces.IHTTPRequest, handler interfaces.IResponseHandler) (interface{}, error) {
	if handler == nil {
		return nil, &models.HTTPError{
			Message: "response handler cannot be nil",
		}
	}

	resp, err := c.Send(request)
	if err != nil {
		// Check if handler can process error responses
		if resp != nil && handler.CanHandle(resp) {
			if handlerErr := handler.HandleError(resp); handlerErr != nil {
				return nil, handlerErr
			}
		}
		return nil, err
	}

	if !handler.CanHandle(resp) {
		return nil, &models.HTTPError{
			Request:  request,
			Response: resp,
			Message:  "handler cannot process this response",
		}
	}

	return handler.Handle(resp)
}

// SetTimeout sets the default timeout for all requests.
func (c *HTTPClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	if c.httpClient != nil {
		c.httpClient.Timeout = timeout
	}
}

// SetHTTPClient sets a custom underlying http.Client.
func (c *HTTPClient) SetHTTPClient(client *http.Client) {
	if client != nil {
		c.httpClient = client
	}
}

// GetHTTPClient returns the underlying http.Client.
func (c *HTTPClient) GetHTTPClient() *http.Client {
	return c.httpClient
}
