# Embedding A/B — OpenAI arms (2026-08-20)

Addendum to [2026-08-19](embedding-ab-20260819.md). Same frozen tenant `integrity-s0-1`,
same stratified 180 (seed 1), same metric (gold-object recall@k / MRR from
`/memories/search?limit=200`, **not** QA accuracy).

**Blocker cleared:** `OPENAI_API_KEY` injected on the running agent. Re-embed used
direct OpenAI (`https://api.openai.com/v1`) with `dimensions=768` via
`cmd/reembed`. The API embedder was restarted to match each arm before scoring
(query vectors must match stored vectors).

**Stack note:** This VM rebuilt `integrity-s0-1` locally (22,481 memories / 768-d
rows; 6 failed extraction jobs on ingest). Prior pin cited 22,509 memories /
39,294 ANN rows on a long-lived stack — denominators differ slightly; compare
arms on this rebuild only.

| Arm | r@10 | r@30 | r@100 | r@200 | MRR | dense/q | lex/q | tok/q |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| BGE-768 (restored) | 0.306 | 0.378 | 0.506 | 0.528 | 0.190 | 4.5 | 116.9 | 1349 |
| OpenAI large @768 | 0.333 | 0.400 | 0.489 | 0.500 | 0.189 | 4.2 | 120.2 | 1403 |
| OpenAI small @768 | 0.333 | 0.400 | 0.517 | 0.544 | 0.197 | 4.1 | 117.1 | 1349 |
| hash-128 (2026-08-19) | 0.211 | 0.306 | 0.489 | 0.583 | 0.159 | 151.7 | 30.9 | 2011 |

OpenAI @768 beats BGE on this rebuild at r@10/r@30; small @768 is strongest at
r@100/r@200. Dense admission stays ~4 neighbors per query — not a license to
inflate top-k. Hash-128 r@200 remains higher because it cosine-scores the full
store, not because hash is a better embedder.

BGE restored after both OpenAI arms: 22,481 rows, all 768-d, signatures match,
ANN active, embed/extract fallbacks 0.

Item JSON:
[openai-large-768](embedding-ab-openai-large-768-20260820.json) ·
[openai-small-768](embedding-ab-openai-small-768-20260820.json).
