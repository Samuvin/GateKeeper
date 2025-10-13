package httpclient

import (
	"fmt"
	"net/http"
	"time"

	"data-plane/internal/httpclient/interfaces"
)

// HTTPClient implements IHTTPClient and provides request execution capabilities.
// This separates request building from request execution following the command pattern.
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
func (c *HTTPClient) Send(request interfaces.IHTTPRequest) (interfaces.IHTTPResponse, error) {
	if request == nil {
		return nil, &HTTPError{
			Message: "request cannot be nil",
		}
	}

	httpReq := request.HTTPRequest()
	if httpReq == nil {
		return nil, &HTTPError{
			Message: "invalid request: HTTPRequest is nil",
		}
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &HTTPError{
			request: request,
			Message: fmt.Sprintf("%s request failed", request.Method()),
			Err:     err,
		}
	}

	resp := &Response{
		httpResponse: httpResp,
		request:      request,
	}

	// Check for HTTP errors (4xx, 5xx)
	if httpResp.StatusCode >= 400 {
		return resp, &HTTPError{
			request:    request,
			response:   resp,
			StatusCode: httpResp.StatusCode,
			Message:    fmt.Sprintf("%s request returned error status %d", request.Method(), httpResp.StatusCode),
		}
	}

	return resp, nil
}

// SendWithHandler executes the request and processes the response with a handler.
func (c *HTTPClient) SendWithHandler(request interfaces.IHTTPRequest, handler interfaces.IResponseHandler) (interface{}, error) {
	if handler == nil {
		return nil, &HTTPError{
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
		return nil, &HTTPError{
			request:  request,
			response: resp,
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
