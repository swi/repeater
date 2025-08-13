package httpaware

import (
	"testing"
	"time"
)

func TestHTTPAwareIntegration_EndToEndScenarios(t *testing.T) {
	tests := []struct {
		name             string
		responses        []string
		expectedDelays   []time.Duration
		fallbackInterval time.Duration
		description      string
	}{
		{
			name: "api_monitoring_with_rate_limits",
			responses: []string{
				"HTTP/1.1 200 OK\r\n\r\n{\"status\": \"healthy\"}",              // Success - use fallback
				"HTTP/1.1 429 Too Many Requests\r\nRetry-After: 60\r\n\r\n",     // Rate limited - wait 60s
				"HTTP/1.1 200 OK\r\n\r\n{\"status\": \"healthy\"}",              // Success - use fallback
				"HTTP/1.1 503 Service Unavailable\r\n\r\n{\"retry_after\": 30}", // Server error - wait 30s
				"HTTP/1.1 200 OK\r\n\r\n{\"status\": \"healthy\"}",              // Success - use fallback
			},
			expectedDelays: []time.Duration{
				15 * time.Second, // Fallback interval
				60 * time.Second, // HTTP rate limit
				15 * time.Second, // Fallback interval
				30 * time.Second, // HTTP server error
				15 * time.Second, // Fallback interval
			},
			fallbackInterval: 15 * time.Second,
			description:      "API monitoring should handle rate limits and server errors intelligently",
		},
		{
			name: "github_api_monitoring",
			responses: []string{
				"HTTP/1.1 200 OK\r\n\r\n{\"rate\": {\"remaining\": 4999}}",                                      // Success
				"HTTP/1.1 403 Forbidden\r\nRetry-After: 3600\r\n\r\n{\"message\": \"API rate limit exceeded\"}", // Rate limited
				"HTTP/1.1 200 OK\r\n\r\n{\"rate\": {\"remaining\": 4999}}",                                      // Success after wait
			},
			expectedDelays: []time.Duration{
				10 * time.Second,   // Fallback interval
				3600 * time.Second, // GitHub rate limit (1 hour)
				10 * time.Second,   // Fallback interval
			},
			fallbackInterval: 10 * time.Second,
			description:      "GitHub API monitoring should respect rate limits",
		},
		{
			name: "aws_api_with_exponential_backoff",
			responses: []string{
				"HTTP/1.1 200 OK\r\n\r\n{\"status\": \"success\"}",                                              // Success
				"HTTP/1.1 429 Too Many Requests\r\nRetry-After: 1\r\n\r\n{\"__type\": \"ThrottlingException\"}", // Throttled
				"HTTP/1.1 429 Too Many Requests\r\nRetry-After: 2\r\n\r\n{\"__type\": \"ThrottlingException\"}", // Throttled again
				"HTTP/1.1 200 OK\r\n\r\n{\"status\": \"success\"}",                                              // Success
			},
			expectedDelays: []time.Duration{
				5 * time.Second, // Fallback interval
				1 * time.Second, // AWS throttling (1s)
				2 * time.Second, // AWS throttling (2s)
				5 * time.Second, // Fallback interval
			},
			fallbackInterval: 5 * time.Second,
			description:      "AWS API should handle throttling with HTTP timing",
		},
		{
			name: "mixed_http_and_non_http_commands",
			responses: []string{
				"Service is healthy\nUptime: 24 hours",                        // Non-HTTP response
				"HTTP/1.1 503 Service Unavailable\r\nRetry-After: 45\r\n\r\n", // HTTP response
				"pong", // Non-HTTP response
				"HTTP/1.1 200 OK\r\n\r\n{\"status\": \"ok\"}", // HTTP response
			},
			expectedDelays: []time.Duration{
				20 * time.Second, // Fallback for non-HTTP
				45 * time.Second, // HTTP timing
				20 * time.Second, // Fallback for non-HTTP
				20 * time.Second, // Fallback for HTTP success
			},
			fallbackInterval: 20 * time.Second,
			description:      "Should handle mixed HTTP and non-HTTP responses correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HTTPAwareConfig{
				MaxDelay: 2 * time.Hour,
				MinDelay: 1 * time.Second,
			}

			scheduler := NewHTTPAwareScheduler(config)
			fallbackScheduler := createIntervalScheduler(tt.fallbackInterval)
			scheduler.SetFallbackScheduler(fallbackScheduler)

			for i, response := range tt.responses {
				scheduler.SetLastResponse(response)
				delay := scheduler.NextDelay()

				expectedDelay := tt.expectedDelays[i]
				if delay != expectedDelay {
					t.Errorf("Response %d: Expected delay %v, got %v", i+1, expectedDelay, delay)
					t.Logf("Response was: %q", response)
				}
			}
		})
	}
}

func TestHTTPAwareIntegration_ConfigurationOptions(t *testing.T) {
	tests := []struct {
		name             string
		config           HTTPAwareConfig
		response         string
		expectedDelay    time.Duration
		expectTimingInfo bool
		description      string
	}{
		{
			name: "json_parsing_disabled",
			config: HTTPAwareConfig{
				MaxDelay:     10 * time.Minute,
				MinDelay:     1 * time.Second,
				ParseJSON:    false,
				ParseHeaders: true,
			},
			response: "HTTP/1.1 503 Service Unavailable\r\n" +
				"Retry-After: 30\r\n" +
				"\r\n" +
				`{"retry_after": 60}`, // Should be ignored
			expectedDelay:    30 * time.Second, // Only header timing
			expectTimingInfo: true,
			description:      "Should ignore JSON when parsing is disabled",
		},
		{
			name: "header_parsing_disabled",
			config: HTTPAwareConfig{
				MaxDelay:     10 * time.Minute,
				MinDelay:     1 * time.Second,
				ParseJSON:    true,
				ParseHeaders: false,
			},
			response: "HTTP/1.1 503 Service Unavailable\r\n" +
				"Retry-After: 30\r\n" + // Should be ignored
				"\r\n" +
				`{"retry_after": 60}`,
			expectedDelay:    60 * time.Second, // Only JSON timing
			expectTimingInfo: true,
			description:      "Should ignore headers when parsing is disabled",
		},
		{
			name: "trust_client_errors_enabled",
			config: HTTPAwareConfig{
				MaxDelay:          10 * time.Minute,
				MinDelay:          1 * time.Second,
				ParseJSON:         true,
				ParseHeaders:      true,
				TrustClientErrors: true,
			},
			response: "HTTP/1.1 404 Not Found\r\n" +
				"Retry-After: 15\r\n" +
				"\r\n",
			expectedDelay:    15 * time.Second, // Should trust 4xx timing
			expectTimingInfo: true,
			description:      "Should trust client error timing when enabled",
		},
		{
			name: "custom_json_fields",
			config: HTTPAwareConfig{
				MaxDelay:     10 * time.Minute,
				MinDelay:     1 * time.Second,
				ParseJSON:    true,
				ParseHeaders: true,
				JSONFields:   []string{"custom_retry", "backoff_seconds"},
			},
			response: "HTTP/1.1 503 Service Unavailable\r\n" +
				"\r\n" +
				`{"custom_retry": 42, "retry_after": 30}`, // Should use custom field
			expectedDelay:    42 * time.Second,
			expectTimingInfo: true,
			description:      "Should use custom JSON fields when configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler := NewHTTPAwareSchedulerWithConfig(tt.config)
			fallbackScheduler := createIntervalScheduler(10 * time.Second)
			scheduler.SetFallbackScheduler(fallbackScheduler)

			scheduler.SetLastResponse(tt.response)
			delay := scheduler.NextDelay()

			if delay != tt.expectedDelay {
				t.Errorf("Expected delay %v, got %v", tt.expectedDelay, delay)
			}

			timingInfo := scheduler.GetTimingInfo()
			if tt.expectTimingInfo && timingInfo == nil {
				t.Errorf("Expected timing info, got nil")
			} else if !tt.expectTimingInfo && timingInfo != nil {
				t.Errorf("Expected no timing info, got: %+v", timingInfo)
			}
		})
	}
}

func TestHTTPAwareIntegration_ErrorHandling(t *testing.T) {
	config := HTTPAwareConfig{
		MaxDelay: 5 * time.Minute,
		MinDelay: 1 * time.Second,
	}

	scheduler := NewHTTPAwareScheduler(config)
	fallbackScheduler := createIntervalScheduler(30 * time.Second)
	scheduler.SetFallbackScheduler(fallbackScheduler)

	tests := []struct {
		name          string
		response      string
		expectedDelay time.Duration
		description   string
	}{
		{
			name:          "malformed_http_response",
			response:      "HTTP/1.1 503\r\nRetry-After: invalid\r\n\r\n",
			expectedDelay: 30 * time.Second, // Fallback
			description:   "Should handle malformed HTTP responses gracefully",
		},
		{
			name:          "invalid_json",
			response:      "HTTP/1.1 503 Service Unavailable\r\n\r\n{invalid json}",
			expectedDelay: 30 * time.Second, // Fallback
			description:   "Should handle invalid JSON gracefully",
		},
		{
			name:          "empty_response",
			response:      "",
			expectedDelay: 30 * time.Second, // Fallback
			description:   "Should handle empty responses gracefully",
		},
		{
			name:          "extremely_large_delay",
			response:      "HTTP/1.1 503 Service Unavailable\r\nRetry-After: 999999\r\n\r\n",
			expectedDelay: 5 * time.Minute, // Capped at MaxDelay
			description:   "Should cap extremely large delays",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler.SetLastResponse(tt.response)
			delay := scheduler.NextDelay()

			if delay != tt.expectedDelay {
				t.Errorf("Expected delay %v, got %v", tt.expectedDelay, delay)
			}
		})
	}
}
