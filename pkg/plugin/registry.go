package plugin

import (
	"fmt"
	"sync"
)

// PluginRegistry manages registered plugins
type PluginRegistry struct {
	schedulerPlugins map[string]SchedulerPlugin
	executorPlugins  map[string]ExecutorPlugin
	outputPlugins    map[string]OutputPlugin
	mu               sync.RWMutex
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		schedulerPlugins: make(map[string]SchedulerPlugin),
		executorPlugins:  make(map[string]ExecutorPlugin),
		outputPlugins:    make(map[string]OutputPlugin),
	}
}

// RegisterSchedulerPlugin registers a scheduler plugin
func (r *PluginRegistry) RegisterSchedulerPlugin(name string, plugin interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	schedulerPlugin, ok := plugin.(SchedulerPlugin)
	if !ok {
		return fmt.Errorf("plugin does not implement SchedulerPlugin interface")
	}

	r.schedulerPlugins[name] = schedulerPlugin
	return nil
}

// GetSchedulerPlugin retrieves a scheduler plugin by name
func (r *PluginRegistry) GetSchedulerPlugin(name string) (SchedulerPlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.schedulerPlugins[name]
	return plugin, exists
}

// RegisterExecutorPlugin registers an executor plugin
func (r *PluginRegistry) RegisterExecutorPlugin(name string, plugin interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	executorPlugin, ok := plugin.(ExecutorPlugin)
	if !ok {
		return fmt.Errorf("plugin does not implement ExecutorPlugin interface")
	}

	r.executorPlugins[name] = executorPlugin
	return nil
}

// GetExecutorPlugin retrieves an executor plugin by name
func (r *PluginRegistry) GetExecutorPlugin(name string) (ExecutorPlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.executorPlugins[name]
	return plugin, exists
}

// RegisterOutputPlugin registers an output plugin
func (r *PluginRegistry) RegisterOutputPlugin(name string, plugin interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	outputPlugin, ok := plugin.(OutputPlugin)
	if !ok {
		return fmt.Errorf("plugin does not implement OutputPlugin interface")
	}

	r.outputPlugins[name] = outputPlugin
	return nil
}

// GetOutputPlugin retrieves an output plugin by name
func (r *PluginRegistry) GetOutputPlugin(name string) (OutputPlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.outputPlugins[name]
	return plugin, exists
}
