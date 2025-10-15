package models

import (
	"net/http"
	"time"

	"data-plane/internal/transport/interfaces"
)

// Request represents an HTTP request with all its components.
// This is the immutable request object that gets built.
// It implements the IRequest interface.
type Request struct {
	HTTPReq    *http.Request
	TimeoutVal time.Duration
}

// Ensure Request implements IHTTPRequest interface
var _ interfaces.IHTTPRequest = (*Request)(nil)

// URL returns the request URL as a string.
func (r *Request) URL() string {
	if r.HTTPReq == nil {
		return ""
	}
	return r.HTTPReq.URL.String()
}

// Method returns the HTTP method of the request.
func (r *Request) Method() string {
	if r.HTTPReq == nil {
		return ""
	}
	return r.HTTPReq.Method
}

// Header returns a specific header value from the request.
func (r *Request) Header(key string) string {
	if r.HTTPReq == nil {
		return ""
	}
	return r.HTTPReq.Header.Get(key)
}

// Headers returns all headers from the request.
func (r *Request) Headers() http.Header {
	if r.HTTPReq == nil {
		return http.Header{}
	}
	return r.HTTPReq.Header
}

// Timeout returns the configured timeout for this request.
func (r *Request) Timeout() time.Duration {
	return r.TimeoutVal
}

// HTTPRequest returns the underlying *http.Request object.
func (r *Request) HTTPRequest() *http.Request {
	return r.HTTPReq
}
