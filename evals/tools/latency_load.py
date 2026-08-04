#!/usr/bin/env python3
"""W6 latency load report against staging search (master-plan).

Ingests a synthetic subject, then hammers /memories/search at concurrency N
and reports p50/p95/p99.
"""
from __future__ import annotations

import argparse
import concurrent.futures
import json
import os
import pathlib
import statistics
import sys
import time
import uuid

ROOT = pathlib.Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))

from httputil import get_json, post_json  # noqa: E402


def percentile(xs: list[float], p: float) -> float:
    if not xs:
        return 0.0
    ys = sorted(xs)
    k = (len(ys) - 1) * (p / 100.0)
    f = int(k)
    c = min(f + 1, len(ys) - 1)
    if f == c:
        return ys[f]
    return ys[f] + (ys[c] - ys[f]) * (k - f)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--base-url", default=os.environ.get("BRAINY_BASE_URL", "http://127.0.0.1:8080"))
    ap.add_argument("--memories", type=int, default=200)
    ap.add_argument("--queries", type=int, default=100)
    ap.add_argument("--concurrency", type=int, default=8)
    ap.add_argument("--top-k", type=int, default=30)
    ap.add_argument("--out", default="")
    args = ap.parse_args()
    base = args.base_url.rstrip("/")
    tenant = f"load-{uuid.uuid4().hex[:8]}"
    subject = "load-subj"

    print(f"ingest {args.memories} memories tenant={tenant}", flush=True)
    for i in range(args.memories):
        post_json(
            base,
            "/ingest",
            {
                "tenant_id": tenant,
                "subject_id": subject,
                "source_type": "conversation",
                "messages": [
                    {
                        "role": "user",
                        "content": (
                            f"Fact {i}: the project Orion uses Postgres and prefers "
                            f"async extract; ticket-{i % 17} needs escalation on SLA breach."
                        ),
                    }
                ],
            },
            timeout=60,
        )

    queries = [
        "Orion Postgres preference",
        "ticket escalation SLA",
        "async extract",
        "project Orion",
        "customer escalation rule",
    ]
    latencies: list[float] = []
    errors = 0

    def one(qi: int) -> float:
        q = queries[qi % len(queries)]
        t0 = time.perf_counter()
        get_json(
            base,
            "/memories/search",
            {
                "tenant_id": tenant,
                "subject_id": subject,
                "q": q,
                "limit": str(args.top_k),
            },
            timeout=60,
        )
        return (time.perf_counter() - t0) * 1000.0

    print(f"search load queries={args.queries} concurrency={args.concurrency}", flush=True)
    with concurrent.futures.ThreadPoolExecutor(max_workers=args.concurrency) as pool:
        futs = [pool.submit(one, i) for i in range(args.queries)]
        for fut in concurrent.futures.as_completed(futs):
            try:
                latencies.append(fut.result())
            except Exception as exc:
                errors += 1
                print(f"err {exc}", flush=True)

    report = {
        "tenant_id": tenant,
        "memories": args.memories,
        "queries": args.queries,
        "concurrency": args.concurrency,
        "top_k": args.top_k,
        "errors": errors,
        "n": len(latencies),
        "p50_ms": round(percentile(latencies, 50), 1),
        "p95_ms": round(percentile(latencies, 95), 1),
        "p99_ms": round(percentile(latencies, 99), 1),
        "mean_ms": round(statistics.fmean(latencies), 1) if latencies else None,
        "slo_p50_met": (percentile(latencies, 50) <= 1000) if latencies else False,
        "slo_p95_met": (percentile(latencies, 95) <= 2500) if latencies else False,
    }
    print(json.dumps(report, indent=2), flush=True)
    if args.out:
        pathlib.Path(args.out).write_text(json.dumps(report, indent=2) + "\n")
    return 0 if errors == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
