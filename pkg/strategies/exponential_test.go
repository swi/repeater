package strategies

import (
	"errors"
	"testing"
	"time"
)

func TestExponentialStrategy_NextDelay(t *testing.T) {
	tests := []struct {
		name        string
		baseDelay   time.Duration
		multiplier  float64
		maxDelay    time.Duration
		attempts    int
		expectedSeq []time.Duration
	}{
		{
			name:       "standard exponential backoff",
			baseDelay:  1 * time.Second,
			multiplier: 2.0,
			maxDelay:   0, // no cap
			attempts:   5,
			expectedSeq: []time.Duration{
				1 * time.Second,  // attempt 1: 1s * 2^0 = 1s
				2 * time.Second,  // attempt 2: 1s * 2^1 = 2s
				4 * time.Second,  // attempt 3: 1s * 2^2 = 4s
				8 * time.Second,  // attempt 4: 1s * 2^3 = 8s
				16 * time.Second, // attempt 5: 1s * 2^4 = 16s
			},
		},
		{
			name:       "with maximum delay cap",
			baseDelay:  1 * time.Second,
			multiplier: 2.0,
			maxDelay:   5 * time.Second,
			attempts:   5,
			expectedSeq: []time.Duration{
				1 * time.Second, // attempt 1: 1s * 2^0 = 1s
				2 * time.Second, // attempt 2: 1s * 2^1 = 2s
				4 * time.Second, // attempt 3: 1s * 2^2 = 4s
				5 * time.Second, // attempt 4: 1s * 2^3 = 8s (capped to 5s)
				5 * time.Second, // attempt 5: 1s * 2^4 = 16s (capped to 5s)
			},
		},
		{
			name:       "custom multiplier",
			baseDelay:  500 * time.Millisecond,
			multiplier: 1.5,
			maxDelay:   0, // no cap
			attempts:   4,
			expectedSeq: []time.Duration{
				500 * time.Millisecond,  // attempt 1: 500ms * 1.5^0 = 500ms
				750 * time.Millisecond,  // attempt 2: 500ms * 1.5^1 = 750ms
				1125 * time.Millisecond, // attempt 3: 500ms * 1.5^2 = 1125ms
				1687 * time.Millisecond, // attempt 4: 500ms * 1.5^3 = 1687.5ms (truncated)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := NewExponentialStrategy(tt.baseDelay, tt.multiplier, tt.maxDelay)

			for i, expected := range tt.expectedSeq {
				attempt := i + 1
				actual := strategy.NextDelay(attempt, 0)

				// Allow small rounding differences for floating point calculations
				diff := actual - expected
				if diff < 0 {
					diff = -diff
				}
				if diff > time.Millisecond {
					t.Errorf("attempt %d: expected %v, got %v (diff: %v)", attempt, expected, actual, diff)
				}
			}
		})
	}
}

func TestExponentialStrategy_ValidateConfig(t *testing.T) {
	strategy := &ExponentialStrategy{}

	tests := []struct {
		name      string
		config    *StrategyConfig
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid config",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Multiplier:  2.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectErr: false,
		},
		{
			name: "zero base delay",
			config: &StrategyConfig{
				BaseDelay:   0,
				Multiplier:  2.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectErr: true,
			errMsg:    "base-delay must be positive",
		},
		{
			name: "multiplier too small",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Multiplier:  1.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectErr: true,
			errMsg:    "multiplier must be greater than 1.0",
		},
		{
			name: "multiplier too large",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Multiplier:  15.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectErr: true,
			errMsg:    "multiplier must be <= 10.0 to prevent overflow",
		},
		{
			name: "max delay less than base delay",
			config: &StrategyConfig{
				BaseDelay:   5 * time.Second,
				Multiplier:  2.0,
				MaxDelay:    2 * time.Second,
				MaxAttempts: 5,
			},
			expectErr: true,
			errMsg:    "max-delay must be greater than base-delay",
		},
		{
			name: "zero max attempts",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
				Multiplier:  2.0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 0,
			},
			expectErr: true,
			errMsg:    "attempts must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := strategy.ValidateConfig(tt.config)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestExponentialStrategy_Name(t *testing.T) {
	strategy := NewExponentialStrategy(1*time.Second, 2.0, 60*time.Second)
	if strategy.Name() != "exponential" {
		t.Errorf("expected name 'exponential', got '%s'", strategy.Name())
	}
}

func TestExponentialStrategy_ShouldRetry(t *testing.T) {
	strategy := NewExponentialStrategy(1*time.Second, 2.0, 60*time.Second)

	// Strategy should always return true - retry logic is handled elsewhere
	if !strategy.ShouldRetry(1, nil, "") {
		t.Error("expected ShouldRetry to return true")
	}

	if !strategy.ShouldRetry(5, errors.New("test error"), "error output") {
		t.Error("expected ShouldRetry to return true even with error")
	}
}
