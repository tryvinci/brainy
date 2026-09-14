#!/usr/bin/env bash
# Protocol-v2 skip-ingest score of the restored n=1540 tenant.
# Incremental JSONL lives under /opt/cursor/artifacts/eval-runs (never /tmp).
# Always force a local BRAINY_BASE_URL (cloud injects staging).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
export PYTHONPATH="$ROOT/evals"
export BRAINY_BASE_URL="${BRAINY_BASE_URL:-http://127.0.0.1:18200}"
export BRAINY_USE_RECALL=1
export BRAINY_EVAL_PROTOCOL=eval-protocol-v2
export N1540_TENANT_PREFIX="${N1540_TENANT_PREFIX:-locomo-n1540-bbe55f7}"
RUN_ID="${RUN_ID:-locomo-full-n1540-bbe55f7-protocol-v2}"
ARTIFACT_ROOT="${ARTIFACT_ROOT:-/opt/cursor/artifacts/eval-runs}"

echo "BRAINY_BASE_URL=$BRAINY_BASE_URL USE_RECALL=$BRAINY_USE_RECALL prefix=$N1540_TENANT_PREFIX run_id=$RUN_ID" >&2
case "$BRAINY_BASE_URL" in
  http://127.0.0.1:*|http://localhost:*) ;;
  *)
    echo "refusing to score: BRAINY_BASE_URL is not local ($BRAINY_BASE_URL)" >&2
    exit 1
    ;;
esac
case "$ARTIFACT_ROOT" in
  /tmp|/*/tmp/*)
    echo "refusing to store qualification artifacts under /tmp" >&2
    exit 1
    ;;
esac

mkdir -p "$ARTIFACT_ROOT"
python3 -m public.locomo.run_staged \
  --base-url "$BRAINY_BASE_URL" \
  --run-id "$RUN_ID" \
  --tenant-prefix "$N1540_TENANT_PREFIX" \
  --conversations 10 \
  --skip-ingest \
  --async-timeout 3600 \
  --top-k 30 \
  --artifact-root "$ARTIFACT_ROOT"
