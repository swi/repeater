package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPluginManagerCreation tests that plugin manager can be created with directories
func TestPluginManagerCreation(t *testing.T) {
	t.Run("create manager with single directory", func(t *testing.T) {
		// This test will fail until we implement PluginManager
		manager := NewPluginManager([]string{"/tmp/plugins"})
		require.NotNil(t, manager, "Manager should not be nil")

		// Test that manager has the directory configured
		dirs := manager.GetPluginDirectories()
		assert.Equal(t, []string{"/tmp/plugins"}, dirs, "Should have configured directory")
	})

	t.Run("create manager with multiple directories", func(t *testing.T) {
		dirs := []string{"/tmp/plugins", "/usr/local/plugins", "~/.repeater/plugins"}
		manager := NewPluginManager(dirs)
		require.NotNil(t, manager, "Manager should not be nil")

		managerDirs := manager.GetPluginDirectories()
		assert.Equal(t, dirs, managerDirs, "Should have all configured directories")
	})
}

// TestPluginDiscovery tests that plugin manager can discover plugins in directories
func TestPluginDiscovery(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir := t.TempDir()
	pluginDir := filepath.Join(tempDir, "plugins")
	err := os.MkdirAll(pluginDir, 0755)
	require.NoError(t, err)

	t.Run("discover plugins in empty directory", func(t *testing.T) {
		// This test will fail until we implement plugin discovery
		manager := NewPluginManager([]string{pluginDir})

		plugins, err := manager.DiscoverPlugins()
		require.NoError(t, err, "Discovery should not error on empty directory")
		assert.Empty(t, plugins, "Should find no plugins in empty directory")
	})

	t.Run("discover plugins with manifest files", func(t *testing.T) {
		// Create mock plugin manifest
		schedulerDir := filepath.Join(pluginDir, "schedulers", "test-scheduler")
		err := os.MkdirAll(schedulerDir, 0755)
		require.NoError(t, err)

		manifestContent := `[plugin]
name = "test-scheduler"
version = "1.0.0"
type = "scheduler"
description = "Test scheduler plugin"

[runtime]
type = "go-plugin"
binary = "test-scheduler.so"
entry_point = "Plugin"
`
		manifestPath := filepath.Join(schedulerDir, "plugin.toml")
		err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)

		manager := NewPluginManager([]string{pluginDir})
		plugins, err := manager.DiscoverPlugins()
		require.NoError(t, err, "Discovery should not error with valid manifest")
		assert.Len(t, plugins, 1, "Should find one plugin")

		plugin := plugins[0]
		assert.Equal(t, "test-scheduler", plugin.Plugin.Name, "Should have correct plugin name")
		assert.Equal(t, "1.0.0", plugin.Plugin.Version, "Should have correct version")
		assert.Equal(t, "scheduler", plugin.Plugin.Type, "Should have correct type")
	})

	t.Run("skip invalid manifest files", func(t *testing.T) {
		// Create invalid manifest
		invalidDir := filepath.Join(pluginDir, "schedulers", "invalid-scheduler")
		err := os.MkdirAll(invalidDir, 0755)
		require.NoError(t, err)

		invalidManifest := `invalid toml content [[[`
		manifestPath := filepath.Join(invalidDir, "plugin.toml")
		err = os.WriteFile(manifestPath, []byte(invalidManifest), 0644)
		require.NoError(t, err)

		manager := NewPluginManager([]string{pluginDir})
		plugins, err := manager.DiscoverPlugins()
		require.NoError(t, err, "Discovery should not error with invalid manifest")
		assert.Len(t, plugins, 1, "Should skip invalid manifest and find valid one")
	})
}

// TestPluginLoading tests that plugin manager can load discovered plugins
func TestPluginLoading(t *testing.T) {
	t.Run("load plugins from discovery", func(t *testing.T) {
		// This test will fail until we implement plugin loading
		manager := NewPluginManager([]string{"/tmp/plugins"})

		// Mock discovered plugin
		discoveredPlugin := &PluginManifest{
			Plugin: PluginInfo{
				Name:        "test-scheduler",
				Version:     "1.0.0",
				Type:        "scheduler",
				Description: "Test scheduler",
			},
			Runtime: RuntimeConfig{
				Type:       "go-plugin",
				Binary:     "test-scheduler.so",
				EntryPoint: "Plugin",
			},
			ManifestPath: "/tmp/plugins/schedulers/test-scheduler/plugin.toml",
		}

		err := manager.LoadPlugin(discoveredPlugin)
		// Should fail because binary doesn't exist, but method should exist
		assert.Error(t, err, "Should error when binary doesn't exist")
		assert.Contains(t, err.Error(), "failed to load", "Error should mention loading failure")
	})

	t.Run("validate plugin before loading", func(t *testing.T) {
		manager := NewPluginManager([]string{"/tmp/plugins"})

		// Invalid plugin manifest (missing required fields)
		invalidPlugin := &PluginManifest{
			Plugin: PluginInfo{
				Name: "invalid-plugin",
				// Missing version, type, etc.
			},
		}

		err := manager.LoadPlugin(invalidPlugin)
		require.Error(t, err, "Should error with invalid plugin manifest")
		assert.Contains(t, err.Error(), "validation failed", "Error should mention validation failure")
	})
}

// TestPluginManagerIntegration tests full plugin manager workflow
func TestPluginManagerIntegration(t *testing.T) {
	t.Run("full discovery and loading workflow", func(t *testing.T) {
		// This test will fail until we implement the full workflow
		tempDir := t.TempDir()
		manager := NewPluginManager([]string{tempDir})

		// Test full workflow
		err := manager.LoadAllPlugins()
		require.NoError(t, err, "LoadAllPlugins should not error on empty directory")

		// Verify no plugins loaded
		schedulerPlugins := manager.GetLoadedSchedulerPlugins()
		assert.Empty(t, schedulerPlugins, "Should have no loaded scheduler plugins")

		executorPlugins := manager.GetLoadedExecutorPlugins()
		assert.Empty(t, executorPlugins, "Should have no loaded executor plugins")

		outputPlugins := manager.GetLoadedOutputPlugins()
		assert.Empty(t, outputPlugins, "Should have no loaded output plugins")
	})
}

// TestPluginValidation tests plugin manifest validation
func TestPluginValidation(t *testing.T) {
	t.Run("validate complete plugin manifest", func(t *testing.T) {
		// This test will fail until we implement validation
		validator := NewPluginValidator()

		validManifest := &PluginManifest{
			Plugin: PluginInfo{
				Name:        "test-scheduler",
				Version:     "1.0.0",
				Type:        "scheduler",
				Description: "Test scheduler plugin",
			},
			Runtime: RuntimeConfig{
				Type:       "go-plugin",
				Binary:     "test-scheduler.so",
				EntryPoint: "Plugin",
			},
		}

		err := validator.ValidateManifest(validManifest)
		assert.NoError(t, err, "Valid manifest should pass validation")
	})

	t.Run("reject invalid plugin manifest", func(t *testing.T) {
		validator := NewPluginValidator()

		invalidManifest := &PluginManifest{
			Plugin: PluginInfo{
				Name: "test-scheduler",
				// Missing required fields
			},
		}

		err := validator.ValidateManifest(invalidManifest)
		require.Error(t, err, "Invalid manifest should fail validation")
		assert.Contains(t, err.Error(), "version", "Should mention missing version")
	})
}
