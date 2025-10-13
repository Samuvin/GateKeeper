package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"data-plane/internal/httpclient/interfaces"
)

// RequestBuilder provides a fluent interface for building HTTP requests.
// It follows the builder pattern to construct Request objects with
// sensible defaults and comprehensive configuration options.
// It implements the IRequestBuilder interface.
type RequestBuilder struct {
	scheme      string
	host        string
	paths       []string
	queryParams url.Values
	headers     http.Header
	body        io.Reader
	method      string
	timeout     time.Duration
	ctx         context.Context
	client      *http.Client
	err         error
}

// Ensure RequestBuilder implements IRequestBuilder interface
var _ interfaces.IRequestBuilder = (*RequestBuilder)(nil)

// NewBuilder creates a new RequestBuilder with sensible defaults.
// The default scheme is "https" and the default timeout is 30 seconds.
func NewBuilder() interfaces.IRequestBuilder {
	return &RequestBuilder{
		scheme:      "https",
		queryParams: url.Values{},
		headers:     http.Header{},
		timeout:     30 * time.Second,
		ctx:         context.Background(),
	}
}

// Host sets the host for the request (e.g., "api.example.com").
// The host should not include the scheme (http/https).
func (rb *RequestBuilder) Host(host string) interfaces.IRequestBuilder {
	if rb.err != nil {
		return rb
	}
	if host == "" {
		rb.err = fmt.Errorf("host cannot be empty")
		return rb
	}
	// Remove any scheme prefix if accidentally included
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	// Remove trailing slash
	rb.host = strings.TrimSuffix(host, "/")
	return rb
}

// Scheme sets the URL scheme (http or https).
// Defaults to https if not specified.
func (rb *RequestBuilder) Scheme(scheme string) interfaces.IRequestBuilder {
	if rb.err != nil {
		return rb
	}
	if scheme != "http" && scheme != "https" {
		rb.err = fmt.Errorf("scheme must be 'http' or 'https', got: %s", scheme)
		return rb
	}
	rb.scheme = scheme
	return rb
}

// AddPath appends a path segment to the URL path.
// Multiple calls will concatenate paths with proper "/" handling.
// Leading and trailing slashes are handled automatically.
func (rb *RequestBuilder) AddPath(path string) interfaces.IRequestBuilder {
	if rb.err != nil {
		return rb
	}
	if path == "" {
		return rb
	}
	// Clean the path segment
	path = strings.Trim(path, "/")
	if path != "" {
		rb.paths = append(rb.paths, path)
	}
	return rb
}

// Path sets the complete path, replacing any previously added paths.
// This is useful when you want to set the entire path at once.
func (rb *RequestBuilder) Path(path string) interfaces.IRequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.paths = []string{}
	return rb.AddPath(path)
}

// QueryParam adds a single query parameter to the request.
// Multiple values for the same key are supported.
func (rb *RequestBuilder) QueryParam(key, value string) interfaces.IRequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.queryParams.Add(key, value)
	return rb
}

// QueryParams sets multiple query parameters at once.
// This replaces any previously set query parameters.
func (rb *RequestBuilder) QueryParams(params map[string]string) interfaces.IRequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.queryParams = url.Values{}
	for key, value := range params {
		rb.queryParams.Set(key, value)
	}
	return rb
}

// Header adds a header to the request.
// Multiple values for the same header are supported.
func (rb *RequestBuilder) Header(key, value string) interfaces.IRequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.headers.Add(key, value)
	return rb
}

// Headers sets multiple headers at once.
// This replaces any previously set headers.
func (rb *RequestBuilder) Headers(headers map[string]string) interfaces.IRequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.headers = http.Header{}
	for key, value := range headers {
		rb.headers.Set(key, value)
	}
	return rb
}

// ContentType sets the Content-Type header.
func (rb *RequestBuilder) ContentType(contentType string) interfaces.IRequestBuilder {
	return rb.Header("Content-Type", contentType)
}

// Accept sets the Accept header.
func (rb *RequestBuilder) Accept(accept string) interfaces.IRequestBuilder {
	return rb.Header("Accept", accept)
}

// Authorization sets the Authorization header.
func (rb *RequestBuilder) Authorization(token string) interfaces.IRequestBuilder {
	return rb.Header("Authorization", token)
}

// BearerToken sets the Authorization header with a Bearer token.
func (rb *RequestBuilder) BearerToken(token string) interfaces.IRequestBuilder {
	return rb.Header("Authorization", fmt.Sprintf("Bearer %s", token))
}

// Body sets the request body from an io.Reader.
func (rb *RequestBuilder) Body(body io.Reader) interfaces.IRequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.body = body
	return rb
}

// BodyBytes sets the request body from a byte slice.
func (rb *RequestBuilder) BodyBytes(data []byte) interfaces.IRequestBuilder {
	return rb.Body(bytes.NewReader(data))
}

// BodyString sets the request body from a string.
func (rb *RequestBuilder) BodyString(data string) interfaces.IRequestBuilder {
	return rb.Body(strings.NewReader(data))
}

// JSON sets the request body from a JSON-encodable object.
// It automatically sets the Content-Type to application/json.
func (rb *RequestBuilder) JSON(v interface{}) interfaces.IRequestBuilder {
	if rb.err != nil {
		return rb
	}
	data, err := json.Marshal(v)
	if err != nil {
		rb.err = fmt.Errorf("failed to marshal JSON body: %w", err)
		return rb
	}
	rb.ContentType("application/json")
	return rb.BodyBytes(data)
}

// Timeout sets the request timeout duration.
func (rb *RequestBuilder) Timeout(timeout time.Duration) interfaces.IRequestBuilder {
	if rb.err != nil {
		return rb
	}
	if timeout <= 0 {
		rb.err = fmt.Errorf("timeout must be positive, got: %v", timeout)
		return rb
	}
	rb.timeout = timeout
	return rb
}

// WithContext sets the context for the request.
// If not set, context.Background() is used by default.
func (rb *RequestBuilder) WithContext(ctx context.Context) interfaces.IRequestBuilder {
	if rb.err != nil {
		return rb
	}
	if ctx == nil {
		rb.err = fmt.Errorf("context cannot be nil")
		return rb
	}
	rb.ctx = ctx
	return rb
}

// Client sets a custom HTTP client to use for requests.
// If not set, a default client with the configured timeout will be used.
func (rb *RequestBuilder) Client(client *http.Client) interfaces.IRequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.client = client
	return rb
}

// buildURL constructs the complete URL from the builder's components.
func (rb *RequestBuilder) buildURL() (string, error) {
	if rb.host == "" {
		return "", fmt.Errorf("host is required")
	}

	u := &url.URL{
		Scheme: rb.scheme,
		Host:   rb.host,
	}

	// Build path
	if len(rb.paths) > 0 {
		u.Path = "/" + strings.Join(rb.paths, "/")
	}

	// Add query parameters
	if len(rb.queryParams) > 0 {
		u.RawQuery = rb.queryParams.Encode()
	}

	return u.String(), nil
}

// Build constructs the IHTTPRequest object.
// Returns an error if any required fields are missing or invalid.
func (rb *RequestBuilder) Build() (interfaces.IHTTPRequest, error) {
	if rb.err != nil {
		return nil, rb.err
	}

	if rb.method == "" {
		return nil, fmt.Errorf("HTTP method is required")
	}

	urlStr, err := rb.buildURL()
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(rb.ctx, rb.method, urlStr, rb.body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Copy headers to request
	httpReq.Header = rb.headers.Clone()

	return &Request{
		httpRequest: httpReq,
		timeout:     rb.timeout,
	}, nil
}

// Execute sends the request and returns a Response or HTTPError.
// This is an internal helper method that builds and executes the request.
func (rb *RequestBuilder) Execute() (interfaces.IHTTPResponse, error) {
	req, err := rb.Build()
	if err != nil {
		return nil, &HTTPError{
			Message: "failed to build request",
			Err:     err,
		}
	}

	client := rb.client
	if client == nil {
		client = &http.Client{
			Timeout: rb.timeout,
		}
	}

	httpResp, err := client.Do(req.HTTPRequest())
	if err != nil {
		return nil, &HTTPError{
			request: req,
			Message: fmt.Sprintf("%s request failed", req.Method()),
			Err:     err,
		}
	}

	resp := &Response{
		httpResponse: httpResp,
		request:      req,
	}

	// Check for HTTP errors (4xx, 5xx)
	if httpResp.StatusCode >= 400 {
		return resp, &HTTPError{
			request:    req,
			response:   resp,
			StatusCode: httpResp.StatusCode,
			Message:    fmt.Sprintf("%s request returned error status", req.Method()),
		}
	}

	return resp, nil
}

// GET sets the HTTP method to GET and returns the builder.
// Call Build() after this to create the request.
func (rb *RequestBuilder) GET() interfaces.IRequestBuilder {
	rb.method = http.MethodGet
	return rb
}

// POST sets the HTTP method to POST and returns the builder.
// Call Build() after this to create the request.
func (rb *RequestBuilder) POST() interfaces.IRequestBuilder {
	rb.method = http.MethodPost
	return rb
}

// PUT sets the HTTP method to PUT and returns the builder.
// Call Build() after this to create the request.
func (rb *RequestBuilder) PUT() interfaces.IRequestBuilder {
	rb.method = http.MethodPut
	return rb
}

// PATCH sets the HTTP method to PATCH and returns the builder.
// Call Build() after this to create the request.
func (rb *RequestBuilder) PATCH() interfaces.IRequestBuilder {
	rb.method = http.MethodPatch
	return rb
}

// DELETE sets the HTTP method to DELETE and returns the builder.
// Call Build() after this to create the request.
func (rb *RequestBuilder) DELETE() interfaces.IRequestBuilder {
	rb.method = http.MethodDelete
	return rb
}

// Method sets a custom HTTP method and returns the builder.
// Call Build() after this to create the request.
func (rb *RequestBuilder) Method(method string) interfaces.IRequestBuilder {
	rb.method = method
	return rb
}
