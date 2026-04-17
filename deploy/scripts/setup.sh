#!/usr/bin/env bash
# ============================================================
# Initial server setup for CodeMind deployment
# ============================================================
set -euo pipefail

echo "==> CodeMind Initial Setup"
echo ""

# Check prerequisites
command -v docker >/dev/null 2>&1 || { echo "Error: docker is required but not installed."; exit 1; }
command -v docker compose >/dev/null 2>&1 || { echo "Error: docker compose is required but not installed."; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_DIR="$(dirname "$DEPLOY_DIR")"

# Create .env if not exists
if [ ! -f "$PROJECT_DIR/.env" ]; then
    echo "==> Creating .env from template..."
    cp "$PROJECT_DIR/.env.example" "$PROJECT_DIR/.env"
    echo "    Please edit .env with your actual values before starting."
    echo ""
fi

# Create config if not exists
if [ ! -f "$DEPLOY_DIR/config/app.yaml" ]; then
    echo "==> Creating app.yaml from template..."
    cp "$DEPLOY_DIR/config/app.yaml.example" "$DEPLOY_DIR/config/app.yaml"
    echo "    Please edit deploy/config/app.yaml with your actual values."
    echo ""
fi

echo "==> Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Edit .env with database passwords, JWT secret, and LLM server URL"
echo "  2. Edit deploy/config/app.yaml with application settings"
echo "  3. Run: docker compose up -d"
echo ""
