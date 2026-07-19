# LOCOMO smoke — `locomo-smoke-post-phase3`

**Timestamp:** 2026-07-19T13:10:41Z  
**Brainy:** `http://127.0.0.1:8080` (commit `185ff48d06df3079fe346b809f3e73695718c3b6`)  
**Dataset:** [https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json](https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json)  
**SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Answerer:** `workers-ai/@cf/openai/gpt-oss-120b@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat`  
**Judge:** `workers-ai/@cf/openai/gpt-oss-120b@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat` (temp=0.0)  
**Schema:** mem0ai/memory-benchmarks@UnifiedResult-1.0  
**Ingest mode:** async (`/ingest/async` + worker provider extract with baseline merge)  
**Embeddings:** local hash (provider embeddings soft-degrade; CF gateway `/embeddings` returns 403 in this env)

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | 0.233 (7/30) |
| Search p50 ms | 17.2 |
| Search p95 ms | 23.8 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.000 | 10 |
| open-domain | 0.750 | 4 |
| temporal | 0.250 | 16 |

## Proveability

Pins present (dataset SHA, judge model, brainy URL/commit).

## Notes (same-pin remeasure ladder)

| Run | Path | Overall | Notes |
| --- | --- | ---: | --- |
| Pre ENG-171 | sync keyword | **2/30** | 28 retrieval miss |
| Post ENG-171 | sync episodes | **4/30** | 25 retrieval miss |
| Post async+relative dates | async | **6/30** | temporal lift |
| Uncapped session expand | async | **4/30** | diluted exact-span (reverted) |
| **Phase 3** (`locomo-smoke-post-phase3`) | async + embedder iface + capped multi-hop neighbors | **7/30 (23.3%)** | best so far |

Pins held fixed: 1 conversation / 30 Q / `gpt-oss-120b` judge+answerer / same dataset SHA.

### Failure taxonomy (7/30)

| Bucket | Count |
| --- | ---: |
| correct | 7 |
| retrieval miss of GT span | 10 |
| answer/judge miss (related text retrieved) | 13 |
| empty retrieval | 0 |

L4 unlock gate (≥12/30) still unmet. Provider `/v1/embeddings` not usable on current CF gateway (403); ship interface + local fallback. Next: working embedding provider + multi-hop synthesis.

## Outlinks

- Dataset upstream: https://github.com/snap-research/locomo
- Paper: https://aclanthology.org/2024.acl-long.747/
- Harness peer: https://github.com/mem0ai/memory-benchmarks
