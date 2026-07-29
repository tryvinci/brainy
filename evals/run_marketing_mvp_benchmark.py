#!/usr/bin/env python3
"""Marketing MVP benchmark: Brainy vertical evals + optional empirical Mem0 counter-run.

By default runs Brainy only (static mem0_has matrix for expected differentiation).
Pass ``--systems brainy,mem0`` (requires MEM0_API_KEY) to *measure* Mem0 on the
same fixtures for a true comparison.
"""
from __future__ import annotations

import argparse
import json
import os
import pathlib
import sys
from datetime import datetime, timezone

from run_eval import run_suite


def load_matrix(path: pathlib.Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def capability_results(
    matrix: dict,
    brainy_fixtures: dict[str, dict],
    mem0_fixtures: dict[str, dict] | None = None,
) -> list[dict]:
    rows = []
    for cap in matrix.get("capabilities", []):
        fixture_names = cap.get("fixtures", [])
        brainy_runs = [brainy_fixtures[name] for name in fixture_names if name in brainy_fixtures]
        brainy_pass = bool(brainy_runs) and all(r.get("passed") for r in brainy_runs)

        mem0_declared = cap.get("mem0_has")
        mem0_empirical = None
        if mem0_fixtures is not None:
            mem0_runs = [mem0_fixtures[name] for name in fixture_names if name in mem0_fixtures]
            if mem0_runs:
                mem0_empirical = all(r.get("passed") for r in mem0_runs)

        # Differentiation: Brainy passes and Mem0 fails (empirical if present, else declared false).
        if mem0_empirical is not None:
            differentiation = brainy_pass and (mem0_empirical is False)
        else:
            differentiation = mem0_declared is False and brainy_pass

        rows.append(
            {
                "id": cap["id"],
                "name": cap["name"],
                "mem0_has_declared": mem0_declared,
                "mem0_empirical_pass": mem0_empirical,
                "gap": cap.get("gap"),
                "fixtures": fixture_names,
                "brainy_pass": brainy_pass,
                "differentiation": differentiation,
            }
        )
    return rows


def render_markdown(report: dict) -> str:
    mem0 = report.get("mem0") or {}
    mem0_mode = report.get("mem0_mode", "declared")
    lines = [
        "# Marketing MVP Benchmark Report",
        "",
        f"- **Benchmark:** `{report['benchmark_id']}`",
        f"- **Generated:** {report['generated_at']}",
        f"- **Mem0 mode:** `{mem0_mode}` "
        + ("(fixtures executed on Mem0 Platform)" if mem0_mode == "empirical" else "(static matrix only)"),
        f"- **Mem0 reference commit:** `{report['mem0_reference_commit']}`",
        "",
        "## Summary",
        "",
        f"| Suite | Brainy | Mem0 |",
        f"| --- | ---: | ---: |",
        f"| Parity | {report['parity']['passed']}/{report['parity']['total']} | "
        f"{_suite_mem0(report, 'parity')} |",
        f"| Vertical (marketing) | {report['vertical']['passed']}/{report['vertical']['total']} | "
        f"{_suite_mem0(report, 'vertical')} |",
        "",
        f"**MVP ready (Brainy):** {'yes' if report['mvp_ready'] else 'no'}",
        "",
        f"**Differentiation score:** {report['differentiation']['passed']}/{report['differentiation']['total']} "
        f"capabilities where Brainy passes and Mem0 does not.",
        "",
        "## Capabilities vs Mem0",
        "",
        "| Capability | Brainy | Mem0 declared | Mem0 empirical | Differentiation |",
        "| --- | --- | --- | --- | --- |",
    ]
    for row in report["capabilities"]:
        brainy = "pass" if row["brainy_pass"] else "fail"
        declared = _mem0_cell(row.get("mem0_has_declared", row.get("mem0_has")))
        emp = row.get("mem0_empirical_pass")
        emp_s = "n/a" if emp is None else ("pass" if emp else "fail")
        diff = "yes" if row["differentiation"] else "no"
        lines.append(f"| {row['name']} | {brainy} | {declared} | {emp_s} | {diff} |")

    lines.extend(["", "## Fixture detail", ""])
    for suite_name in ("parity", "vertical"):
        lines.append(f"### {suite_name} — Brainy")
        lines.append("")
        for item in report[suite_name]["results"]:
            status = "pass" if item.get("passed") else "fail"
            if item.get("skipped"):
                status = "skip"
            lines.append(f"- `{item['fixture']}` — **{status}**")
            for err in item.get("errors", []):
                lines.append(f"  - {err}")
        lines.append("")
        if mem0.get(suite_name):
            lines.append(f"### {suite_name} — Mem0 (empirical)")
            lines.append("")
            for item in mem0[suite_name]["results"]:
                status = "pass" if item.get("passed") else "fail"
                if item.get("skipped"):
                    status = "skip"
                lines.append(f"- `{item['fixture']}` — **{status}**")
                for err in item.get("errors", [])[:3]:
                    lines.append(f"  - {err}")
            lines.append("")

    lines.extend(
        [
            "## Interpretation",
            "",
            "- **Parity suite** — Mem0-like ingest/search/dedupe; both systems should mostly pass.",
            "- **Vertical suite** — marketing pack behavior (Principle > preference, lifecycle, etc.).",
            "- **Declared** `mem0_has` comes from the capability matrix (design expectation).",
            "- **Empirical** runs the *same fixture JSON* against Mem0 Platform when `--systems` includes `mem0`.",
            "- Fixtures that require `explain.primitive` / pack labels will fail on Mem0 by design — that is the moat.",
            "",
            "Reproduce:",
            "",
            "```bash",
            "# Brainy only (declared Mem0 gaps)",
            "python3 evals/run_marketing_mvp_benchmark.py --base-url \"$BRAINY_BASE_URL\"",
            "",
            "# True counter-run (requires MEM0_API_KEY)",
            "python3 evals/run_marketing_mvp_benchmark.py --base-url \"$BRAINY_BASE_URL\" \\",
            "  --systems brainy,mem0",
            "```",
            "",
        ]
    )
    return "\n".join(lines)


def _suite_mem0(report: dict, suite: str) -> str:
    mem0 = report.get("mem0") or {}
    block = mem0.get(suite)
    if not block:
        return report[suite].get("mem0_expected", "n/a")
    return f"{block['passed']}/{block['total']}"


def _mem0_cell(value) -> str:
    if value is True:
        return "yes"
    if value is False:
        return "no"
    if value == "approximate":
        return "approx"
    return str(value)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", default="http://127.0.0.1:8080")
    parser.add_argument("--matrix", default="evals/marketing_mvp_matrix.json")
    parser.add_argument("--json-out", default="docs/vertical/marketing-mvp-benchmark.json")
    parser.add_argument("--md-out", default="docs/vertical/marketing-mvp-benchmark.md")
    parser.add_argument(
        "--systems",
        default="brainy",
        help="Comma list: brainy and/or mem0 (mem0 needs MEM0_API_KEY for empirical counter-run)",
    )
    args = parser.parse_args()

    systems = {s.strip().lower() for s in args.systems.split(",") if s.strip()}
    root = pathlib.Path(__file__).resolve().parent.parent
    matrix_path = root / args.matrix
    matrix = load_matrix(matrix_path)

    parity_dir = root / matrix["parity_suite"]["dir"]
    vertical_dir = root / matrix["vertical_suite"]["dir"]

    if "brainy" not in systems:
        print("brainy must be included in --systems (MVP gate)", file=sys.stderr)
        return 2

    parity_ok, parity_results = run_suite(args.base_url, parity_dir)
    vertical_ok, vertical_results = run_suite(args.base_url, vertical_dir)

    brainy_fixtures = {
        r["fixture"]: r for r in parity_results + vertical_results if r.get("fixture")
    }

    mem0_block = None
    mem0_fixtures = None
    mem0_mode = "declared"
    if "mem0" in systems:
        from competitors.mem0_adapter import Mem0Adapter
        from competitors.run_mem0_fixtures import run_suite as run_mem0_suite

        adapter = Mem0Adapter()
        if not adapter.available():
            print("MEM0_API_KEY required for --systems mem0", file=sys.stderr)
            return 2
        mem0_mode = "empirical"
        # Parity: content-level compare (ignore Brainy-only kind/primitive schema).
        # Vertical: strict_schema so missing pack primitives count as Mem0 fails (moat).
        m_parity_ok, m_parity = run_mem0_suite(adapter, parity_dir, strict_schema=False)
        m_vert_ok, m_vert = run_mem0_suite(adapter, vertical_dir, strict_schema=True)
        mem0_fixtures = {r["fixture"]: r for r in m_parity + m_vert if r.get("fixture")}
        mem0_block = {
            "parity": {
                "passed": sum(1 for r in m_parity if r.get("passed") and not r.get("skipped")),
                "total": len(m_parity),
                "suite_passed": m_parity_ok,
                "results": m_parity,
            },
            "vertical": {
                "passed": sum(1 for r in m_vert if r.get("passed") and not r.get("skipped")),
                "total": len(m_vert),
                "suite_passed": m_vert_ok,
                "results": m_vert,
            },
        }

    caps = capability_results(matrix, brainy_fixtures, mem0_fixtures)

    if mem0_fixtures is not None:
        diff_pass = sum(1 for c in caps if c["differentiation"])
        diff_total = sum(1 for c in caps if c.get("mem0_empirical_pass") is False)
    else:
        diff_pass = sum(1 for c in caps if c["differentiation"])
        diff_total = sum(1 for c in caps if c.get("mem0_has_declared") is False)

    report = {
        "benchmark_id": matrix["benchmark_id"],
        "generated_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "mem0_reference_commit": matrix["mem0_reference_commit"],
        "mem0_mode": mem0_mode,
        "systems": sorted(systems),
        "mvp_ready": parity_ok and vertical_ok,
        "parity": {
            "passed": sum(1 for r in parity_results if r.get("passed") and not r.get("skipped")),
            "total": len(parity_results),
            "mem0_expected": matrix["parity_suite"]["mem0_expected"],
            "results": parity_results,
        },
        "vertical": {
            "passed": sum(1 for r in vertical_results if r.get("passed") and not r.get("skipped")),
            "total": len(vertical_results),
            "mem0_expected": matrix["vertical_suite"]["mem0_expected"],
            "results": vertical_results,
        },
        "differentiation": {
            "passed": diff_pass,
            "total": diff_total,
        },
        "capabilities": caps,
    }
    if mem0_block:
        report["mem0"] = mem0_block

    json_out = root / args.json_out
    md_out = root / args.md_out
    json_out.parent.mkdir(parents=True, exist_ok=True)
    json_out.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    md_out.write_text(render_markdown(report), encoding="utf-8")

    print(
        json.dumps(
            {
                "passed": report["mvp_ready"],
                "mem0_mode": mem0_mode,
                "json": str(json_out),
                "markdown": str(md_out),
            },
            indent=2,
        )
    )
    return 0 if report["mvp_ready"] else 1


if __name__ == "__main__":
    sys.exit(main())
