package interfaces

import (
	"net/http"
	"time"
)

// IHTTPRequest represents the interface for an HTTP request.
// This interface defines the contract for accessing request properties
// and metadata in a read-only fashion.
type IHTTPRequest interface {
	// URL returns the complete request URL as a string.
	URL() string

	// Method returns the HTTP method (GET, POST, PUT, etc.).
	Method() string

	// Header returns a specific header value by key.
	Header(key string) string

	// Headers returns all request headers.
	Headers() http.Header

	// Timeout returns the configured timeout duration for this request.
	Timeout() time.Duration

	// HTTPRequest returns the underlying *http.Request object.
	// Use this when you need direct access to the standard library request.
	HTTPRequest() *http.Request
}
