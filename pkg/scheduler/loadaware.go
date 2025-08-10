package scheduler

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SystemMetrics represents current system resource usage
type SystemMetrics struct {
	CPUUsage       float64   // CPU usage percentage (0-100)
	MemoryUsage    float64   // Memory usage percentage (0-100)
	LoadAverage1m  float64   // 1-minute load average
	LoadAverage5m  float64   // 5-minute load average
	LoadAverage15m float64   // 15-minute load average
	Timestamp      time.Time // When metrics were collected
}

// SystemResourceMonitor monitors system resource usage
type SystemResourceMonitor struct {
	mu sync.RWMutex
}

// NewSystemResourceMonitor creates a new system resource monitor
func NewSystemResourceMonitor() *SystemResourceMonitor {
	return &SystemResourceMonitor{}
}

// GetCurrentMetrics returns current system metrics
func (m *SystemResourceMonitor) GetCurrentMetrics() (*SystemMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	// Get CPU usage (simplified - in production would use more sophisticated monitoring)
	metrics.CPUUsage = m.getCPUUsage()

	// Get memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	metrics.MemoryUsage = m.getMemoryUsage(&memStats)

	// Get load averages (Unix-like systems only)
	loadAvg1, loadAvg5, loadAvg15, err := m.getLoadAverages()
	if err != nil {
		// Default to reasonable values if load average unavailable
		loadAvg1, loadAvg5, loadAvg15 = 0.5, 0.5, 0.5
	}
	metrics.LoadAverage1m = loadAvg1
	metrics.LoadAverage5m = loadAvg5
	metrics.LoadAverage15m = loadAvg15

	return metrics, nil
}

// getCPUUsage returns estimated CPU usage percentage
func (m *SystemResourceMonitor) getCPUUsage() float64 {
	// Simplified CPU usage estimation
	// In production, would use more sophisticated monitoring like /proc/stat
	numCPU := float64(runtime.NumCPU())
	numGoroutine := float64(runtime.NumGoroutine())

	// Rough estimation based on goroutines and CPU count
	usage := (numGoroutine / (numCPU * 10)) * 100
	if usage > 100 {
		usage = 100
	}
	if usage < 5 {
		usage = 5 // Minimum baseline
	}

	return usage
}

// getMemoryUsage returns memory usage percentage
func (m *SystemResourceMonitor) getMemoryUsage(memStats *runtime.MemStats) float64 {
	// Use heap memory as a proxy for overall memory usage
	heapInUse := float64(memStats.HeapInuse)
	heapSys := float64(memStats.HeapSys)

	if heapSys == 0 {
		return 10.0 // Default reasonable value
	}

	usage := (heapInUse / heapSys) * 100
	if usage < 5 {
		usage = 5 // Minimum baseline
	}
	if usage > 100 {
		usage = 100
	}

	return usage
}

// getLoadAverages returns system load averages (Unix-like systems)
func (m *SystemResourceMonitor) getLoadAverages() (float64, float64, float64, error) {
	// Try to read /proc/loadavg on Linux
	if data, err := os.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 3 {
			load1, err1 := strconv.ParseFloat(fields[0], 64)
			load5, err2 := strconv.ParseFloat(fields[1], 64)
			load15, err3 := strconv.ParseFloat(fields[2], 64)

			if err1 == nil && err2 == nil && err3 == nil {
				return load1, load5, load15, nil
			}
		}
	}

	// Fallback: estimate based on CPU count and usage
	numCPU := float64(runtime.NumCPU())
	estimatedLoad := numCPU * 0.3 // Conservative estimate

	return estimatedLoad, estimatedLoad, estimatedLoad, nil
}

// LoadAwareScheduler implements load-aware adaptive scheduling
type LoadAwareScheduler struct {
	mu              sync.RWMutex
	baseInterval    time.Duration
	minInterval     time.Duration
	maxInterval     time.Duration
	targetCPU       float64
	targetMemory    float64
	targetLoad      float64
	currentInterval time.Duration
	monitor         *SystemResourceMonitor
	metricsHistory  []*SystemMetrics
	maxHistorySize  int
	nextChan        chan time.Time
	stopChan        chan struct{}
	stopped         bool
	mockMetrics     *SystemMetrics // For testing
}

// NewLoadAwareScheduler creates a new load-aware scheduler
func NewLoadAwareScheduler(baseInterval time.Duration, targetCPU, targetMemory, targetLoad float64) *LoadAwareScheduler {
	return NewLoadAwareSchedulerWithBounds(
		baseInterval,
		targetCPU,
		targetMemory,
		targetLoad,
		baseInterval/10, // min = base/10
		baseInterval*10, // max = base*10
	)
}

// NewLoadAwareSchedulerWithBounds creates a load-aware scheduler with custom bounds
func NewLoadAwareSchedulerWithBounds(baseInterval time.Duration, targetCPU, targetMemory, targetLoad float64, minInterval, maxInterval time.Duration) *LoadAwareScheduler {
	s := &LoadAwareScheduler{
		baseInterval:    baseInterval,
		minInterval:     minInterval,
		maxInterval:     maxInterval,
		targetCPU:       targetCPU,
		targetMemory:    targetMemory,
		targetLoad:      targetLoad,
		currentInterval: baseInterval,
		monitor:         NewSystemResourceMonitor(),
		metricsHistory:  make([]*SystemMetrics, 0),
		maxHistorySize:  100,
		nextChan:        make(chan time.Time, 1),
		stopChan:        make(chan struct{}),
		stopped:         false,
	}

	go s.scheduleLoop()
	return s
}

// SetMockMetrics sets mock metrics for testing
func (s *LoadAwareScheduler) SetMockMetrics(metrics *SystemMetrics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mockMetrics = metrics
}

// UpdateFromMetrics updates the scheduler based on current system metrics
func (s *LoadAwareScheduler) UpdateFromMetrics() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var metrics *SystemMetrics
	var err error

	if s.mockMetrics != nil {
		// Use mock metrics for testing
		metrics = s.mockMetrics
	} else {
		// Get real system metrics
		metrics, err = s.monitor.GetCurrentMetrics()
		if err != nil {
			return fmt.Errorf("failed to get system metrics: %w", err)
		}
	}

	// Add to history
	s.metricsHistory = append(s.metricsHistory, metrics)
	if len(s.metricsHistory) > s.maxHistorySize {
		s.metricsHistory = s.metricsHistory[1:]
	}

	// Calculate load factor based on current vs target metrics
	cpuFactor := metrics.CPUUsage / s.targetCPU
	memoryFactor := metrics.MemoryUsage / s.targetMemory
	loadFactor := metrics.LoadAverage1m / s.targetLoad

	// Use the highest factor (most constrained resource)
	maxFactor := math.Max(cpuFactor, math.Max(memoryFactor, loadFactor))

	// Adjust interval based on load factor
	// If factor > 1, system is overloaded, increase interval
	// If factor < 1, system has capacity, decrease interval
	newInterval := time.Duration(float64(s.baseInterval) * maxFactor)

	// Apply bounds
	if newInterval < s.minInterval {
		newInterval = s.minInterval
	}
	if newInterval > s.maxInterval {
		newInterval = s.maxInterval
	}

	s.currentInterval = newInterval
	return nil
}

// GetCurrentInterval returns the current interval
func (s *LoadAwareScheduler) GetCurrentInterval() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentInterval
}

// GetMetricsHistory returns the metrics history
func (s *LoadAwareScheduler) GetMetricsHistory() []*SystemMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid race conditions
	history := make([]*SystemMetrics, len(s.metricsHistory))
	copy(history, s.metricsHistory)
	return history
}

// Next returns a channel that delivers the next execution time
func (s *LoadAwareScheduler) Next() <-chan time.Time {
	return s.nextChan
}

// scheduleLoop continuously schedules the next execution based on load-aware intervals
func (s *LoadAwareScheduler) scheduleLoop() {
	ticker := time.NewTicker(5 * time.Second) // Update metrics every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			// Update metrics and adjust interval
			_ = s.UpdateFromMetrics()
		default:
			// Get current interval and wait
			interval := s.GetCurrentInterval()

			select {
			case <-time.After(interval):
				select {
				case s.nextChan <- time.Now():
					// Successfully sent
				case <-s.stopChan:
					return
				}
			case <-s.stopChan:
				return
			}
		}
	}
}

// Stop stops the scheduler
func (s *LoadAwareScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.stopped {
		close(s.stopChan)
		s.stopped = true
	}
}
