package cli

import (
	"testing"
)

func BenchmarkParseArgs_Simple(b *testing.B) {
	args := []string{"interval", "--every", "1s", "--", "echo", "test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseArgs(args)
	}
}

func BenchmarkParseArgs_Complex(b *testing.B) {
	args := []string{"--config", "/path/to/config.toml", "duration", "--for", "2m", "--every", "10s", "--", "curl", "-v", "http://example.com"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseArgs(args)
	}
}

func BenchmarkParseArgs_Help(b *testing.B) {
	args := []string{"--help"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseArgs(args)
	}
}

func BenchmarkParseArgs_Count(b *testing.B) {
	args := []string{"count", "--times", "100", "--", "date"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseArgs(args)
	}
}
