package strategies

import (
	"errors"
	"testing"
	"time"
)

func TestDecorrelatedJitterStrategy_NextDelay(t *testing.T) {
	tests := []struct {
		name              string
		baseDelay         time.Duration
		multiplier        float64
		maxDelay          time.Duration
		attempts          int
		minExpectedDelays []time.Duration // minimum expected delays
		maxExpectedDelays []time.Duration // maximum expected delays
	}{
		{
			name:       "basic decorrelated jitter",
			baseDelay:  1 * time.Second,
			multiplier: 3.0,
			maxDelay:   0, // no cap
			attempts:   4,
			minExpectedDelays: []time.Duration{
				1 * time.Second, // attempt 1: always base_delay
				1 * time.Second, // attempt 2: min(base_delay, previous * multiplier) = min(1s, 3s) = 1s
				1 * time.Second, // attempt 3: varies based on previous
				1 * time.Second, // attempt 4: varies based on previous
			},
			maxExpectedDelays: []time.Duration{
				1 * time.Second,  // attempt 1: always base_delay
				3 * time.Second,  // attempt 2: max is previous * multiplier
				9 * time.Second,  // attempt 3: could be up to 3s * 3 = 9s
				27 * time.Second, // attempt 4: theoretical max keeps growing
			},
		},
		{
			name:       "with maximum delay cap",
			baseDelay:  500 * time.Millisecond,
			multiplier: 2.0,
			maxDelay:   5 * time.Second,
			attempts:   6,
			minExpectedDelays: []time.Duration{
				500 * time.Millisecond, // attempt 1: always base_delay
				500 * time.Millisecond, // attempt 2: minimum is base_delay
				500 * time.Millisecond, // attempt 3+: minimum is base_delay
				500 * time.Millisecond,
				500 * time.Millisecond,
				500 * time.Millisecond,
			},
			maxExpectedDelays: []time.Duration{
				500 * time.Millisecond,  // attempt 1: always base_delay
				1000 * time.Millisecond, // attempt 2: 500ms * 2.0 = 1s
				2000 * time.Millisecond, // attempt 3: could be up to 1s * 2 = 2s
				4000 * time.Millisecond, // attempt 4: could be up to 2s * 2 = 4s
				5000 * time.Millisecond, // attempt 5: capped at 5s
				5000 * time.Millisecond, // attempt 6: capped at 5s
			},
		},
		{
			name:       "low multiplier",
			baseDelay:  2 * time.Second,
			multiplier: 1.1,
			maxDelay:   0, // no cap
			attempts:   3,
			minExpectedDelays: []time.Duration{
				2 * time.Second, // attempt 1: always base_delay
				2 * time.Second, // attempt 2: minimum is base_delay
				2 * time.Second, // attempt 3: minimum is base_delay
			},
			maxExpectedDelays: []time.Duration{
				2 * time.Second, // attempt 1: always base_delay
				time.Duration(2.2 * float64(time.Second)),  // attempt 2: 2s * 1.1 = 2.2s
				time.Duration(2.42 * float64(time.Second)), // attempt 3: theoretical max around 2.2s * 1.1 = 2.42s
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple iterations to test randomness behavior
			for iteration := 0; iteration < 10; iteration++ {
				strategy := NewDecorrelatedJitterStrategy(tt.baseDelay, tt.multiplier, tt.maxDelay)

				for attempt := 1; attempt <= tt.attempts; attempt++ {
					delay := strategy.NextDelay(attempt, 0)

					minExpected := tt.minExpectedDelays[attempt-1]
					maxExpected := tt.maxExpectedDelays[attempt-1]

					if delay < minExpected {
						t.Errorf("Iteration %d, attempt %d: delay %v is less than minimum expected %v",
							iteration, attempt, delay, minExpected)
					}
					if delay > maxExpected {
						t.Errorf("Iteration %d, attempt %d: delay %v is greater than maximum expected %v",
							iteration, attempt, delay, maxExpected)
					}
				}
			}
		})
	}
}

func TestDecorrelatedJitterStrategy_NextDelay_Deterministic(t *testing.T) {
	strategy := NewDecorrelatedJitterStrategy(1*time.Second, 2.0, 10*time.Second)

	// Test first delay is always base delay
	delay1 := strategy.NextDelay(1, 0)
	if delay1 != 1*time.Second {
		t.Errorf("First delay should always be base delay, got %v", delay1)
	}

	// Test zero attempt
	delay0 := strategy.NextDelay(0, 0)
	if delay0 != 1*time.Second {
		t.Errorf("Zero attempt should return base delay, got %v", delay0)
	}

	// Test negative attempt
	delayNeg := strategy.NextDelay(-1, 0)
	if delayNeg != 1*time.Second {
		t.Errorf("Negative attempt should return base delay, got %v", delayNeg)
	}
}

func TestDecorrelatedJitterStrategy_NextDelay_MaxDelayRespected(t *testing.T) {
	strategy := NewDecorrelatedJitterStrategy(100*time.Millisecond, 10.0, 2*time.Second)

	// Run many attempts to ensure max delay is always respected
	for attempt := 1; attempt <= 20; attempt++ {
		delay := strategy.NextDelay(attempt, 0)
		if delay > 2*time.Second {
			t.Errorf("Attempt %d: delay %v exceeds max delay %v", attempt, delay, 2*time.Second)
		}
		if delay < 100*time.Millisecond {
			t.Errorf("Attempt %d: delay %v is less than base delay %v", attempt, delay, 100*time.Millisecond)
		}
	}
}

func TestDecorrelatedJitterStrategy_NextDelay_RandomnessDistribution(t *testing.T) {
	strategy := NewDecorrelatedJitterStrategy(1*time.Second, 2.0, 0)

	// Collect delays from multiple runs to verify randomness
	delays := make([]time.Duration, 100)
	for i := 0; i < 100; i++ {
		strategy = NewDecorrelatedJitterStrategy(1*time.Second, 2.0, 0) // reset for each run
		strategy.NextDelay(1, 0)                                        // first delay (always 1s)
		delays[i] = strategy.NextDelay(2, 0)                            // second delay (should vary between 1s and 2s)
	}

	// Verify we have some variation in delays
	minDelay := delays[0]
	maxDelay := delays[0]
	for _, delay := range delays {
		if delay < minDelay {
			minDelay = delay
		}
		if delay > maxDelay {
			maxDelay = delay
		}
	}

	// Should have some spread (at least 200ms difference in a jittered range)
	if maxDelay-minDelay < 200*time.Millisecond {
		t.Errorf("Delays should show more variation, got range %v to %v", minDelay, maxDelay)
	}

	// All delays should be within expected bounds
	for i, delay := range delays {
		if delay < 1*time.Second || delay > 2*time.Second {
			t.Errorf("Delay %d (%v) outside expected range [1s, 2s]", i, delay)
		}
	}
}

func TestDecorrelatedJitterStrategy_ValidateConfig(t *testing.T) {
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
				Multiplier:  3.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: false,
		},
		{
			name: "zero base delay",
			config: &StrategyConfig{
				BaseDelay:   0,
				Multiplier:  3.0,
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
				Multiplier:  3.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: true,
			errorMsg:    "base-delay must be positive",
		},
		{
			name: "multiplier too small",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Multiplier:  1.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: true,
			errorMsg:    "multiplier must be greater than 1.0",
		},
		{
			name: "multiplier exactly 1.0",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Multiplier:  1.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: true,
			errorMsg:    "multiplier must be greater than 1.0",
		},
		{
			name: "multiplier too large",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Multiplier:  11.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: true,
			errorMsg:    "multiplier must be <= 10.0 to prevent excessive delays",
		},
		{
			name: "max delay less than base delay",
			config: &StrategyConfig{
				BaseDelay:   10 * time.Second,
				Multiplier:  3.0,
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
				Multiplier:  3.0,
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
				Multiplier:  3.0,
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
				Multiplier:  2.5,
				MaxDelay:    0, // unlimited
				MaxAttempts: 3,
			},
			expectError: false,
		},
		{
			name: "boundary multiplier (exactly 10.0)",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Multiplier:  10.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectError: false,
		},
		{
			name: "AWS recommended config",
			config: &StrategyConfig{
				BaseDelay:   100 * time.Millisecond,
				Multiplier:  3.0,
				MaxDelay:    20 * time.Second,
				MaxAttempts: 10,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := NewDecorrelatedJitterStrategy(tt.config.BaseDelay, tt.config.Multiplier, tt.config.MaxDelay)
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

func TestDecorrelatedJitterStrategy_Name(t *testing.T) {
	strategy := NewDecorrelatedJitterStrategy(1*time.Second, 3.0, 60*time.Second)
	if strategy.Name() != "decorrelated-jitter" {
		t.Errorf("Name() = %v, expected 'decorrelated-jitter'", strategy.Name())
	}
}

func TestDecorrelatedJitterStrategy_ShouldRetry(t *testing.T) {
	strategy := NewDecorrelatedJitterStrategy(1*time.Second, 3.0, 60*time.Second)

	// ShouldRetry should always return true for decorrelated-jitter strategy
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
				t.Errorf("ShouldRetry() = false, expected true (decorrelated-jitter strategy should always return true)")
			}
		})
	}
}

func TestDecorrelatedJitterStrategy_RealWorldScenarios(t *testing.T) {
	t.Run("AWS API retry scenario", func(t *testing.T) {
		// AWS recommends: base 100ms, multiplier 3.0, max 20s
		strategy := NewDecorrelatedJitterStrategy(100*time.Millisecond, 3.0, 20*time.Second)

		delays := make([]time.Duration, 8)
		for i := 0; i < 8; i++ {
			delays[i] = strategy.NextDelay(i+1, 0)
		}

		// First delay should always be base delay
		if delays[0] != 100*time.Millisecond {
			t.Errorf("First delay should be base delay (100ms), got %v", delays[0])
		}

		// All delays should be within bounds
		for i, delay := range delays {
			if delay < 100*time.Millisecond {
				t.Errorf("Delay[%d] (%v) should not be less than base delay", i, delay)
			}
			if delay > 20*time.Second {
				t.Errorf("Delay[%d] (%v) should not exceed max delay", i, delay)
			}
		}

		// Later delays should tend to be larger (though jitter adds randomness)
		// Test that at least some delays increase
		foundIncrease := false
		for i := 1; i < len(delays); i++ {
			if delays[i] > delays[i-1] {
				foundIncrease = true
				break
			}
		}
		if !foundIncrease {
			t.Errorf("Expected at least some delays to increase over time")
		}
	})

	t.Run("microservice retry scenario", func(t *testing.T) {
		// Microservice scenario: base 500ms, multiplier 2.0, max 30s
		strategy := NewDecorrelatedJitterStrategy(500*time.Millisecond, 2.0, 30*time.Second)

		// Test that strategy behaves reasonably for typical retry counts
		for attempt := 1; attempt <= 6; attempt++ {
			delay := strategy.NextDelay(attempt, 0)

			if attempt == 1 && delay != 500*time.Millisecond {
				t.Errorf("First attempt should be exactly base delay")
			}

			if delay > 30*time.Second {
				t.Errorf("Attempt %d: delay should not exceed max delay", attempt)
			}

			if delay < 500*time.Millisecond {
				t.Errorf("Attempt %d: delay should not be less than base delay", attempt)
			}
		}
	})

	t.Run("thundering herd prevention", func(t *testing.T) {
		// Test that multiple instances don't converge to same delays
		numInstances := 10
		strategy1Delays := make([]time.Duration, 5)
		strategy2Delays := make([]time.Duration, 5)

		for instance := 0; instance < numInstances; instance++ {
			strategy1 := NewDecorrelatedJitterStrategy(1*time.Second, 2.0, 0)
			strategy2 := NewDecorrelatedJitterStrategy(1*time.Second, 2.0, 0)

			for attempt := 1; attempt <= 5; attempt++ {
				delay1 := strategy1.NextDelay(attempt, 0)
				delay2 := strategy2.NextDelay(attempt, 0)

				strategy1Delays[attempt-1] = delay1
				strategy2Delays[attempt-1] = delay2
			}

			// After first attempt, delays should often be different (due to randomness)
			foundDifference := false
			for i := 1; i < 5; i++ { // skip first attempt (always same)
				if strategy1Delays[i] != strategy2Delays[i] {
					foundDifference = true
					break
				}
			}

			// Due to randomness, we expect to find differences most of the time
			// If we never find differences across multiple instances, that's suspicious
			if instance > 5 && !foundDifference {
				t.Logf("Instance %d: strategies produced identical delays (may indicate insufficient randomness)", instance)
			}
		}
	})
}
