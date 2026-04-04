# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Coding Conventions

- **代码注释使用中文** (Code comments should be written in Chinese)
- 完成工作或有需要澄清的问题时，请主动提问获取进一步指示

## Project Overview

CodeMind is an enterprise-level AI coding service management platform. It acts as a middleware layer that proxies requests to locally deployed LLM models, providing user management, resource control, and usage statistics.

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

# Run all tests with coverage
make test

# Run a single test
go test ./internal/service/ -run TestUserService -v

# Generate coverage report (HTML)
make test-coverage

# Run linter
make lint

# Generate Swagger documentation
make swagger

# Run database migrations
make migrate
```

**Backend Entry Point**: `backend/cmd/server/main.go`

### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Run development server (proxies API to localhost:8080 via Vite)
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
./scripts/dev.sh      # Start development infrastructure (postgres + redis)
./scripts/build.sh    # Build all Docker images
./scripts/test.sh     # Run all tests (backend + frontend)
```

## High-Level Architecture

### Backend Structure (Layered Architecture)

```
backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management (Viper: YAML + env vars)
│   ├── middleware/      # HTTP middleware (auth, CORS, rate limiting, etc.)
│   ├── handler/         # HTTP handlers (controllers)
│   ├── service/         # Business logic layer
│   ├── repository/      # Data access layer (GORM)
│   ├── model/           # Data models & DTOs
│   ├── router/          # Route definitions (Gin)
│   └── pkg/             # Internal utilities (JWT, crypto, response, etc.)
├── pkg/                 # External shared packages (LLM client, token counter)
└── migrations/          # SQL migration files (numbered prefix: 001_, 002_, ...)
```

> Tests are co-located with source files following Go conventions (e.g. `service/user.go` + `service/service_test.go`).

**Request Flow**:
1. Router → Middleware Chain → Handler → Service → Repository → PostgreSQL
2. LLM Proxy: API Key validation → Rate limit check → Concurrency check → Load balancer selects provider → Forward to LLM → Stream response → Record usage

**Dependency Injection**: App initializes bottom-up in `main.go`: Config → DB/Redis → Repositories → Services → Handlers → Router. All dependencies are injected via constructors.

### API Route Groups

| Route Group | Auth Method | Access |
|-------------|------------|--------|
| `/health` | None | Public |
| `/api/v1/auth/*` | None | Login, register, refresh |
| `/api/v1/*` (authenticated) | JWT | All logged-in users |
| `/api/v1/admin/*` | JWT + role | `super_admin` / `dept_manager` |
| `/api/v1/system/*` | JWT + role | `super_admin` only |
| `/v1/*` | API Key | OpenAI-compatible LLM proxy |
| `/mcp/*` | API Key | MCP protocol gateway |

### Frontend Structure

```
frontend/src/
├── pages/               # Page components (route-based, lazy loaded)
├── components/
│   ├── common/          # Reusable components (UsageProgressCards, etc.)
│   └── layout/          # Layout components (DashboardLayout)
├── services/            # API client layer (Axios, base URL: /api/v1)
├── store/               # Zustand state management
│   ├── authStore.ts     # Auth state (token/user in localStorage)
│   └── appStore.ts      # UI state (theme, sidebar)
├── router/              # React Router with AuthGuard (role-based) and GuestGuard
└── types/               # TypeScript type definitions
```

**Frontend Styling**: Ant Design + Tailwind utility classes + inline styles with CSS variables. Preflight disabled. Vite build splits vendor chunks for React, Ant Design, and ECharts.

### Key Design Patterns

**User Roles**: Three-tier RBAC system
- `super_admin`: Full system access
- `dept_manager`: Manages department users and statistics
- `user`: Personal API keys and usage only

**API Key Format**: `cm-{32-char-hex}` - Only shown once at creation, stored as SHA-256 hash

**Rate Limiting Priority**: User limit > Department limit > Global limit

**Soft Delete**: Models use `deleted_at` timestamps (GORM soft delete pattern)

**Login Lockout**: Exponential backoff for failed attempts (5 min → 24h max)

## LLM Proxy & Multi-Provider Architecture

The LLM proxy (`backend/pkg/llm/`) supports multiple providers with intelligent routing:

- **Provider Manager** (`manager.go`): Manages multiple LLM providers (OpenAI/Anthropic), supports model routing rules (e.g., `claude-*` → anthropic provider)
- **Load Balancer** (`balancer.go`): Weighted round-robin with Redis-based user affinity for sticky sessions
- **SSE Streaming**: All LLM responses stream via Server-Sent Events
- **MCP Gateway** (`/mcp/*`): Separate protocol gateway for Model Context Protocol integration
- **Usage Recording**: Token usage parsed from responses and recorded asynchronously

## Configuration

### Environment Setup

1. Copy `.env.example` to `.env` and fill in required values
2. Copy `deploy/config/app.yaml.example` to `deploy/config/app.yaml`
3. Minimum required env vars: `DB_PASSWORD`, `JWT_SECRET`, `LLM_BASE_URL`

**Config Loading Order**: Default values → YAML file → Environment variables (use `_` for `.` separator, e.g., `database.host` → `DB_HOST`)

### Default Credentials

- Username: `admin`
- Password: `Admin@123456` (change on first login)

## Redis Key Patterns

| Pattern | Type | TTL | Purpose |
|---------|------|-----|---------|
| `codemind:jwt:blacklist:{jti}` | String | JWT remaining | JWT blacklist |
| `codemind:apikey:{hash}` | String(JSON) | 300s | API Key cache |
| `codemind:concurrency:{user_id}` | Counter | 300s | Concurrent requests |
| `codemind:usage:{user_id}:daily:{date}` | Counter | 48h | Daily token usage |

## Security Requirements

- Passwords: bcrypt with cost factor 12
- API Keys: SHA-256 hash storage, never log full keys
- JWT: HS256 algorithm, 24-hour expiration, blacklisted on logout/password change
- All sensitive operations: audit log with IP, operator, timestamp

## Testing Strategy

- **Backend**: Unit tests (service/repository) > 80% coverage, integration tests for full request flow
- **Frontend**: Component tests with Vitest + React Testing Library
- **Test Command**: `./scripts/test.sh` runs both backend and frontend tests

## Deployment

Production deployment requires:
1. Set all environment variables in `.env`
2. Configure `deploy/config/app.yaml`
3. Run: `docker compose up -d --build`

Services are orchestrated via `docker-compose.yml` with health checks for PostgreSQL and Redis. Nginx handles:
- Frontend static files with SPA fallback
- `/api/` → Backend management APIs
- `/v1/` → LLM proxy with SSE support and 600s timeout
