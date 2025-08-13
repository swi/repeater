package httpaware

import (
	"testing"
	"time"
)

func TestHTTPResponseParser_RetryAfterHeader(t *testing.T) {
	parser := NewHTTPResponseParser()

	tests := []struct {
		name          string
		httpResponse  string
		expectedDelay time.Duration
		expectedFound bool
		description   string
	}{
		{
			name: "retry_after_seconds_integer",
			httpResponse: "HTTP/1.1 503 Service Unavailable\r\n" +
				"Retry-After: 30\r\n" +
				"\r\n",
			expectedDelay: 30 * time.Second,
			expectedFound: true,
			description:   "Integer seconds should be parsed correctly",
		},
		{
			name: "retry_after_zero",
			httpResponse: "HTTP/1.1 503 Service Unavailable\r\n" +
				"Retry-After: 0\r\n" +
				"\r\n",
			expectedDelay: 1 * time.Second, // Minimum delay
			expectedFound: true,
			description:   "Zero delay should be clamped to minimum",
		},
		{
			name: "retry_after_negative",
			httpResponse: "HTTP/1.1 503 Service Unavailable\r\n" +
				"Retry-After: -30\r\n" +
				"\r\n",
			expectedDelay: 0,
			expectedFound: false,
			description:   "Negative delays should be ignored",
		},
		{
			name: "no_retry_after_header",
			httpResponse: "HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n",
			expectedDelay: 0,
			expectedFound: false,
			description:   "Missing header should return not found",
		},
		{
			name: "success_response_with_retry_after",
			httpResponse: "HTTP/1.1 200 OK\r\n" +
				"Retry-After: 30\r\n" +
				"\r\n",
			expectedDelay: 0,
			expectedFound: false,
			description:   "Success responses should ignore Retry-After",
		},
		{
			name: "client_error_with_retry_after",
			httpResponse: "HTTP/1.1 404 Not Found\r\n" +
				"Retry-After: 30\r\n" +
				"\r\n",
			expectedDelay: 0,
			expectedFound: false,
			description:   "Client errors should ignore Retry-After",
		},
		{
			name: "rate_limit_with_retry_after",
			httpResponse: "HTTP/1.1 429 Too Many Requests\r\n" +
				"Retry-After: 120\r\n" +
				"\r\n",
			expectedDelay: 120 * time.Second,
			expectedFound: true,
			description:   "Rate limiting should respect Retry-After",
		},
		{
			name: "server_error_with_retry_after",
			httpResponse: "HTTP/1.1 502 Bad Gateway\r\n" +
				"Retry-After: 60\r\n" +
				"\r\n",
			expectedDelay: 60 * time.Second,
			expectedFound: true,
			description:   "Server errors should respect Retry-After",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timingInfo, err := parser.ParseResponse(tt.httpResponse)

			if tt.expectedFound {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if timingInfo == nil {
					t.Errorf("Expected timing info, got nil")
					return
				}
				if timingInfo.Delay != tt.expectedDelay {
					t.Errorf("Expected delay %v, got %v", tt.expectedDelay, timingInfo.Delay)
				}
				if timingInfo.Source != TimingSourceRetryAfterHeader {
					t.Errorf("Expected source %v, got %v", TimingSourceRetryAfterHeader, timingInfo.Source)
				}
			} else {
				if timingInfo != nil {
					t.Errorf("Expected no timing info, got: %+v", timingInfo)
				}
			}
		})
	}
}

func TestHTTPResponseParser_JSONRetryInfo(t *testing.T) {
	parser := NewHTTPResponseParser()

	tests := []struct {
		name           string
		httpResponse   string
		expectedDelay  time.Duration
		expectedFound  bool
		expectedSource TimingSource
		description    string
	}{
		{
			name: "json_retry_after_snake_case",
			httpResponse: "HTTP/1.1 503 Service Unavailable\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"retry_after": 45}`,
			expectedDelay:  45 * time.Second,
			expectedFound:  true,
			expectedSource: TimingSourceJSONRetryAfter,
			description:    "Snake case retry_after should be parsed",
		},
		{
			name: "json_retry_after_camel_case",
			httpResponse: "HTTP/1.1 503 Service Unavailable\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"retryAfter": 60}`,
			expectedDelay:  60 * time.Second,
			expectedFound:  true,
			expectedSource: TimingSourceJSONRetryAfter,
			description:    "CamelCase retryAfter should be parsed",
		},
		{
			name: "json_nested_retry_info",
			httpResponse: "HTTP/1.1 429 Too Many Requests\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"error": {"retry_after": 120, "message": "Rate limited"}}`,
			expectedDelay:  120 * time.Second,
			expectedFound:  true,
			expectedSource: TimingSourceJSONRetryAfter,
			description:    "Nested retry information should be found",
		},
		{
			name: "json_rate_limit_reset",
			httpResponse: "HTTP/1.1 429 Too Many Requests\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"rate_limit": {"reset_in": 90, "remaining": 0}}`,
			expectedDelay:  90 * time.Second,
			expectedFound:  true,
			expectedSource: TimingSourceJSONRateLimit,
			description:    "Rate limit reset timing should be used",
		},
		{
			name: "json_backoff_delay",
			httpResponse: "HTTP/1.1 503 Service Unavailable\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"backoff": {"delay": 75}}`,
			expectedDelay:  75 * time.Second,
			expectedFound:  true,
			expectedSource: TimingSourceJSONBackoff,
			description:    "Backoff delay should be parsed",
		},
		{
			name: "json_malformed",
			httpResponse: "HTTP/1.1 503 Service Unavailable\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"retry_after": invalid json}`,
			expectedDelay: 0,
			expectedFound: false,
			description:   "Malformed JSON should be ignored gracefully",
		},
		{
			name: "json_no_retry_info",
			httpResponse: "HTTP/1.1 503 Service Unavailable\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"message": "Service temporarily unavailable"}`,
			expectedDelay: 0,
			expectedFound: false,
			description:   "JSON without retry info should return not found",
		},
		{
			name: "json_success_response_ignored",
			httpResponse: "HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"retry_after": 30}`,
			expectedDelay: 0,
			expectedFound: false,
			description:   "Success responses should ignore JSON timing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timingInfo, err := parser.ParseResponse(tt.httpResponse)

			if tt.expectedFound {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if timingInfo == nil {
					t.Errorf("Expected timing info, got nil")
					return
				}
				if timingInfo.Delay != tt.expectedDelay {
					t.Errorf("Expected delay %v, got %v", tt.expectedDelay, timingInfo.Delay)
				}
				if timingInfo.Source != tt.expectedSource {
					t.Errorf("Expected source %v, got %v", tt.expectedSource, timingInfo.Source)
				}
			} else {
				if timingInfo != nil {
					t.Errorf("Expected no timing info, got: %+v", timingInfo)
				}
			}
		})
	}
}

func TestHTTPResponseParser_PriorityOrder(t *testing.T) {
	parser := NewHTTPResponseParser()

	tests := []struct {
		name           string
		httpResponse   string
		expectedDelay  time.Duration
		expectedSource TimingSource
		description    string
	}{
		{
			name: "header_overrides_json",
			httpResponse: "HTTP/1.1 503 Service Unavailable\r\n" +
				"Retry-After: 30\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"retry_after": 60}`,
			expectedDelay:  30 * time.Second,
			expectedSource: TimingSourceRetryAfterHeader,
			description:    "Retry-After header should take priority over JSON",
		},
		{
			name: "json_retry_after_overrides_rate_limit",
			httpResponse: "HTTP/1.1 429 Too Many Requests\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				`{"retry_after": 45, "rate_limit": {"reset_in": 90}}`,
			expectedDelay:  45 * time.Second,
			expectedSource: TimingSourceJSONRetryAfter,
			description:    "JSON retry_after should take priority over rate_limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timingInfo, err := parser.ParseResponse(tt.httpResponse)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if timingInfo == nil {
				t.Errorf("Expected timing info, got nil")
				return
			}
			if timingInfo.Delay != tt.expectedDelay {
				t.Errorf("Expected delay %v, got %v", tt.expectedDelay, timingInfo.Delay)
			}
			if timingInfo.Source != tt.expectedSource {
				t.Errorf("Expected source %v, got %v", tt.expectedSource, timingInfo.Source)
			}
		})
	}
}

func TestHTTPResponseParser_RealWorldAPIs(t *testing.T) {
	parser := NewHTTPResponseParser()

	tests := []struct {
		name          string
		apiResponse   string
		expectedDelay time.Duration
		description   string
	}{
		{
			name: "github_api_rate_limit",
			apiResponse: "HTTP/1.1 403 Forbidden\r\n" +
				"X-RateLimit-Remaining: 0\r\n" +
				"Retry-After: 3600\r\n" +
				"\r\n" +
				`{"message": "API rate limit exceeded"}`,
			expectedDelay: 3600 * time.Second,
			description:   "GitHub API rate limiting should be respected",
		},
		{
			name: "aws_api_throttling",
			apiResponse: "HTTP/1.1 429 Too Many Requests\r\n" +
				"Retry-After: 120\r\n" +
				"\r\n" +
				`{"__type": "ThrottlingException", "message": "Rate exceeded"}`,
			expectedDelay: 120 * time.Second,
			description:   "AWS API throttling should be handled",
		},
		{
			name: "stripe_api_rate_limit",
			apiResponse: "HTTP/1.1 429 Too Many Requests\r\n" +
				"Retry-After: 1\r\n" +
				"\r\n" +
				`{"error": {"type": "rate_limit_error", "message": "Too many requests"}}`,
			expectedDelay: 1 * time.Second,
			description:   "Stripe API rate limiting should be handled",
		},
		{
			name: "discord_api_rate_limit",
			apiResponse: "HTTP/1.1 429 Too Many Requests\r\n" +
				"\r\n" +
				`{"retry_after": 64.57, "global": false}`,
			expectedDelay: 65 * time.Second, // Rounded up
			description:   "Discord API fractional retry_after should be handled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timingInfo, err := parser.ParseResponse(tt.apiResponse)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if timingInfo == nil {
				t.Errorf("Expected timing info, got nil")
				return
			}
			if timingInfo.Delay != tt.expectedDelay {
				t.Errorf("Expected delay %v, got %v", tt.expectedDelay, timingInfo.Delay)
			}
		})
	}
}

func TestHTTPResponseParser_NonHTTPResponse(t *testing.T) {
	parser := NewHTTPResponseParser()

	tests := []struct {
		name     string
		response string
	}{
		{
			name:     "plain_text_output",
			response: "Service is healthy\nUptime: 24 hours",
		},
		{
			name:     "json_without_http_headers",
			response: `{"status": "ok", "uptime": "24h"}`,
		},
		{
			name:     "empty_response",
			response: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			supported := parser.SupportsResponse(tt.response)
			if supported {
				t.Errorf("Expected non-HTTP response to not be supported")
			}

			timingInfo, err := parser.ParseResponse(tt.response)
			if timingInfo != nil {
				t.Errorf("Expected no timing info for non-HTTP response, got: %+v", timingInfo)
			}
			if err != nil {
				t.Errorf("Expected no error for non-HTTP response, got: %v", err)
			}
		})
	}
}
