# LoCoMo S0 product `/recall` — P3 distinctive query-token admit — 2026-08-22

Same frozen store as [locomo-s0-diag-mh-135-20260822.md](./locomo-s0-diag-mh-135-20260822.md): tenant `diag-mh-135` + conv-30, dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`. Product SHA `5bc28ea` (admit leftover distinctive query tokens into the evidence set; merge covering extras ahead of a full original top-k; second-pass probe on the uncovered token; do not compose or prompt from unproven `search_fallback` hops). Hybrid **on** (`BRAINY_RECALL_LLM=1`). Where / polar stay locked. Count / dual-entity `mh_list` locks stay. Enumerated hop-ground skip from P2b stays.

P2b pair: [locomo-s0-diag-mh-135-p2b-20260822.md](./locomo-s0-diag-mh-135-p2b-20260822.md) (`fb41ece`, **61/180**).

**Not** n=1540. **Not** a Mem0 same-pin. **Not** SOTA. Does not replace integrity 32/180. Does not replace the reader-off 19/180 no-LLM pin.

## Scores vs prior pins on this store

| Lane | Overall | multi-hop | open-domain | single-hop | temporal |
| --- | ---: | ---: | ---: | ---: | ---: |
| product reader **off** (`453a929`) | **19/180 (0.106)** | **12/33** | 0/11 | 5/98 | 2/38 |
| product hybrid **on** P1 (`3d42b17`) | **37/180 (0.206)** | **10/33** | 1/11 | **19/98** | **7/38** |
| product hybrid **on** P2 length-lock (`681028e`) | **56/180 (0.311)** | **11/33** | 1/11 | **23/98** | **21/38** |
| product hybrid **on** P2b (`fb41ece`) | **61/180 (0.339)** | **16/33** | 1/11 | **25/98** | **19/38** |
| product hybrid **on** P3 (`5bc28ea`) | **73/180 (0.406)** | **16/33** | **3/11** | **32/98** | **22/38** |
| industry search+harness (reader-off pin) | 62/180 (0.344) | 10/33 | 3/11 | 27/98 | 22/38 |

MH **held 16/33** (gain `conv-48-q73` polar yes; loss `conv-26-q52` pet names Oliver/Luna/Bailey). SH **25→32**. OD **1→3**. Temporal **19→22** (recovers P2b's mentorship-date dip `conv-26-q36`; new loss `conv-44-q38` weekend-before date). Product overall **now leads this-VM industry 62/180 on the labeled product lane** — still not a Mem0 same-pin.

## Failure ledger (107 misses)

| Primary | P2b | P3 |
| --- | ---: | ---: |
| PROOF_MISS | 42 | 34 |
| RETRIEVAL_MISS | 38 | 29 |
| READER_MISS | 28 | 34 |
| WRITE_MISS | 10 | 8 |
| HARNESS_ERROR | 1 | 2 |

Largest P3 cells: `single-hop:PROOF_MISS` 28 (was 36), `single-hop:READER_MISS` 22 (was 15 — **dip**), `single-hop:RETRIEVAL_MISS` 12 (was 19), `multi-hop:RETRIEVAL_MISS` 11. WRITE 10→8 — do not merge #133. The two `HARNESS_ERROR` rows are oracle mislabels on `not in memory` (`conv-42-q146`, `conv-48-q116`), not harness crashes.

## What this says

1. Distinctive-token admit is the mechanism: compiled facts such as strawberry filling and "joined a gym" were in Postgres but absent from the original top-30 (FTS AND / name-recency ILIKE). After admit they rank (filling at slot 0) and hybrid answers them. Unproven hop dumps no longer replace those answers.
2. 73/180 is still far from 80% (would be 144/180) and is **not** n=1540. SH PROOF 28 + SH READER 22 remain the mass. Wheel of Time / self-care stay misses (gold not in query tokens, or reader picks a nearby wrong fact).
3. MH 16/33 held with a named 1-for-1 swap. Temporal 22/38 matches this-VM industry temporal. Reader-off 19/180 remains the labeled no-LLM product pin.

Report: `locomo-s0-diag-mh-135-p3-product-recall-s1-42c2dd` (summary JSON + failure ledger in this folder). Auto smoke JSON/md dumps are not committed (secret scanner).
