package plugin

import (
	"context"
	"fmt"
	"maps"

	"github.com/swi/repeater/pkg/interfaces"
)

// PluginCoordinator manages plugin lifecycle and coordination
type PluginCoordinator struct {
	schedulerPlugins map[string]SchedulerPlugin
	executorPlugins  map[string]ExecutorPlugin
	outputPlugins    map[string]OutputPlugin
}

// NewPluginCoordinator creates a new plugin coordinator
func NewPluginCoordinator() *PluginCoordinator {
	return &PluginCoordinator{
		schedulerPlugins: make(map[string]SchedulerPlugin),
		executorPlugins:  make(map[string]ExecutorPlugin),
		outputPlugins:    make(map[string]OutputPlugin),
	}
}

// RegisterSchedulerPlugin registers a scheduler plugin
func (pc *PluginCoordinator) RegisterSchedulerPlugin(plugin SchedulerPlugin) error {
	name := plugin.Name()
	if _, exists := pc.schedulerPlugins[name]; exists {
		return fmt.Errorf("scheduler plugin %q already registered", name)
	}
	pc.schedulerPlugins[name] = plugin
	return nil
}

// RegisterExecutorPlugin registers an executor plugin
func (pc *PluginCoordinator) RegisterExecutorPlugin(plugin ExecutorPlugin) error {
	name := plugin.Name()
	if _, exists := pc.executorPlugins[name]; exists {
		return fmt.Errorf("executor plugin %q already registered", name)
	}
	pc.executorPlugins[name] = plugin
	return nil
}

// RegisterOutputPlugin registers an output plugin
func (pc *PluginCoordinator) RegisterOutputPlugin(plugin OutputPlugin) error {
	name := plugin.Name()
	if _, exists := pc.outputPlugins[name]; exists {
		return fmt.Errorf("output plugin %q already registered", name)
	}
	pc.outputPlugins[name] = plugin
	return nil
}

// CreateScheduler creates a scheduler from a plugin
func (pc *PluginCoordinator) CreateScheduler(ctx context.Context, pluginName string, config map[string]any) (interfaces.Scheduler, error) {
	plugin, exists := pc.schedulerPlugins[pluginName]
	if !exists {
		return nil, fmt.Errorf("scheduler plugin %q not found", pluginName)
	}

	if err := plugin.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config for scheduler plugin %q: %w", pluginName, err)
	}

	return plugin.Create(config)
}

// GetSchedulerPlugins returns all registered scheduler plugins
func (pc *PluginCoordinator) GetSchedulerPlugins() map[string]SchedulerPlugin {
	result := make(map[string]SchedulerPlugin, len(pc.schedulerPlugins))
	maps.Copy(result, pc.schedulerPlugins)
	return result
}

// GetExecutorPlugins returns all registered executor plugins
func (pc *PluginCoordinator) GetExecutorPlugins() map[string]ExecutorPlugin {
	result := make(map[string]ExecutorPlugin, len(pc.executorPlugins))
	maps.Copy(result, pc.executorPlugins)
	return result
}

// GetOutputPlugins returns all registered output plugins
func (pc *PluginCoordinator) GetOutputPlugins() map[string]OutputPlugin {
	result := make(map[string]OutputPlugin, len(pc.outputPlugins))
	maps.Copy(result, pc.outputPlugins)
	return result
}

// ValidatePluginConfig validates configuration for a specific plugin
func (pc *PluginCoordinator) ValidatePluginConfig(pluginType, pluginName string, config map[string]any) error {
	switch pluginType {
	case "scheduler":
		if plugin, exists := pc.schedulerPlugins[pluginName]; exists {
			return plugin.ValidateConfig(config)
		}
		return fmt.Errorf("scheduler plugin %q not found", pluginName)
	case "executor":
		if plugin, exists := pc.executorPlugins[pluginName]; exists {
			return plugin.ValidateConfig(config)
		}
		return fmt.Errorf("executor plugin %q not found", pluginName)
	case "output":
		if plugin, exists := pc.outputPlugins[pluginName]; exists {
			return plugin.ValidateConfig(config)
		}
		return fmt.Errorf("output plugin %q not found", pluginName)
	default:
		return fmt.Errorf("unknown plugin type %q", pluginType)
	}
}
