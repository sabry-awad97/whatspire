package unit

import (
	"os"
	"path/filepath"
	"testing"

	"whatspire/internal/infrastructure/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoad_FromYAML(t *testing.T) {
	// Create temporary YAML config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
server:
  host: "127.0.0.1"
  port: 9090

log:
  level: "debug"
  format: "text"

ratelimit:
  enabled: false
  requests_per_second: 20.0

database:
  driver: "sqlite"
  dsn: "/tmp/test.db"

events:
  enabled: true
  retention_days: 7
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Load config from file
	cfg, err := config.LoadWithConfigFile(configFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify values from YAML file
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "text", cfg.Log.Format)
	assert.False(t, cfg.RateLimit.Enabled)
	assert.Equal(t, 20.0, cfg.RateLimit.RequestsPerSecond)
	assert.Equal(t, "sqlite", cfg.Database.Driver)
	assert.Equal(t, "/tmp/test.db", cfg.Database.DSN)
	assert.True(t, cfg.Events.Enabled)
	assert.Equal(t, 7, cfg.Events.RetentionDays)
}

func TestConfigLoad_FromJSON(t *testing.T) {
	// Create temporary JSON config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	jsonContent := `{
  "server": {
    "host": "0.0.0.0",
    "port": 8888
  },
  "log": {
    "level": "warn",
    "format": "json"
  },
  "ratelimit": {
    "enabled": true,
    "requests_per_second": 15.0,
    "burst_size": 30
  },
  "database": {
    "driver": "postgres",
    "dsn": "postgres://localhost/testdb"
  },
  "events": {
    "enabled": false,
    "retention_days": 90
  }
}`

	err := os.WriteFile(configFile, []byte(jsonContent), 0644)
	require.NoError(t, err)

	// Load config from file
	cfg, err := config.LoadWithConfigFile(configFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify values from JSON file
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8888, cfg.Server.Port)
	assert.Equal(t, "warn", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
	assert.True(t, cfg.RateLimit.Enabled)
	assert.Equal(t, 15.0, cfg.RateLimit.RequestsPerSecond)
	assert.Equal(t, 30, cfg.RateLimit.BurstSize)
	assert.Equal(t, "postgres", cfg.Database.Driver)
	assert.Equal(t, "postgres://localhost/testdb", cfg.Database.DSN)
	assert.False(t, cfg.Events.Enabled)
	assert.Equal(t, 90, cfg.Events.RetentionDays)
}

func TestConfigLoad_EnvironmentOverride(t *testing.T) {
	// Create temporary YAML config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
server:
  host: "127.0.0.1"
  port: 9090

log:
  level: "info"
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Set environment variables (should override file config)
	os.Setenv("WHATSAPP_SERVER_PORT", "7777")
	os.Setenv("WHATSAPP_LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("WHATSAPP_SERVER_PORT")
		os.Unsetenv("WHATSAPP_LOG_LEVEL")
	}()

	// Load config from file
	cfg, err := config.LoadWithConfigFile(configFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify environment variables override file config
	assert.Equal(t, "127.0.0.1", cfg.Server.Host) // Not overridden
	assert.Equal(t, 7777, cfg.Server.Port)        // Overridden by env var
	assert.Equal(t, "debug", cfg.Log.Level)       // Overridden by env var
}

func TestConfigLoad_InvalidFile(t *testing.T) {
	// Try to load non-existent file
	cfg, err := config.LoadWithConfigFile("/nonexistent/config.yaml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestConfigLoad_InvalidYAML(t *testing.T) {
	// Create temporary file with invalid YAML
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	invalidYAML := `
server:
  host: "127.0.0.1"
  port: invalid_port
  nested:
    - item1
    - item2
  invalid_indent
`

	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	// Try to load invalid config
	cfg, err := config.LoadWithConfigFile(configFile)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestConfigLoad_ValidationError(t *testing.T) {
	// Create temporary YAML with invalid values
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
server:
  host: "127.0.0.1"
  port: 99999  # Invalid port (> 65535)

log:
  level: "invalid_level"  # Invalid log level
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Try to load config with validation errors
	cfg, err := config.LoadWithConfigFile(configFile)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "validation")
}

func TestConfigLoad_NoFile(t *testing.T) {
	// Load config without file (should use defaults and env vars)
	cfg, err := config.LoadWithConfigFile("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify default values are used
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
}

func TestConfigLoad_Precedence(t *testing.T) {
	// Test config precedence: env > file > defaults
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
server:
  host: "file-host"
  port: 5000

log:
  level: "warn"
  format: "text"
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Set some environment variables
	os.Setenv("WHATSAPP_SERVER_PORT", "6000")
	os.Setenv("WHATSAPP_LOG_FORMAT", "json")
	defer func() {
		os.Unsetenv("WHATSAPP_SERVER_PORT")
		os.Unsetenv("WHATSAPP_LOG_FORMAT")
	}()

	// Load config
	cfg, err := config.LoadWithConfigFile(configFile)
	require.NoError(t, err)

	// Verify precedence: env > file > defaults
	assert.Equal(t, "file-host", cfg.Server.Host) // From file (no env override)
	assert.Equal(t, 6000, cfg.Server.Port)        // From env (overrides file)
	assert.Equal(t, "warn", cfg.Log.Level)        // From file (no env override)
	assert.Equal(t, "json", cfg.Log.Format)       // From env (overrides file)
}
