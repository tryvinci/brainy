#!/usr/bin/env python3
"""S6 freeze orchestrator (qualification, then one-shot full remasure).

Does not invent scores. Does not write SOTA / beats-Mem0 product copy.
Full n=1540 is once per freeze and should not share a worker queue with S0.

Usage:
  python -m public.locomo.run_s6 --base-url http://127.0.0.1:18090
  python -m public.locomo.run_s6 --full --lme20 --mem0 --base-url http://127.0.0.1:18090
"""
from __future__ import annotations

import argparse
import json
import os
import pathlib
import sys
import uuid

ROOT = pathlib.Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "evals"))

from public.locomo.ledger_summary import summarize_ledger  # noqa: E402
from public.locomo.run_smoke import run as run_smoke  # noqa: E402


def plan_steps(
    *,
    qualify: bool = True,
    full: bool = False,
    lme20: bool = False,
    mem0: bool = False,
) -> list[dict]:
    steps: list[dict] = []
    if qualify:
        steps.append(
            {
                "id": "s6a-product-3x90",
                "kind": "locomo",
                "lane": "product-recall",
                "conversations": 10,
                "questions": 90,
                "seeds": 3,
                "system": "brainy",
            }
        )
        steps.append(
            {
                "id": "s6a-industry-3x90",
                "kind": "locomo",
                "lane": "industry-search",
                "conversations": 10,
                "questions": 90,
                "seeds": 3,
                "system": "brainy",
            }
        )
    if full:
        steps.append(
            {
                "id": "s6b-product-full",
                "kind": "locomo",
                "lane": "product-recall",
                "conversations": 10,
                "questions": 0,
                "seeds": 1,
                "system": "brainy",
            }
        )
        steps.append(
            {
                "id": "s6b-industry-full",
                "kind": "locomo",
                "lane": "industry-search",
                "conversations": 10,
                "questions": 0,
                "seeds": 3,
                "system": "brainy",
            }
        )
    if lme20:
        steps.append(
            {
                "id": "s6b-lme20",
                "kind": "lme",
                "limit": 20,
                "seed": 1,
                "product_recall": True,
            }
        )
    if mem0:
        steps.append(
            {
                "id": "s6b-mem0-full",
                "kind": "locomo",
                "lane": "industry-search",
                "conversations": 10,
                "questions": 0,
                "seeds": 1,
                "system": "mem0",
            }
        )
    return steps


def _require_s0(out_dir: pathlib.Path, prefix: str) -> pathlib.Path:
    path = out_dir / f"{prefix}-summary.json"
    if not path.exists():
        raise SystemExit(
            f"S6 requires an S0 summary at {path}. Run "
            "`python -m public.locomo.run_s0` first, or pass --skip-s0-gate."
        )
    return path


def _locomo_ns(args: argparse.Namespace, step: dict, seed: int, run_id: str) -> argparse.Namespace:
    out_dir = pathlib.Path(args.out_dir)
    questions = step.get("questions") or 0
    return argparse.Namespace(
        base_url=args.base_url,
        conversations=int(step.get("conversations") or 10),
        questions=questions if questions > 0 else None,
        stratified=0,
        seed=seed,
        top_k=args.top_k,
        eval_lane=step.get("lane") or "",
        answerer_model=args.answerer_model,
        judge_model=args.judge_model,
        llm_base_url=args.llm_base_url,
        llm_api_key=args.llm_api_key,
        lexical_only=False,
        sync_ingest=args.sync_ingest,
        async_timeout=args.async_timeout,
        system=step.get("system") or "brainy",
        run_id=run_id,
        out_dir=str(out_dir),
        report=str(out_dir / f"{run_id}.md"),
        failure_ledger=str(out_dir / "failure-ledger" / f"{run_id}.jsonl"),
    )


def _run_locomo(args: argparse.Namespace, step: dict) -> list[dict]:
    if step.get("system") == "mem0" and not (
        os.environ.get("MEM0_API_KEY") or os.environ.get("MEM0_PLATFORM_API_KEY")
    ):
        print(f"skip {step['id']}: MEM0_API_KEY unset (not a score)", flush=True)
        return [{"id": step["id"], "skipped": "mem0-credentials"}]
    rows = []
    out_dir = pathlib.Path(args.out_dir)
    seeds = max(1, int(step.get("seeds") or 1))
    for seed in range(seeds):
        run_id = f"{args.run_prefix}-{step['id']}-s{seed}-{uuid.uuid4().hex[:6]}"
        print(f"=== S6 {step['id']} seed={seed} run_id={run_id} ===", flush=True)
        result = run_smoke(_locomo_ns(args, step, seed, run_id))
        ledger = out_dir / "failure-ledger" / f"{run_id}.jsonl"
        rows.append(
            {
                "id": step["id"],
                "seed": seed,
                "run_id": run_id,
                "accuracy": result.metrics.overall_accuracy,
                "correct": result.metrics.correct,
                "total": result.metrics.total,
                "by_group": {
                    k: {"correct": v.correct, "total": v.total}
                    for k, v in (result.metrics.by_group or {}).items()
                },
                "failure_histogram": summarize_ledger(ledger) if ledger.exists() else {},
            }
        )
    return rows


def _run_lme(args: argparse.Namespace, step: dict) -> list[dict]:
    from public.longmemeval.run import run as run_lme

    run_id = f"{args.run_prefix}-{step['id']}-{uuid.uuid4().hex[:6]}"
    print(f"=== S6 {step['id']} run_id={run_id} ===", flush=True)
    ns = argparse.Namespace(
        dataset=args.lme_dataset,
        base_url=args.base_url,
        limit=int(step.get("limit") or 20),
        seed=int(step.get("seed") or 1),
        top_k=30,
        sync_ingest=args.sync_ingest,
        publish=True,
        product_recall=bool(step.get("product_recall", True)),
        async_timeout=max(float(args.async_timeout), 3600.0),
        lexical_only=False,
        answerer_model=args.answerer_model,
        judge_model=args.judge_model,
        llm_base_url=args.llm_base_url,
        llm_api_key=args.llm_api_key,
        run_id=run_id,
        report=str(pathlib.Path(args.out_dir) / f"{run_id}.md"),
        out_dir=args.out_dir,
    )
    result = run_lme(ns)
    return [
        {
            "id": step["id"],
            "run_id": run_id,
            "accuracy": result.metrics.overall_accuracy,
            "correct": result.metrics.correct,
            "total": result.metrics.total,
            "by_group": {
                k: {"correct": v.correct, "total": v.total}
                for k, v in (result.metrics.by_group or {}).items()
            },
        }
    ]


def main() -> int:
    parser = argparse.ArgumentParser(description="S6 freeze orchestrator")
    parser.add_argument("--base-url", default="")
    parser.add_argument("--top-k", type=int, default=None)
    parser.add_argument("--answerer-model", default="")
    parser.add_argument("--judge-model", default="")
    parser.add_argument("--llm-base-url", default="")
    parser.add_argument("--llm-api-key", default="")
    parser.add_argument("--sync-ingest", action="store_true")
    parser.add_argument("--async-timeout", type=float, default=1800.0)
    parser.add_argument("--run-prefix", default="locomo-s6")
    parser.add_argument("--s0-prefix", default="locomo-s0-20260819")
    parser.add_argument(
        "--out-dir",
        default=str(ROOT / "docs" / "benchmarks" / "runs"),
    )
    parser.add_argument(
        "--lme-dataset",
        default="/workspace/datasets/longmemeval/longmemeval_s_cleaned.json",
    )
    parser.add_argument("--qualify", action=argparse.BooleanOptionalAction, default=True)
    parser.add_argument("--full", action="store_true", help="S6b n=1540 both lanes")
    parser.add_argument("--lme20", action="store_true", help="S6b LME-20 product /recall")
    parser.add_argument("--mem0", action="store_true", help="S6b Mem0 Platform same-pin")
    parser.add_argument("--skip-s0-gate", action="store_true")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    out_dir = pathlib.Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    if not args.skip_s0_gate and not args.dry_run:
        _require_s0(out_dir, args.s0_prefix)

    steps = plan_steps(
        qualify=bool(args.qualify),
        full=bool(args.full),
        lme20=bool(args.lme20),
        mem0=bool(args.mem0),
    )
    if args.dry_run:
        print(json.dumps({"steps": steps}, indent=2))
        return 0

    summaries: list[dict] = []
    for step in steps:
        kind = step.get("kind")
        if kind == "lme":
            summaries.extend(_run_lme(args, step))
        else:
            summaries.extend(_run_locomo(args, step))

    summary_path = out_dir / f"{args.run_prefix}-summary.json"
    summary_path.write_text(json.dumps({"runs": summaries}, indent=2) + "\n", encoding="utf-8")
    print(f"wrote {summary_path}", flush=True)
    for row in summaries:
        if row.get("skipped"):
            print(f"{row['id']} skipped={row['skipped']}", flush=True)
            continue
        print(
            f"{row['id']} acc={row['accuracy']:.3f} ({row['correct']}/{row['total']})",
            flush=True,
        )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
