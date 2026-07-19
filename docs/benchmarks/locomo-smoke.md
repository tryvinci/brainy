# LOCOMO smoke — `locomo-smoke-post-eng172-async`

**Timestamp:** 2026-07-19T12:45:45Z  
**Brainy:** `http://127.0.0.1:8080` (commit `a189db805d6f9eb1ce2f3a17ae101fcd73f6ab0f`)  
**Dataset:** [https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json](https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json)  
**SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Answerer:** `workers-ai/@cf/openai/gpt-oss-120b@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat`  
**Judge:** `workers-ai/@cf/openai/gpt-oss-120b@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat` (temp=0.0)  
**Schema:** mem0ai/memory-benchmarks@UnifiedResult-1.0

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | 0.200 (6/30) |
| Search p50 ms | 7.6 |
| Search p95 ms | 10.1 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.100 | 10 |
| open-domain | 0.750 | 4 |
| temporal | 0.125 | 16 |

## Proveability

Pins present (dataset SHA, judge model, brainy URL/commit).


## Notes (same-pin remeasure)

| Run | Commit / path | Overall | Retrieval-miss (of 30) |
| --- | --- | ---: | ---: |
| Baseline (pre ENG-171) | `e8b3bb8` era | **2/30 (6.7%)** | 28 |
| Post ENG-171 | `1dc738c` sync | **4/30 (13.3%)** | 25 |
| Post ENG-172 async+merge | `locomo-smoke-post-eng172-async` | **6/30 (20.0%)** | 18 |

Pins held fixed: 1 conversation / 30 Q / `gpt-oss-120b` judge+answerer / same dataset SHA.

**Failure taxonomy (6/30 run):** 18 retrieval-miss of GT span, 6 answer/judge miss (retrieval had related text), 0 empty retrieval.

**Product notes:** Async ingest + worker resilience landed; provider LLM still flakes (JSON truncate) and soft-degrades to conversational episodes. Remaining misses are mostly rank/temporal resolution (relative dates like "yesterday" vs absolute GT) and multi-hop synthesis — next: embeddings + temporal query/filter.

## Outlinks

- Dataset upstream: https://github.com/snap-research/locomo
- Paper: https://aclanthology.org/2024.acl-long.747/
- Harness peer: https://github.com/mem0ai/memory-benchmarks

