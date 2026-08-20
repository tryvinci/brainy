#!/usr/bin/env bash
# Run OpenAI embedding A/B arms on frozen tenant integrity-s0-1, then restore BGE.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
source /tmp/integrity-stack.env
BASE_URL="${1:-http://127.0.0.1:18100}"
TENANT="${2:-integrity-s0-1}"
DATE="${3:-20260820}"

if [[ -z "${OPENAI_API_KEY:-}" ]]; then
  echo "OPENAI_API_KEY is required" >&2
  exit 1
fi

BGE_BASE="$BRAINY_EMBEDDING_BASE_URL"
BGE_KEY="$BRAINY_EMBEDDING_API_KEY"
BGE_MODEL="$BRAINY_EMBEDDING_MODEL"

restart_api() {
  tmux -f /exec-daemon/tmux.portal.conf send-keys -t integrity-api:0.0 C-c
  sleep 2
  tmux -f /exec-daemon/tmux.portal.conf send-keys -t integrity-api:0.0 \
    "source /tmp/integrity-stack.env && /tmp/brainy-api 2>&1 | tee -a /tmp/integrity-api.log" C-m
  sleep 5
  curl -sf "$BASE_URL/healthz" >/dev/null
}

run_arm() {
  local arm="$1" model="$2"
  export BRAINY_EMBEDDING_BASE_URL="https://api.openai.com/v1"
  export BRAINY_EMBEDDING_API_KEY="$OPENAI_API_KEY"
  export BRAINY_EMBEDDING_MODEL="$model"
  export BRAINY_EMBEDDING_DIMENSIONS=768
  echo "=== reembed $arm ($model @768) ==="
  (cd "$ROOT" && go run ./cmd/reembed)
  restart_api
  echo "=== score $arm ==="
  (cd "$ROOT" && unset BRAINY_API_KEY && python3 evals/public/embedding_ab.py \
    --base-url "$BASE_URL" \
    --tenant-prefix "$TENANT" \
    --arm "$arm" \
    --out "$ROOT/docs/benchmarks/artifacts/embedding-ab-${arm}-${DATE}.json")
}

restore_bge() {
  export BRAINY_EMBEDDING_BASE_URL="$BGE_BASE"
  export BRAINY_EMBEDDING_API_KEY="$BGE_KEY"
  export BRAINY_EMBEDDING_MODEL="$BGE_MODEL"
  unset BRAINY_EMBEDDING_DIMENSIONS
  echo "=== restore BGE reembed ==="
  (cd "$ROOT" && go run ./cmd/reembed)
  restart_api
  curl -sf "$BASE_URL/runtime" | python3 -c "import sys,json; r=json.load(sys.stdin); print('ann', r.get('ann')); print('sig', r.get('signatures'))"
}

run_arm "openai-large-768" "text-embedding-3-large"
run_arm "openai-small-768" "text-embedding-3-small"
restore_bge
