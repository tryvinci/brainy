# LoCoMo S0 product `/recall` — P9 unproven mh_list dumps — 2026-08-22

Same frozen store as [locomo-s0-diag-mh-135-20260822.md](./locomo-s0-diag-mh-135-20260822.md): tenant `diag-mh-135` + conv-30, dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`. Product SHA `bdee669` (unlock `mh_list` when hops are unproven `search_fallback` dumps; treat 4+ short identity fragments and question-echo hop values as dumps; leftover covering skips OCR captions and stored question prompts). Hybrid **on** (`BRAINY_RECALL_LLM=1`). Go default for `BRAINY_RECALL_LLM` stays **off**.

P8 pair: [locomo-s0-diag-mh-135-p8-20260822.md](./locomo-s0-diag-mh-135-p8-20260822.md) (`86eab77`, **93/180**).

**Not** n=1540. **Not** a Mem0 same-pin. **Not** SOTA. Does not replace integrity 32/180. Does not replace the reader-off 19/180 no-LLM pin. Does not replace README 11.4% / 70% 1×30.

## Scores vs prior pins on this store

| Lane | Overall | multi-hop | open-domain | single-hop | temporal |
| --- | ---: | ---: | ---: | ---: | ---: |
| product reader **off** (`453a929`) | **19/180 (0.106)** | **12/33** | 0/11 | 5/98 | 2/38 |
| product hybrid **on** P5 (`5ad07c4`) | **84/180 (0.467)** | **17/33** | 2/11 | **45/98** | **20/38** |
| product hybrid **on** P6 (`45a83b5`) | **87/180 (0.483)** | **13/33 dip** | **3/11** | **52/98** | **19/38 dip** |
| product hybrid **on** P7 (`f3e0a7f`) | **88/180 (0.489)** | **14/33** | **4/11** | **49/98 dip** | **21/38** |
| product hybrid **on** P8 (`86eab77`) | **93/180 (0.517)** | **15/33** | **4/11** | **52/98** | **22/38** |
| product hybrid **on** P9 (`bdee669`) | **94/180 (0.522)** | **15/33** | **4/11** | **53/98** | **22/38** |
| industry search+harness (reader-off pin) | 62/180 (0.344) | 10/33 | 3/11 | 27/98 | 22/38 |

MH **held 15/33** (still a dip vs P5 **17/33**). SH **52→53**. OD **held 4/11**. Temporal **held 22/38**. Product overall still leads this-VM industry 62/180 on the labeled product lane — still not a Mem0 same-pin.

Item flips vs P8: **+1 / −0 = net +1**.

Named P8 SH recovery: studying/time-management strategy (`conv-48-q120`) — question-echo hop value "Any tips on studying…" no longer mh_list-locks hybrid. Calvin/Dave goals (`conv-50-q118`) still incomplete (hard work/dedication for one entity; gold wants both + determination).

Held P8: Shadow, festival, CS:GO, basketball, walking, UK, gym, chili+ring-toss, injured, community yoga+running, mother's hobbies (no yoga extra). Phuket `conv-48-q77` still MISS (`not in memory`).

## Failure ledger (86 misses)

| Primary | P8 | P9 |
| --- | ---: | ---: |
| PROOF_MISS | 28 | 27 |
| RETRIEVAL_MISS | 29 | 29 |
| READER_MISS | 24 | 24 |
| WRITE_MISS | 4 | 4 |
| HARNESS_ERROR | 2 | 2 |

Largest P9 cells: `single-hop:PROOF_MISS` 21 (was 22), `single-hop:RETRIEVAL_MISS` 12, `multi-hop:RETRIEVAL_MISS` 11, `temporal:READER_MISS` 9, `single-hop:READER_MISS` 9, `multi-hop:READER_MISS` 6. WRITE 4 — do not merge #133.

## What this says

1. Unproven search_fallback hops must not mh_list-lock a covering hybrid answer. Short identity fragment lists and stored question prompts are dumps, not typed joins. Typed 2-item community/skill joins stay locked.
2. Leftover covering skipping OCR captions stops photo-bracket answers; it does not recover write-missing golds (Wolves, Monster Hunter, Wheel of Time).
3. 94/180 is still far from 80% (would be 144/180) and is **not** n=1540. Remaining mass is SH **PROOF 21**. MH **17→15** vs P5 remains a dip. Next product step is nearby-wrong / incomplete dual-entity compose without giving back studying / Shadow / festival / CS:GO / basketball / chili / walking.

Report: `locomo-s0-diag-mh-135-p9-product-recall-s1-abbfef` (summary JSON + failure ledger in this folder). Auto smoke JSON/md dumps are not committed (secret scanner).
