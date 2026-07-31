# Master-plan execution status

**Date:** 2026-07-31  
**Program:** `docs/research/master-plan.md` (PR #69)

## Code / product deliverables — complete

| ID | Deliverable | Evidence |
| --- | --- | --- |
| W1 | De-overfit + CI denylist | `overfit_denylist_test.go`; P0 baseline 16/30 |
| W2 | Typed atom taxonomy + index | `predicates.go`, migration v13, golden fixtures |
| W3.2 | Postgres FTS | migration v14 `content_tsv` (GIN ensure gated) |
| W3.4 | Predicate enumeration admit | Search + `/recall` enumerate |
| W4 | `POST /recall` | `docs/conversation-recall.md`; live smoke |
| W5 | Auto-supersede state | `autoSupersedePriorState`; OpMem upd03 → **13/13** |
| W7 | Support pack | `packs/support/v1` fixtures **3/3** |
| Lane A | memory-benchmarks Brainy stub | `evals/public/backends/memory_benchmarks_brainy.py` |
| E2 harness | LongMemEval-S runner | `evals/public/longmemeval/run.py` |
| W6 tool | Latency load harness | `evals/tools/latency_load.py` |

## Publish-stack evals (budget unconstrained — in flight)

| Run | Status | Notes |
| --- | --- | --- |
| Full LoCoMo 10×1540 Q seed 0 | **Running** | `locomo-full-publish-s0-*`; CF gpt-oss judge |
| LoCoMo seeds 1–2 | Pending | After seed 0 |
| LongMemEval-S 100 then 500 | Harness ready | Start after LoCoMo seed 0 or in parallel slot |
| BEAM-100K sample | Pending | memory-benchmarks + Brainy client |
| W6 load SLO | **Measured** | p50/p95 miss under c=8; see `latency-slo.md` |
| AMB / MemoryBench submit | External | Needs account + PR |

## Ops notes

- **Render `brainy-worker-staging` exits ~60s after boot** (SIGTERM) even after FTS boot fix. Async queue drained via **local worker** pointed at Render external Postgres + same CF embed/provider pins.
- Migration 14 applied; GIN index exists; most historical rows still `content_tsv` NULL (trigger fills new writes).

## Staging snapshot

- OpMem: **13/13**
- Support vertical: **3/3**
- `/recall` enumerate: OK
- LOCOMO P0 baseline: **16/30** (honest post-de-overfit)
- W6 load: SLO miss at c=8 (p50 2403 / p95 4997)
