# Brainy Benchmarks

Public evaluation artifacts for Brainy. Two layers:

1. **Own suites** — OpMem, marketing vertical (reproducible in this repo)
2. **Public suites** — LOCOMO, LongMemEval, BEAM (external datasets; adapters WIP)

Honest rule: we do **not** invent scores. Suites without a run are marked **not run**. Do not claim SOTA.

---

## Current results (local R4h pin, 2026-08-15)

1×30 LoCoMo is **measurement, not qualification**. Open-domain still trails.

| Suite | Brainy | Report |
| --- | ---: | --- |
| LoCoMo 1×30 (conv-26) | **20/30** (MH 10/10, OD 0/4, temporal 10/16) | [R4h](./artifacts/locomo-mh-r4h-dev-1x30-20260815.md) |
| OpMem | **13/13** | [R4h OpMem](./artifacts/opmem-mh-r4h-local-20260815.md) |
| Marketing vertical | **17/17** | [R4h marketing](./artifacts/marketing-mh-r4h-local-20260815.md) · [moat](./marketing-moat-report.md) |
| LongMemEval-20 | **0/20** integrity (not re-run this cycle) | [cycle closeout](../research/competitive/cycle-closeout.md) |
| BEAM | **not run** | [ladder](../research/public-bench-ladder.md) |

Historical staging smoke (2026-07, do not mix with R4h): [locomo-smoke.md](./locomo-smoke.md).

When publishing, report accuracy + p95 latency + tokens, and outlink the dataset + judge.

---

## External benchmarks we will outlink (and eventually run)

| Benchmark | What it tests | Upstream |
| --- | --- | --- |
| **LOCOMO** | Very long multi-session dialogues (~300–600 turns); factual / multi-hop / temporal QA | [snap-research/locomo](https://github.com/snap-research/locomo) · [ACL 2024 paper](https://aclanthology.org/2024.acl-long.747/) · [project page](https://snap-research.github.io/locomo/) |
| **LongMemEval** | 500 Qs across info extraction, temporal, multi-session | Original LongMemEval paper/dataset |
| **BEAM** | Real-world retrieval across 100K–10M token chats | Public long-memory harnesses |
| **LongMemBench** | SuperMemory’s long-memory public framing | [supermemory.ai/research/longmembench](https://supermemory.ai/research/longmembench/) |

When we publish LOCOMO numbers, the post must link snap-research + methodology (judge model, temp=0, retrieval vs full-context baselines). Cite the **dataset**, not a vendor blog.

---

## Own suites (shipped)

| Suite | Spec / fixtures | Runner |
| --- | --- | --- |
| OpMem | [opmem-spec](../research/opmem-spec.md) · `fixtures/opmem/` | `evals/run_opmem.py` |
| Parity | `fixtures/parity/` | `evals/run_eval.py` |
| Marketing vertical | `fixtures/vertical/marketing/` | `evals/run_vertical_eval.py` |
| Marketing MVP / moat | [METHODOLOGY](./METHODOLOGY.md) | `evals/run_marketing_mvp_benchmark.py` |

---

## Reproduce own-suite runs

```bash
export BRAINY_BASE_URL=http://localhost:8080
python3 evals/run_opmem.py --systems brainy,verbatim --base-url "$BRAINY_BASE_URL"
python3 evals/run_vertical_eval.py --base-url "$BRAINY_BASE_URL"
python3 evals/run_marketing_mvp_benchmark.py --base-url "$BRAINY_BASE_URL"
```

---

## Research portal

Product-facing narratives live under [`docs/research/`](../research/README.md).
