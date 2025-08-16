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

	// Create a production client with all features
	prodClient := client.NewProductionRESTClient("https://api.example.com")

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
	fmt.Println("Production client created:", prodClient != nil)
	fmt.Println("Custom client created:", customClient != nil)

	// Output:
	// Basic client created: true
	// Production client created: true
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
