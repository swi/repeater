package httpaware

import (
	"testing"
	"time"

	"github.com/swi/repeater/pkg/scheduler"
)

func TestHTTPAwareScheduler_SchedulingBehavior(t *testing.T) {
	tests := []struct {
		name              string
		httpResponse      string
		fallbackScheduler scheduler.Scheduler
		expectedDelay     time.Duration
		description       string
	}{
		{
			name: "http_timing_overrides_fallback",
			httpResponse: "HTTP/1.1 503 Service Unavailable\r\n" +
				"Retry-After: 45\r\n" +
				"\r\n",
			fallbackScheduler: createIntervalScheduler(10 * time.Second),
			expectedDelay:     45 * time.Second,
			description:       "HTTP timing should override fallback scheduler",
		},
		{
			name: "no_http_timing_uses_fallback",
			httpResponse: "HTTP/1.1 200 OK\r\n" +
				"Content-Type: text/plain\r\n" +
				"\r\n" +
				"Service is healthy",
			fallbackScheduler: createIntervalScheduler(30 * time.Second),
			expectedDelay:     30 * time.Second,
			description:       "No HTTP timing should use fallback scheduler",
		},
		{
			name: "rate_limit_timing_respected",
			httpResponse: "HTTP/1.1 429 Too Many Requests\r\n" +
				"Retry-After: 120\r\n" +
				"\r\n",
			fallbackScheduler: createIntervalScheduler(5 * time.Second),
			expectedDelay:     120 * time.Second,
			description:       "Rate limiting should be respected",
		},
		{
			name: "json_timing_used",
			httpResponse: "HTTP/1.1 503 Service Unavailable\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"retry_after": 60}`,
			fallbackScheduler: createIntervalScheduler(15 * time.Second),
			expectedDelay:     60 * time.Second,
			description:       "JSON timing should be used",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HTTPAwareConfig{
				MaxDelay: 10 * time.Minute,
				MinDelay: 1 * time.Second,
			}

			httpScheduler := NewHTTPAwareScheduler(config)
			httpScheduler.SetFallbackScheduler(tt.fallbackScheduler)
			httpScheduler.SetLastResponse(tt.httpResponse)

			delay := httpScheduler.NextDelay()
			if delay != tt.expectedDelay {
				t.Errorf("Expected delay %v, got %v", tt.expectedDelay, delay)
			}
		})
	}
}

func TestHTTPAwareScheduler_MaxDelayCapApplied(t *testing.T) {
	config := HTTPAwareConfig{
		MaxDelay: 30 * time.Minute,
		MinDelay: 1 * time.Second,
	}

	httpScheduler := NewHTTPAwareScheduler(config)
	fallbackScheduler := createIntervalScheduler(10 * time.Second)
	httpScheduler.SetFallbackScheduler(fallbackScheduler)

	// Response with 2 hour delay (7200 seconds)
	httpResponse := "HTTP/1.1 503 Service Unavailable\r\n" +
		"Retry-After: 7200\r\n" +
		"\r\n"

	httpScheduler.SetLastResponse(httpResponse)
	delay := httpScheduler.NextDelay()

	expectedDelay := 30 * time.Minute
	if delay != expectedDelay {
		t.Errorf("Expected delay to be capped at %v, got %v", expectedDelay, delay)
	}
}

func TestHTTPAwareScheduler_MinDelayEnforced(t *testing.T) {
	config := HTTPAwareConfig{
		MaxDelay: 10 * time.Minute,
		MinDelay: 5 * time.Second,
	}

	httpScheduler := NewHTTPAwareScheduler(config)
	fallbackScheduler := createIntervalScheduler(10 * time.Second)
	httpScheduler.SetFallbackScheduler(fallbackScheduler)

	// Response with zero delay
	httpResponse := "HTTP/1.1 503 Service Unavailable\r\n" +
		"Retry-After: 0\r\n" +
		"\r\n"

	httpScheduler.SetLastResponse(httpResponse)
	delay := httpScheduler.NextDelay()

	expectedDelay := 5 * time.Second
	if delay != expectedDelay {
		t.Errorf("Expected delay to be at least %v, got %v", expectedDelay, delay)
	}
}

func TestHTTPAwareScheduler_FallbackIntegration(t *testing.T) {
	config := HTTPAwareConfig{
		MaxDelay: 5 * time.Minute,
		MinDelay: 1 * time.Second,
	}

	httpScheduler := NewHTTPAwareScheduler(config)

	// Create a mock exponential backoff scheduler
	fallbackScheduler := createExponentialScheduler(1*time.Second, 2.0, 60*time.Second)
	httpScheduler.SetFallbackScheduler(fallbackScheduler)

	testCases := []struct {
		response      string
		expectedDelay time.Duration
		description   string
	}{
		{
			response:      "HTTP/1.1 200 OK\r\n\r\n", // No HTTP timing
			expectedDelay: 1 * time.Second,           // Exponential: 1s
			description:   "Should use exponential backoff for success response",
		},
		{
			response:      "HTTP/1.1 503 Service Unavailable\r\nRetry-After: 30\r\n\r\n",
			expectedDelay: 30 * time.Second, // HTTP timing
			description:   "Should use HTTP timing for server error",
		},
		{
			response:      "HTTP/1.1 200 OK\r\n\r\n", // No HTTP timing again
			expectedDelay: 2 * time.Second,           // Exponential: 2s (continues sequence)
			description:   "Should continue exponential sequence after HTTP timing",
		},
	}

	for i, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			httpScheduler.SetLastResponse(tc.response)
			delay := httpScheduler.NextDelay()

			if delay != tc.expectedDelay {
				t.Errorf("Test case %d: Expected delay %v, got %v", i+1, tc.expectedDelay, delay)
			}
		})
	}
}

func TestHTTPAwareScheduler_GetTimingInfo(t *testing.T) {
	config := HTTPAwareConfig{
		MaxDelay: 10 * time.Minute,
		MinDelay: 1 * time.Second,
	}

	httpScheduler := NewHTTPAwareScheduler(config)
	fallbackScheduler := createIntervalScheduler(30 * time.Second)
	httpScheduler.SetFallbackScheduler(fallbackScheduler)

	// Test with HTTP timing
	httpResponse := "HTTP/1.1 503 Service Unavailable\r\n" +
		"Retry-After: 45\r\n" +
		"\r\n"

	httpScheduler.SetLastResponse(httpResponse)
	httpScheduler.NextDelay() // Trigger parsing

	timingInfo := httpScheduler.GetTimingInfo()
	if timingInfo == nil {
		t.Errorf("Expected timing info, got nil")
		return
	}

	if timingInfo.Delay != 45*time.Second {
		t.Errorf("Expected timing info delay %v, got %v", 45*time.Second, timingInfo.Delay)
	}

	if timingInfo.Source != TimingSourceRetryAfterHeader {
		t.Errorf("Expected timing source %v, got %v", TimingSourceRetryAfterHeader, timingInfo.Source)
	}

	// Test with no HTTP timing
	noTimingResponse := "HTTP/1.1 200 OK\r\n\r\n"
	httpScheduler.SetLastResponse(noTimingResponse)
	httpScheduler.NextDelay() // Trigger parsing

	timingInfo = httpScheduler.GetTimingInfo()
	if timingInfo != nil {
		t.Errorf("Expected no timing info for success response, got: %+v", timingInfo)
	}
}

func TestHTTPAwareScheduler_NonHTTPResponse(t *testing.T) {
	config := HTTPAwareConfig{
		MaxDelay: 10 * time.Minute,
		MinDelay: 1 * time.Second,
	}

	httpScheduler := NewHTTPAwareScheduler(config)
	fallbackScheduler := createIntervalScheduler(25 * time.Second)
	httpScheduler.SetFallbackScheduler(fallbackScheduler)

	// Test with non-HTTP response
	nonHTTPResponse := "Service is healthy\nUptime: 24 hours"
	httpScheduler.SetLastResponse(nonHTTPResponse)

	delay := httpScheduler.NextDelay()
	expectedDelay := 25 * time.Second

	if delay != expectedDelay {
		t.Errorf("Expected fallback delay %v for non-HTTP response, got %v", expectedDelay, delay)
	}

	timingInfo := httpScheduler.GetTimingInfo()
	if timingInfo != nil {
		t.Errorf("Expected no timing info for non-HTTP response, got: %+v", timingInfo)
	}
}

// Helper functions to create mock schedulers for testing

func createIntervalScheduler(interval time.Duration) scheduler.Scheduler {
	return &mockIntervalScheduler{interval: interval}
}

func createExponentialScheduler(baseDelay time.Duration, multiplier float64, maxDelay time.Duration) scheduler.Scheduler {
	return &mockExponentialScheduler{
		baseDelay:    baseDelay,
		multiplier:   multiplier,
		maxDelay:     maxDelay,
		currentDelay: baseDelay,
	}
}

// Mock schedulers for testing

type mockIntervalScheduler struct {
	interval time.Duration
	nextCh   chan time.Time
	stopped  bool
}

func (m *mockIntervalScheduler) Next() <-chan time.Time {
	if m.nextCh == nil {
		m.nextCh = make(chan time.Time, 1)
	}
	if !m.stopped {
		// Send immediate tick for testing
		select {
		case m.nextCh <- time.Now():
		default:
		}
	}
	return m.nextCh
}

func (m *mockIntervalScheduler) Stop() {
	m.stopped = true
	if m.nextCh != nil {
		close(m.nextCh)
	}
}

func (m *mockIntervalScheduler) GetNextDelay() time.Duration {
	return m.interval
}

type mockExponentialScheduler struct {
	baseDelay    time.Duration
	multiplier   float64
	maxDelay     time.Duration
	currentDelay time.Duration
	nextCh       chan time.Time
	stopped      bool
}

func (m *mockExponentialScheduler) Next() <-chan time.Time {
	if m.nextCh == nil {
		m.nextCh = make(chan time.Time, 1)
	}
	if !m.stopped {
		// Send immediate tick for testing
		select {
		case m.nextCh <- time.Now():
		default:
		}
	}
	return m.nextCh
}

func (m *mockExponentialScheduler) Stop() {
	m.stopped = true
	if m.nextCh != nil {
		close(m.nextCh)
	}
}

func (m *mockExponentialScheduler) GetNextDelay() time.Duration {
	delay := m.currentDelay
	m.currentDelay = time.Duration(float64(m.currentDelay) * m.multiplier)
	if m.currentDelay > m.maxDelay {
		m.currentDelay = m.maxDelay
	}
	return delay
}
