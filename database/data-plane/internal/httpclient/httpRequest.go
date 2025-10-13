package httpclient

import (
	"net/http"
	"time"

	"data-plane/internal/httpclient/interfaces"
)

// Request represents an HTTP request with all its components.
// This is the immutable request object that gets built.
// It implements the IRequest interface.
type Request struct {
	httpRequest *http.Request
	timeout     time.Duration
}

// Ensure Request implements IHTTPRequest interface
var _ interfaces.IHTTPRequest = (*Request)(nil)

// URL returns the request URL as a string.
func (r *Request) URL() string {
	if r.httpRequest == nil {
		return ""
	}
	return r.httpRequest.URL.String()
}

// Method returns the HTTP method of the request.
func (r *Request) Method() string {
	if r.httpRequest == nil {
		return ""
	}
	return r.httpRequest.Method
}

// Header returns a specific header value from the request.
func (r *Request) Header(key string) string {
	if r.httpRequest == nil {
		return ""
	}
	return r.httpRequest.Header.Get(key)
}

// Headers returns all headers from the request.
func (r *Request) Headers() http.Header {
	if r.httpRequest == nil {
		return http.Header{}
	}
	return r.httpRequest.Header
}

// Timeout returns the configured timeout for this request.
func (r *Request) Timeout() time.Duration {
	return r.timeout
}

// HTTPRequest returns the underlying *http.Request object.
func (r *Request) HTTPRequest() *http.Request {
	return r.httpRequest
}
