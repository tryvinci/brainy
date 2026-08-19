# Extraction ceiling — deterministic vs strict provider (2026-08-19)

Semantic gold coverage on the representation oracle, **not** QA accuracy.
Same LoCoMo 10-conv / stratified 180 (seed 1) as Aug-19 S0.
Dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`.

**Stack:** fail-closed integrity API `:18100`, pgvector active, 768-d hosted embedder.
Extractor identity is Cloudflare gpt-oss-120b via AI Gateway (what `GET /runtime` reports). There is no OpenAI key in this environment.

| Arm | Tenant | Turns | Coverage |
| --- | --- | ---: | ---: |
| deterministic (sync `/ingest`) | `ceil-det-1` | 5882 | **139/180 (0.772)** |
| strict provider (async worker) | `integrity-s0-1` | 5882 | **161/180 (0.894)** |

Provider arm: 1571/1571 extraction jobs completed, 22,509 memories, all 768-d ANN rows. No silent baseline substitution.

## By group

| Group | Deterministic | Provider |
| --- | ---: | ---: |
| multi-hop | 24/33 (0.727) | **32/33 (0.970)** |
| open-domain | 1/11 (0.091) | **5/11 (0.455)** |
| single-hop | 82/98 (0.837) | **90/98 (0.918)** |
| temporal | 32/38 (0.842) | **34/38 (0.895)** |

Facts per question (representation blob): det mean 1026; provider mean 1613.

## What this answers

The Aug-19 S0 pin was riding a **regex compiler whenever the hosted extractor returned empty/unparseable JSON**. Fail-closed mode exposed that (max_tokens=256, kind/value JSON mismatch). After those parser holes were fixed, the same 180 items show the provider covering **every** deterministic hit plus **22** of the 41 det misses (19 remain uncovered on both).

So the benchmark has been riding the regex compiler **when the provider failed**. When the provider actually runs, coverage is higher — especially multi-hop (24→32). Open-domain is still weak (5/11). Do not add a new regex batch from this; the leftover 19 are the P3 audit problem, not a license to grow `providerSystemPrompt`.

## Fail-closed holes found while measuring

- Cloudflare default `max_tokens=256` → gpt-oss reasoning-only, `content=null`.
- `kind` field used for taxonomy slots (`plan`, `belief`).
- Numeric `"value": 2` rejected by a string-only JSON field.

Item rows: [extraction-ceiling-det-20260819.json](extraction-ceiling-det-20260819.json), [extraction-ceiling-prov-20260819.json](extraction-ceiling-prov-20260819.json).
