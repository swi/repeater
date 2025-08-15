package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/httpaware"
)

func TestCLI_HTTPAwareFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
	}{
		{
			name: "basic http-aware flag",
			args: []string{"interval", "--every", "30s", "--http-aware", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:       "interval",
				Every:            30 * time.Second,
				HTTPAware:        true,
				HTTPParseJSON:    true,  // Default
				HTTPParseHeaders: true,  // Default
				HTTPTrustClient:  false, // Default
				Command:          []string{"curl", "api.com"},
			},
		},
		{
			name: "http-aware with custom delays",
			args: []string{"interval", "--every", "30s", "--http-aware", "--http-max-delay", "10m", "--http-min-delay", "5s", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:       "interval",
				Every:            30 * time.Second,
				HTTPAware:        true,
				HTTPMaxDelay:     10 * time.Minute,
				HTTPMinDelay:     5 * time.Second,
				HTTPParseJSON:    true,
				HTTPParseHeaders: true,
				HTTPTrustClient:  false,
				Command:          []string{"curl", "api.com"},
			},
		},
		{
			name: "http-aware with parsing options",
			args: []string{"interval", "--every", "30s", "--http-aware", "--http-no-parse-json", "--http-trust-client", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:       "interval",
				Every:            30 * time.Second,
				HTTPAware:        true,
				HTTPParseJSON:    false, // Disabled
				HTTPParseHeaders: true,  // Default
				HTTPTrustClient:  true,  // Enabled
				Command:          []string{"curl", "api.com"},
			},
		},
		{
			name: "http-aware with custom fields",
			args: []string{"interval", "--every", "30s", "--http-aware", "--http-custom-fields", "custom_retry,backoff_seconds", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:       "interval",
				Every:            30 * time.Second,
				HTTPAware:        true,
				HTTPParseJSON:    true,
				HTTPParseHeaders: true,
				HTTPTrustClient:  false,
				HTTPCustomFields: []string{"custom_retry", "backoff_seconds"},
				Command:          []string{"curl", "api.com"},
			},
		},
		{
			name: "adaptive with http-aware",
			args: []string{"adaptive", "--base-interval", "1s", "--http-aware", "--http-max-delay", "5m", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:       "adaptive",
				BaseInterval:     1 * time.Second,
				HTTPAware:        true,
				HTTPMaxDelay:     5 * time.Minute,
				HTTPParseJSON:    true,
				HTTPParseHeaders: true,
				HTTPTrustClient:  false,
				Command:          []string{"curl", "api.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)
			require.NoError(t, err)

			assert.Equal(t, tt.expected.Subcommand, config.Subcommand)
			assert.Equal(t, tt.expected.Every, config.Every)
			assert.Equal(t, tt.expected.BaseInterval, config.BaseInterval)
			assert.Equal(t, tt.expected.HTTPAware, config.HTTPAware)
			assert.Equal(t, tt.expected.HTTPMaxDelay, config.HTTPMaxDelay)
			assert.Equal(t, tt.expected.HTTPMinDelay, config.HTTPMinDelay)
			assert.Equal(t, tt.expected.HTTPParseJSON, config.HTTPParseJSON)
			assert.Equal(t, tt.expected.HTTPParseHeaders, config.HTTPParseHeaders)
			assert.Equal(t, tt.expected.HTTPTrustClient, config.HTTPTrustClient)
			assert.Equal(t, tt.expected.HTTPCustomFields, config.HTTPCustomFields)
			assert.Equal(t, tt.expected.Command, config.Command)
		})
	}
}

func TestConfig_GetHTTPAwareConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected *httpaware.HTTPAwareConfig
	}{
		{
			name: "http-aware disabled",
			config: Config{
				HTTPAware: false,
			},
			expected: nil,
		},
		{
			name: "http-aware with defaults",
			config: Config{
				HTTPAware:        true,
				HTTPParseJSON:    true,
				HTTPParseHeaders: true,
				HTTPTrustClient:  false,
			},
			expected: &httpaware.HTTPAwareConfig{
				MaxDelay:          30 * time.Minute, // Default
				MinDelay:          1 * time.Second,  // Default
				ParseJSON:         true,
				ParseHeaders:      true,
				TrustClientErrors: false,
				JSONFields:        []string{"retry_after", "retryAfter"}, // Default
			},
		},
		{
			name: "http-aware with custom values",
			config: Config{
				HTTPAware:        true,
				HTTPMaxDelay:     10 * time.Minute,
				HTTPMinDelay:     5 * time.Second,
				HTTPParseJSON:    false,
				HTTPParseHeaders: true,
				HTTPTrustClient:  true,
				HTTPCustomFields: []string{"custom_retry", "backoff_delay"},
			},
			expected: &httpaware.HTTPAwareConfig{
				MaxDelay:          10 * time.Minute,
				MinDelay:          5 * time.Second,
				ParseJSON:         false,
				ParseHeaders:      true,
				TrustClientErrors: true,
				JSONFields:        []string{"custom_retry", "backoff_delay"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetHTTPAwareConfig()

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.MaxDelay, result.MaxDelay)
				assert.Equal(t, tt.expected.MinDelay, result.MinDelay)
				assert.Equal(t, tt.expected.ParseJSON, result.ParseJSON)
				assert.Equal(t, tt.expected.ParseHeaders, result.ParseHeaders)
				assert.Equal(t, tt.expected.TrustClientErrors, result.TrustClientErrors)
				assert.Equal(t, tt.expected.JSONFields, result.JSONFields)
			}
		})
	}
}

func TestCLI_HTTPAwareFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid http-aware flags",
			args:        []string{"interval", "--every", "30s", "--http-aware", "--http-max-delay", "10m", "--", "curl", "api.com"},
			expectError: false,
		},
		{
			name:        "invalid max delay format",
			args:        []string{"interval", "--every", "30s", "--http-aware", "--http-max-delay", "invalid", "--", "curl", "api.com"},
			expectError: true,
			errorMsg:    "invalid duration",
		},
		{
			name:        "invalid min delay format",
			args:        []string{"interval", "--every", "30s", "--http-aware", "--http-min-delay", "invalid", "--", "curl", "api.com"},
			expectError: true,
			errorMsg:    "invalid duration",
		},
		{
			name:        "empty custom fields should work",
			args:        []string{"interval", "--every", "30s", "--http-aware", "--http-custom-fields", "", "--", "curl", "api.com"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArgs(tt.args)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
