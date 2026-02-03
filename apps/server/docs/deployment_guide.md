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
