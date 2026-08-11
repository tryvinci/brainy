# LoCoMo 1×30 — post-cutover staging pin — 2026-08-11

**Render deploy SHA:** `1f2f26f28ecc` (API + worker live; verified via Render Deploys API)  
**Harness workspace tip:** `7cda664…` (docs branch) — treat **Render deploy `1f2f26f`** as the system under test.  
**Flags:** `BRAINY_USE_RECALL=1`, async ingest, top_k=30  
**Run id:** `locomo-staging-postcutover-1x30-20260811`

## Scores

| Metric | Value |
| --- | ---: |
| Overall | **50.0% (15/30)** |
| multi-hop | **50% (5/10)** |
| temporal | **56.2% (9/16)** |
| open-domain | **25% (1/4)** |

## Compare (no tuning)

| Pin | Overall | MH | OD |
| --- | ---: | ---: | ---: |
| Staging Gate 0 (`9bad898`) | 18/30 (60%) | 50% | 25% |
| Harden local (#93–#97) | 14/30 (46.7%) | 50% | 50% |
| **Staging post-cutover (`1f2f26f`)** | **15/30 (50%)** | **50%** | **25%** |

Honest: overall sits between Gate 0 and harden-local. Not an improvement over Gate 0. MH holds 50% on this 1×30 slice; OD still 25%.

Artifacts: [report](./locomo-staging-postcutover-1x30-20260811.md) · run JSON under `docs/benchmarks/runs/`.
