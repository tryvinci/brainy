# LoCoMo S0 product `/recall` — P2-narrow length-lock — 2026-08-22

Same frozen store as [locomo-s0-diag-mh-135-20260822.md](./locomo-s0-diag-mh-135-20260822.md): tenant `diag-mh-135` + conv-30, dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`. Product SHA `681028e` (count / dual-entity / non-shortening list locks; when-event date unlock). Hybrid **on** (`BRAINY_RECALL_LLM=1`). Where / polar stay locked.

This is the **length-lock** remasure. A later extras-lock + skip hop-ground SHA (`fb41ece`) is remasured separately (`locomo-s0-diag-mh-135-p2b`).

**Not** n=1540. **Not** a Mem0 same-pin. **Not** SOTA. Does not replace integrity 32/180. Does not replace the reader-off 19/180 no-LLM pin.

## Scores vs prior pins on this store

| Lane | Overall | multi-hop | open-domain | single-hop | temporal |
| --- | ---: | ---: | ---: | ---: | ---: |
| product reader **off** (`453a929`) | **19/180 (0.106)** | **12/33** | 0/11 | 5/98 | 2/38 |
| product hybrid **on** P1 (`3d42b17`) | **37/180 (0.206)** | **10/33** | 1/11 | **19/98** | **7/38** |
| product hybrid **on** P2 (`681028e`) | **56/180 (0.311)** | **11/33** | 1/11 | **23/98** | **21/38** |
| industry search+harness (reader-off pin) | 62/180 (0.344) | 10/33 | 3/11 | 27/98 | 22/38 |

MH **12→10→11** is still a **dip** vs reader-off 12/33. SH **5→19→23**. Temporal **2→7→21** is the attributed P2 move (weekday/weekend-relative dates vs the collapsed `25 May 2023` lock). OD stuck at 1/11.

## Failure ledger (124 misses)

| Primary | P1 hybrid | P2 length-lock |
| --- | ---: | ---: |
| PROOF_MISS | 44 | 41 |
| RETRIEVAL_MISS | 39 | 39 |
| READER_MISS | 49 | 33 |
| WRITE_MISS | 10 | 10 |
| HARNESS_ERROR | 1 | 1 |

Largest P2 cells: `single-hop:PROOF_MISS` 35, `single-hop:RETRIEVAL_MISS` 19, `single-hop:READER_MISS` 18, `multi-hop:RETRIEVAL_MISS` 13, `temporal:READER_MISS` 7 (was 22). WRITE stays 10 — do not merge #133.

## What this says

1. P2 length-lock moved temporal 7→21/38 on this store. That is product `/recall` with hybrid on, not judge drift: P0 collapsed many “when” answers to the wrong calendar day; P2 emitted weekend/Friday-relative dates (e.g. weekend of 15–16 July vs gold “weekend before 17 July”).
2. Count lock recovered Ferrari **2**. Dual-entity MH lock recovered Tim+John vs a pottery dump, but the typed join still carries extra values (Harry Potter). Instruments / snacks still lost: extras vs typed was 0, then hop-slot grounding re-expanded a hybrid subset into a 6-item dump.
3. 56/180 is still far from 80% and from industry 62/180. MH 11/33 is not the #135 recovery pin.

Report: `locomo-s0-diag-mh-135-p2-product-recall-s1-701b00` (summary JSON + failure ledger in this folder). Auto smoke JSON/md dumps are not committed (secret scanner).
