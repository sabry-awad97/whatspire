# Production Deployment Guide

This guide covers deploying the WhatsApp Service to production using Docker Compose.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Deployment](#deployment)
- [Security Hardening](#security-hardening)
- [Monitoring](#monitoring)
- [Backup and Recovery](#backup-and-recovery)
- [Maintenance](#maintenance)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

- **OS**: Linux (Ubuntu 20.04+, Debian 11+, CentOS 8+, or similar)
- **CPU**: 2+ cores recommended
- **RAM**: 4GB minimum, 8GB recommended
- **Disk**: 20GB+ available space
- **Docker**: 20.10+ with BuildKit support
- **Docker Compose**: 2.0+

### Software Installation

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version
```

### Network Requirements

- Port 80 (HTTP) - if using reverse proxy
- Port 443 (HTTPS) - if using reverse proxy with SSL
- Port 8080 (Application) - internal only, bind to localhost
- Port 5432 (PostgreSQL) - internal only, bind to localhost

## Quick Start

### 1. Clone and Setup

```bash
# Clone the repository
git clone <repository-url>
cd whatspire

# Create data directories
mkdir -p data/postgres data/whatsapp

# Set proper permissions
sudo chown -R 1000:1000 data/
```

### 2. Configure Environment

```bash
# Copy environment template
cp .env.production.example .env.production

# Edit with your settings
nano .env.production
```

**Required settings:**

- `POSTGRES_PASSWORD` - Strong database password
- `TZ` - Your timezone (e.g., America/New_York)

### 3. Configure Application

```bash
# Copy and edit application config
cp apps/server/config.example.yaml apps/server/config.yaml
nano apps/server/config.yaml
```

**Key settings to configure:**

- `server.host` and `server.port`
- `database.dsn` - Database connection string
- `websocket.url` and `websocket.api_key`
- `log.level` - Set to "info" or "warn" for production
- `apikey.enabled` - Enable API key authentication

### 4. Deploy

```bash
# Build and start services
docker-compose -f docker-compose.production.yml --env-file .env.production up -d

# Check status
docker-compose -f docker-compose.production.yml ps

# View logs
docker-compose -f docker-compose.production.yml logs -f
```

## Configuration

### Environment Variables

Create `.env.production` from the template:

```bash
cp .env.production.example .env.production
```

**Critical variables:**

| Variable            | Description                  | Example             |
| ------------------- | ---------------------------- | ------------------- |
| `POSTGRES_PASSWORD` | Database password (required) | `MyStr0ngP@ssw0rd!` |
| `POSTGRES_USER`     | Database username            | `whatspire`         |
| `POSTGRES_DB`       | Database name                | `whatspire_prod`    |
| `TZ`                | Timezone                     | `UTC`               |
| `VERSION`           | Application version          | `2.0.0`             |
| `DATA_PATH`         | Data directory path          | `./data`            |

### Application Configuration

Edit `apps/server/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "postgres" # Use PostgreSQL in production
  dsn: "postgresql://whatspire:password@postgres:5432/whatspire_prod?sslmode=disable"
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: "1h"

log:
  level: "info" # Use "info" or "warn" in production
  format: "json" # JSON format for log aggregation

apikey:
  enabled: true # Enable API key authentication

ratelimit:
  enabled: true
  requests_per_second: 10.0
  burst_size: 20

circuitbreaker:
  enabled: true
  max_requests: 3
  interval: "60s"
  timeout: "30s"
```

### Volume Configuration

The production compose file uses bind mounts for data persistence:

```yaml
volumes:
  postgres_data:
    driver_opts:
      device: ./data/postgres

  whatsapp_data:
    driver_opts:
      device: ./data/whatsapp
```

**Recommended production setup:**

```bash
# Create dedicated data directory
sudo mkdir -p /var/lib/whatspire/{postgres,whatsapp}
sudo chown -R 1000:1000 /var/lib/whatspire

# Update DATA_PATH in .env.production
DATA_PATH=/var/lib/whatspire
```

## Deployment

### Using Docker Compose

```bash
# Start services
docker-compose -f docker-compose.production.yml --env-file .env.production up -d

# Stop services
docker-compose -f docker-compose.production.yml down

# Restart services
docker-compose -f docker-compose.production.yml restart

# View logs
docker-compose -f docker-compose.production.yml logs -f whatsapp-service

# Check health
docker-compose -f docker-compose.production.yml ps
```

### Using Task (Recommended)

Add production tasks to `Taskfile.yml`:

```yaml
prod:deploy:
  desc: "Deploy to production"
  cmds:
    - docker-compose -f docker-compose.production.yml --env-file .env.production up -d
    - docker-compose -f docker-compose.production.yml ps

prod:logs:
  desc: "View production logs"
  cmds:
    - docker-compose -f docker-compose.production.yml logs -f

prod:restart:
  desc: "Restart production services"
  cmds:
    - docker-compose -f docker-compose.production.yml restart
```

Then use:

```bash
task prod:deploy
task prod:logs
task prod:restart
```

### Systemd Service (Optional)

Create `/etc/systemd/system/whatspire.service`:

```ini
[Unit]
Description=Whatspire WhatsApp Service
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/whatspire
ExecStart=/usr/local/bin/docker-compose -f docker-compose.production.yml --env-file .env.production up -d
ExecStop=/usr/local/bin/docker-compose -f docker-compose.production.yml down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable whatspire
sudo systemctl start whatspire
sudo systemctl status whatspire
```

## Security Hardening

### 1. Network Security

**Bind services to localhost only:**

```yaml
ports:
  - "127.0.0.1:8080:8080" # Application
  - "127.0.0.1:5432:5432" # PostgreSQL
```

**Use reverse proxy for external access** (see Nginx section below).

### 2. Secrets Management

**Never commit secrets to git:**

```bash
# Add to .gitignore
echo ".env.production" >> .gitignore
echo "config.yaml" >> .gitignore
```

**Use Docker secrets (Swarm mode):**

```yaml
secrets:
  postgres_password:
    external: true

services:
  postgres:
    secrets:
      - postgres_password
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
```

### 3. Container Security

The production compose file includes:

- **Non-root user**: Containers run as user 1000
- **Read-only root filesystem**: Where possible
- **No new privileges**: `security_opt: no-new-privileges:true`
- **Resource limits**: CPU and memory constraints
- **Network isolation**: Separate backend and frontend networks

### 4. SSL/TLS Configuration

#### Using Nginx Reverse Proxy

Create `config/nginx/conf.d/whatspire.conf`:

```nginx
upstream whatsapp_backend {
    server whatsapp-service:8080;
    keepalive 32;
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL certificates
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Proxy settings
    location / {
        proxy_pass http://whatsapp_backend;
        proxy_http_version 1.1;

        # WebSocket support
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # Headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://whatsapp_backend/health;
        access_log off;
    }
}
```

Enable Nginx in `docker-compose.production.yml` by uncommenting the nginx service.

#### Using Let's Encrypt

```bash
# Install certbot
sudo apt-get install certbot

# Obtain certificate
sudo certbot certonly --standalone -d your-domain.com

# Copy certificates
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem config/ssl/cert.pem
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem config/ssl/key.pem

# Set up auto-renewal
sudo certbot renew --dry-run
```

### 5. Firewall Configuration

```bash
# Allow SSH
sudo ufw allow 22/tcp

# Allow HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable firewall
sudo ufw enable

# Check status
sudo ufw status
```

## Monitoring

### Health Checks

Built-in health checks monitor service status:

```bash
# Check container health
docker inspect --format='{{.State.Health.Status}}' whatspire-whatsapp-prod
docker inspect --format='{{.State.Health.Status}}' whatspire-postgres-prod

# Manual health check
curl http://localhost:8080/health
```

### Prometheus Metrics

Enable Prometheus in `docker-compose.production.yml`:

1. Uncomment the `prometheus` service
2. Create `config/prometheus/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: "whatsapp-service"
    static_configs:
      - targets: ["whatsapp-service:8080"]
    metrics_path: "/metrics"
```

3. Access Prometheus at `http://localhost:9090`

### Grafana Dashboards

Enable Grafana in `docker-compose.production.yml`:

1. Uncomment the `grafana` service
2. Set `GRAFANA_ADMIN_PASSWORD` in `.env.production`
3. Access Grafana at `http://localhost:3000`
4. Add Prometheus as data source
5. Import dashboards for Go applications

### Log Aggregation

**Using Docker logging driver:**

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "5"
    compress: "true"
```

**View logs:**

```bash
# All services
docker-compose -f docker-compose.production.yml logs -f

# Specific service
docker-compose -f docker-compose.production.yml logs -f whatsapp-service

# Last 100 lines
docker-compose -f docker-compose.production.yml logs --tail=100 whatsapp-service
```

**Using external log aggregation (ELK, Loki, etc.):**

Update logging driver in compose file:

```yaml
logging:
  driver: "syslog"
  options:
    syslog-address: "tcp://logserver:514"
    tag: "{{.Name}}"
```

## Backup and Recovery

### Automated Backups

Create backup script `/opt/whatspire/backup.sh`:

```bash
#!/bin/bash

BACKUP_DIR="/backups/whatspire"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup PostgreSQL
docker exec whatspire-postgres-prod pg_dump -U whatspire whatspire_prod | gzip > $BACKUP_DIR/postgres_$DATE.sql.gz

# Backup WhatsApp data
tar -czf $BACKUP_DIR/whatsapp_$DATE.tar.gz /var/lib/whatspire/whatsapp

# Backup configuration
tar -czf $BACKUP_DIR/config_$DATE.tar.gz /opt/whatspire/apps/server/config.yaml

# Remove old backups
find $BACKUP_DIR -name "*.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup completed: $DATE"
```

**Schedule with cron:**

```bash
# Make executable
chmod +x /opt/whatspire/backup.sh

# Add to crontab (daily at 2 AM)
crontab -e
0 2 * * * /opt/whatspire/backup.sh >> /var/log/whatspire-backup.log 2>&1
```

### Manual Backup

```bash
# Stop services
docker-compose -f docker-compose.production.yml down

# Backup data
tar -czf whatspire-backup-$(date +%Y%m%d).tar.gz data/ apps/server/config.yaml .env.production

# Restart services
docker-compose -f docker-compose.production.yml up -d
```

### Restore from Backup

```bash
# Stop services
docker-compose -f docker-compose.production.yml down

# Restore data
tar -xzf whatspire-backup-20240207.tar.gz

# Restore PostgreSQL
gunzip < postgres_20240207.sql.gz | docker exec -i whatspire-postgres-prod psql -U whatspire whatspire_prod

# Restart services
docker-compose -f docker-compose.production.yml up -d
```

## Maintenance

### Updates

```bash
# Pull latest code
git pull

# Rebuild images
docker-compose -f docker-compose.production.yml build --no-cache

# Restart with new images
docker-compose -f docker-compose.production.yml up -d

# Remove old images
docker image prune -f
```

### Database Maintenance

```bash
# Vacuum database
docker exec whatspire-postgres-prod psql -U whatspire -d whatspire_prod -c "VACUUM ANALYZE;"

# Check database size
docker exec whatspire-postgres-prod psql -U whatspire -d whatspire_prod -c "SELECT pg_size_pretty(pg_database_size('whatspire_prod'));"
```

### Log Rotation

Logs are automatically rotated by Docker:

```yaml
logging:
  options:
    max-size: "10m"
    max-file: "5"
```

Manual cleanup:

```bash
# Clean up old logs
docker system prune -f

# Truncate logs
truncate -s 0 $(docker inspect --format='{{.LogPath}}' whatspire-whatsapp-prod)
```

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker-compose -f docker-compose.production.yml logs whatsapp-service

# Check container status
docker-compose -f docker-compose.production.yml ps

# Inspect container
docker inspect whatspire-whatsapp-prod
```

### Database Connection Issues

```bash
# Test database connection
docker exec whatspire-postgres-prod psql -U whatspire -d whatspire_prod -c "SELECT 1;"

# Check database logs
docker-compose -f docker-compose.production.yml logs postgres

# Verify network connectivity
docker exec whatspire-whatsapp-prod ping postgres
```

### Performance Issues

```bash
# Check resource usage
docker stats

# Check container limits
docker inspect whatspire-whatsapp-prod | grep -A 10 Resources

# Increase limits in docker-compose.production.yml
deploy:
  resources:
    limits:
      cpus: '4'
      memory: 2G
```

### Out of Disk Space

```bash
# Check disk usage
df -h

# Clean up Docker resources
docker system prune -a -f --volumes

# Check volume sizes
docker system df -v
```

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Production Guide](https://docs.docker.com/compose/production/)
- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Nginx Security Best Practices](https://nginx.org/en/docs/http/ngx_http_ssl_module.html)
