#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

UNFORMATTED="$(find services -name '*.go' -not -path '*/vendor/*' -print0 | xargs -0 gofmt -l)"

if [[ -n "$UNFORMATTED" ]]; then
  echo "Found unformatted Go files:"
  echo "$UNFORMATTED"
  exit 1
fi

echo "Go formatting check passed"
