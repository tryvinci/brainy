# Multi-hop packet/proof remasure — 2026-08-20

Product `/recall` only. Fail-closed integrity tenant `integrity-s0-1`
(skip-ingest). Same stratified 180 seed-1 MH slice (**n=33**). Not a 1×30
freeze, not n=1540, not Mem0 same-pin, not SOTA.

| Field | Value |
| --- | --- |
| Tenant | `integrity-s0-1` (skip-ingest; same provider ingest as P4 / S0) |
| Dataset SHA | `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` |
| Sample | stratified 180 seed 1, **multi-hop 33** |
| API | local integrity stack, port 18100 |
| ANN | pgvector + 768-d rows only (22,481), signatures match, embed/extract fallbacks 0 |
| Extractor (ingest) | hosted gpt-oss-120b via AI Gateway (unchanged store) |
| Embedder | hosted 768-d BGE (model id not recorded) |
| Judge | hosted OpenAI-compatible gateway, temp 0.0 |
| Oracle probes | **not run** on this slice (full S0 ledger stalled on embed timeouts) |

## Scores

| Lane | Multi-hop | Notes |
| --- | ---: | --- |
| Prior fail-closed S0 product `/recall` | **1/33** | [S0 pin](./locomo-integrity-s0-20260819.md) |
| This remasure | **2/33 (0.061)** | Same tenant, new packet/proof SHA |

CORRECT this run:

- `conv-42-q56` “What animal do both Nate and Joanna like?” gold `Turtles.` → `Turtles, Dairy-free Desserts`
- `conv-49-q15` “What kind of unhealthy snacks does Sam enjoy eating?” gold `soda, candy` → long preference list that includes `Soda And Candy`

The first item is the attributed proof win (hops were resolving the topic noun
`animal` and answering from search-fallback activity lists while “Joanna likes
turtles” sat in hop contents). The second is a judge hit on a crowded
preference list, not a new join.

## What we did not remasure

- Full product S0 n=180 (started twice, `--fail-closed --skip-ingest`; stalled
  after conv-26/30 when `/recall` embed calls hit 120s timeouts). Do not treat
  the partial ledger as a pin.
- Industry search+harness, 3×90, 1×30, LME-20, n=1540.
- OpenAI embedding A/B (already pinned 2026-08-20; not re-run).

Compact judgments: [locomo-mh-packet-proof-20260820.json](./locomo-mh-packet-proof-20260820.json).
