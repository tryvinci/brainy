# Embedding A/B — retrieval only (2026-08-19)

Gold-object recall@k / MRR from `/memories/search?limit=200`, **not** QA accuracy.
Frozen tenant `integrity-s0-1` (strict provider ingest, 22,509 memories, 768-d ANN).
Dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` stratified 180 seed 1.

ANN active, API/worker signatures match, embedder/extractor fallbacks 0.
OpenAI `text-embedding-3-large` skipped (no OpenAI key). Hash-128 control runs after S0 so the fail-closed BGE stack is not disturbed during QA.

| Arm | r@10 | r@30 | r@100 | r@200 | MRR | dense/q | lex/q | tok/q |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| BGE-768 | 0.239 | 0.311 | 0.417 | 0.444 | 0.148 | 4.4 | 107.1 | 1196 |
| hash-128 | _pending_ | | | | | | | |

Dense admission is a small slice of the hybrid candidate list (mean 4.4 dense vs 107 lexical at k=200). This is a retrieval metric, not a reason to inflate top-k.
