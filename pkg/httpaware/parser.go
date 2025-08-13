package httpaware

import (
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// httpResponseParser implements HTTPResponseParser interface
type httpResponseParser struct {
	config HTTPAwareConfig
}

// NewHTTPResponseParser creates a new HTTP response parser with default configuration
func NewHTTPResponseParser() HTTPResponseParser {
	return &httpResponseParser{
		config: HTTPAwareConfig{
			ParseJSON:         true,
			ParseHeaders:      true,
			TrustClientErrors: false,
			JSONFields:        []string{"retry_after", "retryAfter"},
			HeaderNames:       []string{"Retry-After"},
		},
	}
}

// NewHTTPResponseParserWithConfig creates a new HTTP response parser with custom configuration
func NewHTTPResponseParserWithConfig(config HTTPAwareConfig) HTTPResponseParser {
	return &httpResponseParser{
		config: config,
	}
}

// ParseResponse extracts timing information from an HTTP response
func (p *httpResponseParser) ParseResponse(response string) (*TimingInfo, error) {
	if !p.SupportsResponse(response) {
		return nil, nil
	}

	// Try to extract timing information in priority order
	if p.config.ParseHeaders {
		if timingInfo := p.parseRetryAfterHeader(response); timingInfo != nil {
			return timingInfo, nil
		}
	}

	if p.config.ParseJSON {
		if timingInfo := p.parseJSONTiming(response); timingInfo != nil {
			return timingInfo, nil
		}
	}

	return nil, nil
}

// SupportsResponse returns true if the response appears to be HTTP
func (p *httpResponseParser) SupportsResponse(response string) bool {
	return strings.HasPrefix(response, "HTTP/")
}

// parseRetryAfterHeader extracts timing from Retry-After header
func (p *httpResponseParser) parseRetryAfterHeader(response string) *TimingInfo {
	// Extract status code first
	statusCode := p.extractStatusCode(response)

	// Only process server errors (5xx) and rate limiting (429, 403 for GitHub)
	// Skip success (2xx) and client errors (4xx) unless configured
	if statusCode >= 200 && statusCode < 300 {
		return nil // Success responses
	}
	if statusCode >= 400 && statusCode < 500 && statusCode != 429 && statusCode != 403 && !p.config.TrustClientErrors {
		return nil // Client errors (except 429 and 403 for rate limiting)
	}

	// Find Retry-After header
	headerRegex := regexp.MustCompile(`(?i)Retry-After:\s*(\d+)`)
	matches := headerRegex.FindStringSubmatch(response)
	if len(matches) < 2 {
		return nil
	}

	seconds, err := strconv.Atoi(matches[1])
	if err != nil || seconds < 0 {
		return nil
	}

	delay := time.Duration(seconds) * time.Second

	// Apply min delay constraint
	if delay == 0 {
		delay = 1 * time.Second // Minimum delay
	}

	return &TimingInfo{
		Delay:      delay,
		Source:     TimingSourceRetryAfterHeader,
		Confidence: 1.0,
	}
}

// parseJSONTiming extracts timing from JSON response body
func (p *httpResponseParser) parseJSONTiming(response string) *TimingInfo {
	// Extract status code first
	statusCode := p.extractStatusCode(response)

	// Only process server errors (5xx) and rate limiting (429, 403 for GitHub)
	if statusCode >= 200 && statusCode < 300 {
		return nil // Success responses
	}
	if statusCode >= 400 && statusCode < 500 && statusCode != 429 && statusCode != 403 && !p.config.TrustClientErrors {
		return nil // Client errors (except 429 and 403 for rate limiting)
	}

	// Extract JSON body
	jsonBody := p.extractJSONBody(response)
	if jsonBody == "" {
		return nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonBody), &data); err != nil {
		return nil
	}

	// Check custom fields first (highest priority)
	for _, field := range p.config.JSONFields {
		if delay := p.extractDelayFromJSON(data, field); delay > 0 {
			return &TimingInfo{
				Delay:      delay,
				Source:     TimingSourceJSONRetryAfter,
				Confidence: 1.0,
			}
		}
	}

	// Try different JSON field patterns in priority order
	if delay := p.extractDelayFromJSON(data, "retry_after"); delay > 0 {
		return &TimingInfo{
			Delay:      delay,
			Source:     TimingSourceJSONRetryAfter,
			Confidence: 0.9,
		}
	}

	if delay := p.extractDelayFromJSON(data, "retryAfter"); delay > 0 {
		return &TimingInfo{
			Delay:      delay,
			Source:     TimingSourceJSONRetryAfter,
			Confidence: 0.9,
		}
	}

	// Check nested structures
	if errorData, ok := data["error"].(map[string]interface{}); ok {
		if delay := p.extractDelayFromJSON(errorData, "retry_after"); delay > 0 {
			return &TimingInfo{
				Delay:      delay,
				Source:     TimingSourceJSONRetryAfter,
				Confidence: 0.8,
			}
		}
	}

	// Check rate limit structures
	if rateLimitData, ok := data["rate_limit"].(map[string]interface{}); ok {
		if delay := p.extractDelayFromJSON(rateLimitData, "reset_in"); delay > 0 {
			return &TimingInfo{
				Delay:      delay,
				Source:     TimingSourceJSONRateLimit,
				Confidence: 0.8,
			}
		}
	}

	// Check backoff structures
	if backoffData, ok := data["backoff"].(map[string]interface{}); ok {
		if delay := p.extractDelayFromJSON(backoffData, "delay"); delay > 0 {
			return &TimingInfo{
				Delay:      delay,
				Source:     TimingSourceJSONBackoff,
				Confidence: 0.8,
			}
		}
	}

	return nil
}

// extractStatusCode extracts HTTP status code from response
func (p *httpResponseParser) extractStatusCode(response string) int {
	statusRegex := regexp.MustCompile(`HTTP/\d\.\d\s+(\d+)`)
	matches := statusRegex.FindStringSubmatch(response)
	if len(matches) < 2 {
		return 0
	}

	statusCode, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}

	return statusCode
}

// extractJSONBody extracts JSON body from HTTP response
func (p *httpResponseParser) extractJSONBody(response string) string {
	// Find the end of headers (double CRLF)
	headerEndIndex := strings.Index(response, "\r\n\r\n")
	if headerEndIndex == -1 {
		return ""
	}

	body := response[headerEndIndex+4:]
	body = strings.TrimSpace(body)

	// Basic JSON validation
	if !strings.HasPrefix(body, "{") || !strings.HasSuffix(body, "}") {
		return ""
	}

	return body
}

// extractDelayFromJSON extracts delay value from JSON data
func (p *httpResponseParser) extractDelayFromJSON(data map[string]interface{}, field string) time.Duration {
	value, exists := data[field]
	if !exists {
		return 0
	}

	var seconds float64
	switch v := value.(type) {
	case float64:
		seconds = v
	case int:
		seconds = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		seconds = parsed
	default:
		return 0
	}

	if seconds <= 0 {
		return 0
	}

	// Round up fractional seconds (like Discord API)
	return time.Duration(math.Ceil(seconds)) * time.Second
}
