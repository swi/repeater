package httpaware

import (
	"strings"
	"testing"
	"time"

	httpbinTest "github.com/swi/repeater/pkg/testing"
)

// BenchmarkHTTPResponseParser_RealWorldResponses benchmarks parsing with realistic HTTP responses
func BenchmarkHTTPResponseParser_RealWorldResponses(b *testing.B) {
	parser := NewHTTPResponseParser()

	// Realistic HTTPBin-style responses
	responses := []string{
		// 503 with Retry-After header
		"HTTP/1.1 503 Service Unavailable\r\nRetry-After: 30\r\nContent-Type: text/html\r\n\r\n<html><body>Service Unavailable</body></html>",

		// 429 with JSON retry info
		"HTTP/1.1 429 Too Many Requests\r\nContent-Type: application/json\r\n\r\n{\"error\":{\"retry_after\":60,\"message\":\"Rate limit exceeded\"}}",

		// 200 Success response (should be ignored)
		"HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{\"slideshow\":{\"title\":\"Sample\"}}",

		// Large JSON response similar to HTTPBin
		"HTTP/1.1 503 Service Unavailable\r\nContent-Type: application/json\r\n\r\n" + strings.Repeat(`{"retry_after":45,"details":"Service temporarily unavailable"}`, 10),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		response := responses[i%len(responses)]
		timingInfo, _ := parser.ParseResponse(response)
		_ = timingInfo
	}
}

// BenchmarkHTTPResponseParser_LargeResponse benchmarks parsing with large HTTP responses
func BenchmarkHTTPResponseParser_LargeResponse(b *testing.B) {
	parser := NewHTTPResponseParser()

	// Simulate a large response like HTTPBin might return
	largeBody := strings.Repeat(`{"message":"This is a large response body","data":[1,2,3,4,5],"metadata":{"timestamp":"2025-01-08T10:00:00Z"}}`, 100)
	largeResponse := "HTTP/1.1 503 Service Unavailable\r\nContent-Type: application/json\r\nRetry-After: 120\r\n\r\n" + largeBody

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		timingInfo, _ := parser.ParseResponse(largeResponse)
		_ = timingInfo
	}
}

// BenchmarkHTTPAwareScheduler_RealWorldTiming benchmarks HTTP-aware scheduling with realistic scenarios
func BenchmarkHTTPAwareScheduler_RealWorldTiming(b *testing.B) {
	config := &HTTPAwareConfig{
		MaxDelay:     30 * time.Minute,
		MinDelay:     1 * time.Second,
		ParseJSON:    true,
		ParseHeaders: true,
	}

	scheduler := NewHTTPAwareScheduler(*config)

	// Create a fallback scheduler
	fallbackScheduler := newMockScheduler(10 * time.Second)
	scheduler.SetFallbackScheduler(fallbackScheduler)

	// Realistic HTTP responses from HTTPBin-style services
	httpResponses := []string{
		// Rate limiting with retry-after
		"HTTP/1.1 429 Too Many Requests\r\nRetry-After: 60\r\n\r\n",

		// Service unavailable with JSON timing
		"HTTP/1.1 503 Service Unavailable\r\nContent-Type: application/json\r\n\r\n{\"retry_after\": 45}",

		// Success response (no timing)
		"HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{\"status\": \"success\"}",

		// Complex JSON with nested timing info
		"HTTP/1.1 503 Service Unavailable\r\nContent-Type: application/json\r\n\r\n{\"error\":{\"retry_after\":30,\"backoff\":{\"delay\":25}}}",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		response := httpResponses[i%len(httpResponses)]
		scheduler.SetLastResponse(response)
		delay := scheduler.NextDelay()
		_ = delay
	}
}

// BenchmarkHTTPBinScenarios_Integration benchmarks complete HTTPBin integration scenarios
func BenchmarkHTTPBinScenarios_Integration(b *testing.B) {
	helper := httpbinTest.NewHTTPBinHelper(nil)

	// Skip if no network available
	if !helper.IsHTTPBinAvailable() {
		b.Skip("HTTPBin not available - skipping integration benchmarks")
	}

	scenarios := helper.CommonHTTPBinScenarios()

	// Test different scheduling configurations
	configs := []*HTTPAwareConfig{
		{MaxDelay: 30 * time.Second, MinDelay: 1 * time.Second, ParseJSON: true},
		{MaxDelay: 60 * time.Second, MinDelay: 5 * time.Second, ParseHeaders: true},
		{MaxDelay: 120 * time.Second, MinDelay: 2 * time.Second, ParseJSON: true, ParseHeaders: true},
	}

	for configIdx, config := range configs {
		b.Run(b.Name()+"/config_"+string(rune('A'+configIdx)), func(b *testing.B) {
			scheduler := NewHTTPAwareScheduler(*config)
			fallbackScheduler := newMockScheduler(15 * time.Second)
			scheduler.SetFallbackScheduler(fallbackScheduler)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				scenario := scenarios[i%len(scenarios)]

				// Simulate response processing (in real tests, this would come from curl execution)
				var simulatedResponse string
				if strings.Contains(scenario.Endpoint, "status/503") {
					simulatedResponse = "HTTP/1.1 503 Service Unavailable\r\nRetry-After: 30\r\n\r\n"
				} else if strings.Contains(scenario.Endpoint, "status/429") {
					simulatedResponse = "HTTP/1.1 429 Too Many Requests\r\nRetry-After: 60\r\n\r\n"
				} else if strings.Contains(scenario.Endpoint, "json") {
					simulatedResponse = "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{\"slideshow\":{\"title\":\"Sample\"}}"
				} else {
					simulatedResponse = "HTTP/1.1 200 OK\r\n\r\n"
				}

				scheduler.SetLastResponse(simulatedResponse)
				delay := scheduler.NextDelay()
				_ = delay
			}
		})
	}
}

// Mock scheduler for benchmarking
type mockScheduler struct {
	delay  time.Duration
	nextCh chan time.Time
}

func newMockScheduler(delay time.Duration) *mockScheduler {
	return &mockScheduler{
		delay:  delay,
		nextCh: make(chan time.Time, 1),
	}
}

func (m *mockScheduler) Next() <-chan time.Time {
	go func() {
		<-time.After(m.delay)
		select {
		case m.nextCh <- time.Now():
		default:
		}
	}()
	return m.nextCh
}

func (m *mockScheduler) Stop() {
	close(m.nextCh)
}

// BenchmarkHTTPBinEndpoints_URLGeneration benchmarks endpoint URL generation
func BenchmarkHTTPBinEndpoints_URLGeneration(b *testing.B) {
	endpoints := httpbinTest.NewHTTPBinEndpoints("https://httpbin.org")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Cycle through different endpoint types
		switch i % 7 {
		case 0:
			_ = endpoints.Status(503)
		case 1:
			_ = endpoints.Status(429)
		case 2:
			_ = endpoints.JSON()
		case 3:
			_ = endpoints.Delay(5)
		case 4:
			_ = endpoints.Headers()
		case 5:
			_ = endpoints.Get()
		case 6:
			_ = endpoints.UserAgent()
		}
	}
}

// BenchmarkHTTPBinHelper_CurlCommandGeneration benchmarks curl command generation
func BenchmarkHTTPBinHelper_CurlCommandGeneration(b *testing.B) {
	helper := httpbinTest.NewHTTPBinHelper(nil)
	endpoints := []string{
		"https://httpbin.org/status/503",
		"https://httpbin.org/status/429",
		"https://httpbin.org/json",
		"https://httpbin.org/delay/2",
		"https://httpbin.org/headers",
	}

	methods := []string{"GET", "POST", "PUT", "DELETE"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		endpoint := endpoints[i%len(endpoints)]
		method := methods[i%len(methods)]
		curlCmd := helper.GetCurlCommand(endpoint, method)
		_ = curlCmd
	}
}
