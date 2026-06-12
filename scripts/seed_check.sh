#!/usr/bin/env bash
# Basic health checks after seed — see cinema -ticket-flow.MD

set -euo pipefail

API_URL="${API_URL:-http://localhost:8080}"

echo "Checking backend health..."
curl -sf "$API_URL/health" | grep -q '"status":"ok"' && echo "OK: /health" || echo "FAIL: /health"

echo "Checking showtimes..."
curl -sf "$API_URL/api/showtimes" && echo "" || echo "FAIL: /api/showtimes"
