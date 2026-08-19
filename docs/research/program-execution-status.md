# Program execution status — competitive program (2026-08-17)

**External handoff:** [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)  
**This-pass verdict (live):** [external-reviews/2026-08-17-parity-gap-verdict.md](./external-reviews/2026-08-17-parity-gap-verdict.md)  
**This-pass source:** [external-reviews/2026-08-17-parity-gap-review.md](./external-reviews/2026-08-17-parity-gap-review.md)  
**Historical same-day (Wave 1):** [external-reviews/2026-08-17-competitive-archaeology-verdict.md](./external-reviews/2026-08-17-competitive-archaeology-verdict.md)  
**This-pass self-review prompt:** [external-reviews/2026-08-17-full-recall-self-review-prompt.md](./external-reviews/2026-08-17-full-recall-self-review-prompt.md)  
**Dip diagnosis:** [locomo-full-recall-dip-why-20260817.md](../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md)  
**Accepted competitive verdict:** [external-reviews/2026-08-11-competitive-architecture-verdict.md](./external-reviews/2026-08-11-competitive-architecture-verdict.md)  
**Competitive SOP / gap matrix:** [competitive/README.md](./competitive/README.md) · [competitive/competitive-gap-matrix.md](./competitive/competitive-gap-matrix.md)  
**Hardening self-review prompt (historical):** [external-reviews/2026-08-11-hardening-self-review-prompt.md](./external-reviews/2026-08-11-hardening-self-review-prompt.md)  
**Intake SOP:** [external-reviews/README.md](./external-reviews/README.md)  
**Tips:** `main` = `dev` after 2026-08-17 R5A production FF. Product SHA for LoCoMo pins is still **`1b5ab3e`**. Live: LoCoMo 1x30 **21/30** (MH **10/10**, OD **0/4**); full `/recall` **11.4%** (dip vs July search+harness 49.8%; that 49.8% is **not** a current-SHA ceiling). R0-R4 closed; **R5A structured-first `/recall` landed**; **R5B–R10 representation stack landed** (not a 70–80% claim). Next is freeze remasure when requested. Two lanes. Every cycle: [competitive/cycle-closeout.md](./competitive/cycle-closeout.md).

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
13. **2026-08-15:** R4 hops + leftover coverage. Local remasure LoCoMo **19/30 (63.3%)**, MH **9/10 (90.0%)**, OD **0/4**, temporal **10/16 (62.5%)** vs Mem0 same-pin **12/30 / 7/10 / 3/4 / 2/16**. OpMem **13/13**, marketing **17/17**. Lead overall **and** MH this pin. Remaining MH miss is image-gold. **Not SOTA.** Next: R5 OD. [pin](../benchmarks/artifacts/locomo-mh-r4c-dev-1x30-20260815.md).
14. **2026-08-15/16 remasure + 2026-08-17 dip diagnosis:** No product change (`1b5ab3e`). OpMem **13/13** vs Mem0 **10/13**; marketing **17/17** vs Mem0 **4/17**; LoCoMo 1x30 **21/30** vs Mem0 **11/30** (MH **10/10**, OD **0/4**, temporal **11/16**) — 1x30 did **not** drop vs R4h 20/30. Full LoCoMo **175/1540 (11.4%)** product `/recall` — **did** drop vs July search+harness **49.8%** because the path changed (`firstStatementFromPacket` slogans / enumerate / 188 abstains), not because the compiler vanished. LME-20 **4/20** `/recall` (lift vs 0/20). BEAM 100K **8/20**. LME-500 and BEAM 1M/10M **not run** (cost; would not change the diagnosis). Two stacked gaps recorded that day as answer-path then representation; item 16 **modifies** the 11.4% to ~50% numeric ceiling (not proven on current SHA). [why](../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md).
15. **2026-08-17 archaeology review adjudicated (historical for next-work):** Source pin `bd987fa` (Wave 1). **Keep** representation-first. **Reject** re-queue of P0-P7 (R0-R4 already landed). **Accept** two-lane eval. Histogram sum bug accepted and **fixed** on this branch. [historical verdict](./external-reviews/2026-08-17-competitive-archaeology-verdict.md).
16. **2026-08-17 current-SHA parity-gap review adjudicated (live):** Source pin docs `8492ad3` / product `1b5ab3e`. **Keep** representation-first. **Adjust** sequence to **R5A** structured-first `/recall` (retire `firstStatementFromPacket` as a normal factual strategy) then R5B typed packet, R6 compiler coverage V2, R7-R9 canonical identity / relation V2 / hop ID joins, R10 dual-path freeze. **Modify:** 11.4% to ~50% is directional, not a proven current-SHA ceiling; R5-on-OD is a diagnostic not the PR name; proposed v2 DDL is later, not next. Two lanes. Skip LME-500 / BEAM 1M. [verdict](./external-reviews/2026-08-17-parity-gap-verdict.md) · [source](./external-reviews/2026-08-17-parity-gap-review.md).
17. **2026-08-17 R5A landed:** Structured-first `/recall` (no first-packet slogans as normal answers). OpMem **13/13**, marketing **17/17** non-reg on this SHA. LoCoMo/LME/BEAM not re-run. Next: **R5B**. [closeout](./competitive/cycle-closeout.md).
18. **2026-08-18 R6a named-subject compiler:** Dialogue reports bind to the named person / addressee, not the reporter. Held-out audit is the merge gate. Not a n=1540 remasure and not a 70–80% claim. [path](./locomo-full-70-80-path.md) · [closeout](./competitive/cycle-closeout.md).
19. **2026-08-18 R5B–R10 representation stack:** Typed `ContextEvidence`, `she`/`he` last-named-person coref, canonical `ent:` IDs + `memory_entities` (mig v22), relation ID dual-write, hop ID joins (`typed_exact` vs context), dual-path freeze **wiring** (`--eval-lane`). OpMem/marketing non-reg required. LoCoMo n=1540 **not re-run**. Not SOTA. [freeze](./locomo-dual-path-freeze.md).
20. **2026-08-19 merge + plan:** PR #130 merge-committed onto `dev` (`fb3e166`); `dev` fast-forwarded onto **`main`** with explicit approval (2026-08-19). Execution now follows [sota-execution-plan.md](./sota-execution-plan.md): S0 stratified dual-lane baseline on this SHA first; full n=1540 only at S6 with a Mem0 Platform same-pin.
21. **2026-08-19 S0–S5 product increments:** Stratified dual-lane harness (`run_s0`, `--stratified`); compiler coverage audit (`scripts/compiler-audit.sh`, ≥85% held-out gate); entity-id enumerate; OD yes/no from compiled facts; alias lifecycle; hop-plan coverage; KU entity-scoped supersede; cross-session atoms; industry atoms-first + token reporting. **Not a LoCoMo remasure.** S0 ledger / S6 freeze still required before any competitive language.
22. **2026-08-19 S6 freeze wiring:** `python -m public.locomo.run_s6` (3×90 both lanes by default; `--full --lme20 --mem0` for the one-shot remasure). Gates on an S0 summary unless `--skip-s0-gate`. No published % until those runs finish.
23. **2026-08-19 S0 dual-lane ledger:** product `/recall` **17/180 (9.4%)**; industry search+harness **52/180 (28.9%)**, below July 49.8%. WRITE_MISS 120 vs 94. Not n=1540; not a competitive claim. [artifact](../benchmarks/artifacts/locomo-s0-20260819.md)
24. **2026-08-19 S6 posture:** orchestrator ready; 3×90 may run; **n=1540 / Mem0 same-pin deferred** so the once-per-freeze slot is not burned at 9.4% / 28.9%. LME-20 remasure still outstanding.

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
| PR8 Hop executor V3 | PROOF_MISS / PLANNING_MISS | **R4 landed (measurement)** | 1×30 MH **10/10**; full-suite MH still 21/282 (7.4%) |
| PR9 Assistant-generated memories | EXTRACTION_COVERAGE_MISS | **Merged to `dev`** (#105) | Skip phatic assistant `conversation_episode`; keep `assistant_stated` |
| PR10 Frozen competitive qualification | — | **Measured; not claimed** | Fresh Mem0 **11/30** (MH 6/10, OD 3/4) vs Brainy **21/30** (MH 10/10, OD **0/4**). Lead this freeze; trail OD. Full `/recall` **11.4%**. LME-20 quality **4/20**. **No SOTA / beats-Mem0.** |

## Still open (honest)

- Conversational quality: LoCoMo 1×30 **21/30** (lead Mem0 11/30 this freeze; trail OD **0/4 vs 3/4**); full `/recall` **11.4%** (dip vs July search+harness 49.8%; not vs Mem0 92.5% on the same path); LME-20 **4/20** (multi-session 0/5); BEAM 100K **8/20**. 1×30 is measurement, not qualification.
- Two stacked gaps: **answer-path** (`/recall` slogans/enumerate/abstain; 11.4% to ~50% is directional, not a current-SHA ceiling) then **representation** (WRITE_MISS / thin facts; identity/relations/hops). See [dip why](../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md) · [parity-gap verdict](./external-reviews/2026-08-17-parity-gap-verdict.md).
- Reader-only as the SOTA bet is **rejected**. Hard episode-drop before compiler coverage is also **rejected**. See [sota-representation-path.md](./sota-representation-path.md).
- LME-500 and BEAM 1M/10M not run; do not invent those scores.
- Pack authority / procedures / conflict packets (deferred behind conversational parity)
- Hash/128 re-embed residue

## Claims discipline

- Allowed: remasure pins above; 1×30 **21/30** as **lead this freeze** overall/MH/temporal and **trail OD**; full **11.4%** as a named dip vs July 49.8% search+harness; LME **4/20** as lift vs own 0/20 integrity, not vs published 94.4%; OpMem/marketing non-reg vs Mem0 this cycle; historical Gate 0 / harden / Wave 1 / R-series pins as history only.
- Forbidden: unqualified “beats Mem0”; SOTA; “MH solved”; publishing 70% as full LoCoMo; restoring 49.8% as current; calling 11.4% a harness glitch; mixing 92.5 vs 70; comparing LME 4/20 to Mem0 94.4% or SuperMemory 95 Recall@15; comparing BEAM 8/20 to Mem0 64.1 (1M); inventing LME-500 / BEAM 1M; LoCoMo-named product rules; stuffing episodes to restore OD/SH.
