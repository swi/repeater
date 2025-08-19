package executor

import (
	"context"
	"testing"
	"time"

	"github.com/swi/repeater/pkg/patterns"
)

func BenchmarkExecutor_SimpleCommand(b *testing.B) {
	executor, _ := NewExecutor(WithTimeout(5 * time.Second))
	ctx := context.Background()
	command := []string{"echo", "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.Execute(ctx, command)
	}
}

func BenchmarkExecutor_FastCommand(b *testing.B) {
	executor, _ := NewExecutor(WithTimeout(5 * time.Second))
	ctx := context.Background()
	command := []string{"true"} // Fastest possible command

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.Execute(ctx, command)
	}
}

func BenchmarkExecutor_WithTimeout(b *testing.B) {
	executor, _ := NewExecutor(WithTimeout(100 * time.Millisecond))
	ctx := context.Background()
	command := []string{"echo", "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.Execute(ctx, command)
	}
}

func BenchmarkExecutor_PatternMatching(b *testing.B) {
	config := ExecutorConfig{
		Timeout: 5 * time.Second,
		PatternConfig: &patterns.PatternConfig{
			SuccessPattern: "test",
		},
	}
	executor, _ := NewExecutorWithConfig(config)
	ctx := context.Background()
	command := []string{"echo", "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.Execute(ctx, command)
	}
}

func BenchmarkExecutor_NewInstance(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewExecutor(WithTimeout(5 * time.Second))
	}
}

func BenchmarkExecutor_LargeOutput(b *testing.B) {
	executor, _ := NewExecutor(WithTimeout(5 * time.Second))
	ctx := context.Background()
	command := []string{"yes"} // Generates large output until timeout

	// Use shorter timeout for benchmark
	shortCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.Execute(shortCtx, command)
	}
}

func BenchmarkOptimizedExecutor_SimpleCommand(b *testing.B) {
	executor, _ := NewOptimizedExecutor(WithTimeout(5 * time.Second))
	ctx := context.Background()
	command := []string{"echo", "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.ExecuteOptimized(ctx, command)
	}
}

func BenchmarkOptimizedExecutor_FastCommand(b *testing.B) {
	executor, _ := NewOptimizedExecutor(WithTimeout(5 * time.Second))
	ctx := context.Background()
	command := []string{"true"} // Fastest possible command

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.ExecuteOptimized(ctx, command)
	}
}
