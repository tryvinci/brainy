# LOCOMO smoke — `locomo-smoke-entity-linking`

**Timestamp:** 2026-07-20T02:03:20Z  
**Brainy:** `http://127.0.0.1:8080` (commit `a8cdec8821313bf7098670e30fe8a8f4d1ea0248`)  
**Dataset:** [https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json](https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json)  
**SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Answerer:** `[REDACTED]@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat`  
**Judge:** `[REDACTED]@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat` (temp=0.0)  
**Schema:** mem0ai/memory-benchmarks@UnifiedResult-1.0

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | 0.300 (9/30) |
| Search p50 ms | 31.9 |
| Search p95 ms | 44.0 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.200 | 10 |
| open-domain | 0.500 | 4 |
| temporal | 0.312 | 16 |

## Proveability

Pins present (dataset SHA, judge model, brainy URL/commit).

## Outlinks

- Dataset upstream: https://github.com/snap-research/locomo
- Paper: https://aclanthology.org/2024.acl-long.747/
- Harness peer: https://github.com/mem0ai/memory-benchmarks

