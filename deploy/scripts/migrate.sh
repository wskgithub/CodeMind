#!/usr/bin/env bash
# ============================================================
# Run database migrations
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"

echo "==> Running database migrations..."
cd "$PROJECT_DIR/backend"

# Use golang-migrate or GORM auto-migrate (depending on implementation)
go run cmd/server/main.go --migrate

echo "==> Migrations complete."
