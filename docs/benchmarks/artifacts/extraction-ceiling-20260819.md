# Extraction ceiling — deterministic arm (2026-08-19)

Semantic gold coverage on the representation oracle, **not** QA accuracy.
Same LoCoMo 10-conv / stratified 180 (seed 1) as Aug-19 S0.
Dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`.

**Stack:** fail-closed integrity API `:18100`, pgvector active, 768-d hosted embedder.
Sync `/ingest` uses the deterministic extractor. Worker was idle for this arm.

| Arm | Tenant | Turns | Coverage |
| --- | --- | ---: | ---: |
| deterministic | `ceil-det-1` | 5882 | **139/180 (0.772)** |

## By group

| Group | Covered | Rate |
| --- | ---: | ---: |
| multi-hop | 24/33 | 0.727 |
| open-domain | 1/11 | 0.091 |
| single-hop | 82/98 | 0.837 |
| temporal | 32/38 | 0.842 |

Facts per question (representation blob): min 591, max 1285, mean 1026.

Uncovered n=41. Full item rows: [extraction-ceiling-det-20260819.json](extraction-ceiling-det-20260819.json).

Provider arm is scored after async ingest on tenant `integrity-s0-1` (frozen for P2/P5). Do not treat 0.772 as an extractor ceiling until that arm lands.
