package recovery

import (
	"context"
	"time"
)

// FallbackFunc defines the signature for fallback functions
type FallbackFunc func(ctx context.Context, originalErr error) error

// ExecuteFunc defines the signature for functions to be executed with recovery
type ExecuteFunc func(ctx context.Context) error

// RecoveryState tracks the state of recovery operations
type RecoveryState struct {
	TotalAttempts        int           `json:"total_attempts"`
	SuccessfulRecoveries int           `json:"successful_recoveries"`
	FailedRecoveries     int           `json:"failed_recoveries"`
	ConsecutiveSuccesses int           `json:"consecutive_successes"`
	ConsecutiveFailures  int           `json:"consecutive_failures"`
	RecentFailures       []error       `json:"recent_failures"`
	LastSuccessTime      time.Time     `json:"last_success_time"`
	LastFailureTime      time.Time     `json:"last_failure_time"`
	AverageRecoveryTime  time.Duration `json:"average_recovery_time"`
}
