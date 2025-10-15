package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"data-plane/internal/transport"
)

func main() {
	fmt.Println("🚀 HTTP Client Test - Production Ready with Resiliency")

	// Test 0: Three-step pattern (Build → Send → Handle) - JAVA STYLE
	testThreeStepPattern()

	// Test 1: Simple sync request (no resiliency)
	testSimpleSync()

	// Test 2: Request with retry and timeout
	testWithRetry()

	// Test 3: Async request with logging
	testAsync()

	// Test 4: Full resiliency stack (retry, circuit breaker, rate limiter, bulkhead)
	testFullResiliency()

	// Tests 5-8: Legacy examples
	testJSONPlaceholder()
	testPoetryAPI()
	testWeatherAPI()
}

// ============= NEW FLUENT API EXAMPLES =============

// testThreeStepPattern demonstrates the Java-style three-step pattern
// Step 1: Build request (doesn't send)
// Step 2: Send request (separate from building)
// Step 3: Handle response (type-safe handler)
func testThreeStepPattern() {
	fmt.Println("🔷 Test 0: Three-Step Pattern (Build → Send → Handle)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("This matches the Java pattern exactly!")
	fmt.Println()

	// ========== STEP 1: BUILD REQUEST (doesn't send) ==========
	fmt.Println("📦 Step 1: Building request...")
	request, err := transport.NewHTTPBuilder().
		Host("jsonplaceholder.typicode.com").
		AddPath("users").
		AddPath("1").
		Accept("application/json").
		Header("X-Custom-Header", "MyValue").
		GET().
		Build() // ← Only builds, doesn't send!

	if err != nil {
		log.Printf("❌ Failed to build request: %v\n\n", err)
		return
	}
	fmt.Printf("   ✅ Request built: %s %s\n", request.Method(), request.URL())
	fmt.Println()

	// ========== STEP 2: SEND REQUEST (separate from building) ==========
	fmt.Println("📤 Step 2: Sending request...")
	client := transport.NewHTTPClient()
	response, err := client.Send(request) // ← Send the built request

	if err != nil {
		log.Printf("❌ Request failed: %v\n\n", err)
		return
	}
	defer response.Close()
	fmt.Printf("   ✅ Response received: Status %d\n", response.StatusCode())
	fmt.Println()

	// ========== STEP 3: HANDLE RESPONSE (type-safe handler) ==========
	fmt.Println("🔧 Step 3: Handling response with type-safe handler...")

	// Define the expected response structure
	type User struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Website  string `json:"website"`
	}

	// Create a response handler
	handler := transport.NewHTTPResponseHandler().
		WithResponseType(User{}).
		WithAcceptedStatusCodes(200).
		Build()

	// Handle the response
	result, err := handler.Handle(response)
	if err != nil {
		log.Printf("❌ Failed to handle response: %v\n\n", err)
		return
	}

	user := result.(User)
	fmt.Printf("   ✅ Response parsed successfully!\n")
	fmt.Printf("   User ID: %d\n", user.ID)
	fmt.Printf("   Name: %s\n", user.Name)
	fmt.Printf("   Email: %s\n", user.Email)
	fmt.Printf("   Phone: %s\n", user.Phone)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✨ This is the EXACT pattern from your Java code!")
	fmt.Println("   1. Build → 2. Send → 3. Handle")
	fmt.Println()
}

// testSimpleSync demonstrates a simple synchronous request with the fluent API
func testSimpleSync() {
	fmt.Println("🔹 Test 1: Simple Sync Request (Fluent API)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Everything in one fluent chain - Build and execute synchronously
	response, err := transport.NewHTTPBuilder().
		Host("jsonplaceholder.typicode.com").
		AddPath("users").
		AddPath("1").
		Accept("application/json").
		GET().
		Sync() // ← Execute synchronously

	if err != nil {
		log.Printf("❌ Request failed: %v\n\n", err)
		return
	}
	defer response.Close()

	body, _ := response.BodyString()
	fmt.Printf("✅ Success! Status: %d\n", response.StatusCode())
	fmt.Printf("   URL: %s\n", response.Request().URL())
	fmt.Printf("   Body length: %d bytes\n\n", len(body))
}

// testWithRetry demonstrates retry with exponential backoff
func testWithRetry() {
	fmt.Println("🔁 Test 2: Request with Retry and Timeout")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	response, err := transport.NewHTTPBuilder().
		Host("jsonplaceholder.typicode.com").
		Path("/posts/1").
		WithContext(ctx).
		WithRetry(3).              // ← 3 retry attempts with exponential backoff
		Timeout(10 * time.Second). // ← Set timeout
		WithLogging().             // ← Enable logging to see retries
		GET().
		Sync()

	if err != nil {
		log.Printf("❌ Request failed after retries: %v\n\n", err)
		return
	}
	defer response.Close()

	type Post struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		UserID int    `json:"userId"`
	}

	var post Post
	if err := response.JSON(&post); err != nil {
		log.Printf("❌ Failed to parse JSON: %v\n\n", err)
		return
	}

	fmt.Printf("✅ Success! Status: %d\n", response.StatusCode())
	fmt.Printf("   Post ID: %d\n", post.ID)
	fmt.Printf("   Title: %s\n\n", post.Title)
}

// testAsync demonstrates asynchronous request execution with goroutines
func testAsync() {
	fmt.Println("⚡ Test 3: Async Request (Goroutine + Channel)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Execute asynchronously - returns immediately
	resultChan := transport.NewHTTPBuilder().
		Host("poetrydb.org").
		Path("/random").
		WithRetry(2).
		WithLogging().
		GET().
		Async() // ← Returns channel, executes in goroutine

	fmt.Println("📤 Request sent asynchronously, doing other work...")
	time.Sleep(100 * time.Millisecond)
	fmt.Println("⏳ Waiting for result...")

	// Wait for result
	result := <-resultChan

	if result.Error != nil {
		log.Printf("❌ Async request failed: %v\n\n", result.Error)
		return
	}

	fmt.Printf("✅ Success! Status: %d\n", result.Response.StatusCode())
	fmt.Printf("   Duration: %v\n", result.Duration)
	fmt.Printf("   URL: %s\n\n", result.Request.URL())
}

// testFullResiliency demonstrates all resiliency features together
func testFullResiliency() {
	fmt.Println("🛡️  Test 4: Full Resiliency Stack")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	response, err := transport.NewHTTPBuilder().
		Host("jsonplaceholder.typicode.com").
		Path("/comments").
		QueryParam("postId", "1").
		// Resiliency configuration (all in one chain!)
		WithRetry(3).                          // Retry up to 3 times
		WithCircuitBreaker(5, 30*time.Second). // Open after 5 failures
		WithRateLimiter(100, 10).              // 100 req/s, burst 10
		WithBulkhead(50).                      // Max 50 concurrent requests
		Timeout(10 * time.Second).             // 10 second timeout
		WithLogging().                         // Enable logging
		WithMetrics().                         // Enable metrics
		GET().
		Sync()

	if err != nil {
		log.Printf("❌ Request failed: %v\n\n", err)
		return
	}
	defer response.Close()

	type Comment struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Body  string `json:"body"`
	}

	var comments []Comment
	if err := response.JSON(&comments); err != nil {
		log.Printf("❌ Failed to parse JSON: %v\n\n", err)
		return
	}

	fmt.Printf("✅ Success! Status: %d\n", response.StatusCode())
	fmt.Printf("   Comments fetched: %d\n", len(comments))
	if len(comments) > 0 {
		fmt.Printf("   First comment by: %s\n\n", comments[0].Email)
	}
}

// ============= LEGACY EXAMPLES =============

// testJSONPlaceholder demonstrates fetching user data from JSONPlaceholder
func testJSONPlaceholder() {
	fmt.Println("📡 Test 6: Fetching User from JSONPlaceholder API")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Build request using the builder pattern
	request, err := transport.NewHTTPBuilder().
		Host("jsonplaceholder.typicode.com").
		AddPath("users").
		AddPath("1").
		Accept("application/json").
		GET().
		Build()

	if err != nil {
		log.Printf("❌ Failed to build request: %v\n\n", err)
		return
	}

	// Create HTTP client and send request
	client := transport.NewHTTPClient()
	response, err := client.Send(request)

	if err != nil {
		log.Printf("❌ Request failed: %v\n\n", err)
		return
	}
	defer response.Close()

	// Parse response
	type User struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	var user User
	if err := response.JSON(&user); err != nil {
		log.Printf("❌ Failed to parse JSON: %v\n\n", err)
		return
	}

	fmt.Printf("✅ Success! Status: %d\n", response.StatusCode())
	fmt.Printf("   User ID: %d\n", user.ID)
	fmt.Printf("   Name: %s\n", user.Name)
	fmt.Printf("   Email: %s\n\n", user.Email)
}

// testPoetryAPI demonstrates fetching a random poem
func testPoetryAPI() {
	fmt.Println("📚 Test 7: Fetching Random Poem from PoetryDB")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	request, err := transport.NewHTTPBuilder().
		Host("poetrydb.org").
		AddPath("random").
		Accept("application/json").
		WithContext(ctx).
		GET().
		Build()

	if err != nil {
		log.Printf("❌ Failed to build request: %v\n\n", err)
		return
	}

	client := transport.NewHTTPClientWithTimeout(10 * time.Second)
	response, err := client.Send(request)

	if err != nil {
		if httpErr, ok := err.(*transport.HTTPError); ok {
			if httpErr.IsTimeout() {
				log.Printf("❌ Request timed out\n\n")
				return
			}
		}
		log.Printf("❌ Request failed: %v\n\n", err)
		return
	}
	defer response.Close()

	type Poem struct {
		Title  string   `json:"title"`
		Author string   `json:"author"`
		Lines  []string `json:"lines"`
	}

	var poems []Poem
	if err := response.JSON(&poems); err != nil {
		log.Printf("❌ Failed to parse JSON: %v\n\n", err)
		return
	}

	if len(poems) > 0 {
		poem := poems[0]
		fmt.Printf("✅ Success! Status: %d\n", response.StatusCode())
		fmt.Printf("   Title: %s\n", poem.Title)
		fmt.Printf("   Author: %s\n", poem.Author)
		fmt.Printf("   Lines: %d\n", len(poem.Lines))
		if len(poem.Lines) > 0 {
			fmt.Printf("   First line: %s\n\n", poem.Lines[0])
		}
	}
}

// testWeatherAPI demonstrates fetching weather data with type-safe handler
func testWeatherAPI() {
	fmt.Println("🌤️  Test 8: Fetching Weather from Open-Meteo API")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Build request for San Francisco weather
	request, err := transport.NewHTTPBuilder().
		Host("api.open-meteo.com").
		AddPath("v1").
		AddPath("forecast").
		QueryParam("latitude", "37.7749").
		QueryParam("longitude", "-122.4194").
		QueryParam("current_weather", "true").
		Accept("application/json").
		GET().
		Build()

	if err != nil {
		log.Printf("❌ Failed to build request: %v\n\n", err)
		return
	}

	// Use handler for type-safe response processing
	type WeatherResponse struct {
		CurrentWeather struct {
			Temperature float64 `json:"temperature"`
			WindSpeed   float64 `json:"windspeed"`
			WeatherCode int     `json:"weathercode"`
		} `json:"current_weather"`
	}

	client := transport.NewHTTPClient()
	handler := transport.NewHTTPResponseHandler().
		WithResponseType(WeatherResponse{}).
		WithAcceptedStatusCodes(200).
		Build()

	result, err := client.SendWithHandler(request, handler)
	if err != nil {
		log.Printf("❌ Request failed: %v\n\n", err)
		return
	}

	weather := result.(WeatherResponse)
	fmt.Printf("✅ Success! San Francisco Weather:\n")
	fmt.Printf("   Temperature: %.1f°C\n", weather.CurrentWeather.Temperature)
	fmt.Printf("   Wind Speed: %.1f km/h\n", weather.CurrentWeather.WindSpeed)
	fmt.Printf("   Weather Code: %d\n\n", weather.CurrentWeather.WeatherCode)

	fmt.Println("✨ All tests completed successfully!")
}
