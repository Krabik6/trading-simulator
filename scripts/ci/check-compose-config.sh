#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

for env in dev stage prod; do
  file="deploy/docker/docker-compose.${env}.yml"
  if [[ ! -f "$file" ]]; then
    echo "Missing compose file: $file"
    exit 1
  fi

  env_file="deploy/env/${env}.env"
  if [[ -f "$env_file" ]]; then
    docker compose --env-file "$env_file" -f "$file" config >/dev/null
  else
    docker compose -f "$file" config >/dev/null
  fi
  echo "Compose config valid: $file"
done
