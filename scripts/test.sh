#!/usr/bin/env bash
# ============================================================
# Run all tests (backend + frontend)
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

FAILED=0

echo "========================================"
echo "  Running Backend Tests"
echo "========================================"
cd "$PROJECT_DIR/backend"
if go test ./... -v -cover; then
    echo "  ✓ Backend tests passed"
else
    echo "  ✗ Backend tests failed"
    FAILED=1
fi

echo ""
echo "========================================"
echo "  Running Frontend Tests"
echo "========================================"
cd "$PROJECT_DIR/frontend"
if npm test -- --run; then
    echo "  ✓ Frontend tests passed"
else
    echo "  ✗ Frontend tests failed"
    FAILED=1
fi

echo ""
echo "========================================"
if [ $FAILED -eq 0 ]; then
    echo "  All tests passed!"
else
    echo "  Some tests failed. Please check the output above."
    exit 1
fi
