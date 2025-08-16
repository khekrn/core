# Core - Reusable Go Packages for Microservices

[![Go Reference](https://pkg.go.dev/badge/github.com/khekrn/core.svg)](https://pkg.go.dev/github.com/khekrn/core)
[![Go Report Card](https://goreportcard.com/badge/github.com/khekrn/core)](https://goreportcard.com/report/github.com/khekrn/core)

A comprehensive collection of reusable Go packages designed for building robust, production-ready microservices. This core library provides essential utilities for HTTP clients, logging, response handling, JSON operations, and more.

## 📦 Packages

- **[client](#client-package)** - Full-featured REST client with circuit breaker, retry logic, and observability
- **[response](#response-package)** - Standardized API response structures and utilities
- **[logger](#logger-package)** - Structured logging with context support and multiple output formats
- **[helpers](#helpers-package)** - Generic JSON utilities and common helper functions

## 🚀 Quick Start

```bash
go get github.com/khekrn/core
```

## 📖 Package Documentation

### Client Package

A production-ready REST client with advanced features including circuit breaker, retry logic, context support, and Datadog integration.

#### Features

- ✅ All HTTP methods (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS)
- ✅ Circuit breaker for fault tolerance
- ✅ Configurable retry with exponential backoff
- ✅ Context support for timeouts and cancellation
- ✅ Datadog tracing integration
- ✅ Builder pattern for flexible configuration
- ✅ Request/response middleware support

#### Basic Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/khekrn/core/client"
)

func main() {
    // Create a basic REST client
    restClient := client.NewDefaultRESTClient()

    // Simple GET request
    resp, err := restClient.GET("https://jsonplaceholder.typicode.com/users/1")
    if err != nil {
        log.Fatal(err)
    }

    if resp.IsSuccess() {
        fmt.Println("Response:", resp.String())
    }
}
```

#### Production Usage

```go
// Create a production-ready client with all features
prodClient := client.NewProductionRESTClient("https://api.example.com")

// Custom client with specific configuration
customClient := client.NewClientBuilder().
    WithBaseURL("https://api.example.com").
    WithTimeout(30 * time.Second).
    WithDefaultHeader("Authorization", "Bearer token").
    WithDefaultRetry().
    WithDefaultCircuitBreaker("my-service").
    WithDatadog(true).
    Build()

// HTTP requests with options
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := customClient.POST("/users", user,
    client.WithContext(ctx),
    client.WithHeader("X-Request-ID", "12345"),
    client.WithQueryParam("version", "v2"),
)
```

#### Multiple HTTP Methods

```go
// GET with query parameters
resp, err := client.GET("/users",
    client.WithQueryParam("page", "1"),
    client.WithQueryParam("limit", "10"),
)

// POST with JSON body
user := User{Name: "John", Email: "john@example.com"}
resp, err := client.POST("/users", user)

// PUT and PATCH
resp, err := client.PUT("/users/123", updatedUser)
resp, err := client.PATCH("/users/123", partialUpdate)

// DELETE
resp, err := client.DELETE("/users/123")
```

#### Response Handling

```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

resp, err := client.GET("/users/1")
if err != nil {
    return err
}

// Check status
if !resp.IsSuccess() {
    return fmt.Errorf("request failed: %d", resp.StatusCode)
}

// Parse JSON response
var user User
if err := resp.JSON(&user); err != nil {
    return err
}

// Or get as string
responseText := resp.String()
```

### Response Package

Standardized response structures for consistent API responses across microservices.

#### Basic Usage

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/khekrn/core/response"
)

func main() {
    // Success response
    successResp := response.NewSuccessResponse("User created successfully", map[string]int{"id": 123})

    // Error response
    errorResp := response.NewErrorResponse("User not found")

    // Error with validation details
    validationResp := response.NewErrorResponseWithValidationErrors(
        "Validation failed",
        response.ValidationError{Field: "email", Reason: "Invalid format"},
        response.ValidationError{Field: "age", Reason: "Must be positive"},
    )

    // Convert to JSON
    jsonData, _ := json.Marshal(successResp)
    fmt.Println(string(jsonData))
}
```

#### HTTP Handler Example

```go
func CreateUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        resp := response.NewErrorResponse("Invalid JSON format")
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(resp)
        return
    }

    // Validation
    if user.Email == "" {
        resp := response.NewErrorResponseWithValidationErrors(
            "Validation failed",
            response.ValidationError{Field: "email", Reason: "Email is required"},
        )
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(resp)
        return
    }

    // Create user logic...

    // Success response
    resp := response.NewSuccessResponse("User created successfully", user)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(resp)
}
```

### Logger Package

Structured logging with context support, multiple output formats, and production-ready configuration.

#### Basic Usage

```go
package main

import (
    "context"

    "github.com/khekrn/core/logger"
    "go.uber.org/zap"
)

func main() {
    // Initialize logger
    logger.InitLogger("info", "development")
    defer logger.Sync()

    // Basic logging
    logger.Info("Application started")
    logger.Error("Something went wrong", zap.String("error", "connection failed"))
    logger.Debug("Debug information", zap.Int("user_id", 123))
}
```

#### Context-Aware Logging

```go
func handleRequest(ctx context.Context) {
    // Add request ID to context
    ctx = context.WithValue(ctx, "RequestID", "req-12345")

    // Get logger from context (automatically includes request ID)
    log := logger.FromContext(ctx)

    log.Info("Processing request",
        zap.String("operation", "create_user"),
        zap.String("user_id", "user-456"),
    )

    // Logger will automatically include request_id field
}
```

#### Production Configuration

```go
// Production logger with JSON format and file output
logger.InitLogger("info", "production")

// The logger automatically:
// - Uses JSON encoding for structured logs
// - Disables stack traces in production
// - Outputs to both stdout and /tmp/logs
// - Includes caller information
// - Uses ISO8601 timestamps
```

### Helpers Package

Generic utilities for JSON operations and common helper functions with type safety.

#### JSON Operations

```go
package main

import (
    "fmt"
    "log"

    "github.com/khekrn/core/helpers"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    user := User{ID: 1, Name: "John"}

    // Convert to JSON
    jsonData, err := helpers.ToJSON(user)
    if err != nil {
        log.Fatal(err)
    }

    // Convert from JSON (returns pointer)
    userPtr, err := helpers.FromJSON[User](jsonData)
    if err != nil {
        log.Fatal(err)
    }

    // Convert from JSON (returns value)
    userVal, err := helpers.FromJSONValue[User](jsonData)
    if err != nil {
        log.Fatal(err)
    }

    // Pretty print
    prettyJSON, err := helpers.PrettyPrint(user)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(prettyJSON)
}
```

#### String and Reader Operations

```go
// JSON string operations
jsonStr := `{"id":1,"name":"John"}`
user, err := helpers.FromString[User](jsonStr)

userStr, err := helpers.ToString(user)

// Reader operations
reader := strings.NewReader(jsonStr)
user, err := helpers.FromReader[User](reader)

jsonReader, err := helpers.ToReader(user)
```

#### Validation and Utilities

```go
// Validate JSON
isValid := helpers.ValidateJSON([]byte(`{"id":1}`))
isValidStr := helpers.ValidateJSONString(`{"id":1}`)

// Compact JSON (remove whitespace)
compacted, err := helpers.CompactJSON([]byte(`{
    "id": 1,
    "name": "John"
}`))

// Check if JSON is empty
isEmpty := helpers.IsEmptyJSON([]byte(`{}`)) // true
isEmpty = helpers.IsEmptyJSON([]byte(`{"id":1}`)) // false

// Merge JSON objects
json1 := []byte(`{"a":1,"b":2}`)
json2 := []byte(`{"b":3,"c":4}`)
merged, err := helpers.MergeJSON(json1, json2)
// Result: {"a":1,"b":3,"c":4}
```

#### Must Functions (Panic on Error)

```go
// Use carefully - these panic on error
jsonData := helpers.MustToJSON(user)
user := helpers.MustFromJSON[User](jsonData)
prettyJSON := helpers.MustPrettyPrint(user)
```

## 🏗️ Architecture Examples

### Microservice Setup

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/khekrn/core/client"
    "github.com/khekrn/core/logger"
    "github.com/khekrn/core/response"
)

// Service represents a microservice
type Service struct {
    userClient  *client.RESTClient
    orderClient *client.RESTClient
}

func NewService() *Service {
    // Initialize logger
    logger.InitLogger("info", "production")

    return &Service{
        userClient: client.NewClientBuilder().
            WithBaseURL("https://user-service.example.com").
            WithDefaultCircuitBreaker("user-service").
            WithDefaultRetry().
            WithDatadog(true).
            Build(),

        orderClient: client.NewClientBuilder().
            WithBaseURL("https://order-service.example.com").
            WithDefaultCircuitBreaker("order-service").
            WithTimeout(15 * time.Second). // Orders need more time
            WithDefaultRetry().
            WithDatadog(true).
            Build(),
    }
}

func (s *Service) CreateOrder(ctx context.Context, order Order) (*response.Response, error) {
    // Log the operation
    logger.FromContext(ctx).Info("Creating order",
        zap.String("user_id", order.UserID),
        zap.Float64("amount", order.Amount),
    )

    // Call order service
    resp, err := s.orderClient.POST("/orders", order,
        client.WithContext(ctx),
        client.WithHeader("X-Service", "api-gateway"),
    )

    if err != nil {
        logger.FromContext(ctx).Error("Failed to create order", zap.Error(err))
        return nil, err
    }

    if !resp.IsSuccess() {
        return response.NewErrorResponse("Order creation failed"), nil
    }

    var createdOrder Order
    if err := resp.JSON(&createdOrder); err != nil {
        return nil, err
    }

    return response.NewSuccessResponse("Order created successfully", createdOrder), nil
}
```

### HTTP Middleware with Context

```go
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = generateRequestID()
        }

        // Add request ID to context
        ctx := context.WithValue(r.Context(), "RequestID", requestID)

        // Add to response header
        w.Header().Set("X-Request-ID", requestID)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Get logger with context
        log := logger.FromContext(r.Context())

        log.Info("Request started",
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.String("user_agent", r.UserAgent()),
        )

        next.ServeHTTP(w, r)

        log.Info("Request completed",
            zap.Duration("duration", time.Since(start)),
        )
    })
}
```

## 🧪 Testing

All packages include comprehensive tests. Run tests for individual packages:

```bash
# Test all packages
go test ./...

# Test specific package
go test ./client -v
go test ./helpers -v
go test ./response -v
go test ./logger -v

# Run benchmarks
go test ./helpers -bench=.
go test ./client -bench=.
```

## 📋 Dependencies

- `go.uber.org/zap` - High-performance logging
- `github.com/sony/gobreaker` - Circuit breaker implementation
- `github.com/DataDog/dd-trace-go` - Datadog tracing (optional)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 🔗 Links

- [Go Documentation](https://pkg.go.dev/github.com/khekrn/core)
- [Issues](https://github.com/khekrn/core/issues)

---

Built with ❤️ for the Go microservices community.
