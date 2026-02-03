# ADR-003: Viper for Configuration Management

**Date**: 2026-02-03  
**Status**: Accepted  
**Deciders**: Development Team  
**Technical Story**: Configuration Management - Flexible Config Loading

---

## Context and Problem Statement

System administrators needed a flexible way to configure the application across different environments (development, staging, production). The application required support for environment variables, configuration files (YAML/JSON), and hot-reloading of non-critical settings. Which configuration management library should we use?

## Decision Drivers

- Support for multiple configuration sources (env vars, files, defaults)
- Configuration precedence: env > file > defaults
- Hot-reload capability for non-critical settings
- Support for YAML and JSON formats
- Type-safe configuration access
- Validation of configuration values
- Zero-downtime configuration updates

## Considered Options

- **Option 1**: Viper (configuration management library)
- **Option 2**: envconfig (environment variable only)
- **Option 3**: Custom configuration loader
- **Option 4**: Standard library (os.Getenv + flag package)

## Decision Outcome

Chosen option: "**Viper**", because it provides comprehensive configuration management with hot-reload support, multiple formats, and a clean API while being battle-tested in production environments.

### Positive Consequences

- **Multiple Sources**: Supports env vars, files, and defaults in one place
- **Hot Reload**: Non-critical settings update without restart
- **Format Support**: YAML, JSON, TOML, and more
- **Precedence**: Clear configuration override hierarchy
- **Type Safety**: Strongly-typed configuration access
- **Validation**: Built-in validation support
- **Watch Support**: File system watching with fsnotify

### Negative Consequences

- **Dependency**: Adds external dependency (Viper + fsnotify)
- **Complexity**: More complex than simple env var reading
- **Learning Curve**: Team needs to understand Viper API
- **Overhead**: Slight performance overhead vs direct env var access

## Pros and Cons of the Options

### Option 1: Viper

- Good, because supports multiple configuration sources
- Good, because hot-reload with fsnotify integration
- Good, because widely used and well-maintained
- Good, because supports multiple file formats
- Good, because clear precedence rules
- Bad, because adds dependency
- Bad, because more complex than needed for simple cases

### Option 2: envconfig

- Good, because simple and focused
- Good, because minimal dependencies
- Good, because good for 12-factor apps
- Bad, because environment variables only
- Bad, because no hot-reload support
- Bad, because no file-based configuration
- Bad, because limited flexibility

### Option 3: Custom Configuration Loader

- Good, because complete control
- Good, because no dependencies
- Good, because tailored to our needs
- Bad, because reinventing the wheel
- Bad, because maintenance burden
- Bad, because likely to have bugs
- Bad, because no community support

### Option 4: Standard Library

- Good, because no dependencies
- Good, because simple and explicit
- Bad, because massive boilerplate
- Bad, because no hot-reload
- Bad, because no file format support
- Bad, because manual precedence handling

## Implementation Details

### Configuration Structure

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    WhatsApp WhatsAppConfig
    Events   EventsConfig
    Logging  LoggingConfig
}
```

### Configuration Sources (Precedence Order)

1. **Environment Variables** (highest priority)

   ```bash
   export SERVER_PORT=8080
   export DATABASE_DRIVER=postgres
   ```

2. **Configuration File** (YAML or JSON)

   ```yaml
   server:
     port: 8080
   database:
     driver: postgres
   ```

3. **Defaults** (lowest priority)
   ```go
   viper.SetDefault("server.port", 8080)
   ```

### Hot Reload Implementation

```go
viper.WatchConfig()
viper.OnConfigChange(func(e fsnotify.Event) {
    // Reload non-critical settings
    updateLoggingLevel()
    updateRateLimits()
    // Critical settings require restart
})
```

### Critical vs Non-Critical Settings

**Critical Settings** (require restart):

- Server port
- Database connection
- TLS certificates
- API keys

**Non-Critical Settings** (hot-reloadable):

- Log level
- Rate limits
- Webhook URLs
- Feature flags
- Timeouts

## Configuration Files

### Example YAML (config.example.yaml)

```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  driver: "sqlite"
  dsn: "./data/whatspire.db"

events:
  retention_days: 30
  cleanup_interval: "24h"
```

### Example JSON (config.example.json)

```json
{
  "server": {
    "port": 8080,
    "host": "0.0.0.0"
  },
  "database": {
    "driver": "sqlite",
    "dsn": "./data/whatspire.db"
  }
}
```

## Usage Patterns

### Loading Configuration

```go
// Load from file
viper.SetConfigName("config")
viper.SetConfigType("yaml")
viper.AddConfigPath(".")
viper.ReadInConfig()

// Override with environment variables
viper.AutomaticEnv()
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

// Unmarshal into struct
var config Config
viper.Unmarshal(&config)
```

### Validation

```go
func (c *Config) Validate() error {
    if c.Server.Port < 1 || c.Server.Port > 65535 {
        return errors.New("invalid port")
    }
    if c.Database.Driver != "sqlite" && c.Database.Driver != "postgres" {
        return errors.New("unsupported database driver")
    }
    return nil
}
```

## Links

- [Viper Documentation](https://github.com/spf13/viper)
- [fsnotify Documentation](https://github.com/fsnotify/fsnotify)
- Related: ADR-001 (Clean Architecture)
- See: `apps/server/docs/configuration.md` for detailed config reference

---

## Notes

### Environment Variable Mapping

Viper automatically maps environment variables:

- `SERVER_PORT` → `server.port`
- `DATABASE_DRIVER` → `database.driver`
- `EVENTS_RETENTION_DAYS` → `events.retention_days`

### Hot Reload Safety

To prevent race conditions during hot reload:

1. Use atomic operations for config updates
2. Validate new config before applying
3. Log all configuration changes
4. Provide rollback mechanism

### Testing Strategy

- **Unit Tests**: Use in-memory config
- **Integration Tests**: Test file loading and env var override
- **Hot Reload Tests**: Verify non-critical settings update
- **Validation Tests**: Test all validation rules

### Docker Considerations

In Docker environments:

- Mount config file as volume: `-v ./config.yaml:/app/config.yaml`
- Use environment variables for secrets
- Use Docker secrets for sensitive data

### Kubernetes Considerations

In Kubernetes:

- Use ConfigMaps for configuration files
- Use Secrets for sensitive data
- Use environment variables for pod-specific config
- Hot reload works with ConfigMap updates

### Migration from Environment Variables Only

If moving from env-only configuration:

1. Keep env var support (backward compatible)
2. Add config file support gradually
3. Document migration path
4. Provide conversion tool

### Future Enhancements

- **Remote Configuration**: Viper supports etcd, Consul
- **Encrypted Secrets**: Add vault integration
- **Configuration Versioning**: Track config changes
- **A/B Testing**: Feature flags with hot reload
