#!/usr/bin/env bash
# Runs the backend test suite and reports coverage of hand-written code.
# Generated GraphQL code, in-memory test doubles and the main packages are excluded.
# Fails if coverage is below MIN_COVERAGE (default 95).
set -euo pipefail
cd "$(dirname "$0")/../backend"

MIN_COVERAGE="${MIN_COVERAGE:-95}"
go test ./... -race -count=1 -coverpkg=./... -coverprofile=coverage.raw.out "$@"

grep -v -E '/graph/generated\.go|/graph/model/models_gen\.go|/internal/memstore/|/cmd/' coverage.raw.out > coverage.out
total=$(go tool cover -func=coverage.out | awk '/^total:/ {sub("%","",$3); print $3}')
echo "Backend coverage (excluding generated code): ${total}%"
awk -v t="$total" -v m="$MIN_COVERAGE" 'BEGIN { if (t+0 < m+0) { print "coverage below " m "%"; exit 1 } }'
