package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"data-plane/internal/transport/interfaces"
)

// Response wraps http.Response and provides convenient methods
// for handling response data, status codes, and error checking.
// It implements the IHTTPResponse interface.
type Response struct {
	HttpResp   *http.Response
	RequestRef interfaces.IHTTPRequest
	BodyData   []byte
	BodyRead   bool
}

// Ensure Response implements IHTTPResponse interface
var _ interfaces.IHTTPResponse = (*Response)(nil)

// StatusCode returns the HTTP status code of the response.
func (r *Response) StatusCode() int {
	if r.HttpResp == nil {
		return 0
	}
	return r.HttpResp.StatusCode
}

// Status returns the HTTP status string (e.g., "200 OK").
func (r *Response) Status() string {
	if r.HttpResp == nil {
		return ""
	}
	return r.HttpResp.Status
}

// IsSuccess returns true if the status code is 2xx.
func (r *Response) IsSuccess() bool {
	code := r.StatusCode()
	return code >= 200 && code < 300
}

// IsError returns true if the status code is 4xx or 5xx.
func (r *Response) IsError() bool {
	code := r.StatusCode()
	return code >= 400
}

// IsClientError returns true if the status code is 4xx.
func (r *Response) IsClientError() bool {
	code := r.StatusCode()
	return code >= 400 && code < 500
}

// IsServerError returns true if the status code is 5xx.
func (r *Response) IsServerError() bool {
	code := r.StatusCode()
	return code >= 500
}

// Header returns a specific header value from the response.
func (r *Response) Header(key string) string {
	if r.HttpResp == nil {
		return ""
	}
	return r.HttpResp.Header.Get(key)
}

// Headers returns all headers from the response.
func (r *Response) Headers() http.Header {
	if r.HttpResp == nil {
		return http.Header{}
	}
	return r.HttpResp.Header
}

// Body reads and returns the response body as bytes.
// The body is cached after first read.
func (r *Response) Body() ([]byte, error) {
	if r.BodyRead {
		return r.BodyData, nil
	}

	if r.HttpResp == nil || r.HttpResp.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	defer r.HttpResp.Body.Close()

	data, err := io.ReadAll(r.HttpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	r.BodyData = data
	r.BodyRead = true
	return r.BodyData, nil
}

// BodyString reads and returns the response body as a string.
func (r *Response) BodyString() (string, error) {
	body, err := r.Body()
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// JSON unmarshals the response body into the provided interface.
func (r *Response) JSON(v interface{}) error {
	body, err := r.Body()
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return nil
}

// Close closes the response body if it hasn't been read yet.
func (r *Response) Close() error {
	if r.HttpResp != nil && r.HttpResp.Body != nil && !r.BodyRead {
		return r.HttpResp.Body.Close()
	}
	return nil
}

// Request returns the original IHTTPRequest that generated this response.
func (r *Response) Request() interfaces.IHTTPRequest {
	return r.RequestRef
}

// ContentType returns the Content-Type header value.
func (r *Response) ContentType() string {
	return r.Header("Content-Type")
}

// ContentLength returns the Content-Length header value.
func (r *Response) ContentLength() int64 {
	if r.HttpResp == nil {
		return 0
	}
	return r.HttpResp.ContentLength
}

// HTTPResponse returns the underlying *http.Response object.
func (r *Response) HTTPResponse() *http.Response {
	return r.HttpResp
}

// Reader returns an io.ReadCloser for streaming the response body.
// Use this for large responses to avoid loading everything into memory.
func (r *Response) Reader() io.ReadCloser {
	if r.HttpResp == nil {
		return nil
	}
	return r.HttpResp.Body
}
