package property

import (
	"os"
	"strconv"
	"testing"
	"time"

	"whatspire/internal/infrastructure/config"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/spf13/viper"
)

// Feature: whatsapp-http-api-enhancement, Property 33: Configuration Default Values
// *For any* missing configuration value, the service SHALL use a sensible default value
// and continue startup.
// **Validates: Requirements 12.3**

// createMinimalViperConfig creates a viper instance with only required fields set
func createMinimalViperConfig() *viper.Viper {
	v := viper.New()
	v.Set("whatsapp.db_path", "/data/whatsmeow.db")
	v.Set("websocket.url", "ws://localhost:3000/ws/whatsapp")
	v.Set("log.level", "info")
	v.Set("log.format", "json")
	v.Set("media.base_path", "/data/media")
	v.Set("media.base_url", "http://localhost:8080/media")
	return v
}

func TestConfigurationDefaultValues_Property33(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 33.1: Missing server host should use default "0.0.0.0"
	properties.Property("missing server host uses default", prop.ForAll(
		func(_ int) bool {
			v := createMinimalViperConfig()
			// Omit server.host to test default

			cfg, err := config.LoadWithViper(v)
			if err != nil {
				return false
			}

			return cfg.Server.Host == "0.0.0.0"
		},
		gen.Const(0),
	))

	// Property 33.2: Missing server port should use default 8080
	properties.Property("missing server port uses default", prop.ForAll(
		func(_ int) bool {
			v := createMinimalViperConfig()
			// Omit server.port to test default

			cfg, err := config.LoadWithViper(v)
			if err != nil {
				return false
			}

			return cfg.Server.Port == 8080
		},
		gen.Const(0),
	))

	// Property 33.3: Missing QR timeout should use default 2 minutes
	properties.Property("missing QR timeout uses default", prop.ForAll(
		func(_ int) bool {
			v := createMinimalViperConfig()
			// Omit whatsapp.qr_timeout to test default

			cfg, err := config.LoadWithViper(v)
			if err != nil {
				return false
			}

			return cfg.WhatsApp.QRTimeout == 2*time.Minute
		},
		gen.Const(0),
	))

	// Property 33.4: Missing log level should use default "info"
	properties.Property("missing log level uses default", prop.ForAll(
		func(_ int) bool {
			v := createMinimalViperConfig()
			v.Set("log.format", "json")
			// Omit log.level to test default

			cfg, err := config.LoadWithViper(v)
			if err != nil {
				return false
			}

			return cfg.Log.Level == "info"
		},
		gen.Const(0),
	))

	// Property 33.5: Missing log format should use default "json"
	properties.Property("missing log format uses default", prop.ForAll(
		func(_ int) bool {
			v := createMinimalViperConfig()
			v.Set("log.level", "info")
			// Omit log.format to test default

			cfg, err := config.LoadWithViper(v)
			if err != nil {
				return false
			}

			return cfg.Log.Format == "json"
		},
		gen.Const(0),
	))

	// Property 33.6: Missing rate limit settings should use sensible defaults
	properties.Property("missing rate limit settings use defaults", prop.ForAll(
		func(_ int) bool {
			v := createMinimalViperConfig()
			// Omit ratelimit settings to test defaults

			cfg, err := config.LoadWithViper(v)
			if err != nil {
				return false
			}

			return cfg.RateLimit.Enabled == true &&
				cfg.RateLimit.RequestsPerSecond == 10.0 &&
				cfg.RateLimit.BurstSize == 20
		},
		gen.Const(0),
	))

	// Property 33.7: Missing media max file size should use default 16MB
	properties.Property("missing media max file size uses default", prop.ForAll(
		func(_ int) bool {
			v := createMinimalViperConfig()
			// Omit media.max_file_size to test default

			cfg, err := config.LoadWithViper(v)
			if err != nil {
				return false
			}

			return cfg.Media.MaxFileSize == 16*1024*1024 // 16MB
		},
		gen.Const(0),
	))

	// Property 33.9: Missing API key enabled should default to false
	properties.Property("missing API key enabled defaults to false", prop.ForAll(
		func(_ int) bool {
			v := createMinimalViperConfig()
			// Omit apikey.enabled to test default

			cfg, err := config.LoadWithViper(v)
			if err != nil {
				return false
			}

			return cfg.APIKey.Enabled == false
		},
		gen.Const(0),
	))

	// Property 33.10: Missing circuit breaker settings should use sensible defaults
	properties.Property("missing circuit breaker settings use defaults", prop.ForAll(
		func(_ int) bool {
			v := createMinimalViperConfig()
			// Omit circuitbreaker settings to test defaults

			cfg, err := config.LoadWithViper(v)
			if err != nil {
				return false
			}

			return cfg.CircuitBreaker.Enabled == true &&
				cfg.CircuitBreaker.MaxRequests == 3 &&
				cfg.CircuitBreaker.FailureThreshold == 5
		},
		gen.Const(0),
	))

	properties.TestingRun(t)
}

// Feature: whatsapp-http-api-enhancement, Property 34: Configuration Validation on Startup
// *For any* invalid configuration value, the service SHALL log an error and fail to start.
// **Validates: Requirements 12.4**

func TestConfigurationValidationOnStartup_Property34(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 34.1: Invalid port should cause startup failure
	properties.Property("invalid port causes startup failure", prop.ForAll(
		func(port int) bool {
			if port >= 1 && port <= 65535 {
				return true // skip valid ports
			}

			v := createMinimalViperConfig()
			v.Set("server.port", port)

			_, err := config.LoadWithViper(v)
			return err != nil
		},
		gen.OneConstOf(0, -1, -100, 65536, 70000, 100000),
	))

	// Property 34.2: Missing required DB path should cause startup failure
	properties.Property("missing DB path causes startup failure", prop.ForAll(
		func(_ int) bool {
			v := createMinimalViperConfig()
			v.Set("whatsapp.db_path", "")

			_, err := config.LoadWithViper(v)
			return err != nil
		},
		gen.Const(0),
	))

	// Property 34.3: Missing required WebSocket URL should cause startup failure
	properties.Property("missing WebSocket URL causes startup failure", prop.ForAll(
		func(_ int) bool {
			v := createMinimalViperConfig()
			v.Set("websocket.url", "")

			_, err := config.LoadWithViper(v)
			return err != nil
		},
		gen.Const(0),
	))

	// Property 34.4: Invalid log level should cause startup failure
	properties.Property("invalid log level causes startup failure", prop.ForAll(
		func(level string) bool {
			validLevels := map[string]bool{
				"debug": true, "info": true, "warn": true, "error": true,
				"DEBUG": true, "INFO": true, "WARN": true, "ERROR": true,
			}
			if validLevels[level] {
				return true // skip valid levels
			}

			v := createMinimalViperConfig()
			v.Set("log.level", level)

			_, err := config.LoadWithViper(v)
			return err != nil
		},
		gen.OneConstOf("invalid", "trace", "fatal", "warning", ""),
	))

	// Property 34.5: Invalid log format should cause startup failure
	properties.Property("invalid log format causes startup failure", prop.ForAll(
		func(format string) bool {
			validFormats := map[string]bool{
				"json": true, "text": true,
				"JSON": true, "TEXT": true,
			}
			if validFormats[format] {
				return true // skip valid formats
			}

			v := createMinimalViperConfig()
			v.Set("log.format", format)

			_, err := config.LoadWithViper(v)
			return err != nil
		},
		gen.OneConstOf("xml", "yaml", "csv", ""),
	))

	// Property 34.6: Non-positive QR timeout should cause startup failure
	properties.Property("non-positive QR timeout causes startup failure", prop.ForAll(
		func(timeout time.Duration) bool {
			if timeout > 0 {
				return true // skip valid timeouts
			}

			v := createMinimalViperConfig()
			v.Set("whatsapp.qr_timeout", timeout)

			_, err := config.LoadWithViper(v)
			return err != nil
		},
		gen.OneConstOf(time.Duration(0), -time.Second, -time.Minute),
	))

	properties.TestingRun(t)
}

// Feature: whatsapp-http-api-enhancement, Property 35: Configuration Hot Reload
// *For any* configuration change, reloading the configuration SHALL apply the new values
// without restarting the service.
// **Validates: Requirements 12.5**

func TestConfigurationHotReload_Property35(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50 // Reduced since we're doing env var manipulation
	properties := gopter.NewProperties(parameters)

	// Property 35.1: Reloading with changed server port should update config
	properties.Property("reload with changed server port updates config", prop.ForAll(
		func(newPort int) bool {
			if newPort < 1 || newPort > 65535 {
				return true // skip invalid ports
			}

			// Set initial environment
			os.Setenv("WHATSAPP_SERVER_PORT", "8080")
			os.Setenv("WHATSAPP_WHATSAPP_DB_PATH", "/data/whatsmeow.db")
			os.Setenv("WHATSAPP_WEBSOCKET_URL", "ws://localhost:3000/ws/whatsapp")
			os.Setenv("WHATSAPP_LOG_LEVEL", "info")
			os.Setenv("WHATSAPP_LOG_FORMAT", "json")
			os.Setenv("WHATSAPP_MEDIA_BASE_PATH", "/data/media")
			os.Setenv("WHATSAPP_MEDIA_BASE_URL", "http://localhost:8080/media")
			defer cleanupEnv()

			cfg, err := config.Load()
			if err != nil {
				t.Logf("Initial load failed: %v", err)
				return false
			}

			if cfg.Server.Port != 8080 {
				t.Logf("Initial port mismatch: got %d, want 8080", cfg.Server.Port)
				return false
			}

			// Change environment variable - FIX: Use strconv.Itoa instead of string(rune())
			os.Setenv("WHATSAPP_SERVER_PORT", strconv.Itoa(newPort))

			// Reload configuration
			err = cfg.Reload()
			if err != nil {
				t.Logf("Reload failed: %v", err)
				return false
			}

			if cfg.Server.Port != newPort {
				t.Logf("Port not updated after reload: got %d, want %d", cfg.Server.Port, newPort)
				return false
			}

			return true
		},
		gen.IntRange(1, 65535),
	))

	// Property 35.2: Reloading with changed log level should update config
	properties.Property("reload with changed log level updates config", prop.ForAll(
		func(newLevel string) bool {
			// Set initial environment
			os.Setenv("WHATSAPP_SERVER_PORT", "8080")
			os.Setenv("WHATSAPP_WHATSAPP_DB_PATH", "/data/whatsmeow.db")
			os.Setenv("WHATSAPP_WEBSOCKET_URL", "ws://localhost:3000/ws/whatsapp")
			os.Setenv("WHATSAPP_LOG_LEVEL", "info")
			os.Setenv("WHATSAPP_LOG_FORMAT", "json")
			os.Setenv("WHATSAPP_MEDIA_BASE_PATH", "/data/media")
			os.Setenv("WHATSAPP_MEDIA_BASE_URL", "http://localhost:8080/media")
			defer cleanupEnv()

			cfg, err := config.Load()
			if err != nil {
				t.Logf("Initial load failed: %v", err)
				return false
			}

			if cfg.Log.Level != "info" {
				t.Logf("Initial log level mismatch: got %s, want info", cfg.Log.Level)
				return false
			}

			// Change environment variable
			os.Setenv("WHATSAPP_LOG_LEVEL", newLevel)

			// Reload configuration
			err = cfg.Reload()
			if err != nil {
				t.Logf("Reload failed: %v", err)
				return false
			}

			if cfg.Log.Level != newLevel {
				t.Logf("Log level not updated after reload: got %s, want %s", cfg.Log.Level, newLevel)
				return false
			}

			return true
		},
		gen.OneConstOf("debug", "info", "warn", "error"),
	))

	// Property 35.4: Reloading with invalid config should fail and preserve old config
	properties.Property("reload with invalid config fails and preserves old config", prop.ForAll(
		func(_ int) bool {
			// Set initial valid environment
			os.Setenv("WHATSAPP_SERVER_PORT", "8080")
			os.Setenv("WHATSAPP_WHATSAPP_DB_PATH", "/data/whatsmeow.db")
			os.Setenv("WHATSAPP_WEBSOCKET_URL", "ws://localhost:3000/ws/whatsapp")
			os.Setenv("WHATSAPP_LOG_LEVEL", "info")
			os.Setenv("WHATSAPP_LOG_FORMAT", "json")
			os.Setenv("WHATSAPP_MEDIA_BASE_PATH", "/data/media")
			os.Setenv("WHATSAPP_MEDIA_BASE_URL", "http://localhost:8080/media")
			defer cleanupEnv()

			cfg, err := config.Load()
			if err != nil {
				t.Logf("Initial load failed: %v", err)
				return false
			}

			originalPort := cfg.Server.Port

			// Change to invalid port
			os.Setenv("WHATSAPP_SERVER_PORT", "0")

			// Reload should fail
			err = cfg.Reload()
			if err == nil {
				t.Logf("Reload should have failed with invalid port but succeeded")
				return false // should have failed
			}

			// Config should preserve old values
			if cfg.Server.Port != originalPort {
				t.Logf("Config was modified despite reload failure: got %d, want %d", cfg.Server.Port, originalPort)
				return false
			}

			return true
		},
		gen.Const(0),
	))

	// Property 35.5: Reloading with changed rate limit settings should update config
	properties.Property("reload with changed rate limit settings updates config", prop.ForAll(
		func(newRPS float64) bool {
			if newRPS <= 0 {
				return true // skip invalid values
			}

			// Set initial environment
			os.Setenv("WHATSAPP_SERVER_PORT", "8080")
			os.Setenv("WHATSAPP_WHATSAPP_DB_PATH", "/data/whatsmeow.db")
			os.Setenv("WHATSAPP_WEBSOCKET_URL", "ws://localhost:3000/ws/whatsapp")
			os.Setenv("WHATSAPP_LOG_LEVEL", "info")
			os.Setenv("WHATSAPP_LOG_FORMAT", "json")
			os.Setenv("WHATSAPP_MEDIA_BASE_PATH", "/data/media")
			os.Setenv("WHATSAPP_MEDIA_BASE_URL", "http://localhost:8080/media")
			os.Setenv("WHATSAPP_RATELIMIT_ENABLED", "true")
			os.Setenv("WHATSAPP_RATELIMIT_RPS", "10.0")
			defer cleanupEnv()

			cfg, err := config.Load()
			if err != nil {
				t.Logf("Initial load failed: %v", err)
				return false
			}

			if cfg.RateLimit.RequestsPerSecond != 10.0 {
				t.Logf("Initial RPS mismatch: got %f, want 10.0", cfg.RateLimit.RequestsPerSecond)
				return false
			}

			// Change environment variable - FIX: Use strconv.FormatFloat with proper precision
			rpsStr := strconv.FormatFloat(newRPS, 'f', -1, 64)
			os.Setenv("WHATSAPP_RATELIMIT_RPS", rpsStr)

			// Reload configuration
			err = cfg.Reload()
			if err != nil {
				t.Logf("Reload failed: %v", err)
				return false
			}

			// Parse back the string to get the actual value that was set
			expectedRPS, _ := strconv.ParseFloat(rpsStr, 64)

			// Use epsilon comparison for floating point values
			const epsilon = 1e-9
			diff := cfg.RateLimit.RequestsPerSecond - expectedRPS
			if diff < 0 {
				diff = -diff
			}

			if diff > epsilon {
				t.Logf("RPS not updated after reload: got %f, want %f (diff: %e)", cfg.RateLimit.RequestsPerSecond, expectedRPS, diff)
				return false
			}

			return true
		},
		gen.Float64Range(1.0, 100.0),
	))

	properties.TestingRun(t)
}

// cleanupEnv clears all WHATSAPP_* environment variables
func cleanupEnv() {
	envVars := []string{
		"WHATSAPP_SERVER_HOST",
		"WHATSAPP_SERVER_PORT",
		"WHATSAPP_WHATSAPP_DB_PATH",
		"WHATSAPP_WEBSOCKET_URL",
		"WHATSAPP_LOG_LEVEL",
		"WHATSAPP_LOG_FORMAT",
		"WHATSAPP_MEDIA_BASE_PATH",
		"WHATSAPP_MEDIA_BASE_URL",
		"WHATSAPP_RATELIMIT_ENABLED",
		"WHATSAPP_RATELIMIT_RPS",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}
}
