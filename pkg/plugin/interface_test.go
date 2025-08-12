package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchedulerPluginInterface tests that the SchedulerPlugin interface exists and has required methods
func TestSchedulerPluginInterface(t *testing.T) {
	t.Run("interface methods exist", func(t *testing.T) {
		// Test with a concrete implementation
		plugin := &MockSchedulerPlugin{
			name:    "test-scheduler",
			version: "1.0.0",
		}

		// Test that interface methods exist and return expected types
		name := plugin.Name()
		assert.IsType(t, "", name, "Name() should return string")
		assert.Equal(t, "test-scheduler", name, "Name() should return correct name")

		version := plugin.Version()
		assert.IsType(t, "", version, "Version() should return string")
		assert.Equal(t, "1.0.0", version, "Version() should return correct version")

		// Test config validation method exists
		err := plugin.ValidateConfig(map[string]interface{}{})
		assert.IsType(t, error(nil), err, "ValidateConfig() should return error")

		// Test scheduler creation method exists
		scheduler, err := plugin.Create(map[string]interface{}{})
		assert.IsType(t, (*MockScheduler)(nil), scheduler, "Create() should return Scheduler interface")
		assert.IsType(t, error(nil), err, "Create() should return error")
	})
}

// TestExecutorPluginInterface tests that the ExecutorPlugin interface exists and has required methods
func TestExecutorPluginInterface(t *testing.T) {
	t.Run("interface methods exist", func(t *testing.T) {
		// Test with a concrete implementation
		plugin := &MockExecutorPlugin{
			name: "test-executor",
		}

		name := plugin.Name()
		assert.IsType(t, "", name, "Name() should return string")
		assert.Equal(t, "test-executor", name, "Name() should return correct name")

		// Test execution method exists
		ctx := context.Background()
		result, err := plugin.Execute(ctx, []string{"echo", "test"}, ExecutorOptions{})
		assert.IsType(t, (*ExecutionResult)(nil), result, "Execute() should return ExecutionResult")
		assert.IsType(t, error(nil), err, "Execute() should return error")

		// Test capability methods
		streaming := plugin.SupportsStreaming()
		assert.IsType(t, false, streaming, "SupportsStreaming() should return bool")

		platforms := plugin.SupportedPlatforms()
		assert.IsType(t, []string{}, platforms, "SupportedPlatforms() should return []string")
	})
}

// TestOutputPluginInterface tests that the OutputPlugin interface exists and has required methods
func TestOutputPluginInterface(t *testing.T) {
	t.Run("interface methods exist", func(t *testing.T) {
		// Test with a concrete implementation
		plugin := &MockOutputPlugin{
			name: "test-output",
		}

		name := plugin.Name()
		assert.IsType(t, "", name, "Name() should return string")
		assert.Equal(t, "test-output", name, "Name() should return correct name")

		// Test output processing method exists
		err := plugin.ProcessOutput(&ExecutionResult{}, OutputConfig{})
		assert.IsType(t, error(nil), err, "ProcessOutput() should return error")

		// Test capability methods
		streaming := plugin.SupportsStreaming()
		assert.IsType(t, false, streaming, "SupportsStreaming() should return bool")

		required := plugin.RequiredConfig()
		assert.IsType(t, []string{}, required, "RequiredConfig() should return []string")
	})
}

// TestPluginRegistration tests plugin registration fails with invalid interface
func TestPluginRegistration(t *testing.T) {
	t.Run("registration fails with invalid interface", func(t *testing.T) {
		// This test will fail until we implement plugin registration
		registry := NewPluginRegistry()

		// Try to register something that doesn't implement the interface
		err := registry.RegisterSchedulerPlugin("invalid", &InvalidPlugin{})
		require.Error(t, err, "Should fail to register invalid plugin")
		assert.Contains(t, err.Error(), "does not implement", "Error should mention interface implementation")
	})

	t.Run("registration succeeds with valid interface", func(t *testing.T) {
		registry := NewPluginRegistry()

		// Register a valid plugin
		validPlugin := &MockSchedulerPlugin{
			name:    "test-scheduler",
			version: "1.0.0",
		}

		err := registry.RegisterSchedulerPlugin("test-scheduler", validPlugin)
		require.NoError(t, err, "Should successfully register valid plugin")

		// Verify plugin was registered
		plugin, exists := registry.GetSchedulerPlugin("test-scheduler")
		assert.True(t, exists, "Plugin should exist in registry")
		assert.Equal(t, validPlugin, plugin, "Should return the same plugin instance")
	})
}

// TestPluginImplementsRequiredMethods tests that plugins implement all required methods
func TestPluginImplementsRequiredMethods(t *testing.T) {
	t.Run("scheduler plugin implements all methods", func(t *testing.T) {
		plugin := &MockSchedulerPlugin{
			name:    "test-scheduler",
			version: "1.0.0",
		}

		// Test all required methods are implemented
		assert.NotEmpty(t, plugin.Name(), "Name() should return non-empty string")
		assert.NotEmpty(t, plugin.Version(), "Version() should return non-empty string")

		// Test config validation
		err := plugin.ValidateConfig(map[string]interface{}{})
		assert.NoError(t, err, "ValidateConfig() should not error with empty config")

		// Test scheduler creation
		scheduler, err := plugin.Create(map[string]interface{}{})
		assert.NoError(t, err, "Create() should not error with empty config")
		assert.NotNil(t, scheduler, "Create() should return non-nil scheduler")
	})
}

// Mock types for testing (these will fail until we define the real interfaces)

type MockScheduler struct{}

func (m *MockScheduler) Next() <-chan time.Time {
	ch := make(chan time.Time, 1)
	ch <- time.Now()
	return ch
}

func (m *MockScheduler) Stop() {}

type MockSchedulerPlugin struct {
	name    string
	version string
}

func (m *MockSchedulerPlugin) Name() string {
	return m.name
}

func (m *MockSchedulerPlugin) Version() string {
	return m.version
}

func (m *MockSchedulerPlugin) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (m *MockSchedulerPlugin) Create(config map[string]interface{}) (Scheduler, error) {
	return &MockScheduler{}, nil
}

type MockExecutorPlugin struct {
	name string
}

func (m *MockExecutorPlugin) Name() string {
	return m.name
}

func (m *MockExecutorPlugin) Execute(ctx context.Context, cmd []string, opts ExecutorOptions) (*ExecutionResult, error) {
	return &ExecutionResult{
		ExitCode: 0,
		Output:   "mock output",
		Duration: time.Millisecond,
	}, nil
}

func (m *MockExecutorPlugin) SupportsStreaming() bool {
	return false
}

func (m *MockExecutorPlugin) SupportedPlatforms() []string {
	return []string{"linux", "darwin"}
}

type MockOutputPlugin struct {
	name string
}

func (m *MockOutputPlugin) Name() string {
	return m.name
}

func (m *MockOutputPlugin) ProcessOutput(result *ExecutionResult, config OutputConfig) error {
	return nil
}

func (m *MockOutputPlugin) SupportsStreaming() bool {
	return false
}

func (m *MockOutputPlugin) RequiredConfig() []string {
	return []string{}
}

type InvalidPlugin struct{}
