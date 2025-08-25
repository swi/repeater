// Package testing provides utilities for real-world HTTP testing with HTTPBin
package testing

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"testing"
	"time"
)

// HTTPBinConfig holds configuration for HTTPBin testing
type HTTPBinConfig struct {
	BaseURL        string
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
	MaxRetries     int
}

// DefaultHTTPBinConfig returns a sensible default configuration
func DefaultHTTPBinConfig() *HTTPBinConfig {
	return &HTTPBinConfig{
		BaseURL:        "https://httpbin.org",
		ConnectTimeout: 3 * time.Second,
		RequestTimeout: 10 * time.Second,
		MaxRetries:     2,
	}
}

// HTTPBinHelper provides utilities for HTTPBin-based testing
type HTTPBinHelper struct {
	config *HTTPBinConfig
	client *http.Client
}

// NewHTTPBinHelper creates a new HTTPBin testing helper
func NewHTTPBinHelper(config *HTTPBinConfig) *HTTPBinHelper {
	if config == nil {
		config = DefaultHTTPBinConfig()
	}

	client := &http.Client{
		Timeout: config.RequestTimeout,
	}

	return &HTTPBinHelper{
		config: config,
		client: client,
	}
}

// SkipIfNoNetwork skips the test if network connectivity is unavailable
func (h *HTTPBinHelper) SkipIfNoNetwork(t *testing.T) {
	t.Helper()

	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping network tests in short mode (use -test.short=false to enable)")
	}

	// Skip if HTTPBin is not reachable
	if !h.IsHTTPBinAvailable() {
		t.Skip("HTTPBin not available - skipping real-world HTTP tests")
	}
}

// IsHTTPBinAvailable checks if HTTPBin is reachable
func (h *HTTPBinHelper) IsHTTPBinAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), h.config.ConnectTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", h.config.BaseURL+"/status/200", nil)
	if err != nil {
		return false
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == 200
}

// HTTPBinEndpoints provides commonly used HTTPBin endpoints for testing
type HTTPBinEndpoints struct {
	baseURL string
}

// NewHTTPBinEndpoints creates endpoint helpers with the given base URL
func NewHTTPBinEndpoints(baseURL string) *HTTPBinEndpoints {
	return &HTTPBinEndpoints{baseURL: baseURL}
}

// Status returns an endpoint that responds with the given HTTP status code
func (e *HTTPBinEndpoints) Status(code int) string {
	return fmt.Sprintf("%s/status/%d", e.baseURL, code)
}

// Delay returns an endpoint that delays response by the given seconds
func (e *HTTPBinEndpoints) Delay(seconds int) string {
	return fmt.Sprintf("%s/delay/%d", e.baseURL, seconds)
}

// JSON returns an endpoint that responds with JSON data
func (e *HTTPBinEndpoints) JSON() string {
	return fmt.Sprintf("%s/json", e.baseURL)
}

// Headers returns an endpoint that returns request headers
func (e *HTTPBinEndpoints) Headers() string {
	return fmt.Sprintf("%s/headers", e.baseURL)
}

// UserAgent returns an endpoint that returns user agent info
func (e *HTTPBinEndpoints) UserAgent() string {
	return fmt.Sprintf("%s/user-agent", e.baseURL)
}

// Get returns an endpoint that accepts GET requests and echoes request data
func (e *HTTPBinEndpoints) Get() string {
	return fmt.Sprintf("%s/get", e.baseURL)
}

// TestScenario represents a real-world HTTP testing scenario
type TestScenario struct {
	Name            string
	Endpoint        string
	Method          string
	ExpectedPattern string
	Description     string
}

// CommonHTTPBinScenarios returns a set of common testing scenarios
func (h *HTTPBinHelper) CommonHTTPBinScenarios() []TestScenario {
	endpoints := NewHTTPBinEndpoints(h.config.BaseURL)

	return []TestScenario{
		{
			Name:            "service_unavailable_503",
			Endpoint:        endpoints.Status(503),
			Method:          "GET",
			ExpectedPattern: "503|Service Unavailable",
			Description:     "Service temporarily unavailable - should trigger retry",
		},
		{
			Name:            "rate_limited_429",
			Endpoint:        endpoints.Status(429),
			Method:          "GET",
			ExpectedPattern: "429|Too Many Requests",
			Description:     "Rate limiting - should respect retry timing",
		},
		{
			Name:            "server_error_502",
			Endpoint:        endpoints.Status(502),
			Method:          "GET",
			ExpectedPattern: "502|Bad Gateway",
			Description:     "Server error - should trigger exponential backoff",
		},
		{
			Name:            "success_response",
			Endpoint:        endpoints.Status(200),
			Method:          "GET",
			ExpectedPattern: "200|OK",
			Description:     "Successful response - should complete successfully",
		},
		{
			Name:            "json_response_parsing",
			Endpoint:        endpoints.JSON(),
			Method:          "GET",
			ExpectedPattern: "slideshow|origin",
			Description:     "JSON response - should parse structured data",
		},
		{
			Name:            "delayed_response",
			Endpoint:        endpoints.Delay(2),
			Method:          "GET",
			ExpectedPattern: "origin|args",
			Description:     "Delayed response - should handle timing correctly",
		},
		{
			Name:            "headers_inspection",
			Endpoint:        endpoints.Headers(),
			Method:          "GET",
			ExpectedPattern: "User-Agent|headers",
			Description:     "Headers inspection - should show request details",
		},
	}
}

// GetCurlCommand returns a curl command string for the given endpoint
func (h *HTTPBinHelper) GetCurlCommand(endpoint, method string) []string {
	// Determine curl binary location
	curlBinary := "curl"
	if runtime.GOOS == "windows" {
		curlBinary = "curl.exe"
	}

	baseArgs := []string{
		curlBinary,
		"-s",               // Silent mode
		"-L",               // Follow redirects
		"--max-time", "30", // Maximum time for operation
		"--connect-timeout", "5", // Connection timeout
	}

	switch method {
	case "POST":
		baseArgs = append(baseArgs, "-X", "POST")
	case "PUT":
		baseArgs = append(baseArgs, "-X", "PUT")
	case "DELETE":
		baseArgs = append(baseArgs, "-X", "DELETE")
		// GET is default, no need to specify
	}

	baseArgs = append(baseArgs, endpoint)
	return baseArgs
}

// ValidatePrerequisites checks if system prerequisites are available
func (h *HTTPBinHelper) ValidatePrerequisites(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test HTTPBin connectivity
	req, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/status/200", nil)
	if err != nil {
		t.Skipf("Cannot validate network prerequisites: %v", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		t.Skipf("Network prerequisites not available: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Skipf("HTTPBin connectivity test failed: status %d", resp.StatusCode)
	}
}
