#!/usr/bin/env python3
"""Full LOCOMO publish-stack runner (master-plan E1).

Features:
- 1–10 conversations, question budget, resume from partial JSON
- Multiple seeds (re-runs) for confidence intervals
- Pluggable judge/answerer models
- Writes UnifiedResult + manifest + markdown report

This is the dry-runable publish path; prefer staging Brainy + pinned LLM.
"""
from __future__ import annotations

import argparse
import json
import pathlib
import sys
import uuid

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "evals"))

from public.locomo.run_smoke import run as run_smoke  # noqa: E402


def main() -> int:
    parser = argparse.ArgumentParser(description="Full LOCOMO publish runner")
    parser.add_argument("--base-url", default="")
    parser.add_argument("--conversations", type=int, default=10)
    parser.add_argument("--questions", type=int, default=0, help="0 = all questions in selected convs")
    parser.add_argument("--stratified", type=int, default=0)
    parser.add_argument("--seed", type=int, default=1)
    parser.add_argument("--top-k", type=int, default=None)
    parser.add_argument(
        "--eval-lane",
        choices=("product-recall", "industry-search"),
        default="",
        help="R10 freeze label. industry-search defaults --top-k 200 when --top-k is omitted.",
    )
    parser.add_argument("--seeds", type=int, default=1)
    parser.add_argument("--answerer-model", default="")
    parser.add_argument("--judge-model", default="")
    parser.add_argument("--llm-base-url", default="")
    parser.add_argument("--llm-api-key", default="")
    parser.add_argument("--sync-ingest", action="store_true")
    parser.add_argument("--async-timeout", type=float, default=1800.0)
    parser.add_argument("--system", choices=("brainy", "mem0"), default="brainy")
    parser.add_argument("--run-prefix", default="locomo-full")
    parser.add_argument(
        "--out-dir",
        default=str(ROOT / "docs" / "benchmarks" / "runs"),
    )
    parser.add_argument(
        "--resume",
        action="store_true",
        help="Skip a seed if run JSON already exists under out-dir",
    )
    args = parser.parse_args()

    out_dir = pathlib.Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    summaries = []

    for seed in range(args.seeds):
        run_id = f"{args.run_prefix}-s{seed}-{uuid.uuid4().hex[:6]}"
        result_path = out_dir / f"{run_id}.json"
        if args.resume and result_path.exists():
            print(f"resume skip {result_path}", flush=True)
            continue
        # Build a Namespace compatible with run_smoke.run
        ns = argparse.Namespace(
            base_url=args.base_url,
            conversations=args.conversations,
            questions=args.questions if args.questions > 0 else None,
            stratified=getattr(args, "stratified", 0),
            seed=getattr(args, "seed", 1),
            top_k=args.top_k,
            eval_lane=getattr(args, "eval_lane", ""),
            failure_ledger=str(pathlib.Path(args.out_dir) / "failure-ledger" / f"{run_id}.jsonl"),
            answerer_model=args.answerer_model,
            judge_model=args.judge_model,
            llm_base_url=args.llm_base_url,
            llm_api_key=args.llm_api_key,
            lexical_only=False,
            sync_ingest=args.sync_ingest,
            async_timeout=args.async_timeout,
            system=args.system,
            run_id=run_id,
            out_dir=str(out_dir),
            report=str(out_dir / f"{run_id}.md"),
        )
        print(f"=== seed {seed} run_id={run_id} ===", flush=True)
        result = run_smoke(ns)
        summaries.append(
            {
                "seed": seed,
                "run_id": run_id,
                "accuracy": result.metrics.overall_accuracy,
                "correct": result.metrics.correct,
                "total": result.metrics.total,
                "by_group": {
                    k: {"correct": v.correct, "total": v.total}
                    for k, v in (result.metrics.by_group or {}).items()
                },
            }
        )

    summary_path = out_dir / f"{args.run_prefix}-summary.json"
    summary_path.write_text(json.dumps({"runs": summaries}, indent=2) + "\n", encoding="utf-8")
    print(f"wrote {summary_path}", flush=True)
    for row in summaries:
        print(
            f"seed={row['seed']} acc={row['accuracy']:.3f} ({row['correct']}/{row['total']})",
            flush=True,
        )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
