package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthServer_StartStop(t *testing.T) {
	server := NewHealthServer(8080)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test that server is running
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		t.Fatalf("Expected server to be running, got error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Stop server
	cancel()

	// Wait for server to stop
	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			t.Errorf("Expected clean shutdown, got error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Server did not stop within timeout")
	}
}

func TestHealthEndpoint_Response(t *testing.T) {
	server := NewHealthServer(0) // Use random port for testing

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call handler directly
	server.healthHandler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse JSON response
	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Verify response structure
	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}

	if response.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	if response.Version == "" {
		t.Error("Expected non-empty version")
	}

	if response.Uptime < 0 {
		t.Errorf("Expected non-negative uptime, got %v", response.Uptime)
	}
}

func TestHealthEndpoint_WithMetrics(t *testing.T) {
	server := NewHealthServer(0)

	// Set some test metrics
	server.SetExecutionStats(ExecutionStats{
		TotalExecutions:      42,
		SuccessfulExecutions: 38,
		FailedExecutions:     4,
		AverageResponseTime:  150 * time.Millisecond,
		LastExecution:        time.Now().Add(-30 * time.Second),
	})

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.healthHandler(w, req)

	// Parse response
	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Verify metrics are included
	if response.Metrics == nil {
		t.Fatal("Expected metrics to be included")
	}

	if response.Metrics.TotalExecutions != 42 {
		t.Errorf("Expected 42 total executions, got %d", response.Metrics.TotalExecutions)
	}

	if response.Metrics.SuccessfulExecutions != 38 {
		t.Errorf("Expected 38 successful executions, got %d", response.Metrics.SuccessfulExecutions)
	}

	if response.Metrics.FailedExecutions != 4 {
		t.Errorf("Expected 4 failed executions, got %d", response.Metrics.FailedExecutions)
	}

	if response.Metrics.AverageResponseTime != 150*time.Millisecond {
		t.Errorf("Expected 150ms average response time, got %v", response.Metrics.AverageResponseTime)
	}

	if response.Metrics.LastExecution.IsZero() {
		t.Error("Expected non-zero last execution time")
	}
}

func TestHealthEndpoint_ReadinessCheck(t *testing.T) {
	server := NewHealthServer(0)

	// Test readiness when not ready
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	server.readinessHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503 when not ready, got %d", w.Code)
	}

	// Set as ready
	server.SetReady(true)

	// Test readiness when ready
	w = httptest.NewRecorder()
	server.readinessHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 when ready, got %d", w.Code)
	}

	var response ReadinessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if !response.Ready {
		t.Error("Expected ready to be true")
	}
}

func TestHealthEndpoint_LivenessCheck(t *testing.T) {
	server := NewHealthServer(0)

	// Test liveness - should always be OK if server is responding
	req := httptest.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()

	server.livenessHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for liveness, got %d", w.Code)
	}

	var response LivenessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if !response.Alive {
		t.Error("Expected alive to be true")
	}
}

func TestHealthServer_CustomPort(t *testing.T) {
	// Test that server can be configured with custom port
	server := NewHealthServer(9999)

	if server.port != 9999 {
		t.Errorf("Expected port 9999, got %d", server.port)
	}
}

func TestHealthServer_ConcurrentRequests(t *testing.T) {
	server := NewHealthServer(0)

	// Test concurrent access to health endpoint
	const numRequests = 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			server.healthHandler(w, req)
			results <- w.Code
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		select {
		case code := <-results:
			if code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", code)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Request timed out")
		}
	}
}
