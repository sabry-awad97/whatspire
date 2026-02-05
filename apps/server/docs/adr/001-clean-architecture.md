# ADR-001: Clean Architecture Adoption

**Date**: 2026-02-03  
**Status**: Accepted  
**Deciders**: Development Team  
**Technical Story**: Platform Improvements - Code Maintainability

---

## Context and Problem Statement

The Whatspire WhatsApp Service codebase was growing rapidly with increasing complexity. We needed an architectural pattern that would support long-term maintainability, testability, and allow multiple developers to work on different features without conflicts. How should we structure our Go application to achieve these goals?

## Decision Drivers

- Need for clear separation of concerns
- Requirement for high testability (70%+ coverage for business logic, 100% for domain)
- Support for multiple database backends (SQLite, PostgreSQL)
- Ability to swap infrastructure components without affecting business logic
- Team scalability - multiple developers working simultaneously
- Constitution requirement: no file exceeds 800 lines

## Considered Options

- **Option 1**: Clean Architecture (Hexagonal/Ports & Adapters)
- **Option 2**: Traditional Layered Architecture (MVC-style)
- **Option 3**: Modular Monolith without strict boundaries
- **Option 4**: Microservices Architecture

## Decision Outcome

Chosen option: "**Clean Architecture (Hexagonal/Ports & Adapters)**", because it provides the best balance of maintainability, testability, and flexibility while keeping the application as a monolith for operational simplicity.

### Positive Consequences

- **Domain Independence**: Business logic has zero external dependencies
- **Testability**: Domain and application layers can be tested without infrastructure
- **Flexibility**: Easy to swap databases, add new interfaces (HTTP, gRPC, CLI)
- **Team Scalability**: Clear boundaries allow parallel development
- **Code Organization**: Enforced structure prevents architectural drift
- **Migration Path**: Can evolve to microservices if needed

### Negative Consequences

- **Initial Complexity**: More files and interfaces than simpler architectures
- **Learning Curve**: Team needs to understand Clean Architecture principles
- **Boilerplate**: More code for dependency injection and interface definitions
- **Indirection**: Multiple layers can make code flow harder to trace initially

## Pros and Cons of the Options

### Option 1: Clean Architecture

- Good, because domain logic is completely isolated and testable
- Good, because infrastructure can be swapped without touching business logic
- Good, because enforces dependency rule (dependencies point inward)
- Good, because supports multiple interfaces (HTTP, WebSocket, CLI)
- Bad, because requires more upfront design and structure
- Bad, because more files and interfaces to maintain

### Option 2: Traditional Layered Architecture

- Good, because simpler and more familiar to most developers
- Good, because less boilerplate code
- Bad, because business logic often leaks into presentation layer
- Bad, because harder to test without infrastructure
- Bad, because database changes affect multiple layers

### Option 3: Modular Monolith

- Good, because flexible and pragmatic
- Good, because less ceremony than Clean Architecture
- Bad, because boundaries are not enforced by structure
- Bad, because can lead to architectural drift over time
- Bad, because harder to maintain consistency across team

### Option 4: Microservices

- Good, because ultimate flexibility and scalability
- Good, because independent deployment of services
- Bad, because massive operational complexity for our use case
- Bad, because distributed system challenges (networking, consistency)
- Bad, because overkill for current scale and team size

## Implementation Structure

```
apps/server/internal/
├── domain/              # Enterprise Business Rules
│   ├── entity/         # Business objects
│   ├── valueobject/    # Immutable value types
│   ├── repository/     # Repository interfaces
│   └── errors/         # Domain errors
├── application/         # Application Business Rules
│   ├── dto/            # Data Transfer Objects
│   └── usecase/        # Use cases (orchestration)
├── infrastructure/      # Frameworks & Drivers
│   ├── persistence/    # Database implementations
│   ├── whatsapp/       # WhatsApp client
│   ├── webhook/        # Webhook publisher
│   ├── websocket/      # WebSocket publisher
│   ├── config/         # Configuration
│   └── logger/         # Logging
└── presentation/        # Interface Adapters
    ├── http/           # HTTP handlers
    └── ws/             # WebSocket handlers
```

## Dependency Rule

Dependencies must point inward:

1. **Domain** → No dependencies (pure Go)
2. **Application** → Depends on Domain only
3. **Infrastructure** → Depends on Domain and Application
4. **Presentation** → Depends on Application and Domain

## Links

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- Related: ADR-002 (GORM for Database Abstraction)
- Related: ADR-005 (Event Persistence Strategy)

---

## Notes

### File Size Management

To comply with the 800-line limit, we split large files:

- `client.go` (1,373 lines) → 6 files (max 604 lines)
- `handler.go` (554 lines) → 6 files (max 381 lines)

### Testing Strategy

- **Domain Layer**: 100% coverage (pure business logic)
- **Application Layer**: 70%+ coverage (use cases)
- **Infrastructure Layer**: Integration tests with real dependencies
- **Presentation Layer**: HTTP handler tests with mocks

### Migration Path

If we need to scale to microservices:

1. Each use case can become a service
2. Domain entities remain shared or duplicated
3. Infrastructure adapters become service clients
4. Presentation layer becomes API gateway

This architecture provides a clear path forward without requiring a rewrite.
