package config

import (
	"context"
	"log"
	"os"

	"go.uber.org/fx"
)

// configFileFlag holds the config file path (can be set programmatically for testing)
var configFileFlag string

func init() {
	// Check if --config flag is provided
	for i, arg := range os.Args {
		if arg == "--config" || arg == "-config" {
			if i+1 < len(os.Args) {
				configFileFlag = os.Args[i+1]
				break
			}
		}
	}
}

// Module provides configuration dependencies
var Module = fx.Module("config",
	fx.Provide(
		ProvideConfig,
		ProvideConfigWatcher,
	),
	fx.Invoke(StartConfigWatcher),
)

// ProvideConfig provides the configuration instance
func ProvideConfig() (*Config, error) {
	// Load config from file (if specified) or auto-discover
	cfg, err := LoadWithConfigFile(configFileFlag)
	if err != nil {
		return nil, err
	}

	if configFileFlag != "" {
		log.Printf("✅ Configuration loaded from: %s", configFileFlag)
	} else {
		log.Println("✅ Configuration loaded from environment variables and defaults")
	}

	return cfg, nil
}

// ProvideConfigWatcher provides the configuration watcher for hot reload
func ProvideConfigWatcher() (*ConfigWatcher, error) {
	// Create watcher
	watcher, err := NewConfigWatcher(configFileFlag)
	if err != nil {
		return nil, err
	}

	return watcher, nil
}

// StartConfigWatcher starts the configuration watcher if a config file is being used
func StartConfigWatcher(lc fx.Lifecycle, watcher *ConfigWatcher) {
	// Only start watcher if a config file is being used
	if watcher.GetViper().ConfigFileUsed() == "" {
		log.Println("ℹ️  Config watcher not started (no config file in use)")
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := watcher.Start(ctx); err != nil {
				log.Printf("⚠️  Failed to start config watcher: %v", err)
				// Don't fail startup if watcher fails
				return nil
			}
			log.Printf("✅ Config watcher monitoring: %s", watcher.GetViper().ConfigFileUsed())
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return watcher.Stop()
		},
	})
}

// SetConfigFile sets the config file path (for testing)
func SetConfigFile(path string) {
	configFileFlag = path
}
