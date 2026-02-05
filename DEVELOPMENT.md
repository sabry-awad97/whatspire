# Development Guide

Quick reference for running the WhatsApp API project in development mode.

---

## 🚀 Quick Start

### Option 1: Using Taskfile (Recommended)

**Install Task** (if not already installed):

- **Windows**: `choco install go-task` or download from [taskfile.dev](https://taskfile.dev/installation/)
- **macOS**: `brew install go-task`
- **Linux**: `sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin`

**Start Development Servers**:

```bash
task dev
```

This will start both backend and frontend servers in parallel.

**Other Useful Commands**:

```bash
task dev:backend      # Start only backend
task dev:frontend     # Start only frontend
task build            # Build both projects
task test             # Run all tests
task lint             # Lint all code
task setup            # Initial project setup
task --list           # Show all available tasks
```

---

### Option 2: Using Startup Scripts

**Windows (PowerShell)**:

```powershell
.\start-dev.ps1
```

**Linux/macOS (Bash)**:

```bash
chmod +x start-dev.sh
./start-dev.sh
```

These scripts will:

1. Check if Go and Node.js are installed
2. Install npm dependencies if needed
3. Start both servers in separate terminal windows (Windows) or background processes (Linux/macOS)

---

### Option 3: Manual Start

**Terminal 1 - Backend**:

```bash
cd apps/server
go run cmd/whatsapp/main.go
```

**Terminal 2 - Frontend**:

```bash
cd apps/web
npm run dev
```

---

## 📍 Server URLs

Once started, the servers will be available at:

- **Backend API**: http://localhost:8080
  - Health Check: http://localhost:8080/health
  - API Docs: http://localhost:8080/api
  - Metrics: http://localhost:8080/metrics

- **Frontend**: http://localhost:5173
  - Main App: http://localhost:5173
  - Sessions: http://localhost:5173/sessions
  - New Session: http://localhost:5173/sessions/new

---

## 🔧 Prerequisites

### Required Software

1. **Go** (1.21 or higher)
   - Download: https://go.dev/dl/
   - Verify: `go version`

2. **Node.js** (18 or higher)
   - Download: https://nodejs.org/
   - Verify: `node --version`

3. **npm** (comes with Node.js)
   - Verify: `npm --version`

### Optional Tools

- **Task** - Task runner (recommended)
- **Docker** - For containerized deployment
- **Git** - Version control

---

## 📦 Initial Setup

### First Time Setup

1. **Clone the repository** (if not already done):

   ```bash
   git clone <repository-url>
   cd whatspire
   ```

2. **Install dependencies**:

   ```bash
   # Using Task
   task install

   # Or manually
   cd apps/server && go mod download && cd ../..
   cd apps/web && npm install && cd ../..
   ```

3. **Configure environment**:

   ```bash
   # Backend config already created at apps/server/config.yaml
   # Frontend env already configured at apps/web/.env
   ```

4. **Start development servers**:
   ```bash
   task dev
   # or use startup scripts
   ```

---

## 🧪 Testing

### Run All Tests

```bash
task test
```

### Backend Tests Only

```bash
task test:backend
# or
cd apps/server && go test ./... -v
```

### Frontend Tests Only

```bash
task test:frontend
# or
cd apps/web && npm run test
```

### Backend Tests with Coverage

```bash
task test:backend:coverage
# Opens coverage.html in browser
```

---

## 🔍 Linting & Formatting

### Lint All Code

```bash
task lint
```

### Backend Linting

```bash
task lint:backend
# or
cd apps/server && go vet ./... && go fmt ./...
```

### Frontend Linting

```bash
task lint:frontend
# or
cd apps/web && npm run lint
```

---

## 🏗️ Building

### Build Both Projects

```bash
task build
```

### Backend Build

```bash
task build:backend
# Creates: bin/whatsapp-server
```

### Frontend Build

```bash
task build:frontend
# Creates: apps/web/dist/
```

---

## 🐳 Docker

### Start with Docker Compose

```bash
task docker:up
# or
docker-compose up -d
```

### Stop Docker Containers

```bash
task docker:down
# or
docker-compose down
```

### View Docker Logs

```bash
task docker:logs
# or
docker-compose logs -f
```

### Build Docker Images

```bash
task docker:build
# or
docker-compose build
```

---

## 🗄️ Database

### Run Migrations

```bash
task db:migrate
```

### Reset Database (⚠️ Deletes all data)

```bash
# Windows
task db:reset:windows

# Linux/macOS
task db:reset
```

---

## 🧹 Cleanup

### Clean All Build Artifacts

```bash
task clean
```

### Clean Backend Only

```bash
task clean:backend
```

### Clean Frontend Only

```bash
task clean:frontend
```

---

## 📊 Monitoring

### Check Service Status

```bash
task status
```

### View Application Logs

```bash
task logs
```

### Health Checks

```bash
# Backend health
curl http://localhost:8080/health

# Backend readiness
curl http://localhost:8080/ready

# Metrics
curl http://localhost:8080/metrics
```

---

## 🔥 Troubleshooting

### Port Already in Use

**Backend (8080)**:

```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/macOS
lsof -ti:8080 | xargs kill -9
```

**Frontend (5173)**:

```bash
# Windows
netstat -ano | findstr :5173
taskkill /PID <PID> /F

# Linux/macOS
lsof -ti:5173 | xargs kill -9
```

### Go Module Issues

```bash
cd apps/server
go mod tidy
go mod download
```

### npm Issues

```bash
cd apps/web
rm -rf node_modules package-lock.json
npm install
```

### CORS Errors

- Check `apps/server/config.yaml` - ensure `cors.allowed_origins` includes `http://localhost:5173`
- Restart backend server after config changes

### Database Connection Issues

- Check database path in `apps/server/config.yaml`
- Ensure directory exists and has write permissions
- For SQLite: Check `/data/whatsmeow.db` path

---

## 📚 Additional Resources

- **Project Documentation**: See `apps/server/docs/` folder
- **API Specification**: `apps/server/docs/api_specification.md`
- **Configuration Guide**: `apps/server/docs/configuration.md`
- **Troubleshooting**: `apps/server/docs/troubleshooting.md`
- **Implementation Summary**: `session-integration-fixes/IMPLEMENTATION_SUMMARY.md`

---

## 🎯 Development Workflow

### Typical Development Session

1. **Start servers**:

   ```bash
   task dev
   ```

2. **Make changes** to code

3. **Backend changes**: Server auto-restarts (if using air/nodemon)
   - Or manually restart: Ctrl+C and `task dev:backend`

4. **Frontend changes**: Hot reload automatic (Vite HMR)

5. **Test changes**:

   ```bash
   task test
   ```

6. **Lint code**:

   ```bash
   task lint
   ```

7. **Commit changes**:
   ```bash
   git add .
   git commit -m "Your commit message"
   git push
   ```

---

## 🚦 Git Workflow

### Feature Development

1. **Create feature branch**:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make changes and commit**:

   ```bash
   git add .
   git commit -m "feat: your feature description"
   ```

3. **Run checks before push**:

   ```bash
   task check  # Runs lint, test, build
   ```

4. **Push to remote**:

   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create Pull Request** on GitHub/GitLab

---

## 💡 Tips

- Use `task --list` to see all available commands
- Backend logs are in `apps/server/logs/`
- Frontend uses Vite's fast HMR for instant updates
- Use browser DevTools Network tab to debug API calls
- Check `apps/server/config.yaml` for all configuration options
- Use `task status` to quickly check if servers are running

---

## 🆘 Getting Help

If you encounter issues:

1. Check the troubleshooting section above
2. Review `apps/server/docs/troubleshooting.md`
3. Check application logs
4. Verify all prerequisites are installed
5. Try `task clean` and `task setup` to reset

---

**Happy Coding! 🚀**
