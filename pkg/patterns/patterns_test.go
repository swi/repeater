package patterns

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatternMatcher_BasicSuccessPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		successPattern string
		output         string
		exitCode       int
		expectedResult bool
		expectedCode   int
	}{
		{
			name:           "success pattern matches should return success",
			successPattern: "deployment successful",
			output:         "INFO: deployment successful\nCompleted at 2025-01-08",
			exitCode:       1, // Original command failed
			expectedResult: true,
			expectedCode:   0, // Should be overridden to success
		},
		{
			name:           "success pattern no match should use exit code",
			successPattern: "deployment successful",
			output:         "ERROR: deployment failed\nExiting with error",
			exitCode:       1,
			expectedResult: false,
			expectedCode:   1, // Should preserve original exit code
		},
		{
			name:           "success pattern matches with zero exit code",
			successPattern: "build completed",
			output:         "build completed successfully",
			exitCode:       0,
			expectedResult: true,
			expectedCode:   0, // Should remain success
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewPatternMatcher(PatternConfig{
				SuccessPattern: tt.successPattern,
			})
			require.NoError(t, err)

			result := matcher.EvaluateResult(tt.output, tt.exitCode)

			assert.Equal(t, tt.expectedResult, result.Success)
			assert.Equal(t, tt.expectedCode, result.ExitCode)
		})
	}
}

func TestPatternMatcher_BasicFailurePattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		failurePattern string
		output         string
		exitCode       int
		expectedResult bool
		expectedCode   int
	}{
		{
			name:           "failure pattern matches should return failure",
			failurePattern: "(?i)error|failed",
			output:         "Process completed\nERROR: connection timeout",
			exitCode:       0, // Original command succeeded
			expectedResult: false,
			expectedCode:   1, // Should be overridden to failure
		},
		{
			name:           "failure pattern no match should use exit code",
			failurePattern: "(?i)error|failed",
			output:         "Process completed successfully",
			exitCode:       0,
			expectedResult: true,
			expectedCode:   0, // Should preserve original exit code
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewPatternMatcher(PatternConfig{
				FailurePattern: tt.failurePattern,
			})
			require.NoError(t, err)

			result := matcher.EvaluateResult(tt.output, tt.exitCode)

			assert.Equal(t, tt.expectedResult, result.Success)
			assert.Equal(t, tt.expectedCode, result.ExitCode)
		})
	}
}

func TestPatternMatcher_PatternPrecedence(t *testing.T) {
	// Failure patterns should override success patterns
	matcher, err := NewPatternMatcher(PatternConfig{
		SuccessPattern: "deployment",
		FailurePattern: "(?i)error", // Case-insensitive to match "ERROR"
	})
	require.NoError(t, err)

	// Output contains both success and failure patterns
	output := "deployment started\nERROR: network timeout\ndeployment failed"
	result := matcher.EvaluateResult(output, 0)

	// Failure pattern should take precedence
	assert.False(t, result.Success)
	assert.Equal(t, 1, result.ExitCode)
}

func TestPatternMatcher_NoPatterns(t *testing.T) {
	// When no patterns are configured, should use exit code
	matcher, err := NewPatternMatcher(PatternConfig{})
	require.NoError(t, err)

	tests := []struct {
		name          string
		output        string
		exitCode      int
		expectSuccess bool
	}{
		{
			name:          "exit code 0 should succeed",
			output:        "some output",
			exitCode:      0,
			expectSuccess: true,
		},
		{
			name:          "exit code 1 should fail",
			output:        "some output",
			exitCode:      1,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.EvaluateResult(tt.output, tt.exitCode)
			assert.Equal(t, tt.expectSuccess, result.Success)
			assert.Equal(t, tt.exitCode, result.ExitCode)
		})
	}
}

func TestPatternMatcher_CaseInsensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		pattern         string
		output          string
		caseInsensitive bool
		shouldMatch     bool
	}{
		{
			name:            "case sensitive should not match different case",
			pattern:         "SUCCESS",
			output:          "Process completed with success",
			caseInsensitive: false,
			shouldMatch:     false,
		},
		{
			name:            "case insensitive should match different case",
			pattern:         "SUCCESS",
			output:          "Process completed with success",
			caseInsensitive: true,
			shouldMatch:     true,
		},
		{
			name:            "case insensitive should match mixed case",
			pattern:         "deployment successful",
			output:          "DEPLOYMENT SUCCESSFUL at 2025-01-08",
			caseInsensitive: true,
			shouldMatch:     true,
		},
		{
			name:            "case insensitive should match exact case",
			pattern:         "SUCCESS",
			output:          "Process completed with SUCCESS",
			caseInsensitive: true,
			shouldMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewPatternMatcher(PatternConfig{
				SuccessPattern:  tt.pattern,
				CaseInsensitive: tt.caseInsensitive,
			})
			require.NoError(t, err)

			result := matcher.EvaluateResult(tt.output, 1)

			if tt.shouldMatch {
				assert.True(t, result.Success, "Pattern should have matched")
				assert.Equal(t, 0, result.ExitCode)
				assert.Equal(t, "success pattern matched", result.Reason)
			} else {
				assert.False(t, result.Success, "Pattern should not have matched")
				assert.Equal(t, 1, result.ExitCode)
				assert.Equal(t, "exit code used", result.Reason)
			}
		})
	}
}

func TestPatternMatcher_CaseInsensitiveFailurePattern(t *testing.T) {
	tests := []struct {
		name            string
		pattern         string
		output          string
		caseInsensitive bool
		shouldMatch     bool
	}{
		{
			name:            "case sensitive failure pattern should not match different case",
			pattern:         "ERROR",
			output:          "Process failed with error",
			caseInsensitive: false,
			shouldMatch:     false,
		},
		{
			name:            "case insensitive failure pattern should match different case",
			pattern:         "ERROR",
			output:          "Process failed with error",
			caseInsensitive: true,
			shouldMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewPatternMatcher(PatternConfig{
				FailurePattern:  tt.pattern,
				CaseInsensitive: tt.caseInsensitive,
			})
			require.NoError(t, err)

			result := matcher.EvaluateResult(tt.output, 0)

			if tt.shouldMatch {
				assert.False(t, result.Success, "Failure pattern should have matched")
				assert.Equal(t, 1, result.ExitCode)
				assert.Equal(t, "failure pattern matched", result.Reason)
			} else {
				assert.True(t, result.Success, "Failure pattern should not have matched")
				assert.Equal(t, 0, result.ExitCode)
				assert.Equal(t, "exit code used", result.Reason)
			}
		})
	}
}

func TestPatternMatcher_InvalidRegex(t *testing.T) {
	tests := []struct {
		name    string
		config  PatternConfig
		wantErr bool
	}{
		{
			name: "invalid success pattern",
			config: PatternConfig{
				SuccessPattern: "[invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid failure pattern",
			config: PatternConfig{
				FailurePattern: "*invalid",
			},
			wantErr: true,
		},
		{
			name: "valid patterns",
			config: PatternConfig{
				SuccessPattern: "success",
				FailurePattern: "(?i)error",
			},
			wantErr: false,
		},
		{
			name: "valid patterns with case insensitive",
			config: PatternConfig{
				SuccessPattern:  "SUCCESS",
				FailurePattern:  "ERROR",
				CaseInsensitive: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPatternMatcher(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
