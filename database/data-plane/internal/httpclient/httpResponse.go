package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"data-plane/internal/httpclient/interfaces"
)

// Response wraps http.Response and provides convenient methods
// for handling response data, status codes, and error checking.
// It implements the IHTTPResponse interface.
type Response struct {
	httpResponse *http.Response
	request      interfaces.IHTTPRequest
	body         []byte
	bodyRead     bool
}

// Ensure Response implements IHTTPResponse interface
var _ interfaces.IHTTPResponse = (*Response)(nil)

// StatusCode returns the HTTP status code of the response.
func (r *Response) StatusCode() int {
	if r.httpResponse == nil {
		return 0
	}
	return r.httpResponse.StatusCode
}

// Status returns the HTTP status string (e.g., "200 OK").
func (r *Response) Status() string {
	if r.httpResponse == nil {
		return ""
	}
	return r.httpResponse.Status
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
	if r.httpResponse == nil {
		return ""
	}
	return r.httpResponse.Header.Get(key)
}

// Headers returns all headers from the response.
func (r *Response) Headers() http.Header {
	if r.httpResponse == nil {
		return http.Header{}
	}
	return r.httpResponse.Header
}

// Body reads and returns the response body as bytes.
// The body is cached after first read.
func (r *Response) Body() ([]byte, error) {
	if r.bodyRead {
		return r.body, nil
	}

	if r.httpResponse == nil || r.httpResponse.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	defer r.httpResponse.Body.Close()

	data, err := io.ReadAll(r.httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	r.body = data
	r.bodyRead = true
	return r.body, nil
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
	if r.httpResponse != nil && r.httpResponse.Body != nil && !r.bodyRead {
		return r.httpResponse.Body.Close()
	}
	return nil
}

// Request returns the original IHTTPRequest that generated this response.
func (r *Response) Request() interfaces.IHTTPRequest {
	return r.request
}

// ContentType returns the Content-Type header value.
func (r *Response) ContentType() string {
	return r.Header("Content-Type")
}

// ContentLength returns the Content-Length header value.
func (r *Response) ContentLength() int64 {
	if r.httpResponse == nil {
		return 0
	}
	return r.httpResponse.ContentLength
}

// HTTPResponse returns the underlying *http.Response object.
func (r *Response) HTTPResponse() *http.Response {
	return r.httpResponse
}

// Reader returns an io.ReadCloser for streaming the response body.
// Use this for large responses to avoid loading everything into memory.
func (r *Response) Reader() io.ReadCloser {
	if r.httpResponse == nil {
		return nil
	}
	return r.httpResponse.Body
}
