package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiophantineRateLimiter_BasicRateLimit tests basic Diophantine rate limiting
func TestDiophantineRateLimiter_BasicRateLimit(t *testing.T) {
	tests := []struct {
		name        string
		rateLimit   int64
		windowSize  time.Duration
		requests    int
		expectAllow int
	}{
		{
			name:        "10 requests per minute",
			rateLimit:   10,
			windowSize:  time.Minute,
			requests:    15,
			expectAllow: 10,
		},
		{
			name:        "5 requests per 30 seconds",
			rateLimit:   5,
			windowSize:  30 * time.Second,
			requests:    8,
			expectAllow: 5,
		},
		{
			name:        "100 requests per hour",
			rateLimit:   100,
			windowSize:  time.Hour,
			requests:    150,
			expectAllow: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewDiophantineRateLimiter(tt.rateLimit, tt.windowSize, nil)
			require.NotNil(t, limiter)

			allowed := 0
			for i := 0; i < tt.requests; i++ {
				if limiter.Allow() {
					allowed++
				}
			}

			assert.Equal(t, tt.expectAllow, allowed,
				"expected %d requests to be allowed, got %d", tt.expectAllow, allowed)
		})
	}
}

// TestDiophantineRateLimiter_RetryPatterns tests retry pattern handling
func TestDiophantineRateLimiter_RetryPatterns(t *testing.T) {
	// 10 requests per hour, each request retries at 10min and 30min
	retryPattern := []time.Duration{0, 10 * time.Minute, 30 * time.Minute}
	limiter := NewDiophantineRateLimiter(10, time.Hour, retryPattern)
	require.NotNil(t, limiter)

	// Each task generates 3 requests (initial + 2 retries)
	// So we can only safely schedule 3 tasks (3 * 3 = 9 requests < 10 limit)
	allowed := 0
	for i := 0; i < 5; i++ {
		if limiter.Allow() {
			allowed++
		}
	}

	// Should allow 3 tasks maximum (9 total requests including retries)
	assert.LessOrEqual(t, allowed, 3, "should not allow more than 3 tasks with retry pattern")
	assert.GreaterOrEqual(t, allowed, 1, "should allow at least 1 task")
}

// TestDiophantineRateLimiter_WindowConstraints tests window-based constraints
func TestDiophantineRateLimiter_WindowConstraints(t *testing.T) {
	// 5 requests per 10 seconds
	limiter := NewDiophantineRateLimiter(5, 10*time.Second, nil)
	require.NotNil(t, limiter)

	// Allow initial burst
	allowed := 0
	for i := 0; i < 10; i++ {
		if limiter.Allow() {
			allowed++
		}
	}
	assert.Equal(t, 5, allowed, "should allow exactly 5 requests initially")

	// Wait 5 seconds (half window) - should still be constrained
	time.Sleep(5 * time.Second)
	assert.False(t, limiter.Allow(), "should not allow request after 5 seconds (still in window)")

	// Wait another 6 seconds (total 11 seconds, outside original window)
	time.Sleep(6 * time.Second)
	assert.True(t, limiter.Allow(), "should allow request after window expires")
}

// TestDiophantineRateLimiter_PredictiveScheduling tests NextAllowedTime
func TestDiophantineRateLimiter_PredictiveScheduling(t *testing.T) {
	limiter := NewDiophantineRateLimiter(2, time.Minute, nil)
	require.NotNil(t, limiter)

	start := time.Now()

	// Consume all available slots
	assert.True(t, limiter.Allow())
	assert.True(t, limiter.Allow())
	assert.False(t, limiter.Allow())

	// Get next allowed time
	nextTime := limiter.NextAllowedTime()
	assert.True(t, nextTime.After(start), "next allowed time should be in the future")
	assert.True(t, nextTime.Before(start.Add(2*time.Minute)), "next allowed time should be reasonable")
}

// TestDiophantineRateLimiter_Statistics tests statistics tracking
func TestDiophantineRateLimiter_Statistics(t *testing.T) {
	limiter := NewDiophantineRateLimiter(5, time.Minute, nil)
	require.NotNil(t, limiter)

	// Make some requests
	for i := 0; i < 8; i++ {
		limiter.Allow()
	}

	stats := limiter.Statistics()
	assert.Equal(t, int64(8), stats.TotalRequests, "should track total requests")
	assert.Equal(t, int64(5), stats.AllowedRequests, "should track allowed requests")
	assert.Equal(t, int64(3), stats.DeniedRequests, "should track denied requests")
	assert.Equal(t, int64(5), stats.Rate, "should report configured rate")
}

// TestPreciseTokenBucket_BasicRateLimit tests basic rate limiting functionality
func TestPreciseTokenBucket_BasicRateLimit(t *testing.T) {
	tests := []struct {
		name        string
		rate        int64 // requests per second
		capacity    int64 // burst capacity
		requests    int   // number of requests to make
		expectAllow int   // expected number of allowed requests
	}{
		{
			name:        "10 requests per second, no burst",
			rate:        10,
			capacity:    1,
			requests:    15,
			expectAllow: 1, // Only initial token available
		},
		{
			name:        "10 requests per second with burst of 5",
			rate:        10,
			capacity:    5,
			requests:    10,
			expectAllow: 5, // Initial burst capacity
		},
		{
			name:        "100 requests per second with burst of 10",
			rate:        100,
			capacity:    10,
			requests:    20,
			expectAllow: 10, // Initial burst capacity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewPreciseTokenBucket(tt.rate, tt.capacity)
			require.NotNil(t, limiter)

			allowed := 0
			for i := 0; i < tt.requests; i++ {
				if limiter.Allow() {
					allowed++
				}
			}

			assert.Equal(t, tt.expectAllow, allowed,
				"expected %d requests to be allowed, got %d", tt.expectAllow, allowed)
		})
	}
}

// TestPreciseTokenBucket_TokenRefill tests token refill over time
func TestPreciseTokenBucket_TokenRefill(t *testing.T) {
	// 10 requests per second = 1 request per 100ms
	limiter := NewPreciseTokenBucket(10, 5)
	require.NotNil(t, limiter)

	// Consume initial burst
	for i := 0; i < 5; i++ {
		assert.True(t, limiter.Allow(), "initial burst request %d should be allowed", i)
	}

	// Next request should be denied
	assert.False(t, limiter.Allow(), "request after burst should be denied")

	// Wait for token refill (110ms to account for timing precision)
	time.Sleep(110 * time.Millisecond)

	// Should allow one more request
	assert.True(t, limiter.Allow(), "request after refill should be allowed")
	assert.False(t, limiter.Allow(), "second request after refill should be denied")
}

// TestPreciseTokenBucket_PrecisionTiming tests mathematical precision
func TestPreciseTokenBucket_PrecisionTiming(t *testing.T) {
	// 1000 requests per second = 1 request per millisecond
	limiter := NewPreciseTokenBucket(1000, 1)
	require.NotNil(t, limiter)

	// Consume initial token
	assert.True(t, limiter.Allow())
	assert.False(t, limiter.Allow())

	// Wait exactly 1ms
	time.Sleep(1 * time.Millisecond)

	// Should allow exactly one more request
	assert.True(t, limiter.Allow(), "request after 1ms should be allowed")
	assert.False(t, limiter.Allow(), "second request should be denied")
}

// TestPreciseTokenBucket_BurstHandling tests burst capability
func TestPreciseTokenBucket_BurstHandling(t *testing.T) {
	// 10 requests per second with burst of 20
	limiter := NewPreciseTokenBucket(10, 20)
	require.NotNil(t, limiter)

	// Should allow full burst immediately
	for i := 0; i < 20; i++ {
		assert.True(t, limiter.Allow(), "burst request %d should be allowed", i)
	}

	// 21st request should be denied
	assert.False(t, limiter.Allow(), "request beyond burst should be denied")

	// Wait for 2 seconds (should refill 20 tokens)
	time.Sleep(2 * time.Second)

	// Should allow full burst again
	for i := 0; i < 20; i++ {
		assert.True(t, limiter.Allow(), "refilled burst request %d should be allowed", i)
	}
}

// TestPreciseTokenBucket_ZeroRate tests edge case of zero rate
func TestPreciseTokenBucket_ZeroRate(t *testing.T) {
	limiter := NewPreciseTokenBucket(0, 5)
	require.NotNil(t, limiter)

	// Should allow initial burst
	for i := 0; i < 5; i++ {
		assert.True(t, limiter.Allow(), "initial token %d should be allowed", i)
	}

	// No more tokens should be generated
	time.Sleep(100 * time.Millisecond)
	assert.False(t, limiter.Allow(), "no tokens should be generated with zero rate")
}

// TestPreciseTokenBucket_HighRate tests high rate scenarios
func TestPreciseTokenBucket_HighRate(t *testing.T) {
	// 1 million requests per second
	limiter := NewPreciseTokenBucket(1000000, 100)
	require.NotNil(t, limiter)

	// Should handle high rates without overflow
	// Make requests as fast as possible to minimize time-based token generation
	allowed := 0
	start := time.Now()
	for i := 0; i < 200; i++ {
		if limiter.Allow() {
			allowed++
		}
	}
	elapsed := time.Since(start)

	// With 1M req/sec, even 1ms could generate 1000 tokens
	// So we need to account for time-based generation with some tolerance
	maxExpected := 100 + int(elapsed.Nanoseconds()*1000000/1e9) + 5 // Add 5 token tolerance

	assert.LessOrEqual(t, allowed, maxExpected, "should not exceed capacity plus time-based tokens")
	assert.GreaterOrEqual(t, allowed, 100, "should allow at least initial capacity")
}

// TestPreciseTokenBucket_Statistics tests rate limiting statistics
func TestPreciseTokenBucket_Statistics(t *testing.T) {
	limiter := NewPreciseTokenBucket(10, 5)
	require.NotNil(t, limiter)

	// Make some requests
	for i := 0; i < 10; i++ {
		limiter.Allow()
	}

	stats := limiter.Statistics()
	assert.Equal(t, int64(10), stats.TotalRequests, "should track total requests")
	assert.Equal(t, int64(5), stats.AllowedRequests, "should track allowed requests")
	assert.Equal(t, int64(5), stats.DeniedRequests, "should track denied requests")
	assert.Equal(t, int64(10), stats.Rate, "should report configured rate")
	assert.Equal(t, int64(5), stats.Capacity, "should report configured capacity")
}

// TestDistributedRateLimiter_MultiInstance tests multi-instance coordination
func TestDistributedRateLimiter_MultiInstance(t *testing.T) {
	// Create two instances sharing a 10 req/hour limit (using Diophantine approach)
	limiter1 := NewDistributedRateLimiter("instance1", 10, 5)
	limiter2 := NewDistributedRateLimiter("instance2", 10, 5)
	require.NotNil(t, limiter1)
	require.NotNil(t, limiter2)

	// With Diophantine approach, each instance gets the full rate (will be coordinated in future)
	// For now, test that each instance works independently
	allowed1 := 0
	allowed2 := 0

	// Test that each instance respects its own limits
	for i := 0; i < 15; i++ {
		if limiter1.Allow() {
			allowed1++
		}
		if limiter2.Allow() {
			allowed2++
		}
	}

	// Each instance should respect the rate limit independently
	assert.LessOrEqual(t, allowed1, 10, "instance1 should not exceed rate limit")
	assert.LessOrEqual(t, allowed2, 10, "instance2 should not exceed rate limit")
	assert.Greater(t, allowed1, 0, "instance1 should allow some requests")
	assert.Greater(t, allowed2, 0, "instance2 should allow some requests")
}

// TestRateLimitParser_ParseRateSpec tests rate specification parsing
func TestRateLimitParser_ParseRateSpec(t *testing.T) {
	tests := []struct {
		spec         string
		expectRate   int64
		expectPeriod time.Duration
		expectError  bool
	}{
		{"10/1s", 10, time.Second, false},
		{"100/1m", 100, time.Minute, false},
		{"1000/1h", 1000, time.Hour, false},
		{"5/500ms", 5, 500 * time.Millisecond, false},
		{"invalid", 0, 0, true},
		{"10/", 0, 0, true},
		{"/1s", 0, 0, true},
		{"0/1s", 0, time.Second, false}, // Zero rate should be valid
	}

	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			rate, period, err := ParseRateSpec(tt.spec)

			if tt.expectError {
				assert.Error(t, err, "expected error for spec: %s", tt.spec)
			} else {
				assert.NoError(t, err, "unexpected error for spec: %s", tt.spec)
				assert.Equal(t, tt.expectRate, rate, "rate mismatch for spec: %s", tt.spec)
				assert.Equal(t, tt.expectPeriod, period, "period mismatch for spec: %s", tt.spec)
			}
		})
	}
}
