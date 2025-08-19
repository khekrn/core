package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/khekrn/core/client"
)

// Example usage demonstrating the REST client capabilities
func ExampleRESTClient() {
	// Create a basic client
	basicClient := client.NewDefaultRESTClient()

	// Create a custom client with specific configuration
	customClient := client.NewClientBuilder().
		WithBaseURL("https://api.example.com").
		WithTimeout(30*time.Second).
		WithDefaultHeader("Authorization", "Bearer token").
		WithDefaultRetry().
		WithDefaultCircuitBreaker("my-service").
		Build()

	// Example API calls
	fmt.Println("Basic client created:", basicClient != nil)
	fmt.Println("Custom client created:", customClient != nil)

	// Output:
	// Basic client created: true
	// Custom client created: true
}

func TestRESTClient_GET(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"message": "success"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	restClient := client.NewClientBuilder().
		WithBaseURL(server.URL).
		Build()

	// Make GET request
	resp, err := restClient.GET("/test")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}

	if !resp.IsSuccess() {
		t.Errorf("Expected successful response, got status %d", resp.StatusCode)
	}

	// Parse JSON response
	var result map[string]string
	if err := resp.JSON(&result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["message"] != "success" {
		t.Errorf("Expected message 'success', got '%s'", result["message"])
	}
}

func TestRESTClient_POST(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if body["name"] != "test" {
			t.Errorf("Expected name 'test', got '%s'", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"id": "123", "name": body["name"]}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	restClient := client.NewClientBuilder().
		WithBaseURL(server.URL).
		Build()

	// Make POST request
	payload := map[string]string{"name": "test"}
	resp, err := restClient.POST("/users", payload)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}

	if !resp.IsSuccess() {
		t.Errorf("Expected successful response, got status %d", resp.StatusCode)
	}

	// Parse JSON response
	var result map[string]string
	if err := resp.JSON(&result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["id"] != "123" {
		t.Errorf("Expected id '123', got '%s'", result["id"])
	}
}

func TestRESTClient_WithOptions(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check custom header
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Errorf("Expected X-Custom-Header 'custom-value', got '%s'", r.Header.Get("X-Custom-Header"))
		}

		// Check query parameter
		if r.URL.Query().Get("param1") != "value1" {
			t.Errorf("Expected param1 'value1', got '%s'", r.URL.Query().Get("param1"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Create client
	restClient := client.NewClientBuilder().
		WithBaseURL(server.URL).
		Build()

	// Make request with options
	resp, err := restClient.GET("/test",
		client.WithHeader("X-Custom-Header", "custom-value"),
		client.WithQueryParam("param1", "value1"),
		client.WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}

	if !resp.IsSuccess() {
		t.Errorf("Expected successful response, got status %d", resp.StatusCode)
	}
}

func TestRESTClient_Context(t *testing.T) {
	// Create a test server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Create client
	restClient := client.NewClientBuilder().
		WithBaseURL(server.URL).
		Build()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Make request with context
	_, err := restClient.GET("/test", client.WithContext(ctx))
	if err == nil {
		t.Error("Expected context timeout error, got nil")
	}
}

func TestRESTClient_AllHTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					t.Errorf("Expected %s method, got %s", method, r.Method)
				}

				w.WriteHeader(http.StatusOK)
				if method != "HEAD" {
					w.Write([]byte("OK"))
				}
			}))
			defer server.Close()

			// Create client
			restClient := client.NewClientBuilder().
				WithBaseURL(server.URL).
				Build()

			var resp *client.Response
			var err error

			// Test each HTTP method
			switch method {
			case "GET":
				resp, err = restClient.GET("/test")
			case "POST":
				resp, err = restClient.POST("/test", map[string]string{"key": "value"})
			case "PUT":
				resp, err = restClient.PUT("/test", map[string]string{"key": "value"})
			case "PATCH":
				resp, err = restClient.PATCH("/test", map[string]string{"key": "value"})
			case "DELETE":
				resp, err = restClient.DELETE("/test")
			case "HEAD":
				resp, err = restClient.HEAD("/test")
			case "OPTIONS":
				resp, err = restClient.OPTIONS("/test")
			}

			if err != nil {
				t.Fatalf("%s request failed: %v", method, err)
			}

			if !resp.IsSuccess() {
				t.Errorf("Expected successful response for %s, got status %d", method, resp.StatusCode)
			}
		})
	}
}

func TestFromSharedClient(t *testing.T) {
	// Create test servers for different services
	baseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only verify headers for non-minimal requests
		// Minimal client requests go to "/minimal" and don't have default headers
		if r.URL.Path != "/minimal" {
			// Verify base client headers are present
			if r.Header.Get("Authorization") != "Bearer base-token" {
				t.Errorf("Expected Authorization header from base client, got '%s'", r.Header.Get("Authorization"))
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type header from base client, got '%s'", r.Header.Get("Content-Type"))
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service": "base", "status": "success"}`))
	}))
	defer baseServer.Close()

	plexServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify inherited headers
		if r.Header.Get("Authorization") != "Bearer base-token" {
			t.Errorf("Expected inherited Authorization header, got '%s'", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected inherited Content-Type header, got '%s'", r.Header.Get("Content-Type"))
		}
		// Verify service-specific headers
		if r.Header.Get("X-Service") != "plex" {
			t.Errorf("Expected X-Service header 'plex', got '%s'", r.Header.Get("X-Service"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service": "plex", "status": "success"}`))
	}))
	defer plexServer.Close()

	// Step 1: Create base sample client with configuration
	sampleClient := client.NewClientBuilder().
		WithBaseURL(baseServer.URL).
		WithTimeout(20*time.Second).
		WithDefaultHeader("Authorization", "Bearer base-token").
		WithDefaultHeader("Content-Type", "application/json").
		WithDefaultHeader("User-Agent", "base-client/1.0").
		Build()

	// Step 2: Test base client works
	resp, err := sampleClient.GET("/test")
	if err != nil {
		t.Fatalf("Base client GET failed: %v", err)
	}
	if !resp.IsSuccess() {
		t.Errorf("Expected successful response from base client, got %d", resp.StatusCode)
	}

	// Step 3: Create plexClient inheriting from sampleClient
	plexClient := client.FromSharedClient(sampleClient, "plex-service", plexServer.URL).
		WithTimeout(30*time.Second).            // Override timeout
		WithDefaultHeader("X-Service", "plex"). // Add service-specific header
		Build()

	// Step 4: Test plexClient inherits configuration
	resp2, err := plexClient.GET("/plex/test")
	if err != nil {
		t.Fatalf("Plex client GET failed: %v", err)
	}
	if !resp2.IsSuccess() {
		t.Errorf("Expected successful response from plex client, got %d", resp2.StatusCode)
	}

	// Step 5: Verify configuration inheritance and overrides
	// Test that plexClient inherited the base timeout but overrode it
	baseTimeout := sampleClient.GetInstance().Timeout
	plexTimeout := plexClient.GetInstance().Timeout

	if baseTimeout != 20*time.Second {
		t.Errorf("Expected base client timeout 20s, got %v", baseTimeout)
	}
	if plexTimeout != 30*time.Second {
		t.Errorf("Expected plex client timeout 30s, got %v", plexTimeout)
	}

	// Step 6: Create orderClient inheriting base URL
	orderClient := client.FromSharedClient(sampleClient, "order-service", "").
		WithDefaultHeader("X-API-Version", "v2").
		Build()

	// Test that orderClient inherited base URL
	resp3, err := orderClient.GET("/orders")
	if err != nil {
		t.Fatalf("Order client GET failed: %v", err)
	}
	if !resp3.IsSuccess() {
		t.Errorf("Expected successful response from order client, got %d", resp3.StatusCode)
	}

	// Step 7: Test clients without retry/circuit breaker
	minimalClient := client.NewClientBuilder().
		WithBaseURL(baseServer.URL).
		WithoutRetry().
		WithoutCircuitBreaker().
		Build()

	resp4, err := minimalClient.GET("/minimal")
	if err != nil {
		t.Fatalf("Minimal client GET failed: %v", err)
	}
	if !resp4.IsSuccess() {
		t.Errorf("Expected successful response from minimal client, got %d", resp4.StatusCode)
	}
}

func TestDefaultRetryAndCircuitBreaker(t *testing.T) {
	// Create a server that returns errors
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error"))
	}))
	defer errorServer.Close()

	// Create client with the error server and default settings
	errorClient := client.NewClientBuilder().
		WithBaseURL(errorServer.URL).
		Build()

	// This should retry multiple times due to default retry configuration
	start := time.Now()
	_, err := errorClient.GET("/error")
	duration := time.Since(start)

	// Should take longer than a single request due to retries
	if duration < 100*time.Millisecond {
		t.Error("Expected retry behavior with backoff, but request completed too quickly")
	}

	if err == nil {
		t.Error("Expected error due to server errors, but got success")
	}
}
