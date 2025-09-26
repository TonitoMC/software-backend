#!/usr/bin/env bash
set -euo pipefail

# Install k6 on Ubuntu/Debian (Codespaces)
if command -v k6 >/dev/null 2>&1; then
  echo "k6 already installed: $(k6 version)"
  exit 0
fi

. /etc/os-release
if [[ ${ID:-} =~ (ubuntu|debian) ]]; then
  sudo apt-get update -y
  sudo apt-get install -y ca-certificates gnupg curl
  sudo mkdir -p /usr/share/keyrings
  curl -fsSL https://dl.k6.io/key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/k6-archive-keyring.gpg
  echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb/ stable main" | sudo tee /etc/apt/sources.list.d/k6.list > /dev/null
  sudo apt-get update -y
  sudo apt-get install -y k6
else
  echo "Please install k6 for your distro manually: https://k6.io/docs/get-started/installation/"
  exit 1
fi

echo "Installed k6: $(k6 version)"
