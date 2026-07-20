# Entity linking / graph propagation — honest A/B

SOTA memory systems (Mem0, Zep/Graphiti, A-MEM, HippoRAG) rely on entity linking
and graph propagation. We implemented both generically and measured same-pin
LOCOMO smoke (1 conversation / 30 Q, async ingest, identical judge/answerer).

| Config | Overall |
| --- | ---: |
| Entity ranking OFF (default) | **13/30 (43.3%)** |
| Entity overlap boost ON | 8/30 |
| Entity + IDF weighting ON | 10/30 |
| Entity graph propagation (rerank-only) ON | 10/30 |

## Conclusion

Naive/graph entity reranking **regresses** conversational recall in our stack:
distinctive-entity mentions are frequently not the answer, and reshuffling displaces
strong lexical/temporal hits. HippoRAG/Mem0 gains depend on **dense semantic
embeddings** feeding the graph — which is currently blocked in this environment
(provider `/embeddings` returns 403).

Decision (product-first, anti-benchmax): keep entity **extraction + persistence**
(provenance + future graph layer) always on. Entity **ranking** now auto-enables
**only when a provider embedding model is configured** (`BRAINY_EMBEDDING_MODEL`),
because the A/B shows it regresses with the local hash embedder but is the
documented SOTA path with real dense semantics. `BRAINY_ENTITY_RANKING=true|false`
overrides either way.

This is correct-by-construction rather than tuned to the 30-question smoke:
with a hash embedder it stays off (where it hurt), and with real embeddings it
turns on (where Mem0/Zep/HippoRAG show gains). Re-measure on staging once a
dense embedding endpoint is available.

Entities are not tuned to any benchmark — extraction is generic (proper nouns,
quoted spans, years).

## Update: real dense embeddings (local bge-small, 2026-07-20)

Unblocked dense embeddings by running a local open-source model
(`BAAI/bge-small-en-v1.5`, 384-d, CPU) behind an OpenAI-compatible endpoint,
since the hosted gateway blocks `/embeddings`.

**Critical finding:** modern embedding models have a high baseline cosine
(bge-small ≈ 0.49 for *unrelated* English). The prior absolute thresholds
(`>=0.15`) flooded candidates and compressed ranking. Fix: per-query similarity
**calibration** (rescale relative to the candidate mean) — now model-agnostic.

Same-pin LOCOMO smoke (1 convo / 30 Q):

| Config | Overall | multi-hop |
| --- | ---: | ---: |
| hash embedder (baseline) | 13/30 | 2/10 |
| real emb, uncalibrated | 11/30 | — |
| real emb, calibrated, entity OFF | 12/30 | 2/10 |
| real emb, calibrated, entity ON | 11/30 | 3/10 |

Takeaways:
- **Calibration is a required correctness fix** for any real embedding provider.
- A small CPU embedding model (bge-small) is ~on par with the tuned lexical+hash
  stack on this smoke; entity ranking helps multi-hop (+1) but is net-neutral.
- SOTA numbers (Mem0 92) use larger hosted embedding models + larger retrieval
  budgets; the path is a stronger embedding endpoint on staging, then re-measure.
- Repro: run an OpenAI-compatible embeddings server and set
  `BRAINY_EMBEDDING_BASE_URL` / `BRAINY_EMBEDDING_MODEL` (entity ranking auto-on).
