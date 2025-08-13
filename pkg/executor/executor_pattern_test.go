package executor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/patterns"
)

func TestExecutor_WithPatternMatching(t *testing.T) {
	tests := []struct {
		name           string
		command        []string
		patternConfig  patterns.PatternConfig
		expectedResult bool
		expectedCode   int
		description    string
	}{
		{
			name:    "success pattern overrides non-zero exit code",
			command: []string{"sh", "-c", "echo 'deployment successful'; exit 1"},
			patternConfig: patterns.PatternConfig{
				SuccessPattern: "deployment successful",
			},
			expectedResult: true,
			expectedCode:   0,
			description:    "Commands that print success but exit with 1 should be treated as successful",
		},
		{
			name:    "failure pattern overrides zero exit code",
			command: []string{"sh", "-c", "echo 'ERROR: connection failed'; exit 0"},
			patternConfig: patterns.PatternConfig{
				FailurePattern: "(?i)error",
			},
			expectedResult: false,
			expectedCode:   1,
			description:    "Commands that print errors but exit with 0 should be treated as failed",
		},
		{
			name:    "no pattern match uses original exit code success",
			command: []string{"sh", "-c", "echo 'process completed'; exit 0"},
			patternConfig: patterns.PatternConfig{
				SuccessPattern: "deployment successful",
			},
			expectedResult: true,
			expectedCode:   0,
			description:    "When no pattern matches, should use original exit code",
		},
		{
			name:    "no pattern match uses original exit code failure",
			command: []string{"sh", "-c", "echo 'process completed'; exit 1"},
			patternConfig: patterns.PatternConfig{
				SuccessPattern: "deployment successful",
			},
			expectedResult: false,
			expectedCode:   1,
			description:    "When no pattern matches, should use original exit code",
		},
		{
			name:    "case insensitive pattern matching",
			command: []string{"sh", "-c", "echo 'DEPLOYMENT SUCCESSFUL'; exit 1"},
			patternConfig: patterns.PatternConfig{
				SuccessPattern:  "deployment successful",
				CaseInsensitive: true,
			},
			expectedResult: true,
			expectedCode:   0,
			description:    "Case insensitive patterns should work correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewExecutorWithConfig(ExecutorConfig{
				Timeout:       30 * time.Second,
				PatternConfig: &tt.patternConfig,
			})
			require.NoError(t, err)

			ctx := context.Background()
			result, err := executor.Execute(ctx, tt.command)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedResult, result.Success, tt.description)
			assert.Equal(t, tt.expectedCode, result.ExitCode, tt.description)
			assert.NotEmpty(t, result.Reason, "Reason should be provided")
		})
	}
}

func TestExecutor_PatternMatchingWithStreaming(t *testing.T) {
	patternConfig := patterns.PatternConfig{
		SuccessPattern: "build completed",
	}

	executor, err := NewExecutorWithConfig(ExecutorConfig{
		Timeout:       30 * time.Second,
		PatternConfig: &patternConfig,
		Streaming:     true,
	})
	require.NoError(t, err)

	ctx := context.Background()
	command := []string{"sh", "-c", "echo 'starting build'; echo 'build completed'; exit 1"}

	result, err := executor.Execute(ctx, command)
	require.NoError(t, err)

	// Should succeed due to pattern match despite exit code 1
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Output, "build completed")
	assert.Equal(t, "success pattern matched", result.Reason)
}

func TestExecutor_PatternPrecedence(t *testing.T) {
	// Test that failure patterns take precedence over success patterns
	patternConfig := patterns.PatternConfig{
		SuccessPattern: "completed",
		FailurePattern: "error",
	}

	executor, err := NewExecutorWithConfig(ExecutorConfig{
		Timeout:       30 * time.Second,
		PatternConfig: &patternConfig,
	})
	require.NoError(t, err)

	ctx := context.Background()
	command := []string{"sh", "-c", "echo 'Process completed with error'; exit 0"}

	result, err := executor.Execute(ctx, command)
	require.NoError(t, err)

	// Should fail due to failure pattern taking precedence
	assert.False(t, result.Success)
	assert.Equal(t, 1, result.ExitCode)
	assert.Equal(t, "failure pattern matched", result.Reason)
}

func TestExecutor_NoPatternConfig(t *testing.T) {
	// Test that executor works normally when no pattern config is provided
	executor, err := NewExecutorWithConfig(ExecutorConfig{
		Timeout: 30 * time.Second,
		// No PatternConfig provided
	})
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name          string
		command       []string
		expectSuccess bool
		expectCode    int
	}{
		{
			name:          "successful command",
			command:       []string{"echo", "test"},
			expectSuccess: true,
			expectCode:    0,
		},
		{
			name:          "failing command",
			command:       []string{"sh", "-c", "exit 1"},
			expectSuccess: false,
			expectCode:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(ctx, tt.command)
			require.NoError(t, err)

			assert.Equal(t, tt.expectSuccess, result.Success)
			assert.Equal(t, tt.expectCode, result.ExitCode)
			assert.Equal(t, "exit code used", result.Reason)
		})
	}
}
