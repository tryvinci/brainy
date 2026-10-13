#!/usr/bin/env python3
"""Dry-run all OpMem external lanes: operation counts + token/cost worksheet (no vendor APIs)."""
from __future__ import annotations

import argparse
import json
import pathlib
import sys

ROOT = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(ROOT))

from opmem_external import LANE_ADAPTERS  # noqa: E402
from opmem_external.cost_budget import build_lane_budgets, build_worksheet_from_budgets  # noqa: E402
from opmem_external.worksheet import build_worksheet, write_worksheet  # noqa: E402
from run_opmem import run_task  # noqa: E402


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--fixture-dir", default="fixtures/opmem")
    parser.add_argument("--out-dir", default="docs/benchmarks/artifacts/opmem-external-lanes")
    parser.add_argument("--run-id", default="opmem-dryrun-0001")
    args = parser.parse_args()

    fixture_dir = pathlib.Path(args.fixture_dir)
    if not fixture_dir.is_absolute():
        fixture_dir = ROOT.parent / fixture_dir
    tasks = [json.loads(p.read_text(encoding="utf-8")) for p in sorted(fixture_dir.glob("*.json"))]
    step_count = sum(len(t["steps"]) for t in tasks)

    out_dir = pathlib.Path(args.out_dir)
    if not out_dir.is_absolute():
        out_dir = ROOT.parent / out_dir
    out_dir.mkdir(parents=True, exist_ok=True)

    lane_stats = []
    per_lane_results = {}
    for lane_name, factory in sorted(LANE_ADAPTERS.items()):
        adapter = factory(dry_run=True, run_id=args.run_id)
        results = []
        infra = 0
        for task in tasks:
            try:
                outcome = run_task(adapter, task)
                results.append({"task": task["name"], "outcome": outcome})
            except Exception as exc:  # noqa: BLE001
                infra += 1
                results.append({"task": task["name"], "error": f"{type(exc).__name__}: {exc}"})
        per_lane_results[lane_name] = {
            "infrastructure_errors": infra,
            "results": results,
        }
        log_path = out_dir / f"raw-{lane_name}-dry-run.jsonl"
        adapter.write_raw_log(log_path.with_suffix(".json"))
        lane_stats.append(adapter.dry_run_stats())

    worksheet = build_worksheet(lane_stats, task_count=len(tasks), step_count=step_count)
    write_worksheet(out_dir / "cost-worksheet-dry-run.json", worksheet)
    budgets = build_lane_budgets(tasks)
    payload_ws = build_worksheet_from_budgets(
        budgets, task_count=len(tasks), step_count=step_count,
    )
    (out_dir / "cost-worksheet-payload-budget.json").write_text(
        json.dumps(payload_ws, indent=2) + "\n", encoding="utf-8",
    )
    (out_dir / "dry-run-results.json").write_text(
        json.dumps(
            {
                "benchmark": "opmem-v0",
                "mode": "dry_run",
                "run_id": args.run_id,
                "task_count": len(tasks),
                "step_count": step_count,
                "lanes": per_lane_results,
            },
            indent=2,
        )
        + "\n",
        encoding="utf-8",
    )
    print(json.dumps(worksheet.to_dict(), indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
