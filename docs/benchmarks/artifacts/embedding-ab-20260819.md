# Embedding A/B — retrieval only (2026-08-19)

Gold-object recall@k / MRR from `/memories/search?limit=200`, **not** QA accuracy.
Frozen tenant `integrity-s0-1` (strict provider ingest, 22,509 memories).
Dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` stratified 180 seed 1.

OpenAI `text-embedding-3-large` / `-small` skipped (no OpenAI key). Hash-128 is
the local 128-d control after `cmd/reembed` (NULLs `embedding_vec_768`). BGE is
the fail-closed 768-d hosted arm with ANN.

| Arm | r@10 | r@30 | r@100 | r@200 | MRR | dense/q | lex/q | tok/q |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| BGE-768 (ANN) | 0.239 | 0.311 | 0.417 | 0.444 | 0.148 | 4.4 | 107.1 | 1196 |
| hash-128 (in-process cosine) | 0.211 | 0.306 | 0.489 | 0.583 | 0.159 | 151.7 | 30.9 | 2011 |

BGE ANN is active and signatures match on the 768 arm. Dense admission is a
small slice of hybrid at k=200 (mean 4.4 dense vs 107 lexical). Hash r@10 is
slightly worse; hash r@100/200 is higher because the control scores the full
store with cosine instead of admitting ~4 dense neighbors. This is a retrieval
metric, not a reason to inflate top-k, and not a claim that hash is a better
embedder.

Item JSON: [bge-768](embedding-ab-bge-768-20260819.json) · [hash-128](embedding-ab-hash-128-20260819.json).
