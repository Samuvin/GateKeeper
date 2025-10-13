package interfaces

// IHTTPError represents the interface for HTTP request errors.
// This interface extends the standard error interface with additional
// context about HTTP failures, network issues, and retry capabilities.
type IHTTPError interface {
	// error embeds the standard error interface
	error

	// GetRequest returns the request that caused this error.
	GetRequest() IHTTPRequest

	// GetResponse returns the response if available (may be nil for network errors).
	GetResponse() IHTTPResponse

	// GetStatusCode returns the HTTP status code if available (0 for network errors).
	GetStatusCode() int

	// GetMessage returns a human-readable error message.
	GetMessage() string

	// GetError returns the underlying error if available.
	GetError() error

	// IsTimeout returns true if the error was caused by a timeout.
	IsTimeout() bool

	// IsTemporary returns true if the error is temporary and can be retried.
	IsTemporary() bool

	// IsClientError returns true if this is a 4xx client error.
	IsClientError() bool

	// IsServerError returns true if this is a 5xx server error.
	IsServerError() bool

	// IsNetworkError returns true if this is a network-related error.
	IsNetworkError() bool

	// GetResponseBody attempts to read and return the response body if available.
	GetResponseBody() (string, error)

	// Unwrap returns the underlying error for error chain support.
	Unwrap() error
}

