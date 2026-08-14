# Program execution status — competitive program (2026-08-13)

**External handoff:** [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)  
**Accepted competitive verdict:** [external-reviews/2026-08-11-competitive-architecture-verdict.md](./external-reviews/2026-08-11-competitive-architecture-verdict.md)  
**Competitive SOP / gap matrix:** [competitive/README.md](./competitive/README.md) · [competitive/competitive-gap-matrix.md](./competitive/competitive-gap-matrix.md)  
**Hardening self-review prompt (historical):** [external-reviews/2026-08-11-hardening-self-review-prompt.md](./external-reviews/2026-08-11-hardening-self-review-prompt.md)  
**Intake SOP:** [external-reviews/README.md](./external-reviews/README.md)  
**Tips:** `main` = `dev` after 2026-08-14 production FF (compiler-quality gate; LoCoMo **11/30**). Next is held-out **coverage**, not more ranking. Every cycle: [competitive/cycle-closeout.md](./competitive/cycle-closeout.md).

**Course-correction (2026-08-14):** [sota-representation-path.md](./sota-representation-path.md) — Wave 1 was ranking around a transcript index. Next is **representation-first**: compile interactions into facts/entities/relations, retrieve those, keep episodes as provenance + fallback. Not reader-first, not retrieval-tuning-first.  
**External amendment (same day):** [external-reviews/2026-08-14-representation-path-additions.md](./external-reviews/2026-08-14-representation-path-additions.md) — R1c is fact-priority with episode fallback on incomplete coverage, **not** hard episode drop before R1b. R0 coverage oracle before the next LoCoMo category read.

## Course correction (accepted)

1. Architect PR1–PR7 remain **closed**.
2. Recall-contract + V3 hardening (#93–#98) **landed** on `dev` + `main`.
3. **New PoR (2026-08-11):** competitive architecture program — Mem0-quality conversational recall + Graphiti-quality relations + Brainy governed truth. Principle: **recall broadly → represent explicitly → prove narrowly → answer truthfully.**
4. **2026-08-13:** PR #100 (PR1 LME-20 integrity pin + PR2 write policy) merged to **`dev` only**. `main` untouched.  
5. **2026-08-13 later:** Wave 1 (#101–#105) merge-committed onto `dev` (`a7a5184`).  
6. **2026-08-14:** `dev` fast-forwarded onto **`main`** (`bd987fa`) with explicit approval.  
7. **2026-08-14:** Wave 1 deferred PR6–PR8 on a misleading `READER_MISS` oracle. Course-correct to [sota-representation-path.md](./sota-representation-path.md).  
8. **2026-08-14 later:** External review accepted. Sequence is **R0 → R1a → R1b → R1c → R2–R5 → R6**. Do **not** ship unconditional episode suppression before held-out compiler coverage.
9. **2026-08-14 later:** PR #113 merge-committed onto `dev` + `main` (`21a632b`): R0 fact-aware oracles, R1a primitive semantics, coverage-gated R1c. Local remasure: OpMem **13/13**, marketing **17/17**, LoCoMo 1×30 **10/30** (MH **2/10**, OD **0/4**, temporal **8/16**) — [pin](../benchmarks/artifacts/locomo-r1c-dev-1x30-20260814.md). Dip vs Wave 1 14/30. Next real milestone is **R1b** (atomic compiler + held-out coverage). The 19 `SOURCE_MISS` labels on that ledger were an evidence-dump over-label, not a write outage.
10. **2026-08-14 later:** Compiler quality gate (malformed `has done going at` / failed gerund stems are not recall-primary). Local remasure **11/30** (temporal **9/16**, q10 recovered). Packet junk templates 45→6. Ledger **15 WRITE_MISS + 4 READER_MISS**. [pin](../benchmarks/artifacts/locomo-atomq-dev-1x30-20260814.md).
11. **2026-08-14 later:** `dev` fast-forwarded onto **`main`** with explicit approval (compiler-quality + cycle-closeout SOP). Competitor compare is required every cycle: [competitive/cycle-closeout.md](./competitive/cycle-closeout.md).
12. **2026-08-14 later:** R1b coverage + R3 relation projection. Held-out audit green. Local remasure LoCoMo **15/30 (50.0%)**, MH **6/10 (60.0%)**, OD **0/4**, temporal **9/16 (56.2%)** vs Mem0 same-pin **12/30 (40.0%) / 7/10 (70.0%) / 3/4 (75.0%) / 2/16 (12.5%)**. OpMem **13/13**, marketing **17/17**. Lead overall this pin; trail MH by 1. **Not SOTA.** Next: R4 ID hops. [pin](../benchmarks/artifacts/locomo-r1b-dev-1x30-20260814.md).

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

Wave 1 remasure (`a7a5184`) still showed **coverage supported / READER_MISS**. That oracle meant gold sat in a **transcript blob**, not that entities/edges were unnecessary. **2026-08-14:** un-defer PR6–PR8 **after** the representation compiler (R1b) and coverage-gated fact-primary recall (R1c). R0 must stop calling that pattern `READER_MISS`. See [sota-representation-path.md](./sota-representation-path.md).

LoCoMo 1×30 remasure on merged PR2 (`24be5ab`): **6/30** (MH 4/10, temporal 1/16). Ledger **24/24 READER_MISS** with oracles supported. Pin: [locomo-pr2-dev-1x30-20260813.md](../benchmarks/artifacts/locomo-pr2-dev-1x30-20260813.md). Dip vs staging 15/30; do not invent lift.

LoCoMo 1×30 remasure on merged Wave 1 (`a7a5184`): **14/30** (MH **3/10**, OD 2/4, temporal **9/16**). Ledger **16/16 READER_MISS** with all four oracles **supported**. Pin: [locomo-wave1-dev-1x30-20260813.md](../benchmarks/artifacts/locomo-wave1-dev-1x30-20260813.md). Local OpMem **13/13** + marketing **passed**. **Confound:** this run had API `BRAINY_RECALL_LLM=1`; the 6/30 PR2 local run did not. Temporal moved; MH did not. Not an improvement vs Gate 0 18/30.

## Competitive program — accepted sequence

| PR | Failure class | Status | Notes |
| --- | --- | --- | --- |
| PR1 LME-20 measurement integrity | MEASUREMENT / WRITE_PIPELINE | **Done** | Publishable pin: 20/20 `/recall`, jobs 4829=4829 failed=0; accuracy **0/20** |
| PR2 Conversational vs governed write policy | REPRESENTATION_MISS | **Merged to `dev`** (#100) | Core/conversation append-only; non-core verticals keep #94 ops. Local OpMem **13/13** + marketing **passed**. Auto-supersede still runs on conversational ingest. |
| PR3 Temporal features V1 | TEMPORAL_RETRIEVAL_MISS | **Merged to `dev`** (#103) | Intent → `IncludeHistorical`; `temporal_score`. Local 1×30 temporal **9/16** (was 1/16 on PR2 local; hybrid-reader confound). |
| PR4 Retrieval V4 budgets | RETRIEVAL_MISS | **Merged to `dev`** (#104) | `MaxEvidenceTokens`; candidate pools 30/50/100/200 cap 200; episode −0.15 |
| PR5 Context vs proof split | EVIDENCE_COVERAGE_MISS | **Merged to `dev`** (#102) | `ContextEvidence` + `ProofChain`. Failures still READER_MISS with coverage supported. |
| PR6 Canonical entity store V2 | ENTITY_LINK_MISS | **R2 first slice** | subject/value on atoms + entity links. Canonical IDs still next |
| PR7 Relation memory V1 | RELATION_MISS | **R3 first slice** | `memory_relations` projection; MH 6/10 vs Mem0 7/10 |
| PR8 Hop executor V3 | PROOF_MISS / PLANNING_MISS | **R4 next** | q11 origin is READER_MISS with gold in facts |
| PR9 Assistant-generated memories | EXTRACTION_COVERAGE_MISS | **Merged to `dev`** (#105) | Skip phatic assistant `conversation_episode`; keep `assistant_stated` |
| PR10 Frozen competitive qualification | — | **Measured; not claimed** | Fresh Mem0 **12/30** (MH 70%) vs Wave 1 local **14/30** (MH **30%**). Mem0 still leads MH. LME-20 quality still **0/20**. **No SOTA / beats-Mem0.** |

## Still open (honest)

- Conversational quality (LME-20 **0/20** publishable; LoCoMo staging 15/30 / 33/90; local PR2 **6/30**; Wave 1 local **14/30**, MH **3/10**; R1c local **10/30**; compiler-quality local **11/30**, MH **2/10**; **R1b local 15/30 (50.0%)**, MH **6/10 (60.0%)** — lead Mem0 same-pin overall, trail MH 6 vs 7)
- Reader-only as the SOTA bet is **rejected**. Remaining conversational gap is **representation** (compiler coverage, then entities/edges). Hard episode-drop before R1b is also **rejected**. See [sota-representation-path.md](./sota-representation-path.md).
- Optional cleaner compare: staging 1×30 vs post-cutover 15/30 after `main` deploy
- PR10: Mem0 **12/30** same-pin; Wave 1 local **14/30** is not a qualification (MH 3/10 vs 7/10)
- Pack authority / procedures / conflict packets (deferred behind conversational parity)
- Hash/128 re-embed residue

## Claims discipline

- Allowed: Gate 0 18/30 + 32/90; post-cutover 15/30 + 33/90 with honesty; local PR2 remasure **6/30**; Wave 1 local **14/30** (MH 3/10, temporal 9/16) with hybrid-reader confound vs 6/30; R1c local **10/30** (MH 2/10, OD 0/4, temporal 8/16) as an honest dip after PR #113; compiler-quality local **11/30** (q10 recovered; junk 45→6; still a dip vs Wave 1); R1b local **15/30 (50.0%)** (MH 6/10, OD 0/4, temporal 9/16) as **lead this Mem0 freeze overall** and **trail MH 6 vs 7** — not SOTA / not qualification; fresh Mem0 **12/30 (40.0%)** (MH 7/10); OpMem/marketing non-reg; **publishable LME-20 0/20** (integrity, not quality); PR #113 on `dev`+`main` (`21a632b`); Wave D histogram.
- Forbidden: unqualified “beats Mem0”; SOTA; “MH solved”; inventing LME accuracy; calling 0/20 a quality improvement; calling post-cutover 1×30, local 6/30, Wave 1 14/30, R1c 10/30, compiler-quality 11/30, or R1b 15/30 an improvement vs Gate 0; calling 3×90 MH 50%; promising 75% LoCoMo/LME; treating Wave 1 ranking PRs as the SOTA architecture; treating R1c’s 19 `SOURCE_MISS` ledger labels as a write-pipeline outage.
