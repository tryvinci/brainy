# LOCOMO 1×30 — R4 hops + leftover coverage — `locomo-mh-r4c-dev-1x30-20260815`

**Honest pin.** Local **19/30 (63.3%)** is **+4 vs R1b 15/30 (50.0%)**. On this frozen Mem0 same-pin it **leads overall 19 vs 12** and **leads multi-hop 9/10 vs 7/10**. The remaining MH miss is image-gold (`"Nothing is Impossible"` is not in transcript, BLIP, or image query). It is **not** a SOTA / beats-Mem0 claim. 1×30 remains measurement, not qualification.

**Timestamp:** 2026-08-15T04:33:44Z
**Stack:** local API+worker at commit `d48e202` (R4 hops + researched-topic / pronoun-excited atoms + async `memory_relations`)
**Dataset SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` (same as Wave 1 / R1c / R1b / Mem0 freeze)
**Answerer / judge:** same pinned LLM (temp=0.0); `BRAINY_USE_RECALL=1`; **API `BRAINY_RECALL_LLM=1`**
**Ingest:** async, 419 turns / 19 sessions, conv-26 (image `query` / `blip_caption` kept as turn text)
**Failure ledger:** [locomo-mh-r4c-dev-1x30.jsonl](./failure-ledger/locomo-mh-r4c-dev-1x30.jsonl)

All 30 scored items answered via product `/recall`. `metrics.errors: 1` is q8 `JUDGE_MISS` (unparseable judge); not counted correct.

Do **not** ship the earlier r4 run `locomo-mh-r4-dev-1x30-20260815` (**6/30**): when-question hops overwrote dated answers. r4b (**17/30**, MH 7/10) recovered when-skip but dumped identity on research and missed the exhibit noun.

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | **0.633 (19/30)** |
| Search p50 ms | 128.4 |
| Search p95 ms | 186.8 |

| Category | Acc | n |
| --- | ---: | --- |
| multi-hop | 0.900 (**9/10**) | 10 |
| open-domain | 0.000 (**0/4**) | 4 |
| temporal | 0.625 (**10/16**) | 16 |

## What the product change actually did

R1b compiled joinable facts but MH still trailed Mem0 6/10 vs 7/10: origin anaphora was not cited, research answers dumped identity, kids pronoun-excited objects did not persist as preferences, and async extract wrote **0** `memory_relations` rows so `follow_relation` had nothing to walk.

This cycle: hops bind typed destinations; when/how-long skip event hops; researching X is a plan atom; they-were-stoked objects are preferences; list hops keep only compatible predicates; the worker projects relation edges; image alt-text is ingested with the turn.

Ledger this run: **7 WRITE_MISS + 2 READER_MISS + 1 RETRIEVAL_MISS + 1 JUDGE_MISS**.

Measured recoveries vs R1b (same 30 items):

| Item | Group | R1b | Now | Mechanism |
| --- | --- | --- | --- | --- |
| q3 research topic | MH | CORRECT on R1b; **WRONG on r4b** | **CORRECT** (`adoption agencies`) | plan atom + scoped hops (no identity dump) |
| q11 origin | MH | READER_MISS | **CORRECT** (Sweden) | origin hop destination, not “home country” |
| q13 career qualifier | MH | WRITE_MISS | **CORRECT** | occupation + identity / population join |
| q19 kids likes | MH | WRITE_MISS | **CORRECT** (nature + dinosaur exhibit) | they-stoked preference + image query text |
| q29 workshop date | temporal | READER_MISS | **CORRECT** | Friday-before stamp cited |

Still wrong (not a text-join hole): q23 second titled work — gold is **not in the source text**. Multimodal WRITE_MISS.

## Compare (do not blend)

| Pin | Overall | MH | OD | Temporal | Notes |
| --- | ---: | ---: | ---: | ---: | --- |
| Gate 0 staging | 18/30 (60.0%) | ~5/10 | 1/4 | — | pre-cutover; different stack |
| Wave 1 local (`a7a5184`) | 14/30 (46.7%) | 3/10 (30.0%) | 2/4 (50.0%) | 9/16 (56.2%) | hybrid reader on |
| R1c local (`21a632b`) | 10/30 (33.3%) | 2/10 (20.0%) | 0/4 (0.0%) | 8/16 (50.0%) | junk atoms crowded provenance |
| Compiler-quality (`d82f7d6`) | 11/30 (36.7%) | 2/10 (20.0%) | 0/4 (0.0%) | 9/16 (56.2%) | junk 45→6 |
| R1b (`5c5f561` / `b23677e`) | 15/30 (50.0%) | 6/10 (60.0%) | 0/4 (0.0%) | 9/16 (56.2%) | WRITE_MISS 15→10; MH trail 6 vs 7 |
| r4 (not shipped) | 6/30 (20.0%) | 5/10 | 0/4 | 1/16 | when-hops overwrote dates |
| r4b (not shipped) | 17/30 (56.7%) | 7/10 (70.0%) | 0/4 | 10/16 (62.5%) | q3 identity dump; q19 exhibit miss |
| **This R4 remasure (`d48e202`)** | **19/30 (63.3%)** | **9/10 (90.0%)** | **0/4 (0.0%)** | **10/16 (62.5%)** | text-join MH; q23 image-gold remains |
| Mem0 Platform freeze 2026-08-13 | 12/30 (40.0%) | 7/10 (70.0%) | 3/4 (75.0%) | 2/16 (12.5%) | same dataset SHA |

## Non-reg (same stack)

- OpMem **13/13 (100%)** — [opmem-mh-r4c-local-20260815.md](./opmem-mh-r4c-local-20260815.md)
- Marketing vertical **17/17 (100%) passed** — [marketing-mh-r4c-local-20260815.md](./marketing-mh-r4c-local-20260815.md)

## Claims

Allowed: lead **this frozen Mem0 same-pin overall** 19/30 (63.3%) vs 12/30 (40.0%); MH **lead** 9/10 (90.0%) vs 7/10 (70.0%); OD **trail** 0/4 vs 3/4; temporal **lead** 10/16 vs 2/16. Remaining MH miss is image-only gold.

Forbidden: SOTA; beats-Mem0 as a general claim; “MH solved”; treating 1×30 as qualification; mixing with 3×90 or Mem0 blog 90+; inventing Graphiti/Zep LoCoMo numbers; claiming q23 from text.
