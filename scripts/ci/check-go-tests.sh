#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

run_market_data_tests() {
  echo "[ci] running market-data tests"
  pushd "$ROOT_DIR/services/market-data" >/dev/null
  go test ./...
  popd >/dev/null
}

run_trading_tests() {
  echo "[ci] running trading tests (excluding integration test package)"
  pushd "$ROOT_DIR/services/trading" >/dev/null
  pkgs="$(go list ./... | grep -v '/internal/integration_test' | tr '\n' ' ')"
  # shellcheck disable=SC2086
  go test $pkgs
  popd >/dev/null
}

run_market_data_tests
run_trading_tests

echo "Go tests check passed"
