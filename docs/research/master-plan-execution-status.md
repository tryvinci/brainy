# Master-plan execution status

> **Superseding program of record (2026-08):** [sota-end-to-end-program.md](./sota-end-to-end-program.md) — see [program-execution-status.md](./program-execution-status.md) for Phase 0–5 landings.


**Date:** 2026-08-01  
**Program:** `docs/research/master-plan.md` (PR #69)

## Publish-stack portfolio — complete (honest baselines)

### Full LoCoMo (10 conv, 1540 Q, cats 1–4) — 3 seeds

| Seed | Overall | MH | Temp | Open | Single |
| --- | ---: | ---: | ---: | ---: | ---: |
| `2a6a04` | **49.4%** | 25.2% | 54.8% | 38.5% | 56.7% |
| `e7ba5b` | **49.4%** | 26.2% | 54.8% | 34.4% | 56.7% |
| `9b61f5` | **50.6%** | 27.7% | 55.8% | 38.5% | 57.8% |
| **Mean** | **≈49.8%** | **≈26%** | **≈55%** | **≈37%** | **≈57%** |

Gate R1 (LoCoMo ≥75): **not met**.

### LongMemEval-S (100 stratified)

| Metric | Value |
| --- | ---: |
| Overall | **4.0% (4/100)** |
| multi-session | 2/27 |
| temporal-reasoning | 1/26 |
| single-session-user | 1/14 |
| knowledge-update | 0/16 |
| single-session-assistant | 0/11 |
| single-session-preference | 0/6 |

Gate R1 (LME ≥75): **not met**. **Skip LME-500** until long-haystack retrieval improves.

### BEAM-100K (conversation 0, 20 Q)

| Metric | Value |
| --- | ---: |
| Overall | **40% (8/20)** |
| Strong | contradiction 2/2, summarization 2/2 |
| Weak | temporal 0/2, multi-session 0/2, event_ordering 0/2 |

### W6 latency

p50 **2403** / p95 **4997** ms @ concurrency 8 — SLO **miss**.

## Product takeaway

1. **Multi-hop / long-haystack** is the gap (LoCoMo MH ~26%, LME ~4%).
2. Do **not** reintroduce LOCOMO surface-forms.
3. Next product work: extraction completeness + `/recall` enumerate for list/multi-hop; long-context indexing for LME-scale haystacks.
4. Staging worker SIGTERM flap remains an ops issue (local drain workaround works).

## Artifacts

- `docs/benchmarks/artifacts/locomo-full-publish-s0-2a6a04.md`
- `docs/benchmarks/artifacts/locomo-full-s12-s0-e7ba5b.md`
- `docs/benchmarks/artifacts/locomo-full-s12-s1-9b61f5.md`
- `docs/benchmarks/artifacts/locomo-full-publish-summary.json`
- `docs/benchmarks/artifacts/beam-100k-c0-async.md`
- `docs/benchmarks/artifacts/lme-s-100.md`
- `docs/benchmarks/artifacts/latency-load-20260731T065251Z.json`
