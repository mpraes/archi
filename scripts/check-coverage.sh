#!/usr/bin/env bash
set -euo pipefail

MIN_COVERAGE="${MIN_COVERAGE:-75}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

PKGS=$(go list ./internal/...)
go test -count=1 -coverprofile=coverage.out $PKGS

TOTAL=$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%","",$3); print $3}')
echo "Total internal coverage: ${TOTAL}% (min ${MIN_COVERAGE}%)"

awk -v total="$TOTAL" -v min="$MIN_COVERAGE" 'BEGIN {
  if (total + 0 < min + 0) {
    printf "Coverage %.1f%% is below minimum %s%%\n", total, min
    exit 1
  }
}'
