# LoCoMo S0 product `/recall` — P2b extras-lock + skip hop-ground — 2026-08-22

Same frozen store as [locomo-s0-diag-mh-135-20260822.md](./locomo-s0-diag-mh-135-20260822.md): tenant `diag-mh-135` + conv-30, dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`. Product SHA `fb41ece` (comma-split extras coverage; lock multi-item extras / short-list expansion; do not hop-ground enumerated hybrid lists). Hybrid **on** (`BRAINY_RECALL_LLM=1`). Where / polar stay locked. Count / dual-entity `mh_list` locks stay.

Length-lock pair: [locomo-s0-diag-mh-135-p2-20260822.md](./locomo-s0-diag-mh-135-p2-20260822.md) (`681028e`, **56/180**).

**Not** n=1540. **Not** a Mem0 same-pin. **Not** SOTA. Does not replace integrity 32/180. Does not replace the reader-off 19/180 no-LLM pin.

## Scores vs prior pins on this store

| Lane | Overall | multi-hop | open-domain | single-hop | temporal |
| --- | ---: | ---: | ---: | ---: | ---: |
| product reader **off** (`453a929`) | **19/180 (0.106)** | **12/33** | 0/11 | 5/98 | 2/38 |
| product hybrid **on** P1 (`3d42b17`) | **37/180 (0.206)** | **10/33** | 1/11 | **19/98** | **7/38** |
| product hybrid **on** P2 length-lock (`681028e`) | **56/180 (0.311)** | **11/33** | 1/11 | **23/98** | **21/38** |
| product hybrid **on** P2b (`fb41ece`) | **61/180 (0.339)** | **16/33** | 1/11 | **25/98** | **19/38** |
| industry search+harness (reader-off pin) | 62/180 (0.344) | 10/33 | 3/11 | 27/98 | 22/38 |

MH **12→10→11→16**. The P1 MH dip is **closed** on this store (clarinet/violin, Tim+John collectible, Sam snacks recovered; Ferrari count stayed 2). Temporal **21→19 is a dip** vs P2 length-lock (mentorship date 10 July vs weekend of 15–16 July; art-show day). SH **23→25**. OD stuck at 1/11.

## Failure ledger (119 misses)

| Primary | P1 | P2 | P2b |
| --- | ---: | ---: | ---: |
| PROOF_MISS | 44 | 41 | 42 |
| RETRIEVAL_MISS | 39 | 39 | 38 |
| READER_MISS | 49 | 33 | 28 |
| WRITE_MISS | 10 | 10 | 10 |
| HARNESS_ERROR | 1 | 1 | 1 |

Largest P2b cells: `single-hop:PROOF_MISS` 36, `single-hop:RETRIEVAL_MISS` 19, `single-hop:READER_MISS` 15, `multi-hop:RETRIEVAL_MISS` 12, `temporal:READER_MISS` 9. WRITE stays 10 — do not merge #133.

## What this says

1. Skip hop-grounding on enumerated hybrid answers is the mechanism for the MH recovery: hybrid returned `clarinet, violin` / `Soda and candy`, and P2 `groundToHopValues` had been replacing those subsets with 6-slot dumps (`pottery, beach`).
2. 61/180 is still far from 80% and still trails this-VM industry 62/180 on overall. Product MH **16/33** now leads this-VM industry MH **10/33**. OD remains the trail axis on this 180.
3. Temporal 21→19 is a real dip to name, not noise to hide. SH PROOF 36 is still the mass for P3.

Report: `locomo-s0-diag-mh-135-p2b-product-recall-s1-59d01f` (summary JSON + failure ledger in this folder). Auto smoke JSON/md dumps are not committed (secret scanner).
