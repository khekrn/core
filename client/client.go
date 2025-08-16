// Package client provides a full-featured REST client with circuit breaker,
// retry logic, context support, and comprehensive HTTP method support.
//
// This package offers production-ready HTTP client functionality with
// fault tolerance, observability, and flexible configuration options.
//
// Example usage:
//
//	// Basic client
//	client := client.NewDefaultRESTClient()
//	resp, err := client.GET("https://api.example.com/users")
//
//	// Production client with all features
//	prodClient := client.NewProductionRESTClient("https://api.example.com")
//
//	// Custom client
//	customClient := client.NewClientBuilder().
//		WithBaseURL("https://api.example.com").
//		WithTimeout(30 * time.Second).
//		WithDefaultRetry().
//		WithDefaultCircuitBreaker("my-service").
//		WithDatadog(true).
//		Build()
//
//	// Request with options
//	resp, err := client.POST("/users", userData,
//		client.WithContext(ctx),
//		client.WithHeader("Authorization", "Bearer token"),
//		client.WithQueryParam("version", "v2"),
//	)
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	ddhttp "github.com/DataDog/dd-trace-go/contrib/net/http/v2"
	"github.com/sony/gobreaker"
)

// HTTPMethod represents supported HTTP methods
type HTTPMethod string

// Supported HTTP methods
const (
	GET     HTTPMethod = "GET"     // GET method for retrieving resources
	POST    HTTPMethod = "POST"    // POST method for creating resources
	PUT     HTTPMethod = "PUT"     // PUT method for updating/replacing resources
	PATCH   HTTPMethod = "PATCH"   // PATCH method for partial updates
	DELETE  HTTPMethod = "DELETE"  // DELETE method for removing resources
	HEAD    HTTPMethod = "HEAD"    // HEAD method for metadata only
	OPTIONS HTTPMethod = "OPTIONS" // OPTIONS method for capability discovery
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Name        string
	MaxRequests uint32
	Interval    time.Duration
	Timeout     time.Duration
	ReadyToTrip func(counts gobreaker.Counts) bool
}

// RequestConfig holds configuration for a single request
type RequestConfig struct {
	Method      HTTPMethod
	URL         string
	Body        interface{}
	Headers     map[string]string
	QueryParams map[string]string
	Timeout     time.Duration
	Context     context.Context
}

// Response wraps HTTP response with additional metadata
type Response struct {
	*http.Response
	Body       []byte
	StatusCode int
	Headers    http.Header
}

// RESTClient provides a full-featured HTTP client
type RESTClient struct {
	client         *http.Client
	baseURL        string
	defaultHeaders map[string]string
	retry          *RetryConfig
	circuitBreaker *gobreaker.CircuitBreaker
}

// ClientBuilder provides a fluent interface for building REST clients
type ClientBuilder struct {
	timeout             time.Duration
	maxIdleConns        int
	maxIdleConnsPerHost int
	idleConnTimeout     time.Duration
	enableDatadog       bool
	transport           http.RoundTripper
	baseURL             string
	defaultHeaders      map[string]string
	retry               *RetryConfig
	circuitBreaker      *CircuitBreakerConfig
}

// NewClientBuilder creates a new client builder with sensible defaults
func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{
		timeout:             30 * time.Second,
		maxIdleConns:        100,
		maxIdleConnsPerHost: 100,
		idleConnTimeout:     90 * time.Second,
		enableDatadog:       false,
		defaultHeaders:      make(map[string]string),
	}
}

// WithTimeout sets the client timeout
func (b *ClientBuilder) WithTimeout(timeout time.Duration) *ClientBuilder {
	b.timeout = timeout
	return b
}

// WithMaxIdleConns sets the maximum number of idle connections
func (b *ClientBuilder) WithMaxIdleConns(maxIdleConns int) *ClientBuilder {
	b.maxIdleConns = maxIdleConns
	return b
}

// WithMaxIdleConnsPerHost sets the maximum number of idle connections per host
func (b *ClientBuilder) WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) *ClientBuilder {
	b.maxIdleConnsPerHost = maxIdleConnsPerHost
	return b
}

// WithIdleConnTimeout sets the idle connection timeout
func (b *ClientBuilder) WithIdleConnTimeout(idleConnTimeout time.Duration) *ClientBuilder {
	b.idleConnTimeout = idleConnTimeout
	return b
}

// WithDatadog enables Datadog tracing for the HTTP client
func (b *ClientBuilder) WithDatadog(enable bool) *ClientBuilder {
	b.enableDatadog = enable
	return b
}

// WithTransport sets a custom transport
func (b *ClientBuilder) WithTransport(transport http.RoundTripper) *ClientBuilder {
	b.transport = transport
	return b
}

// WithBaseURL sets the base URL for all requests
func (b *ClientBuilder) WithBaseURL(baseURL string) *ClientBuilder {
	b.baseURL = strings.TrimSuffix(baseURL, "/")
	return b
}

// WithDefaultHeader adds a default header to all requests
func (b *ClientBuilder) WithDefaultHeader(key, value string) *ClientBuilder {
	b.defaultHeaders[key] = value
	return b
}

// WithDefaultHeaders sets multiple default headers
func (b *ClientBuilder) WithDefaultHeaders(headers map[string]string) *ClientBuilder {
	for k, v := range headers {
		b.defaultHeaders[k] = v
	}
	return b
}

// WithRetry configures retry behavior
func (b *ClientBuilder) WithRetry(config RetryConfig) *ClientBuilder {
	b.retry = &config
	return b
}

// WithDefaultRetry configures retry with sensible defaults
func (b *ClientBuilder) WithDefaultRetry() *ClientBuilder {
	b.retry = &RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		BackoffFactor:  2.0,
	}
	return b
}

// WithCircuitBreaker configures circuit breaker
func (b *ClientBuilder) WithCircuitBreaker(config CircuitBreakerConfig) *ClientBuilder {
	b.circuitBreaker = &config
	return b
}

// WithDefaultCircuitBreaker configures circuit breaker with sensible defaults
func (b *ClientBuilder) WithDefaultCircuitBreaker(name string) *ClientBuilder {
	b.circuitBreaker = &CircuitBreakerConfig{
		Name:        name,
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	}
	return b
}

// Build creates the REST client with the configured options
func (b *ClientBuilder) Build() *RESTClient {
	var transport http.RoundTripper

	if b.transport != nil {
		transport = b.transport
	} else {
		transport = &http.Transport{
			MaxIdleConns:        b.maxIdleConns,
			MaxIdleConnsPerHost: b.maxIdleConnsPerHost,
			IdleConnTimeout:     b.idleConnTimeout,
		}
	}

	client := &http.Client{
		Timeout:   b.timeout,
		Transport: transport,
	}

	if b.enableDatadog {
		client = ddhttp.WrapClient(client)
	}

	restClient := &RESTClient{
		client:         client,
		baseURL:        b.baseURL,
		defaultHeaders: b.defaultHeaders,
		retry:          b.retry,
	}

	// Configure circuit breaker if specified
	if b.circuitBreaker != nil {
		settings := gobreaker.Settings{
			Name:        b.circuitBreaker.Name,
			MaxRequests: b.circuitBreaker.MaxRequests,
			Interval:    b.circuitBreaker.Interval,
			Timeout:     b.circuitBreaker.Timeout,
			ReadyToTrip: b.circuitBreaker.ReadyToTrip,
		}
		restClient.circuitBreaker = gobreaker.NewCircuitBreaker(settings)
	}

	return restClient
}

// NewDefaultRESTClient creates a default REST client
func NewDefaultRESTClient() *RESTClient {
	return NewClientBuilder().Build()
}

// NewProductionRESTClient creates a production-ready REST client with all features
func NewProductionRESTClient(baseURL string) *RESTClient {
	return NewClientBuilder().
		WithBaseURL(baseURL).
		WithDatadog(true).
		WithDefaultRetry().
		WithDefaultCircuitBreaker("production-client").
		WithDefaultHeader("User-Agent", "core-rest-client/1.0").
		Build()
}

// GetInstance returns the underlying http.Client instance
func (rc *RESTClient) GetInstance() *http.Client {
	return rc.client
}

// buildURL constructs the full URL from base URL and path
func (rc *RESTClient) buildURL(path string) string {
	if rc.baseURL == "" {
		return path
	}
	return rc.baseURL + "/" + strings.TrimPrefix(path, "/")
}

// createRequest creates an HTTP request with proper headers and body
func (rc *RESTClient) createRequest(config RequestConfig) (*http.Request, error) {
	url := rc.buildURL(config.URL)

	var body io.Reader
	if config.Body != nil {
		switch v := config.Body.(type) {
		case string:
			body = strings.NewReader(v)
		case []byte:
			body = bytes.NewReader(v)
		case io.Reader:
			body = v
		default:
			// JSON encode the body
			jsonData, err := json.Marshal(config.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			body = bytes.NewReader(jsonData)
		}
	}

	ctx := config.Context
	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, string(config.Method), url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add default headers
	for k, v := range rc.defaultHeaders {
		req.Header.Set(k, v)
	}

	// Add request-specific headers
	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	// Add query parameters
	if len(config.QueryParams) > 0 {
		q := req.URL.Query()
		for k, v := range config.QueryParams {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	// Set JSON content type if body was auto-marshaled
	if config.Body != nil && req.Header.Get("Content-Type") == "" {
		switch config.Body.(type) {
		case string, []byte, io.Reader:
			// Don't auto-set content type for raw data
		default:
			req.Header.Set("Content-Type", "application/json")
		}
	}

	return req, nil
}

// executeWithRetry executes a request with retry logic
func (rc *RESTClient) executeWithRetry(req *http.Request) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt < rc.getMaxAttempts(); attempt++ {
		if attempt > 0 {
			// Calculate backoff delay
			delay := rc.calculateBackoff(attempt)
			select {
			case <-time.After(delay):
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}

		resp, err := rc.executeRequest(req)
		if err == nil && !rc.shouldRetry(resp.StatusCode) {
			return resp, nil
		}

		lastErr = err
		if err == nil {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// executeRequest executes a single HTTP request
func (rc *RESTClient) executeRequest(req *http.Request) (*Response, error) {
	var resp *http.Response
	var err error

	if rc.circuitBreaker != nil {
		result, cbErr := rc.circuitBreaker.Execute(func() (interface{}, error) {
			return rc.client.Do(req)
		})
		if cbErr != nil {
			return nil, fmt.Errorf("circuit breaker: %w", cbErr)
		}
		resp = result.(*http.Response)
	} else {
		resp, err = rc.client.Do(req)
		if err != nil {
			return nil, err
		}
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		Response:   resp,
		Body:       bodyBytes,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}, nil
}

// getMaxAttempts returns the maximum number of retry attempts
func (rc *RESTClient) getMaxAttempts() int {
	if rc.retry == nil {
		return 1
	}
	return rc.retry.MaxAttempts
}

// calculateBackoff calculates the backoff delay for retry attempts
func (rc *RESTClient) calculateBackoff(attempt int) time.Duration {
	if rc.retry == nil {
		return 0
	}

	delay := time.Duration(float64(rc.retry.InitialBackoff) *
		(rc.retry.BackoffFactor * float64(attempt-1)))

	if delay > rc.retry.MaxBackoff {
		delay = rc.retry.MaxBackoff
	}

	return delay
}

// shouldRetry determines if a status code warrants a retry
func (rc *RESTClient) shouldRetry(statusCode int) bool {
	return statusCode >= 500 || statusCode == 429 || statusCode == 408
}

// Request executes a generic HTTP request
func (rc *RESTClient) Request(config RequestConfig) (*Response, error) {
	req, err := rc.createRequest(config)
	if err != nil {
		return nil, err
	}

	if rc.retry != nil {
		return rc.executeWithRetry(req)
	}

	return rc.executeRequest(req)
}

// GET executes a GET request
func (rc *RESTClient) GET(url string, options ...RequestOption) (*Response, error) {
	config := RequestConfig{Method: GET, URL: url}
	for _, opt := range options {
		opt(&config)
	}
	return rc.Request(config)
}

// POST executes a POST request
func (rc *RESTClient) POST(url string, body interface{}, options ...RequestOption) (*Response, error) {
	config := RequestConfig{Method: POST, URL: url, Body: body}
	for _, opt := range options {
		opt(&config)
	}
	return rc.Request(config)
}

// PUT executes a PUT request
func (rc *RESTClient) PUT(url string, body interface{}, options ...RequestOption) (*Response, error) {
	config := RequestConfig{Method: PUT, URL: url, Body: body}
	for _, opt := range options {
		opt(&config)
	}
	return rc.Request(config)
}

// PATCH executes a PATCH request
func (rc *RESTClient) PATCH(url string, body interface{}, options ...RequestOption) (*Response, error) {
	config := RequestConfig{Method: PATCH, URL: url, Body: body}
	for _, opt := range options {
		opt(&config)
	}
	return rc.Request(config)
}

// DELETE executes a DELETE request
func (rc *RESTClient) DELETE(url string, options ...RequestOption) (*Response, error) {
	config := RequestConfig{Method: DELETE, URL: url}
	for _, opt := range options {
		opt(&config)
	}
	return rc.Request(config)
}

// HEAD executes a HEAD request
func (rc *RESTClient) HEAD(url string, options ...RequestOption) (*Response, error) {
	config := RequestConfig{Method: HEAD, URL: url}
	for _, opt := range options {
		opt(&config)
	}
	return rc.Request(config)
}

// OPTIONS executes an OPTIONS request
func (rc *RESTClient) OPTIONS(url string, options ...RequestOption) (*Response, error) {
	config := RequestConfig{Method: OPTIONS, URL: url}
	for _, opt := range options {
		opt(&config)
	}
	return rc.Request(config)
}

// RequestOption is a function type for configuring requests
type RequestOption func(*RequestConfig)

// WithHeader adds a header to the request
func WithHeader(key, value string) RequestOption {
	return func(config *RequestConfig) {
		if config.Headers == nil {
			config.Headers = make(map[string]string)
		}
		config.Headers[key] = value
	}
}

// WithHeaders adds multiple headers to the request
func WithHeaders(headers map[string]string) RequestOption {
	return func(config *RequestConfig) {
		if config.Headers == nil {
			config.Headers = make(map[string]string)
		}
		for k, v := range headers {
			config.Headers[k] = v
		}
	}
}

// WithQueryParam adds a query parameter to the request
func WithQueryParam(key, value string) RequestOption {
	return func(config *RequestConfig) {
		if config.QueryParams == nil {
			config.QueryParams = make(map[string]string)
		}
		config.QueryParams[key] = value
	}
}

// WithQueryParams adds multiple query parameters to the request
func WithQueryParams(params map[string]string) RequestOption {
	return func(config *RequestConfig) {
		if config.QueryParams == nil {
			config.QueryParams = make(map[string]string)
		}
		for k, v := range params {
			config.QueryParams[k] = v
		}
	}
}

// WithTimeout sets a request-specific timeout
func WithTimeout(timeout time.Duration) RequestOption {
	return func(config *RequestConfig) {
		config.Timeout = timeout
	}
}

// WithContext sets the request context
func WithContext(ctx context.Context) RequestOption {
	return func(config *RequestConfig) {
		config.Context = ctx
	}
}

// JSON parses the response body as JSON
func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// String returns the response body as a string
func (r *Response) String() string {
	return string(r.Body)
}

// IsSuccess returns true if the status code is in the 2xx range
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}
