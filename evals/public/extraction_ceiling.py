#!/usr/bin/env python3
"""Extraction ceiling: deterministic vs provider, semantic gold coverage.

Ingests the same LoCoMo sessions twice (two tenant prefixes):
  - deterministic: --sync-ingest (no provider)
  - provider: async worker path (strict recommended)

Scores representation-oracle semantic coverage, not QA accuracy.
"""
from __future__ import annotations

import argparse
import json
import os
import pathlib
import sys

ROOT = pathlib.Path(__file__).resolve().parents[2]
EVALS = ROOT / "evals"
sys.path.insert(0, str(EVALS))

from public.backends.brainy import BrainyBackend  # noqa: E402
from public.locomo.dataset import (  # noqa: E402
    ensure_dataset,
    iter_sessions,
    load_conversations,
    scored_question_pool,
    stratified_questions,
)
from public.locomo.run_smoke import ingest_conversation  # noqa: E402
from public.oracle import gold_semantically_in_texts  # noqa: E402
from public.stage_oracle import oracle_recall_request, post_recall  # noqa: E402


def coverage_for_arm(
    *,
    base_url: str,
    tenant_prefix: str,
    conversations: list[dict],
    sample: list[dict],
    async_ingest: bool,
    timeout: float,
    skip_ingest: bool = False,
) -> dict:
    backend = BrainyBackend(
        base_url,
        tenant_prefix=tenant_prefix,
        async_ingest=async_ingest,
        async_timeout_s=timeout,
    )
    ingested = 0
    keep = {int(row["conv_idx"]) for row in sample}
    if not skip_ingest:
        for idx in sorted(keep):
            conv = conversations[idx]
            sample_id = str(conv.get("sample_id") or f"c{idx}")
            ingested += ingest_conversation(backend, sample_id, iter_sessions(conv))
            print(f"[{tenant_prefix} {sample_id}] ingested {ingested} cumulative", flush=True)
    else:
        print(f"[{tenant_prefix}] skip-ingest coverage only", flush=True)

    hits = 0
    rows = []
    for item in sample:
        tenant_id = backend._tenant(str(item["sample_id"]))
        body = oracle_recall_request(tenant_id, str(item["sample_id"]), item["question"], "representation")
        resp = post_recall(base_url, body)
        explain = resp.get("explain") or {}
        blob = str(explain.get("oracle_fact_blob") or "")
        ok = gold_semantically_in_texts(item.get("answer") or "", blob)
        hits += int(ok)
        rows.append(
            {
                "id": f"{item['sample_id']}-{item['id']}",
                "group": item.get("group"),
                "covered": ok,
                "facts": int(explain.get("oracle_fact_count") or 0),
            }
        )
    n = max(len(sample), 1)
    return {
        "tenant_prefix": tenant_prefix,
        "async_ingest": async_ingest,
        "ingested_turns": ingested,
        "skip_ingest": skip_ingest,
        "covered": hits,
        "n": len(sample),
        "coverage": hits / n,
        "items": rows,
    }


def main() -> int:
    p = argparse.ArgumentParser(description="Deterministic vs provider extraction ceiling")
    p.add_argument("--base-url", default="")
    p.add_argument("--conversations", type=int, default=10)
    p.add_argument("--stratified", type=int, default=180)
    p.add_argument("--seed", type=int, default=1)
    p.add_argument("--async-timeout", type=float, default=1800.0)
    p.add_argument("--skip-deterministic", action="store_true")
    p.add_argument("--skip-provider", action="store_true")
    p.add_argument("--skip-ingest", action="store_true")
    p.add_argument("--tenant-prefix-det", default="")
    p.add_argument("--tenant-prefix-prov", default="")
    p.add_argument(
        "--out",
        default=str(ROOT / "docs" / "benchmarks" / "runs" / "extraction-ceiling.json"),
    )
    args = p.parse_args()
    base_url = (args.base_url or os.environ.get("BRAINY_BASE_URL") or "http://127.0.0.1:8080").rstrip("/")
    dataset_path, dataset_sha = ensure_dataset()
    conversations = load_conversations(dataset_path)
    if args.conversations:
        conversations = conversations[: args.conversations]
    sample = stratified_questions(scored_question_pool(conversations), args.stratified, args.seed)

    arms = {}
    if not args.skip_deterministic:
        print("=== deterministic sync ingest ===", flush=True)
        # Force sync path: BrainyBackend async_ingest=False uses /ingest.
        arms["deterministic"] = coverage_for_arm(
            base_url=base_url,
            tenant_prefix=args.tenant_prefix_det or f"ceil-det-{args.seed}",
            conversations=conversations,
            sample=sample,
            async_ingest=False,
            timeout=args.async_timeout,
            skip_ingest=args.skip_ingest,
        )
    if not args.skip_provider:
        print("=== provider async ingest ===", flush=True)
        arms["provider"] = coverage_for_arm(
            base_url=base_url,
            tenant_prefix=args.tenant_prefix_prov or f"ceil-prov-{args.seed}",
            conversations=conversations,
            sample=sample,
            async_ingest=True,
            timeout=args.async_timeout,
            skip_ingest=args.skip_ingest,
        )
    out = {
        "dataset_sha256": dataset_sha,
        "n": len(sample),
        "seed": args.seed,
        "arms": arms,
    }
    path = pathlib.Path(args.out)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(out, indent=2) + "\n", encoding="utf-8")
    for name, arm in arms.items():
        print(f"{name} coverage={arm['coverage']:.3f} ({arm['covered']}/{arm['n']})", flush=True)
    print(f"wrote {path}", flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
