#!/bin/bash

# Bash script to start both backend and frontend servers
# Usage: ./start-dev.sh

set -e

echo "🚀 Starting WhatsApp API Development Servers..."
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go first.${NC}"
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}❌ Node.js is not installed. Please install Node.js first.${NC}"
    exit 1
fi

# Check if npm dependencies are installed
if [ ! -d "apps/web/node_modules" ]; then
    echo -e "${YELLOW}📦 Installing frontend dependencies...${NC}"
    cd apps/web
    npm install
    cd ../..
fi

# Function to cleanup on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}🛑 Stopping servers...${NC}"
    kill $(jobs -p) 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

# Start backend server
echo -e "${GREEN}📦 Starting Backend Server (Go)...${NC}"
cd apps/server
go run cmd/whatsapp/main.go &
BACKEND_PID=$!
cd ../..

# Wait a bit for backend to start
sleep 2

# Start frontend server
echo -e "${BLUE}📦 Starting Frontend Server (React + Vite)...${NC}"
cd apps/web
npm run dev &
FRONTEND_PID=$!
cd ../..

echo ""
echo -e "${GREEN}✅ Development servers started!${NC}"
echo ""
echo -e "${CYAN}📍 Backend:  http://localhost:8080${NC}"
echo -e "${CYAN}📍 Frontend: http://localhost:5173${NC}"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop all servers${NC}"
echo ""

# Wait for both processes
wait $BACKEND_PID $FRONTEND_PID
