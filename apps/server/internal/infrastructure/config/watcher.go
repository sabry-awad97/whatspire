package config

import (
	"context"
	"log"
	"strings"
	"sync"

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
}

// GetViper returns the underlying viper instance (for checking if config file is used)
func (w *ConfigWatcher) GetViper() *viper.Viper {
	return w.viper
}

// NewConfigWatcher creates a new configuration watcher
func NewConfigWatcher(configFile string) (*ConfigWatcher, error) {
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
			log.Println("⚠️  No config file found, using defaults and environment variables")
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
		log.Printf("🔄 Config file changed: %s", e.Name)

		// Reload configuration
		if err := w.reload(); err != nil {
			log.Printf("⚠️  Failed to reload config: %v", err)
			return
		}

		log.Println("✅ Configuration reloaded successfully")
	})

	log.Println("✅ Config watcher started")

	// Keep watcher running
	go func() {
		<-ctx.Done()
		if err := w.Stop(); err != nil {
			log.Printf("⚠️ Error stopping config watcher: %v", err)
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
	log.Println("✅ Config watcher stopped")
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
		log.Println("⚠️  Critical configuration changes detected - restart required for full effect")
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
		log.Printf("📝 Log level changed: %s -> %s", oldCfg.Log.Level, newCfg.Log.Level)
	}

	// Rate limit changes
	if oldCfg.RateLimit.RequestsPerSecond != newCfg.RateLimit.RequestsPerSecond {
		log.Printf("🚦 Rate limit changed: %.2f -> %.2f requests/second",
			oldCfg.RateLimit.RequestsPerSecond, newCfg.RateLimit.RequestsPerSecond)
	}

	// Event retention changes
	if oldCfg.Events.RetentionDays != newCfg.Events.RetentionDays {
		log.Printf("🗄️  Event retention changed: %d -> %d days",
			oldCfg.Events.RetentionDays, newCfg.Events.RetentionDays)
	}

	// Webhook changes
	if oldCfg.Webhook.Enabled != newCfg.Webhook.Enabled {
		log.Printf("🔔 Webhook enabled changed: %v -> %v", oldCfg.Webhook.Enabled, newCfg.Webhook.Enabled)
	}
}

// IsRunning returns whether the watcher is currently running
func (w *ConfigWatcher) IsRunning() bool {
	return w.running
}
