package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the WhatsApp service
type Config struct {
	// Server configuration
	Server ServerConfig `mapstructure:"server"`

	// WhatsApp client configuration (includes whatsmeow database path)
	WhatsApp WhatsAppConfig `mapstructure:"whatsapp"`

	// WebSocket configuration for API server connection
	WebSocket WebSocketConfig `mapstructure:"websocket"`

	// Logging configuration
	Log LogConfig `mapstructure:"log"`

	// Rate limiting configuration
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`

	// CORS configuration
	CORS CORSConfig `mapstructure:"cors"`

	// API Key authentication configuration
	APIKey APIKeyConfig `mapstructure:"apikey"`

	// Metrics configuration
	Metrics MetricsConfig `mapstructure:"metrics"`

	// Circuit breaker configuration
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuitbreaker"`

	// Media storage configuration
	Media MediaConfig `mapstructure:"media"`

	// Webhook configuration
	Webhook WebhookConfig `mapstructure:"webhook"`

	// Database configuration
	Database DatabaseConfig `mapstructure:"database"`

	// Event storage configuration
	Events EventsConfig `mapstructure:"events"`
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	MaxRequests      uint32        `mapstructure:"max_requests"`      // Max requests in half-open state
	Interval         time.Duration `mapstructure:"interval"`          // Interval for clearing counts in closed state
	Timeout          time.Duration `mapstructure:"timeout"`           // Timeout before transitioning from open to half-open
	FailureThreshold uint32        `mapstructure:"failure_threshold"` // Consecutive failures to open circuit
	SuccessThreshold uint32        `mapstructure:"success_threshold"` // Consecutive successes to close circuit
}

// MediaConfig holds media storage configuration
type MediaConfig struct {
	BasePath    string `mapstructure:"base_path"`     // Local directory for storing media files
	BaseURL     string `mapstructure:"base_url"`      // Public URL prefix for accessing media
	MaxFileSize int64  `mapstructure:"max_file_size"` // Maximum file size in bytes (default: 16MB)
}

// WebhookConfig holds webhook delivery configuration
type WebhookConfig struct {
	Enabled bool     `mapstructure:"enabled"` // Enable webhook delivery
	URL     string   `mapstructure:"url"`     // Webhook endpoint URL
	Secret  string   `mapstructure:"secret"`  // Secret for HMAC signing (optional)
	Events  []string `mapstructure:"events"`  // Event types to deliver (empty = all events)
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`            // Database driver: "sqlite" or "postgres"
	DSN             string        `mapstructure:"dsn"`               // Data Source Name (connection string)
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`    // Maximum idle connections
	MaxOpenConns    int           `mapstructure:"max_open_conns"`    // Maximum open connections
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"` // Connection maximum lifetime
	LogLevel        string        `mapstructure:"log_level"`         // GORM log level: "silent", "error", "warn", "info"
}

// EventsConfig holds event storage and retention configuration
type EventsConfig struct {
	Enabled         bool          `mapstructure:"enabled"`          // Enable event storage
	RetentionDays   int           `mapstructure:"retention_days"`   // Days to retain events (0 = forever)
	CleanupTime     string        `mapstructure:"cleanup_time"`     // Daily cleanup time in UTC (HH:MM format, e.g., "02:00")
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"` // Cleanup check interval (default: 1 hour)
}

// MetricsConfig holds Prometheus metrics configuration
type MetricsConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Path      string `mapstructure:"path"`      // Metrics endpoint path (default: /metrics)
	Namespace string `mapstructure:"namespace"` // Prometheus namespace (default: whatsapp)
}

// Role represents an API key role for authorization
type Role string

const (
	RoleRead  Role = "read"
	RoleWrite Role = "write"
	RoleAdmin Role = "admin"
)

// APIKeyConfig holds API key authentication configuration
// Note: API keys are now managed in the database. This config only controls
// whether API key authentication is enabled and which header to use.
type APIKeyConfig struct {
	Enabled bool   `mapstructure:"enabled"` // Enable API key authentication
	Header  string `mapstructure:"header"`  // Header name for API key (default: X-API-Key)
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"` // seconds
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	RequestsPerSecond float64       `mapstructure:"requests_per_second"`
	BurstSize         int           `mapstructure:"burst_size"`
	ByIP              bool          `mapstructure:"by_ip"`
	ByAPIKey          bool          `mapstructure:"by_api_key"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
	MaxAge            time.Duration `mapstructure:"max_age"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// WhatsAppConfig holds WhatsApp client configuration
type WhatsAppConfig struct {
	DBPath           string        `mapstructure:"db_path"` // Whatsmeow SQLite database path
	QRTimeout        time.Duration `mapstructure:"qr_timeout"`
	ReconnectDelay   time.Duration `mapstructure:"reconnect_delay"`
	MaxReconnects    int           `mapstructure:"max_reconnects"`
	MessageRateLimit int           `mapstructure:"message_rate_limit"` // messages per minute
}

// WebSocketConfig holds WebSocket configuration for API server connection
type WebSocketConfig struct {
	URL            string        `mapstructure:"url"`
	APIKey         string        `mapstructure:"api_key"` // API key for authentication with Node.js API
	PingInterval   time.Duration `mapstructure:"ping_interval"`
	PongTimeout    time.Duration `mapstructure:"pong_timeout"`
	ReconnectDelay time.Duration `mapstructure:"reconnect_delay"`
	MaxReconnects  int           `mapstructure:"max_reconnects"`
	QueueSize      int           `mapstructure:"queue_size"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // json or text
}

// Address returns the server address in host:port format
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("config validation error: %s - %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Validate validates the configuration and returns any errors
func (c *Config) Validate() error {
	var errs ValidationErrors

	// Validate Server config
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, ValidationError{
			Field:   "server.port",
			Message: "must be between 1 and 65535",
		})
	}

	// Validate WhatsApp config
	if c.WhatsApp.DBPath == "" {
		errs = append(errs, ValidationError{
			Field:   "whatsapp.db_path",
			Message: "is required",
		})
	}
	if c.WhatsApp.QRTimeout <= 0 {
		errs = append(errs, ValidationError{
			Field:   "whatsapp.qr_timeout",
			Message: "must be positive",
		})
	}
	if c.WhatsApp.MaxReconnects < 0 {
		errs = append(errs, ValidationError{
			Field:   "whatsapp.max_reconnects",
			Message: "must be non-negative",
		})
	}
	if c.WhatsApp.MessageRateLimit < 0 {
		errs = append(errs, ValidationError{
			Field:   "whatsapp.message_rate_limit",
			Message: "must be non-negative",
		})
	}

	// Validate WebSocket config
	if c.WebSocket.URL == "" {
		errs = append(errs, ValidationError{
			Field:   "websocket.url",
			Message: "is required",
		})
	}
	if c.WebSocket.PingInterval <= 0 {
		errs = append(errs, ValidationError{
			Field:   "websocket.ping_interval",
			Message: "must be positive",
		})
	}
	if c.WebSocket.PongTimeout <= 0 {
		errs = append(errs, ValidationError{
			Field:   "websocket.pong_timeout",
			Message: "must be positive",
		})
	}
	if c.WebSocket.QueueSize < 0 {
		errs = append(errs, ValidationError{
			Field:   "websocket.queue_size",
			Message: "must be non-negative",
		})
	}

	// Validate Log config
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLogLevels[strings.ToLower(c.Log.Level)] {
		errs = append(errs, ValidationError{
			Field:   "log.level",
			Message: "must be one of: debug, info, warn, error",
		})
	}
	validLogFormats := map[string]bool{
		"json": true, "text": true,
	}
	if !validLogFormats[strings.ToLower(c.Log.Format)] {
		errs = append(errs, ValidationError{
			Field:   "log.format",
			Message: "must be one of: json, text",
		})
	}

	// Validate RateLimit config
	if c.RateLimit.Enabled {
		if c.RateLimit.RequestsPerSecond <= 0 {
			errs = append(errs, ValidationError{
				Field:   "ratelimit.requests_per_second",
				Message: "must be positive when rate limiting is enabled",
			})
		}
		if c.RateLimit.BurstSize <= 0 {
			errs = append(errs, ValidationError{
				Field:   "ratelimit.burst_size",
				Message: "must be positive when rate limiting is enabled",
			})
		}
	}

	// Validate Media config
	if c.Media.BasePath == "" {
		errs = append(errs, ValidationError{
			Field:   "media.base_path",
			Message: "is required",
		})
	}
	if c.Media.BaseURL == "" {
		errs = append(errs, ValidationError{
			Field:   "media.base_url",
			Message: "is required",
		})
	}
	if c.Media.MaxFileSize <= 0 {
		errs = append(errs, ValidationError{
			Field:   "media.max_file_size",
			Message: "must be positive",
		})
	}

	// Validate Webhook config
	if c.Webhook.Enabled {
		if c.Webhook.URL == "" {
			errs = append(errs, ValidationError{
				Field:   "webhook.url",
				Message: "is required when webhooks are enabled",
			})
		}
		// Validate event types if specified
		validEvents := map[string]bool{
			"message.received":     true,
			"message.sent":         true,
			"message.delivered":    true,
			"message.read":         true,
			"message.reaction":     true,
			"presence.update":      true,
			"session.connected":    true,
			"session.disconnected": true,
		}
		for _, event := range c.Webhook.Events {
			if !validEvents[event] {
				errs = append(errs, ValidationError{
					Field:   "webhook.events",
					Message: fmt.Sprintf("invalid event type: %s", event),
				})
			}
		}
	}

	// Validate API Key config
	if c.APIKey.Enabled {
		// API keys are now managed in the database
		// No validation needed for config-based keys
		if c.APIKey.Header == "" {
			c.APIKey.Header = "X-API-Key" // Set default header
		}
	}

	// Validate Database config
	validDrivers := map[string]bool{
		"sqlite":   true,
		"postgres": true,
	}
	if !validDrivers[c.Database.Driver] {
		errs = append(errs, ValidationError{
			Field:   "database.driver",
			Message: "must be one of: sqlite, postgres",
		})
	}
	if c.Database.DSN == "" {
		errs = append(errs, ValidationError{
			Field:   "database.dsn",
			Message: "is required",
		})
	}
	if c.Database.MaxIdleConns < 0 {
		errs = append(errs, ValidationError{
			Field:   "database.max_idle_conns",
			Message: "must be non-negative",
		})
	}
	if c.Database.MaxOpenConns < 0 {
		errs = append(errs, ValidationError{
			Field:   "database.max_open_conns",
			Message: "must be non-negative",
		})
	}
	if c.Database.ConnMaxLifetime < 0 {
		errs = append(errs, ValidationError{
			Field:   "database.conn_max_lifetime",
			Message: "must be non-negative",
		})
	}

	// Validate Events config
	if c.Events.Enabled {
		if c.Events.RetentionDays < 0 {
			errs = append(errs, ValidationError{
				Field:   "events.retention_days",
				Message: "must be non-negative (0 = forever)",
			})
		}
		if c.Events.CleanupTime != "" {
			// Validate HH:MM format
			parts := strings.Split(c.Events.CleanupTime, ":")
			if len(parts) != 2 {
				errs = append(errs, ValidationError{
					Field:   "events.cleanup_time",
					Message: "must be in HH:MM format (e.g., 02:00)",
				})
			}
		}
		if c.Events.CleanupInterval <= 0 {
			errs = append(errs, ValidationError{
				Field:   "events.cleanup_interval",
				Message: "must be positive",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	return LoadWithConfigFile("")
}

// LoadWithConfigFile loads configuration from a file (if provided) and environment variables
func LoadWithConfigFile(configFile string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Load from config file if provided
	if configFile != "" {
		v.SetConfigFile(configFile)

		// Read config file
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		// Try to find config file in standard locations
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/whatspire")
		v.AddConfigPath("$HOME/.whatspire")

		// Read config file (ignore error if not found)
		_ = v.ReadInConfig()
	}

	// Enable reading from environment variables (overrides file config)
	v.SetEnvPrefix("WHATSAPP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind environment variables explicitly
	bindEnvVars(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadWithViper loads configuration using a provided viper instance (for testing)
func LoadWithViper(v *viper.Viper) (*Config, error) {
	setDefaults(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)

	// WhatsApp defaults
	v.SetDefault("whatsapp.db_path", "/data/whatsmeow.db")
	v.SetDefault("whatsapp.qr_timeout", 2*time.Minute)
	v.SetDefault("whatsapp.reconnect_delay", 5*time.Second)
	v.SetDefault("whatsapp.max_reconnects", 10)
	v.SetDefault("whatsapp.message_rate_limit", 30)

	// WebSocket defaults
	v.SetDefault("websocket.url", "ws://localhost:3000/ws/whatsapp")
	v.SetDefault("websocket.api_key", "") // API key for authenticating with Node.js API
	v.SetDefault("websocket.ping_interval", 30*time.Second)
	v.SetDefault("websocket.pong_timeout", 10*time.Second)
	v.SetDefault("websocket.reconnect_delay", 5*time.Second)
	v.SetDefault("websocket.max_reconnects", 0) // 0 = unlimited
	v.SetDefault("websocket.queue_size", 1000)

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	// RateLimit defaults
	v.SetDefault("ratelimit.enabled", true)
	v.SetDefault("ratelimit.requests_per_second", 10.0)
	v.SetDefault("ratelimit.burst_size", 20)
	v.SetDefault("ratelimit.by_ip", true)
	v.SetDefault("ratelimit.by_api_key", false)
	v.SetDefault("ratelimit.cleanup_interval", 5*time.Minute)
	v.SetDefault("ratelimit.max_age", time.Hour)

	// CORS defaults
	v.SetDefault("cors.allowed_origins", []string{"*"})
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowed_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID", "X-API-Key"})
	v.SetDefault("cors.expose_headers", []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"})
	v.SetDefault("cors.allow_credentials", false)
	v.SetDefault("cors.max_age", 86400) // 24 hours

	// APIKey defaults
	v.SetDefault("apikey.enabled", true)
	v.SetDefault("apikey.header", "X-API-Key")

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.path", "/metrics")
	v.SetDefault("metrics.namespace", "whatsapp")

	// CircuitBreaker defaults
	v.SetDefault("circuitbreaker.enabled", true)
	v.SetDefault("circuitbreaker.max_requests", 3)
	v.SetDefault("circuitbreaker.interval", 60*time.Second)
	v.SetDefault("circuitbreaker.timeout", 30*time.Second)
	v.SetDefault("circuitbreaker.failure_threshold", 5)
	v.SetDefault("circuitbreaker.success_threshold", 2)

	// Media defaults
	v.SetDefault("media.base_path", "/data/media")
	v.SetDefault("media.base_url", "http://localhost:8080/media")
	v.SetDefault("media.max_file_size", 16*1024*1024) // 16MB

	// Webhook defaults
	v.SetDefault("webhook.enabled", false)
	v.SetDefault("webhook.url", "")
	v.SetDefault("webhook.secret", "")
	v.SetDefault("webhook.events", []string{}) // Empty = all events

	// Database defaults
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", "/data/application.db")
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.conn_max_lifetime", time.Hour)
	v.SetDefault("database.log_level", "warn")

	// Events defaults
	v.SetDefault("events.enabled", false)
	v.SetDefault("events.retention_days", 30)
	v.SetDefault("events.cleanup_time", "02:00")
	v.SetDefault("events.cleanup_interval", time.Hour)
}

func bindEnvVars(v *viper.Viper) {
	// Server
	_ = v.BindEnv("server.host", "WHATSAPP_SERVER_HOST")
	_ = v.BindEnv("server.port", "WHATSAPP_SERVER_PORT")

	// WhatsApp
	_ = v.BindEnv("whatsapp.db_path", "WHATSAPP_DB_PATH")
	_ = v.BindEnv("whatsapp.qr_timeout", "WHATSAPP_QR_TIMEOUT")
	_ = v.BindEnv("whatsapp.reconnect_delay", "WHATSAPP_RECONNECT_DELAY")
	_ = v.BindEnv("whatsapp.max_reconnects", "WHATSAPP_MAX_RECONNECTS")
	_ = v.BindEnv("whatsapp.message_rate_limit", "WHATSAPP_MESSAGE_RATE_LIMIT")

	// WebSocket
	_ = v.BindEnv("websocket.url", "WHATSAPP_WEBSOCKET_URL", "API_WEBHOOK_URL")
	_ = v.BindEnv("websocket.api_key", "WHATSAPP_WEBSOCKET_API_KEY")
	_ = v.BindEnv("websocket.ping_interval", "WHATSAPP_WEBSOCKET_PING_INTERVAL")
	_ = v.BindEnv("websocket.pong_timeout", "WHATSAPP_WEBSOCKET_PONG_TIMEOUT")
	_ = v.BindEnv("websocket.reconnect_delay", "WHATSAPP_WEBSOCKET_RECONNECT_DELAY")
	_ = v.BindEnv("websocket.max_reconnects", "WHATSAPP_WEBSOCKET_MAX_RECONNECTS")
	_ = v.BindEnv("websocket.queue_size", "WHATSAPP_WEBSOCKET_QUEUE_SIZE")

	// Log
	_ = v.BindEnv("log.level", "WHATSAPP_LOG_LEVEL", "LOG_LEVEL")
	_ = v.BindEnv("log.format", "WHATSAPP_LOG_FORMAT", "LOG_FORMAT")

	// RateLimit
	_ = v.BindEnv("ratelimit.enabled", "WHATSAPP_RATELIMIT_ENABLED")
	_ = v.BindEnv("ratelimit.requests_per_second", "WHATSAPP_RATELIMIT_RPS")
	_ = v.BindEnv("ratelimit.burst_size", "WHATSAPP_RATELIMIT_BURST")
	_ = v.BindEnv("ratelimit.by_ip", "WHATSAPP_RATELIMIT_BY_IP")
	_ = v.BindEnv("ratelimit.by_api_key", "WHATSAPP_RATELIMIT_BY_API_KEY")
	_ = v.BindEnv("ratelimit.cleanup_interval", "WHATSAPP_RATELIMIT_CLEANUP_INTERVAL")
	_ = v.BindEnv("ratelimit.max_age", "WHATSAPP_RATELIMIT_MAX_AGE")

	// CORS
	_ = v.BindEnv("cors.allowed_origins", "WHATSAPP_CORS_ORIGINS")
	_ = v.BindEnv("cors.allowed_methods", "WHATSAPP_CORS_METHODS")
	_ = v.BindEnv("cors.allowed_headers", "WHATSAPP_CORS_HEADERS")
	_ = v.BindEnv("cors.expose_headers", "WHATSAPP_CORS_EXPOSE_HEADERS")
	_ = v.BindEnv("cors.allow_credentials", "WHATSAPP_CORS_ALLOW_CREDENTIALS")
	_ = v.BindEnv("cors.max_age", "WHATSAPP_CORS_MAX_AGE")

	// APIKey
	_ = v.BindEnv("apikey.enabled", "WHATSAPP_API_KEY_ENABLED")
	_ = v.BindEnv("apikey.header", "WHATSAPP_API_KEY_HEADER")

	// Metrics
	_ = v.BindEnv("metrics.enabled", "WHATSAPP_METRICS_ENABLED")
	_ = v.BindEnv("metrics.path", "WHATSAPP_METRICS_PATH")
	_ = v.BindEnv("metrics.namespace", "WHATSAPP_METRICS_NAMESPACE")

	// CircuitBreaker
	_ = v.BindEnv("circuitbreaker.enabled", "WHATSAPP_CIRCUIT_BREAKER_ENABLED")
	_ = v.BindEnv("circuitbreaker.max_requests", "WHATSAPP_CIRCUIT_BREAKER_MAX_REQUESTS")
	_ = v.BindEnv("circuitbreaker.interval", "WHATSAPP_CIRCUIT_BREAKER_INTERVAL")
	_ = v.BindEnv("circuitbreaker.timeout", "WHATSAPP_CIRCUIT_BREAKER_TIMEOUT")
	_ = v.BindEnv("circuitbreaker.failure_threshold", "WHATSAPP_CIRCUIT_BREAKER_FAILURE_THRESHOLD")
	_ = v.BindEnv("circuitbreaker.success_threshold", "WHATSAPP_CIRCUIT_BREAKER_SUCCESS_THRESHOLD")

	// Media
	_ = v.BindEnv("media.base_path", "WHATSAPP_MEDIA_BASE_PATH")
	_ = v.BindEnv("media.base_url", "WHATSAPP_MEDIA_BASE_URL")
	_ = v.BindEnv("media.max_file_size", "WHATSAPP_MEDIA_MAX_FILE_SIZE")

	// Webhook
	_ = v.BindEnv("webhook.enabled", "WHATSAPP_WEBHOOK_ENABLED")
	_ = v.BindEnv("webhook.url", "WHATSAPP_WEBHOOK_URL")
	_ = v.BindEnv("webhook.secret", "WHATSAPP_WEBHOOK_SECRET")
	_ = v.BindEnv("webhook.events", "WHATSAPP_WEBHOOK_EVENTS")

	// Database
	_ = v.BindEnv("database.driver", "WHATSAPP_DATABASE_DRIVER")
	_ = v.BindEnv("database.dsn", "WHATSAPP_DATABASE_DSN")
	_ = v.BindEnv("database.max_idle_conns", "WHATSAPP_DATABASE_MAX_IDLE_CONNS")
	_ = v.BindEnv("database.max_open_conns", "WHATSAPP_DATABASE_MAX_OPEN_CONNS")
	_ = v.BindEnv("database.conn_max_lifetime", "WHATSAPP_DATABASE_CONN_MAX_LIFETIME")
	_ = v.BindEnv("database.log_level", "WHATSAPP_DATABASE_LOG_LEVEL")

	// Events
	_ = v.BindEnv("events.enabled", "WHATSAPP_EVENTS_ENABLED")
	_ = v.BindEnv("events.retention_days", "WHATSAPP_EVENTS_RETENTION_DAYS")
	_ = v.BindEnv("events.cleanup_time", "WHATSAPP_EVENTS_CLEANUP_TIME")
	_ = v.BindEnv("events.cleanup_interval", "WHATSAPP_EVENTS_CLEANUP_INTERVAL")
}

// MustLoad loads configuration and panics on error (for use in main)
func MustLoad() *Config {
	return MustLoadWithConfigFile("")
}

// MustLoadWithConfigFile loads configuration from a file and panics on error
func MustLoadWithConfigFile(configFile string) *Config {
	cfg, err := LoadWithConfigFile(configFile)
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
	return cfg
}

// Reload reloads configuration from environment variables
// This allows configuration changes without restarting the service
func (c *Config) Reload() error {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Enable reading from environment variables
	v.SetEnvPrefix("WHATSAPP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind environment variables explicitly
	bindEnvVars(v)

	var newCfg Config
	if err := v.Unmarshal(&newCfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate new configuration before applying
	if err := newCfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Apply new configuration
	*c = newCfg

	return nil
}
