#!/usr/bin/env bash
# ============================================================
# Build all services for production
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "==> Building all Docker images..."
cd "$PROJECT_DIR"
docker compose build --no-cache

echo ""
echo "Build complete! Start with:"
echo "  docker compose up -d"
echo ""
