package interfaces

import (
	"io"
	"net/http"
)

// IHTTPResponse represents the interface for an HTTP response.
// This interface provides methods for accessing response data,
// status codes, headers, and convenient parsing utilities.
type IHTTPResponse interface {
	// StatusCode returns the HTTP status code (e.g., 200, 404, 500).
	StatusCode() int

	// Status returns the HTTP status string (e.g., "200 OK").
	Status() string

	// IsSuccess returns true if the status code is 2xx.
	IsSuccess() bool

	// IsError returns true if the status code is 4xx or 5xx.
	IsError() bool

	// IsClientError returns true if the status code is 4xx.
	IsClientError() bool

	// IsServerError returns true if the status code is 5xx.
	IsServerError() bool

	// Header returns a specific header value from the response.
	Header(key string) string

	// Headers returns all response headers.
	Headers() http.Header

	// Body reads and returns the response body as bytes.
	// The body is cached after the first read.
	Body() ([]byte, error)

	// BodyString reads and returns the response body as a string.
	BodyString() (string, error)

	// JSON unmarshals the response body into the provided interface.
	JSON(v interface{}) error

	// Close closes the response body if it hasn't been read yet.
	Close() error

	// Request returns the original IHTTPRequest that generated this response.
	Request() IHTTPRequest

	// ContentType returns the Content-Type header value.
	ContentType() string

	// ContentLength returns the Content-Length header value.
	ContentLength() int64

	// HTTPResponse returns the underlying *http.Response object.
	HTTPResponse() *http.Response

	// Reader returns an io.ReadCloser for streaming the response body.
	// Use this for large responses to avoid loading everything into memory.
	Reader() io.ReadCloser
}
