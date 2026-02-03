# Development Setup Guide

**Version**: 1.0.0  
**Last Updated**: 2026-02-03  
**Estimated Setup Time**: 2-3 hours

---

## Prerequisites

### Required Software

1. **Go 1.24+**

   ```bash
   # Verify installation
   go version
   # Should output: go version go1.24.0 or higher
   ```

   Installation:
   - macOS: `brew install go`
   - Linux: Download from [golang.org](https://golang.org/dl/)
   - Windows: Download installer from [golang.org](https://golang.org/dl/)

2. **Node.js 18+** (for frontend)

   ```bash
   # Verify installation
   node --version
   # Should output: v18.0.0 or higher
   ```

   Installation:
   - macOS: `brew install node`
   - Linux: Use [nvm](https://github.com/nvm-sh/nvm)
   - Windows: Download from [nodejs.org](https://nodejs.org/)

3. **Bun** (package manager for frontend)

   ```bash
   # Install Bun
   curl -fsSL https://bun.sh/install | bash

   # Verify installation
   bun --version
   ```

4. **Docker Desktop** (for PostgreSQL)

   ```bash
   # Verify installation
   docker --version
   # Should output: Docker version 20.0.0 or higher
   ```

   Installation:
   - Download from [docker.com](https://www.docker.com/products/docker-desktop)

5. **Git**
   ```bash
   # Verify installation
   git --version
   ```

### Optional Tools

- **golangci-lint**: For code linting

  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```

- **swag**: For OpenAPI documentation

  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```

- **air**: For hot-reload during development
  ```bash
  go install github.com/cosmtrek/air@latest
  ```

---

## Quick Start (5 minutes)

```bash
# 1. Clone repository
git clone https://github.com/your-org/whatspire.git
cd whatspire

# 2. Install backend dependencies
cd apps/server
go mod download

# 3. Install frontend dependencies
cd ../web
bun install

# 4. Start backend (SQLite)
cd ../server
go run cmd/whatsapp/main.go

# 5. Start frontend (in new terminal)
cd ../web
bun run dev
```

Access the application:

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- API Docs: http://localhost:8080/docs (if Swagger UI is configured)

---

## Detailed Setup

### 1. Repository Setup

```bash
# Clone the repository
git clone https://github.com/your-org/whatspire.git
cd whatspire

# Create feature branch
git checkout -b feature/your-feature-name
```

### 2. Backend Setup

#### Install Dependencies

```bash
cd apps/server
go mod download
go mod verify
```

#### Configuration

Create a configuration file:

```bash
# Copy example config
cp config.example.yaml config.yaml

# Edit configuration
nano config.yaml
```

Example `config.yaml`:

```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  driver: "sqlite" # or "postgres"
  dsn: "./data/whatspire.db"

events:
  retention_days: 30
  cleanup_interval: "24h"

logging:
  level: "debug"
  format: "json"
```

#### Environment Variables (Alternative)

```bash
# Create .env file
cat > .env << EOF
SERVER_PORT=8080
DATABASE_DRIVER=sqlite
DATABASE_DSN=./data/whatspire.db
LOG_LEVEL=debug
EOF

# Load environment variables
export $(cat .env | xargs)
```

#### Run Backend

```bash
# Development mode (with hot-reload if air is installed)
air

# Or run directly
go run cmd/whatsapp/main.go

# Or build and run
go build -o bin/whatspire cmd/whatsapp/main.go
./bin/whatspire
```

#### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/domain/...

# Run with verbose output
go test -v ./...

# Run benchmarks
go test -bench=. ./test/benchmark/...
```

### 3. Frontend Setup

#### Install Dependencies

```bash
cd apps/web
bun install
```

#### Configuration

Create `.env` file:

```bash
cat > .env << EOF
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
EOF
```

#### Run Frontend

```bash
# Development mode
bun run dev

# Type checking
bun run check-types

# Build for production
bun run build

# Preview production build
bun run serve
```

#### Run Desktop App

```bash
# Development mode
bun run desktop:dev

# Build for production
bun run desktop:build
```

### 4. Database Setup

#### SQLite (Default)

No setup required! SQLite database is created automatically.

```bash
# Database location
./data/whatspire.db
```

#### PostgreSQL (Production)

Start PostgreSQL with Docker:

```bash
# Start PostgreSQL
docker-compose up -d postgres

# Verify it's running
docker ps

# View logs
docker logs whatspire-postgres
```

Update configuration:

```yaml
database:
  driver: "postgres"
  dsn: "host=localhost port=5432 user=postgres password=postgres dbname=whatspire sslmode=disable"
```

Or use environment variables:

```bash
export DATABASE_DRIVER=postgres
export DATABASE_DSN="host=localhost port=5432 user=postgres password=postgres dbname=whatspire sslmode=disable"
```

#### Database Migrations

Migrations run automatically on startup using GORM AutoMigrate.

To manually run migrations:

```go
// In your code
db.AutoMigrate(
    &models.Session{},
    &models.Event{},
    &models.Presence{},
    // ...
)
```

---

## Development Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Changes

Follow the project structure:

```
apps/server/internal/
├── domain/              # Business logic (no dependencies)
├── application/         # Use cases
├── infrastructure/      # External services
└── presentation/        # HTTP/WebSocket handlers
```

### 3. Run Tests

```bash
# Backend tests
cd apps/server
go test ./...

# Frontend tests
cd apps/web
bun test
```

### 4. Lint Code

```bash
# Backend
golangci-lint run

# Frontend
bun run lint
```

### 5. Commit Changes

```bash
git add .
git commit -m "feat: add new feature"
```

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `refactor:` - Code refactoring
- `test:` - Tests
- `chore:` - Maintenance

### 6. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

---

## IDE Setup

### VS Code

Recommended extensions:

```json
{
  "recommendations": [
    "golang.go",
    "dbaeumer.vscode-eslint",
    "esbenp.prettier-vscode",
    "bradlc.vscode-tailwindcss",
    "tauri-apps.tauri-vscode"
  ]
}
```

Settings (`.vscode/settings.json`):

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  },
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  }
}
```

### GoLand / IntelliJ IDEA

1. Open project
2. Enable Go modules: Settings → Go → Go Modules
3. Set GOROOT to Go 1.24+
4. Enable golangci-lint: Settings → Tools → File Watchers

---

## Debugging

### Backend Debugging

#### VS Code

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/apps/server/cmd/whatsapp",
      "env": {
        "DATABASE_DRIVER": "sqlite",
        "LOG_LEVEL": "debug"
      }
    }
  ]
}
```

#### Delve (CLI)

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug
dlv debug cmd/whatsapp/main.go
```

### Frontend Debugging

#### Browser DevTools

1. Open http://localhost:5173
2. Press F12 to open DevTools
3. Use Console, Network, and React DevTools

#### VS Code

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Chrome",
      "type": "chrome",
      "request": "launch",
      "url": "http://localhost:5173",
      "webRoot": "${workspaceFolder}/apps/web/src"
    }
  ]
}
```

---

## Common Tasks

### Generate OpenAPI Documentation

```bash
cd apps/server
swag init -g cmd/whatsapp/main.go
```

### Run Benchmarks

```bash
go test -bench=. -benchmem ./test/benchmark/...
```

### Check Code Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Update Dependencies

```bash
# Backend
go get -u ./...
go mod tidy

# Frontend
bun update
```

### Clean Build Artifacts

```bash
# Backend
go clean
rm -rf bin/

# Frontend
rm -rf apps/web/dist/
rm -rf apps/web/node_modules/
```

---

## Troubleshooting

See [troubleshooting.md](./troubleshooting.md) for common issues and solutions.

---

## Next Steps

1. Read [Architecture Decision Records](./adr/README.md)
2. Review [API Specification](./api_specification.md)
3. Check [Contributing Guidelines](../../CONTRIBUTING.md)
4. Join team Slack/Discord channel

---

## Resources

- [Go Documentation](https://golang.org/doc/)
- [React Documentation](https://react.dev/)
- [Tauri Documentation](https://tauri.app/)
- [GORM Documentation](https://gorm.io/)
- [TanStack Documentation](https://tanstack.com/)

---

**Questions?** Contact the development team or open an issue on GitHub.
