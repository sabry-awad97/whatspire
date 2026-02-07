# Docker Deployment Guide

This guide covers building and deploying the WhatsApp Service using Docker.

## Table of Contents

- [Quick Start](#quick-start)
- [Building the Image](#building-the-image)
- [Running the Container](#running-the-container)
- [Docker Compose](#docker-compose)
- [Configuration](#configuration)
- [Production Deployment](#production-deployment)
- [Troubleshooting](#troubleshooting)

## Quick Start

The fastest way to get started:

```bash
# 1. Copy the example configuration
cp config.example.yaml config.yaml

# 2. Edit config.yaml with your settings
# (especially websocket.url and websocket.api_key)

# 3. Build and run with Docker Compose
docker-compose up -d

# 4. Check logs
docker-compose logs -f
```

## Building the Image

### Using Build Scripts

**Linux/macOS:**

```bash
chmod +x docker-build.sh
./docker-build.sh
```

**Windows (PowerShell):**

```powershell
.\docker-build.ps1
```

### Manual Build

```bash
# Enable BuildKit for optimized builds
export DOCKER_BUILDKIT=1

# Build the image
docker build -t whatsapp-service:2.0.0 -t whatsapp-service:latest .

# Check the image size
docker images whatsapp-service
```

### Build Arguments

The Dockerfile supports the following build arguments:

- `BUILDKIT_INLINE_CACHE=1` - Enable build cache for faster rebuilds

## Running the Container

### Using Docker Compose (Recommended)

```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f whatsapp-service

# Stop the service
docker-compose down

# Restart the service
docker-compose restart whatsapp-service
```

### Manual Docker Run

```bash
docker run -d \
  --name whatsapp-service \
  -p 8080:8080 \
  -v $(pwd)/data:/data \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -e GIN_MODE=release \
  -e TZ=UTC \
  --restart unless-stopped \
  whatsapp-service:latest
```

**Windows (PowerShell):**

```powershell
docker run -d `
  --name whatsapp-service `
  -p 8080:8080 `
  -v ${PWD}/data:/data `
  -v ${PWD}/config.yaml:/app/config.yaml:ro `
  -e GIN_MODE=release `
  -e TZ=UTC `
  --restart unless-stopped `
  whatsapp-service:latest
```

## Docker Compose

### Configuration

The `docker-compose.yml` file includes:

- **Port mapping**: 8080:8080
- **Volume mounts**:
  - `./data:/data` - Persistent data (databases, media)
  - `./config.yaml:/app/config.yaml:ro` - Configuration (read-only)
- **Health checks**: Automatic health monitoring
- **Resource limits**: CPU and memory constraints
- **Logging**: JSON file logging with rotation
- **Restart policy**: Automatic restart on failure

### Customization

Edit `docker-compose.yml` to customize:

```yaml
# Change port mapping
ports:
  - "9090:8080" # Host:Container

# Adjust resource limits
deploy:
  resources:
    limits:
      cpus: "4"
      memory: 2G
    reservations:
      cpus: "1"
      memory: 512M

# Add environment variables
environment:
  - LOG_LEVEL=debug
  - TZ=America/New_York
```

## Configuration

### Required Files

1. **config.yaml** - Main configuration file

   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your settings
   ```

2. **data/** - Persistent data directory
   - Created automatically if it doesn't exist
   - Contains databases and media files

### Environment Variables

| Variable   | Default   | Description                        |
| ---------- | --------- | ---------------------------------- |
| `GIN_MODE` | `release` | Gin framework mode (debug/release) |
| `TZ`       | `UTC`     | Timezone for the container         |

### Volume Mounts

| Host Path       | Container Path     | Purpose                                    |
| --------------- | ------------------ | ------------------------------------------ |
| `./data`        | `/data`            | Persistent storage for databases and media |
| `./config.yaml` | `/app/config.yaml` | Application configuration (read-only)      |

## Production Deployment

### Security Best Practices

1. **Run as non-root user** (already configured in Dockerfile)
2. **Use read-only configuration mount**
3. **Set resource limits** to prevent resource exhaustion
4. **Enable health checks** for automatic recovery
5. **Use secrets management** for sensitive data

### Example Production Setup

```yaml
version: "3.9"

services:
  whatsapp-service:
    image: whatsapp-service:2.0.0
    container_name: whatsapp-service
    restart: always

    ports:
      - "127.0.0.1:8080:8080" # Bind to localhost only

    environment:
      - GIN_MODE=release
      - TZ=UTC

    volumes:
      - /var/lib/whatsapp/data:/data
      - /etc/whatsapp/config.yaml:/app/config.yaml:ro

    deploy:
      resources:
        limits:
          cpus: "2"
          memory: 1G
        reservations:
          cpus: "0.5"
          memory: 256M

    healthcheck:
      test:
        [
          "CMD",
          "wget",
          "--no-verbose",
          "--tries=1",
          "--spider",
          "http://localhost:8080/health",
        ]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "5"

    networks:
      - whatsapp-network

networks:
  whatsapp-network:
    driver: bridge
```

### Using with Reverse Proxy

**Nginx example:**

```nginx
upstream whatsapp_backend {
    server 127.0.0.1:8080;
}

server {
    listen 443 ssl http2;
    server_name whatsapp.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://whatsapp_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Container Registry

Push to a container registry for production:

```bash
# Tag for registry
docker tag whatsapp-service:2.0.0 registry.example.com/whatsapp-service:2.0.0

# Push to registry
docker push registry.example.com/whatsapp-service:2.0.0

# Pull on production server
docker pull registry.example.com/whatsapp-service:2.0.0
```

## Troubleshooting

### Check Container Status

```bash
# View running containers
docker ps

# View all containers (including stopped)
docker ps -a

# Check container logs
docker logs whatsapp-service

# Follow logs in real-time
docker logs -f whatsapp-service

# Check last 100 lines
docker logs --tail 100 whatsapp-service
```

### Health Check

```bash
# Check health status
docker inspect whatsapp-service | grep -A 10 Health

# Manual health check
curl http://localhost:8080/health
```

### Common Issues

#### Container won't start

1. Check logs: `docker logs whatsapp-service`
2. Verify config.yaml exists and is valid
3. Ensure data directory has correct permissions
4. Check port 8080 is not already in use

#### Permission denied errors

```bash
# Fix data directory permissions
sudo chown -R 1000:1000 ./data
```

#### Out of memory

Increase memory limits in docker-compose.yml:

```yaml
deploy:
  resources:
    limits:
      memory: 2G
```

#### Database locked errors

Ensure only one instance is running:

```bash
docker ps | grep whatsapp-service
docker-compose down
docker-compose up -d
```

### Debugging

Enter the container for debugging:

```bash
# Start a shell in the running container
docker exec -it whatsapp-service sh

# Check files
ls -la /app
ls -la /data

# Check processes
ps aux

# Check network
netstat -tlnp
```

### Performance Monitoring

```bash
# View resource usage
docker stats whatsapp-service

# View detailed container info
docker inspect whatsapp-service
```

## Maintenance

### Updating the Service

```bash
# Pull latest code
git pull

# Rebuild image
docker-compose build

# Restart with new image
docker-compose up -d

# Remove old images
docker image prune -f
```

### Backup

```bash
# Backup data directory
tar -czf whatsapp-backup-$(date +%Y%m%d).tar.gz data/

# Backup configuration
cp config.yaml config.yaml.backup
```

### Cleanup

```bash
# Stop and remove containers
docker-compose down

# Remove volumes (WARNING: deletes data)
docker-compose down -v

# Remove images
docker rmi whatsapp-service:latest

# Clean up unused resources
docker system prune -a
```

## Multi-Stage Build Details

The Dockerfile uses a multi-stage build for optimization:

1. **Builder Stage** (golang:1.24-alpine)
   - Installs build dependencies
   - Downloads Go modules
   - Compiles the application with optimizations
   - Creates a static binary

2. **Runtime Stage** (alpine:3.21)
   - Minimal base image (~5MB)
   - Only runtime dependencies
   - Non-root user for security
   - Final image size: ~30-40MB

### Build Optimizations

- **Layer caching**: go.mod/go.sum copied separately
- **Static binary**: No external dependencies
- **Stripped binary**: Debug symbols removed (-ldflags "-s -w")
- **CGO enabled**: Required for SQLite support
- **BuildKit**: Faster builds with better caching

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Go Docker Best Practices](https://docs.docker.com/language/golang/)
- [Alpine Linux](https://alpinelinux.org/)
