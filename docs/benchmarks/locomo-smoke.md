# LOCOMO smoke — `locomo-smoke-post-embeddings`

**Timestamp:** 2026-07-19T13:07:25Z  
**Brainy:** `http://127.0.0.1:8080` (commit `74d992fdcced8d9b3be0e3ef962e531f85d9d665`)  
**Dataset:** [https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json](https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json)  
**SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Answerer:** `workers-ai/@cf/openai/gpt-oss-120b@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat`  
**Judge:** `workers-ai/@cf/openai/gpt-oss-120b@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat` (temp=0.0)  
**Schema:** mem0ai/memory-benchmarks@UnifiedResult-1.0

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | 0.133 (4/30) |
| Search p50 ms | 28.8 |
| Search p95 ms | 31.1 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.000 | 10 |
| open-domain | 0.250 | 4 |
| temporal | 0.188 | 16 |

## Proveability

Pins present (dataset SHA, judge model, brainy URL/commit).

## Outlinks

- Dataset upstream: https://github.com/snap-research/locomo
- Paper: https://aclanthology.org/2024.acl-long.747/
- Harness peer: https://github.com/mem0ai/memory-benchmarks

