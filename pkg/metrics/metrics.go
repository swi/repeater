package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// MetricsServer provides Prometheus-compatible metrics endpoint
type MetricsServer struct {
	port   int
	server *http.Server
	mu     sync.RWMutex

	// Execution metrics
	successCount int64
	failureCount int64
	durations    []time.Duration

	// Scheduler metrics
	currentInterval time.Duration

	// Rate limit metrics
	rateLimitHits    int64
	rateLimitAllowed int64
}

// NewMetricsServer creates a new metrics server instance
func NewMetricsServer(port int) *MetricsServer {
	return &MetricsServer{
		port:      port,
		durations: make([]time.Duration, 0),
	}
}

// Start starts the metrics server
func (m *MetricsServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", m.metricsHandler)

	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.port),
		Handler: mux,
	}

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return m.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// RecordExecution records an execution result and duration
func (m *MetricsServer) RecordExecution(success bool, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if success {
		m.successCount++
	} else {
		m.failureCount++
	}

	m.durations = append(m.durations, duration)
}

// RecordSchedulerInterval records the current scheduler interval
func (m *MetricsServer) RecordSchedulerInterval(interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentInterval = interval
}

// RecordRateLimitHit records a rate limit hit
func (m *MetricsServer) RecordRateLimitHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rateLimitHits++
}

// RecordRateLimitAllowed records a rate limit allow
func (m *MetricsServer) RecordRateLimitAllowed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rateLimitAllowed++
}

// Reset resets all metrics to zero
func (m *MetricsServer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.successCount = 0
	m.failureCount = 0
	m.durations = make([]time.Duration, 0)
	m.currentInterval = 0
	m.rateLimitHits = 0
	m.rateLimitAllowed = 0
}

// metricsHandler handles the /metrics endpoint
func (m *MetricsServer) metricsHandler(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// Execution counters
	_, _ = fmt.Fprintf(w, "# HELP rpr_executions_total Total number of command executions\n")
	_, _ = fmt.Fprintf(w, "# TYPE rpr_executions_total counter\n")
	_, _ = fmt.Fprintf(w, "rpr_executions_total{status=\"success\"} %d\n", m.successCount)
	_, _ = fmt.Fprintf(w, "rpr_executions_total{status=\"failure\"} %d\n", m.failureCount)

	// Execution duration histogram
	_, _ = fmt.Fprintf(w, "# HELP rpr_execution_duration_seconds Duration of command executions\n")
	_, _ = fmt.Fprintf(w, "# TYPE rpr_execution_duration_seconds histogram\n")

	// Calculate histogram buckets
	buckets := []float64{0.001, 0.01, 0.1, 1.0, 10.0, 60.0, 300.0}
	bucketCounts := make([]int, len(buckets))
	totalDuration := time.Duration(0)

	for _, duration := range m.durations {
		totalDuration += duration
		seconds := duration.Seconds()
		for i, bucket := range buckets {
			if seconds <= bucket {
				bucketCounts[i]++
			}
		}
	}

	// Output histogram buckets
	cumulativeCount := 0
	for i, bucket := range buckets {
		cumulativeCount += bucketCounts[i]
		_, _ = fmt.Fprintf(w, "rpr_execution_duration_seconds_bucket{le=\"%g\"} %d\n", bucket, cumulativeCount)
	}
	_, _ = fmt.Fprintf(w, "rpr_execution_duration_seconds_bucket{le=\"+Inf\"} %d\n", len(m.durations))

	// Output sum and count
	_, _ = fmt.Fprintf(w, "# HELP rpr_execution_duration_seconds_sum Sum of execution durations\n")
	_, _ = fmt.Fprintf(w, "# TYPE rpr_execution_duration_seconds_sum counter\n")
	_, _ = fmt.Fprintf(w, "rpr_execution_duration_seconds_sum %g\n", totalDuration.Seconds())

	_, _ = fmt.Fprintf(w, "# HELP rpr_execution_duration_seconds_count Count of execution durations\n")
	_, _ = fmt.Fprintf(w, "# TYPE rpr_execution_duration_seconds_count counter\n")
	_, _ = fmt.Fprintf(w, "rpr_execution_duration_seconds_count %d\n", len(m.durations))

	// Scheduler interval gauge
	_, _ = fmt.Fprintf(w, "# HELP rpr_scheduler_interval_seconds Current scheduler interval\n")
	_, _ = fmt.Fprintf(w, "# TYPE rpr_scheduler_interval_seconds gauge\n")
	_, _ = fmt.Fprintf(w, "rpr_scheduler_interval_seconds %g\n", m.currentInterval.Seconds())

	// Rate limit counters
	_, _ = fmt.Fprintf(w, "# HELP rpr_rate_limit_total Rate limit events\n")
	_, _ = fmt.Fprintf(w, "# TYPE rpr_rate_limit_total counter\n")
	_, _ = fmt.Fprintf(w, "rpr_rate_limit_total{result=\"hit\"} %d\n", m.rateLimitHits)
	_, _ = fmt.Fprintf(w, "rpr_rate_limit_total{result=\"allowed\"} %d\n", m.rateLimitAllowed)
}
