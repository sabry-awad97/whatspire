# Deployment Guide

> Production deployment instructions for Whatspire WhatsApp Service

## Prerequisites

- Go 1.24+ (for building)
- Docker (recommended)
- Persistent storage for WhatsApp session data
- Network access to WhatsApp servers

---

## Environment Configuration

### Required Variables

| Variable                   | Description              | Example                         |
| -------------------------- | ------------------------ | ------------------------------- |
| `WHATSAPP_DB_PATH`         | SQLite database path     | `/data/whatsmeow.db`            |
| `WHATSAPP_WEBSOCKET_URL`   | API server WebSocket URL | `ws://api:3000/ws/whatsapp`     |
| `WHATSAPP_MEDIA_BASE_PATH` | Media storage directory  | `/data/media`                   |
| `WHATSAPP_MEDIA_BASE_URL`  | Public media URL         | `https://api.example.com/media` |

### Optional Variables

| Variable               | Default   | Description                       |
| ---------------------- | --------- | --------------------------------- |
| `WHATSAPP_SERVER_HOST` | `0.0.0.0` | Listen address                    |
| `WHATSAPP_SERVER_PORT` | `8080`    | Listen port                       |
| `WHATSAPP_LOG_LEVEL`   | `info`    | Log level (debug/info/warn/error) |
| `WHATSAPP_LOG_FORMAT`  | `json`    | Log format (json/text)            |

### Security Variables

| Variable                     | Default | Description              |
| ---------------------------- | ------- | ------------------------ |
| `WHATSAPP_API_KEY_ENABLED`   | `false` | Enable API key auth      |
| `WHATSAPP_API_KEYS`          | -       | Comma-separated API keys |
| `WHATSAPP_RATELIMIT_ENABLED` | `true`  | Enable rate limiting     |
| `WHATSAPP_RATELIMIT_RPS`     | `10`    | Requests per second      |

### Circuit Breaker Variables

| Variable                                     | Default | Description            |
| -------------------------------------------- | ------- | ---------------------- |
| `WHATSAPP_CIRCUIT_BREAKER_ENABLED`           | `true`  | Enable circuit breaker |
| `WHATSAPP_CIRCUIT_BREAKER_FAILURE_THRESHOLD` | `5`     | Failures to open       |
| `WHATSAPP_CIRCUIT_BREAKER_TIMEOUT`           | `30s`   | Open state duration    |

### Webhook Variables

| Variable                   | Default | Description                 |
| -------------------------- | ------- | --------------------------- |
| `WHATSAPP_WEBHOOK_ENABLED` | `false` | Enable webhooks             |
| `WHATSAPP_WEBHOOK_URL`     | -       | Webhook endpoint            |
| `WHATSAPP_WEBHOOK_SECRET`  | -       | HMAC signing secret         |
| `WHATSAPP_WEBHOOK_EVENTS`  | all     | Comma-separated event types |

---

## Deployment Options

### Option 1: Docker (Recommended)

**Dockerfile**

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY apps/server/go.mod apps/server/go.sum ./
RUN go mod download

COPY apps/server/ ./
RUN CGO_ENABLED=0 go build -o /whatsapp-service ./cmd/whatsapp

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /whatsapp-service /usr/local/bin/

RUN mkdir -p /data/media

USER nobody
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/whatsapp-service"]
```

**docker-compose.yml**

```yaml
version: "3.8"

services:
  whatsapp:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    volumes:
      - whatsapp-data:/data
    environment:
      - WHATSAPP_DB_PATH=/data/whatsmeow.db
      - WHATSAPP_MEDIA_BASE_PATH=/data/media
      - WHATSAPP_MEDIA_BASE_URL=http://localhost:8080/media
      - WHATSAPP_WEBSOCKET_URL=ws://api:3000/ws/whatsapp
      - WHATSAPP_API_KEY_ENABLED=true
      - WHATSAPP_API_KEYS=your-secret-key
      - WHATSAPP_LOG_LEVEL=info
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  whatsapp-data:
```

**Run**

```bash
docker-compose up -d
```

### Option 2: Binary Deployment

**Build**

```bash
cd apps/server
CGO_ENABLED=0 GOOS=linux go build -o whatsapp-service ./cmd/whatsapp
```

**Systemd Service** (`/etc/systemd/system/whatsapp.service`)

```ini
[Unit]
Description=Whatspire WhatsApp Service
After=network.target

[Service]
Type=simple
User=whatsapp
Group=whatsapp
WorkingDirectory=/opt/whatsapp
ExecStart=/opt/whatsapp/whatsapp-service
Restart=always
RestartSec=5

Environment=WHATSAPP_DB_PATH=/var/lib/whatsapp/whatsmeow.db
Environment=WHATSAPP_MEDIA_BASE_PATH=/var/lib/whatsapp/media
Environment=WHATSAPP_MEDIA_BASE_URL=https://api.example.com/media
Environment=WHATSAPP_WEBSOCKET_URL=ws://localhost:3000/ws/whatsapp

[Install]
WantedBy=multi-user.target
```

**Enable and Start**

```bash
sudo systemctl daemon-reload
sudo systemctl enable whatsapp
sudo systemctl start whatsapp
```

---

## Kubernetes Deployment

**deployment.yaml**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whatsapp-service
spec:
  replicas: 1 # Single instance recommended
  selector:
    matchLabels:
      app: whatsapp
  template:
    metadata:
      labels:
        app: whatsapp
    spec:
      containers:
        - name: whatsapp
          image: your-registry/whatsapp-service:latest
          ports:
            - containerPort: 8080
          env:
            - name: WHATSAPP_DB_PATH
              value: /data/whatsmeow.db
            - name: WHATSAPP_MEDIA_BASE_PATH
              value: /data/media
            - name: WHATSAPP_API_KEY_ENABLED
              value: "true"
            - name: WHATSAPP_API_KEYS
              valueFrom:
                secretKeyRef:
                  name: whatsapp-secrets
                  key: api-keys
          volumeMounts:
            - name: data
              mountPath: /data
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "512Mi"
              cpu: "500m"
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: whatsapp-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: whatsapp-service
spec:
  selector:
    app: whatsapp
  ports:
    - port: 8080
      targetPort: 8080
```

---

## Production Checklist

### Security

- [ ] Enable API key authentication
- [ ] Configure role-based API keys
- [ ] Set up HTTPS (via reverse proxy)
- [ ] Configure CORS allowed origins
- [ ] Enable webhook HMAC signing

### Reliability

- [ ] Enable circuit breaker
- [ ] Configure rate limiting
- [ ] Set up health check monitoring
- [ ] Configure log aggregation
- [ ] Enable Prometheus metrics

### Data

- [ ] Use persistent storage for `/data`
- [ ] Set up backup for SQLite database
- [ ] Configure media retention policy
- [ ] Plan storage capacity

### Monitoring

- [ ] Scrape `/metrics` with Prometheus
- [ ] Set up Grafana dashboards
- [ ] Configure alerting rules
- [ ] Monitor disk usage

---

## Backup & Recovery

### Database Backup

```bash
# Stop service (or use SQLite backup)
sqlite3 /data/whatsmeow.db ".backup /backup/whatsmeow-$(date +%Y%m%d).db"
```

### Restore

```bash
cp /backup/whatsmeow-20260203.db /data/whatsmeow.db
```

---

## Troubleshooting

### Common Issues

| Issue            | Cause              | Solution                       |
| ---------------- | ------------------ | ------------------------------ |
| QR timeout       | Phone not scanning | Increase `WHATSAPP_QR_TIMEOUT` |
| Connection drops | Network issues     | Check circuit breaker settings |
| Rate limited     | Too many requests  | Increase limits or add delay   |
| 401 Unauthorized | Invalid API key    | Check `WHATSAPP_API_KEYS`      |

### Debug Mode

```bash
export WHATSAPP_LOG_LEVEL=debug
export WHATSAPP_LOG_FORMAT=text
```

### Health Check

```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

---

## Scaling Considerations

> ⚠️ **Important**: This service maintains persistent WhatsApp connections. Running multiple replicas requires careful session management.

- Use **single replica** per WhatsApp session
- Session state is stored in SQLite (local to instance)
- For multi-session scaling, use session affinity

---

## Database Configuration

### SQLite (Development/Small Scale)

SQLite is the default database and suitable for:

- Development environments
- Small deployments (< 1000 sessions)
- Single-server setups

**Configuration:**

```bash
export WHATSAPP_DATABASE_DRIVER=sqlite
export WHATSAPP_DATABASE_DSN=/data/whatspire.db
```

**Advantages:**

- No separate database server required
- Simple setup and maintenance
- Single file for easy backups

**Limitations:**

- Limited concurrent write performance
- Not suitable for distributed deployments
- File-based storage

### PostgreSQL (Production/Large Scale)

PostgreSQL is recommended for:

- Production environments
- Large deployments (> 1000 sessions)
- High-availability setups
- Distributed architectures

**Configuration:**

```bash
export WHATSAPP_DATABASE_DRIVER=postgres
export WHATSAPP_DATABASE_DSN="host=postgres user=whatspire password=secret dbname=whatspire port=5432 sslmode=require"
```

**Docker Image:**

The service uses `pgvector/pgvector:pg18-trixie` which includes:

- PostgreSQL 18 (latest stable version)
- pgvector extension for vector similarity search
- Debian Trixie base for stability
- Future-ready for AI/ML features

**Advantages:**

- Better concurrent performance
- Advanced indexing and query optimization
- Replication and high availability
- Better monitoring and management tools
- Vector similarity search support (pgvector extension)

**Setup with Docker Compose:**

```yaml
name: whatspire

services:
  postgres:
    image: pgvector/pgvector:pg18-trixie
    environment:
      POSTGRES_USER: whatspire
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: whatspire
    volumes:
      - postgres_data:/var/lib/postgresql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U whatspire"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  whatsapp:
    build: .
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      WHATSAPP_DATABASE_DRIVER: postgres
      WHATSAPP_DATABASE_DSN: "host=postgres user=whatspire password=${DB_PASSWORD} dbname=whatspire port=5432 sslmode=disable"
      WHATSAPP_DATABASE_MAX_IDLE_CONNS: 10
      WHATSAPP_DATABASE_MAX_OPEN_CONNS: 100
      WHATSAPP_DATABASE_CONN_MAX_LIFETIME: 1h
    volumes:
      - whatsapp_data:/data
    restart: unless-stopped

volumes:
  postgres_data:
  whatsapp_data:
```

### Database Configuration Options

| Variable                              | Default | Description                       |
| ------------------------------------- | ------- | --------------------------------- |
| `WHATSAPP_DATABASE_DRIVER`            | sqlite  | Database driver (sqlite/postgres) |
| `WHATSAPP_DATABASE_DSN`               | -       | Database connection string        |
| `WHATSAPP_DATABASE_MAX_IDLE_CONNS`    | 10      | Maximum idle connections          |
| `WHATSAPP_DATABASE_MAX_OPEN_CONNS`    | 100     | Maximum open connections          |
| `WHATSAPP_DATABASE_CONN_MAX_LIFETIME` | 1h      | Connection maximum lifetime       |
| `WHATSAPP_DATABASE_LOG_LEVEL`         | warn    | GORM log level                    |

### PostgreSQL Connection String Format

**Standard format:**

```
host=localhost user=username password=secret dbname=database port=5432 sslmode=require
```

**Connection URI format:**

```
postgres://username:password@localhost:5432/database?sslmode=require
```

**SSL Modes:**

- `disable`: No SSL (development only)
- `require`: Require SSL (recommended)
- `verify-ca`: Verify CA certificate
- `verify-full`: Verify CA and hostname

### Database Migrations

Migrations run automatically on startup. See [Database Migrations Guide](database_migrations.md) for details.

**Migration logs:**

```
🔄 Running database migrations...
📊 Current migration version: 1707123456
✅ Migration recorded: version 1707123457
✅ Database migrations completed
```

### Backup and Recovery

#### SQLite Backup

```bash
# Backup
cp /data/whatspire.db /backup/whatspire-$(date +%Y%m%d).db

# Restore
cp /backup/whatspire-20240203.db /data/whatspire.db
```

#### PostgreSQL Backup

```bash
# Backup
pg_dump -h localhost -U whatspire whatspire > backup-$(date +%Y%m%d).sql

# Restore
psql -h localhost -U whatspire whatspire < backup-20240203.sql
```

### Monitoring

#### Database Health Check

The service includes database health checks:

```bash
curl http://localhost:8080/ready
```

Response:

```json
{
  "success": true,
  "data": {
    "status": "ready",
    "components": {
      "database": "healthy",
      "whatsapp_client": "healthy"
    }
  }
}
```

#### PostgreSQL Monitoring

Monitor connection pool usage:

```sql
SELECT
    count(*) as total_connections,
    count(*) FILTER (WHERE state = 'active') as active,
    count(*) FILTER (WHERE state = 'idle') as idle
FROM pg_stat_activity
WHERE datname = 'whatspire';
```

### Performance Tuning

#### PostgreSQL Configuration

For production deployments, tune PostgreSQL settings:

```ini
# postgresql.conf
max_connections = 200
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 2621kB
min_wal_size = 1GB
max_wal_size = 4GB
```

#### Connection Pool Tuning

Adjust based on your workload:

```bash
# For high-concurrency workloads
export WHATSAPP_DATABASE_MAX_OPEN_CONNS=200
export WHATSAPP_DATABASE_MAX_IDLE_CONNS=50

# For low-latency requirements
export WHATSAPP_DATABASE_CONN_MAX_LIFETIME=30m
```

---

## Migration from SQLite to PostgreSQL

### Step 1: Backup SQLite Database

```bash
cp /data/whatspire.db /backup/whatspire-migration.db
```

### Step 2: Export Data

```bash
sqlite3 /data/whatspire.db .dump > export.sql
```

### Step 3: Convert SQL (if needed)

SQLite and PostgreSQL have some syntax differences. Review and adjust:

- `AUTOINCREMENT` → `SERIAL`
- `TEXT` → `VARCHAR` or `TEXT`
- Date/time formats
- Boolean values

### Step 4: Import to PostgreSQL

```bash
psql -h localhost -U whatspire whatspire < export.sql
```

### Step 5: Update Configuration

```bash
export WHATSAPP_DATABASE_DRIVER=postgres
export WHATSAPP_DATABASE_DSN="host=localhost user=whatspire password=secret dbname=whatspire port=5432 sslmode=require"
```

### Step 6: Restart Service

```bash
docker-compose restart whatsapp
```

### Step 7: Verify

```bash
# Check logs
docker-compose logs whatsapp | grep migration

# Test API
curl http://localhost:8080/health
```

---

## See Also

- [Database Migrations Guide](database_migrations.md)
- [Configuration Guide](configuration.md)
- [API Specification](api_specification.md)
