package patterns

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	httpbinTest "github.com/swi/repeater/pkg/testing"
)

// TestPatternMatcher_HTTPBin_RealWorldJSON tests pattern matching with real JSON responses
func TestPatternMatcher_HTTPBin_RealWorldJSON(t *testing.T) {
	t.Parallel()

	helper := httpbinTest.NewHTTPBinHelper(nil)
	helper.SkipIfNoNetwork(t)

	tests := []struct {
		name            string
		endpoint        string
		successPattern  string
		failurePattern  string
		expectedSuccess bool
		description     string
	}{
		{
			name:            "json_slideshow_success_pattern",
			endpoint:        helper.CommonHTTPBinScenarios()[4].Endpoint, // JSON endpoint
			successPattern:  "slideshow",
			expectedSuccess: true,
			description:     "Should match slideshow in HTTPBin JSON response",
		},
		{
			name:            "json_origin_success_pattern",
			endpoint:        helper.CommonHTTPBinScenarios()[4].Endpoint,
			successPattern:  "origin",
			expectedSuccess: true,
			description:     "Should match origin field in JSON response",
		},
		{
			name:            "json_nonexistent_failure_pattern",
			endpoint:        helper.CommonHTTPBinScenarios()[4].Endpoint,
			successPattern:  "nonexistent_field_xyz",
			expectedSuccess: true, // HTTPBin returns exit code 0, so fallback to exit code
			description:     "Should fallback to exit code when pattern doesn't match",
		},
		{
			name:            "headers_user_agent_pattern",
			endpoint:        httpbinTest.NewHTTPBinEndpoints("https://httpbin.org").Headers(),
			successPattern:  "User-Agent",
			expectedSuccess: true,
			description:     "Should match User-Agent in headers response",
		},
		{
			name:            "case_insensitive_pattern_matching",
			endpoint:        helper.CommonHTTPBinScenarios()[4].Endpoint,
			successPattern:  "(?i)SLIDESHOW", // Case insensitive
			expectedSuccess: true,
			description:     "Should match case-insensitive patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create pattern matcher
			config := PatternConfig{
				SuccessPattern: tt.successPattern,
				FailurePattern: tt.failurePattern,
			}

			matcher, err := NewPatternMatcher(config)
			require.NoError(t, err, "Failed to create pattern matcher")

			// Note: For pattern testing, we simulate the response content that HTTPBin typically returns
			// In production, this would come from actual curl execution via our executor
			var simulatedOutput string

			if strings.Contains(tt.endpoint, "json") {
				// Simulate HTTPBin JSON response structure
				simulatedOutput = `{
					"slideshow": {
						"title": "Sample Slide Show"
					},
					"origin": "192.168.1.1"
				}`
			} else if strings.Contains(tt.endpoint, "headers") {
				// Simulate headers response
				simulatedOutput = `{
					"headers": {
						"User-Agent": "curl/7.68.0",
						"Accept": "*/*"
					}
				}`
			} else {
				simulatedOutput = "test output"
			}

			// Test pattern evaluation
			result := matcher.EvaluateResult(simulatedOutput, 0)

			assert.Equal(t, tt.expectedSuccess, result.Success,
				"Pattern matching result for %s", tt.description)

			if tt.expectedSuccess {
				assert.Equal(t, 0, result.ExitCode, "Success should return exit code 0")
				if tt.successPattern == "nonexistent_field_xyz" {
					assert.Contains(t, result.Reason, "exit code used", "Should indicate fallback to exit code")
				} else {
					assert.Contains(t, result.Reason, "pattern matched", "Should indicate pattern match")
				}
			}

			t.Logf("✅ %s: Pattern '%s' matching result: %v (reason: %s)",
				tt.description, tt.successPattern, result.Success, result.Reason)
		})
	}
}

// TestPatternMatcher_HTTPBin_ErrorResponses tests pattern matching with HTTP error responses
func TestPatternMatcher_HTTPBin_ErrorResponses(t *testing.T) {
	t.Parallel()

	helper := httpbinTest.NewHTTPBinHelper(nil)
	helper.SkipIfNoNetwork(t)

	errorScenarios := []struct {
		name           string
		statusCode     int
		failurePattern string
		successPattern string
		expectedResult bool
		description    string
	}{
		{
			name:           "503_service_unavailable",
			statusCode:     503,
			failurePattern: "503|Service Unavailable",
			expectedResult: false, // Should match failure pattern
			description:    "Should detect 503 service unavailable responses",
		},
		{
			name:           "429_rate_limit_detection",
			statusCode:     429,
			failurePattern: "429|Too Many Requests",
			expectedResult: false, // Should match failure pattern
			description:    "Should detect 429 rate limiting responses",
		},
		{
			name:           "404_not_found_handling",
			statusCode:     404,
			failurePattern: "404|Not Found",
			expectedResult: false, // Should match failure pattern
			description:    "Should handle 404 not found responses",
		},
		{
			name:           "200_success_override",
			statusCode:     200,
			successPattern: "200|OK",
			expectedResult: true, // Should match success pattern
			description:    "Should override exit code with success pattern",
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Create pattern matcher
			config := PatternConfig{
				SuccessPattern: scenario.successPattern,
				FailurePattern: scenario.failurePattern,
			}

			matcher, err := NewPatternMatcher(config)
			require.NoError(t, err, "Failed to create pattern matcher")

			// Simulate HTTP error response content
			var simulatedOutput string
			var simulatedExitCode int

			switch scenario.statusCode {
			case 503:
				simulatedOutput = "HTTP/1.1 503 Service Unavailable\r\n\r\n"
				simulatedExitCode = 1
			case 429:
				simulatedOutput = "HTTP/1.1 429 Too Many Requests\r\n\r\n"
				simulatedExitCode = 1
			case 404:
				simulatedOutput = "HTTP/1.1 404 Not Found\r\n\r\n"
				simulatedExitCode = 1
			case 200:
				simulatedOutput = "HTTP/1.1 200 OK\r\n\r\n"
				simulatedExitCode = 0
			default:
				simulatedOutput = "Unknown response"
				simulatedExitCode = 1
			}

			// Test pattern evaluation
			result := matcher.EvaluateResult(simulatedOutput, simulatedExitCode)

			assert.Equal(t, scenario.expectedResult, result.Success,
				"Error response pattern matching for %s", scenario.description)

			// Verify exit code transformation
			if scenario.expectedResult {
				assert.Equal(t, 0, result.ExitCode, "Success pattern should return exit code 0")
			} else {
				assert.Equal(t, 1, result.ExitCode, "Failure pattern should return exit code 1")
			}

			t.Logf("✅ %s: HTTP %d response handled correctly (success: %v, reason: %s)",
				scenario.description, scenario.statusCode, result.Success, result.Reason)
		})
	}
}

// TestPatternMatcher_HTTPBin_ComplexPatterns tests complex regex patterns with real data
func TestPatternMatcher_HTTPBin_ComplexPatterns(t *testing.T) {
	t.Parallel()

	helper := httpbinTest.NewHTTPBinHelper(nil)
	helper.SkipIfNoNetwork(t)

	complexTests := []struct {
		name        string
		pattern     string
		testData    string
		shouldMatch bool
		description string
	}{
		{
			name:        "json_nested_field_extraction",
			pattern:     `"slideshow"\s*:\s*{[^}]*"title"\s*:\s*"[^"]*"`,
			testData:    `{"slideshow": {"title": "Sample Slide Show", "author": "Test"}}`,
			shouldMatch: true,
			description: "Should match nested JSON structure patterns",
		},
		{
			name:        "http_status_code_pattern",
			pattern:     `HTTP/1\.[01]\s+(200|201|202)\s+`,
			testData:    "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n",
			shouldMatch: true,
			description: "Should match HTTP success status patterns",
		},
		{
			name:        "ip_address_extraction",
			pattern:     `"origin"\s*:\s*"(\d{1,3}\.){3}\d{1,3}"`,
			testData:    `{"origin": "192.168.1.100", "url": "https://example.com"}`,
			shouldMatch: true,
			description: "Should extract IP addresses from JSON responses",
		},
		{
			name:        "error_message_detection",
			pattern:     `(error|fail|exception|timeout)`,
			testData:    "Request completed successfully",
			shouldMatch: false,
			description: "Should not match success messages for error patterns",
		},
		{
			name:        "case_insensitive_header_matching",
			pattern:     `(?i)content-type\s*:\s*application/json`,
			testData:    "Content-Type: application/json\r\nContent-Length: 123",
			shouldMatch: true,
			description: "Should match headers case-insensitively",
		},
	}

	for _, tt := range complexTests {
		t.Run(tt.name, func(t *testing.T) {
			// Create pattern matcher with complex pattern
			config := PatternConfig{
				SuccessPattern: tt.pattern,
			}

			matcher, err := NewPatternMatcher(config)
			require.NoError(t, err, "Failed to create pattern matcher for complex pattern: %s", tt.pattern)

			// Test pattern evaluation
			result := matcher.EvaluateResult(tt.testData, 1) // Start with failure exit code

			assert.Equal(t, tt.shouldMatch, result.Success,
				"Complex pattern matching for %s", tt.description)

			if tt.shouldMatch {
				assert.Equal(t, 0, result.ExitCode, "Pattern match should override exit code to 0")
				assert.Contains(t, result.Reason, "success pattern", "Should indicate success pattern match")
			} else {
				assert.Equal(t, 1, result.ExitCode, "No pattern match should preserve original exit code")
			}

			t.Logf("✅ %s: Complex pattern '%s' evaluation completed (match: %v)",
				tt.description, tt.pattern, result.Success)
		})
	}
}

// TestPatternMatcher_HTTPBin_PerformanceWithRealData tests performance with realistic data sizes
func TestPatternMatcher_HTTPBin_PerformanceWithRealData(t *testing.T) {
	helper := httpbinTest.NewHTTPBinHelper(nil)
	helper.SkipIfNoNetwork(t)

	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create pattern matcher for JSON content
	config := PatternConfig{
		SuccessPattern: "slideshow|origin",
		FailurePattern: "error|failed",
	}

	matcher, err := NewPatternMatcher(config)
	require.NoError(t, err)

	// Simulate large JSON response (like HTTPBin might return)
	largeJSONResponse := strings.Repeat(`{
		"slideshow": {
			"title": "Sample Slide Show",
			"author": "Yours Truly", 
			"date": "date of publication",
			"slides": [
				{"title": "Wake up to WonderWidgets!", "type": "all"},
				{"title": "Overview", "type": "all", "items": ["Why WonderWidgets?", "Who buys them"]}
			]
		},
		"origin": "192.168.1.100"
	}`, 100) // 100 copies for realistic size

	// Performance test
	iterations := 1000
	startTime := time.Now()

	for i := 0; i < iterations; i++ {
		result := matcher.EvaluateResult(largeJSONResponse, 0)
		assert.True(t, result.Success, "Should match success pattern in large response")
	}

	duration := time.Since(startTime)
	avgTime := duration / time.Duration(iterations)

	// Performance assertions
	assert.Less(t, avgTime, 10*time.Millisecond,
		"Pattern matching should be fast even with large responses")

	t.Logf("✅ Performance test: %d pattern evaluations in %v (avg: %v per evaluation)",
		iterations, duration, avgTime)
}

// BenchmarkPatternMatcher_HTTPBin_RealWorldJSON benchmarks pattern matching with realistic JSON
func BenchmarkPatternMatcher_HTTPBin_RealWorldJSON(b *testing.B) {
	config := PatternConfig{
		SuccessPattern: "slideshow|origin|args",
		FailurePattern: "error|failed|exception",
	}

	matcher, err := NewPatternMatcher(config)
	if err != nil {
		b.Fatalf("Failed to create pattern matcher: %v", err)
	}

	// Realistic HTTPBin JSON response
	realWorldJSON := `{
		"args": {}, 
		"headers": {
			"Accept": "*/*", 
			"Host": "httpbin.org", 
			"User-Agent": "curl/7.68.0"
		}, 
		"origin": "192.168.1.100", 
		"url": "https://httpbin.org/get"
	}`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := matcher.EvaluateResult(realWorldJSON, 0)
		_ = result.Success
	}
}

// BenchmarkPatternMatcher_HTTPBin_LargeResponse benchmarks with large HTTP responses
func BenchmarkPatternMatcher_HTTPBin_LargeResponse(b *testing.B) {
	config := PatternConfig{
		SuccessPattern: "slideshow",
		FailurePattern: "error",
	}

	matcher, err := NewPatternMatcher(config)
	if err != nil {
		b.Fatalf("Failed to create pattern matcher: %v", err)
	}

	// Large response similar to what HTTPBin might return
	largeResponse := strings.Repeat("This is a line of HTTP response data. ", 1000) + "slideshow content here"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := matcher.EvaluateResult(largeResponse, 0)
		_ = result.Success
	}
}
