#!/usr/bin/env bash
# Score product n=1540 against an already-ingested tenant prefix.
# Always force a local BRAINY_BASE_URL (cloud injects staging).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
export PYTHONPATH="$ROOT/evals"
export BRAINY_BASE_URL="${BRAINY_BASE_URL:-http://127.0.0.1:18200}"
export BRAINY_USE_RECALL=1
export N1540_TENANT_PREFIX="${N1540_TENANT_PREFIX:-locomo-n1540-bbe55f7}"
RUN_ID="${RUN_ID:-locomo-full-n1540-bbe55f7-s0-complete}"
OUT="${OUT:-$ROOT/docs/benchmarks/runs}"
mkdir -p "$OUT/failure-ledger"

echo "BRAINY_BASE_URL=$BRAINY_BASE_URL USE_RECALL=$BRAINY_USE_RECALL prefix=$N1540_TENANT_PREFIX run_id=$RUN_ID" >&2
case "$BRAINY_BASE_URL" in
  http://127.0.0.1:*|http://localhost:*) ;;
  *)
    echo "refusing to score: BRAINY_BASE_URL is not local ($BRAINY_BASE_URL)" >&2
    exit 1
    ;;
esac

python3 -m public.locomo.run_smoke \
  --base-url "$BRAINY_BASE_URL" \
  --conversations 10 \
  --questions 0 \
  --eval-lane product-recall \
  --tenant-prefix "$N1540_TENANT_PREFIX" \
  --skip-ingest \
  --async-timeout 3600 \
  --run-id "$RUN_ID" \
  --out-dir "$OUT" \
  --report "$OUT/${RUN_ID}.md" \
  --failure-ledger "$OUT/failure-ledger/${RUN_ID}.jsonl"
