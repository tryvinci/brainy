# Brainy Benchmarks

Public evaluation artifacts for Brainy. Two layers:

1. **Own suites** — OpMem, marketing vertical moat (reproducible in this repo)
2. **Public suites** — LOCOMO, LongMemEval, BEAM (external datasets; adapters WIP)

Honest rule: we do **not** invent scores. Suites without a run are marked **not run**.

**Industry bake-off (honest):** [competitive-positioning-20260806.md](./competitive-positioning-20260806.md) — same-pin leads vs Mem0 on OpMem + marketing; conversational SOTA not yet claimable.

---

## Current results (local R4h pin, 2026-08-15)

1×30 LoCoMo is **measurement, not qualification**. Same dataset SHA as the
frozen Mem0 Platform pin. Open-domain still trails. Not SOTA.

| Suite | Brainy | Mem0 same-pin | Report |
| --- | ---: | ---: | --- |
| LoCoMo 1×30 (conv-26) | **20/30** (MH 10/10, OD 0/4, temporal 10/16) | 12/30 (MH 7/10, OD 3/4, temporal 2/16) | [R4h](./artifacts/locomo-mh-r4h-dev-1x30-20260815.md) · [Mem0 freeze](./artifacts/locomo-mem0-samepin-pr10-20260813.md) |
| OpMem | **13/13** | — | [R4h OpMem](./artifacts/opmem-mh-r4h-local-20260815.md) |
| Marketing vertical | **17/17** | — | [R4h marketing](./artifacts/marketing-mh-r4h-local-20260815.md) · [moat](./marketing-moat-report.md) |
| LongMemEval-20 | **0/20** integrity (not re-run this cycle) | — | [cycle closeout](../research/competitive/cycle-closeout.md) |
| BEAM | **not run** | — | [ladder](../research/public-bench-ladder.md) |

Historical staging smoke (2026-07, do not mix with R4h): [locomo-smoke.md](./locomo-smoke.md) · [staging competitive](./staging-competitive-report.md).

Full write-up style target (accuracy + p95 latency + tokens): [SuperMemory research](https://supermemory.ai/research/) · Mem0 LOCOMO blog format.

---

## External benchmarks we will outlink (and eventually run)

| Benchmark | What it tests | Upstream | Common runner |
| --- | --- | --- | --- |
| **LOCOMO** | Very long multi-session dialogues (~300–600 turns); factual / multi-hop / temporal QA | [snap-research/locomo](https://github.com/snap-research/locomo) · [ACL 2024 paper](https://aclanthology.org/2024.acl-long.747/) · [project page](https://snap-research.github.io/locomo/) | [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks) |
| **LongMemEval** | 500 Qs across info extraction, temporal, multi-session | Original LongMemEval paper/dataset (via Mem0 harness) | [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks) |
| **BEAM** | Real-world retrieval across 100K–10M token chats | Via Mem0 harness | [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks) |
| **LongMemBench** | SuperMemory’s long-memory public framing | [supermemory.ai/research/longmembench](https://supermemory.ai/research/longmembench/) | External |

**Cite the dataset, not only Mem0’s scores.** When we publish LOCOMO numbers, the post must link snap-research + methodology (judge model, temp=0, retrieval vs full-context baselines).

---

## Own suites (shipped)

| Suite | Spec / fixtures | Runner |
| --- | --- | --- |
| OpMem | [opmem-spec](../research/opmem-spec.md) · `fixtures/opmem/` | `evals/run_opmem.py` |
| Parity | `fixtures/parity/` | `evals/run_eval.py`, `evals/run_competitor_benchmark.py` |
| Marketing vertical | `fixtures/vertical/marketing/` | `evals/run_vertical_eval.py` |
| Marketing MVP / moat | [METHODOLOGY](./METHODOLOGY.md) | `evals/run_marketing_mvp_benchmark.py` |

---

## Reproduce staging competitive run

```bash
export BRAINY_BASE_URL=https://brainy-api-staging.onrender.com
export MEM0_API_KEY=...   # rotate; never commit
python3 evals/run_competitor_benchmark.py --brainy-url "$BRAINY_BASE_URL"
python3 evals/run_opmem.py --systems brainy,mem0,verbatim --base-url "$BRAINY_BASE_URL"
```

---

## Research portal

Product-facing narratives live under [`docs/research/`](../research/README.md) — the SuperMemory-style index for Brainy.
