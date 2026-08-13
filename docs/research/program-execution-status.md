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

### PR6–PR8 deferral (measured)

Wave 1 does **not** start with canonical entities / Postgres SPO / hop V3:

- LME-20 single-session-user is 0/3 with search hits present.
- LoCoMo 1×30 WRONG items are **24/24 `READER_MISS`** with coverage oracles supported.
- Mem0 still leads MH on the same pin (7/10 vs 4/10), but Brainy’s oracles say the facts are already in the packet. That is PR5/PR4/PR9, not a graph DB.

Revisit PR6–PR8 only after Wave 1 remasure shows **coverage** (not reader-parse) as the remaining MH driver on 3×90.

LoCoMo 1×30 remasure on merged PR2 (`24be5ab`): **6/30** (MH 4/10, temporal 1/16). Ledger **24/24 READER_MISS** with oracles supported. Pin: [locomo-pr2-dev-1x30-20260813.md](../benchmarks/artifacts/locomo-pr2-dev-1x30-20260813.md). Dip vs staging 15/30; do not invent lift.

## Competitive program — accepted sequence

| PR | Failure class | Status | Notes |
| --- | --- | --- | --- |
| PR1 LME-20 measurement integrity | MEASUREMENT / WRITE_PIPELINE | **Done** | Publishable pin: 20/20 `/recall`, jobs 4829=4829 failed=0; accuracy **0/20** |
| PR2 Conversational vs governed write policy | REPRESENTATION_MISS | **Merged to `dev`** (#100) | Core/conversation append-only; non-core verticals keep #94 ops. Local OpMem **13/13** + marketing **passed**. Auto-supersede still runs on conversational ingest. |
| PR3 Temporal features V1 | TEMPORAL_RETRIEVAL_MISS | **Opened** #103 | Reuse mig-16; intent → `IncludeHistorical`; `temporal_score` |
| PR4 Retrieval V4 budgets | RETRIEVAL_MISS | **Opened** #104 | Wire `MaxEvidenceTokens`; candidate matrix 30/50/100/200 @ fixed tokens; episode penalty |
| PR5 Context vs proof split | EVIDENCE_COVERAGE_MISS | **Opened** #102 | `ContextEvidence` + `ProofChain`; stop hop packet replacement |
| PR6 Canonical entity store V2 | ENTITY_RESOLUTION_MISS | **Deferred** | Wave D + 1×30: oracles supported / READER_MISS. Not missing edges. |
| PR7 Relation memory V1 | MULTIHOP_REPRESENTATION_MISS | **Deferred** | 1×30 MH 4/10 vs Mem0 7/10 but 24/24 failures are READER_MISS with coverage supported — not an SPO-table first move |
| PR8 Hop executor V3 | MULTIHOP_PLANNING_MISS | **Deferred** | `follow_relation` after PR7, only if Wave 1 pins show coverage (not reader) as the MH driver |
| PR9 Assistant-generated memories | EXTRACTION_COVERAGE_MISS | **Opened** #105 | Do not persist assistant boilerplate as recall-primary episodes |
| PR10 Frozen competitive qualification | — | **Measured; not claimed** | Fresh Mem0 **12/30** (MH 70%) vs Brainy local **6/30**. LME-20 quality still **0/20**. **No SOTA / beats-Mem0.** |

## Still open (honest)

- Conversational quality (LME-20 **0/20** publishable; LoCoMo staging 15/30 / 33/90; **local PR2 remasure 6/30**)
- Merge Wave 1 PRs #102–#105 to `dev` and remasure (do not invent lift beforehand)
- PR6–PR8 still deferred: 1×30 failures are READER_MISS with oracles supported
- PR10: Mem0 **12/30** same-pin done; Brainy does **not** win; LME-20 quality still 0/20 so LME-100 is not a quality run
- Pack authority / procedures / conflict packets (deferred behind conversational parity)
- Hash/128 re-embed residue

## Claims discipline

- Allowed: Gate 0 18/30 + 32/90; post-cutover 15/30 + 33/90 with honesty; local PR2 remasure **6/30**; fresh Mem0 **12/30** (MH 7/10); OpMem/marketing non-reg; **publishable LME-20 0/20** (integrity, not quality); #100 merged to `dev` only; Wave D histogram; Wave 1 PRs opened against `dev`.
- Forbidden: unqualified “beats Mem0”; SOTA; “MH solved”; inventing LME accuracy; calling 0/20 a quality improvement; calling post-cutover 1×30 or local 6/30 an improvement vs Gate 0; calling 3×90 MH 50%; promising 75% LoCoMo/LME.
