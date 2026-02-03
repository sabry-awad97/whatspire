# Troubleshooting Guide

**Version**: 1.0.0  
**Last Updated**: 2026-02-03

This guide covers the top 10 most common issues developers encounter and their solutions.

---

## Table of Contents

1. [Go Module Issues](#1-go-module-issues)
2. [Database Connection Errors](#2-database-connection-errors)
3. [Port Already in Use](#3-port-already-in-use)
4. [Frontend Build Failures](#4-frontend-build-failures)
5. [WebSocket Connection Issues](#5-websocket-connection-issues)
6. [CORS Errors](#6-cors-errors)
7. [Test Failures](#7-test-failures)
8. [Docker Issues](#8-docker-issues)
9. [Hot Reload Not Working](#9-hot-reload-not-working)
10. [Performance Issues](#10-performance-issues)

---

## 1. Go Module Issues

### Problem: `go: module not found` or `cannot find package`

**Symptoms**:

```
go: module github.com/example/package@latest: reading ...: 404 Not Found
```

**Solutions**:

#### Solution A: Update Go modules

```bash
cd apps/server
go mod download
go mod tidy
go mod verify
```

#### Solution B: Clear module cache

```bash
go clean -modcache
go mod download
```

#### Solution C: Check Go version

```bash
go version
# Should be 1.24.0 or higher
```

#### Solution D: Set GOPRIVATE for private repos

```bash
export GOPRIVATE=github.com/your-org/*
```

**Prevention**:

- Always run `go mod tidy` after adding dependencies
- Commit `go.mod` and `go.sum` to version control

---

## 2. Database Connection Errors

### Problem A: SQLite database locked

**Symptoms**:

```
database is locked
```

**Solutions**:

#### Solution 1: Close other connections

```bash
# Find processes using the database
lsof | grep whatspire.db

# Kill the process
kill -9 <PID>
```

#### Solution 2: Enable WAL mode

```go
db, err := gorm.Open(sqlite.Open("whatspire.db?_journal_mode=WAL"), &gorm.Config{})
```

#### Solution 3: Increase timeout

```go
db, err := gorm.Open(sqlite.Open("whatspire.db?_timeout=5000"), &gorm.Config{})
```

### Problem B: PostgreSQL connection refused

**Symptoms**:

```
dial tcp 127.0.0.1:5432: connect: connection refused
```

**Solutions**:

#### Solution 1: Check if PostgreSQL is running

```bash
docker ps | grep postgres
# If not running:
docker-compose up -d postgres
```

#### Solution 2: Check connection string

```yaml
database:
  driver: "postgres"
  dsn: "host=localhost port=5432 user=postgres password=postgres dbname=whatspire sslmode=disable"
```

#### Solution 3: Check Docker network

```bash
docker network ls
docker network inspect whatspire_default
```

#### Solution 4: Use host.docker.internal (Docker Desktop)

```yaml
database:
  dsn: "host=host.docker.internal port=5432 ..."
```

**Prevention**:

- Always start Docker services before running the app
- Use health checks in docker-compose.yml

---

## 3. Port Already in Use

### Problem: `bind: address already in use`

**Symptoms**:

```
listen tcp :8080: bind: address already in use
```

**Solutions**:

#### Solution 1: Find and kill the process

```bash
# macOS/Linux
lsof -i :8080
kill -9 <PID>

# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

#### Solution 2: Use a different port

```yaml
server:
  port: 8081
```

Or with environment variable:

```bash
export SERVER_PORT=8081
```

#### Solution 3: Stop all Go processes

```bash
killall go
```

**Prevention**:

- Always stop the server before restarting
- Use process managers like `air` for hot-reload

---

## 4. Frontend Build Failures

### Problem A: `Cannot find module` errors

**Symptoms**:

```
Error: Cannot find module '@tanstack/react-query'
```

**Solutions**:

#### Solution 1: Reinstall dependencies

```bash
cd apps/web
rm -rf node_modules bun.lockb
bun install
```

#### Solution 2: Clear Bun cache

```bash
bun pm cache rm
bun install
```

#### Solution 3: Check Node version

```bash
node --version
# Should be v18.0.0 or higher
```

### Problem B: TypeScript errors

**Symptoms**:

```
TS2307: Cannot find module './types' or its corresponding type declarations
```

**Solutions**:

#### Solution 1: Run type checking

```bash
bun run check-types
```

#### Solution 2: Restart TypeScript server (VS Code)

- Press `Cmd+Shift+P` (macOS) or `Ctrl+Shift+P` (Windows/Linux)
- Type "TypeScript: Restart TS Server"

#### Solution 3: Check tsconfig.json

```json
{
  "compilerOptions": {
    "moduleResolution": "bundler",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

**Prevention**:

- Run `bun run check-types` before committing
- Use IDE with TypeScript support

---

## 5. WebSocket Connection Issues

### Problem: WebSocket connection fails

**Symptoms**:

```
WebSocket connection to 'ws://localhost:8080/ws' failed
```

**Solutions**:

#### Solution 1: Check backend is running

```bash
curl http://localhost:8080/health
```

#### Solution 2: Check WebSocket endpoint

```typescript
// Correct
const ws = new WebSocket("ws://localhost:8080/ws");

// Incorrect (missing /ws)
const ws = new WebSocket("ws://localhost:8080");
```

#### Solution 3: Check CORS settings

```go
// In router.go
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:5173"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type"},
    AllowCredentials: true,
}))
```

#### Solution 4: Check firewall

```bash
# macOS
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --add /path/to/whatspire

# Linux
sudo ufw allow 8080
```

**Prevention**:

- Test WebSocket connection in browser console
- Use WebSocket debugging tools

---

## 6. CORS Errors

### Problem: CORS policy blocks requests

**Symptoms**:

```
Access to fetch at 'http://localhost:8080/api/sessions' from origin 'http://localhost:5173' has been blocked by CORS policy
```

**Solutions**:

#### Solution 1: Update CORS configuration

```go
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}))
```

#### Solution 2: Use proxy in development

```typescript
// vite.config.ts
export default defineConfig({
  server: {
    proxy: {
      "/api": "http://localhost:8080",
      "/ws": {
        target: "ws://localhost:8080",
        ws: true,
      },
    },
  },
});
```

#### Solution 3: Check request headers

```typescript
// Add headers to requests
axios.get("/api/sessions", {
  headers: {
    "Content-Type": "application/json",
  },
});
```

**Prevention**:

- Configure CORS properly in development and production
- Use environment-specific CORS settings

---

## 7. Test Failures

### Problem A: Tests fail with database errors

**Symptoms**:

```
Error: database is locked
```

**Solutions**:

#### Solution 1: Use in-memory SQLite for tests

```go
db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
```

#### Solution 2: Clean up after tests

```go
func TestSomething(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    // ...
}
```

#### Solution 3: Run tests sequentially

```bash
go test -p 1 ./...
```

### Problem B: Property tests fail

**Symptoms**:

```
rapid: failed after 100 tests
```

**Solutions**:

#### Solution 1: Check test data

```go
// Ensure generated data is valid
rapid.Check(t, func(t *rapid.T) {
    value := rapid.String().Draw(t, "value")
    require.NotEmpty(t, value)
    // ...
})
```

#### Solution 2: Increase test iterations

```go
rapid.Check(t, func(t *rapid.T) {
    // ...
}, rapid.Trials(1000))
```

**Prevention**:

- Write deterministic tests
- Use test fixtures for complex scenarios
- Run tests before committing

---

## 8. Docker Issues

### Problem A: Docker daemon not running

**Symptoms**:

```
Cannot connect to the Docker daemon
```

**Solutions**:

#### Solution 1: Start Docker Desktop

- macOS: Open Docker Desktop from Applications
- Windows: Open Docker Desktop from Start Menu
- Linux: `sudo systemctl start docker`

#### Solution 2: Check Docker status

```bash
docker info
```

### Problem B: Container won't start

**Symptoms**:

```
Error response from daemon: driver failed programming external connectivity
```

**Solutions**:

#### Solution 1: Remove and recreate container

```bash
docker-compose down
docker-compose up -d
```

#### Solution 2: Check port conflicts

```bash
docker ps -a
# Look for containers using the same port
```

#### Solution 3: Clean up Docker

```bash
docker system prune -a
docker volume prune
```

**Prevention**:

- Use `docker-compose down` before `docker-compose up`
- Regularly clean up unused containers and volumes

---

## 9. Hot Reload Not Working

### Problem A: Backend changes not reflected

**Symptoms**:

- Code changes don't trigger restart
- Old code still running

**Solutions**:

#### Solution 1: Use air for hot-reload

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with air
air
```

#### Solution 2: Create .air.toml

```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ./cmd/whatsapp"
  bin = "tmp/main"
  include_ext = ["go", "yaml"]
  exclude_dir = ["tmp", "vendor"]
```

#### Solution 3: Manual restart

```bash
# Kill and restart
killall go
go run cmd/whatsapp/main.go
```

### Problem B: Frontend changes not reflected

**Symptoms**:

- UI doesn't update after code changes
- Browser shows old version

**Solutions**:

#### Solution 1: Hard refresh browser

- Chrome/Firefox: `Cmd+Shift+R` (macOS) or `Ctrl+Shift+R` (Windows/Linux)

#### Solution 2: Clear Vite cache

```bash
rm -rf apps/web/node_modules/.vite
bun run dev
```

#### Solution 3: Restart dev server

```bash
# Stop with Ctrl+C
# Restart
bun run dev
```

**Prevention**:

- Use hot-reload tools (air, Vite)
- Check file watchers aren't at limit (Linux)

---

## 10. Performance Issues

### Problem A: Slow database queries

**Symptoms**:

- API responses take >1 second
- Database queries timeout

**Solutions**:

#### Solution 1: Add indexes

```go
// Add indexes to frequently queried fields
type Event struct {
    SessionID string `gorm:"index"`
    Type      string `gorm:"index"`
    Timestamp time.Time `gorm:"index"`
}
```

#### Solution 2: Use query optimization

```go
// Bad: N+1 query
events, _ := repo.List()
for _, event := range events {
    session, _ := sessionRepo.GetByID(event.SessionID)
}

// Good: Preload
db.Preload("Session").Find(&events)
```

#### Solution 3: Add pagination

```go
// Limit results
db.Limit(100).Offset(page * 100).Find(&events)
```

#### Solution 4: Run benchmarks

```bash
go test -bench=. -benchmem ./test/benchmark/...
```

### Problem B: High memory usage

**Symptoms**:

- Application uses >1GB RAM
- Out of memory errors

**Solutions**:

#### Solution 1: Profile memory usage

```bash
go test -memprofile=mem.prof
go tool pprof mem.prof
```

#### Solution 2: Limit result sets

```go
// Don't load all records
db.Limit(1000).Find(&records)
```

#### Solution 3: Use streaming

```go
// Stream large result sets
rows, _ := db.Model(&Event{}).Rows()
defer rows.Close()
for rows.Next() {
    var event Event
    db.ScanRows(rows, &event)
    // Process event
}
```

**Prevention**:

- Run performance tests regularly
- Monitor memory usage in production
- Use profiling tools

---

## Getting Help

If you can't find a solution here:

1. **Check Logs**

   ```bash
   # Backend logs
   tail -f logs/app.log

   # Docker logs
   docker logs whatspire-server
   ```

2. **Search Issues**
   - Check GitHub Issues for similar problems
   - Search Stack Overflow

3. **Ask the Team**
   - Post in team Slack/Discord
   - Create a GitHub issue with:
     - Problem description
     - Steps to reproduce
     - Error messages
     - Environment details (OS, Go version, etc.)

4. **Debug Mode**
   ```bash
   # Enable debug logging
   export LOG_LEVEL=debug
   go run cmd/whatsapp/main.go
   ```

---

## Additional Resources

- [Development Setup Guide](./development_setup.md)
- [Architecture Decision Records](./adr/README.md)
- [API Specification](./api_specification.md)
- [Contributing Guidelines](../../CONTRIBUTING.md)

---

**Last Updated**: 2026-02-03  
**Maintainers**: Development Team
