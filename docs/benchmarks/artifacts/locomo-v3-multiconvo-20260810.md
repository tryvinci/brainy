# LOCOMO smoke — `locomo-v3-multiconvo-20260810`

**Timestamp:** 2026-08-10T13:23:29Z  
**Brainy:** `[REDACTED:BRAINY_BASE_URL]` (commit `e3749de1a6a861c065f477869b225d83d1e21dd2`)  
**Dataset:** [https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json](https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json)  
**SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Answerer:** `[REDACTED:LLM_MODEL]@[REDACTED:BRAINY_EMBEDDING_BASE_URL]`  
**Judge:** `[REDACTED:LLM_MODEL]@[REDACTED:BRAINY_EMBEDDING_BASE_URL]` (temp=0.0)  
**Schema:** mem0ai/memory-benchmarks@UnifiedResult-1.0

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | 0.344 (31/90) |
| Search p50 ms | 169.3 |
| Search p95 ms | 211.0 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.222 | 36 |
| open-domain | 0.286 | 7 |
| single-hop | 0.500 | 2 |
| temporal | 0.444 | 45 |

## Proveability

Pins present (dataset SHA, judge model, brainy URL/commit).

## Outlinks

- Dataset upstream: https://github.com/snap-research/locomo
- Paper: https://aclanthology.org/2024.acl-long.747/
- Harness peer: https://github.com/mem0ai/memory-benchmarks

