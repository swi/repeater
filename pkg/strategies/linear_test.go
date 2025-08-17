package strategies

import (
	"errors"
	"testing"
	"time"
)

func TestLinearStrategy_NextDelay(t *testing.T) {
	tests := []struct {
		name        string
		increment   time.Duration
		maxDelay    time.Duration
		attempts    int
		expectedSeq []time.Duration
	}{
		{
			name:      "basic linear progression",
			increment: 2 * time.Second,
			maxDelay:  0, // no cap
			attempts:  5,
			expectedSeq: []time.Duration{
				2 * time.Second,  // attempt 1: 1 * 2s = 2s
				4 * time.Second,  // attempt 2: 2 * 2s = 4s
				6 * time.Second,  // attempt 3: 3 * 2s = 6s
				8 * time.Second,  // attempt 4: 4 * 2s = 8s
				10 * time.Second, // attempt 5: 5 * 2s = 10s
			},
		},
		{
			name:      "with maximum delay cap",
			increment: 3 * time.Second,
			maxDelay:  7 * time.Second,
			attempts:  4,
			expectedSeq: []time.Duration{
				3 * time.Second, // attempt 1: 1 * 3s = 3s
				6 * time.Second, // attempt 2: 2 * 3s = 6s
				7 * time.Second, // attempt 3: 3 * 3s = 9s (capped to 7s)
				7 * time.Second, // attempt 4: 4 * 3s = 12s (capped to 7s)
			},
		},
		{
			name:      "small increment",
			increment: 500 * time.Millisecond,
			maxDelay:  0, // no cap
			attempts:  3,
			expectedSeq: []time.Duration{
				500 * time.Millisecond,  // attempt 1: 1 * 500ms = 500ms
				1000 * time.Millisecond, // attempt 2: 2 * 500ms = 1000ms
				1500 * time.Millisecond, // attempt 3: 3 * 500ms = 1500ms
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := NewLinearStrategy(tt.increment, tt.maxDelay)

			for i, expected := range tt.expectedSeq {
				attempt := i + 1
				actual := strategy.NextDelay(attempt, 0)

				if actual != expected {
					t.Errorf("attempt %d: expected %v, got %v", attempt, expected, actual)
				}
			}
		})
	}
}

func TestLinearStrategy_ValidateConfig(t *testing.T) {
	strategy := &LinearStrategy{}

	tests := []struct {
		name      string
		config    *StrategyConfig
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid config",
			config: &StrategyConfig{
				Increment:   2 * time.Second,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectErr: false,
		},
		{
			name: "zero increment",
			config: &StrategyConfig{
				Increment:   0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectErr: true,
			errMsg:    "increment must be positive",
		},
		{
			name: "negative increment",
			config: &StrategyConfig{
				Increment:   -1 * time.Second,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectErr: true,
			errMsg:    "increment must be positive",
		},
		{
			name: "max delay less than increment",
			config: &StrategyConfig{
				Increment:   5 * time.Second,
				MaxDelay:    2 * time.Second,
				MaxAttempts: 5,
			},
			expectErr: true,
			errMsg:    "max-delay must be greater than increment",
		},
		{
			name: "zero max attempts",
			config: &StrategyConfig{
				Increment:   2 * time.Second,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 0,
			},
			expectErr: true,
			errMsg:    "attempts must be positive",
		},
		{
			name: "no max delay (unlimited)",
			config: &StrategyConfig{
				Increment:   2 * time.Second,
				MaxDelay:    0, // unlimited
				MaxAttempts: 5,
			},
			expectErr: false,
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

func TestLinearStrategy_Name(t *testing.T) {
	strategy := NewLinearStrategy(2*time.Second, 60*time.Second)
	if strategy.Name() != "linear" {
		t.Errorf("expected name 'linear', got '%s'", strategy.Name())
	}
}

func TestLinearStrategy_ShouldRetry(t *testing.T) {
	strategy := NewLinearStrategy(2*time.Second, 60*time.Second)

	// Strategy should always return true - retry logic is handled elsewhere
	if !strategy.ShouldRetry(1, nil, "") {
		t.Error("expected ShouldRetry to return true")
	}

	if !strategy.ShouldRetry(5, errors.New("test error"), "error output") {
		t.Error("expected ShouldRetry to return true even with error")
	}
}
