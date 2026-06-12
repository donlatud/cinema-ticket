#!/usr/bin/env bash
# Concurrent seat lock test — see Phase 11 in cinema -ticket-flow.MD
# Usage: ./scripts/concurrency_test.sh <showtime_id> <seat_no> <token>

set -euo pipefail

SHOWTIME_ID="${1:-}"
SEAT_NO="${2:-A5}"
TOKEN="${3:-}"

if [[ -z "$SHOWTIME_ID" || -z "$TOKEN" ]]; then
  echo "Usage: $0 <showtime_id> <seat_no> <bearer_token>"
  exit 1
fi

API_URL="${API_URL:-http://localhost:8080}"

for i in $(seq 1 10); do
  curl -s -o /dev/null -w "Request $i: %{http_code}\n" \
    -X POST "$API_URL/api/showtimes/$SHOWTIME_ID/seats/lock" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"seat_nos\":[\"$SEAT_NO\"]}" &
done
wait
