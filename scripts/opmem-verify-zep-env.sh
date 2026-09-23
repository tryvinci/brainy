#!/usr/bin/env bash
# Verify Zep lane prerequisites without printing secrets.
set -euo pipefail
echo "CLOUD_AGENT_INJECTED_SECRET_NAMES:"
echo "${CLOUD_AGENT_INJECTED_SECRET_NAMES:-}" | tr ',' '\n' | sort
if [ -n "${ZEP_API_KEY:-}" ]; then
  echo "ZEP_API_KEY: present (length=${#ZEP_API_KEY})"
else
  echo "ZEP_API_KEY: MISSING"
  exit 1
fi
if ! echo "${CLOUD_AGENT_INJECTED_SECRET_NAMES:-}" | tr ',' '\n' | grep -qx 'ZEP_API_KEY'; then
  echo "WARNING: ZEP_API_KEY not listed in CLOUD_AGENT_INJECTED_SECRET_NAMES"
fi
exit 0
