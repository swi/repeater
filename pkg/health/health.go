package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// HealthServer provides HTTP endpoints for health checking and monitoring
type HealthServer struct {
	port      int
	server    *http.Server
	startTime time.Time
	ready     bool
	mu        sync.RWMutex
	stats     *ExecutionStats
}

// ExecutionStats contains execution metrics for health reporting
type ExecutionStats struct {
	TotalExecutions      int64         `json:"total_executions"`
	SuccessfulExecutions int64         `json:"successful_executions"`
	FailedExecutions     int64         `json:"failed_executions"`
	AverageResponseTime  time.Duration `json:"average_response_time"`
	LastExecution        time.Time     `json:"last_execution"`
}

// HealthResponse represents the response from the /health endpoint
type HealthResponse struct {
	Status    string          `json:"status"`
	Timestamp time.Time       `json:"timestamp"`
	Version   string          `json:"version"`
	Uptime    time.Duration   `json:"uptime"`
	Metrics   *ExecutionStats `json:"metrics,omitempty"`
}

// ReadinessResponse represents the response from the /ready endpoint
type ReadinessResponse struct {
	Ready     bool      `json:"ready"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
}

// LivenessResponse represents the response from the /live endpoint
type LivenessResponse struct {
	Alive     bool      `json:"alive"`
	Timestamp time.Time `json:"timestamp"`
}

// NewHealthServer creates a new health server instance
func NewHealthServer(port int) *HealthServer {
	return &HealthServer{
		port:      port,
		startTime: time.Now(),
		ready:     false,
	}
}

// Start starts the health server
func (h *HealthServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.healthHandler)
	mux.HandleFunc("/ready", h.readinessHandler)
	mux.HandleFunc("/live", h.livenessHandler)

	// Create listener to get actual port when using port 0
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", h.port))
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// Update port with actual assigned port
	h.mu.Lock()
	h.port = listener.Addr().(*net.TCPAddr).Port
	h.mu.Unlock()

	h.server = &http.Server{
		Handler: mux,
	}

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := h.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return h.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// SetReady sets the readiness state
func (h *HealthServer) SetReady(ready bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.ready = ready
}

// SetExecutionStats updates the execution statistics
func (h *HealthServer) SetExecutionStats(stats ExecutionStats) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stats = &stats
}

// healthHandler handles the /health endpoint
func (h *HealthServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "0.2.0", // TODO: Get from build info
		Uptime:    time.Since(h.startTime),
		Metrics:   h.stats,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// readinessHandler handles the /ready endpoint
func (h *HealthServer) readinessHandler(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	ready := h.ready
	h.mu.RUnlock()

	response := ReadinessResponse{
		Ready:     ready,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")

	if ready {
		response.Message = "Service is ready to accept requests"
		w.WriteHeader(http.StatusOK)
	} else {
		response.Message = "Service is not ready"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	_ = json.NewEncoder(w).Encode(response)
}

// livenessHandler handles the /live endpoint
func (h *HealthServer) livenessHandler(w http.ResponseWriter, r *http.Request) {
	response := LivenessResponse{
		Alive:     true,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetPort returns the port the health server is configured to use
func (h *HealthServer) GetPort() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.port
}
