# LoCoMo S0 dual-lane — fail-closed integrity remasure — 2026-08-19

Invalidates the Aug-19 product **17/180** / industry **52/180** pin. That run
had no pgvector / `embedding_vec_768`, so dense retrieval scanned the last
64–256 writes (~13–19% of each subject). Extraction could silently substitute
the deterministic baseline. This remasure is one frozen tenant, fail-closed
runtime, ANN active, hosted extractor actually running.

**Not** a Mem0 same-pin and **not** a 1×30 / n=1540 quality claim.

| Field | Value |
| --- | --- |
| Tenant | `integrity-s0-1` (skip-ingest; same provider ingest as P4) |
| Dataset SHA | `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` |
| Sample | stratified 180, seed 1, 10 convs |
| SHA (harness) | `5aedc0c` |
| API | local integrity stack, port 18100 |
| ANN | pgvector + `embedding_vec_768`, signatures match, embedder/extractor fallbacks 0 |
| Extractor | hosted gpt-oss-120b via Cloudflare AI Gateway (what `GET /runtime` reports) |
| Embedder | hosted 768-d BGE (model id redacted) |
| Answerer/judge | same gateway path, temp 0.0 |

## Scores

| Lane | Overall | multi-hop | open-domain | single-hop | temporal | search p50/p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| product `/recall` | **32/180 (0.178)** | 1/33 | 1/11 | 19/98 | 11/38 | 168 / 204 ms |
| industry search+harness top-k 200 | **62/180 (0.344)** | 4/33 | 4/11 | 33/98 | 21/38 | 168 / 226 ms |

Invalidated Aug-19 (no ANN, silent extract degrade): product 17/180, industry 52/180.

## Failure ledger (P3 stage order)

Oracle is evidence → retrieval → representation → coverage, with a semantic gold
check. WRITE_MISS is no longer a lexical substring dump.

| Primary | Product (148 misses) | Industry (118 misses) |
| --- | ---: | ---: |
| PROOF_MISS | 112 | 93 |
| RETRIEVAL_MISS | 22 | 18 |
| READER_MISS | 11 | 5 |
| WRITE_MISS | 3 | 2 |

P4 representation coverage on this tenant is **161/180**. Product QA is 32/180
and industry QA is 62/180: the gold is usually written; the packet/reader does
not prove it. Multi-hop coverage 32/33 vs product QA 1/33 is the headline gap.

## Merge gates (this stack)

- OpMem **13/13**
- Marketing **17/17**

## What this is not

- Not n=1540, not Mem0 same-pin, not SOTA, not a license to grow compiler rules.
- Do not mix 32/180 or 62/180 with frozen Mem0 1×30 **12/30** (different n).
- 3×90 / LME-20 follow this pin on the same fail-closed stack when scored.

Reports: [product](./locomo-integrity-s0-product-recall-s1-dee145.md) ·
[industry](./locomo-integrity-s0-industry-search-s1-6c38c5.md) ·
[summary](./locomo-integrity-s0-summary-20260819.json).
