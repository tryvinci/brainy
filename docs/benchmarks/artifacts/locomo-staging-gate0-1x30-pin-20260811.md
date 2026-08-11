# LoCoMo 1×30 — Gate 0 staging pin — 2026-08-11

**Render deploy SHA:** `9bad8987483a` (API + worker live; verified via Render Deploys API)  
**`origin/dev` tip at Gate 0:** `9bad8987483abe11e241dcc13efac763d3450f43` — **match**  
**Harness note:** `brainy_commit` in the JSON manifest reflects the agent workspace tip at run time (`ae99f9e…`), not the Render image. Treat **Render deploy `9bad898`** as the system under test.

**Flags:** `BRAINY_USE_RECALL=1`, async ingest, top_k=30  
**Run id:** `locomo-staging-gate0-1x30-20260811`

## Scores

| Metric | Value |
| --- | ---: |
| Overall | **60.0% (18/30)** |
| multi-hop | **50% (5/10)** |
| temporal | **75% (12/16)** |
| open-domain | **25% (1/4)** |

Answer path: all 30 items via `brainy-recall` (product `/recall`).

## Compare (no tuning)

| Pin | Overall | MH | OD |
| --- | ---: | ---: | ---: |
| Local V3 early | 16/30 (53.3%) | 50% | 25% |
| **Staging Gate 0** | **18/30 (60%)** | 50% | 25% |

MH and OD remain the co-equal conversational gaps. No code/tuning from this pin.

Artifacts: [report](./locomo-staging-gate0-1x30-20260811.md) · run JSON under `docs/benchmarks/runs/`.
