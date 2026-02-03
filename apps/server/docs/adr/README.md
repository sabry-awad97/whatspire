# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records (ADRs) for the Whatspire WhatsApp Service platform. ADRs document significant architectural decisions made during the development of the project.

---

## What is an ADR?

An Architecture Decision Record (ADR) is a document that captures an important architectural decision made along with its context and consequences. ADRs help teams:

- Understand why decisions were made
- Onboard new team members faster
- Avoid revisiting settled decisions
- Learn from past decisions
- Maintain architectural consistency

---

## ADR Format

Each ADR follows a consistent format (see `template.md`):

1. **Title**: Short, descriptive title
2. **Status**: Proposed, Accepted, Deprecated, or Superseded
3. **Context**: Problem statement and background
4. **Decision Drivers**: Factors influencing the decision
5. **Considered Options**: Alternatives that were evaluated
6. **Decision Outcome**: Chosen option and justification
7. **Consequences**: Positive and negative impacts
8. **Pros and Cons**: Detailed analysis of each option

---

## Index of ADRs

### ADR-001: Clean Architecture Adoption

**Status**: Accepted  
**Date**: 2026-02-03

Adopted Clean Architecture (Hexagonal/Ports & Adapters) pattern to achieve:

- Clear separation of concerns
- High testability (70%+ business logic, 100% domain)
- Support for multiple databases
- Team scalability

**Key Decisions**:

- Domain layer has zero external dependencies
- Application layer orchestrates use cases
- Infrastructure layer implements adapters
- Presentation layer handles HTTP/WebSocket

**Impact**: Reduced largest file from 1,373 to 604 lines, improved maintainability

---

### ADR-002: GORM for Database Abstraction

**Status**: Accepted  
**Date**: 2026-02-03

Selected GORM as the ORM to support:

- Multiple databases (SQLite, PostgreSQL)
- Type-safe query building
- Auto-migrations
- Good performance (<100ms p95)

**Key Decisions**:

- Use dialector pattern for database abstraction
- AutoMigrate for schema management
- Optimize with indexes and preloading

**Impact**: Single codebase works with both SQLite and PostgreSQL

---

### ADR-003: Viper for Configuration Management

**Status**: Accepted  
**Date**: 2026-02-03

Chose Viper for configuration management to enable:

- Multiple configuration sources (env, files, defaults)
- Hot-reload of non-critical settings
- YAML and JSON support
- Clear precedence rules

**Key Decisions**:

- Configuration precedence: env > file > defaults
- Hot-reload for non-critical settings only
- Validation before applying changes

**Impact**: Flexible configuration without service restarts

---

### ADR-004: React + Tauri for Desktop Application

**Status**: Accepted  
**Date**: 2026-02-03

Selected React + Tauri for desktop app to achieve:

- Cross-platform support (Windows, macOS, Linux)
- Small bundle size (~10MB vs 100MB+ for Electron)
- Modern UI with glassmorphic design
- Native system integration

**Key Decisions**:

- React 19 for UI framework
- TanStack Router/Query/Form for data management
- Zustand for state management
- OKLCH colors for glassmorphic theme

**Impact**: Fast, lightweight desktop app with modern UI

---

### ADR-005: Event Persistence Strategy

**Status**: Accepted  
**Date**: 2026-02-03

Implemented database persistence for events to support:

- Event query and filtering (<100ms p95)
- Event replay for debugging
- Retention policy (30 days default)
- Audit trail

**Key Decisions**:

- Store events in database (not event sourcing)
- Async persistence (non-blocking)
- Automated cleanup with retention policy
- Composite indexes for query performance

**Impact**: Support engineers can debug issues by querying/replaying events

---

## Creating a New ADR

1. Copy `template.md` to a new file: `NNN-short-title.md`
2. Fill in all sections
3. Submit for review
4. Update this README with a summary

---

## ADR Lifecycle

### Proposed

Initial draft, under discussion

### Accepted

Decision has been made and implemented

### Deprecated

No longer relevant, but kept for historical reference

### Superseded

Replaced by a newer ADR (link to the new one)

---

## Best Practices

1. **Write ADRs Early**: Document decisions as they're made
2. **Keep It Concise**: Focus on the decision, not implementation details
3. **Include Context**: Explain why the decision was needed
4. **List Alternatives**: Show what options were considered
5. **Document Consequences**: Both positive and negative
6. **Link Related ADRs**: Show relationships between decisions
7. **Update Status**: Mark as deprecated/superseded when appropriate

---

## References

- [ADR GitHub Organization](https://adr.github.io/)
- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [Architecture Decision Records (ADRs)](https://github.com/joelparkerhenderson/architecture-decision-record)

---

**Last Updated**: 2026-02-03  
**Total ADRs**: 5 (all accepted)
