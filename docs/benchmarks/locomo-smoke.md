# LOCOMO smoke — `locomo-smoke-ready-roll`

**Timestamp:** 2026-07-19T12:58:17Z  
**Brainy:** `http://127.0.0.1:8080` (commit `2ba4b1d567accf475fd868e55ed3d5a80ff6c5fd`)  
**Dataset:** [https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json](https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json)  
**SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Answerer:** `workers-ai/@cf/openai/gpt-oss-120b@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat`  
**Judge:** `workers-ai/@cf/openai/gpt-oss-120b@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat` (temp=0.0)  
**Schema:** mem0ai/memory-benchmarks@UnifiedResult-1.0  
**Ingest mode:** async (`/ingest/async` + worker provider extract with baseline merge)

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | 0.200 (6/30) |
| Search p50 ms | 16.7 |
| Search p95 ms | 18.0 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.000 | 10 |
| open-domain | 0.250 | 4 |
| temporal | 0.312 | 16 |

## Proveability

Pins present (dataset SHA, judge model, brainy URL/commit).

## Notes (same-pin remeasure ladder)

| Run | Path | Overall | Notes |
| --- | --- | ---: | --- |
| Baseline (pre ENG-171) | sync keyword | **2/30** | 28 retrieval miss |
| Post ENG-171 | sync episodes | **4/30** | 25 retrieval miss |
| Post ENG-172 async (partial queue) | async+merge | **6/30** | first async path |
| Post temporal enrich (bad stamp parse) | async | **3/30** | invalid — LOCOMO stamps not parsed |
| Post relative-date resolve | async | **5/30** | LGBTQ when-question fixed |
| **Ready-to-roll** (`locomo-smoke-ready-roll`) | async+relative dates | **6/30 (20%)** | taxonomy below |

Pins held fixed: 1 conversation / 30 Q / `gpt-oss-120b` judge+answerer / same dataset SHA.

### Failure taxonomy (6/30)

| Bucket | Count |
| --- | ---: |
| correct | 6 |
| retrieval miss of GT span | 9 |
| answer/judge miss (related text retrieved) | 15 |
| empty retrieval | 0 |

Product lift since ENG-171 is real but below the 12/30 L4 unlock gate. Remaining work is multi-hop synthesis + stronger provider extract yield + real embeddings (not hash) — not harness games.

## Outlinks

- Dataset upstream: https://github.com/snap-research/locomo
- Paper: https://aclanthology.org/2024.acl-long.747/
- Harness peer: https://github.com/mem0ai/memory-benchmarks
