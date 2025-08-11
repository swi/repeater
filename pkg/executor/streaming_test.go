package executor

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestExecutor_StreamingOutput(t *testing.T) {
	// Create executor with streaming enabled
	var outputBuffer bytes.Buffer
	executor, err := NewExecutor(WithStreaming(&outputBuffer))
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	// Execute command that produces output
	ctx := context.Background()
	result, err := executor.Execute(ctx, []string{"echo", "Hello, World!"})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify output was streamed to buffer
	streamedOutput := outputBuffer.String()
	if !strings.Contains(streamedOutput, "Hello, World!") {
		t.Errorf("Expected streamed output to contain 'Hello, World!', got: %s", streamedOutput)
	}

	// Verify output is also captured in result
	if !strings.Contains(result.Stdout, "Hello, World!") {
		t.Errorf("Expected result stdout to contain 'Hello, World!', got: %s", result.Stdout)
	}
}

func TestExecutor_StreamingWithPrefix(t *testing.T) {
	var outputBuffer bytes.Buffer
	executor, err := NewExecutor(
		WithStreaming(&outputBuffer),
		WithOutputPrefix("[TEST] "),
	)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	_, err = executor.Execute(ctx, []string{"echo", "test output"})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	streamedOutput := outputBuffer.String()
	if !strings.Contains(streamedOutput, "[TEST] test output") {
		t.Errorf("Expected prefixed output, got: %s", streamedOutput)
	}
}

func TestExecutor_StreamingStderr(t *testing.T) {
	var outputBuffer bytes.Buffer
	executor, err := NewExecutor(WithStreaming(&outputBuffer))
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	// Command that writes to stderr
	result, err := executor.Execute(ctx, []string{"sh", "-c", "echo 'error message' >&2"})

	// Command should succeed but write to stderr
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify stderr was streamed
	streamedOutput := outputBuffer.String()
	if !strings.Contains(streamedOutput, "error message") {
		t.Errorf("Expected streamed stderr to contain 'error message', got: %s", streamedOutput)
	}

	// Verify stderr is also captured in result
	if !strings.Contains(result.Stderr, "error message") {
		t.Errorf("Expected result stderr to contain 'error message', got: %s", result.Stderr)
	}
}

func TestExecutor_StreamingMultipleLines(t *testing.T) {
	var outputBuffer bytes.Buffer
	executor, err := NewExecutor(WithStreaming(&outputBuffer))
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	// Command that produces multiple lines
	_, err = executor.Execute(ctx, []string{"sh", "-c", "echo 'line 1'; echo 'line 2'; echo 'line 3'"})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	streamedOutput := outputBuffer.String()
	expectedLines := []string{"line 1", "line 2", "line 3"}

	for _, line := range expectedLines {
		if !strings.Contains(streamedOutput, line) {
			t.Errorf("Expected streamed output to contain '%s', got: %s", line, streamedOutput)
		}
	}
}

func TestExecutor_StreamingRealTime(t *testing.T) {
	var outputBuffer bytes.Buffer
	executor, err := NewExecutor(WithStreaming(&outputBuffer))
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()

	// Start a command that produces output over time
	go func() {
		_, _ = executor.Execute(ctx, []string{"sh", "-c", "echo 'start'; sleep 0.1; echo 'middle'; sleep 0.1; echo 'end'"})
	}()

	// Check that output appears progressively
	time.Sleep(50 * time.Millisecond)
	output1 := outputBuffer.String()
	if !strings.Contains(output1, "start") {
		t.Errorf("Expected 'start' to appear first, got: %s", output1)
	}

	time.Sleep(150 * time.Millisecond)
	output2 := outputBuffer.String()
	if !strings.Contains(output2, "middle") {
		t.Errorf("Expected 'middle' to appear second, got: %s", output2)
	}

	time.Sleep(150 * time.Millisecond)
	output3 := outputBuffer.String()
	if !strings.Contains(output3, "end") {
		t.Errorf("Expected 'end' to appear last, got: %s", output3)
	}
}

func TestExecutor_NoStreamingByDefault(t *testing.T) {
	var outputBuffer bytes.Buffer
	// Create executor without streaming
	executor, err := NewExecutor()
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	result, err := executor.Execute(ctx, []string{"echo", "Hello, World!"})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify nothing was streamed to buffer (since streaming not enabled)
	streamedOutput := outputBuffer.String()
	if streamedOutput != "" {
		t.Errorf("Expected no streamed output without streaming enabled, got: %s", streamedOutput)
	}

	// But output should still be captured in result
	if !strings.Contains(result.Stdout, "Hello, World!") {
		t.Errorf("Expected result stdout to contain 'Hello, World!', got: %s", result.Stdout)
	}
}

func TestExecutor_StreamingWithQuietMode(t *testing.T) {
	var outputBuffer bytes.Buffer
	executor, err := NewExecutor(
		WithStreaming(&outputBuffer),
		WithQuietMode(),
	)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	result, err := executor.Execute(ctx, []string{"echo", "Hello, World!"})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// In quiet mode, nothing should be streamed
	streamedOutput := outputBuffer.String()
	if streamedOutput != "" {
		t.Errorf("Expected no streamed output in quiet mode, got: %s", streamedOutput)
	}

	// But output should still be captured in result
	if !strings.Contains(result.Stdout, "Hello, World!") {
		t.Errorf("Expected result stdout to contain 'Hello, World!', got: %s", result.Stdout)
	}
}

func TestExecutor_StreamingWithVerboseMode(t *testing.T) {
	var outputBuffer bytes.Buffer
	executor, err := NewExecutor(
		WithStreaming(&outputBuffer),
		WithVerboseMode(),
	)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx := context.Background()
	_, err = executor.Execute(ctx, []string{"echo", "test"})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	streamedOutput := outputBuffer.String()

	// In verbose mode, should include execution metadata
	if !strings.Contains(streamedOutput, "test") {
		t.Errorf("Expected output to contain command output, got: %s", streamedOutput)
	}

	// Should also include verbose information (command, timing, etc.)
	if !strings.Contains(streamedOutput, "echo") {
		t.Errorf("Expected verbose output to include command name, got: %s", streamedOutput)
	}
}
