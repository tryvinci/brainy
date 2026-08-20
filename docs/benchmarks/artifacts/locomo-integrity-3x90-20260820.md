# LoCoMo 3×90 — fail-closed integrity stack — 2026-08-20

First three locomo10 conversations, 30 questions each (90 total). Skip-ingest
on frozen tenant `integrity-s0-1`. Same fail-closed ANN stack as S0.
Dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`.
Harness commit `2ea9d12`.

**Not** a Mem0 same-pin. Slice is MH+temporal heavy (36 MH, 45 temporal, 7 OD, 2 SH).

| Lane | Overall | multi-hop | open-domain | single-hop | temporal | p50/p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| product `/recall` | **21/90 (0.233)** | 9/36 | 1/7 | 0/2 | 11/45 | 194 / 242 ms |
| industry search+harness top-k 200 | **33/90 (0.367)** | 8/36 | 2/7 | 1/2 | 22/45 | 186 / 236 ms |

Industry overall **33/90 (36.7%)**, MH **22.2%** matches the 2026-08-11 post-cutover
staging pin on overall/MH. That is a same-n observation on a **different SHA /
extractor / ANN stack**, not a claim that quality is unchanged.

## Failure ledger (P3 order)

| Primary | Product (69 misses) | Industry (57 misses) |
| --- | ---: | ---: |
| PROOF_MISS | 36 | 30 |
| RETRIEVAL_MISS | 19 | 15 |
| READER_MISS | 14 | 11 |
| WRITE_MISS | 0 | 1 |

Sanitized lane reports: [product](./locomo-integrity-3x90-product.md) ·
[industry](./locomo-integrity-3x90-industry.md).
