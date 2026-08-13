# Program execution status — competitive program (2026-08-11)

**External handoff:** [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)  
**Accepted competitive verdict:** [external-reviews/2026-08-11-competitive-architecture-verdict.md](./external-reviews/2026-08-11-competitive-architecture-verdict.md)  
**Competitive SOP / gap matrix:** [competitive/README.md](./competitive/README.md) · [competitive/competitive-gap-matrix.md](./competitive/competitive-gap-matrix.md)  
**Hardening self-review prompt (historical):** [external-reviews/2026-08-11-hardening-self-review-prompt.md](./external-reviews/2026-08-11-hardening-self-review-prompt.md)  
**Intake SOP:** [external-reviews/README.md](./external-reviews/README.md)  
**Tips:** `main` `308d3a1` (production = V3 hardening cutover) · `dev` `1f2f26f` (staging Render live)

## Course correction (accepted)

1. Architect PR1–PR7 remain **closed**.
2. Recall-contract + V3 hardening (#93–#98) **landed** on `dev` + `main`.
3. **New PoR (2026-08-11):** competitive architecture program — Mem0-quality conversational recall + Graphiti-quality relations + Brainy governed truth. Principle: **recall broadly → represent explicitly → prove narrowly → answer truthfully.**

## Hardening cycle — closed

| Item | Status |
| --- | --- |
| #93–#98 | **Merged** |
| Post-cutover OpMem / marketing | **13/13** / **passed** |
| Post-cutover LoCoMo 1×30 | **15/30** (MH 50%, OD 25%) — dip vs Gate 0 18/30 |
| Post-cutover LoCoMo 3×90 | **33/90** (MH **22.2%**, OD 42.9%) |
| LME-20 | **Publishable** 0/20 `/recall` (`jobs` 4829/4829 failed=0) — [pin](../benchmarks/artifacts/lme20-product-recall-pr1-20260812-pin.md) |

## Competitive program — accepted sequence

| PR | Failure class | Status | Notes |
| --- | --- | --- | --- |
| PR1 LME-20 measurement integrity | MEASUREMENT / WRITE_PIPELINE | **Done** | Publishable pin `lme20-product-recall-pr1-20260812`: 20/20 `/recall`, jobs 4829=4829 failed=0; accuracy **0/20** (honest, not a quality win) |
| PR2 Conversational vs governed write policy | REPRESENTATION_MISS | **Landed (code)** | Core/conversation append-only; non-core verticals keep #94 ops. Local OpMem **13/13** + marketing **passed**. No LoCoMo/LME quality pin yet. |
| PR3 Temporal features V1 | TEMPORAL_RETRIEVAL_MISS | Queued | Reuse mig-16; add `temporal_score` |
| PR4 Retrieval V4 budgets | RETRIEVAL_MISS | Queued | Extend fusion_v2; candidate matrix @ fixed tokens |
| PR5 Context vs proof split | EVIDENCE_COVERAGE_MISS | Queued | `ContextEvidence` + `ProofChain` |
| PR6 Canonical entity store V2 | ENTITY_RESOLUTION_MISS | Queued | mig 20+ |
| PR7 Relation memory V1 | MULTIHOP_REPRESENTATION_MISS | Queued | Postgres SPO |
| PR8 Hop executor V3 | MULTIHOP_PLANNING_MISS | Queued | Relation traversal invariant |
| PR9 Assistant-generated memories | EXTRACTION_COVERAGE_MISS | Queued | Stage metrics |
| PR10 Frozen competitive qualification | — | Blocked on PR1–PR9 | Fresh Mem0 same-pin + multi-seed + LME |

## Still open (honest)

- Conversational quality (LME-20 **0/20** publishable; LoCoMo post-cutover 15/30 / 33/90)  
- PR2 write-policy code landed; local OpMem 13/13 + marketing passed; remeasure LoCoMo/LME only after a dedicated pin (do not invent lift)  
- PR3–PR10 execution  
- Mem0 same-pin under post-harden stack  
- LME-100 (measurement gate open; quality not ready)  
- Pack authority / procedures / conflict packets (deferred behind conversational parity)  
- Hash/128 re-embed residue  

## Claims discipline

- Allowed: Gate 0 18/30 + 32/90; post-cutover 15/30 + 33/90 with honesty; OpMem/marketing non-reg; **publishable LME-20 0/20** (integrity, not quality); competitive program adopted with modifications.  
- Forbidden: unqualified “beats Mem0”; SOTA; “MH solved”; inventing LME accuracy; calling 0/20 a quality improvement; calling post-cutover 1×30 an improvement vs Gate 0; calling 3×90 MH 50%.
