#!/usr/bin/env python3
"""OpMem: operational memory correctness benchmark runner.

Runs neutral operation-script tasks (fixtures/opmem/) against one or more
memory systems and reports pass/fail per task and category.

Task assertion failures are diagnostic results; the process exits non-zero
only on infrastructure errors (unreachable API, adapter exception), so the
runner can double as a harness-health check in CI.

Spec: docs/research/opmem-spec.md
"""
from __future__ import annotations

import argparse
import json
import pathlib
import sys

ROOT = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(ROOT))

from opmem_adapters import BrainyAdapter, Mem0OpAdapter, VerbatimBaseline


def resolve_actor(task: dict, name: str) -> tuple[str, str]:
    spec = task.get("actors", {}).get(name)
    if spec:
        return spec["tenant"], spec["subject"]
    return "t1", name


def check_expect(results: list[dict], expect: dict, label: str) -> list[str]:
    failures = []
    contents = [r["content"].lower() for r in results]

    if len(results) < expect.get("min_results", 0):
        failures.append(f"{label}: expected at least {expect['min_results']} results, got {len(results)}")
    if "max_results" in expect and len(results) > expect["max_results"]:
        failures.append(f"{label}: expected at most {expect['max_results']} results, got {len(results)}")
    if "top_contains" in expect:
        needle = expect["top_contains"].lower()
        if not contents or needle not in contents[0]:
            failures.append(f"{label}: top result does not contain {expect['top_contains']!r}")
    if "top_not_contains" in expect and contents:
        needle = expect["top_not_contains"].lower()
        if needle in contents[0]:
            failures.append(f"{label}: top result contains forbidden {expect['top_not_contains']!r}")
    for needle in expect.get("must_include", []):
        if not any(needle.lower() in content for content in contents):
            failures.append(f"{label}: no result contains {needle!r}")
    for needle in expect.get("must_exclude", []):
        if any(needle.lower() in content for content in contents):
            failures.append(f"{label}: a result contains excluded {needle!r}")
    if expect.get("unique_contents") and len(set(contents)) != len(contents):
        failures.append(f"{label}: duplicate result contents")
    return failures


def run_task(adapter, task: dict) -> dict:
    adapter.begin_task(task["name"])
    step_memories: dict[str, list[str]] = {}
    failures: list[str] = []

    for index, step in enumerate(task["steps"]):
        op = step["op"]
        actor = resolve_actor(task, step.get("actor", "a1"))
        label = f"step{index}({op})"

        if op == "remember":
            ids = adapter.remember(actor, step["content"])
            if step.get("id"):
                step_memories[step["id"]] = ids
            if not ids:
                failures.append(f"{label}: remember produced no memories")
        elif op == "recall":
            results = adapter.recall(actor, step["query"])
            failures.extend(check_expect(results, step.get("expect", {}), label))
        elif op == "forget":
            adapter.forget(actor, step_memories.get(step["target"], []))
        elif op == "revise":
            adapter.revise(actor, step_memories.get(step["target"], []), step["content"])
        else:
            raise ValueError(f"unknown op {op!r} in task {task['name']}")

    return {"passed": not failures, "failures": failures}


def build_adapters(names: list[str], base_url: str) -> list:
    registry = {
        "verbatim": lambda: VerbatimBaseline(),
        "brainy": lambda: BrainyAdapter(base_url),
        "mem0": lambda: Mem0OpAdapter(),
    }
    adapters = []
    for name in names:
        if name not in registry:
            raise SystemExit(f"unknown system {name!r}; choose from {sorted(registry)}")
        adapters.append(registry[name]())
    return adapters


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", default="http://127.0.0.1:8080")
    parser.add_argument("--systems", default="verbatim,brainy")
    parser.add_argument("--fixture-dir", default="fixtures/opmem")
    parser.add_argument("--json-out", default="docs/benchmarks/opmem-latest.json")
    args = parser.parse_args()

    fixture_dir = pathlib.Path(args.fixture_dir)
    if not fixture_dir.is_absolute():
        fixture_dir = ROOT.parent / fixture_dir
    tasks = [json.loads(path.read_text(encoding="utf-8")) for path in sorted(fixture_dir.glob("*.json"))]
    if not tasks:
        print(json.dumps({"error": "no tasks found"}, indent=2))
        return 1

    adapters = build_adapters([n.strip() for n in args.systems.split(",") if n.strip()], args.base_url)

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
            except Exception as exc:  # noqa: BLE001 - report and continue
                infrastructure_errors += 1
                entry["systems"][adapter.name] = {"error": f"{type(exc).__name__}: {exc}"}
        results.append(entry)

    summary = {}
    for adapter in adapters:
        by_category: dict[str, dict[str, int]] = {}
        passed_total = total = 0
        for entry in results:
            outcome = entry["systems"].get(adapter.name, {})
            if outcome.get("skipped"):
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

    report = {
        "benchmark": "opmem-v0",
        "systems": [adapter.name for adapter in adapters],
        "summary": summary,
        "infrastructure_errors": infrastructure_errors,
        "results": results,
    }
    output = json.dumps(report, indent=2)
    print(output)
    if args.json_out:
        out = pathlib.Path(args.json_out)
        if not out.is_absolute():
            out = ROOT.parent / out
        out.parent.mkdir(parents=True, exist_ok=True)
        out.write_text(output + "\n", encoding="utf-8")
    return 1 if infrastructure_errors else 0


if __name__ == "__main__":
    sys.exit(main())
