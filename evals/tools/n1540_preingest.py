#!/usr/bin/env python3
"""Enqueue all 10 LoCoMo conversations without per-conv job wait.

Different convs are different subjects, so workers can extract them in
parallel (FIFO is per tenant+subject). Score later with run_smoke --skip-ingest.

  N1540_TENANT_PREFIX  default locomo-n1540-bbe55f7
  BRAINY_BASE_URL      must be the local API, not staging
"""
from __future__ import annotations

import os
import sys
import time

sys.path.insert(0, str(__import__("pathlib").Path(__file__).resolve().parents[1]))

from public.backends.brainy import BrainyBackend
from public.locomo.dataset import ensure_dataset, iter_sessions, load_conversations
from public.locomo.run_smoke import ingest_conversation

PREFIX = os.environ.get("N1540_TENANT_PREFIX", "locomo-n1540-bbe55f7")
BASE = os.environ.get("BRAINY_BASE_URL", "http://127.0.0.1:18200")


def main() -> int:
    path, sha = ensure_dataset()
    convs = load_conversations(path)[:10]
    backend = BrainyBackend(
        BASE,
        tenant_prefix=PREFIX,
        async_ingest=True,
        async_timeout_s=float(os.environ.get("N1540_ASYNC_TIMEOUT", "86400")),
        async_poll_s=2.0,
    )
    print(f"preingest prefix={PREFIX} base={BASE} dataset={sha} convs={len(convs)}", flush=True)
    t0 = time.time()
    for conv in convs:
        sample_id = str(conv.get("sample_id") or "c?")
        sessions = iter_sessions(conv)
        n = ingest_conversation(backend, sample_id, sessions, chunk=8, wait_jobs=False)
        print(f"  enqueued {sample_id} turns={n} elapsed={time.time()-t0:.0f}s", flush=True)
    print(f"enqueue_done elapsed={time.time()-t0:.0f}s", flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
