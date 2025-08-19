package ratelimit

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DiophantineRateLimiter implements mathematically precise rate limiting using constraint satisfaction
// It ensures no time window exceeds the rate limit, preventing server overwhelm
type DiophantineRateLimiter struct {
	mu             sync.Mutex
	rateLimit      int64           // maximum requests per window
	windowSize     time.Duration   // time window for rate limiting
	retryPattern   []time.Duration // retry offsets (e.g., [0, 10m, 30m])
	scheduledTimes []time.Time     // all scheduled request times

	// Statistics
	totalRequests   int64
	allowedRequests int64
	deniedRequests  int64
}

// PreciseTokenBucket implements mathematically precise rate limiting using integer arithmetic
// DEPRECATED: Use DiophantineRateLimiter for server-friendly rate limiting
type PreciseTokenBucket struct {
	mu             sync.Mutex
	capacityTokens int64 // maximum tokens in bucket
	currentTokens  int64 // current available tokens
	ratePerNanosec int64 // tokens per nanosecond * 1e9 for precision
	lastUpdateNs   int64 // last update timestamp in nanoseconds
	tokenDebtNs    int64 // accumulated fractional tokens in nanoseconds

	// Statistics
	totalRequests   int64
	allowedRequests int64
	deniedRequests  int64
	rate            int64 // original rate for reporting
}

// Statistics holds rate limiting statistics
type Statistics struct {
	TotalRequests   int64
	AllowedRequests int64
	DeniedRequests  int64
	Rate            int64
	Capacity        int64
	CurrentTokens   int64
}

// NewDiophantineRateLimiter creates a new Diophantine constraint-based rate limiter
func NewDiophantineRateLimiter(rateLimit int64, windowSize time.Duration, retryPattern []time.Duration) *DiophantineRateLimiter {
	if retryPattern == nil {
		retryPattern = []time.Duration{0} // Default: single attempt, no retries
	}

	return &DiophantineRateLimiter{
		rateLimit:      rateLimit,
		windowSize:     windowSize,
		retryPattern:   retryPattern,
		scheduledTimes: make([]time.Time, 0),
	}
}

// NewPreciseTokenBucket creates a new precise token bucket rate limiter
// DEPRECATED: Use NewDiophantineRateLimiter for server-friendly rate limiting
func NewPreciseTokenBucket(rate, capacity int64) *PreciseTokenBucket {
	now := time.Now().UnixNano()

	return &PreciseTokenBucket{
		capacityTokens: capacity,
		currentTokens:  capacity, // Start with full bucket
		ratePerNanosec: rate,     // Will be converted to per-nanosecond in addTokens
		lastUpdateNs:   now,
		tokenDebtNs:    0,
		rate:           rate,
	}
}

// Allow checks if a request can be scheduled without violating any time window constraints
func (d *DiophantineRateLimiter) Allow() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.totalRequests++

	now := time.Now()

	// Clean up old scheduled times that are outside any relevant window
	d.cleanupOldTimes(now)

	// Check if we can safely schedule this request
	if d.canScheduleAt(now) {
		d.scheduledTimes = append(d.scheduledTimes, now)
		d.allowedRequests++
		return true
	}

	d.deniedRequests++
	return false
}

// canScheduleAt checks if a request can be scheduled at the given time without violating rate limits
func (d *DiophantineRateLimiter) canScheduleAt(requestTime time.Time) bool {
	// Count how many requests would be in the window starting at requestTime
	// This is the critical window that determines if we can schedule this request

	windowStart := requestTime
	windowEnd := requestTime.Add(d.windowSize)

	count := int64(0)

	// Count existing scheduled requests that would have attempts in this window
	for _, scheduledTime := range d.scheduledTimes {
		for _, offset := range d.retryPattern {
			attemptTime := scheduledTime.Add(offset)
			if !attemptTime.Before(windowStart) && attemptTime.Before(windowEnd) {
				count++
			}
		}
	}

	// Count attempts from this new request that would be in this window
	for _, offset := range d.retryPattern {
		attemptTime := requestTime.Add(offset)
		if !attemptTime.Before(windowStart) && attemptTime.Before(windowEnd) {
			count++
		}
	}

	// If this window would exceed the limit, reject the request
	if count > d.rateLimit {
		return false
	}

	// Also need to check if this new request would cause any existing windows to be violated
	// Check windows that start before requestTime but could include our new attempts
	for _, scheduledTime := range d.scheduledTimes {
		for _, existingOffset := range d.retryPattern {
			existingAttemptTime := scheduledTime.Add(existingOffset)
			// Check window starting at this existing attempt
			existingWindowStart := existingAttemptTime
			existingWindowEnd := existingAttemptTime.Add(d.windowSize)

			// Count all attempts in this existing window (including our new ones)
			existingWindowCount := int64(0)

			// Count existing attempts in this window
			for _, otherScheduledTime := range d.scheduledTimes {
				for _, otherOffset := range d.retryPattern {
					otherAttemptTime := otherScheduledTime.Add(otherOffset)
					if !otherAttemptTime.Before(existingWindowStart) && otherAttemptTime.Before(existingWindowEnd) {
						existingWindowCount++
					}
				}
			}

			// Count new attempts that would fall in this existing window
			for _, newOffset := range d.retryPattern {
				newAttemptTime := requestTime.Add(newOffset)
				if !newAttemptTime.Before(existingWindowStart) && newAttemptTime.Before(existingWindowEnd) {
					existingWindowCount++
				}
			}

			if existingWindowCount > d.rateLimit {
				return false
			}
		}
	}

	return true
}

// cleanupOldTimes removes scheduled times that are outside any relevant window
func (d *DiophantineRateLimiter) cleanupOldTimes(now time.Time) {
	cutoff := now.Add(-d.windowSize - d.maxRetryOffset())

	// Keep only times that could still affect future windows
	filtered := d.scheduledTimes[:0]
	for _, t := range d.scheduledTimes {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	d.scheduledTimes = filtered
}

// maxRetryOffset returns the maximum retry offset
func (d *DiophantineRateLimiter) maxRetryOffset() time.Duration {
	max := time.Duration(0)
	for _, offset := range d.retryPattern {
		if offset > max {
			max = offset
		}
	}
	return max
}

// NextAllowedTime calculates the earliest time a request can be safely scheduled
func (d *DiophantineRateLimiter) NextAllowedTime() time.Time {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	d.cleanupOldTimes(now)

	// Try scheduling at increasingly later times until we find a safe slot
	candidate := now
	increment := time.Second // Start with 1-second increments

	for attempts := 0; attempts < 3600; attempts++ { // Max 1 hour ahead
		if d.canScheduleAt(candidate) {
			return candidate
		}
		candidate = candidate.Add(increment)
	}

	// If we can't find a slot within an hour, return far future
	return now.Add(time.Hour)
}

// Statistics returns current rate limiting statistics for Diophantine limiter
func (d *DiophantineRateLimiter) Statistics() Statistics {
	d.mu.Lock()
	defer d.mu.Unlock()

	return Statistics{
		TotalRequests:   d.totalRequests,
		AllowedRequests: d.allowedRequests,
		DeniedRequests:  d.deniedRequests,
		Rate:            d.rateLimit,
		Capacity:        d.rateLimit,                                // For compatibility
		CurrentTokens:   d.rateLimit - int64(len(d.scheduledTimes)), // Approximate
	}
}

// Allow checks if a request should be allowed and consumes a token if so
// DEPRECATED: Use DiophantineRateLimiter.Allow() for server-friendly rate limiting
func (tb *PreciseTokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.totalRequests++

	now := time.Now().UnixNano()
	tb.addTokens(now)

	if tb.currentTokens >= 1 {
		tb.currentTokens--
		tb.allowedRequests++
		return true
	}

	tb.deniedRequests++
	return false
}

// addTokens adds tokens based on elapsed time using precise integer arithmetic
func (tb *PreciseTokenBucket) addTokens(now int64) {
	if now <= tb.lastUpdateNs {
		return
	}

	elapsed := now - tb.lastUpdateNs

	// Calculate new tokens using integer arithmetic to avoid floating point errors
	// rate is requests per second, so we need to convert to per nanosecond
	newTokensNumerator := elapsed * tb.ratePerNanosec
	newTokens := newTokensNumerator / 1e9
	tb.tokenDebtNs += newTokensNumerator % 1e9

	// Handle accumulated fractional tokens
	if tb.tokenDebtNs >= 1e9 {
		newTokens += tb.tokenDebtNs / 1e9
		tb.tokenDebtNs %= 1e9
	}

	// Add tokens but don't exceed capacity
	tb.currentTokens = min(tb.capacityTokens, tb.currentTokens+newTokens)
	tb.lastUpdateNs = now
}

// Statistics returns current rate limiting statistics
func (tb *PreciseTokenBucket) Statistics() Statistics {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Update tokens before reporting
	now := time.Now().UnixNano()
	tb.addTokens(now)

	return Statistics{
		TotalRequests:   tb.totalRequests,
		AllowedRequests: tb.allowedRequests,
		DeniedRequests:  tb.deniedRequests,
		Rate:            tb.rate,
		Capacity:        tb.capacityTokens,
		CurrentTokens:   tb.currentTokens,
	}
}

// DistributedRateLimiter coordinates rate limiting across multiple instances using Diophantine constraints
type DistributedRateLimiter struct {
	instanceID   string
	totalRate    int64
	instanceRate int64
	localLimiter *DiophantineRateLimiter
	// Future: Add coordination mechanism for multi-instance rate limiting
}

// NewDistributedRateLimiter creates a new distributed rate limiter using Diophantine approach
func NewDistributedRateLimiter(instanceID string, totalRate, capacity int64) *DistributedRateLimiter {
	// For now, assume single instance (will be enhanced in future cycles)
	instanceRate := totalRate
	windowSize := time.Hour            // Default 1-hour window
	retryPattern := []time.Duration{0} // Default: no retries

	return &DistributedRateLimiter{
		instanceID:   instanceID,
		totalRate:    totalRate,
		instanceRate: instanceRate,
		localLimiter: NewDiophantineRateLimiter(instanceRate, windowSize, retryPattern),
	}
}

// Allow checks if a request should be allowed across distributed instances
func (drl *DistributedRateLimiter) Allow() bool {
	// For now, just use local limiter (will be enhanced in future cycles)
	return drl.localLimiter.Allow()
}

// ParseRateSpec parses rate specifications like "10/1s", "100/1m", "1000/1h"
func ParseRateSpec(spec string) (rate int64, period time.Duration, err error) {
	parts := strings.Split(spec, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid rate spec format: %s (expected format: rate/period)", spec)
	}

	rateStr := strings.TrimSpace(parts[0])
	periodStr := strings.TrimSpace(parts[1])

	if rateStr == "" {
		return 0, 0, fmt.Errorf("missing rate in spec: %s", spec)
	}

	if periodStr == "" {
		return 0, 0, fmt.Errorf("missing period in spec: %s", spec)
	}

	// Parse rate
	rate, err = strconv.ParseInt(rateStr, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid rate '%s': %w", rateStr, err)
	}

	// Parse period
	period, err = time.ParseDuration(periodStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid period '%s': %w", periodStr, err)
	}

	return rate, period, nil
}

// min returns the minimum of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
