package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"whatspire/internal/infrastructure/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestConfigIntegration_WithFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
server:
  host: "127.0.0.1"
  port: 9999

log:
  level: "debug"

database:
  driver: "sqlite"
  dsn: "/tmp/integration-test.db"
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Set config file
	config.SetConfigFile(configFile)
	defer config.SetConfigFile("") // Reset after test

	// Create FX app with config module
	var loadedConfig *config.Config
	var watcher *config.ConfigWatcher

	app := fx.New(
		config.Module,
		fx.Populate(&loadedConfig, &watcher),
		fx.NopLogger,
	)

	// Start app
	startCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = app.Start(startCtx)
	require.NoError(t, err)

	// Verify config was loaded from file
	require.NotNil(t, loadedConfig)
	assert.Equal(t, "127.0.0.1", loadedConfig.Server.Host)
	assert.Equal(t, 9999, loadedConfig.Server.Port)
	assert.Equal(t, "debug", loadedConfig.Log.Level)
	assert.Equal(t, "sqlite", loadedConfig.Database.Driver)
	assert.Equal(t, "/tmp/integration-test.db", loadedConfig.Database.DSN)

	// Verify watcher was created
	require.NotNil(t, watcher)
	assert.True(t, watcher.IsRunning())

	// Stop app
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	err = app.Stop(stopCtx)
	require.NoError(t, err)

	// Verify watcher was stopped
	assert.False(t, watcher.IsRunning())
}

func TestConfigIntegration_WithoutFile(t *testing.T) {
	// Reset config file
	config.SetConfigFile("")

	// Create FX app with config module
	var loadedConfig *config.Config
	var watcher *config.ConfigWatcher

	app := fx.New(
		config.Module,
		fx.Populate(&loadedConfig, &watcher),
		fx.NopLogger,
	)

	// Start app
	startCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(startCtx)
	require.NoError(t, err)

	// Verify config was loaded with defaults
	require.NotNil(t, loadedConfig)
	assert.Equal(t, "0.0.0.0", loadedConfig.Server.Host)
	assert.Equal(t, 8080, loadedConfig.Server.Port)
	assert.Equal(t, "info", loadedConfig.Log.Level)

	// Verify watcher was created but not started (no config file)
	require.NotNil(t, watcher)
	assert.False(t, watcher.IsRunning())

	// Stop app
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	err = app.Stop(stopCtx)
	require.NoError(t, err)
}

func TestConfigIntegration_EnvironmentOverride(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
server:
  host: "127.0.0.1"
  port: 8888

log:
  level: "info"
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Set environment variables
	os.Setenv("WHATSAPP_SERVER_PORT", "7777")
	os.Setenv("WHATSAPP_LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("WHATSAPP_SERVER_PORT")
		os.Unsetenv("WHATSAPP_LOG_LEVEL")
	}()

	// Set config file
	config.SetConfigFile(configFile)
	defer config.SetConfigFile("") // Reset after test

	// Create FX app with config module
	var loadedConfig *config.Config

	app := fx.New(
		config.Module,
		fx.Populate(&loadedConfig),
		fx.NopLogger,
	)

	// Start app
	startCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = app.Start(startCtx)
	require.NoError(t, err)

	// Verify environment variables override file config
	require.NotNil(t, loadedConfig)
	assert.Equal(t, "127.0.0.1", loadedConfig.Server.Host) // From file
	assert.Equal(t, 7777, loadedConfig.Server.Port)        // From env (overrides file)
	assert.Equal(t, "debug", loadedConfig.Log.Level)       // From env (overrides file)

	// Stop app
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	err = app.Stop(stopCtx)
	require.NoError(t, err)
}
