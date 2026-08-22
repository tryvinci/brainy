# LoCoMo S0 product `/recall` — P1 hybrid reader A/B — 2026-08-22

Same frozen store as [locomo-s0-diag-mh-135-20260822.md](./locomo-s0-diag-mh-135-20260822.md): tenant `diag-mh-135` + conv-30, dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`. Product SHA `3d42b17` (enumerate hybrid + max_tokens). Hybrid **on** (`BRAINY_RECALL_LLM=1`). Date / where / polar locks still applied.

**Not** n=1540. **Not** a Mem0 same-pin. **Not** SOTA. Does not replace integrity 32/180.

## Scores vs reader-off baseline on this store

| Lane | Overall | multi-hop | open-domain | single-hop | temporal |
| --- | ---: | ---: | ---: | ---: | ---: |
| product reader **off** (`453a929`) | **19/180 (0.106)** | **12/33** | 0/11 | 5/98 | 2/38 |
| product hybrid **on** (`3d42b17`) | **37/180 (0.206)** | **10/33** | 1/11 | **19/98** | **7/38** |
| industry search+harness (reader-off pin) | 62/180 (0.344) | 10/33 | 3/11 | 27/98 | 22/38 |

MH **12→10 is a dip** (lost clarinet/violin, dual collectible, community join, snack filter, Ferrari count; gained polar-teach, family-injury who, who-supports). SH **5→19** and temporal **2→7** are the attributed P1 moves. OD 0→1.

## Failure ledger (143 misses)

| Primary | Reader-off | Hybrid on |
| --- | ---: | ---: |
| PROOF_MISS | 59 | 44 |
| READER_MISS | 52 | 49 |
| RETRIEVAL_MISS | 39 | 39 |
| WRITE_MISS | 10 | 10 |
| HARNESS_ERROR | 1 | 1 |

Largest hybrid-on cells: `single-hop:PROOF_MISS` 37, `temporal:READER_MISS` 22, `single-hop:READER_MISS` 19, `single-hop:RETRIEVAL_MISS` 19. Temporal READER is still mostly locked calendar dates (`25 May 2023` vs `Sunday before 25 May 2023`).

## What this says

1. Hybrid-on product doubled this-VM `/recall` (19→37) and is still far from 80% and from industry 62/180.
2. P2 is **justified** for SH/temporal, but must **not** keep overwriting typed MH lists/counts.
3. WRITE is still 10 — do not merge #133.

Report: `locomo-s0-diag-mh-135-llm-product-recall-s1-edff06` (summary JSON + failure ledger in this folder).
