#!/usr/bin/env bash
# Start pinned Letta REST server for OpMem external lane (no Docker).
set -euo pipefail
export PATH="${HOME}/.local/bin:${PATH}"
LETTA_VERSION="${OPMEM_LETTA_SERVER_VERSION:-0.11.7}"
LETTA_CLIENT_VERSION="${OPMEM_LETTA_CLIENT_VERSION:-0.1.307}"
PORT="${LETTA_PORT:-8283}"
HOST="${LETTA_HOST:-127.0.0.1}"
LOG="${OPMEM_LETTA_SERVER_LOG:-/tmp/opmem-letta-server.log}"

python3 -m pip install -q "letta==${LETTA_VERSION}" "letta-client==${LETTA_CLIENT_VERSION}" sqlite-vec

if ! command -v letta >/dev/null; then
  echo "letta CLI not on PATH; ensure ~/.local/bin is exported" >&2
  exit 1
fi

if curl -sf "http://${HOST}:${PORT}/v1/health" >/dev/null 2>&1; then
  echo "Letta already listening on http://${HOST}:${PORT}"
  exit 0
fi

nohup letta server --type rest --port "${PORT}" --host "${HOST}" --no-secure >"${LOG}" 2>&1 &
echo "Started Letta ${LETTA_VERSION} at http://${HOST}:${PORT} (log: ${LOG})"

for _ in $(seq 1 30); do
  if curl -sfL "http://${HOST}:${PORT}/v1/health" | grep -q ok; then
    echo "Letta health OK"
    exit 0
  fi
  sleep 1
done
echo "Letta failed to become healthy; see ${LOG}" >&2
exit 1
