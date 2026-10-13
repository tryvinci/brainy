#!/usr/bin/env python3
"""OpMem manuscript experiments: manifests, lane ablations, stability repeats.

Uses the same task runner as run_opmem.py. Intended to run against a live Brainy
API or the httptest server spawned by go test (see scripts/run-opmem-manuscript.sh).
"""
from __future__ import annotations

import argparse
import json
import pathlib
import subprocess
import sys
import time
from datetime import datetime, timezone

ROOT = pathlib.Path(__file__).resolve().parent.parent
sys.path.insert(0, str(ROOT / "evals"))

from run_opmem import build_adapters, run_task  # noqa: E402


def git_sha() -> str:
    try:
        return (
            subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT, text=True)
            .strip()
        )
    except (subprocess.CalledProcessError, FileNotFoundError):
        return "unknown"


def load_tasks(fixture_dir: pathlib.Path) -> list[dict]:
    return [
        json.loads(path.read_text(encoding="utf-8"))
        for path in sorted(fixture_dir.glob("*.json"))
    ]


def run_suite(
    base_url: str,
    system_names: list[str],
    tasks: list[dict],
) -> dict:
    adapters = build_adapters(system_names, base_url)
    results = []
    infrastructure_errors = 0
    for task in tasks:
        entry = {"task": task["name"], "category": task["category"], "systems": {}}
        for adapter in adapters:
            if not adapter.available():
                entry["systems"][adapter.name] = {"skipped": True, "reason": "not configured"}
                continue
            try:
                entry["systems"][adapter.name] = run_task(adapter, task)
            except Exception as exc:  # noqa: BLE001
                infrastructure_errors += 1
                entry["systems"][adapter.name] = {"error": f"{type(exc).__name__}: {exc}"}
        results.append(entry)

    summary: dict[str, dict] = {}
    for adapter in adapters:
        by_category: dict[str, dict[str, int]] = {}
        passed_total = total = 0
        for entry in results:
            outcome = entry["systems"].get(adapter.name, {})
            if outcome.get("skipped") or outcome.get("error"):
                continue
            bucket = by_category.setdefault(entry["category"], {"passed": 0, "total": 0})
            bucket["total"] += 1
            total += 1
            if outcome.get("passed"):
                bucket["passed"] += 1
                passed_total += 1
        summary[adapter.name] = {
            "overall": f"{passed_total}/{total}",
            "by_category": {cat: f"{v['passed']}/{v['total']}" for cat, v in sorted(by_category.items())},
        }

    return {
        "benchmark": "opmem-v0",
        "systems": [a.name for a in adapters],
        "summary": summary,
        "infrastructure_errors": infrastructure_errors,
        "results": results,
    }


def stability_runs(
    base_url: str,
    tasks: list[dict],
    repeats: int,
) -> dict:
    runs = []
    for index in range(repeats):
        report = run_suite(base_url, ["brainy"], tasks)
        passed = report["summary"].get("brainy", {}).get("overall", "0/0")
        runs.append({"repeat": index + 1, "overall": passed, "infrastructure_errors": report["infrastructure_errors"]})
        time.sleep(0.05)
    overalls = [r["overall"] for r in runs]
    return {
        "repeats": repeats,
        "system": "brainy",
        "recall_lane": "search",
        "revise_mode": "correct",
        "runs": runs,
        "stable": len(set(overalls)) == 1,
        "unique_scores": sorted(set(overalls)),
    }


def mechanism_view(report: dict) -> dict:
    """Category-level pass rates as a lightweight mechanism breakdown."""
    out: dict[str, dict[str, str]] = {}
    for system, summ in report.get("summary", {}).items():
        out[system] = summ.get("by_category", {})
    out["_note"] = (
        "verbatim is the rank-only baseline (no correct/suppress lifecycle); "
        "correction/suppression/staleness categories map to revise/forget semantics."
    )
    return out


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", default="http://127.0.0.1:8080")
    parser.add_argument("--fixture-dir", default="fixtures/opmem")
    parser.add_argument("--out-dir", required=True)
    parser.add_argument("--stability-repeats", type=int, default=5)
    parser.add_argument("--skip-mem0", action="store_true")
    args = parser.parse_args()

    fixture_dir = pathlib.Path(args.fixture_dir)
    if not fixture_dir.is_absolute():
        fixture_dir = ROOT / fixture_dir
    tasks = load_tasks(fixture_dir)
    out_dir = pathlib.Path(args.out_dir)
    if not out_dir.is_absolute():
        out_dir = ROOT / out_dir
    out_dir.mkdir(parents=True, exist_ok=True)

    sha = git_sha()
    manifest = {
        "benchmark": "opmem-v0",
        "git_sha": sha,
        "run_utc": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "fixture_dir": str(fixture_dir.relative_to(ROOT) if fixture_dir.is_relative_to(ROOT) else fixture_dir),
        "task_count": len(tasks),
        "base_url": args.base_url,
        "protocol": "docs/research/opmem/PUBLICATION_READINESS.md",
        "scoring": "substring assertions on ranked recall results",
    }
    (out_dir / "run-manifest.json").write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")

    stability = stability_runs(args.base_url, tasks, args.stability_repeats)
    valid = [r for r in stability["runs"] if r.get("infrastructure_errors", 0) == 0]
    stability["valid_runs"] = len(valid)
    stability["valid_scores"] = sorted({r["overall"] for r in valid})
    stability["interpretation"] = (
        "Runs with infrastructure_errors>0 are invalid/incomplete; "
        "partial denominators (e.g. 9/10) are not comparable to a clean 13/13 pin."
    )
    stability["manifest"] = manifest
    (out_dir / "opmem-stability.json").write_text(json.dumps(stability, indent=2) + "\n", encoding="utf-8")

    lanes = run_suite(
        args.base_url,
        ["brainy", "brainy-recall", "brainy-supersede"],
        tasks,
    )
    lanes["manifest"] = manifest
    (out_dir / "opmem-lane-ablation.json").write_text(json.dumps(lanes, indent=2) + "\n", encoding="utf-8")

    primary_systems = ["verbatim", "brainy"]
    if not args.skip_mem0:
        primary_systems.append("mem0")
    primary = run_suite(args.base_url, primary_systems, tasks)
    primary["manifest"] = manifest
    (out_dir / "opmem-primary.json").write_text(json.dumps(primary, indent=2) + "\n", encoding="utf-8")

    mechanism = {
        "manifest": manifest,
        "by_category": mechanism_view(primary),
        "verbatim_overall": primary["summary"].get("verbatim", {}).get("overall"),
    }
    (out_dir / "opmem-mechanism.json").write_text(json.dumps(mechanism, indent=2) + "\n", encoding="utf-8")

    log_lines = [
        f"git_sha={sha}",
        f"tasks={len(tasks)}",
        f"primary={json.dumps(primary['summary'])}",
        f"lanes={json.dumps(lanes['summary'])}",
        f"stability={json.dumps(stability)}",
    ]
    (out_dir / "run.log").write_text("\n".join(log_lines) + "\n", encoding="utf-8")

    return 1 if primary["infrastructure_errors"] else 0


if __name__ == "__main__":
    sys.exit(main())
