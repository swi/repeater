// Package interfaces defines error handling contracts for Repeater components.
package interfaces

import "context"

// ErrorHandler defines the interface for centralized error handling
type ErrorHandler interface {
	// HandleError processes an error with context and severity
	HandleError(ctx context.Context, err error, severity ErrorSeverity) error

	// ShouldRetry determines if an operation should be retried based on the error
	ShouldRetry(err error, attempt int) bool

	// GetErrorCategory classifies the error type
	GetErrorCategory(err error) ErrorCategory
}

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	ErrorSeverityInfo ErrorSeverity = iota
	ErrorSeverityLow
	ErrorSeverityMedium
	ErrorSeverityHigh
	ErrorSeverityCritical
)

// ErrorCategory represents the category of an error
type ErrorCategory int

const (
	ErrorCategoryUnknown ErrorCategory = iota
	ErrorCategoryTimeout
	ErrorCategoryNetwork
	ErrorCategoryPermission
	ErrorCategoryResource
	ErrorCategoryCommand
	ErrorCategorySystem
)

// RepeaterError defines a structured error with additional context
type RepeaterError interface {
	error

	// GetCategory returns the error category
	GetCategory() ErrorCategory

	// GetSeverity returns the error severity
	GetSeverity() ErrorSeverity

	// GetContext returns additional error context
	GetContext() map[string]any

	// IsRetryable indicates if the error is retryable
	IsRetryable() bool
}
