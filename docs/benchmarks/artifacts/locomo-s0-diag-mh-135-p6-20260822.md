# LoCoMo S0 product `/recall` — P6 dump-lock skip — 2026-08-22

Same frozen store as [locomo-s0-diag-mh-135-20260822.md](./locomo-s0-diag-mh-135-20260822.md): tenant `diag-mh-135` + conv-30, dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`. Product SHA `45a83b5` (skip dual-entity **activity** dumps unless hops are a typed skill/possession/preference join; unlock hybrid when the typed answer is a hop dump; cap the hybrid prompt; promote proper-noun/venue facts ahead of generic leftover-cover; do not compose crowded hop dumps when hybrid abstains). Hybrid **on** (`BRAINY_RECALL_LLM=1`). Where / polar stay locked except when the typed answer is itself a hop dump. Count / dual-entity `mh_list` locks stay for short typed joins. Enumerated hop-ground skip from P2b stays. Distinctive-token admit from P3 stays. Identity/garbage skip from P4 stays. Activity-dump skip from P5 stays.

P5 pair: [locomo-s0-diag-mh-135-p5-20260822.md](./locomo-s0-diag-mh-135-p5-20260822.md) (`5ad07c4`, **84/180**).

**Not** n=1540. **Not** a Mem0 same-pin. **Not** SOTA. Does not replace integrity 32/180. Does not replace the reader-off 19/180 no-LLM pin.

## Scores vs prior pins on this store

| Lane | Overall | multi-hop | open-domain | single-hop | temporal |
| --- | ---: | ---: | ---: | ---: | ---: |
| product reader **off** (`453a929`) | **19/180 (0.106)** | **12/33** | 0/11 | 5/98 | 2/38 |
| product hybrid **on** P1 (`3d42b17`) | **37/180 (0.206)** | **10/33** | 1/11 | **19/98** | **7/38** |
| product hybrid **on** P2 length-lock (`681028e`) | **56/180 (0.311)** | **11/33** | 1/11 | **23/98** | **21/38** |
| product hybrid **on** P2b (`fb41ece`) | **61/180 (0.339)** | **16/33** | 1/11 | **25/98** | **19/38** |
| product hybrid **on** P3 (`5bc28ea`) | **73/180 (0.406)** | **16/33** | **3/11** | **32/98** | **22/38** |
| product hybrid **on** P4 (`6f74024`) | **79/180 (0.439)** | **16/33** | **3/11** | **37/98** | **23/38** |
| product hybrid **on** P5 (`5ad07c4`) | **84/180 (0.467)** | **17/33** | **2/11** | **45/98** | **20/38** |
| product hybrid **on** P6 (`45a83b5`) | **87/180 (0.483)** | **13/33** | **3/11** | **52/98** | **19/38** |
| industry search+harness (reader-off pin) | 62/180 (0.344) | 10/33 | 3/11 | 27/98 | 22/38 |

MH **17→13 dip**. SH **45→52**. OD **2→3** (recovers the P5 OD dip). Temporal **20→19 dip**. Product overall still leads this-VM industry 62/180 on the labeled product lane — still not a Mem0 same-pin.

Item flips vs P5: **+12 / −9 = net +3**.

Named P6 SH/OD gains: Andrew+Buddy walking (`conv-44-q110`, sushi dump → walks); James after Toronto Vancouver (`conv-47-q104`); Yoga (`conv-49-q152`); partner pregnant (`conv-49-q133`); Frank Ocean festival (`conv-50-q125`); Melanie LGBTQ (`conv-26-q30`). Tim-UK `conv-43-q38` **held**.

Named MH losses: Maria fundraiser events incomplete (`conv-41-q29`); Tim and John signed basketball → not in memory (`conv-43-q25`); Deborah's mother's hobbies incomplete (`conv-48-q14`); Jolene/partner Phuket diving → not in memory (`conv-48-q77`); Evan family injured incomplete (`conv-49-q49`).

Locks held: `clarinet, violin`; Ferrari `2`; strawberry filling; gym; Coco/Shadow; Tim-UK United Kingdom; Audrey Max.

## Failure ledger (93 misses)

| Primary | P5 | P6 |
| --- | ---: | ---: |
| PROOF_MISS | 32 | 28 |
| RETRIEVAL_MISS | 29 | 28 |
| READER_MISS | 26 | 29 |
| WRITE_MISS | 7 | 6 |
| HARNESS_ERROR | 2 | 2 |

Largest P6 cells: `single-hop:PROOF_MISS` 22 (was 25), `single-hop:RETRIEVAL_MISS` 12, `temporal:READER_MISS` 11 (was 10), `multi-hop:RETRIEVAL_MISS` 10, `multi-hop:READER_MISS` 9 (MH dip). WRITE 6 — do not merge #133.

## What this says

1. P5 skipped activity dumps for **one** entity. Two-name SH questions still planned as multi-hop, so `mh_list` / `where` locks kept sushi and title-case place dumps over covering hybrid answers. Unlocking those dumps and ranking proper-noun/venue lines ahead of generic leftover-cover (visa "countries he wants to visit") recovers walking, Vancouver, Yoga, pregnant, and keeps Tim-UK/gym.
2. 87/180 is still far from 80% (would be 144/180) and is **not** n=1540. SH PROOF 22 remains the mass. MH **17→13** is a named dip — next product step is recover those dual-entity typed joins without giving back SH 52.
3. Fair Mem0 Platform 180 (`fair2`) **died on HTTP 429 usage quota** (SEARCH quota 1000/1000, reset 2026-09-01). No same-n Mem0 pin this cycle. Do not refresh lead/trail from 21/30 vs 11/30.

Report: `locomo-s0-diag-mh-135-p6-product-recall-s1-b22074` (summary JSON + failure ledger in this folder). Auto smoke JSON/md dumps are not committed (secret scanner).
