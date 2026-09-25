#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

case "${1:-}" in
  test)
    exec go test -race -tags "all_inputs all_sinks all_storage all_crypto" ./...
    ;;
  bench)
    exec go test -bench=. ./...
    ;;
  arch-check)
    exec make -C "$ROOT_DIR" arch-check
    ;;
  run)
    exec go run -tags "input_json sink_stdout" ./cmd/sparkbridged -config configs/sparkbridge.yaml
    ;;
  *)
    cat <<'EOF'
SparkBridge helper

Usage:
  scripts/dev.sh test       Run the full race test suite
  scripts/dev.sh bench      Run benchmarks
  scripts/dev.sh arch-check Run the architecture check
  scripts/dev.sh run        Start the daemon with configs/sparkbridge.yaml
EOF
    ;;
esac
