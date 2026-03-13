#!/bin/bash

set -e

# Setup mode
if [ "$1" = "--setup" ]; then
    echo "=== Interview Platform Setup ==="
    
    command -v psql >/dev/null 2>&1 || { echo "PostgreSQL not installed. Install: sudo apt install postgresql"; exit 1; }
    command -v go >/dev/null 2>&1 || { echo "Go not installed. Install from: https://go.dev/dl/"; exit 1; }
    command -v bun >/dev/null 2>&1 || { echo "Bun not installed. Install: curl -fsSL https://bun.sh/install | bash"; exit 1; }
    
    echo "Setting up backend..."
    cd backend
    [ ! -f .env ] && cp .env.example .env && echo "Edit backend/.env and add your API keys"
    go mod download
    cd ..
    
    echo "Setting up frontend..."
    cd frontend
    [ ! -f .env ] && cp .env.example .env
    bun install
    cd ..
    
    echo ""
    echo "Setup complete!"
    echo "Edit .env files, then run: ./start.sh"
    exit 0
fi

# Start mode
echo "Starting portless proxy..."
bun x portless proxy start

cleanup() {
    echo -e "\nStopping servers"
    kill $BACKEND_PID 2>/dev/null
    kill $FRONTEND_PID 2>/dev/null
    exit 0
}

trap cleanup SIGINT

echo "Starting backend: api.simulate-interview.localhost"
cd backend
bun x portless api.simulate-interview --force go run main.go &
BACKEND_PID=$!
cd ..

echo "Starting frontend: simulate-interview.localhost"
cd frontend
bun x portless simulate-interview --force bun run dev &
FRONTEND_PID=$!
cd ..

echo "Press ctrl+c to exit"
wait $BACKEND_PID $FRONTEND_PID
