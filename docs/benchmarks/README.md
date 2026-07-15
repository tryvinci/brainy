# Brainy Benchmarks

Public evaluation artifacts for Brainy. Two layers:

1. **Own suites** — OpMem, marketing vertical moat (reproducible in this repo)
2. **Public suites** — LOCOMO, LongMemEval, BEAM (external datasets; adapters WIP)

Honest rule: we do **not** invent scores. Suites without a run are marked **not run**.

---

## Current results (staging, 2026-07-14)

| Suite | Brainy | Mem0 | Verbatim | Report |
| --- | ---: | ---: | ---: | --- |
| Parity | 4/4 | 4/4 | — | [staging competitive](./staging-competitive-report.md) |
| OpMem v0 | **12/12** | 9/12 | 9/12 | [staging competitive](./staging-competitive-report.md) |
| Marketing vertical | 16/16 | N/A (moat) | — | [moat report](./marketing-moat-report.md) |
| LOCOMO | **not run** | — | — | [ladder](../research/public-bench-ladder.md) |
| LongMemEval | **not run** | — | — | [ladder](../research/public-bench-ladder.md) |
| BEAM | **not run** | — | — | [ladder](../research/public-bench-ladder.md) |

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
