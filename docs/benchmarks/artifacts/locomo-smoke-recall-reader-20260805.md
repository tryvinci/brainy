# LoCoMo smoke — product `/recall` reader closeout (2026-08-05)

**Run ID:** `locomo-smoke-f722342a`  
**When:** 2026-08-05T23:50:30Z  
**Brainy:** staging Render commit `3725489` (architect gap closeout on `dev`)  
**Ingest:** sync (`--sync-ingest`) — async backlog (~217 pending jobs) avoided for closeout  
**Answer path:** `BRAINY_USE_RECALL=1` → product `POST /recall` (`brainy-recall+answer` / `+enumerate`)  
**Dataset:** locomo10.json SHA256 `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Scope:** 1 conversation / 30 scored questions  

## Scores

| Metric | Value |
| --- | ---: |
| Overall | **43.3% (13/30)** |
| multi-hop | **50%** (5/10) |
| temporal | **43.8%** (7/16) |
| open-domain | **25%** (1/4) |
| Search p50 / p95 | ≈ **683 / 1108 ms** |

Vs prior planner/packs smoke (`c223da3d`, LLM answerer over search): 60% (18/30). Deterministic packet reader is **honest but weaker** on this pin — expected risk called in the closeout plan.

## Failure ledger

Path: `docs/benchmarks/artifacts/failure-ledger/locomo-recall-smoke.jsonl` (17 WRONG rows).

| Finding | Detail |
| --- | --- |
| Generation models | all `brainy-recall+*` (product path confirmed) |
| Primary taxonomy | still `READER_MISS` (17/17) — extractive packet compile ≠ judge gold |
| WRONG by group | temporal 9 · multi-hop 5 · open-domain 3 |

## Architect sequence status after this run

PR1–PR7 structural items are **landed** (async temporal guard, packet reader, FTS lexical honesty, pack sidecars+FSMs). Remaining conversational lift is **next-agent** work (better packet→answer composition / optional bounded LLM reader over packet IDs only), not re-opening the 2026-08-04 sequence.

Open beyond architect pass: pack authority/procedures/conflict packets; evidence-as-search-primary; LME-100/Mem0 same-pin proof; Phase-6 multi-seed gates.
