#!/usr/bin/env bash
# ============================================================
# Start development environment
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "==> Starting infrastructure services (PostgreSQL + Redis)..."
cd "$PROJECT_DIR"
docker compose up -d postgres redis

echo "==> Waiting for PostgreSQL to be ready..."
until docker compose exec postgres pg_isready -U codemind > /dev/null 2>&1; do
    sleep 1
done
echo "    PostgreSQL is ready."

echo "==> Waiting for Redis to be ready..."
until docker compose exec redis redis-cli ping > /dev/null 2>&1; do
    sleep 1
done
echo "    Redis is ready."

echo ""
echo "Infrastructure is ready! Now start the services:"
echo ""
echo "  Backend:  cd backend && go run cmd/server/main.go"
echo "  Frontend: cd frontend && npm run dev"
echo ""
