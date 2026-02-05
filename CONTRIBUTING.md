# Contributing to Whatspire

Thank you for your interest in contributing to Whatspire! This document provides guidelines and instructions for contributing to the project.

---

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Process](#development-process)
4. [Coding Standards](#coding-standards)
5. [Testing Guidelines](#testing-guidelines)
6. [Pull Request Process](#pull-request-process)
7. [Architecture Guidelines](#architecture-guidelines)
8. [Documentation](#documentation)

---

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inspiring community for all. Please be respectful and constructive in all interactions.

### Expected Behavior

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on what is best for the community
- Show empathy towards other community members

### Unacceptable Behavior

- Harassment, discrimination, or offensive comments
- Trolling, insulting, or derogatory comments
- Public or private harassment
- Publishing others' private information

---

## Getting Started

### Prerequisites

1. Read the [Development Setup Guide](./apps/server/docs/development_setup.md)
2. Set up your development environment
3. Read the [Architecture Decision Records](./apps/server/docs/adr/README.md)
4. Familiarize yourself with the codebase structure

### Finding Issues to Work On

- Check [GitHub Issues](https://github.com/your-org/whatspire/issues)
- Look for issues labeled `good first issue` or `help wanted`
- Ask in team chat if you need help finding something to work on

---

## Development Process

### 1. Create an Issue

Before starting work, create or comment on an issue to:

- Describe the problem or feature
- Discuss the approach
- Get feedback from maintainers

### 2. Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR-USERNAME/whatspire.git
cd whatspire

# Add upstream remote
git remote add upstream https://github.com/your-org/whatspire.git
```

### 3. Create a Branch

```bash
# Update main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name
```

Branch naming conventions:

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions or changes

### 4. Make Changes

Follow the [Coding Standards](#coding-standards) and [Architecture Guidelines](#architecture-guidelines).

### 5. Test Your Changes

```bash
# Backend tests
cd apps/server
go test ./...
go test -cover ./...

# Frontend tests
cd apps/web
bun test
bun run check-types
```

### 6. Commit Your Changes

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git add .
git commit -m "feat: add new feature"
```

Commit message format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `style:` - Formatting, missing semicolons, etc.
- `refactor:` - Code refactoring
- `test:` - Adding tests
- `chore:` - Maintenance tasks

Examples:

```
feat(sessions): add QR code auto-refresh
fix(database): resolve connection pool leak
docs(api): update OpenAPI annotations
refactor(handlers): split handler into multiple files
test(events): add integration tests for event replay
```

### 7. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

---

## Coding Standards

### Go Code Style

#### General Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Maximum file size: 800 lines
- Write clear, self-documenting code

#### Naming Conventions

```go
// Good
type UserRepository interface { ... }
func GetUserByID(id string) (*User, error) { ... }
const MaxRetries = 3

// Bad
type user_repository interface { ... }
func get_user(id string) (*User, error) { ... }
const max_retries = 3
```

#### Error Handling

```go
// Good
if err != nil {
    return nil, fmt.Errorf("failed to get user: %w", err)
}

// Bad
if err != nil {
    return nil, err  // No context
}
```

#### Comments

```go
// Good
// GetUserByID retrieves a user by their unique identifier.
// Returns ErrUserNotFound if the user doesn't exist.
func GetUserByID(id string) (*User, error) { ... }

// Bad
// get user
func GetUserByID(id string) (*User, error) { ... }
```

### TypeScript/React Code Style

#### General Guidelines

- Use TypeScript for type safety
- Follow React best practices
- Use functional components with hooks
- Maximum file size: 800 lines

#### Naming Conventions

```typescript
// Good
interface UserProps { ... }
function UserCard({ user }: UserProps) { ... }
const API_BASE_URL = "http://localhost:8080";

// Bad
interface userprops { ... }
function user_card({ user }: userprops) { ... }
const api_base_url = "http://localhost:8080";
```

#### Component Structure

```typescript
// Good structure
import { useState } from "react";
import { Button } from "@/components/ui/button";

interface Props {
  user: User;
  onUpdate: (user: User) => void;
}

export function UserCard({ user, onUpdate }: Props) {
  const [isEditing, setIsEditing] = useState(false);

  // Event handlers
  const handleEdit = () => setIsEditing(true);

  // Render
  return (
    <div>
      {/* Component JSX */}
    </div>
  );
}
```

---

## Testing Guidelines

### Backend Testing

#### Unit Tests

```go
func TestUserRepository_GetByID(t *testing.T) {
    // Arrange
    db := setupTestDB(t)
    repo := NewUserRepository(db)
    user := &User{ID: "123", Name: "Test"}
    repo.Save(user)

    // Act
    result, err := repo.GetByID("123")

    // Assert
    require.NoError(t, err)
    assert.Equal(t, user.ID, result.ID)
    assert.Equal(t, user.Name, result.Name)
}
```

#### Integration Tests

```go
func TestSessionAPI_CreateSession(t *testing.T) {
    // Setup
    app := setupTestApp(t)
    defer app.Cleanup()

    // Test
    resp := app.POST("/api/sessions", map[string]string{
        "session_id": "test-session",
    })

    // Assert
    assert.Equal(t, 201, resp.StatusCode)
}
```

#### Test Coverage Requirements

- Domain layer: 100% coverage
- Application layer: 70%+ coverage
- Infrastructure layer: Integration tests
- Presentation layer: Handler tests

### Frontend Testing

```typescript
import { render, screen } from "@testing-library/react";
import { UserCard } from "./user-card";

describe("UserCard", () => {
  it("renders user name", () => {
    const user = { id: "1", name: "Test User" };
    render(<UserCard user={user} />);
    expect(screen.getByText("Test User")).toBeInTheDocument();
  });
});
```

---

## Pull Request Process

### Before Submitting

1. **Update Documentation**
   - Update README if needed
   - Add/update ADRs for architectural changes
   - Update API documentation

2. **Run All Tests**

   ```bash
   # Backend
   go test ./...
   golangci-lint run

   # Frontend
   bun test
   bun run check-types
   ```

3. **Update CHANGELOG** (if applicable)

### PR Description Template

```markdown
## Description

Brief description of changes

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Related Issues

Fixes #123

## Testing

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist

- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests pass locally
- [ ] Dependent changes merged
```

### Review Process

1. **Automated Checks**
   - CI/CD pipeline runs tests
   - Linters check code style
   - Coverage reports generated

2. **Code Review**
   - At least one approval required
   - Address all review comments
   - Keep discussions constructive

3. **Merge**
   - Squash and merge (default)
   - Rebase and merge (for clean history)
   - Delete branch after merge

---

## Architecture Guidelines

### Clean Architecture Principles

Follow the dependency rule:

```
Domain ← Application ← Infrastructure
                    ← Presentation
```

- **Domain**: No external dependencies
- **Application**: Depends on Domain only
- **Infrastructure**: Implements Domain interfaces
- **Presentation**: Orchestrates Application use cases

### File Organization

```
apps/server/internal/
├── domain/              # Business logic
│   ├── entity/         # Business objects
│   ├── valueobject/    # Immutable values
│   ├── repository/     # Repository interfaces
│   └── errors/         # Domain errors
├── application/         # Use cases
│   ├── dto/            # Data transfer objects
│   └── usecase/        # Use case implementations
├── infrastructure/      # External services
│   ├── persistence/    # Database
│   ├── whatsapp/       # WhatsApp client
│   └── config/         # Configuration
└── presentation/        # API layer
    ├── http/           # HTTP handlers
    └── ws/             # WebSocket handlers
```

### Adding New Features

1. **Define Domain Entities** (if needed)
2. **Create Repository Interface** (in domain)
3. **Implement Use Case** (in application)
4. **Implement Repository** (in infrastructure)
5. **Add HTTP Handler** (in presentation)
6. **Write Tests** (all layers)
7. **Update Documentation**

---

## Documentation

### Code Documentation

- Add godoc comments for public functions
- Document complex algorithms
- Explain non-obvious decisions
- Link to relevant ADRs

### API Documentation

- Add OpenAPI annotations to handlers
- Keep examples up-to-date
- Document error responses

### ADRs

Create an ADR for:

- Architectural decisions
- Technology choices
- Design patterns
- Breaking changes

---

## Questions?

- Check [Troubleshooting Guide](./apps/server/docs/troubleshooting.md)
- Ask in team chat
- Create a GitHub issue
- Email: dev@whatspire.com

---

**Thank you for contributing to Whatspire!** 🎉
