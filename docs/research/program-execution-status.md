# Program execution status — competitive program (2026-08-13)

**External handoff:** [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)  
**Accepted competitive verdict:** [external-reviews/2026-08-11-competitive-architecture-verdict.md](./external-reviews/2026-08-11-competitive-architecture-verdict.md)  
**Competitive SOP / gap matrix:** [competitive/README.md](./competitive/README.md) · [competitive/competitive-gap-matrix.md](./competitive/competitive-gap-matrix.md)  
**Hardening self-review prompt (historical):** [external-reviews/2026-08-11-hardening-self-review-prompt.md](./external-reviews/2026-08-11-hardening-self-review-prompt.md)  
**Intake SOP:** [external-reviews/README.md](./external-reviews/README.md)  
**Tips:** `main` `308d3a1` (production = V3 hardening cutover) · `dev` `24be5ab` (PR #100 merged; staging = not production)

## Course correction (accepted)

1. Architect PR1–PR7 remain **closed**.
2. Recall-contract + V3 hardening (#93–#98) **landed** on `dev` + `main`.
3. **New PoR (2026-08-11):** competitive architecture program — Mem0-quality conversational recall + Graphiti-quality relations + Brainy governed truth. Principle: **recall broadly → represent explicitly → prove narrowly → answer truthfully.**
4. **2026-08-13:** PR #100 (PR1 LME-20 integrity pin + PR2 write policy) merged to **`dev` only**. `main` untouched.

## Hardening cycle — closed

| Item | Status |
| --- | --- |
| #93–#98 | **Merged** |
| Post-cutover OpMem / marketing | **13/13** / **passed** |
| Post-cutover LoCoMo 1×30 | **15/30** (MH 50%, OD 25%) — dip vs Gate 0 18/30 |
| Post-cutover LoCoMo 3×90 | **33/90** (MH **22.2%**, OD 42.9%) |
| LME-20 | **Publishable** 0/20 `/recall` (`jobs` 4829/4829 failed=0) — [pin](../benchmarks/artifacts/lme20-product-recall-pr1-20260812-pin.md) |

## Wave D — diagnosis (2026-08-13)

Mined the publishable LME-20 json (gitignored). **0/20 is not empty ingest.**

| Signal | Value |
| --- | --- |
| Mean search hits | **24.45** (range 6–30) |
| Gold in top-k (strict) | **6/20** |
| Empty retrieval | **0/20** |
| Primary labels | **EVIDENCE_COVERAGE_MISS 12** · **READER_MISS 6** · **ABSTENTION_MISS 2** |

Item ledger: [lme20-product-recall-pr1-20260812-failure-ledger.md](../benchmarks/artifacts/lme20-product-recall-pr1-20260812-failure-ledger.md)

**Decision gate** (this is the execution order; the competitive sequence is architecture, not a sufficient run order until this histogram):

| If | Then |
| --- | --- |
| Hops replace packet / reader sees hop-narrowed text; oracles can still look supported | **PR5 first** |
| Temporal / knowledge-update 0; `IncludeHistorical` not set from intent; auto-supersede hides priors | **PR3** (reuse mig-16; `temporal_score`) |
| ~240 extracts/subject; episode **+0.1 boost** promotes assistant pollution | **PR4** at fixed context tokens; downrank untyped episodes |
| Assistant turns stored as recall-primary `conversation_episode` | **PR9 indicated** (drop boilerplate episodes; keep assistant-stated facts) |
| LME single-session fails without missing edges | **Do not start PR6–PR8** as the first move |

LoCoMo 1×30 remasure on merged PR2 (`24be5ab`): run id `locomo-pr2-dev-1x30-20260813`. Pin when the json lands; do not invent lift vs 15/30.

## Competitive program — accepted sequence

| PR | Failure class | Status | Notes |
| --- | --- | --- | --- |
| PR1 LME-20 measurement integrity | MEASUREMENT / WRITE_PIPELINE | **Done** | Publishable pin: 20/20 `/recall`, jobs 4829=4829 failed=0; accuracy **0/20** |
| PR2 Conversational vs governed write policy | REPRESENTATION_MISS | **Merged to `dev`** (#100) | Core/conversation append-only; non-core verticals keep #94 ops. Local OpMem **13/13** + marketing **passed**. Auto-supersede still runs on conversational ingest. |
| PR3 Temporal features V1 | TEMPORAL_RETRIEVAL_MISS | **Wave D: after PR5** | Reuse mig-16; intent → `IncludeHistorical`; `temporal_score` |
| PR4 Retrieval V4 budgets | RETRIEVAL_MISS | Queued after PR3 | Wire `MaxEvidenceTokens`; candidate matrix 30/50/100/200 @ fixed tokens; episode penalty |
| PR5 Context vs proof split | EVIDENCE_COVERAGE_MISS | **Wave D: first** | `ContextEvidence` + `ProofChain`; stop hop packet replacement |
| PR6 Canonical entity store V2 | ENTITY_RESOLUTION_MISS | **Deferred** | Not the LME driver (single-session 0 without missing edges) |
| PR7 Relation memory V1 | MULTIHOP_REPRESENTATION_MISS | **Deferred** | Postgres SPO only after Wave 1 shows MH coverage as the remaining driver |
| PR8 Hop executor V3 | MULTIHOP_PLANNING_MISS | **Deferred** | `follow_relation` after PR7 |
| PR9 Assistant-generated memories | EXTRACTION_COVERAGE_MISS | **Indicated (inverted)** | Do not persist assistant boilerplate as recall-primary episodes |
| PR10 Frozen competitive qualification | — | Blocked on Wave 1 pins | Fresh Mem0 same-pin + publishable LME-20 quality; **no SOTA / beats-Mem0 until that pin** |

## Still open (honest)

- Conversational quality (LME-20 **0/20** publishable; LoCoMo post-cutover 15/30 / 33/90)
- Wave D LoCoMo 1×30 remasure in flight (`locomo-pr2-dev-1x30-20260813`)
- PR3–PR5–PR4–PR9 follow-up branches off `dev` `24be5ab`
- PR6–PR8 deferred pending Wave 1 MH coverage evidence
- Mem0 same-pin under post-harden / PR2 stack (old V3 pin is not this stack)
- LME-100 (measurement gate open; quality not ready)
- Pack authority / procedures / conflict packets (deferred behind conversational parity)
- Hash/128 re-embed residue

## Claims discipline

- Allowed: Gate 0 18/30 + 32/90; post-cutover 15/30 + 33/90 with honesty; OpMem/marketing non-reg; **publishable LME-20 0/20** (integrity, not quality); #100 merged to `dev` only; Wave D histogram above.
- Forbidden: unqualified “beats Mem0”; SOTA; “MH solved”; inventing LME accuracy; calling 0/20 a quality improvement; calling post-cutover 1×30 an improvement vs Gate 0; calling 3×90 MH 50%; promising 75% LoCoMo/LME.
