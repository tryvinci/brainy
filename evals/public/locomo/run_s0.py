#!/usr/bin/env python3
"""S0 dual-lane stratified baseline (measurement only).

Runs product-recall then industry-search on the same stratified sample and
writes a stage-oracle ledger histogram. Does not change product code.

Usage:
  python -m public.locomo.run_s0 --base-url http://127.0.0.1:18090 --stratified 180 --seed 1
"""
from __future__ import annotations

import argparse
import json
import pathlib
import sys
import uuid

ROOT = pathlib.Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "evals"))

from public.locomo.ledger_summary import summarize_ledger  # noqa: E402
from public.locomo.run_smoke import run as run_smoke  # noqa: E402


def _ns(
    args: argparse.Namespace,
    *,
    lane: str,
    run_id: str,
    tenant_prefix: str,
    skip_ingest: bool,
) -> argparse.Namespace:
    out_dir = pathlib.Path(args.out_dir)
    return argparse.Namespace(
        base_url=args.base_url,
        conversations=args.conversations,
        questions=None,
        stratified=args.stratified,
        seed=args.seed,
        top_k=args.top_k,
        eval_lane=lane,
        answerer_model=args.answerer_model,
        judge_model=args.judge_model,
        llm_base_url=args.llm_base_url,
        llm_api_key=args.llm_api_key,
        lexical_only=args.lexical_only,
        sync_ingest=args.sync_ingest,
        async_timeout=args.async_timeout,
        system="brainy",
        run_id=run_id,
        out_dir=str(out_dir),
        report=str(out_dir / f"{run_id}.md"),
        failure_ledger=str(out_dir / "failure-ledger" / f"{run_id}.jsonl"),
        tenant_prefix=tenant_prefix,
        skip_ingest=skip_ingest,
        fail_closed=bool(getattr(args, "fail_closed", False)),
    )


def main() -> int:
    parser = argparse.ArgumentParser(description="S0 dual-lane stratified LoCoMo baseline")
    parser.add_argument("--base-url", default="")
    parser.add_argument("--conversations", type=int, default=10)
    parser.add_argument("--stratified", type=int, default=180)
    parser.add_argument("--seed", type=int, default=1)
    parser.add_argument("--top-k", type=int, default=None)
    parser.add_argument("--answerer-model", default="")
    parser.add_argument("--judge-model", default="")
    parser.add_argument("--llm-base-url", default="")
    parser.add_argument("--llm-api-key", default="")
    parser.add_argument("--lexical-only", action="store_true")
    parser.add_argument("--sync-ingest", action="store_true")
    parser.add_argument("--async-timeout", type=float, default=1800.0)
    parser.add_argument("--run-prefix", default="locomo-s0")
    parser.add_argument(
        "--out-dir",
        default=str(ROOT / "docs" / "benchmarks" / "runs"),
    )
    parser.add_argument(
        "--lanes",
        default="product-recall,industry-search",
        help="Comma-separated eval lanes",
    )
    parser.add_argument(
        "--tenant-prefix",
        default="",
        help="Reuse a frozen ingest. Empty generates integrity-s0-<seed>-<nonce>.",
    )
    parser.add_argument(
        "--skip-ingest",
        action="store_true",
        help="Score without ingesting (requires --tenant-prefix)",
    )
    parser.add_argument(
        "--fail-closed",
        action="store_true",
        help="Abort if runtime signatures mismatch, fallbacks>0, or ANN inactive",
    )
    args = parser.parse_args()

    lanes = [p.strip() for p in args.lanes.split(",") if p.strip()]
    tenant_prefix = (args.tenant_prefix or "").strip() or f"integrity-s0-{args.seed}-{uuid.uuid4().hex[:6]}"
    summaries = []
    for i, lane in enumerate(lanes):
        run_id = f"{args.run_prefix}-{lane}-s{args.seed}-{uuid.uuid4().hex[:6]}"
        skip = bool(args.skip_ingest or i > 0)
        print(
            f"=== S0 lane={lane} run_id={run_id} stratified={args.stratified} "
            f"tenant_prefix={tenant_prefix} skip_ingest={skip} ===",
            flush=True,
        )
        result = run_smoke(
            _ns(
                args,
                lane=lane,
                run_id=run_id,
                tenant_prefix=tenant_prefix,
                skip_ingest=skip,
            )
        )
        ledger = pathlib.Path(args.out_dir) / "failure-ledger" / f"{run_id}.jsonl"
        hist = summarize_ledger(ledger) if ledger.exists() else {}
        summaries.append(
            {
                "lane": lane,
                "run_id": run_id,
                "accuracy": result.metrics.overall_accuracy,
                "correct": result.metrics.correct,
                "total": result.metrics.total,
                "by_group": {
                    k: {"correct": v.correct, "total": v.total}
                    for k, v in (result.metrics.by_group or {}).items()
                },
                "failure_histogram": hist,
            }
        )

    out_dir = pathlib.Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    summary_path = out_dir / f"{args.run_prefix}-summary.json"
    summary_path.write_text(
        json.dumps({"tenant_prefix": tenant_prefix, "runs": summaries}, indent=2) + "\n",
        encoding="utf-8",
    )
    print(f"wrote {summary_path}", flush=True)
    for row in summaries:
        print(
            f"lane={row['lane']} acc={row['accuracy']:.3f} "
            f"({row['correct']}/{row['total']}) hist={row['failure_histogram']}",
            flush=True,
        )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
