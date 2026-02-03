# Whatspire WhatsApp Service

A production-ready WhatsApp Business API service with a modern desktop application for managing WhatsApp sessions, messages, contacts, and groups.

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Node Version](https://img.shields.io/badge/Node-18+-339933?style=flat&logo=node.js)](https://nodejs.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Features

### Backend (Go)

- **Multi-Session Support**: Manage multiple WhatsApp accounts simultaneously
- **Clean Architecture**: Domain-driven design with clear separation of concerns
- **Database Flexibility**: Support for SQLite (development) and PostgreSQL (production)
- **Event Persistence**: Query and replay events for debugging
- **Real-Time Updates**: WebSocket support for live updates
- **Webhook Integration**: Push events to external services
- **Configuration Management**: YAML/JSON config with hot-reload
- **API Documentation**: OpenAPI/Swagger documentation
- **High Performance**: <100ms query latency, optimized for production

### Desktop Application (React + Tauri)

- **Cross-Platform**: Windows, macOS, and Linux support
- **Modern UI**: Glassmorphic design with OKLCH colors
- **Session Management**: Create, manage, and monitor WhatsApp sessions
- **QR Code Scanning**: Real-time QR code display and updates
- **Message Viewer**: View and filter messages with real-time updates
- **Contact Management**: Manage contacts with profile pictures
- **Group Management**: View and manage WhatsApp groups
- **Settings**: Configure API, theme, and notifications
- **Small Bundle**: ~10MB (vs 100MB+ for Electron)

---

## Quick Start

### Prerequisites

- Go 1.24+
- Node.js 18+
- Bun (package manager)
- Docker Desktop (for PostgreSQL)

### Installation

```bash
# Clone repository
git clone https://github.com/your-org/whatspire.git
cd whatspire

# Backend setup
cd apps/server
go mod download
go run cmd/whatsapp/main.go

# Frontend setup (in new terminal)
cd apps/web
bun install
bun run dev
```

Access the application:

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- API Docs: http://localhost:8080/docs

---

## Documentation

### Getting Started

- [Development Setup Guide](./apps/server/docs/development_setup.md) - Complete setup instructions
- [Troubleshooting Guide](./apps/server/docs/troubleshooting.md) - Common issues and solutions
- [Contributing Guidelines](./CONTRIBUTING.md) - How to contribute

### Architecture

- [Architecture Decision Records](./apps/server/docs/adr/README.md) - Key architectural decisions
  - [ADR-001: Clean Architecture](./apps/server/docs/adr/001-clean-architecture.md)
  - [ADR-002: GORM for Database](./apps/server/docs/adr/002-gorm-database-abstraction.md)
  - [ADR-003: Viper Configuration](./apps/server/docs/adr/003-viper-configuration.md)
  - [ADR-004: React + Tauri Desktop](./apps/server/docs/adr/004-react-tauri-desktop.md)
  - [ADR-005: Event Persistence](./apps/server/docs/adr/005-event-persistence-strategy.md)

### API Documentation

- [API Specification](./apps/server/docs/api_specification.md) - Complete API reference
- [Configuration Guide](./apps/server/docs/configuration.md) - Configuration options
- [Database Migrations](./apps/server/docs/database_migrations.md) - Migration guide
- [Event Persistence](./apps/server/docs/event_persistence.md) - Event system documentation
- [Deployment Guide](./apps/server/docs/deployment_guide.md) - Production deployment

### Frontend Documentation

- [Phase 7 Summary](./apps/web/PHASE7_SUMMARY.md) - Complete implementation details
- [Verification Checklist](./apps/web/VERIFICATION_CHECKLIST.md) - Quality assurance
- [Quick Start Guide](./apps/web/QUICK_START.md) - Frontend quick reference

---

## Project Structure

```
whatspire/
├── apps/
│   ├── server/              # Go backend
│   │   ├── cmd/            # Application entry points
│   │   ├── internal/       # Private application code
│   │   │   ├── domain/    # Business logic (no dependencies)
│   │   │   ├── application/ # Use cases
│   │   │   ├── infrastructure/ # External services
│   │   │   └── presentation/ # HTTP/WebSocket handlers
│   │   ├── docs/          # Documentation
│   │   └── test/          # Tests
│   └── web/               # React + Tauri desktop app
│       ├── src/
│       │   ├── components/ # UI components
│       │   ├── routes/    # TanStack Router routes
│       │   ├── lib/       # API client, WebSocket
│       │   └── stores/    # Zustand state management
│       └── src-tauri/     # Tauri backend
├── docs/                  # Project documentation
└── docker-compose.yml     # Docker services
```

---

## Technology Stack

### Backend

- **Language**: Go 1.24+
- **Framework**: Gin (HTTP), Gorilla WebSocket
- **ORM**: GORM
- **Database**: SQLite, PostgreSQL
- **Configuration**: Viper
- **Testing**: Go testing, testify
- **Documentation**: Swagger/OpenAPI

### Frontend

- **Framework**: React 19
- **Desktop**: Tauri 2.9
- **Routing**: TanStack Router
- **Data Fetching**: TanStack Query
- **Forms**: TanStack Form
- **State**: Zustand
- **Styling**: Tailwind CSS 4 (OKLCH colors)
- **UI Components**: shadcn UI
- **Build Tool**: Vite 7

---

## Development

### Backend Development

```bash
cd apps/server

# Run with hot-reload
air

# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Run linter
golangci-lint run

# Generate OpenAPI docs
swag init -g cmd/whatsapp/main.go
```

### Frontend Development

```bash
cd apps/web

# Development server
bun run dev

# Type checking
bun run check-types

# Build for production
bun run build

# Desktop app development
bun run desktop:dev

# Build desktop app
bun run desktop:build
```

---

## Configuration

### Environment Variables

```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database
DATABASE_DRIVER=sqlite  # or postgres
DATABASE_DSN=./data/whatspire.db

# Events
EVENTS_RETENTION_DAYS=30
EVENTS_CLEANUP_INTERVAL=24h

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Configuration File

Create `config.yaml`:

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

logging:
  level: "info"
  format: "json"
```

---

## Deployment

### Docker Compose (Recommended)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Production Deployment

See [Deployment Guide](./apps/server/docs/deployment_guide.md) for detailed instructions.

---

## Testing

### Backend Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/domain/...

# Run benchmarks
go test -bench=. ./test/benchmark/...
```

### Frontend Tests

```bash
# Run tests
bun test

# Type checking
bun run check-types

# E2E tests (if configured)
bun run test:e2e
```

---

## Performance

### Backend Performance

- Event GetByID: 0.016ms (target: <100ms) ✅
- Event List (100 records): 1.04ms (target: <100ms) ✅
- Session queries: <5ms average ✅

### Frontend Performance

- Bundle size: ~10MB (Tauri) vs 100MB+ (Electron)
- Startup time: <1 second
- Memory usage: ~50MB base + app memory

---

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

### Development Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Write/update tests
5. Submit a pull request

### Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` and `golangci-lint`
- Maximum file size: 800 lines
- Write clear, self-documenting code

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Support

- **Documentation**: Check the [docs](./apps/server/docs/) directory
- **Issues**: [GitHub Issues](https://github.com/your-org/whatspire/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/whatspire/discussions)
- **Email**: dev@whatspire.com

---

## Roadmap

### Completed ✅

- Clean Architecture implementation
- Multi-database support (SQLite, PostgreSQL)
- Event persistence and replay
- Configuration management with hot-reload
- Desktop application with glassmorphic UI
- Session management
- Message viewer
- Contact and group management
- Settings and preferences

### In Progress 🚧

- OpenAPI documentation
- Swagger UI integration
- Developer onboarding improvements

### Planned 📋

- Mobile app (iOS, Android)
- Message sending UI
- File upload/download
- Voice/video call UI
- Group creation/management
- Auto-updates for desktop app
- Crash reporting and analytics

---

## Acknowledgments

- [whatsmeow](https://github.com/tulir/whatsmeow) - WhatsApp Web API library
- [Tauri](https://tauri.app/) - Desktop application framework
- [shadcn/ui](https://ui.shadcn.com/) - UI component library
- [TanStack](https://tanstack.com/) - React utilities

---

## Project Status

**Status**: Active Development  
**Version**: 2.0.0  
**Last Updated**: 2026-02-03

---

**Built with ❤️ by the Whatspire Team**
