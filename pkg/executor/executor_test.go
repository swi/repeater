package executor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutor_Creation(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		wantErr bool
	}{
		{
			name:    "default executor",
			options: nil,
			wantErr: false,
		},
		{
			name:    "executor with timeout",
			options: []Option{WithTimeout(5 * time.Second)},
			wantErr: false,
		},
		{
			name:    "executor with zero timeout should error",
			options: []Option{WithTimeout(0)},
			wantErr: true,
		},
		{
			name:    "executor with negative timeout should error",
			options: []Option{WithTimeout(-1 * time.Second)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewExecutor(tt.options...)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, executor)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, executor)
			}
		})
	}
}

func TestExecutor_BasicExecution(t *testing.T) {
	tests := []struct {
		name         string
		command      []string
		expectedCode int
		expectedOut  string
		wantErr      bool
	}{
		{
			name:         "simple echo command",
			command:      []string{"echo", "hello"},
			expectedCode: 0,
			expectedOut:  "hello\n",
			wantErr:      false,
		},
		{
			name:         "echo with multiple args",
			command:      []string{"echo", "hello", "world"},
			expectedCode: 0,
			expectedOut:  "hello world\n",
			wantErr:      false,
		},
		{
			name:         "command that returns non-zero exit code",
			command:      []string{"sh", "-c", "exit 42"},
			expectedCode: 42,
			expectedOut:  "",
			wantErr:      false,
		},
		{
			name:         "empty command should error",
			command:      []string{},
			expectedCode: 0,
			expectedOut:  "",
			wantErr:      true,
		},
		{
			name:         "nonexistent command should error",
			command:      []string{"nonexistent-command-12345"},
			expectedCode: 0,
			expectedOut:  "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewExecutor()
			require.NoError(t, err)

			result, err := executor.Execute(context.Background(), tt.command)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCode, result.ExitCode)
			assert.Equal(t, tt.expectedOut, result.Stdout)
		})
	}
}

func TestExecutor_TimeoutHandling(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		command       []string
		shouldTimeout bool
	}{
		{
			name:          "fast command within timeout",
			timeout:       1 * time.Second,
			command:       []string{"echo", "fast"},
			shouldTimeout: false,
		},
		{
			name:          "slow command exceeds timeout",
			timeout:       100 * time.Millisecond,
			command:       []string{"sleep", "1"},
			shouldTimeout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewExecutor(WithTimeout(tt.timeout))
			require.NoError(t, err)

			start := time.Now()
			result, err := executor.Execute(context.Background(), tt.command)
			elapsed := time.Since(start)

			if tt.shouldTimeout {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "timeout")
				assert.True(t, elapsed >= tt.timeout)
				assert.True(t, elapsed < tt.timeout+50*time.Millisecond) // Should timeout quickly
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 0, result.ExitCode)
				assert.True(t, elapsed < tt.timeout)
			}
		})
	}
}

func TestExecutor_OutputCapture(t *testing.T) {
	tests := []struct {
		name           string
		command        []string
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "stdout only",
			command:        []string{"echo", "stdout message"},
			expectedStdout: "stdout message\n",
			expectedStderr: "",
		},
		{
			name:           "stderr only",
			command:        []string{"sh", "-c", "echo 'stderr message' >&2"},
			expectedStdout: "",
			expectedStderr: "stderr message\n",
		},
		{
			name:           "both stdout and stderr",
			command:        []string{"sh", "-c", "echo 'stdout'; echo 'stderr' >&2"},
			expectedStdout: "stdout\n",
			expectedStderr: "stderr\n",
		},
		{
			name:           "no output",
			command:        []string{"true"},
			expectedStdout: "",
			expectedStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewExecutor()
			require.NoError(t, err)

			result, err := executor.Execute(context.Background(), tt.command)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStdout, result.Stdout)
			assert.Equal(t, tt.expectedStderr, result.Stderr)
		})
	}
}

func TestExecutor_ContextCancellation(t *testing.T) {
	executor, err := NewExecutor()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// Start a long-running command
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	result, err := executor.Execute(ctx, []string{"sleep", "10"})
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Nil(t, result)
	assert.True(t, elapsed < 1*time.Second) // Should cancel quickly
}

func TestExecutor_LargeOutput(t *testing.T) {
	executor, err := NewExecutor()
	require.NoError(t, err)

	// Generate large output (1000 lines)
	command := []string{"sh", "-c", "for i in $(seq 1 1000); do echo \"line $i\"; done"}

	result, err := executor.Execute(context.Background(), command)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	assert.Equal(t, 1000, len(lines))
	assert.Equal(t, "line 1", lines[0])
	assert.Equal(t, "line 1000", lines[999])
}

func TestExecutor_ConcurrentExecution(t *testing.T) {
	executor, err := NewExecutor()
	require.NoError(t, err)

	const numGoroutines = 10
	results := make(chan *ExecutionResult, numGoroutines)
	errors := make(chan error, numGoroutines)

	// Start multiple concurrent executions
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			command := []string{"echo", "goroutine", string(rune('0' + id))}
			result, err := executor.Execute(context.Background(), command)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		select {
		case result := <-results:
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, result.Stdout, "goroutine")
		case err := <-errors:
			t.Errorf("Unexpected error: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}
}
