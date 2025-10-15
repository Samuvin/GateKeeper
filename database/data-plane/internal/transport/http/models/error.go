package models

import (
	"errors"
	"fmt"
	"net"

	"data-plane/internal/transport/interfaces"
)

// HTTPError represents an error that occurred during an HTTP request.
// It provides detailed context about the request, response, and error.
// It implements the IHTTPError interface.
type HTTPError struct {
	Request    interfaces.IHTTPRequest
	Response   interfaces.IHTTPResponse
	StatusCode int
	Message    string
	Err        error
}

// Ensure HTTPError implements IHTTPError interface
var _ interfaces.IHTTPError = (*HTTPError)(nil)

// Error implements the error interface for HTTPError.
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("%s (status: %d)", e.Message, e.StatusCode)
	}
	return e.Message
}

// Unwrap returns the underlying error for error chain support.
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// IsTimeout returns true if the error was caused by a timeout.
func (e *HTTPError) IsTimeout() bool {
	if e.Err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(e.Err, &netErr) {
		return netErr.Timeout()
	}

	return false
}

// IsTemporary returns true if the error is temporary and the request can be retried.
// For network errors, checks if it's a timeout. 5xx errors are considered temporary.
func (e *HTTPError) IsTemporary() bool {
	if e.Err == nil {
		// 5xx errors are typically temporary
		if e.StatusCode >= 500 && e.StatusCode < 600 {
			return true
		}
		return false
	}

	// Timeout errors are temporary
	if e.IsTimeout() {
		return true
	}

	// 5xx errors are typically temporary
	if e.StatusCode >= 500 && e.StatusCode < 600 {
		return true
	}

	return false
}

// IsClientError returns true if the error is a 4xx client error.
func (e *HTTPError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError returns true if the error is a 5xx server error.
func (e *HTTPError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// IsNetworkError returns true if the error is a network-related error.
func (e *HTTPError) IsNetworkError() bool {
	if e.Err == nil {
		return false
	}

	var netErr net.Error
	return errors.As(e.Err, &netErr)
}

// GetRequest returns the request that caused this error.
func (e *HTTPError) GetRequest() interfaces.IHTTPRequest {
	return e.Request
}

// GetResponse returns the response if available (may be nil for network errors).
func (e *HTTPError) GetResponse() interfaces.IHTTPResponse {
	return e.Response
}

// GetStatusCode returns the HTTP status code if available (0 for network errors).
func (e *HTTPError) GetStatusCode() int {
	return e.StatusCode
}

// GetMessage returns a human-readable error message.
func (e *HTTPError) GetMessage() string {
	return e.Message
}

// GetError returns the underlying error if available.
func (e *HTTPError) GetError() error {
	return e.Err
}

// GetResponseBody attempts to read and return the response body if available.
func (e *HTTPError) GetResponseBody() (string, error) {
	if e.Response == nil {
		return "", fmt.Errorf("no response available")
	}
	return e.Response.BodyString()
}

// NewHTTPError creates a new HTTPError with the given message.
func NewHTTPError(message string) *HTTPError {
	return &HTTPError{
		Message: message,
	}
}

// NewHTTPErrorWithStatus creates a new HTTPError with message and status code.
func NewHTTPErrorWithStatus(message string, statusCode int) *HTTPError {
	return &HTTPError{
		Message:    message,
		StatusCode: statusCode,
	}
}

// WrapError wraps an existing error with additional context.
func WrapError(message string, err error) *HTTPError {
	return &HTTPError{
		Message: message,
		Err:     err,
	}
}
