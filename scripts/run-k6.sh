#!/usr/bin/env bash
set -euo pipefail

SCRIPT=${1:-smoke-test.js}
OUT_DIR=${OUT_DIR:-.k6-results}
BASE_URL=${BASE_URL:-http://localhost:4000}

mkdir -p "$OUT_DIR"
TS=$(date +%Y%m%d-%H%M%S)
SUMMARY="$OUT_DIR/summary-$TS.json"
RESULTS="$OUT_DIR/results-$TS.json"

export BASE_URL
export OUT_DIR

if ! command -v k6 >/dev/null 2>&1; then
  echo "k6 not found. Run software-backend/scripts/install-k6.sh first." >&2
  exit 1
fi

k6 run --summary-export="$SUMMARY" --out json="$RESULTS" "software-backend/tests/k6/$SCRIPT"

echo "Summary: $SUMMARY"

echo "Samples: $RESULTS"
