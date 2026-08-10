# Recall Contract V3 qualification status — 2026-08-10

**Branch / commit:** `pr/recall-contract-v3-a6c7` @ `4909832` (+ hybrid JSON salvage follow-up)  
**PR:** https://github.com/tryvinci/brainy/pull/92

## What qualified

| Pin | Result | Artifact |
| --- | --- | --- |
| LoCoMo 1×30 early | **16/30 (53.3%)**, MH 50%, hybrid **17/30** firing | [locomo-v3-early-pin-20260810.md](./locomo-v3-early-pin-20260810.md) |
| OpMem | **13/13** | [opmem-v3-nonreg-20260810.md](./opmem-v3-nonreg-20260810.md) |
| Marketing | **passed** | [marketing-v3-nonreg-20260810.md](./marketing-v3-nonreg-20260810.md) |
| Hybrid MH smoke | `reader_source=hybrid_llm_packet` | local fixture |

## In flight / deferred (honest)

| Pin | Status | Notes |
| --- | --- | --- |
| LoCoMo 3×90 | **running** | `locomo-v3-multiconvo-20260810` |
| LME-20 | **running** | async + job barrier; not publishable until complete |
| LME-100 | **deferred** | needs empty queue + capacity after LME-20 |
| Mem0 same-pin | **deferred** | re-run after 3×90 completes; prior Mem0 MH 70% still the gap marker |

## Claims discipline

- Do **not** claim beats Mem0 / conversational SOTA.
- Do **not** publish LME from partial runs.
- Mark hybrid reader **Landed + early-qualified (local)**; staging/production re-pin after merge to `dev`/`main`.
