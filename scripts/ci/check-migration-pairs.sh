#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

failed=0

while IFS= read -r -d '' up_file; do
  down_file="${up_file/.up.sql/.down.sql}"
  if [[ ! -f "$down_file" ]]; then
    echo "Missing down migration for: $up_file"
    failed=1
  fi
done < <(find services -path '*/migrations/*.up.sql' -print0)

while IFS= read -r -d '' down_file; do
  up_file="${down_file/.down.sql/.up.sql}"
  if [[ ! -f "$up_file" ]]; then
    echo "Missing up migration for: $down_file"
    failed=1
  fi
done < <(find services -path '*/migrations/*.down.sql' -print0)

if [[ "$failed" -ne 0 ]]; then
  echo "Migration pair check failed"
  exit 1
fi

echo "Migration pair check passed"
