#!/usr/bin/env python3
"""Run marketing MVP benchmark: Brainy vertical evals + Mem0 gap report (ENG-93)."""
from __future__ import annotations

import argparse
import json
import pathlib
import sys
from datetime import datetime, timezone

from run_eval import run_suite


def load_matrix(path: pathlib.Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def capability_results(matrix: dict, fixture_results: dict[str, dict]) -> list[dict]:
    rows = []
    for cap in matrix.get("capabilities", []):
        fixture_names = cap.get("fixtures", [])
        runs = [fixture_results[name] for name in fixture_names if name in fixture_results]
        brainy_pass = bool(runs) and all(r.get("passed") for r in runs)
        rows.append(
            {
                "id": cap["id"],
                "name": cap["name"],
                "mem0_has": cap.get("mem0_has"),
                "gap": cap.get("gap"),
                "fixtures": fixture_names,
                "brainy_pass": brainy_pass,
                "differentiation": cap.get("mem0_has") is False and brainy_pass,
            }
        )
    return rows


def render_markdown(report: dict) -> str:
    lines = [
        "# Marketing MVP Benchmark Report",
        "",
        f"- **Benchmark:** `{report['benchmark_id']}`",
        f"- **Generated:** {report['generated_at']}",
        f"- **Mem0 reference commit:** `{report['mem0_reference_commit']}`",
        "",
        "## Summary",
        "",
        f"| Suite | Brainy pass | Total | Mem0 expected |",
        f"| --- | ---: | ---: | --- |",
        f"| Parity | {report['parity']['passed']} | {report['parity']['total']} | pass |",
        f"| Vertical (marketing) | {report['vertical']['passed']} | {report['vertical']['total']} | fail |",
        "",
        f"**MVP ready:** {'yes' if report['mvp_ready'] else 'no'}",
        "",
        f"**Differentiation score:** {report['differentiation']['passed']}/{report['differentiation']['total']} "
        f"capabilities where Brainy passes and Mem0 lacks equivalent behavior.",
        "",
        "## Capabilities vs Mem0",
        "",
        "| Capability | Brainy | Mem0 | Differentiation |",
        "| --- | --- | --- | --- |",
    ]
    for row in report["capabilities"]:
        brainy = "pass" if row["brainy_pass"] else "fail"
        mem0 = _mem0_cell(row["mem0_has"])
        diff = "yes" if row["differentiation"] else "no"
        lines.append(f"| {row['name']} | {brainy} | {mem0} | {diff} |")

    lines.extend(["", "## Fixture detail", ""])
    for suite_name in ("parity", "vertical"):
        lines.append(f"### {suite_name}")
        lines.append("")
        for item in report[suite_name]["results"]:
            status = "pass" if item.get("passed") else "fail"
            if item.get("skipped"):
                status = "skip"
            lines.append(f"- `{item['fixture']}` — **{status}**")
            for err in item.get("errors", []):
                lines.append(f"  - {err}")
        lines.append("")

    lines.extend(
        [
            "## Interpretation",
            "",
            "- **Parity suite** exercises Mem0-like ingest/search/dedupe behavior Brainy must not regress.",
            "- **Vertical suite** exercises marketing pack capabilities Mem0 does not model.",
            "- **Differentiation** = Brainy passes and `mem0_has: false` in the capability matrix.",
            "",
            "Reproduce:",
            "",
            "```bash",
            "go run ./cmd/api  # with Postgres",
            "python3 evals/run_marketing_mvp_benchmark.py --base-url http://127.0.0.1:8080",
            "```",
            "",
        ]
    )
    return "\n".join(lines)


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
    args = parser.parse_args()

    root = pathlib.Path(__file__).resolve().parent.parent
    matrix_path = root / args.matrix
    matrix = load_matrix(matrix_path)

    parity_dir = root / matrix["parity_suite"]["dir"]
    vertical_dir = root / matrix["vertical_suite"]["dir"]

    parity_ok, parity_results = run_suite(args.base_url, parity_dir)
    vertical_ok, vertical_results = run_suite(args.base_url, vertical_dir)

    fixture_results = {r["fixture"]: r for r in parity_results + vertical_results}
    caps = capability_results(matrix, fixture_results)

    diff_pass = sum(1 for c in caps if c["differentiation"])
    diff_total = sum(1 for c in caps if c["mem0_has"] is False)

    report = {
        "benchmark_id": matrix["benchmark_id"],
        "generated_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "mem0_reference_commit": matrix["mem0_reference_commit"],
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

    json_out = root / args.json_out
    md_out = root / args.md_out
    json_out.parent.mkdir(parents=True, exist_ok=True)
    json_out.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    md_out.write_text(render_markdown(report), encoding="utf-8")

    print(json.dumps({"passed": report["mvp_ready"], "json": str(json_out), "markdown": str(md_out)}, indent=2))
    return 0 if report["mvp_ready"] else 1


if __name__ == "__main__":
    sys.exit(main())
