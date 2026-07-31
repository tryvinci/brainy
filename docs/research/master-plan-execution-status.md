# Master-plan execution status

**Date:** 2026-07-31  
**Program:** `docs/research/master-plan.md` (PR #69)

## Code / product deliverables — complete

| ID | Deliverable | Evidence |
| --- | --- | --- |
| W1–W5, W7 | As prior | See previous rows / `master-plan.md` |
| E2 harness | LongMemEval-S runner | `evals/public/longmemeval/run.py` |
| E3 harness | BEAM sample runner | `evals/public/beam/run.py` (100K cached) |
| W6 tool | Latency load | `evals/tools/latency_load.py` |
| Staging ops | Worker boot + respawn | FTS GIN deferred; `scripts/worker-respawn.sh` |

## Publish-stack evals (budget unconstrained — in flight)

| Run | Status | Notes |
| --- | --- | --- |
| Full LoCoMo 10×1540 Q seed 0 | **Running** (~2h+) | `locomo-full-publish-s0-60d21f`; ingested conv-26/30/41; QA in progress. Async+provider extract via local drain workers. |
| LoCoMo seeds 1–2 | Pending | After seed 0 |
| LongMemEval-S smoke (5 Q) | Partial | Harness OK end-to-end; 3/5 judged WRONG on early smoke. Full 100/500 deferred until LoCoMo frees staging (provider extract ~1–6 jobs/min). |
| BEAM-100K conv 0 | Partial | Dataset cached (20 convs). Ingest hits intermittent WAF 403 on some turns (skip+retry shipped). Syncfer full QA until after LoCoMo. |
| W6 load SLO | **Measured** | p50 **2403** / p95 **4997** ms @ c=8 — SLO **miss** under load (`docs/benchmarks/artifacts/latency-load-20260731T065251Z.json`) |

## Ops notes

- Render `brainy-worker-staging` still SIGTERMs ~60s after boot; local drain workers on external Postgres keep the async queue alive.
- Provider extract on CF gpt-oss is the throughput bottleneck for large haystacks (LME/BEAM). Prefer sequential large runs, not parallel with full LoCoMo.
- Migration 14 applied; GIN present; most historical `content_tsv` still NULL.

## Staging snapshot

- OpMem: **13/13**
- Support vertical: **3/3**
- LOCOMO P0 baseline: **16/30**
- W6 load: SLO miss at c=8
- Full LoCoMo seed 0: **in progress**
