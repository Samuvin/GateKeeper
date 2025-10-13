package interfaces

import (
	"net/http"
	"time"
)

// IHTTPClient represents an HTTP client that can execute requests.
// This interface separates request building from request execution,
// following the command pattern for better testability and flexibility.
type IHTTPClient interface {
	// Send executes the given request and returns a response.
	// This method performs the actual HTTP call.
	Send(request IHTTPRequest) (IHTTPResponse, error)

	// SendWithHandler executes the request and processes the response with a handler.
	// This allows for custom response processing and type-safe response objects.
	SendWithHandler(request IHTTPRequest, handler IResponseHandler) (interface{}, error)

	// SetTimeout sets the default timeout for all requests.
	SetTimeout(timeout time.Duration)

	// SetHTTPClient sets a custom underlying http.Client.
	SetHTTPClient(client *http.Client)

	// GetHTTPClient returns the underlying http.Client.
	GetHTTPClient() *http.Client
}
