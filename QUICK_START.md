# 🚀 Quick Start Guide

Get the WhatsApp API project running in 3 simple steps!

---

## ⚡ Super Quick Start

```bash
# 1. Install dependencies
task install

# 2. Start development servers
task dev

# 3. Open browser
# Frontend: http://localhost:5173
# Backend:  http://localhost:8080
```

---

## 📋 Prerequisites

- ✅ Go 1.21+ ([Download](https://go.dev/dl/))
- ✅ Node.js 18+ ([Download](https://nodejs.org/))
- ✅ Task (optional but recommended) ([Install](https://taskfile.dev/installation/))

---

## 🎯 Three Ways to Start

### 1️⃣ Using Task (Recommended)

```bash
task dev
```

### 2️⃣ Using Startup Scripts

**Windows**:

```powershell
.\start-dev.ps1
```

**Linux/macOS**:

```bash
./start-dev.sh
```

### 3️⃣ Manual Start

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

## 🌐 Access Points

| Service      | URL                           | Description        |
| ------------ | ----------------------------- | ------------------ |
| Frontend     | http://localhost:5173         | React web app      |
| Backend API  | http://localhost:8080         | Go REST API        |
| Health Check | http://localhost:8080/health  | Server health      |
| Metrics      | http://localhost:8080/metrics | Prometheus metrics |

---

## 🧪 Test the Integration

1. **Open Frontend**: http://localhost:5173
2. **Navigate to**: Sessions → New Session
3. **Create Session**: Fill form and click "Create Session"
4. **Scan QR Code**: Use WhatsApp mobile app
5. **Verify**: Session status updates to "connected"

---

## 📚 More Information

- **Full Development Guide**: See [DEVELOPMENT.md](DEVELOPMENT.md)
- **Implementation Details**: See [session-integration-fixes/IMPLEMENTATION_SUMMARY.md](session-integration-fixes/IMPLEMENTATION_SUMMARY.md)
- **API Documentation**: See [apps/server/docs/](apps/server/docs/)

---

## 🆘 Common Issues

### Port Already in Use

```bash
# Kill process on port 8080 (backend)
# Windows: netstat -ano | findstr :8080
# Linux/macOS: lsof -ti:8080 | xargs kill -9

# Kill process on port 5173 (frontend)
# Windows: netstat -ano | findstr :5173
# Linux/macOS: lsof -ti:5173 | xargs kill -9
```

### Dependencies Not Installed

```bash
task install
# or
cd apps/server && go mod download && cd ../..
cd apps/web && npm install && cd ../..
```

### CORS Errors

Check `apps/server/config.yaml`:

```yaml
cors:
  allowed_origins:
    - "http://localhost:5173"
    - "http://localhost:3000"
```

---

## 🎉 You're Ready!

The project is now running. Start developing! 🚀

For detailed information, see [DEVELOPMENT.md](DEVELOPMENT.md)
