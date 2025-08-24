package patterns

import (
	"strings"
	"testing"
)

// BenchmarkPatternMatcherMemory tests memory usage of pattern matching operations
func BenchmarkPatternMatcherMemory(b *testing.B) {
	config := PatternConfig{
		SuccessPattern: `success|completed|ok`,
		FailurePattern: `error|failed|exception`,
	}

	matcher, err := NewPatternMatcher(config)
	if err != nil {
		b.Fatalf("Failed to create pattern matcher: %v", err)
	}

	testOutput := "Process completed successfully with status code 0"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := matcher.EvaluateResult(testOutput, 0)
		_ = result.Success
	}
}

// BenchmarkPatternMatcherCaseInsensitiveMemory tests case insensitive pattern matching
func BenchmarkPatternMatcherCaseInsensitiveMemory(b *testing.B) {
	config := PatternConfig{
		SuccessPattern:  `SUCCESS|COMPLETED|OK`,
		FailurePattern:  `ERROR|FAILED|EXCEPTION`,
		CaseInsensitive: true,
	}

	matcher, err := NewPatternMatcher(config)
	if err != nil {
		b.Fatalf("Failed to create pattern matcher: %v", err)
	}

	testOutput := "process completed successfully"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := matcher.EvaluateResult(testOutput, 0)
		_ = result.Success
	}
}

// BenchmarkPatternMatcherLargeOutputMemory tests pattern matching with large output
func BenchmarkPatternMatcherLargeOutputMemory(b *testing.B) {
	config := PatternConfig{
		SuccessPattern: `success|completed`,
		FailurePattern: `error|failed`,
	}

	matcher, err := NewPatternMatcher(config)
	if err != nil {
		b.Fatalf("Failed to create pattern matcher: %v", err)
	}

	// Create a large output string (10KB)
	largeOutput := strings.Repeat("This is a line of log output that we need to process. ", 200) + "Operation completed successfully."

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := matcher.EvaluateResult(largeOutput, 0)
		_ = result.Success
	}
}

// BenchmarkPatternMatcherComplexRegexMemory tests complex regex pattern matching
func BenchmarkPatternMatcherComplexRegexMemory(b *testing.B) {
	config := PatternConfig{
		SuccessPattern: `\b(success|completed|finished)\b.*\d{3}.*ok`,
		FailurePattern: `\b(error|failed|exception)\b.*\d{3}.*failed`,
	}

	matcher, err := NewPatternMatcher(config)
	if err != nil {
		b.Fatalf("Failed to create pattern matcher: %v", err)
	}

	testOutput := "Operation completed with status 200 and everything is ok"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := matcher.EvaluateResult(testOutput, 0)
		_ = result.Success
	}
}

// BenchmarkPatternMatcherConcurrentMemory tests concurrent pattern matching
func BenchmarkPatternMatcherConcurrentMemory(b *testing.B) {
	config := PatternConfig{
		SuccessPattern: `success|completed|ok`,
		FailurePattern: `error|failed|exception`,
	}

	matcher, err := NewPatternMatcher(config)
	if err != nil {
		b.Fatalf("Failed to create pattern matcher: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		outputs := []string{
			"Operation completed successfully",
			"Process failed with error",
			"Task finished ok",
			"System exception occurred",
			"All tests passed successfully",
		}
		i := 0
		for pb.Next() {
			output := outputs[i%len(outputs)]
			exitCode := i % 2 // Alternate exit codes
			result := matcher.EvaluateResult(output, exitCode)
			_ = result.Success
			i++
		}
	})
}

// BenchmarkNewPatternMatcherMemory tests memory usage of creating pattern matchers
func BenchmarkNewPatternMatcherMemory(b *testing.B) {
	config := PatternConfig{
		SuccessPattern: `success|completed|ok`,
		FailurePattern: `error|failed|exception`,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		matcher, err := NewPatternMatcher(config)
		if err != nil {
			b.Fatalf("Failed to create pattern matcher: %v", err)
		}
		_ = matcher
	}
}
