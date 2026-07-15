# LOCOMO smoke — `locomo-smoke-post-eng171`

**Timestamp:** 2026-07-15T17:21:54Z  
**Brainy:** `https://brainy-api-staging.onrender.com` (commit `1dc738cbd66c68b359b592f149f1a818b8bbd631`)  
**Dataset:** [https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json](https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json)  
**SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Answerer:** `workers-ai/@cf/openai/gpt-oss-120b@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat`  
**Judge:** `workers-ai/@cf/openai/gpt-oss-120b@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat` (temp=0.0)  
**Schema:** mem0ai/memory-benchmarks@UnifiedResult-1.0

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | 0.133 (4/30) |
| Search p50 ms | 422.7 |
| Search p95 ms | 543.0 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.000 | 10 |
| open-domain | 0.500 | 4 |
| temporal | 0.125 | 16 |

## Proveability

Pins present (dataset SHA, judge model, brainy URL/commit).

## Notes (same-pin remeasure)

| Run | Commit | Overall | Retrieval-miss (of 30) |
| --- | --- | ---: | ---: |
| Baseline (pre ENG-171) | `e8b3bb8` era | **2/30 (6.7%)** | 28 |
| Post ENG-171 | `1dc738c` | **4/30 (13.3%)** | 25 |

Pins held fixed: 1 conversation / 30 Q / `gpt-oss-120b` judge+answerer / same dataset SHA.

Product lift is real but small — remaining failures are still mostly **retrieval miss of GT span**. Next product track: ENG-172/92 (provider extract) + ENG-173 (ranking).

## Outlinks

- Dataset upstream: https://github.com/snap-research/locomo
- Paper: https://aclanthology.org/2024.acl-long.747/
- Harness peer: https://github.com/mem0ai/memory-benchmarks

