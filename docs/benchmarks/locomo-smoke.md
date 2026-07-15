# LOCOMO smoke — `locomo-smoke-gpt-oss-120b`

**Timestamp:** 2026-07-15T05:20:20Z  
**Brainy:** `https://brainy-api-staging.onrender.com` (commit `e8b3bb89e29f9c39992c91d8dc3a017887b5b7ab`)  
**Dataset:** [https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json](https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json)  
**SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Answerer:** `workers-ai/@cf/openai/gpt-oss-120b@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat`  
**Judge:** `workers-ai/@cf/openai/gpt-oss-120b@https://gateway.ai.cloudflare.com/v1/9c8c17cfa5b8c29c830e072acce42a3d/brainy-staging/compat` (temp=0.0)  
**Schema:** mem0ai/memory-benchmarks@UnifiedResult-1.0

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | 0.067 (2/30) |
| Search p50 ms | 270.0 |
| Search p95 ms | 681.2 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.100 | 10 |
| open-domain | 0.250 | 4 |
| temporal | 0.000 | 16 |

## Proveability

Pins present (dataset SHA, judge model, brainy URL/commit).

## Notes (2026-07-15 smoke)

Honest mid/low result — expected for first LOCOMO pass, **not** a Mem0 head-to-head.

**Failure taxonomy (30 Qs):** 2 correct · **28 retrieval miss** (GT span absent from top-k) · 0 “had evidence but bad answer.”

**Product root causes (not harness games):**

1. Deterministic keyword extractor (`internal/memory/extractor.go`) — most casual dialogue turns never become memories; date clauses detach from topical clauses.
2. Search ranks topical hash/ILIKE overlap highly; episodic facts under-boosted.
3. No first-class event/session time — only ingest `UpdatedAt` recency → temporal 0/16 is expected.

**Policy:** fix these as product capabilities for all conversational clients. Do not LOCOMO-special-case. Re-measure with the same pins after product work.

## Outlinks

- Dataset upstream: https://github.com/snap-research/locomo
- Paper: https://aclanthology.org/2024.acl-long.747/
- Harness peer: https://github.com/mem0ai/memory-benchmarks

