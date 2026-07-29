# Master-plan execution status

**Date:** 2026-07-29  
**Program:** `docs/research/master-plan.md` (PR #69)

## Code / product deliverables — complete

| ID | Deliverable | Evidence |
| --- | --- | --- |
| W1 | De-overfit + CI denylist | `overfit_denylist_test.go`; P0 baseline 16/30 |
| W2 | Typed atom taxonomy + index | `predicates.go`, migration v13, golden fixtures |
| W3.2 | Postgres FTS | migration v14 `content_tsv` |
| W3.4 | Predicate enumeration admit | Search + `/recall` enumerate |
| W4 | `POST /recall` | `docs/conversation-recall.md`; live smoke |
| W5 | Auto-supersede state | `autoSupersedePriorState`; OpMem upd03 → **13/13** |
| W7 | Support pack | `packs/support/v1` fixtures **3/3** |
| Lane A | memory-benchmarks Brainy stub | `evals/public/backends/memory_benchmarks_brainy.py` |

## Ops-gated (not code-blocked)

These remain for publish claims / GA, requiring budget or external accounts:

1. Full LoCoMo 10-convo × ≥3 seeds on publish-stack judge (`run_full.py`)
2. LongMemEval-S / BEAM evaluation runs
3. AMB / MemoryBench public submission
4. W6 load SLOs at 10K memories/subject
5. Design partner live on support pack

## Staging snapshot

- OpMem: **13/13**
- Support vertical: **3/3**
- `/recall` enumerate: OK
- LOCOMO P0 baseline: **16/30** (honest post-de-overfit)
