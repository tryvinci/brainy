# LOCOMO 1×30 — fresh remasure — `locomo-fresh-1x30-20260815`

**Honest pin.** Local **21/30 (70.0%)** is **+1 vs R4h 20/30 (66.7%)**. Multi-hop stays **10/10**. Open-domain stays **0/4**. Temporal **11/16** (was 10/16 on R4h). Same-pin Mem0 Platform this cycle: **11/30 (36.7%)** ([locomo-mem0-fresh-1x30-20260815.md](./locomo-mem0-fresh-1x30-20260815.md)). Brainy **leads overall 21 vs 11** and **MH 10/10 vs 6/10**; **trails OD 0/4 vs 3/4**. This is **not** a SOTA / beats-Mem0 claim. 1×30 remains measurement, not qualification.

**Timestamp:** 2026-08-15T15:32:29Z
**Stack:** dedicated local API+worker rebuilt from product SHA `1b5ab3e` (docs HEAD at run time `7525f92`); `BRAINY_USE_RECALL=1`; API `BRAINY_RECALL_LLM=1`; async ingest
**Dataset SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` (same as Wave 1 / R4h / Mem0 freeze)
**Judge:** temp=0.0; categories 1–4 (adversarial excluded)
**Ingest:** async, 419 turns / 19 sessions, conv-26
**Failure ledger:** [locomo-fresh-1x30-20260815.jsonl](./failure-ledger/locomo-fresh-1x30-20260815.jsonl)

All 30 scored items answered via product `/recall` (`brainy-recall+answer` 17, `+enumerate` 10, `+abstain` 3). `metrics.errors: 1` is q8 `JUDGE_MISS` (unparseable judge); not counted correct.

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | **0.700 (21/30)** |
| Search p50 ms | 174.5 |
| Search p95 ms | 201.3 |

| Category | Acc | n |
| --- | ---: | --- |
| multi-hop | 1.000 (**10/10**) | 10 |
| open-domain | 0.000 (**0/4**) | 4 |
| temporal | 0.688 (**11/16**) | 16 |

## Failures this pin

| Item | Group | Judgment | Primary |
| --- | --- | --- | --- |
| q2 | OD | WRONG | WRITE_MISS |
| q5 | temporal | WRONG | WRITE_MISS |
| q6 | temporal | WRONG | READER_MISS (abstain) |
| q8 | temporal | JUDGE_MISS | JUDGE_MISS |
| q9 | temporal | WRONG | WRITE_MISS |
| q14 | OD | WRONG | WRITE_MISS (abstain) |
| q22 | OD | WRONG | WRITE_MISS |
| q27 | OD | WRONG | WRITE_MISS (abstain) |
| q29 | temporal | WRONG | RETRIEVAL_MISS (pottery date) |

R4h recoveries held: q15 activities CORRECT, q23 titled works CORRECT, q26 read date CORRECT, q10 friends duration CORRECT. q29 pottery date remains WRONG.

## Compare (do not blend)

| Pin | Overall | MH | OD | Temporal |
| --- | ---: | ---: | ---: | ---: |
| R4h (`f4ec4d7`) | 20/30 (66.7%) | 10/10 | 0/4 | 10/16 |
| **This pin (`1b5ab3e`)** | **21/30 (70.0%)** | **10/10** | **0/4** | **11/16** |
| Mem0 Platform 2026-08-13 freeze | 12/30 (40.0%) | 7/10 | 3/4 | 2/16 |
| **Mem0 Platform this cycle** | **11/30 (36.7%)** | **6/10** | **3/4** | **2/16** |

OD remains the trail axis (0/4 vs Mem0 3/4). Search p50: Brainy 175 ms local vs Mem0 492 ms platform — harness observation, not an SLO.

## Claims

Allowed: lead **this frozen Mem0 same-pin overall** 21/30 vs 11/30; MH **lead** 10/10 vs 6/10; OD **trail** 0/4 vs 3/4; temporal **lead** 11/16 vs 2/16. Forbidden: SOTA; beats-Mem0 as a general claim; treating 1×30 as qualification; mixing with full-suite 1540 or vendor 90+ headlines.
