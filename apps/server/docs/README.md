# Whatspire WhatsApp Service

> Enterprise-grade WhatsApp integration service built with Go and Clean Architecture

## Overview

Whatspire is a high-performance WhatsApp integration platform that provides:

- **Multi-session management** - Handle multiple WhatsApp accounts simultaneously
- **Real-time messaging** - Send/receive messages, reactions, read receipts
- **Media support** - Images, documents, audio, video
- **WebSocket events** - Real-time event streaming to clients
- **Webhook delivery** - Push events to external services
- **QR authentication** - Seamless WhatsApp pairing via QR codes

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                       │
│  ┌─────────────────┐  ┌─────────────────────────────────┐   │
│  │   HTTP API      │  │   WebSocket Handlers            │   │
│  │   (Gin Router)  │  │   (QR + Events)                 │   │
│  └────────┬────────┘  └────────────────┬────────────────┘   │
├───────────┼────────────────────────────┼────────────────────┤
│           │      Application Layer     │                    │
│  ┌────────▼────────────────────────────▼────────────────┐   │
│  │  SessionUC │ MessageUC │ GroupsUC │ ContactUC │ ...  │   │
│  └────────────────────────┬─────────────────────────────┘   │
├───────────────────────────┼─────────────────────────────────┤
│                    Domain Layer                             │
│  ┌────────────────────────▼─────────────────────────────┐   │
│  │  Entities │ Repository Interfaces │ Value Objects    │   │
│  └────────────────────────┬─────────────────────────────┘   │
├───────────────────────────┼─────────────────────────────────┤
│                 Infrastructure Layer                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌───────────┐    │
│  │ WhatsApp │  │   GORM   │  │ WebSocket│  │  Webhook  │    │
│  │ (meow)   │  │ (SQLite) │  │ Publisher│  │ Publisher │    │
│  └──────────┘  └──────────┘  └──────────┘  └───────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.24+
- SQLite (bundled via modernc.org/sqlite)

### Installation

```bash
cd apps/server
go mod download
```

### Configuration

Set environment variables (see [Configuration Guide](./configuration.md)):

```bash
export WHATSAPP_SERVER_PORT=8080
export WHATSAPP_DB_PATH=/data/whatsmeow.db
export WHATSAPP_WEBSOCKET_URL=ws://localhost:3000/ws/whatsapp
```

### Run

```bash
go run cmd/whatsapp/main.go
```

## Project Structure

```
apps/server/
├── cmd/whatsapp/          # Application entry point
├── internal/
│   ├── app/               # Uber/fx module composition
│   ├── application/       # Use cases & DTOs
│   │   ├── dto/           # Data transfer objects
│   │   └── usecase/       # Business logic
│   ├── domain/            # Core business rules
│   │   ├── entity/        # Domain entities
│   │   ├── repository/    # Interface definitions
│   │   ├── valueobject/   # Value objects
│   │   └── errors/        # Domain errors
│   ├── infrastructure/    # External adapters
│   │   ├── config/        # Configuration
│   │   ├── persistence/   # Database repositories
│   │   ├── whatsapp/      # WhatsApp client
│   │   ├── websocket/     # WebSocket server
│   │   └── webhook/       # Webhook publisher
│   └── presentation/      # HTTP/WS handlers
│       ├── http/          # REST API
│       └── ws/            # WebSocket handlers
├── pkg/                   # Shared utilities
└── test/                  # Test suites
    ├── unit/              # Unit tests
    ├── integration/       # Integration tests
    └── property/          # Property-based tests
```

## Key Features

### Message Types

| Type     | Send | Receive | Description             |
| -------- | ---- | ------- | ----------------------- |
| Text     | ✅   | ✅      | Plain text messages     |
| Image    | ✅   | ✅      | JPEG, PNG, WebP         |
| Document | ✅   | ✅      | PDF, DOC, etc.          |
| Audio    | ✅   | ✅      | Voice messages          |
| Video    | ✅   | ✅      | MP4, MKV                |
| Reaction | ✅   | ✅      | Emoji reactions         |
| Receipt  | ✅   | ✅      | Read/delivered receipts |

### Resilience Patterns

- **Circuit Breaker** - Prevents cascade failures
- **Rate Limiting** - Token bucket per IP/API key
- **Retry with Backoff** - Exponential retry for transient errors
- **Graceful Shutdown** - Clean connection termination

### Security

- **API Key Authentication** - Header-based auth
- **Role-Based Access Control** - Read/Write/Admin roles
- **HMAC Webhook Signing** - Secure webhook delivery
- **Audit Logging** - Track all API operations

## Documentation

| Document                                                        | Description                        |
| --------------------------------------------------------------- | ---------------------------------- |
| [API Specification](./api_specification.md)                     | Complete REST API reference        |
| [Deployment Guide](./deployment_guide.md)                       | Production deployment instructions |
| [Configuration](./configuration.md)                             | Environment variables reference    |
| [Database Migrations](./database_migrations.md)                 | Database setup and migrations      |
| [Event Persistence](./event_persistence.md)                     | Event storage and replay           |
| [Property Tests Known Issues](./property_tests_known_issues.md) | Known issues with property tests   |
| [Project Analysis](./project_analysis.md)                       | Architecture deep-dive             |
| [Troubleshooting](./troubleshooting.md)                         | Common issues and solutions        |

## Testing

```bash
# Run all tests (skips problematic property tests)
go test -short ./...

# Run with coverage
go test -short -cover ./...

# Run all tests including property tests
go test ./...

# Run specific test suites
go test ./test/unit/...
go test ./test/integration/...
go test ./test/property/...

# Run specific property test (bypassing skip)
go test ./test/property -run TestSessionPersistenceRoundTrip_Property2 -v
```

**Note**: Some property-based tests are skipped in short mode due to known issues with gopter's shrinking mechanism. See [Property Tests Known Issues](./property_tests_known_issues.md) for details.

## License

Proprietary - All rights reserved
