# LOCOMO 1×30 — R4h image WRITE + copula titles — `locomo-mh-r4h-dev-1x30-20260815`

**Honest pin.** Local **20/30 (66.7%)** is **+1 vs R4c 19/30 (63.3%)**. Multi-hop on this 1×30 is **10/10 (100%)**. On the frozen Mem0 same-pin it **leads overall 20 vs 12** and **leads multi-hop 10/10 vs 7/10**. Open-domain is still **0/4 vs Mem0 3/4**. This is **not** a SOTA / beats-Mem0 claim. 1×30 remains measurement, not qualification.

**Timestamp:** 2026-08-15T05:55:31Z
**Stack:** local API+worker at commit `f4ec4d7` (OCR cover-face windows + deictic titled-work bind + enumerate keeps copula titles)
**Dataset SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` (same as Wave 1 / R1c / R1b / R4c / Mem0 freeze)
**Answerer / judge:** same pinned LLM (temp=0.0); `BRAINY_USE_RECALL=1`; **API `BRAINY_RECALL_LLM=1`**
**Ingest:** async, 419 turns / 19 sessions, conv-26 (`image_urls` + `query` / `blip_caption` kept as turn text)
**Failure ledger:** [locomo-mh-r4h-dev-1x30.jsonl](./failure-ledger/locomo-mh-r4h-dev-1x30.jsonl)

All 30 scored items answered via product `/recall`. `metrics.errors: 1` is q8 `JUDGE_MISS` (unparseable judge); not counted correct.

Do **not** ship r4d (**MH 8/10**): single-crop OCR accepted letter shards (`IS VV`) and polluted activity lists. Do **not** ship r4e (**MH 8/10**): WRITE landed both titles but enumerate split copula titles and caption pose-places crowded swimming off the list.

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | **0.667 (20/30)** |
| Search p50 ms | 124.7 |
| Search p95 ms | 149.4 |

| Category | Acc | n |
| --- | ---: | --- |
| multi-hop | 1.000 (**10/10**) | 10 |
| open-domain | 0.000 (**0/4**) | 4 |
| temporal | 0.625 (**10/16**) | 16 |

## What the product change actually did

R4c text-join MH was 9/10. The remaining miss was image WRITE: the second titled work exists as cover lettering, not transcript/BLIP/`query`. A single upper-center crop of the 3D cover returned shards; `titleFromVisibleText` treated `IS VV` as a title and skipped better windows.

This cycle: overlapping cover-face OCR; persist only a well-formed title in `[visible text:]`; attach that block to the deictic sentence; reject shard titles; enumerate treats quoted / visible-text titles as atomic (do not split on `is`); pose gerunds and scene nouns from image captions do not compile as activities.

Held-out compiler/recall tests use Riley / Quiet Orchard / Life is Elsewhere. No benchmark gold strings in product `.go`.

Measured recoveries vs R4c (same 30 items):

| Item | Group | R4c | Now | Mechanism |
| --- | --- | --- | --- | --- |
| q23 titled works | MH | WRONG (Charlotte's Web only) | **CORRECT** (both titles) | OCR WRITE + copula-safe enumerate |
| q15 activities | MH | CORRECT | **CORRECT** (swimming kept) | caption pose/place atoms no longer crowd the list |
| q26 read date | temporal | WRONG | **CORRECT** (`12 July 2022`) | dated titled-work atom from the cover turn |
| q29 workshop date | temporal | CORRECT | **WRONG** (dip) | pottery date still in the packet; reader cited the activity atom |

## Compare (do not blend)

| Pin | Overall | MH | OD | Temporal | Notes |
| --- | ---: | ---: | ---: | ---: | --- |
| R1b | 15/30 (50.0%) | 6/10 (60.0%) | 0/4 | 9/16 (56.2%) | MH trail 6 vs 7 |
| R4c (`d48e202` / `4c96984`) | 19/30 (63.3%) | 9/10 (90.0%) | 0/4 | 10/16 (62.5%) | text-join MH; q23 image-gold |
| r4d (not shipped) | 19/30 (63.3%) | 8/10 (80.0%) | 0/4 | 11/16 (68.8%) | OCR shard titles; q15 dip |
| r4e (not shipped) | 19/30 (63.3%) | 8/10 (80.0%) | 1/4 | 10/16 (62.5%) | WRITE ok; enumerate split `is` |
| **This pin (`f4ec4d7`)** | **20/30 (66.7%)** | **10/10 (100%)** | **0/4 (0.0%)** | **10/16 (62.5%)** | image WRITE + reader |
| Mem0 Platform freeze 2026-08-13 | 12/30 (40.0%) | 7/10 (70.0%) | 3/4 (75.0%) | 2/16 (12.5%) | same dataset SHA |

## Non-reg (same stack)

- OpMem **13/13 (100%)** — [opmem-mh-r4h-local-20260815.md](./opmem-mh-r4h-local-20260815.md)
- Marketing vertical **17/17 (100%) passed** — [marketing-mh-r4h-local-20260815.md](./marketing-mh-r4h-local-20260815.md)

## Claims

Allowed: lead **this frozen Mem0 same-pin overall** 20/30 (66.7%) vs 12/30 (40.0%); MH **lead** 10/10 vs 7/10; OD **trail** 0/4 vs 3/4; temporal **lead** 10/16 vs 2/16. This 1×30 MH axis is closed. OD hypotheticals are not.

Forbidden: SOTA; beats-Mem0 as a general claim; treating 1×30 as qualification; mixing with 3×90 or Mem0 blog 90+; inventing Graphiti/Zep LoCoMo numbers; hardcoding titled-work gold in product code.
