package strategies

import (
	"errors"
	"testing"
	"time"
)

func TestFibonacciStrategy_NextDelay(t *testing.T) {
	tests := []struct {
		name        string
		baseDelay   time.Duration
		maxDelay    time.Duration
		attempt     int
		expectedSeq []time.Duration
	}{
		{
			name:      "basic fibonacci sequence",
			baseDelay: 1 * time.Second,
			maxDelay:  0, // no cap
			attempt:   6,
			expectedSeq: []time.Duration{
				1 * time.Second, // attempt 1: 1 * 1s = 1s
				1 * time.Second, // attempt 2: 1 * 1s = 1s
				2 * time.Second, // attempt 3: 2 * 1s = 2s
				3 * time.Second, // attempt 4: 3 * 1s = 3s
				5 * time.Second, // attempt 5: 5 * 1s = 5s
				8 * time.Second, // attempt 6: 8 * 1s = 8s
			},
		},
		{
			name:      "with maximum delay cap",
			baseDelay: 1 * time.Second,
			maxDelay:  4 * time.Second,
			attempt:   6,
			expectedSeq: []time.Duration{
				1 * time.Second, // attempt 1: 1 * 1s = 1s
				1 * time.Second, // attempt 2: 1 * 1s = 1s
				2 * time.Second, // attempt 3: 2 * 1s = 2s
				3 * time.Second, // attempt 4: 3 * 1s = 3s
				4 * time.Second, // attempt 5: 5 * 1s = 5s (capped to 4s)
				4 * time.Second, // attempt 6: 8 * 1s = 8s (capped to 4s)
			},
		},
		{
			name:      "different base delay",
			baseDelay: 500 * time.Millisecond,
			maxDelay:  0, // no cap
			attempt:   4,
			expectedSeq: []time.Duration{
				500 * time.Millisecond,  // attempt 1: 1 * 500ms = 500ms
				500 * time.Millisecond,  // attempt 2: 1 * 500ms = 500ms
				1000 * time.Millisecond, // attempt 3: 2 * 500ms = 1000ms
				1500 * time.Millisecond, // attempt 4: 3 * 500ms = 1500ms
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := NewFibonacciStrategy(tt.baseDelay, tt.maxDelay)

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

func TestFibonacciStrategy_fibonacci(t *testing.T) {
	strategy := &FibonacciStrategy{}

	tests := []struct {
		n        int
		expected int64
	}{
		{0, 1},   // edge case
		{1, 1},   // F(1) = 1
		{2, 1},   // F(2) = 1
		{3, 2},   // F(3) = 2
		{4, 3},   // F(4) = 3
		{5, 5},   // F(5) = 5
		{6, 8},   // F(6) = 8
		{7, 13},  // F(7) = 13
		{8, 21},  // F(8) = 21
		{10, 55}, // F(10) = 55
	}

	for _, tt := range tests {
		actual := strategy.fibonacci(tt.n)
		if actual != tt.expected {
			t.Errorf("fibonacci(%d): expected %d, got %d", tt.n, tt.expected, actual)
		}
	}
}

func TestFibonacciStrategy_ValidateConfig(t *testing.T) {
	strategy := &FibonacciStrategy{}

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
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectErr: false,
		},
		{
			name: "zero base delay",
			config: &StrategyConfig{
				BaseDelay:   0,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectErr: true,
			errMsg:    "base-delay must be positive",
		},
		{
			name: "negative base delay",
			config: &StrategyConfig{
				BaseDelay:   -1 * time.Second,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 5,
			},
			expectErr: true,
			errMsg:    "base-delay must be positive",
		},
		{
			name: "max delay less than base delay",
			config: &StrategyConfig{
				BaseDelay:   5 * time.Second,
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
				MaxDelay:    60 * time.Second,
				MaxAttempts: 0,
			},
			expectErr: true,
			errMsg:    "attempts must be positive",
		},
		{
			name: "no max delay (unlimited)",
			config: &StrategyConfig{
				BaseDelay:   1 * time.Second,
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

func TestFibonacciStrategy_Name(t *testing.T) {
	strategy := NewFibonacciStrategy(1*time.Second, 60*time.Second)
	if strategy.Name() != "fibonacci" {
		t.Errorf("expected name 'fibonacci', got '%s'", strategy.Name())
	}
}

func TestFibonacciStrategy_ShouldRetry(t *testing.T) {
	strategy := NewFibonacciStrategy(1*time.Second, 60*time.Second)

	// Strategy should always return true - retry logic is handled elsewhere
	if !strategy.ShouldRetry(1, nil, "") {
		t.Error("expected ShouldRetry to return true")
	}

	if !strategy.ShouldRetry(5, errors.New("test error"), "error output") {
		t.Error("expected ShouldRetry to return true even with error")
	}
}
