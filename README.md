# Interview Platform

AI-powered coding interview practice platform with personalized problem generation.

## Prerequisites

- PostgreSQL 14+
- Go 1.21+
- Bun 1.0+
- OpenrouterAPI Key

## Quick Start

```bash
# Clone repository
git clone <repo-url>
cd simulate-interview

# First time setup only
./start.sh --setup

# Edit environment variables
# backend/.env - Add API key

# Start application
./start.sh
```

## Manual Setup

### Backend

```bash
cd backend
cp .env.example .env
# Edit .env with your configuration
go mod download
go run main.go
```

### Frontend

```bash
cd frontend
cp .env.example .env
bun install
bun run dev
```
