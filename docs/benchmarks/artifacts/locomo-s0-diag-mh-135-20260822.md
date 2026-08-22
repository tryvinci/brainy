# LoCoMo S0 dual-lane — this-VM `diag-mh-135` + conv-30 — 2026-08-22

**Not** `integrity-s0-1`. **Not** a Mem0 same-pin. **Not** n=1540. Product `/recall` SHA for the binary is **`453a929`** (#135 merge). Harness SHA is `98d5db8`.

Tenant `diag-mh-135` on local `:18100` / DB `brainy_mh`. Frozen 10/10 conversations after ingesting `conv-30` (369 turns). One extract job of four `session_17` turns failed (provider body started with `!`); 98/99 conv-30 jobs completed. ANN active, signatures match, embedder/extractor fallbacks 0. Hybrid reader **off** (`BRAINY_RECALL_LLM` unset).

Dataset SHA: `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
Sample: stratified 180, seed 1 (98 SH / 33 MH / 38 temporal / 11 OD)

## Scores

| Lane | Overall | multi-hop | open-domain | single-hop | temporal |
| --- | ---: | ---: | ---: | ---: | ---: |
| product `/recall` top-k 30 | **19/180 (0.106)** | **12/33** | **0/11** | **5/98** | **2/38** |
| industry search+harness top-k 200 | **62/180 (0.344)** | 10/33 | 3/11 | 27/98 | 22/38 |

Integrity-VM S0 (different tenant, do not mix): product **32/180**, industry **62/180**. Industry matched. Product is lower here with the reader off; MH 12/33 matches the labeled diagnostic skip-ingest on this tenant.

## Failure ledger (product 161 misses / industry 118)

| Primary | Product | Industry |
| --- | ---: | ---: |
| PROOF_MISS | 59 | 43 |
| READER_MISS | 52 | 30 |
| RETRIEVAL_MISS | 39 | 35 |
| WRITE_MISS | 10 | 8 |
| HARNESS_ERROR | 1 | 1 |
| JUDGE_MISS | 0 | 1 |

Product by category×stage (largest): `single-hop:PROOF_MISS` 48, `temporal:READER_MISS` 24, `single-hop:READER_MISS` 22, `single-hop:RETRIEVAL_MISS` 19, `multi-hop:RETRIEVAL_MISS` 13.

## What this says

1. **MH-only is not the 80% path.** Product MH 12/33 is the #135 slot-recovery pin at S0 n. Perfect MH still leaves SH 5/98 and temporal 2/38.
2. **Industry 62/180 is the current-SHA search+harness ceiling on this store** (same overall as integrity). Historical 49.8% is still not this SHA.
3. **Product vs industry is 19 vs 62.** Reader-off `/recall` is the labeled gap. P1 reader A/B is justified; P2 is not committed until that A/B moves SH/temporal.
4. WRITE is 10/180 here vs 3/180 on integrity — still not the mass. Do not merge #133.

Reports: [product](./locomo-s0-diag-mh-135-product-recall-s1-3973ae.md) · [industry](./locomo-s0-diag-mh-135-industry-search-s1-d59657.md) · [summary](./locomo-s0-diag-mh-135-summary.json)
