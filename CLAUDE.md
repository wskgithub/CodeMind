# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CodeMind (度影智能编码服务) is an enterprise-level AI coding service management platform. It acts as a middleware layer that proxies requests to locally deployed LLM models, providing user management, resource control, and usage statistics.

**Tech Stack:**
- **Frontend**: React 18 + TypeScript + Vite + Ant Design 5 + TailwindCSS
- **Backend**: Go 1.23 + Gin + GORM
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Deployment**: Docker + Docker Compose + Nginx

## Common Development Commands

### Infrastructure (Docker)

```bash
# Start infrastructure services only (PostgreSQL + Redis)
docker compose up -d postgres redis

# Start all services (production)
docker compose up -d --build

# Stop all services
docker compose down

# View logs
docker compose logs -f [service]
```

### Backend

```bash
cd backend

# Run development server (requires infrastructure running)
go run cmd/server/main.go

# Build binary
make build

# Run tests with coverage
make test

# Generate coverage report (HTML)
make test-coverage

# Run linter
make lint

# Generate Swagger documentation
make swagger

# Clean build artifacts
make clean

# Run database migrations
make migrate
```

**Backend Entry Point**: `backend/cmd/server/main.go`

### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Run tests
npm test -- --run

# Run linter
npm run lint
```

### Project-Wide Scripts

```bash
# Start development infrastructure (postgres + redis)
./scripts/dev.sh

# Build all Docker images
./scripts/build.sh

# Run all tests (backend + frontend)
./scripts/test.sh
```

## High-Level Architecture

### Backend Structure (Layered Architecture)

```
backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management (Viper)
│   ├── middleware/      # HTTP middleware (auth, CORS, rate limiting, etc.)
│   ├── handler/         # HTTP handlers (controllers)
│   ├── service/         # Business logic layer
│   ├── repository/      # Data access layer (GORM)
│   ├── model/           # Data models & DTOs
│   ├── router/          # Route definitions (Gin)
│   └── pkg/             # Internal utilities (JWT, crypto, response, etc.)
├── pkg/                 # External shared packages (LLM client, token counter)
└── migrations/          # SQL migration files
```

> Tests are co-located with source files following Go conventions (e.g. `service/user.go` + `service/service_test.go`).

**Request Flow**:
1. Router → Middleware Chain → Handler → Service → Repository → PostgreSQL
2. LLM Proxy: API Key validation → Rate limit check → Concurrency check → Forward to LLM → Stream response → Record usage

### Frontend Structure

```
frontend/src/
├── pages/               # Page components (route-based)
│   ├── home/            # Landing page
│   ├── login/           # Authentication
│   ├── dashboard/       # User dashboard
│   ├── keys/            # API Key management
│   ├── usage/           # Usage statistics
│   ├── profile/         # User profile
│   ├── docs/            # Documentation viewer
│   └── admin/           # Admin pages (users, departments, limits, etc.)
├── components/
│   ├── common/          # Reusable components (UsageProgressCards, etc.)
│   └── layout/          # Layout components (DashboardLayout)
├── assets/styles/       # Global and module CSS
├── services/            # API client layer (Axios)
├── store/               # Zustand state management
├── hooks/               # Custom React hooks
├── types/               # TypeScript type definitions
└── router/              # React Router configuration
```

### Key Design Patterns

**User Roles**: Three-tier RBAC system
- `super_admin`: Full system access
- `dept_manager`: Manages department users and statistics
- `user`: Personal API keys and usage only

**API Key Format**: `cm-{32-char-hex}` - Only shown once at creation, stored as SHA-256 hash

**Rate Limiting Priority**: User limit > Department limit > Global limit

**LLM Proxy**: OpenAI-compatible API at `/v1/*` endpoints with SSE streaming support

## Configuration

### Environment Setup

1. Copy `.env.example` to `.env` and fill in required values
2. Copy `deploy/config/app.yaml.example` to `deploy/config/app.yaml`
3. Minimum required env vars: `DB_PASSWORD`, `JWT_SECRET`, `LLM_BASE_URL`

### Default Credentials

- Username: `admin`
- Password: `Admin@123456` (change on first login)

## Important Implementation Notes

### LLM Proxy Module

The core LLM proxy (`pkg/llm/`) must:
- Support SSE (Server-Sent Events) streaming
- Parse token usage from LLM responses
- Record usage asynchronously after request completion
- Handle concurrent request limits via Redis counter
- Return OpenAI-compatible error responses

### Redis Key Patterns

| Pattern | Type | TTL | Purpose |
|---------|------|-----|---------|
| `codemind:jwt:blacklist:{jti}` | String | JWT remaining | JWT blacklist |
| `codemind:apikey:{hash}` | String(JSON) | 300s | API Key cache |
| `codemind:concurrency:{user_id}` | Counter | 300s | Concurrent requests |
| `codemind:usage:{user_id}:daily:{date}` | Counter | 48h | Daily token usage |

### Security Requirements

- Passwords: bcrypt with cost factor 12
- API Keys: SHA-256 hash storage, never log full keys
- JWT: HS256 algorithm, 24-hour expiration
- All sensitive operations: audit log with IP, operator, timestamp

### Nginx Configuration

The Nginx reverse proxy (`deploy/docker/nginx/nginx.conf`) handles:
- Frontend static files with SPA fallback
- `/api/` → Backend management APIs
- `/v1/` → LLM proxy with SSE support and 600s timeout

## Testing Strategy

- **Backend**: Unit tests (service/repository) > 80% coverage, integration tests for full request flow
- **Frontend**: Component tests with Vitest + React Testing Library
- **Test Command**: `./scripts/test.sh` runs both backend and frontend tests

## Deployment

Production deployment requires:
1. Set all environment variables in `.env`
2. Configure `deploy/config/app.yaml`
3. Run: `docker compose up -d --build`

Services are orchestrated via `docker-compose.yml` with health checks for PostgreSQL and Redis.
