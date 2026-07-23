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
embeddings** feeding the graph. Hosted gateway embeddings are unblocked as of
2026-07-23 via `workers-ai/@cf/baai/bge-base-en-v1.5` on CF AI Gateway
(`/compat/embeddings`); earlier 403/`Invalid provider` failures were model-id
format, not a Brainy bug.

Decision (product-first, anti-benchmax): keep entity **extraction + persistence**
(provenance + future graph layer) always on. Entity **ranking** stays **OFF by
default** — including when a provider embedding model is configured — until a
staging re-tune shows lift under a comparable judge. Opt in with
`BRAINY_ENTITY_RANKING=true`.

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

## Update: embedding model size (bge-base 768-d, 2026-07-20)

| Config | Overall |
| --- | ---: |
| real emb bge-small (384-d), entity OFF | 12/30 |
| real emb bge-base (768-d), entity OFF | 12/30 |

A larger local embedding model does **not** move this smoke: retrieval already
surfaces the ground-truth spans (retrieval-miss is low); the remaining failures
are answer/synthesis and temporal precision under the pinned `gpt-oss-120b`
answerer/judge — not embedding recall. Scores are **not** comparable to Mem0's
published 92 (different judge/answerer, retrieval budget, managed platform).

Guidance: on staging, point `BRAINY_EMBEDDING_*` at a strong hosted embeddings
endpoint and re-measure with a comparable answerer/judge before drawing SOTA
conclusions. Reproduce embedding-backed runs locally via
`evals/tools/local_embeddings_server.py`.

## Update: CF Workers AI dense embeddings (2026-07-23)

Gateway unlock: `BRAINY_EMBEDDING_MODEL=workers-ai/@cf/baai/bge-base-en-v1.5`
(768-d) on CF AI Gateway `/compat`. Same pins as prior smokes (`gpt-oss-120b`
answerer/judge, async ingest, 1 conv / 30 Q). Eval harness now waits for the
async extract queue to settle before QA.

| Config | Overall | multi-hop | temporal | open-domain |
| --- | ---: | ---: | ---: | ---: |
| hash baseline (entity gated off) | **13/30** | 2/10 | 8/16 | 3/4 |
| dense emb + entity auto-ON (drained) | 11/30 | 2/10 | 8/16 | 1/4 |
| dense emb + entity forced OFF | **13/30** | 2/10 | 9/16 | 2/4 |

Takeaways:
- Hosted embeddings are unblocked; Render staging still needs the three
  `BRAINY_EMBEDDING_*` secrets (runbook step 8).
- Under the **same** `gpt-oss-120b` judge, dense emb alone is **net-neutral**
  (13/30); entity ranking still **regresses** (11/30). Premature 14/30 (QA
  before queue drain) is not a pin.
- Product default: entity ranking OFF even with `BRAINY_EMBEDDING_MODEL` set.
  Re-enable only after staging re-tune + GPT-class judge shows lift.
- Anti-benchmax: no boost re-tuning against this 30-Q smoke.

Artifacts: `docs/benchmarks/runs/locomo-smoke-cf-bge-base-drained.*`,
`docs/benchmarks/runs/locomo-smoke-cf-bge-base-entity-off.*`

## Update: IDF-weighted lexical coverage (2026-07-20)

Implemented BM25-style IDF weighting (distinctive query terms dominate). Same-pin
smoke (hash embedder, entity off): **10/30** vs plain-coverage **13/30**; OpMem 12/12.
The additive boost stack (exact-span, date, kind) was tuned to count-ratio, so
IDF over-rewards lone rare-term matches without a full re-tune. Gated off
(`BRAINY_IDF_RANKING`) pending a staging re-tune. Not shipped as default — no
regression, no benchmax re-tuning to the 30-Q smoke.

## Update: staging CF dense embeddings (2026-07-23)

Wired on Render (`brainy-api-staging` + `brainy-worker-staging`):

- `BRAINY_EMBEDDING_BASE_URL` = CF AI Gateway `/compat`
- `BRAINY_EMBEDDING_MODEL` = `workers-ai/@cf/baai/bge-base-en-v1.5` (768-d)
- `BRAINY_ENTITY_RANKING=false`

Same-pin LOCOMO smoke on staging: **13/30** (matches hash baseline). OpMem **12/12**.
Search explain showed `embedding_similarity` / `ranking_basis=hybrid` — dense path live.
