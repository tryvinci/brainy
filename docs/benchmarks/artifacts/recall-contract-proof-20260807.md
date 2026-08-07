# Recall-contract proof pins — 2026-08-07

**Commit:** `87c33e5` (+ harness fix for Mem0 `_tenant`)  
**Stack:** local API+worker, `BRAINY_USE_RECALL=1`, `BRAINY_RECALL_LLM=1`, provider extract + BGE embeddings  
**Judge/answerer:** same LLM pin for Brainy and Mem0  

## Same-pin LoCoMo smoke (1×30, conv-26)

| System | Overall | Temporal | Multi-hop | Open |
| --- | ---: | ---: | ---: | ---: |
| **Brainy** (`ef197919`) | **16/30 (53.3%)** | 10/16 (62.5%) | 4/10 (40%) | 2/4 |
| **Mem0 Platform** (`c2b2e5f7`) | 11/30 (36.7%) | 2/16 (12.5%) | **7/10 (70%)** | 2/4 |

**Interpretation**
- Fresh same-pin on post–recall-contract stack: **Brainy leads overall and temporal**.
- Mem0 still stronger on multi-hop — expected until packet bridge-chains + hybrid composition mature.
- Prior product `/recall` deterministic baseline was 43.3% (13/30); this pin is **+10pp** directional under hybrid recall (single smoke — not multi-seed SOTA).
- Do **not** compare either number to Mem0 blog ~92%.

Artifacts:
- `docs/benchmarks/runs/locomo-smoke-ef197919.*` (Brainy)
- `docs/benchmarks/runs/locomo-smoke-c2b2e5f7.*` (Mem0)
- Failure ledger: `docs/benchmarks/artifacts/failure-ledger/locomo-recall-contract-20260807.jsonl`

## LongMemEval-S (stratified 20)

Started locally; async job backlog large (~130–180 pending under one worker). First scored item: knowledge-update **WRONG**. Full 20-Q pin deferred until queue drains — do not cite partial as a gate.

Prior absolute LME-100 pin remains **4%** until re-run completes under the new contract.

## Reproduce

```bash
# Brainy (local or staging after deploy)
export BRAINY_BASE_URL=... BRAINY_USE_RECALL=1 BRAINY_RECALL_LLM=1
python3 -m public.locomo.run_smoke --conversations 1 --questions 30 --top-k 30

# Mem0 same pin
export MEM0_API_KEY=...
python3 -m public.locomo.run_smoke --system mem0 --conversations 1 --questions 30 --top-k 30
```

## Next measurement

1. Finish LME-20/100 after worker backlog clears (or scale workers).
2. Larger LoCoMo slice / 3-seed full after merge to `dev` staging.
3. Keep OpMem + marketing green as non-regression.
