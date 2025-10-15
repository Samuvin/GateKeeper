# 🎓 Learning Guide: Transport Layer Architecture

**A comprehensive guide to understanding the concepts, patterns, and design decisions**

---

## 📚 Table of Contents

1. [Design Patterns Used](#design-patterns-used)
2. [SOLID Principles](#solid-principles)
3. [Go-Specific Patterns](#go-specific-patterns)
4. [Architecture Decisions](#architecture-decisions)
5. [Resiliency Patterns](#resiliency-patterns)
6. [How It All Works Together](#how-it-all-works-together)

---

## 🎨 Design Patterns Used

### 1. **Builder Pattern** 🏗️

**What is it?**  
A creational pattern that constructs complex objects step by step. It separates the construction from representation.

**Why use it?**  
- Makes object creation readable and intuitive
- Handles optional parameters elegantly
- Allows method chaining (fluent API)
- Prevents telescoping constructors

**Our Implementation:**

```go
// ❌ WITHOUT Builder Pattern (ugly!)
client := NewHTTPClient("https://api.example.com", "/users", 
    map[string]string{"Authorization": "Bearer token"}, 
    map[string]string{"page": "1"}, 
    30*time.Second, context.Background(), nil)

// ✅ WITH Builder Pattern (beautiful!)
request, err := transport.NewHTTPBuilder().
    Host("api.example.com").
    AddPath("users").
    BearerToken("token").
    QueryParam("page", "1").
    Timeout(30 * time.Second).
    GET().
    Build()
```

**Key Components:**

```go
// Builder holds the construction state
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
}

// Fluent methods return *RequestBuilder
func (rb *RequestBuilder) Host(host string) interfaces.IRequestBuilder {
    rb.host = host
    return rb  // ← Return self for chaining
}

// Build() creates the final object
func (rb *RequestBuilder) Build() (interfaces.IHTTPRequest, error) {
    // Validate and construct the Request
    return &models.Request{...}, nil
}
```

**Learning Points:**
1. Each method returns `*RequestBuilder` to enable chaining
2. State is accumulated during chaining
3. `Build()` validates and creates the final object
4. Errors can be accumulated and checked at build time

---

### 2. **Factory Pattern** 🏭

**What is it?**  
A creational pattern that provides an interface for creating objects without specifying their exact class.

**Why use it?**  
- Encapsulates object creation logic
- Enables Dependency Injection
- Makes testing easier (inject mocks)
- Centralizes configuration

**Our Implementation:**

```go
// Factory interface
type ClientFactory interface {
    CreateHTTPClient(client *http.Client, timeout time.Duration) interfaces.IHTTPClient
    CreateRetryPolicy(maxAttempts int) interfaces.IRetryPolicy
    CreateCircuitBreaker(threshold int, timeout time.Duration) interfaces.ICircuitBreaker
    CreateRateLimiter(rate float64, burst int) interfaces.IRateLimiter
    CreateBulkhead(maxConcurrency int) interfaces.IBulkhead
}

// Default implementation
type DefaultClientFactory struct{}

func (f *DefaultClientFactory) CreateHTTPClient(client *http.Client, timeout time.Duration) interfaces.IHTTPClient {
    if client != nil {
        return &HTTPClient{httpClient: client, timeout: timeout}
    }
    return NewHTTPClientWithTimeout(timeout)
}

// Usage in RequestBuilder
func NewBuilder() interfaces.IRequestBuilder {
    return &RequestBuilder{
        factory: client.GetDefaultFactory(),  // ← Injected factory
        // ...
    }
}

// Using the factory
func (rb *RequestBuilder) WithRetry(maxAttempts int) interfaces.IRequestBuilder {
    rb.retryPolicy = rb.factory.CreateRetryPolicy(maxAttempts)  // ← Factory creates it
    return rb
}
```

**Learning Points:**
1. Factory creates objects based on interfaces
2. Client code doesn't know concrete types
3. Easy to swap implementations (production vs test)
4. Global default factory for convenience

**For Testing:**

```go
// Mock factory for tests
type MockClientFactory struct{}

func (m *MockClientFactory) CreateRetryPolicy(maxAttempts int) interfaces.IRetryPolicy {
    return &MockRetryPolicy{} // ← Return mock instead of real implementation
}

// Use in tests
builder := NewBuilderWithFactory(&MockClientFactory{})
```

---

### 3. **Decorator Pattern** 🎁

**What is it?**  
A structural pattern that adds behavior to objects dynamically by wrapping them.

**Why use it?**  
- Adds functionality without modifying original code (Open/Closed Principle)
- Composes behaviors flexibly
- Each decorator has single responsibility
- Easy to add/remove features

**Our Implementation:**

```go
// Base interface
type IHTTPClient interface {
    Send(request IHTTPRequest) (IHTTPResponse, error)
}

// Base implementation (simple HTTP client)
type HTTPClient struct {
    httpClient *http.Client
    timeout    time.Duration
}

func (c *HTTPClient) Send(request IHTTPRequest) (IHTTPResponse, error) {
    // Just does HTTP call, nothing else
    return executeHTTP(request)
}

// Retry Decorator (wraps client, adds retry logic)
type RetryDecorator struct {
    client      IHTTPClient        // ← Wraps another client
    retryPolicy interfaces.IRetryPolicy
}

func (d *RetryDecorator) Send(request IHTTPRequest) (IHTTPResponse, error) {
    for attempt := 0; attempt <= d.retryPolicy.MaxAttempts(); attempt++ {
        resp, err := d.client.Send(request)  // ← Delegates to wrapped client
        
        if err == nil || !d.retryPolicy.ShouldRetry(err, attempt) {
            return resp, err
        }
        
        // Wait before retry
        time.Sleep(d.retryPolicy.GetDelay(attempt))
    }
    return nil, err
}

// Circuit Breaker Decorator (wraps client, adds circuit breaking)
type CircuitBreakerDecorator struct {
    client         IHTTPClient  // ← Wraps another client
    circuitBreaker interfaces.ICircuitBreaker
}

func (d *CircuitBreakerDecorator) Send(request IHTTPRequest) (IHTTPResponse, error) {
    if !d.circuitBreaker.AllowRequest() {
        return nil, errors.New("circuit breaker open")
    }
    
    resp, err := d.client.Send(request)  // ← Delegates to wrapped client
    
    if err != nil {
        d.circuitBreaker.RecordFailure()
    } else {
        d.circuitBreaker.RecordSuccess()
    }
    
    return resp, err
}
```

**Decorator Chain:**

```go
// Creating a decorated client (inside RequestBuilder)
func (rb *RequestBuilder) createClientWithResiliency() IHTTPClient {
    // 1. Start with base client
    client := NewHTTPClient()
    
    // 2. Wrap with rate limiter
    if rb.rateLimiter != nil {
        client = NewRateLimiterDecorator(client, rb.rateLimiter)
    }
    
    // 3. Wrap with bulkhead
    if rb.bulkhead != nil {
        client = NewBulkheadDecorator(client, rb.bulkhead)
    }
    
    // 4. Wrap with circuit breaker
    if rb.circuitBreaker != nil {
        client = NewCircuitBreakerDecorator(client, rb.circuitBreaker)
    }
    
    // 5. Wrap with retry
    if rb.retryPolicy != nil {
        client = NewRetryDecorator(client, rb.retryPolicy)
    }
    
    // 6. Wrap with logging
    if rb.enableLogging {
        client = NewLoggingDecorator(client)
    }
    
    return client
}
```

**Flow Visualization:**

```
Request → Logging → Retry → Circuit Breaker → Bulkhead → Rate Limiter → HTTP → Network
                                                                               ↓
Response ← Logging ← Retry ← Circuit Breaker ← Bulkhead ← Rate Limiter ← HTTP ← Network
```

**Learning Points:**
1. Each decorator implements the same interface as the base
2. Decorators wrap other decorators (Russian doll pattern)
3. Order matters (retry should wrap circuit breaker, not vice versa)
4. Each layer adds ONE responsibility

---

### 4. **Strategy Pattern** 🎯

**What is it?**  
A behavioral pattern that defines a family of algorithms and makes them interchangeable.

**Why use it?**  
- Swaps algorithms at runtime
- Encapsulates algorithm logic
- Follows Open/Closed Principle
- Makes testing easier

**Our Implementation:**

```go
// Strategy interface
type IRetryPolicy interface {
    ShouldRetry(err error, attempt int) bool
    GetDelay(attempt int) time.Duration
    MaxAttempts() int
}

// Concrete Strategy 1: Exponential Backoff
type ExponentialBackoffRetry struct {
    maxAttempts  int
    initialDelay time.Duration
    multiplier   float64
}

func (r *ExponentialBackoffRetry) GetDelay(attempt int) time.Duration {
    return time.Duration(float64(r.initialDelay) * math.Pow(r.multiplier, float64(attempt)))
}

// Concrete Strategy 2: Linear Backoff (could add)
type LinearBackoffRetry struct {
    maxAttempts int
    delay       time.Duration
}

func (r *LinearBackoffRetry) GetDelay(attempt int) time.Duration {
    return r.delay * time.Duration(attempt)
}

// Context uses the strategy
type RetryDecorator struct {
    client      IHTTPClient
    retryPolicy IRetryPolicy  // ← Strategy can be swapped
}

// Usage
client1 := NewRetryDecorator(baseClient, &ExponentialBackoffRetry{...})
client2 := NewRetryDecorator(baseClient, &LinearBackoffRetry{...})
```

**Learning Points:**
1. Interface defines the contract
2. Multiple implementations provide different behaviors
3. Client code doesn't know which implementation it's using
4. Easy to add new strategies without changing existing code

---

### 5. **Facade Pattern** 🏛️

**What is it?**  
A structural pattern that provides a simplified interface to a complex subsystem.

**Why use it?**  
- Hides complexity
- Provides clean public API
- Decouples clients from subsystem
- Makes migration easier

**Our Implementation:**

```go
// Complex subsystem
internal/transport/
├── http/
│   ├── builder/
│   ├── client/
│   ├── handler/
│   └── models/
├── interfaces/
├── middleware/
└── resiliency/

// Facade (transport.go)
package transport

// Simple, unified API
func NewHTTPClient() interfaces.IHTTPClient {
    return client.NewHTTPClient()  // ← Hides internal packages
}

func NewHTTPBuilder() interfaces.IRequestBuilder {
    return builder.NewBuilder()  // ← Hides internal packages
}

// Type aliases for convenience
type (
    HTTPClient = client.HTTPClient
    Request    = models.Request
    Response   = models.Response
)

// Usage (client doesn't know about internal structure)
import "data-plane/internal/transport"

client := transport.NewHTTPClient()  // Simple!
```

**Learning Points:**
1. Single entry point for complex system
2. Internal packages are hidden
3. Type aliases provide convenience
4. Makes refactoring easier (change internals, keep facade)

---

### 6. **Command Pattern** 📋

**What is it?**  
A behavioral pattern that encapsulates a request as an object.

**Why use it?**  
- Separates request building from execution
- Requests can be queued, logged, or undone
- Supports batching and async execution
- Makes testing easier

**Our Implementation:**

```go
// Command (Request is a command object)
type Request struct {
    HTTPReq    *http.Request
    TimeoutVal time.Duration
}

// Command methods (query the command, don't execute)
func (r *Request) Method() string {
    return r.HTTPReq.Method
}

func (r *Request) URL() *url.URL {
    return r.HTTPReq.URL
}

// Invoker (Client executes the command)
type HTTPClient struct {
    httpClient *http.Client
}

func (c *HTTPClient) Send(request IHTTPRequest) (IHTTPResponse, error) {
    // Execute the command
    return c.httpClient.Do(request.HTTPRequest())
}

// Usage (3-step pattern)
// 1. Build command (doesn't execute)
request, err := transport.NewHTTPBuilder().
    Host("api.example.com").
    GET().
    Build()

// 2. Execute command separately
client := transport.NewHTTPClient()
response, err := client.Send(request)

// 3. Process result
handler := transport.NewHTTPResponseHandler()
result, err := handler.Handle(response)
```

**Benefits:**
```go
// Can store commands
var pendingRequests []IHTTPRequest
pendingRequests = append(pendingRequests, request)

// Can execute later
for _, req := range pendingRequests {
    client.Send(req)
}

// Can log commands
log.Printf("Executing: %s %s", request.Method(), request.URL())

// Can queue async execution
go func() {
    client.Send(request)
}()
```

**Learning Points:**
1. Request building != Request execution
2. Requests are first-class objects
3. Enables queuing, logging, batching
4. Matches Java pattern exactly

---

## 🔷 SOLID Principles

### **S - Single Responsibility Principle**

**Definition:** A class should have one, and only one, reason to change.

**Our Implementation:**

```go
// ✅ HTTPClient - ONLY does HTTP calls
type HTTPClient struct {
    httpClient *http.Client
    timeout    time.Duration
}

func (c *HTTPClient) Send(request IHTTPRequest) (IHTTPResponse, error) {
    // ONLY HTTP execution, nothing else
    return c.httpClient.Do(request.HTTPRequest())
}

// ✅ RetryDecorator - ONLY does retry logic
type RetryDecorator struct {
    client      IHTTPClient
    retryPolicy IRetryPolicy
}

func (d *RetryDecorator) Send(request IHTTPRequest) (IHTTPResponse, error) {
    // ONLY retry logic, delegates HTTP to wrapped client
    for attempt := 0; attempt <= maxAttempts; attempt++ {
        resp, err := d.client.Send(request)  // ← Delegates
        if shouldRetry(err) {
            continue
        }
        return resp, err
    }
}

// ✅ RequestBuilder - ONLY builds requests
type RequestBuilder struct { ... }

// ✅ ResponseHandler - ONLY handles responses
type ResponseHandler struct { ... }
```

**Why This Matters:**
- Easy to understand (one purpose)
- Easy to test (one thing to test)
- Easy to change (change only affects one thing)
- Easy to reuse (do one thing well)

---

### **O - Open/Closed Principle**

**Definition:** Software entities should be open for extension but closed for modification.

**Our Implementation:**

```go
// ✅ Base client is CLOSED for modification
type HTTPClient struct {
    httpClient *http.Client
}

// ✅ But OPEN for extension via decorators
type LoggingDecorator struct {
    client IHTTPClient  // ← Extends without modifying
}

type RetryDecorator struct {
    client IHTTPClient  // ← Extends without modifying
}

// Adding new feature? Just create new decorator!
type CachingDecorator struct {
    client IHTTPClient
    cache  Cache
}

func (d *CachingDecorator) Send(request IHTTPRequest) (IHTTPResponse, error) {
    // Check cache first
    if cached := d.cache.Get(request.URL()); cached != nil {
        return cached, nil
    }
    
    // Delegate to wrapped client
    resp, err := d.client.Send(request)
    
    // Store in cache
    d.cache.Set(request.URL(), resp)
    
    return resp, err
}

// Use it without modifying HTTPClient!
client := NewHTTPClient()
client = NewCachingDecorator(client, cache)
client = NewRetryDecorator(client, retryPolicy)
```

**Why This Matters:**
- Add features without risk
- Existing code stays stable
- No breaking changes
- Compose features flexibly

---

### **L - Liskov Substitution Principle**

**Definition:** Objects should be replaceable with instances of their subtypes without altering correctness.

**Our Implementation:**

```go
// Interface contract
type IHTTPClient interface {
    Send(request IHTTPRequest) (IHTTPResponse, error)
}

// Any implementation should work identically
func processRequest(client IHTTPClient, request IHTTPRequest) {
    resp, err := client.Send(request)
    // This works with ANY IHTTPClient implementation
}

// All of these are substitutable
client1 := NewHTTPClient()                          // Base client
client2 := NewRetryDecorator(client1, retry)        // With retry
client3 := NewLoggingDecorator(client2)             // With logging

// All work the same way
processRequest(client1, request)  // ✅
processRequest(client2, request)  // ✅
processRequest(client3, request)  // ✅
```

**Why This Matters:**
- Polymorphism works correctly
- Mock implementations in tests
- Swap implementations safely
- Interface contracts are reliable

---

### **I - Interface Segregation Principle**

**Definition:** Clients shouldn't be forced to depend on methods they don't use.

**Our Implementation:**

```go
// ❌ BAD: Fat interface (forces implementation of unused methods)
type HTTPClientInterface interface {
    Send(request IHTTPRequest) (IHTTPResponse, error)
    SendAsync(request IHTTPRequest) <-chan AsyncResult
    SendBatch(requests []IHTTPRequest) []IHTTPResponse
    SendWithRetry(request IHTTPRequest, maxRetries int) (IHTTPResponse, error)
    SendWithCircuitBreaker(request IHTTPRequest) (IHTTPResponse, error)
}

// ✅ GOOD: Small, focused interfaces
type IHTTPClient interface {
    Send(request IHTTPRequest) (IHTTPResponse, error)
}

type IRetryPolicy interface {
    ShouldRetry(err error, attempt int) bool
    GetDelay(attempt int) time.Duration
}

type ICircuitBreaker interface {
    AllowRequest() bool
    RecordSuccess()
    RecordFailure()
}

// Clients depend only on what they need
type RetryDecorator struct {
    client      IHTTPClient      // ← Only needs Send()
    retryPolicy IRetryPolicy     // ← Only needs retry methods
}
```

**Why This Matters:**
- Easier to implement (fewer methods)
- Easier to test (mock only what's needed)
- More flexible (combine small interfaces)
- Clearer intent (explicit dependencies)

---

### **D - Dependency Inversion Principle**

**Definition:** Depend on abstractions, not concretions.

**Our Implementation:**

```go
// ❌ BAD: Depend on concrete type
type RequestBuilder struct {
    retryPolicy *RetryPolicy  // ← Concrete type
}

// ✅ GOOD: Depend on interface
type RequestBuilder struct {
    retryPolicy interfaces.IRetryPolicy  // ← Interface
}

// ❌ BAD: Create concrete types directly
func NewBuilder() *RequestBuilder {
    return &RequestBuilder{
        retryPolicy: NewRetryPolicy(3),  // ← Tight coupling
    }
}

// ✅ GOOD: Inject via factory
func NewBuilder() *RequestBuilder {
    return &RequestBuilder{
        factory: client.GetDefaultFactory(),  // ← Factory creates types
    }
}

func (rb *RequestBuilder) WithRetry(maxAttempts int) *RequestBuilder {
    rb.retryPolicy = rb.factory.CreateRetryPolicy(maxAttempts)  // ← Factory
    return rb
}
```

**For Testing:**

```go
// Easy to inject mocks
type MockRetryPolicy struct{}

func (m *MockRetryPolicy) ShouldRetry(err error, attempt int) bool {
    return true  // Always retry for testing
}

// Inject in tests
builder := &RequestBuilder{
    retryPolicy: &MockRetryPolicy{},  // ← Mock injection
}
```

**Why This Matters:**
- Easy to test (inject mocks)
- Easy to change (swap implementations)
- Loose coupling (depend on contracts)
- Flexible composition (mix and match)

---

## 🐹 Go-Specific Patterns

### 1. **Interface-Based Design**

**Go Philosophy:** Accept interfaces, return structs.

```go
// ✅ Accept interface (flexible input)
func Send(client IHTTPClient, request IHTTPRequest) (IHTTPResponse, error) {
    return client.Send(request)
}

// ✅ Return concrete type (clear output)
func NewHTTPClient() *HTTPClient {  // Not IHTTPClient
    return &HTTPClient{...}
}
```

**Why?**
- Input flexibility (accept any implementation)
- Output clarity (caller knows exact type)
- Better error messages
- No unnecessary boxing

---

### 2. **Error Wrapping (Go 1.13+)**

```go
// ✅ Wrap errors with %w
return fmt.Errorf("failed to build request: %w", err)

// ✅ Provide Unwrap()
func (e *HTTPError) Unwrap() error {
    return e.Err
}

// ✅ Use errors.As() for type checking
var httpErr *HTTPError
if errors.As(err, &httpErr) {
    if httpErr.IsTimeout() {
        // Handle timeout
    }
}
```

---

### 3. **Context Propagation**

```go
// ✅ Context as explicit parameter (not embedded in struct)
func (rb *RequestBuilder) WithContext(ctx context.Context) IRequestBuilder {
    rb.ctx = ctx  // Store temporarily
    return rb
}

// ✅ Use context for cancellation
select {
case resultChan <- result:
case <-ctx.Done():  // Respect cancellation
    return
}

// ✅ Context in HTTP request
httpReq, err := http.NewRequestWithContext(rb.ctx, method, url, body)
```

---

### 4. **Method Receivers**

```go
// ✅ Pointer receiver for mutating methods
func (rb *RequestBuilder) Host(host string) *RequestBuilder {
    rb.host = host  // Mutates
    return rb
}

// ✅ Value receiver for query methods (non-mutating)
func (r Request) Method() string {
    return r.HTTPReq.Method  // Read-only
}
```

**Rule of thumb:**
- Pointer: if it modifies, or if struct is large
- Value: if it's read-only and struct is small

---

### 5. **Goroutines and Channels**

```go
// ✅ Async execution pattern
func (rb *RequestBuilder) Async() <-chan AsyncResult {
    resultChan := make(chan AsyncResult, 1)  // Buffered
    
    go func() {
        defer close(resultChan)  // Always close
        
        result := doWork()
        
        // Non-blocking send with context
        select {
        case resultChan <- result:
        case <-rb.ctx.Done():
        }
    }()
    
    return resultChan  // Return receive-only channel
}

// Usage
resultChan := builder.Async()
result := <-resultChan  // Receive
```

---

### 6. **Atomic Operations for Concurrency**

```go
// ✅ Use sync/atomic for counters
import "sync/atomic"

type CircuitBreaker struct {
    failures uint32  // Must be first for alignment
}

// Atomic increment (thread-safe)
atomic.AddUint32(&cb.failures, 1)

// Atomic read
failures := atomic.LoadUint32(&cb.failures)

// Atomic compare-and-swap
atomic.CompareAndSwapInt64(&cb.state, StateClosed, StateOpen)
```

**Why atomic over mutex?**
- Faster for simple operations
- Lock-free
- Better for high-contention scenarios

---

## 🏗️ Architecture Decisions

### 1. **Why Protocol-Agnostic Core?**

**Decision:** Separate interfaces/resiliency from HTTP implementation.

```
internal/transport/
├── interfaces/        # Protocol-agnostic
├── resiliency/        # Protocol-agnostic
├── middleware/        # Protocol-agnostic
└── http/              # HTTP-specific
```

**Benefits:**
- ✅ Add gRPC without duplicating resiliency logic
- ✅ Add WebSocket reusing same patterns
- ✅ Consistent behavior across protocols
- ✅ Test resiliency independently

**Example: Adding gRPC**

```go
// Reuse existing interfaces
type GRPCClient struct {
    conn *grpc.ClientConn
}

func (c *GRPCClient) Send(request IRequest) (IResponse, error) {
    // gRPC-specific implementation
}

// Reuse same decorators!
grpcClient := NewGRPCClient()
grpcClient = NewRetryDecorator(grpcClient, retryPolicy)       // ✅ Works!
grpcClient = NewCircuitBreakerDecorator(grpcClient, breaker)  // ✅ Works!
```

---

### 2. **Why Three-Step Pattern?**

**Pattern:** Build → Send → Handle (separate steps)

**Rationale:**
1. **Matches Java pattern** (your requirement)
2. **Testability:** Can test building without sending
3. **Flexibility:** Can build once, send multiple times
4. **Clarity:** Each step has clear responsibility
5. **Debugging:** Easy to inspect request before sending

**Comparison:**

```go
// ❌ One-step (less flexible)
response := http.Get("https://api.example.com/users")

// ✅ Three-step (more flexible)
request := builder.Build()     // 1. Build (can inspect/log)
response := client.Send(request)  // 2. Send (can retry/modify)
result := handler.Handle(response)  // 3. Handle (type-safe)
```

---

### 3. **Why Decorator Over Inheritance?**

**Decision:** Use composition (decorators) instead of inheritance.

**Why?**
- Go doesn't have inheritance
- Composition is more flexible
- Can mix and match features
- Follows Open/Closed Principle

**Inheritance (if Go had it):**

```java
// ❌ Java-style (inflexible)
class HTTPClient { }
class RetryHTTPClient extends HTTPClient { }
class LoggingRetryHTTPClient extends RetryHTTPClient { }
class CircuitBreakerLoggingRetryHTTPClient extends LoggingRetryHTTPClient { }
// Combinatorial explosion! 🤯
```

**Composition (Go way):**

```go
// ✅ Go-style (flexible)
client := NewHTTPClient()
client = NewRetryDecorator(client, retry)
client = NewLoggingDecorator(client)
client = NewCircuitBreakerDecorator(client, breaker)
// Mix and match freely! 🎉
```

---

### 4. **Why Factory Pattern?**

**Decision:** Use factory for creating components.

**Benefits:**
1. **Testability:** Inject mock factory in tests
2. **Configuration:** Centralize creation logic
3. **Flexibility:** Swap implementations easily
4. **Encapsulation:** Hide construction details

**Without Factory:**

```go
// ❌ Tight coupling
type RequestBuilder struct {
    retryPolicy *RetryPolicy
}

func (rb *RequestBuilder) WithRetry(max int) {
    rb.retryPolicy = NewRetryPolicy(max)  // ← Hard-coded
}

// Hard to test!
```

**With Factory:**

```go
// ✅ Loose coupling
type RequestBuilder struct {
    factory ClientFactory  // ← Interface
}

func (rb *RequestBuilder) WithRetry(max int) {
    rb.retryPolicy = rb.factory.CreateRetryPolicy(max)  // ← Factory
}

// Easy to test!
func TestRetry(t *testing.T) {
    mockFactory := &MockFactory{}
    builder := NewBuilderWithFactory(mockFactory)
    // Test with mocks
}
```

---

## 🛡️ Resiliency Patterns

### 1. **Retry with Exponential Backoff**

**What:** Automatically retry failed requests with increasing delays.

**Formula:**
```
delay = initialDelay × (multiplier ^ attempt)

Example:
Attempt 1: 100ms × (2^0) = 100ms
Attempt 2: 100ms × (2^1) = 200ms
Attempt 3: 100ms × (2^2) = 400ms
Attempt 4: 100ms × (2^3) = 800ms
```

**Implementation:**

```go
type RetryPolicy struct {
    maxAttempts  int             // Max retry attempts
    initialDelay time.Duration   // Starting delay
    maxDelay     time.Duration   // Cap delay
    multiplier   float64         // Exponential factor
}

func (rp *RetryPolicy) GetDelay(attempt int) time.Duration {
    delay := float64(rp.initialDelay) * math.Pow(rp.multiplier, float64(attempt))
    if delay > float64(rp.maxDelay) {
        return rp.maxDelay
    }
    return time.Duration(delay)
}

func (rp *RetryPolicy) ShouldRetry(err error, attempt int) bool {
    if attempt >= rp.maxAttempts {
        return false
    }
    
    // Retry on specific errors
    if httpErr, ok := err.(*HTTPError); ok {
        return httpErr.IsTimeout() || httpErr.IsServerError()
    }
    
    return false
}
```

**When to Use:**
- ✅ Transient failures (network glitches)
- ✅ Server overload (503 errors)
- ✅ Rate limiting (429 errors)
- ❌ Client errors (400, 404) - won't help

---

### 2. **Circuit Breaker**

**What:** Stop calling a failing service temporarily to let it recover.

**States:**

```
┌─────────┐
│ Closed  │ ← Normal operation
└────┬────┘
     │ Failures > threshold
     ▼
┌─────────┐
│  Open   │ ← Reject all requests (fast fail)
└────┬────┘
     │ After timeout
     ▼
┌──────────┐
│Half-Open │ ← Try one request
└────┬─────┘
     │ Success: close
     │ Failure: open again
     ▼
```

**Implementation:**

```go
const (
    StateClosed   = 0  // Normal
    StateOpen     = 1  // Failing
    StateHalfOpen = 2  // Testing
)

type CircuitBreaker struct {
    state            int64   // atomic
    failures         uint32  // atomic
    failureThreshold int
    timeout          time.Duration
    lastFailureTime  int64   // atomic
}

func (cb *CircuitBreaker) AllowRequest() bool {
    state := atomic.LoadInt64(&cb.state)
    
    if state == StateClosed {
        return true  // Normal operation
    }
    
    if state == StateOpen {
        // Check if timeout passed
        lastFailure := atomic.LoadInt64(&cb.lastFailureTime)
        if time.Since(time.Unix(0, lastFailure)) > cb.timeout {
            atomic.StoreInt64(&cb.state, StateHalfOpen)
            return true  // Try one request
        }
        return false  // Still open
    }
    
    // Half-open: allow one request
    return true
}

func (cb *CircuitBreaker) RecordFailure() {
    failures := atomic.AddUint32(&cb.failures, 1)
    atomic.StoreInt64(&cb.lastFailureTime, time.Now().UnixNano())
    
    if int(failures) >= cb.failureThreshold {
        atomic.StoreInt64(&cb.state, StateOpen)
    }
}
```

**When to Use:**
- ✅ Prevent cascading failures
- ✅ Give failing service time to recover
- ✅ Fail fast instead of hanging
- ✅ Monitor service health

---

### 3. **Rate Limiting**

**What:** Limit requests per second to avoid overwhelming service.

**Algorithm:** Token Bucket
- Bucket holds tokens
- Tokens added at constant rate
- Each request consumes one token
- If no tokens, request waits or fails

**Implementation:**

```go
import "golang.org/x/time/rate"

type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter(rps float64, burst int) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(rate.Limit(rps), burst),
    }
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
    return rl.limiter.Wait(ctx)  // Blocks until token available
}

func (rl *RateLimiter) Allow() bool {
    return rl.limiter.Allow()  // Non-blocking
}
```

**Example:**

```
Rate: 10 req/s, Burst: 5

Time  Tokens  Action
0s    5       Start (full bucket)
0.1s  4       Request (consume 1)
0.2s  3       Request (consume 1)
0.3s  2       Request (consume 1)
1.0s  12      Wait 0.7s (refill at 10/s)
1.1s  11      Request (consume 1)
```

**When to Use:**
- ✅ Protect your own services
- ✅ Comply with API limits
- ✅ Prevent abuse
- ✅ Fair resource allocation

---

### 4. **Bulkhead Pattern**

**What:** Isolate resources to prevent total failure (like ship compartments).

**Concept:**
```
Without Bulkhead:
┌──────────────┐
│ 1000 threads │ ← All threads can be used by one service
└──────────────┘
If service hangs, entire app hangs! ❌

With Bulkhead:
┌────────┬────────┬────────┐
│ 100    │ 100    │ 100    │
│Service1│Service2│Service3│
└────────┴────────┴────────┘
If Service1 hangs, others still work! ✅
```

**Implementation:**

```go
type Bulkhead struct {
    semaphore chan struct{}  // Buffered channel as semaphore
    timeout   time.Duration
}

func NewBulkhead(maxConcurrency int) *Bulkhead {
    return &Bulkhead{
        semaphore: make(chan struct{}, maxConcurrency),
        timeout:   5 * time.Second,
    }
}

func (bh *Bulkhead) Acquire(ctx context.Context) error {
    select {
    case bh.semaphore <- struct{}{}:  // Acquire slot
        return nil
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(bh.timeout):
        return errors.New("bulkhead timeout")
    }
}

func (bh *Bulkhead) Release() {
    <-bh.semaphore  // Release slot
}

// Usage
err := bulkhead.Acquire(ctx)
if err != nil {
    return err
}
defer bulkhead.Release()

// Do work (protected by bulkhead)
resp, err := client.Send(request)
```

**When to Use:**
- ✅ Limit concurrent requests
- ✅ Prevent resource exhaustion
- ✅ Isolate different services
- ✅ Protect critical resources

---

## 🔗 How It All Works Together

### Complete Request Flow

```go
// 1. User builds request
request := transport.NewHTTPBuilder().
    Host("api.example.com").
    AddPath("users").
    BearerToken("token").
    WithRetry(3).
    WithCircuitBreaker(5, 30*time.Second).
    WithRateLimiter(100, 10).
    WithLogging().
    GET().
    Build()

// 2. Builder creates decorated client
client := createClientWithResiliency()
// client = Logging(Retry(CircuitBreaker(RateLimit(HTTPClient))))

// 3. Send request (flows through decorators)
response := client.Send(request)
```

**Decorator Flow:**

```
Request
  │
  ▼
┌──────────────┐
│   Logging    │ ← Logs "→ GET /users"
└──────┬───────┘
       ▼
┌──────────────┐
│    Retry     │ ← Retries on failure
└──────┬───────┘
       ▼
┌──────────────┐
│Circuit Breaker│ ← Opens on too many failures
└──────┬───────┘
       ▼
┌──────────────┐
│Rate Limiter  │ ← Waits if over limit
└──────┬───────┘
       ▼
┌──────────────┐
│  HTTPClient  │ ← Actual HTTP call
└──────┬───────┘
       ▼
    Network
       │
       ▼
    Response
       │
       ▼
┌──────────────┐
│   Logging    │ ← Logs "← 200 OK in 50ms"
└──────────────┘
```

---

## 🎯 Key Takeaways

### **Design Patterns = Reusable Solutions**

1. **Builder** → Complex object construction
2. **Factory** → Centralized object creation
3. **Decorator** → Add behavior dynamically
4. **Strategy** → Swap algorithms
5. **Facade** → Simplify complex system
6. **Command** → Encapsulate requests

### **SOLID Principles = Good Design**

1. **S**ingle Responsibility → One job per class
2. **O**pen/Closed → Extend, don't modify
3. **L**iskov Substitution → Swap implementations safely
4. **I**nterface Segregation → Small, focused interfaces
5. **D**ependency Inversion → Depend on abstractions

### **Go Patterns = Idiomatic Code**

1. Accept interfaces, return structs
2. Error wrapping with `%w`
3. Context propagation
4. Goroutines + channels for async
5. Atomic operations for concurrency

### **Resiliency Patterns = Robust Systems**

1. **Retry** → Handle transient failures
2. **Circuit Breaker** → Prevent cascading failures
3. **Rate Limiting** → Control request rate
4. **Bulkhead** → Isolate failures

---

## 📚 Further Reading

### **Books**
- "Design Patterns" by Gang of Four
- "Clean Architecture" by Robert C. Martin
- "Release It!" by Michael Nygard (Resiliency)
- "Concurrency in Go" by Katherine Cox-Buday

### **Online Resources**
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Resiliency Patterns](https://docs.microsoft.com/en-us/azure/architecture/patterns/category/resiliency)

### **Practice**
- Read the code in `internal/transport/`
- Modify decorators to add features
- Write unit tests for patterns
- Build your own HTTP client using these patterns

---

**You've learned professional-grade software engineering! 🎓**

These patterns are used in production systems at Google, Netflix, Amazon, and more. Master them and you'll write better code anywhere.

**Next Steps:**
1. Study each pattern implementation in the code
2. Draw diagrams of how they interact
3. Write tests for each component
4. Add a new feature using these patterns
5. Refactor your other projects with these principles

Happy coding! 🚀

