# LoCoMo 3×90 — post-cutover staging pin — 2026-08-11

**Render deploy SHA:** `1f2f26f28ecc` (API + worker live; verified via Render Deploys API)  
**Harness workspace tip (not deploy SHA):** recorded in run manifest  
**Flags:** `BRAINY_USE_RECALL=1`, async ingest, top_k=30  
**Run id:** `locomo-staging-postcutover-3x90-20260811`

## Scores

| Metric | Value |
| --- | ---: |
| Overall | **36.7% (33/90)** |
| multi-hop | **22.2% (8/36)** |
| temporal | **46.7% (21/45)** |
| open-domain | **42.9% (3/7)** |

## Compare (no tuning)

| Pin | Overall | MH | OD |
| --- | ---: | ---: | ---: |
| Staging Gate 0 (`9bad898`) | 32/90 (35.6%) | 19.4% | 42.9% |
| LoCoMo V3 local 3×90 | 31/90 (34.4%) | — | — |
| **Staging post-cutover (`1f2f26f`)** | **33/90 (36.7%)** | **22.2%** | **42.9%** |

Honest: overall roughly flat/slightly up vs Gate 0; MH remains weak on the multi-convo slice (~22%). Do not call this SOTA or “MH solved.”

Artifacts: [report](./locomo-staging-postcutover-3x90-20260811.md) · run JSON under `docs/benchmarks/runs/`.
