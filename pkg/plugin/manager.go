package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

// PluginManager manages plugin discovery, loading, and lifecycle
type PluginManager struct {
	directories   []string
	registry      *PluginRegistry
	loadedPlugins map[string]*PluginManifest
	mu            sync.RWMutex
}

// PluginManifest represents a plugin's manifest file
type PluginManifest struct {
	Plugin       PluginInfo    `toml:"plugin"`
	Runtime      RuntimeConfig `toml:"runtime"`
	Config       ConfigInfo    `toml:"config"`
	Dependencies Dependencies  `toml:"dependencies"`
	Permissions  Permissions   `toml:"permissions"`
	ManifestPath string        `toml:"-"` // Not in TOML, set during discovery
}

// PluginInfo represents the main plugin information
type PluginInfo struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Type        string `toml:"type"`
	Description string `toml:"description"`
	Author      string `toml:"author"`
}

// RuntimeConfig represents plugin runtime configuration
type RuntimeConfig struct {
	Type       string `toml:"type"`        // "go-plugin", "wasm", "external"
	Binary     string `toml:"binary"`      // Binary file name
	EntryPoint string `toml:"entry_point"` // Entry point symbol/function
}

// ConfigInfo represents plugin configuration requirements
type ConfigInfo struct {
	Required []string `toml:"required"`
	Optional []string `toml:"optional"`
}

// Dependencies represents plugin dependencies
type Dependencies struct {
	MinRepeaterVersion string   `toml:"min_repeater_version"`
	ExternalDeps       []string `toml:"external_deps"`
}

// Permissions represents plugin security permissions
type Permissions struct {
	Network     bool     `toml:"network"`
	Filesystem  string   `toml:"filesystem"` // "read-only", "read-write", "none"
	SystemCalls []string `toml:"system_calls"`
}

// NewPluginManager creates a new plugin manager with specified directories
func NewPluginManager(directories []string) *PluginManager {
	return &PluginManager{
		directories:   directories,
		registry:      NewPluginRegistry(),
		loadedPlugins: make(map[string]*PluginManifest),
	}
}

// GetPluginDirectories returns the configured plugin directories
func (pm *PluginManager) GetPluginDirectories() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy to prevent external modification
	dirs := make([]string, len(pm.directories))
	copy(dirs, pm.directories)
	return dirs
}

// DiscoverPlugins discovers all plugins in the configured directories
func (pm *PluginManager) DiscoverPlugins() ([]*PluginManifest, error) {
	var allPlugins []*PluginManifest

	for _, dir := range pm.directories {
		plugins, err := pm.discoverPluginsInDirectory(dir)
		if err != nil {
			// Log error but continue with other directories
			continue
		}
		allPlugins = append(allPlugins, plugins...)
	}

	return allPlugins, nil
}

// discoverPluginsInDirectory discovers plugins in a specific directory
func (pm *PluginManager) discoverPluginsInDirectory(dir string) ([]*PluginManifest, error) {
	var plugins []*PluginManifest

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return plugins, nil // Empty result for non-existent directory
	}

	// Walk through directory looking for plugin.toml files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		if info.Name() == "plugin.toml" {
			manifest, err := pm.loadManifest(path)
			if err != nil {
				// Skip invalid manifests, continue discovery
				return nil
			}
			plugins = append(plugins, manifest)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	return plugins, nil
}

// loadManifest loads a plugin manifest from a TOML file
func (pm *PluginManager) loadManifest(manifestPath string) (*PluginManifest, error) {
	var manifest PluginManifest

	_, err := toml.DecodeFile(manifestPath, &manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode manifest %s: %w", manifestPath, err)
	}

	manifest.ManifestPath = manifestPath
	return &manifest, nil
}

// LoadPlugin loads a single plugin from its manifest
func (pm *PluginManager) LoadPlugin(manifest *PluginManifest) error {
	// Validate manifest first
	validator := NewPluginValidator()
	if err := validator.ValidateManifest(manifest); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if binary exists
	binaryPath := pm.getBinaryPath(manifest)
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("failed to load plugin: binary %s does not exist", binaryPath)
	}

	// For now, just mark as loaded (actual loading will be implemented later)
	pm.mu.Lock()
	pm.loadedPlugins[manifest.Plugin.Name] = manifest
	pm.mu.Unlock()

	return nil
}

// getBinaryPath constructs the full path to the plugin binary
func (pm *PluginManager) getBinaryPath(manifest *PluginManifest) string {
	manifestDir := filepath.Dir(manifest.ManifestPath)
	return filepath.Join(manifestDir, manifest.Runtime.Binary)
}

// LoadAllPlugins discovers and loads all plugins
func (pm *PluginManager) LoadAllPlugins() error {
	plugins, err := pm.DiscoverPlugins()
	if err != nil {
		return fmt.Errorf("failed to discover plugins: %w", err)
	}

	for _, plugin := range plugins {
		if err := pm.LoadPlugin(plugin); err != nil {
			// Log error but continue loading other plugins
			continue
		}
	}

	return nil
}

// GetLoadedSchedulerPlugins returns all loaded scheduler plugins
func (pm *PluginManager) GetLoadedSchedulerPlugins() map[string]*PluginManifest {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	schedulerPlugins := make(map[string]*PluginManifest)
	for name, plugin := range pm.loadedPlugins {
		if plugin.Plugin.Type == "scheduler" {
			schedulerPlugins[name] = plugin
		}
	}

	return schedulerPlugins
}

// GetLoadedExecutorPlugins returns all loaded executor plugins
func (pm *PluginManager) GetLoadedExecutorPlugins() map[string]*PluginManifest {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	executorPlugins := make(map[string]*PluginManifest)
	for name, plugin := range pm.loadedPlugins {
		if plugin.Plugin.Type == "executor" {
			executorPlugins[name] = plugin
		}
	}

	return executorPlugins
}

// GetLoadedOutputPlugins returns all loaded output plugins
func (pm *PluginManager) GetLoadedOutputPlugins() map[string]*PluginManifest {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	outputPlugins := make(map[string]*PluginManifest)
	for name, plugin := range pm.loadedPlugins {
		if plugin.Plugin.Type == "output" {
			outputPlugins[name] = plugin
		}
	}

	return outputPlugins
}

// PluginValidator validates plugin manifests
type PluginValidator struct{}

// NewPluginValidator creates a new plugin validator
func NewPluginValidator() *PluginValidator {
	return &PluginValidator{}
}

// ValidateManifest validates a plugin manifest
func (pv *PluginValidator) ValidateManifest(manifest *PluginManifest) error {
	if manifest.Plugin.Name == "" {
		return fmt.Errorf("plugin name is required")
	}

	if manifest.Plugin.Version == "" {
		return fmt.Errorf("plugin version is required")
	}

	if manifest.Plugin.Type == "" {
		return fmt.Errorf("plugin type is required")
	}

	// Validate plugin type
	validTypes := []string{"scheduler", "executor", "output"}
	validType := false
	for _, t := range validTypes {
		if manifest.Plugin.Type == t {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("invalid plugin type: %s (must be one of: %s)",
			manifest.Plugin.Type, strings.Join(validTypes, ", "))
	}
	// Validate runtime configuration
	if manifest.Runtime.Type == "" {
		return fmt.Errorf("runtime type is required")
	}

	if manifest.Runtime.Binary == "" {
		return fmt.Errorf("runtime binary is required")
	}

	return nil
}
