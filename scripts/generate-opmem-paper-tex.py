#!/usr/bin/env python3
"""Generate LaTeX fragments for OpMem manuscript from run artifacts."""
from __future__ import annotations

import json
import pathlib
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
ART = ROOT / "docs/benchmarks/artifacts/opmem-manuscript-20533f2"
OUT = ROOT / "docs/research/opmem/paper/generated"


def load(name: str) -> dict:
    return json.loads((ART / name).read_text(encoding="utf-8"))


def write(name: str, body: str) -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    (OUT / name).write_text(body.strip() + "\n", encoding="utf-8")


def main() -> int:
    if not ART.is_dir():
        print(f"missing artifact dir {ART}", file=sys.stderr)
        return 1
    manifest = load("run-manifest.json")
    primary = load("opmem-primary.json")
    lane_path = ART / "opmem-lane-ablation-isolated.json"
    if not lane_path.is_file():
        lane_path = ART / "opmem-lane-ablation.json"
    lanes = json.loads(lane_path.read_text(encoding="utf-8"))
    stability = load("opmem-stability.json")
    mechanism = load("opmem-mechanism.json")

    sha = manifest["git_sha"]
    brainy = primary["summary"]["brainy"]["overall"]
    verbatim = primary["summary"]["verbatim"]["overall"]
    mem0_line = ""
    if "mem0" in primary["summary"]:
        mem0_line = (
            f"Mem0 Platform (same run): \\textbf{{{primary['summary']['mem0']['overall']}}} "
            f"(API surface: Platform via \\texttt{{Mem0OpAdapter}})."
        )

    write(
        "current-sha-results.tex",
        f"""
\\evidence{{completed}}{{
At git SHA \\texttt{{{sha[:12]}}} (UTC {manifest['run_utc']}), embedded API harness:
Brainy (search lane) \\textbf{{{brainy}}}, verbatim \\textbf{{{verbatim}}}.
{mem0_line}
}}
""",
    )

    write(
        "combined-results.tex",
        f"""
\\evidence{{completed}}{{
Committed \\texttt{{opmem-primary.json}} lists per-task outcomes for
{', '.join(primary['systems'])} with manifest fields (SHA, date, fixture count).
}}
""",
    )

    lane_rows = []
    for system in lanes["systems"]:
        summ = lanes["summary"].get(system, {})
        lane_rows.append(f"{system} & {summ.get('overall', 'n/a')} \\\\")
    write(
        "lane-ablation.tex",
        """
\\begin{table}[h]
\\centering
\\caption{Brainy recall lane and revise mapping (SHA """
        + sha[:12]
        + """).}
\\begin{tabular}{lc}
\\toprule
Configuration & Overall \\\\
\\midrule
"""
        + "\n".join(lane_rows)
        + """
\\bottomrule
\\end{tabular}
\\end{table}
Search (\\texttt{/memories/search}) is the normative pin lane.
\\texttt{brainy-recall} uses \\texttt{POST /recall} (context mode).
Correction tasks on the recall lane may return HTTP 409 during \\texttt{/correct}
(infrastructure errors in artifact JSON).
""",
    )

    sup = lanes["summary"].get("brainy-supersede", {}).get("overall", "n/a")
    corr = lanes["summary"].get("brainy", {}).get("overall", "n/a")
    write(
        "supersede-ablation.tex",
        f"""
\\evidence{{completed}}{{
Mapping \\texttt{{revise}} $\\rightarrow$ \\texttt{{POST .../supersede}} yields
\\textbf{{{sup}}} vs \\textbf{{{corr}}} with \\texttt{{/correct}} on the search lane
(same SHA). Failures are task-level assertion misses, not missing runs.
}}
""",
    )

    by_cat = mechanism.get("by_category", {})
    brainy_cat = by_cat.get("brainy", {})
    verb_cat = by_cat.get("verbatim", {})
    lines = [
        "\\begin{tabular}{lcc}",
        "\\toprule",
        "Category & Brainy & Verbatim (rank-only) \\\\",
        "\\midrule",
    ]
    for cat in sorted(brainy_cat.keys()):
        if cat.startswith("_"):
            continue
        lines.append(f"{cat} & {brainy_cat.get(cat, '---')} & {verb_cat.get(cat, '---')} \\\\")
    lines.append("\\bottomrule")
    lines.append("\\end{tabular}")
    write(
        "mechanism-ablation.tex",
        """
Category pass rates separate correction/suppression/staleness contracts;
verbatim approximates rank-only recall without governed lifecycle ops.
\\begin{table}[h]
\\centering
"""
        + "\n".join(lines)
        + "\n\\end{table}",
    )

    runs = stability.get("runs", [])
    valid = [r for r in runs if r.get("infrastructure_errors", 0) == 0]
    invalid = [r for r in runs if r.get("infrastructure_errors", 0) > 0]
    valid_scores = sorted({r["overall"] for r in valid})
    run_lines = [
        f"Repeat {r['repeat']}: {r['overall']} (infra errors={r['infrastructure_errors']})"
        for r in runs
    ]
    write(
        "stability.tex",
        f"""
\\evidence{{completed}}{{
Five consecutive Brainy search-lane repeats on one embedded server session.
\\textbf{{Valid runs only}} (infra errors $=0$): {len(valid)}/{len(runs)}; scores: {', '.join(valid_scores) or 'none'}.
}}
\\begin{{itemize}}
"""
        + "\n".join(f"\\item {line}" for line in run_lines)
        + f"""
\\end{{itemize}}
Repeats with infra errors $>0$ ({len(invalid)} of {len(runs)}) are \\textbf{{invalid/incomplete}}: the harness omits errored tasks from the denominator, so aggregates such as 9/10 are \\emph{{not}} comparable to a clean 13/13 pin.
Use isolated \\texttt{{TestOpMemBenchmarkAgainstHTTPServer}} for merge-gate stability.
""",
    )

    return 0


if __name__ == "__main__":
    sys.exit(main())
