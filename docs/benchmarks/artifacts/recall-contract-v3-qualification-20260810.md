# Recall Contract V3 qualification status — 2026-08-10

**Branch / commit:** `pr/recall-contract-v3-a6c7` @ `e3749de`  
**PR:** https://github.com/tryvinci/brainy/pull/92

## What qualified

| Pin | Result | Artifact |
| --- | --- | --- |
| LoCoMo 1×30 early | **16/30 (53.3%)**, MH 50%, hybrid **17/30** | [early pin](./locomo-v3-early-pin-20260810.md) |
| Mem0 same-pin 1×30 | **12/30 (40%)**, MH **70%** | [mem0](./locomo-mem0-samepin-v3-20260810.md) · [compare](./locomo-v3-samepin-vs-mem0-20260810.md) |
| LoCoMo 3×90 | **31/90 (34.4%)** (was 27/90) | [multiconvo pin](./locomo-v3-multiconvo-pin-20260810.md) |
| OpMem | **13/13** | [opmem](./opmem-v3-nonreg-20260810.md) |
| Marketing | **passed** | [marketing](./marketing-v3-nonreg-20260810.md) |

## Deferred (honest)

| Pin | Status | Notes |
| --- | --- | --- |
| LME-20 / LME-100 | **Deferred** | Queue contention; barrier+concurrency shipped — [note](./lme20-v3-deferred-20260810.md) |

## Claims discipline

- Brainy beats Mem0 **overall** on this 1×30 pin (16 vs 12); Mem0 still leads **MH** (70% vs 50%).
- Do **not** claim conversational SOTA.
- Do **not** publish LME from incomplete runs.
- Staging/production re-pin after merge to `dev`/`main`.
