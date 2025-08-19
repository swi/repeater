package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/swi/repeater/pkg/interfaces"
)

// Mock plugins for testing
type mockSchedulerPlugin struct {
	name string
}

func (m *mockSchedulerPlugin) Name() string                               { return m.name }
func (m *mockSchedulerPlugin) Version() string                            { return "1.0.0" }
func (m *mockSchedulerPlugin) Description() string                        { return "Mock scheduler" }
func (m *mockSchedulerPlugin) ValidateConfig(config map[string]any) error { return nil }
func (m *mockSchedulerPlugin) ConfigSchema() *ConfigSchema                { return &ConfigSchema{} }
func (m *mockSchedulerPlugin) Create(config map[string]any) (interfaces.Scheduler, error) {
	return &mockScheduler{}, nil
}

type mockScheduler struct{}

func (m *mockScheduler) Next() <-chan time.Time {
	ch := make(chan time.Time, 1)
	ch <- time.Now()
	return ch
}
func (m *mockScheduler) Stop() {}

type mockExecutorPlugin struct {
	name string
}

func (m *mockExecutorPlugin) Name() string                               { return m.name }
func (m *mockExecutorPlugin) ValidateConfig(config map[string]any) error { return nil }
func (m *mockExecutorPlugin) Execute(ctx context.Context, cmd []string, opts ExecutorOptions) (*ExecutionResult, error) {
	return &ExecutionResult{}, nil
}
func (m *mockExecutorPlugin) SupportsStreaming() bool      { return false }
func (m *mockExecutorPlugin) SupportedPlatforms() []string { return []string{"linux"} }

func TestPluginCoordinator_RegisterSchedulerPlugin(t *testing.T) {
	coordinator := NewPluginCoordinator()

	plugin := &mockSchedulerPlugin{name: "test-scheduler"}

	// Test successful registration
	err := coordinator.RegisterSchedulerPlugin(plugin)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test duplicate registration
	err = coordinator.RegisterSchedulerPlugin(plugin)
	if err == nil {
		t.Error("Expected error for duplicate registration, got nil")
	}
}

func TestPluginCoordinator_RegisterExecutorPlugin(t *testing.T) {
	coordinator := NewPluginCoordinator()

	plugin := &mockExecutorPlugin{name: "test-executor"}

	// Test successful registration
	err := coordinator.RegisterExecutorPlugin(plugin)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test duplicate registration
	err = coordinator.RegisterExecutorPlugin(plugin)
	if err == nil {
		t.Error("Expected error for duplicate registration, got nil")
	}
}

func TestPluginCoordinator_CreateScheduler(t *testing.T) {
	coordinator := NewPluginCoordinator()
	plugin := &mockSchedulerPlugin{name: "test-scheduler"}

	// Register plugin
	err := coordinator.RegisterSchedulerPlugin(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Test successful creation
	ctx := context.Background()
	scheduler, err := coordinator.CreateScheduler(ctx, "test-scheduler", map[string]any{})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if scheduler == nil {
		t.Error("Expected scheduler, got nil")
	}

	// Test creation with non-existent plugin
	_, err = coordinator.CreateScheduler(ctx, "non-existent", map[string]any{})
	if err == nil {
		t.Error("Expected error for non-existent plugin, got nil")
	}
}

func TestPluginCoordinator_GetPlugins(t *testing.T) {
	coordinator := NewPluginCoordinator()

	// Register plugins
	schedulerPlugin := &mockSchedulerPlugin{name: "scheduler1"}
	executorPlugin := &mockExecutorPlugin{name: "executor1"}

	_ = coordinator.RegisterSchedulerPlugin(schedulerPlugin)
	_ = coordinator.RegisterExecutorPlugin(executorPlugin)

	// Test getting scheduler plugins
	schedulers := coordinator.GetSchedulerPlugins()
	if len(schedulers) != 1 {
		t.Errorf("Expected 1 scheduler plugin, got %d", len(schedulers))
	}
	if schedulers["scheduler1"] == nil {
		t.Error("Expected scheduler1 to be registered")
	}

	// Test getting executor plugins
	executors := coordinator.GetExecutorPlugins()
	if len(executors) != 1 {
		t.Errorf("Expected 1 executor plugin, got %d", len(executors))
	}
	if executors["executor1"] == nil {
		t.Error("Expected executor1 to be registered")
	}
}
