# LOCOMO 1×30 — Wave 1 local remasure — `locomo-wave1-dev-1x30-20260813`

**Honest pin.** Local 14/30 is **not** an improvement vs Gate 0 staging 18/30. It is **not** a SOTA / beats-Mem0 claim. Multi-hop did not improve.

**Timestamp:** 2026-08-13T13:04:21Z  
**Stack:** local API+worker at commit `a7a5184` (Wave 1 merged to `dev`: #101–#105 / PR5, PR3, PR4, PR9)  
**Dataset SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Answerer / judge:** same pinned LLM (temp=0.0); `BRAINY_USE_RECALL=1`; **API `BRAINY_RECALL_LLM=1`**  
**Ingest:** async, 419 turns / 19 sessions, conv-26  
**Failure ledger:** [locomo-wave1-dev-1x30.jsonl](./failure-ledger/locomo-wave1-dev-1x30.jsonl)

All 30 scored items answered via product `/recall` (`brainy-recall+answer` / `+enumerate` / `+abstain`).

## Confound vs local PR2 6/30

The same-day PR2 remasure (`24be5ab`, 6/30) used `/recall` from the harness but the **API process did not have `BRAINY_RECALL_LLM=1`**. This run enabled the hybrid reader so Wave 1 PR5 (context vs proof) is actually exercised. Do **not** attribute the 6→14 jump to Wave 1 code alone.

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | **0.467 (14/30)** |
| Search p50 ms | 171.1 |
| Search p95 ms | 230.7 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.300 (3/10) | 10 |
| open-domain | 0.500 (2/4) | 4 |
| temporal | 0.562 (9/16) | 16 |

## Failure ledger

**16/16 WRONG labeled `READER_MISS`.** Stage oracles (evidence / semantic / retrieval / coverage) all **supported** (typically 30 hits). MH misses are still chat-continuation / enumerate dumps, not empty retrieval.

## Compare (do not blend)

| Pin | Overall | MH | Temporal | Notes |
| --- | ---: | ---: | ---: | --- |
| Gate 0 staging | 18/30 | 50% | — | pre-cutover |
| Post-cutover staging | 15/30 | 50% | 56.2% | `1f2f26f`; hybrid reader on staging |
| Harden local | 14/30 | 5/10 | — | hop-join dip |
| Local PR2 remasure | **6/30** | **4/10** | **1/16** | `24be5ab`; hybrid reader **off** on API |
| **This Wave 1 local remasure** | **14/30** | **3/10** | **9/16** | `a7a5184`; hybrid reader **on** |
| Fresh Mem0 same-pin (this day) | 12/30 | 70% | — | [pr10 pin](./locomo-mem0-samepin-pr10-20260813.md) |

Temporal moved (1/16 → 9/16). Multi-hop did **not** (4/10 → 3/10). Mem0 still leads MH on the same pin. Do not call 14/30 a win vs Gate 0 or vs Mem0.

## Claims

Forbidden: SOTA; beats-Mem0; MH solved; treating 14/30 as an improvement vs Gate 0; starting PR6–PR8 from this pin (coverage oracles still supported).
