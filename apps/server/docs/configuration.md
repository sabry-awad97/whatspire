# Configuration Reference

> Complete configuration reference for Whatspire WhatsApp Service

## Configuration Methods

The service supports three configuration methods with the following precedence:

1. **Environment Variables** (highest priority)
2. **Configuration Files** (YAML or JSON)
3. **Default Values** (lowest priority)

### File-Based Configuration

The service can load configuration from YAML or JSON files. Configuration files are searched in the following locations (in order):

1. File specified via `--config` flag (if provided)
2. `./config.yaml` or `./config.json` (current directory)
3. `./config/config.yaml` or `./config/config.json`
4. `/etc/whatspire/config.yaml` or `/etc/whatspire/config.json`
5. `$HOME/.whatspire/config.yaml` or `$HOME/.whatspire/config.json`

**Example: Using a config file**

```bash
# Start with specific config file
./whatsapp --config /path/to/config.yaml

# Or place config.yaml in current directory
./whatsapp
```

**Example config files:**

- See `config.example.yaml` for YAML format
- See `config.example.json` for JSON format

### Hot Reload

The service supports hot reload for non-critical configuration changes. When a configuration file is modified, the service automatically reloads the configuration without requiring a restart.

**Non-Critical Settings** (can be hot-reloaded):

- Log level and format
- Rate limiting settings
- CORS configuration
- Webhook settings
- Event retention settings
- Circuit breaker settings

**Critical Settings** (require restart):

- Server host and port
- Database connection (driver, DSN)
- WhatsApp database path
- WebSocket URL

When critical settings are changed, the service will log a warning indicating that a restart is required.

### Configuration Precedence Example

Given the following configurations:

**config.yaml:**

```yaml
server:
  port: 8080
log:
  level: "info"
```

**Environment variable:**

```bash
export WHATSAPP_LOG_LEVEL=debug
```

**Result:**

- `server.port` = 8080 (from file)
- `log.level` = "debug" (from environment, overrides file)

---

## Configuration Reference

### Server Configuration

| Variable               | Type   | Default   | Description    |
| ---------------------- | ------ | --------- | -------------- |
| `WHATSAPP_SERVER_HOST` | string | `0.0.0.0` | Listen address |
| `WHATSAPP_SERVER_PORT` | int    | `8080`    | Listen port    |

## WhatsApp Client

| Variable                      | Type     | Default              | Description               |
| ----------------------------- | -------- | -------------------- | ------------------------- |
| `WHATSAPP_DB_PATH`            | string   | `/data/whatsmeow.db` | SQLite database path      |
| `WHATSAPP_QR_TIMEOUT`         | duration | `2m`                 | QR code expiration        |
| `WHATSAPP_RECONNECT_DELAY`    | duration | `5s`                 | Delay between reconnects  |
| `WHATSAPP_MAX_RECONNECTS`     | int      | `10`                 | Max reconnection attempts |
| `WHATSAPP_MESSAGE_RATE_LIMIT` | int      | `30`                 | Messages per minute       |

## WebSocket

| Variable                           | Type     | Default                           | Description         |
| ---------------------------------- | -------- | --------------------------------- | ------------------- |
| `WHATSAPP_WEBSOCKET_URL`           | string   | `ws://localhost:3000/ws/whatsapp` | API server endpoint |
| `WHATSAPP_WEBSOCKET_API_KEY`       | string   | -                                 | Authentication key  |
| `WHATSAPP_WEBSOCKET_PING_INTERVAL` | duration | `30s`                             | Ping interval       |
| `WHATSAPP_WEBSOCKET_PONG_TIMEOUT`  | duration | `10s`                             | Pong timeout        |
| `WHATSAPP_WEBSOCKET_QUEUE_SIZE`    | int      | `1000`                            | Event queue size    |

## Logging

| Variable              | Type   | Default | Description              |
| --------------------- | ------ | ------- | ------------------------ |
| `WHATSAPP_LOG_LEVEL`  | string | `info`  | debug, info, warn, error |
| `WHATSAPP_LOG_FORMAT` | string | `json`  | json or text             |

## Rate Limiting

| Variable                        | Type  | Default | Description           |
| ------------------------------- | ----- | ------- | --------------------- |
| `WHATSAPP_RATELIMIT_ENABLED`    | bool  | `true`  | Enable rate limiting  |
| `WHATSAPP_RATELIMIT_RPS`        | float | `10.0`  | Requests per second   |
| `WHATSAPP_RATELIMIT_BURST`      | int   | `20`    | Burst capacity        |
| `WHATSAPP_RATELIMIT_BY_IP`      | bool  | `true`  | Rate limit by IP      |
| `WHATSAPP_RATELIMIT_BY_API_KEY` | bool  | `false` | Rate limit by API key |

## CORS

| Variable                          | Type     | Default                       | Description               |
| --------------------------------- | -------- | ----------------------------- | ------------------------- |
| `WHATSAPP_CORS_ORIGINS`           | []string | `*`                           | Allowed origins           |
| `WHATSAPP_CORS_METHODS`           | []string | `GET,POST,PUT,DELETE,OPTIONS` | Allowed methods           |
| `WHATSAPP_CORS_HEADERS`           | []string | See defaults                  | Allowed headers           |
| `WHATSAPP_CORS_ALLOW_CREDENTIALS` | bool     | `false`                       | Allow credentials         |
| `WHATSAPP_CORS_MAX_AGE`           | int      | `86400`                       | Preflight cache (seconds) |

## API Key Authentication

| Variable                   | Type     | Default     | Description           |
| -------------------------- | -------- | ----------- | --------------------- |
| `WHATSAPP_API_KEY_ENABLED` | bool     | `false`     | Enable authentication |
| `WHATSAPP_API_KEYS`        | []string | -           | Comma-separated keys  |
| `WHATSAPP_API_KEY_HEADER`  | string   | `X-API-Key` | Header name           |

## Metrics

| Variable                     | Type   | Default    | Description       |
| ---------------------------- | ------ | ---------- | ----------------- |
| `WHATSAPP_METRICS_ENABLED`   | bool   | `true`     | Enable Prometheus |
| `WHATSAPP_METRICS_PATH`      | string | `/metrics` | Endpoint path     |
| `WHATSAPP_METRICS_NAMESPACE` | string | `whatsapp` | Metric prefix     |

## Circuit Breaker

| Variable                                     | Type     | Default | Description            |
| -------------------------------------------- | -------- | ------- | ---------------------- |
| `WHATSAPP_CIRCUIT_BREAKER_ENABLED`           | bool     | `true`  | Enable circuit breaker |
| `WHATSAPP_CIRCUIT_BREAKER_MAX_REQUESTS`      | int      | `3`     | Half-open requests     |
| `WHATSAPP_CIRCUIT_BREAKER_INTERVAL`          | duration | `60s`   | Closed state interval  |
| `WHATSAPP_CIRCUIT_BREAKER_TIMEOUT`           | duration | `30s`   | Open to half-open      |
| `WHATSAPP_CIRCUIT_BREAKER_FAILURE_THRESHOLD` | int      | `5`     | Failures to open       |
| `WHATSAPP_CIRCUIT_BREAKER_SUCCESS_THRESHOLD` | int      | `2`     | Successes to close     |

## Media Storage

| Variable                       | Type   | Default                       | Description       |
| ------------------------------ | ------ | ----------------------------- | ----------------- |
| `WHATSAPP_MEDIA_BASE_PATH`     | string | `/data/media`                 | Storage directory |
| `WHATSAPP_MEDIA_BASE_URL`      | string | `http://localhost:8080/media` | Public URL        |
| `WHATSAPP_MEDIA_MAX_FILE_SIZE` | int    | `16777216`                    | Max size (16MB)   |

## Webhooks

| Variable                   | Type     | Default | Description     |
| -------------------------- | -------- | ------- | --------------- |
| `WHATSAPP_WEBHOOK_ENABLED` | bool     | `false` | Enable webhooks |
| `WHATSAPP_WEBHOOK_URL`     | string   | -       | Endpoint URL    |
| `WHATSAPP_WEBHOOK_SECRET`  | string   | -       | HMAC secret     |
| `WHATSAPP_WEBHOOK_EVENTS`  | []string | all     | Event filter    |

**Supported Events**: `message.received`, `message.sent`, `message.delivered`, `message.read`, `message.reaction`, `presence.update`, `session.connected`, `session.disconnected`
