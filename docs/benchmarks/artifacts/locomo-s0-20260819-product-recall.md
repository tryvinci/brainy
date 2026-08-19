# S0 product `/recall` stratified baseline — 2026-08-19

**Not a full LoCoMo pin. Not a SOTA / beats-Mem0 claim.** Stratified n=180, seed=1, 10 conversations.

| Field | Value |
| --- | --- |
| Run id | `locomo-s0-20260819-product-recall-s1-828403` |
| SHA | `df42f65` |
| Lane | `--eval-lane product-recall` (`POST /recall`, top-k 30) |
| Dataset SHA | `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` |
| Judge temp | 0.0 |
| Mix | SH 98 · MH 33 · temporal 38 · OD 11 |

## Scores

| Slice | Score |
| --- | ---: |
| Overall | **17/180 = 9.4%** |
| Single-hop | 5/98 = 5.1% |
| Multi-hop | 3/33 = 9.1% |
| Temporal | 9/38 = 23.7% |
| Open-domain | 0/11 = 0.0% |
| Search p50 / p95 | 172.5 / 222.4 ms (local) |

Do not paste this next to Mem0 92.5% or the July industry 49.8% band. Do not treat 9.4% as n=1540. The prior full `/recall` pin remains **175/1540 = 11.4%** at `1b5ab3e`.

## Failure ledger (163 misses)

| Primary | n |
| --- | ---: |
| WRITE_MISS | 120 |
| RETRIEVAL_MISS | 28 |
| PROOF_MISS | 12 |
| READER_MISS | 3 |

Largest group×stage: `single-hop:WRITE_MISS` 65. Histogram confirms WRITE_MISS is still the mass bucket (S1), not hop-planner absence.

Hop-plan coverage on this MH slice: **32/33** items emitted a typed hop plan (`hop_plan_count` 2 or 3; one item 0). Accuracy on those hops is still 3/33 — plans fire; compiled facts often do not.

No `JUDGE_MISS` on this lane.

## Industry lane

In flight on the same SHA / same 180 items (`locomo-s0-20260819-industry-search-s1-88882c`). No industry % until that run writes a report.

## Outlinks

- Report: [../runs/locomo-s0-20260819-product-recall-s1-828403.md](../runs/locomo-s0-20260819-product-recall-s1-828403.md)
- Ledger: [../runs/failure-ledger/locomo-s0-20260819-product-recall-s1-828403.jsonl](../runs/failure-ledger/locomo-s0-20260819-product-recall-s1-828403.jsonl)
