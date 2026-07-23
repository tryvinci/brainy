# Path to SOTA conversational memory

Consolidated plan derived from studying the open-source leaders and from honest
same-pin measurement of Brainy. **Doctrine: improve the product, never tailor the
benchmark.** LOCOMO/OpMem are diagnostics, not targets.

## Where Brainy is (2026-07-23, `dev`)

- Conversational path: async provider extract, atomic episodes, event-time
  (`observed_at` incl. relative dates), content-bearing + hybrid ranking.
- Entity linking: generic extraction + persistence; graph propagation available,
  gated to real embedders.
- Embeddings: pluggable OpenAI-compatible provider with per-query similarity
  calibration (works with any model); local hash fallback for CI/offline.
  **Unblocked on CF AI Gateway** via
  `BRAINY_EMBEDDING_MODEL=workers-ai/@cf/baai/bge-base-en-v1.5` (768-d).
  Bare `@cf/...` models fail with `Invalid provider`; see staging runbook step 8.
- Own suites: OpMem 12/12, marketing vertical 16/16.
- LOCOMO smoke (1 conv / 30 Q, `gpt-oss-120b` judge): ~12–13/30 on hash/local.
  Re-measure with gateway dense embeddings + entity auto-on is in progress;
  a GPT-class judge is still required for a Mem0-comparable claim.

## What the leaders do (and Brainy's status)

| Technique | Source | Brainy |
| --- | --- | --- |
| Single-pass ADD-only extraction | Mem0 v3 | Done (provider extract + episodes) |
| Entity linking + entity-boosted retrieval | Mem0, Zep, A-MEM | Infra done; ranking gated to embedders |
| Multi-signal fusion (semantic + BM25 + entity) | Mem0 | Partial: token + calibrated dense; BM25/IDF not yet |
| Temporal metadata + query temporal-intent scoring | Mem0 | Partial: `observed_at` + when-query boosts |
| Bi-temporal fact invalidation / supersession | Zep/Graphiti | Backlog (ENG-86/ENG-59) |
| Entity graph + Personalized PageRank multi-hop | HippoRAG | Prototype (rerank-only, gated) |
| Note construction + memory evolution | A-MEM | Not started |

## Ordered path (highest leverage first)

1. **Strong hosted embeddings on staging.** Set `BRAINY_EMBEDDING_*` to CF AI
   Gateway Workers AI (`workers-ai/@cf/baai/bge-base-en-v1.5` — runbook step 8).
   Same-pin local remeasure (`gpt-oss-120b`): dense emb alone **13/30**
   (net-neutral vs hash); entity ranking still regresses (**11/30**) so it stays
   opt-in. Re-measure on staging with a **GPT-class judge** for a Mem0-comparable
   claim — the answerer/judge remains the gate.
2. **BM25/IDF lexical signal.** Replace match-count ratio with IDF-weighted term
   scoring (Mem0's keyword signal). Generic; A/B before default-on.
3. **Temporal supersession (ENG-86).** Mark contradicted facts superseded so
   retrieval prefers current state (knowledge-update correctness).
4. **Entity graph + PageRank (HippoRAG).** Convert the entity infra into real
   multi-hop propagation once dense embeddings feed it; prove non-regression.
5. **Full LOCOMO + LongMemEval** on staging, then publish with pins.

## Measurement discipline

- Same pins across product iterations; attribute deltas to product.
- Never special-case dataset speakers/answers in product or the eval answerer.
- Publish honest scores; a local score is not comparable across judges/budgets.
- Reproduce embedding-backed runs with `evals/tools/local_embeddings_server.py`.

## Why not "SOTA" yet, honestly

Mem0's 92 uses a managed platform, hosted embeddings, a top-200 retrieval budget,
and a GPT-class answerer/judge. Hosted embeddings are no longer blocked (CF
Workers AI via AI Gateway works with the `workers-ai/` model prefix). Remaining
gaps for a fair claim: wire the same model on Render staging, re-tune boosts
there, and judge with a GPT-class model — not the 30-question `gpt-oss-120b`
smoke alone.
