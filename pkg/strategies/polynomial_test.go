package strategies

import (
	"errors"
	"testing"
	"time"
)

func TestPolynomialStrategy_NextDelay(t *testing.T) {
	tests := []struct {
		name        string
		baseDelay   time.Duration
		exponent    float64
		maxDelay    time.Duration
		attempts    int
		expectedSeq []time.Duration
	}{
		{
			name:      "quadratic polynomial (exponent 2.0)",
			baseDelay: 1 * time.Second,
			exponent:  2.0,
			maxDelay:  0, // no cap
			attempts:  5,
			expectedSeq: []time.Duration{
				1 * time.Second,  // attempt 1: 1s * 1^2 = 1s
				4 * time.Second,  // attempt 2: 1s * 2^2 = 4s
				9 * time.Second,  // attempt 3: 1s * 3^2 = 9s
				16 * time.Second, // attempt 4: 1s * 4^2 = 16s
				25 * time.Second, // attempt 5: 1s * 5^2 = 25s
			},
		},
		{
			name:      "cubic polynomial (exponent 3.0)",
			baseDelay: 500 * time.Millisecond,
			exponent:  3.0,
			maxDelay:  0, // no cap
			attempts:  4,
			expectedSeq: []time.Duration{
				500 * time.Millisecond,   // attempt 1: 500ms * 1^3 = 500ms
				4000 * time.Millisecond,  // attempt 2: 500ms * 2^3 = 4000ms
				13500 * time.Millisecond, // attempt 3: 500ms * 3^3 = 13500ms
				32000 * time.Millisecond, // attempt 4: 500ms * 4^3 = 32000ms
			},
		},
		{
			name:      "square root polynomial (exponent 0.5)",
			baseDelay: 2 * time.Second,
			exponent:  0.5,
			maxDelay:  0, // no cap
			attempts:  4,
			expectedSeq: []time.Duration{
				2 * time.Second, // attempt 1: 2s * 1^0.5 = 2s
				time.Duration(2.0 * 1.414 * float64(time.Second)), // attempt 2: 2s * 2^0.5 ≈ 2.83s
				time.Duration(2.0 * 1.732 * float64(time.Second)), // attempt 3: 2s * 3^0.5 ≈ 3.46s
				4 * time.Second, // attempt 4: 2s * 4^0.5 = 4s
			},
		},
		{
			name:      "linear polynomial (exponent 1.0)",
			baseDelay: 1 * time.Second,
			exponent:  1.0,
			maxDelay:  0, // no cap
			attempts:  5,
			expectedSeq: []time.Duration{
				1 * time.Second, // attempt 1: 1s * 1^1 = 1s
				2 * time.Second, // attempt 2: 1s * 2^1 = 2s
				3 * time.Second, // attempt 3: 1s * 3^1 = 3s
				4 * time.Second, // attempt 4: 1s * 4^1 = 4s
				5 * time.Second, // attempt 5: 1s * 5^1 = 5s
			},
		},
		{
			name:      "with maximum delay cap",
			baseDelay: 1 * time.Second,
			exponent:  2.0,
			maxDelay:  10 * time.Second,
			attempts:  5,
			expectedSeq: []time.Duration{
				1 * time.Second,  // attempt 1: 1s * 1^2 = 1s
				4 * time.Second,  // attempt 2: 1s * 2^2 = 4s
				9 * time.Second,  // attempt 3: 1s * 3^2 = 9s
				10 * time.Second, // attempt 4: 1s * 4^2 = 16s (capped to 10s)
				10 * time.Second, // attempt 5: 1s * 5^2 = 25s (capped to 10s)
			},
		},
		{
			name:      "high exponent with small base",
			baseDelay: 100 * time.Millisecond,
			exponent:  1.5,
			maxDelay:  0, // no cap
			attempts:  4,
			expectedSeq: []time.Duration{
				100 * time.Millisecond,                                 // attempt 1: 100ms * 1^1.5 = 100ms
				time.Duration(100 * 2.828 * float64(time.Millisecond)), // attempt 2: 100ms * 2^1.5 ≈ 283ms
				time.Duration(100 * 5.196 * float64(time.Millisecond)), // attempt 3: 100ms * 3^1.5 ≈ 520ms
				800 * time.Millisecond,                                 // attempt 4: 100ms * 4^1.5 = 800ms
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := NewPolynomialStrategy(tt.baseDelay, tt.exponent, tt.maxDelay)

			for i, expected := range tt.expectedSeq {
				attempt := i + 1
				actual := strategy.NextDelay(attempt, 0)

				// Allow for small floating point variations
				tolerance := time.Duration(float64(expected) * 0.01) // 1% tolerance
				if tolerance < time.Millisecond {
					tolerance = time.Millisecond
				}

				diff := actual - expected
				if diff < 0 {
					diff = -diff
				}

				if diff > tolerance {
					t.Errorf("NextDelay(attempt=%d) = %v, expected %v (tolerance: %v)",
						attempt, actual, expected, tolerance)
				}
			}
		})
	}
}

func TestPolynomialStrategy_NextDelay_EdgeCases(t *testing.T) {
	strategy := NewPolynomialStrategy(1*time.Second, 2.0, 10*time.Second)

	// Test zero attempt
	delay := strategy.NextDelay(0, 0)
	if delay != 1*time.Second {
		t.Errorf("NextDelay(0) = %v, expected %v", delay, 1*time.Second)
	}

	// Test negative attempt (should be treated as zero/one)
	delay = strategy.NextDelay(-1, 0)
	if delay != 1*time.Second {
		t.Errorf("NextDelay(-1) = %v, expected %v", delay, 1*time.Second)
	}

	// Test very large attempt number (should be capped by maxDelay)
	delay = strategy.NextDelay(100, 0)
	if delay != 10*time.Second {
		t.Errorf("NextDelay(100) = %v, expected %v (max cap)", delay, 10*time.Second)
	}
}

func TestPolynomialStrategy_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *StrategyConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Exponent:    2.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: false,
		},
		{
			name: "zero base delay",
			config: &StrategyConfig{
				BaseDelay:   0,
				Exponent:    2.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: true,
			errorMsg:    "base-delay must be positive",
		},
		{
			name: "negative base delay",
			config: &StrategyConfig{
				BaseDelay:   -1 * time.Second,
				Exponent:    2.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: true,
			errorMsg:    "base-delay must be positive",
		},
		{
			name: "zero exponent",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Exponent:    0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: true,
			errorMsg:    "exponent must be positive",
		},
		{
			name: "negative exponent",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Exponent:    -1.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: true,
			errorMsg:    "exponent must be positive",
		},
		{
			name: "exponent too large",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Exponent:    11.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: true,
			errorMsg:    "exponent must be <= 10 to prevent overflow",
		},
		{
			name: "max delay less than base delay",
			config: &StrategyConfig{
				BaseDelay:   10 * time.Second,
				Exponent:    2.0,
				MaxDelay:    5 * time.Second,
				MaxAttempts: 5,
			},
			expectError: true,
			errorMsg:    "max-delay must be greater than base-delay",
		},
		{
			name: "zero max attempts",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Exponent:    2.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 0,
			},
			expectError: true,
			errorMsg:    "attempts must be positive",
		},
		{
			name: "negative max attempts",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Exponent:    2.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: -1,
			},
			expectError: true,
			errorMsg:    "attempts must be positive",
		},
		{
			name: "no max delay (unlimited)",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Exponent:    1.5,
				MaxDelay:    0, // unlimited
				MaxAttempts: 3,
			},
			expectError: false,
		},
		{
			name: "fractional exponent",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Exponent:    0.5,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: false,
		},
		{
			name: "boundary exponent (exactly 10)",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Exponent:    10.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := NewPolynomialStrategy(tt.config.BaseDelay, tt.config.Exponent, tt.config.MaxDelay)
			err := strategy.ValidateConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestPolynomialStrategy_Name(t *testing.T) {
	strategy := NewPolynomialStrategy(1*time.Second, 2.0, 60*time.Second)
	if strategy.Name() != "polynomial" {
		t.Errorf("Name() = %v, expected 'polynomial'", strategy.Name())
	}
}

func TestPolynomialStrategy_ShouldRetry(t *testing.T) {
	strategy := NewPolynomialStrategy(1*time.Second, 2.0, 60*time.Second)

	// ShouldRetry should always return true for polynomial strategy
	// (retry logic is handled by the main retry system)
	tests := []struct {
		name    string
		attempt int
		err     error
		output  string
	}{
		{"first attempt with error", 1, errors.New("test error"), "error output"},
		{"second attempt with error", 2, errors.New("another error"), "more error output"},
		{"first attempt without error", 1, nil, "success output"},
		{"high attempt count", 10, errors.New("persistent error"), "error output"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldRetry := strategy.ShouldRetry(tt.attempt, tt.err, tt.output)
			if !shouldRetry {
				t.Errorf("ShouldRetry() = false, expected true (polynomial strategy should always return true)")
			}
		})
	}
}

func TestPolynomialStrategy_RealWorldScenarios(t *testing.T) {
	t.Run("API retry with moderate growth", func(t *testing.T) {
		// Simulate API retries with 1.5 exponent for moderate growth
		strategy := NewPolynomialStrategy(500*time.Millisecond, 1.5, 5*time.Second)

		delays := make([]time.Duration, 6)
		for i := 0; i < 6; i++ {
			delays[i] = strategy.NextDelay(i+1, 0)
		}

		// Verify growth pattern is reasonable for API retries
		if delays[0] != 500*time.Millisecond {
			t.Errorf("First delay should be base delay")
		}

		// Each delay should be larger than previous (until cap)
		for i := 1; i < len(delays)-1; i++ {
			if delays[i] <= delays[i-1] && delays[i] < 5*time.Second {
				t.Errorf("Delay[%d] (%v) should be greater than delay[%d] (%v) unless capped",
					i, delays[i], i-1, delays[i-1])
			}
		}

		// Should eventually hit cap (check last few delays)
		foundCappedDelay := false
		for _, delay := range delays[3:] { // check last 3 delays
			if delay == 5*time.Second {
				foundCappedDelay = true
				break
			}
		}
		if !foundCappedDelay {
			t.Errorf("Should hit max delay cap of 5s, delays: %v", delays)
		}
	})

	t.Run("database connection retry with quadratic growth", func(t *testing.T) {
		// Simulate database reconnection with quadratic growth
		strategy := NewPolynomialStrategy(1*time.Second, 2.0, 0) // no cap

		// Test first few delays match quadratic pattern
		expected := []time.Duration{1 * time.Second, 4 * time.Second, 9 * time.Second, 16 * time.Second}
		for i, exp := range expected {
			actual := strategy.NextDelay(i+1, 0)
			if actual != exp {
				t.Errorf("Attempt %d: got %v, expected %v", i+1, actual, exp)
			}
		}
	})
}
