#!/usr/bin/env bash
# Run OpMem manuscript experiments via embedded API (same stack as CI).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
SHA="$(git rev-parse HEAD)"
OUT_DIR="${1:-docs/benchmarks/artifacts/opmem-manuscript-${SHA:0:8}}"
mkdir -p "$OUT_DIR"

{
  echo "git_sha=$SHA"
  echo "date_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "go test ./internal/api/ -run TestOpMemBenchmarkAgainstHTTPServer -count=1"
  echo "OPMEM_MANUSCRIPT_OUT=$OUT_DIR go test ./internal/api/ -run TestOpMemPaperExperiments -count=1"
} > "$OUT_DIR/commands.log"

go test ./internal/api/ -run TestOpMemBenchmarkAgainstHTTPServer -count=1 2>&1 | tee "$OUT_DIR/go-test-opmem.log"

export OPMEM_MANUSCRIPT_OUT="$OUT_DIR"
go test ./internal/api/ -run TestOpMemPaperExperiments -count=1 2>&1 | tee "$OUT_DIR/go-test-paper.log"

# Mem0 Platform counter-run when MEM0_API_KEY is set (optional; not required for CI).
if [[ -n "${MEM0_API_KEY:-}" ]]; then
  echo "mem0 run: python3 evals/run_opmem.py --systems mem0 --json-out $OUT_DIR/opmem-mem0-rerun.json" >> "$OUT_DIR/commands.log"
  python3 evals/run_opmem.py --systems mem0 --json-out "$OUT_DIR/opmem-mem0-rerun.json" 2>&1 | tee "$OUT_DIR/mem0-rerun.log" || true
fi

echo "Artifacts written to $OUT_DIR"
