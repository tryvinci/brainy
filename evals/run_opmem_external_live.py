#!/usr/bin/env python3
"""Phase 2: smoke or full OpMem external lane runs with preflight + artifacts."""

from __future__ import annotations

import argparse
import json
import os
import platform
import subprocess
import sys
import time
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parent
sys.path.insert(0, str(ROOT))

from opmem_external import LANE_ADAPTERS  # noqa: E402
from opmem_external.cost_budget import (  # noqa: E402
    build_lane_budgets,
    build_worksheet_from_budgets,
    load_tasks,
)
from opmem_external.pins import LANE_ENV_REQUIRED, SPEND_CEILING_USD  # noqa: E402
from run_opmem import run_task  # noqa: E402

DEFAULT_OUT = ROOT.parent / "docs/benchmarks/artifacts/opmem-external-lanes"
SMOKE_TASK = "dup01_idempotent_remember.json"


def _utc_now() -> str:
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def _package_versions() -> dict[str, str]:
    names = (
        "mem0ai",
        "qdrant-client",
        "zep-cloud",
        "tiktoken",
        "langmem",
        "langgraph-checkpoint-postgres",
        "openai",
        "langchain-openai",
    )
    out: dict[str, str] = {}
    for name in names:
        try:
            import importlib.metadata as md

            out[name] = md.version(name)
        except Exception:
            out[name] = "not_installed"
    return out


def _preflight_budget(tasks: list[dict]) -> dict:
    step_count = sum(len(t["steps"]) for t in tasks)
    budgets = build_lane_budgets(tasks)
    return build_worksheet_from_budgets(budgets, task_count=len(tasks), step_count=step_count)


def _lane_blocked(lane: str) -> tuple[bool, str]:
    if lane == "zep":
        if not os.environ.get("ZEP_API_KEY", "").strip():
            return True, "ZEP_API_KEY not present in runtime (vault not injected on this VM)"
    if lane == "letta":
        if not os.environ.get("LETTA_APP_SERVER_TOKEN", "").strip():
            return True, "LETTA_APP_SERVER_TOKEN not present in runtime"
        base = os.environ.get("LETTA_BASE_URL", "http://127.0.0.1:8283").rstrip("/")
        try:
            import urllib.request

            urllib.request.urlopen(f"{base}/v1/health", timeout=3)
        except Exception as exc:
            return True, f"Letta App Server unreachable at {base}: {exc}"
    if lane == "langmem":
        if not os.environ.get("DATABASE_URL", "").strip():
            default = "postgresql://opmem:opmem@127.0.0.1:5432/opmem_langmem"
            os.environ.setdefault("DATABASE_URL", default)
    for key in LANE_ENV_REQUIRED.get(lane, ()):
        if not os.environ.get(key, "").strip():
            return True, f"missing required env {key}"
    return False, ""


def _run_lane(
    lane: str,
    tasks: list[dict],
    *,
    run_id: str,
    out_dir: Path,
    mode: str,
) -> dict:
    blocked, reason = _lane_blocked(lane)
    if blocked:
        return {
            "lane": lane,
            "status": "blocked",
            "reason": reason,
            "infrastructure_errors": 0,
            "task_results": [],
        }

    factory = LANE_ADAPTERS[lane]
    adapter = factory(dry_run=False, run_id=run_id)
    if not adapter.available():
        return {
            "lane": lane,
            "status": "blocked",
            "reason": f"credentials: {adapter.dry_run_stats().missing_credentials}",
            "infrastructure_errors": 0,
            "task_results": [],
        }

    if lane == "zep":
        zep = adapter
        try:
            zep.verify_free_tier()
        except Exception as exc:
            return {
                "lane": lane,
                "status": "blocked",
                "reason": f"Zep free-tier preflight failed: {exc}",
                "infrastructure_errors": 1,
                "task_results": [],
            }

    infra = 0
    results = []
    started = time.time()
    for task in tasks:
        try:
            outcome = run_task(adapter, task)
            results.append(
                {
                    "task": task["name"],
                    "passed": outcome["passed"],
                    "failures": outcome.get("failures", []),
                }
            )
        except Exception as exc:
            infra += 1
            results.append(
                {
                    "task": task["name"],
                    "error": f"{type(exc).__name__}: {exc}",
                    "invalid_incomplete": True,
                }
            )
            if lane == "zep" and "hard-stop" in str(exc).lower():
                break

    elapsed = round(time.time() - started, 2)
    raw_path = out_dir / f"raw-{lane}-{mode}.json"
    adapter.write_raw_log(raw_path)
    telemetry = getattr(adapter, "telemetry", None)
    if telemetry is not None:
        telemetry.write(out_dir / f"telemetry-{lane}-{mode}.json")

    passed = sum(1 for r in results if r.get("passed"))
    return {
        "lane": lane,
        "status": "completed" if infra == 0 else "invalid_incomplete",
        "mode": mode,
        "run_id": run_id,
        "elapsed_seconds": elapsed,
        "infrastructure_errors": infra,
        "tasks_run": len(results),
        "tasks_passed": passed,
        "task_results": results,
        "product_constraints": adapter.dry_run_stats().product_constraints,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--fixture-dir", default="fixtures/opmem")
    parser.add_argument("--out-dir", default=str(DEFAULT_OUT))
    parser.add_argument("--run-id", default="")
    parser.add_argument(
        "--lanes",
        default="mem0-oss,zep,letta,langmem",
        help="Comma-separated lane ids",
    )
    parser.add_argument("--smoke", action="store_true", help="Run one smallest task per lane")
    parser.add_argument("--full", action="store_true", help="Run all 13 tasks per lane")
    parser.add_argument("--skip-preflight", action="store_true")
    args = parser.parse_args()

    if not args.smoke and not args.full:
        parser.error("Specify --smoke and/or --full")

    fixture_dir = Path(args.fixture_dir)
    if not fixture_dir.is_absolute():
        fixture_dir = ROOT.parent / fixture_dir
    all_tasks = load_tasks(fixture_dir)
    run_id = args.run_id or f"opmem-live-{datetime.now(timezone.utc).strftime('%Y%m%d%H%M%S')}"
    os.environ["OPMEM_RUN_ID"] = run_id

    out_dir = Path(args.out_dir)
    if not out_dir.is_absolute():
        out_dir = ROOT.parent / out_dir
    out_dir.mkdir(parents=True, exist_ok=True)

    if not args.skip_preflight:
        budget = _preflight_budget(all_tasks)
        (out_dir / "cost-worksheet-payload-budget.json").write_text(
            json.dumps(budget, indent=2) + "\n", encoding="utf-8",
        )
        total = next(r for r in budget["lanes"] if r["lane"] == "_total_openai_worst_case")
        if not total["within_ceiling"]:
            print("ABORT: projected spend exceeds owner ceiling", file=sys.stderr)
            print(json.dumps(budget, indent=2))
            return 2
        for row in budget["lanes"]:
            lane = row.get("lane", "")
            if lane in SPEND_CEILING_USD and not row.get("within_ceiling", True):
                print(f"ABORT: lane {lane} exceeds ceiling", file=sys.stderr)
                return 2

    lanes = [x.strip() for x in args.lanes.split(",") if x.strip()]
    manifest = {
        "benchmark": "opmem-v0",
        "generated_utc": _utc_now(),
        "run_id": run_id,
        "host": platform.platform(),
        "python": platform.python_version(),
        "package_versions": _package_versions(),
        "pinned_models": {
            "chat": "gpt-4.1-mini-2025-04-14",
            "embed": "text-embedding-3-small",
        },
        "ceilings_usd": SPEND_CEILING_USD,
        "modes": [],
    }

    smoke_tasks = [t for t in all_tasks if t["name"] == SMOKE_TASK.replace(".json", "")]
    if not smoke_tasks:
        smoke_path = fixture_dir / SMOKE_TASK
        smoke_tasks = [json.loads(smoke_path.read_text(encoding="utf-8"))]

    if args.smoke:
        smoke_out = []
        for lane in lanes:
            smoke_out.append(_run_lane(lane, smoke_tasks, run_id=run_id, out_dir=out_dir, mode="smoke"))
        manifest["modes"].append({"mode": "smoke", "task": smoke_tasks[0]["name"], "lanes": smoke_out})
        (out_dir / f"live-smoke-{run_id}.json").write_text(
            json.dumps(smoke_out, indent=2) + "\n", encoding="utf-8",
        )

    if args.full:
        full_out = []
        for lane in lanes:
            prior = next(
                (m for m in manifest.get("modes", []) if m.get("mode") == "smoke"),
                None,
            )
            if prior:
                lane_smoke = next((x for x in prior["lanes"] if x["lane"] == lane), None)
                if lane_smoke and lane_smoke.get("status") != "completed":
                    full_out.append(
                        {
                            "lane": lane,
                            "status": "skipped",
                            "reason": "smoke did not complete cleanly",
                        }
                    )
                    continue
            full_out.append(_run_lane(lane, all_tasks, run_id=run_id, out_dir=out_dir, mode="full"))
        manifest["modes"].append({"mode": "full", "task_count": len(all_tasks), "lanes": full_out})
        (out_dir / f"live-full-{run_id}.json").write_text(
            json.dumps(full_out, indent=2) + "\n", encoding="utf-8",
        )

    (out_dir / f"run-manifest-{run_id}.json").write_text(
        json.dumps(manifest, indent=2) + "\n", encoding="utf-8",
    )
    commands_log = out_dir / "commands.log"
    with commands_log.open("a", encoding="utf-8") as fh:
        fh.write(f"\n# {_utc_now()} run_id={run_id}\n")
        fh.write(" ".join(subprocess.list2cmdline(sys.argv)) + "\n")
    print(json.dumps(manifest, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
