#!/bin/bash

# Ensure portless proxy is running
echo "Starting portless proxy..."
bun x portless proxy start

cleanup() {
    echo -e "\nStopping frontend and backend servers"
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
