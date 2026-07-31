# Master-plan execution status

**Date:** 2026-07-31  
**Program:** `docs/research/master-plan.md` (PR #69)

## Code / product deliverables — complete

| ID | Deliverable | Evidence |
| --- | --- | --- |
| W1–W5, W7 | Prior | `master-plan.md` |
| E2 harness | LongMemEval-S | `evals/public/longmemeval/run.py` |
| E3 harness | BEAM sample | `evals/public/beam/run.py` |
| W6 tool | Latency load | `evals/tools/latency_load.py` |
| Staging ops | Worker boot + respawn | FTS GIN deferred; `scripts/worker-respawn.sh` |

## Publish-stack evals

| Run | Status | Notes |
| --- | --- | --- |
| Full LoCoMo 10×1540 seed 0 | **Done** | **49.4% (761/1540)** — MH 25.2%, temp 54.8%, open 38.5%, single 56.7%. Gate R1 (≥75) **not met**. Report: `docs/benchmarks/artifacts/locomo-full-publish-s0-2a6a04.md` |
| LoCoMo seeds 1–2 | **Running** | `locomo-full-publish` ×2 for error bars |
| LongMemEval-S 100/500 | Pending | After LoCoMo multi-seed (serialize) |
| BEAM-100K sample | Pending | Dataset cached; after LME |
| W6 load SLO | **Measured** | p50 2403 / p95 4997 @ c=8 — **miss** |

## Honest read vs Gate R1

- Gate R1 asked LoCoMo ≥75 on publish stack. Seed 0 landed **~49** on CF gpt-oss judge after de-overfit — expected gap vs Mem0’s published 92.5 (different stack + years of platform tuning).
- Multi-hop remains the main hole (**25%**). Temporal/single-hop ~55–57%.
- Next product work after seed bars: extraction completeness + `/recall` enumerate for list/multi-hop — not more LOCOMO surface-forms.

## Ops notes

- Render worker still SIGTERM~60s; local drain workers used for async extract.
- Do not parallelize LME/BEAM with full LoCoMo (queue + WAF contention).
