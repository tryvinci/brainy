# LOCOMO 1×30 — compiler atom quality — `locomo-atomq-dev-1x30-20260814`

**Honest pin.** Local **11/30** is **+1 vs R1c 10/30** and still a **dip** vs Wave 1 local 14/30 and Gate 0 staging 18/30. It is **not** a SOTA / beats-Mem0 claim. 1×30 remains measurement, not qualification.

**Timestamp:** 2026-08-14T10:59:25Z
**Stack:** local API+worker at commit `d82f7d6` (malformed compiler-atom gate)
**Dataset SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` (same as Wave 1 / R1c)
**Answerer / judge:** same pinned LLM (temp=0.0); `BRAINY_USE_RECALL=1`; **API `BRAINY_RECALL_LLM=1`**
**Ingest:** async, 419 turns / 19 sessions, conv-26
**Failure ledger:** [locomo-atomq-dev-1x30.jsonl](./failure-ledger/locomo-atomq-dev-1x30.jsonl)

All 30 scored items answered via product `/recall`.

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | **0.367 (11/30)** |
| Search p50 ms | 165.3 |
| Search p95 ms | 200.4 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.200 (**2/10**) | 10 |
| open-domain | 0.000 (**0/4**) | 4 |
| temporal | 0.562 (**9/16**) | 16 |

## What the product change actually did

R1c’s dip was junk compiler facts crowding provenance (`has done going at …`, `participates in runn`), not empty retrieval.

Same-pin packet compare (conv-26, top-k=30):

| Metric | Wave 1 | R1c | This run |
| --- | ---: | ---: | ---: |
| Correct | 14 | 10 | **11** |
| Mean hits | 29.3 | 27.9 | 27.8 |
| Mean packet chars | 2012 | 1712 | **1907** |
| Dialogue-shaped hits | 639 | 507 | **550** |
| Junk template hits | 12 | **45** | **6** |

The diagnosed miss came back: q10 (`How long has Caroline had her current group of friends` → `4 years`) was CORRECT on Wave 1, WRONG on R1c with gold **absent** from the packet, CORRECT here with gold **present**. Live sync probe on the same compiler: duration utterance ranks first; origin atoms (`moved from canada`) persist; light-verb templates do not.

Remaining losses vs Wave 1 (q6, q13, q22, q27) still have gold missing as a **fact** or present as an episode the reader did not use. Ledger this run: **15 WRITE_MISS + 4 READER_MISS**. That WRITE_MISS mass is the real R1b coverage gap (durable claims still live only in transcripts).

## Compare (do not blend)

| Pin | Overall | MH | OD | Temporal | Notes |
| --- | ---: | ---: | ---: | ---: | --- |
| Gate 0 staging | 18/30 | 50% | 25% | — | pre-cutover |
| Wave 1 local (`a7a5184`) | **14/30** | **3/10** | **2/4** | **9/16** | hybrid reader on |
| R1c local (`21a632b`) | **10/30** | **2/10** | **0/4** | **8/16** | junk atoms crowded provenance |
| **This compiler-quality remasure (`d82f7d6`)** | **11/30** | **2/10** | **0/4** | **9/16** | junk 45→6; q10 recovered |

## Non-reg (same stack)

- OpMem **13/13** (stale-fact June vs May kept; numeric date tails are not malformed)
- Marketing vertical **17/17 passed**

## Claims

Forbidden: SOTA; beats-Mem0; MH solved; treating 11/30 as an improvement vs Gate 0 or Wave 1; treating 1×30 as a merge gate; LoCoMo-named product rules.
