# LoCoMo smoke — post planner/packs (2026-08-04)

**Run ID:** `locomo-smoke-c223da3d`  
**When:** 2026-08-04T22:40–22:44Z (≈4 min; not a multi-hour overnight job)  
**Brainy:** staging Render from `dev` commit `e8ecb82`  
**Ingest:** async (provider extract on worker)  
**Dataset:** locomo10.json SHA256 `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Scope:** 1 conversation / 30 scored questions (categories 1–4)

## Scores

| Metric | Value |
| --- | ---: |
| Overall | **60% (18/30)** |
| multi-hop | **40%** (4/10) |
| temporal | **68.8%** (11/16) |
| open-domain | **75%** (3/4) |
| Search p50 / p95 | ≈ **807 / 1509 ms** |

Vs prior pins: publish-stack full ≈49.8%; 2026-08-04 progress smoke 50% (15/30). This smoke is **directional**, not a 3-seed publish claim.

## Failure ledger

Path: `docs/benchmarks/artifacts/failure-ledger/locomo-smoke.jsonl` (12 WRONG rows).

| Finding | Detail |
| --- | --- |
| Primary taxonomy | **READER_MISS on all 12** |
| Stage oracles | evidence / semantic / retrieval / coverage all **supported** on every WRONG |
| WRONG by group | multi-hop 6 · temporal 5 · open-domain 1 |

**Interpretation:** For this smoke, the bottleneck moved past “nothing retrieved.” Evidence is present; synthesis/reader still misses identity slots, enumeration completeness, date precision, and multi-hop composition.

## Recommended next build (do not fusion-retune)

1. **Reader / synthesis over evidence packets** — make `/recall` (and harness answerer path preferring recall) consume `query_plan` + `evidence_packet` + temporal resolver; enforce enumeration coverage and abstain-vs-gold honesty.
2. Finish stratified **LME-100** adjudication (worker searchability / async timeouts observed mid-run).
3. Optional ops: hosted re-embed of remaining hash/128 residue; Mem0 same-pin compare (`MEM0_API_KEY` available).

Rejected unless new evidence: fusion constant retune, category dictionaries, graph DB, top-k inflation.
