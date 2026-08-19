#!/usr/bin/env python3
"""Embedding A/B retrieval metrics independent of QA accuracy.

Assumes one frozen ingest already landed. Re-embed between arms with
`go run ./cmd/reembed`, then run this scorer.

Reports gold-object recall@k / MRR from /memories/search contents,
plus dense vs lexical admission and candidate token counts.
"""
from __future__ import annotations

import argparse
import json
import os
import pathlib
import sys
import time

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "evals"))

from httputil import get_json  # noqa: E402
from public.backends.brainy import BrainyBackend  # noqa: E402
from public.locomo.dataset import (  # noqa: E402
    ensure_dataset,
    load_conversations,
    scored_question_pool,
    stratified_questions,
)
from public.oracle import gold_semantically_in_texts  # noqa: E402
from public.runtime_manifest import fetch_runtime  # noqa: E402


def _hit_at(contents: list[str], gold: str, k: int) -> bool:
    return gold_semantically_in_texts(gold, contents[:k])


def _mrr(contents: list[str], gold: str) -> float:
    for i, text in enumerate(contents, start=1):
        if gold_semantically_in_texts(gold, text):
            return 1.0 / i
    if gold_semantically_in_texts(gold, contents):
        return 1.0 / max(len(contents), 1)
    return 0.0


def _is_dense(explain: dict) -> bool:
    if not explain:
        return False
    if explain.get("embedding_similarity") or explain.get("signal_semantic"):
        return True
    basis = str(explain.get("ranking_basis") or "")
    return basis in {"hybrid_embedding", "hybrid"}


def _token_count(text: str) -> int:
    return len((text or "").split())


def run(args: argparse.Namespace) -> dict:
    dataset_path, dataset_sha = ensure_dataset()
    conversations = load_conversations(dataset_path)
    if args.conversations:
        conversations = conversations[: args.conversations]
    pool = scored_question_pool(conversations)
    sample = stratified_questions(pool, args.stratified, args.seed)
    base_url = (args.base_url or os.environ.get("BRAINY_BASE_URL") or "http://127.0.0.1:8080").rstrip("/")
    backend = BrainyBackend(base_url, tenant_prefix=args.tenant_prefix, async_ingest=False)
    runtime = {}
    try:
        runtime = fetch_runtime(base_url)
    except Exception as exc:  # noqa: BLE001
        runtime = {"error": str(exc)}

    ks = (10, 30, 100, 200)
    hits = {k: 0 for k in ks}
    mrr_sum = 0.0
    dense_n = 0
    lexical_n = 0
    candidate_n = 0
    context_tokens = 0
    rows = []
    for item in sample:
        user_id = str(item["sample_id"])
        tenant = backend._tenant(user_id)
        t0 = time.perf_counter()
        body = get_json(
            base_url,
            "/memories/search",
            {
                "tenant_id": tenant,
                "subject_id": user_id,
                "q": item["question"],
                "limit": str(max(ks)),
            },
            timeout=120,
        )
        latency_ms = (time.perf_counter() - t0) * 1000
        results = list(body.get("results") or [])
        contents = [str(r.get("content") or "") for r in results]
        gold = str(item.get("answer") or "")
        dense_hits = 0
        lexical_hits = 0
        for result in results:
            explain = result.get("explain") or {}
            if _is_dense(explain):
                dense_hits += 1
            else:
                lexical_hits += 1
        dense_n += dense_hits
        lexical_n += lexical_hits
        candidate_n += len(results)
        context_tokens += sum(_token_count(c) for c in contents)
        row = {
            "id": f"{item['sample_id']}-{item['id']}",
            "group": item.get("group"),
            "latency_ms": latency_ms,
            "n": len(contents),
            "dense_admitted": dense_hits,
            "lexical_admitted": lexical_hits,
            "context_tokens": sum(_token_count(c) for c in contents),
        }
        for k in ks:
            ok = _hit_at(contents, gold, k)
            row[f"recall@{k}"] = ok
            if ok:
                hits[k] += 1
        rr = _mrr(contents, gold)
        row["rr"] = rr
        mrr_sum += rr
        rows.append(row)

    n = max(len(sample), 1)
    summary = {
        "arm": args.arm,
        "dataset_sha256": dataset_sha,
        "n": len(sample),
        "seed": args.seed,
        "tenant_prefix": args.tenant_prefix,
        "runtime": {
            "signatures": (runtime.get("signatures") or {}),
            "ann": runtime.get("ann"),
            "embedder": ((runtime.get("api") or {}).get("embedder") or {}),
        },
        "recall": {f"@{k}": hits[k] / n for k in ks},
        "mrr": mrr_sum / n,
        "admission": {
            "dense_mean": dense_n / n,
            "lexical_mean": lexical_n / n,
            "candidates_mean": candidate_n / n,
            "context_tokens_mean": context_tokens / n,
        },
        "items": rows,
    }
    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    print(
        f"arm={args.arm} n={len(sample)} "
        + " ".join(f"r@{k}={hits[k]/n:.3f}" for k in ks)
        + f" mrr={mrr_sum/n:.3f}"
        + f" dense={dense_n/n:.1f} lex={lexical_n/n:.1f} tok={context_tokens/n:.0f}",
        flush=True,
    )
    print(f"wrote {out}", flush=True)
    return summary


def main() -> int:
    p = argparse.ArgumentParser(description="Retrieval-only embedding A/B scorer")
    p.add_argument("--base-url", default="")
    p.add_argument("--arm", default="current")
    p.add_argument("--tenant-prefix", default="locomo-ab")
    p.add_argument("--conversations", type=int, default=10)
    p.add_argument("--stratified", type=int, default=180)
    p.add_argument("--seed", type=int, default=1)
    p.add_argument(
        "--out",
        default=str(ROOT / "docs" / "benchmarks" / "runs" / "embedding-ab.json"),
    )
    run(p.parse_args())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
