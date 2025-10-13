package requestbuilder

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	rb := New()
	if rb == nil {
		t.Fatal("New() returned nil")
	}
	if rb.scheme != "https" {
		t.Errorf("expected default scheme 'https', got '%s'", rb.scheme)
	}
	if rb.timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", rb.timeout)
	}
}

func TestHost(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{"valid host", "api.example.com", "api.example.com", false},
		{"host with http prefix", "http://api.example.com", "api.example.com", false},
		{"host with https prefix", "https://api.example.com", "api.example.com", false},
		{"host with trailing slash", "api.example.com/", "api.example.com", false},
		{"empty host", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := New().Host(tt.input)
			if tt.expectError {
				if rb.err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if rb.err != nil {
					t.Errorf("unexpected error: %v", rb.err)
				}
				if rb.host != tt.expected {
					t.Errorf("expected host '%s', got '%s'", tt.expected, rb.host)
				}
			}
		})
	}
}

func TestScheme(t *testing.T) {
	tests := []struct {
		name        string
		scheme      string
		expectError bool
	}{
		{"http scheme", "http", false},
		{"https scheme", "https", false},
		{"invalid scheme", "ftp", true},
		{"invalid scheme empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := New().Scheme(tt.scheme)
			if tt.expectError {
				if rb.err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if rb.err != nil {
					t.Errorf("unexpected error: %v", rb.err)
				}
				if rb.scheme != tt.scheme {
					t.Errorf("expected scheme '%s', got '%s'", tt.scheme, rb.scheme)
				}
			}
		})
	}
}

func TestAddPath(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		expected string
	}{
		{"single path", []string{"users"}, "/users"},
		{"multiple paths", []string{"api", "v1", "users"}, "/api/v1/users"},
		{"paths with slashes", []string{"/api/", "/v1/", "/users/"}, "/api/v1/users"},
		{"empty path ignored", []string{"api", "", "users"}, "/api/users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := New().Host("example.com")
			for _, p := range tt.paths {
				rb.AddPath(p)
			}
			url, err := rb.buildURL()
			if err != nil {
				t.Fatalf("buildURL failed: %v", err)
			}
			if !strings.Contains(url, tt.expected) {
				t.Errorf("expected path '%s' in URL '%s'", tt.expected, url)
			}
		})
	}
}

func TestQueryParam(t *testing.T) {
	rb := New().
		Host("example.com").
		QueryParam("key1", "value1").
		QueryParam("key2", "value2")

	url, err := rb.buildURL()
	if err != nil {
		t.Fatalf("buildURL failed: %v", err)
	}

	if !strings.Contains(url, "key1=value1") {
		t.Error("expected query param key1=value1")
	}
	if !strings.Contains(url, "key2=value2") {
		t.Error("expected query param key2=value2")
	}
}

func TestQueryParams(t *testing.T) {
	params := map[string]string{
		"foo": "bar",
		"baz": "qux",
	}

	rb := New().
		Host("example.com").
		QueryParams(params)

	url, err := rb.buildURL()
	if err != nil {
		t.Fatalf("buildURL failed: %v", err)
	}

	if !strings.Contains(url, "foo=bar") {
		t.Error("expected query param foo=bar")
	}
	if !strings.Contains(url, "baz=qux") {
		t.Error("expected query param baz=qux")
	}
}

func TestHeader(t *testing.T) {
	rb := New().
		Header("X-Custom-Header", "value1").
		Header("X-Another-Header", "value2")

	if rb.headers.Get("X-Custom-Header") != "value1" {
		t.Error("expected X-Custom-Header to be set")
	}
	if rb.headers.Get("X-Another-Header") != "value2" {
		t.Error("expected X-Another-Header to be set")
	}
}

func TestContentType(t *testing.T) {
	rb := New().ContentType("application/json")
	if rb.headers.Get("Content-Type") != "application/json" {
		t.Error("expected Content-Type to be application/json")
	}
}

func TestBearerToken(t *testing.T) {
	token := "my-secret-token"
	rb := New().BearerToken(token)
	expected := "Bearer " + token
	if rb.headers.Get("Authorization") != expected {
		t.Errorf("expected Authorization header '%s', got '%s'", expected, rb.headers.Get("Authorization"))
	}
}

func TestBodyString(t *testing.T) {
	body := "test body content"
	rb := New().BodyString(body)
	if rb.body == nil {
		t.Error("expected body to be set")
	}
}

func TestBodyJSON(t *testing.T) {
	data := map[string]string{
		"key": "value",
	}
	rb := New().BodyJSON(data)
	if rb.err != nil {
		t.Errorf("unexpected error: %v", rb.err)
	}
	if rb.headers.Get("Content-Type") != "application/json" {
		t.Error("expected Content-Type to be set to application/json")
	}
}

func TestBodyJSONError(t *testing.T) {
	// Use a channel which cannot be marshaled to JSON
	invalidData := make(chan int)
	rb := New().BodyJSON(invalidData)
	if rb.err == nil {
		t.Error("expected error when marshaling invalid JSON")
	}
}

func TestTimeout(t *testing.T) {
	duration := 10 * time.Second
	rb := New().Timeout(duration)
	if rb.timeout != duration {
		t.Errorf("expected timeout %v, got %v", duration, rb.timeout)
	}
}

func TestTimeoutInvalid(t *testing.T) {
	rb := New().Timeout(-1 * time.Second)
	if rb.err == nil {
		t.Error("expected error for negative timeout")
	}
}

func TestContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "key", "value")
	rb := New().Context(ctx)
	if rb.ctx != ctx {
		t.Error("expected context to be set")
	}
}

func TestContextNil(t *testing.T) {
	// Test that passing nil context returns an error
	var nilCtx context.Context
	rb := New().Context(nilCtx)
	if rb.err == nil {
		t.Error("expected error for nil context")
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		builder  *RequestBuilder
		expected string
		hasError bool
	}{
		{
			name: "simple URL",
			builder: New().
				Host("example.com").
				AddPath("api"),
			expected: "https://example.com/api",
			hasError: false,
		},
		{
			name: "URL with multiple paths",
			builder: New().
				Host("example.com").
				AddPath("api").
				AddPath("v1").
				AddPath("users"),
			expected: "https://example.com/api/v1/users",
			hasError: false,
		},
		{
			name: "URL with query params",
			builder: New().
				Host("example.com").
				AddPath("search").
				QueryParam("q", "golang"),
			expected: "https://example.com/search?q=golang",
			hasError: false,
		},
		{
			name: "HTTP scheme",
			builder: New().
				Scheme("http").
				Host("example.com"),
			expected: "http://example.com",
			hasError: false,
		},
		{
			name:     "missing host",
			builder:  New(),
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := tt.builder.buildURL()
			if tt.hasError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if url != tt.expected {
					t.Errorf("expected URL '%s', got '%s'", tt.expected, url)
				}
			}
		})
	}
}

func TestBuild(t *testing.T) {
	rb := New().
		Host("example.com").
		AddPath("api").
		Header("X-Custom", "value")

	rb.method = http.MethodGet
	req, err := rb.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if req.Method != http.MethodGet {
		t.Errorf("expected method GET, got %s", req.Method)
	}
	if req.Header.Get("X-Custom") != "value" {
		t.Error("expected custom header to be set")
	}
}

func TestBuildWithoutMethod(t *testing.T) {
	rb := New().Host("example.com")
	_, err := rb.Build()
	if err == nil {
		t.Error("expected error when building without method")
	}
}

func TestErrorPropagation(t *testing.T) {
	// Set an error early in the chain
	rb := New().
		Host("").  // This will set an error
		AddPath("api").
		QueryParam("key", "value")

	if rb.err == nil {
		t.Error("expected error to be set")
	}

	// Error should prevent further operations
	_, err := rb.Build()
	if err == nil {
		t.Error("expected error to propagate to Build")
	}
}

func TestGetRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	rb := New().
		Scheme("http").
		Host(strings.TrimPrefix(server.URL, "http://")).
		AddPath("test")

	resp, err := rb.Get()
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "success" {
		t.Errorf("expected body 'success', got '%s'", string(body))
	}
}

func TestPostRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "test data" {
			t.Errorf("expected body 'test data', got '%s'", string(body))
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	rb := New().
		Scheme("http").
		Host(strings.TrimPrefix(server.URL, "http://")).
		AddPath("users").
		BodyString("test data")

	resp, err := rb.Post()
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
}

func TestPostRequestWithJSON(t *testing.T) {
	type TestPayload struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("expected Content-Type application/json")
		}

		var payload TestPayload
		json.NewDecoder(r.Body).Decode(&payload)
		if payload.Name != "John" || payload.Email != "john@example.com" {
			t.Error("unexpected JSON payload")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	payload := TestPayload{
		Name:  "John",
		Email: "john@example.com",
	}

	rb := New().
		Scheme("http").
		Host(strings.TrimPrefix(server.URL, "http://")).
		AddPath("users").
		BodyJSON(payload)

	resp, err := rb.Post()
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestPutRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rb := New().
		Scheme("http").
		Host(strings.TrimPrefix(server.URL, "http://"))

	resp, err := rb.Put()
	if err != nil {
		t.Fatalf("PUT request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestDeleteRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	rb := New().
		Scheme("http").
		Host(strings.TrimPrefix(server.URL, "http://"))

	resp, err := rb.Delete()
	if err != nil {
		t.Fatalf("DELETE request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", resp.StatusCode)
	}
}

func TestPatchRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rb := New().
		Scheme("http").
		Host(strings.TrimPrefix(server.URL, "http://"))

	resp, err := rb.Patch()
	if err != nil {
		t.Fatalf("PATCH request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestDoWithCustomClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	customClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	rb := New().
		Scheme("http").
		Host(strings.TrimPrefix(server.URL, "http://"))

	resp, err := rb.Do(http.MethodGet, customClient)
	if err != nil {
		t.Fatalf("Do request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRequestWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Error("expected custom header to be present")
		}
		if r.Header.Get("Authorization") != "Bearer token123" {
			t.Error("expected authorization header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rb := New().
		Scheme("http").
		Host(strings.TrimPrefix(server.URL, "http://")).
		Header("X-Custom-Header", "custom-value").
		BearerToken("token123")

	resp, err := rb.Get()
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()
}

func TestRequestWithQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("search") != "golang" {
			t.Error("expected query param 'search=golang'")
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Error("expected query param 'limit=10'")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rb := New().
		Scheme("http").
		Host(strings.TrimPrefix(server.URL, "http://")).
		QueryParam("search", "golang").
		QueryParam("limit", "10")

	resp, err := rb.Get()
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()
}

func TestComplexRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		// Verify path
		if r.URL.Path != "/api/v1/users" {
			t.Errorf("expected path '/api/v1/users', got '%s'", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("expected Content-Type application/json")
		}
		if r.Header.Get("Authorization") != "Bearer secret-token" {
			t.Error("expected Authorization header")
		}

		// Verify query params
		if r.URL.Query().Get("include") != "profile" {
			t.Error("expected query param 'include=profile'")
		}

		// Verify body
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Alice" {
			t.Error("expected name 'Alice' in body")
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	payload := map[string]string{
		"name":  "Alice",
		"email": "alice@example.com",
	}

	rb := New().
		Scheme("http").
		Host(strings.TrimPrefix(server.URL, "http://")).
		AddPath("api").
		AddPath("v1").
		AddPath("users").
		QueryParam("include", "profile").
		BearerToken("secret-token").
		BodyJSON(payload)

	resp, err := rb.Post()
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
}

