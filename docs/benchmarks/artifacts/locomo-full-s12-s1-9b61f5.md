# LOCOMO smoke — `locomo-full-s12-s1-9b61f5`

**Timestamp:** 2026-08-01T07:19:01Z  
**Brainy:** `[REDACTED:BRAINY_BASE_URL]` (commit `ac40806fdef3e3b4097c280ac7db70bc5382202c`)  
**Dataset:** [https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json](https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json)  
**SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Answerer:** `[REDACTED:LLM_MODEL]@[REDACTED:LLM_BASE_URL]`  
**Judge:** `[REDACTED:LLM_MODEL]@[REDACTED:LLM_BASE_URL]` (temp=0.0)  
**Schema:** mem0ai/memory-benchmarks@UnifiedResult-1.0

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | 0.506 (780/1540) |
| Search p50 ms | 1853.0 |
| Search p95 ms | 2817.1 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.277 | 282 |
| open-domain | 0.385 | 96 |
| single-hop | 0.578 | 841 |
| temporal | 0.558 | 321 |

## Proveability

Pins present (dataset SHA, judge model, brainy URL/commit).

## Outlinks

- Dataset upstream: https://github.com/snap-research/locomo
- Paper: https://aclanthology.org/2024.acl-long.747/
- Harness peer: https://github.com/mem0ai/memory-benchmarks

