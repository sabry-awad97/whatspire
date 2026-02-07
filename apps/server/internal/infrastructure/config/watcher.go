package config

import (
	"context"
	"strings"
	"sync"
	"whatspire/internal/infrastructure/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ConfigWatcher watches for configuration file changes and reloads config
type ConfigWatcher struct {
	viper       *viper.Viper
	config      *Config
	configMutex sync.RWMutex
	stopCh      chan struct{}
	callbacks   []func(*Config)
	running     bool
	logger      *logger.Logger
}

// GetViper returns the underlying viper instance (for checking if config file is used)
func (w *ConfigWatcher) GetViper() *viper.Viper {
	return w.viper
}

// NewConfigWatcher creates a new configuration watcher
func NewConfigWatcher(configFile string, log *logger.Logger) (*ConfigWatcher, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Load from config file if provided
	if configFile != "" {
		v.SetConfigFile(configFile)

		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	} else {
		// Try to find config file in standard locations
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/whatspire")
		v.AddConfigPath("$HOME/.whatspire")

		if err := v.ReadInConfig(); err != nil {
			// No config file found, use defaults and env vars only
			log.Warn("No configuration file found, using defaults and environment variables")
		}
	}

	// Enable reading from environment variables
	v.SetEnvPrefix("WHATSAPP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind environment variables explicitly
	bindEnvVars(v)

	// Unmarshal initial config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &ConfigWatcher{
		viper:     v,
		config:    &cfg,
		stopCh:    make(chan struct{}),
		callbacks: make([]func(*Config), 0),
		logger:    log.Sub("config_watcher"),
	}, nil
}

// Start starts watching for configuration changes
func (w *ConfigWatcher) Start(ctx context.Context) error {
	if w.running {
		return nil
	}

	w.running = true

	// Watch for config file changes
	w.viper.WatchConfig()
	w.viper.OnConfigChange(func(e fsnotify.Event) {
		w.logger.WithFields(map[string]interface{}{
			"file":      e.Name,
			"operation": e.Op.String(),
		}).Info("Configuration file changed, reloading")

		// Reload configuration
		if err := w.reload(); err != nil {
			w.logger.WithError(err).Error("Failed to reload configuration after file change")
			return
		}

		w.logger.Info("Configuration reloaded successfully")
	})

	w.logger.Info("Configuration watcher started successfully")

	// Keep watcher running
	go func() {
		<-ctx.Done()
		if err := w.Stop(); err != nil {
			w.logger.WithError(err).Warn("Error occurred while stopping configuration watcher")
		}
	}()

	return nil
}

// Stop stops the configuration watcher
func (w *ConfigWatcher) Stop() error {
	if !w.running {
		return nil
	}

	w.running = false
	close(w.stopCh)
	w.logger.Info("Configuration watcher stopped gracefully")
	return nil
}

// reload reloads the configuration from file
func (w *ConfigWatcher) reload() error {
	w.configMutex.Lock()
	defer w.configMutex.Unlock()

	// Unmarshal new config
	var newCfg Config
	if err := w.viper.Unmarshal(&newCfg); err != nil {
		return err
	}

	// Validate new configuration
	if err := newCfg.Validate(); err != nil {
		return err
	}

	// Check if critical settings changed (these require restart)
	if w.hasCriticalChanges(&newCfg) {
		w.logger.Warn("Critical configuration changes detected - application restart required for full effect")
	}

	// Update config
	oldCfg := w.config
	w.config = &newCfg

	// Notify callbacks
	for _, callback := range w.callbacks {
		callback(&newCfg)
	}

	// Log non-critical changes
	w.logConfigChanges(oldCfg, &newCfg)

	return nil
}

// GetConfig returns the current configuration (thread-safe)
func (w *ConfigWatcher) GetConfig() *Config {
	w.configMutex.RLock()
	defer w.configMutex.RUnlock()
	return w.config
}

// OnConfigChange registers a callback for configuration changes
func (w *ConfigWatcher) OnConfigChange(callback func(*Config)) {
	w.callbacks = append(w.callbacks, callback)
}

// hasCriticalChanges checks if critical settings that require restart have changed
func (w *ConfigWatcher) hasCriticalChanges(newCfg *Config) bool {
	oldCfg := w.config

	// Critical settings that require restart
	criticalChanges := []bool{
		oldCfg.Server.Host != newCfg.Server.Host,
		oldCfg.Server.Port != newCfg.Server.Port,
		oldCfg.WhatsApp.DBPath != newCfg.WhatsApp.DBPath,
		oldCfg.Database.Driver != newCfg.Database.Driver,
		oldCfg.Database.DSN != newCfg.Database.DSN,
		oldCfg.WebSocket.URL != newCfg.WebSocket.URL,
	}

	for _, changed := range criticalChanges {
		if changed {
			return true
		}
	}

	return false
}

// logConfigChanges logs non-critical configuration changes
func (w *ConfigWatcher) logConfigChanges(oldCfg, newCfg *Config) {
	// Log level changes
	if oldCfg.Log.Level != newCfg.Log.Level {
		w.logger.WithFields(map[string]interface{}{
			"old_level": oldCfg.Log.Level,
			"new_level": newCfg.Log.Level,
		}).Info("Log level configuration changed")
	}

	// Rate limit changes
	if oldCfg.RateLimit.RequestsPerSecond != newCfg.RateLimit.RequestsPerSecond {
		w.logger.WithFields(map[string]interface{}{
			"old_rate": oldCfg.RateLimit.RequestsPerSecond,
			"new_rate": newCfg.RateLimit.RequestsPerSecond,
		}).Info("Rate limit configuration changed")
	}

	// Event retention changes
	if oldCfg.Events.RetentionDays != newCfg.Events.RetentionDays {
		w.logger.WithFields(map[string]interface{}{
			"old_retention_days": oldCfg.Events.RetentionDays,
			"new_retention_days": newCfg.Events.RetentionDays,
		}).Info("Event retention policy changed")
	}
}

// IsRunning returns whether the watcher is currently running
func (w *ConfigWatcher) IsRunning() bool {
	return w.running
}
