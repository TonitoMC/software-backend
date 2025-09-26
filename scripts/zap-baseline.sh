#!/usr/bin/env bash
set -euo pipefail

TARGET=${1:-http://localhost:4000}
CONTAINER_NAME=zap-baseline

# Requires Docker in Codespaces
if ! command -v docker >/dev/null 2>&1; then
  echo "Docker is required to run OWASP ZAP baseline scan in this mode." >&2
  exit 1
fi

# Pull and run ZAP Baseline scan
# Docs: https://www.zaproxy.org/docs/docker/baseline-scan/
docker run --rm --name "$CONTAINER_NAME" -t owasp/zap2docker-stable zap-baseline.py \
  -t "$TARGET" \
  -I \
  -r zap-baseline-report.html \
  -w zap-warnings.md

echo "Reports written to current directory: zap-baseline-report.html, zap-warnings.md"
