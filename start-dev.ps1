# PowerShell script to start both backend and frontend servers
# Usage: .\start-dev.ps1

Write-Host "Starting WhatsApp API Development Servers..." -ForegroundColor Cyan
Write-Host ""

# Function to start backend
function Start-Backend {
    Write-Host "Starting Backend Server (Go)..." -ForegroundColor Green
    Set-Location "apps/server"
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "Write-Host 'Backend Server' -ForegroundColor Green; go run cmd/whatsapp/main.go"
    Set-Location "../.."
}

# Function to start frontend
function Start-Frontend {
    Write-Host "Starting Frontend Server (React + Vite)..." -ForegroundColor Blue
    Set-Location "apps/web"
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "Write-Host 'Frontend Server' -ForegroundColor Blue; npm run dev"
    Set-Location "../.."
}

# Check if Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Go is not installed. Please install Go first." -ForegroundColor Red
    exit 1
}

# Check if Node.js is installed
if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
    Write-Host "Node.js is not installed. Please install Node.js first." -ForegroundColor Red
    exit 1
}

# Check if npm dependencies are installed
if (-not (Test-Path "apps/web/node_modules")) {
    Write-Host "Installing frontend dependencies..." -ForegroundColor Yellow
    Set-Location "apps/web"
    npm install
    Set-Location "../.."
}

# Start servers
Start-Backend
Start-Sleep -Seconds 2
Start-Frontend

Write-Host ""
Write-Host "Development servers starting..." -ForegroundColor Green
Write-Host ""
Write-Host "Backend:  http://localhost:8080" -ForegroundColor Cyan
Write-Host "Frontend: http://localhost:5173" -ForegroundColor Cyan
Write-Host ""
Write-Host "Press Ctrl+C in each terminal window to stop the servers" -ForegroundColor Yellow
