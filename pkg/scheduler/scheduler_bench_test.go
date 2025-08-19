package scheduler

import (
	"testing"
	"time"

	"github.com/swi/repeater/pkg/strategies"
)

func BenchmarkIntervalScheduler_Next(b *testing.B) {
	scheduler, _ := NewIntervalScheduler(100*time.Millisecond, 0.0, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := scheduler.Next()
		// Consume channel to avoid blocking
		select {
		case <-ch:
		case <-time.After(1 * time.Millisecond):
		}
		scheduler.Stop()
		scheduler, _ = NewIntervalScheduler(100*time.Millisecond, 0.0, false) // Reset for next iteration
	}
}

func BenchmarkStrategyScheduler_Exponential(b *testing.B) {
	strategy := &strategies.ExponentialStrategy{}
	config := &strategies.StrategyConfig{
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
		MaxAttempts: 5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheduler, _ := NewStrategyScheduler(strategy, config)
		ch := scheduler.Next()
		select {
		case <-ch:
		case <-time.After(1 * time.Millisecond):
		}
		scheduler.Stop()
	}
}

func BenchmarkLoadAwareScheduler_UpdateFromMetrics(b *testing.B) {
	scheduler := NewLoadAwareScheduler(
		1*time.Second, // baseInterval
		0.7,           // targetCPU
		0.8,           // targetMemory
		0.5,           // alpha
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scheduler.UpdateFromMetrics()
	}
}

func BenchmarkSchedulerCreation_Interval(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheduler, _ := NewIntervalScheduler(100*time.Millisecond, 0.0, false)
		scheduler.Stop()
	}
}

func BenchmarkSchedulerCreation_LoadAware(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheduler := NewLoadAwareScheduler(1*time.Second, 0.7, 0.8, 0.5)
		scheduler.Stop()
	}
}
