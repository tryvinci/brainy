# Brainy — External Agent Assessment Pack

**Status:** Canonical handoff artifact for external agents / reviewers  
**Date:** 2026-08-17 (current-SHA parity-gap review received; R0-R4 **closed**; next is **R5A** structured-first `/recall`)  
**How to use:** Start from **CURRENT (2026-08-17)** and the [parity-gap verdict](./external-reviews/2026-08-17-parity-gap-verdict.md). Use this pack for architecture context + reproduce commands. Do **not** treat Gate 0 / "next is R1b" / "LME 0/20" / Wave 1 P0-P7 / "cite-facts vs R5 OD as two bets" language below as live truth. Do **not** re-queue R0-R4 because a review was pinned to `bd987fa`.

| Related doc | Role |
| --- | --- |
| [external-reviews/2026-08-17-parity-gap-verdict.md](./external-reviews/2026-08-17-parity-gap-verdict.md) | **This pass received (live)** — keep course; adjust sequence R5A-R10; 49.8% is not a current-SHA ceiling |
| [external-reviews/2026-08-17-parity-gap-review.md](./external-reviews/2026-08-17-parity-gap-review.md) | Source report (docs `8492ad3`, product `1b5ab3e`) |
| [external-reviews/2026-08-17-competitive-archaeology-verdict.md](./external-reviews/2026-08-17-competitive-archaeology-verdict.md) | **Historical** — Wave 1 `bd987fa`; keep "do not re-queue P0-P7" |
| [external-reviews/2026-08-17-full-recall-self-review-prompt.md](./external-reviews/2026-08-17-full-recall-self-review-prompt.md) | Commission prompt used for the dip (now answered) |
| [locomo-full-recall-dip-why-20260817.md](../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md) | Why full LoCoMo is 11.4% on `/recall` (two stacked gaps; ceiling honesty) |
| [sota-representation-path.md](./sota-representation-path.md) | Accepted execution path (R5A first; R5B-R10 follow) |
| [external-reviews/2026-08-11-competitive-architecture-verdict.md](./external-reviews/2026-08-11-competitive-architecture-verdict.md) | Accepted competitive program (KEEP V3; adjust next sequence) |
| [competitive/README.md](./competitive/README.md) | Competitive archaeology SOP |
| [competitive/competitive-gap-matrix.md](./competitive/competitive-gap-matrix.md) | Living Mem0/Graphiti/Brainy gap matrix |
| [competitive/cycle-closeout.md](./competitive/cycle-closeout.md) | Required cycle closeout (see **2026-08-15/16**) |
| [external-reviews/README.md](./external-reviews/README.md) | Standing intake SOP + current priority |
| [external-reviews/2026-08-14-representation-path-additions.md](./external-reviews/2026-08-14-representation-path-additions.md) | Accepted R1c amendment (historical) |
| [external-reviews/2026-08-11-hardening-self-review-prompt.md](./external-reviews/2026-08-11-hardening-self-review-prompt.md) | Hardening cutover self-review prompt (historical) |
| [external-reviews/2026-08-10-v3-rereview-brief.md](./external-reviews/2026-08-10-v3-rereview-brief.md) | Prior V3 re-review brief (historical) |
| [external-reviews/2026-08-07-recall-contract-verdict.md](./external-reviews/2026-08-07-recall-contract-verdict.md) | Accepted recall-contract course correction |
| [external-reviews/2026-08-04-architecture-verdict.md](./external-reviews/2026-08-04-architecture-verdict.md) | Architecture course correction (accepted; **sequence closed**) |
| [recall-contract-v3-hardening-qualification-20260811.md](../benchmarks/artifacts/recall-contract-v3-hardening-qualification-20260811.md) | Hardening qualification + pins (historical) |
| [codebase-graph.md](./codebase-graph.md) | Visual/structural map |
| [codebase-graph.json](./codebase-graph.json) | Machine-readable graph |
| [sota-end-to-end-program.md](./sota-end-to-end-program.md) | Prior program of record (superseded for next sequence by competitive verdict) |
| [program-execution-status.md](./program-execution-status.md) | Execution + measurement notes |

---

## CURRENT (2026-08-17) — live truth; read this first

**Repo tips:** `dev` = `main` = `8492ad3` (docs remasure + production FF 2026-08-17, explicit approval). Product code for those pins is still **`1b5ab3e`**. No product change in the remasure.

**Do not use as live:** Gate 0 18/30, harden 14/30, “next is R1b”, LME **0/20**, Wave 1 “READER_MISS-dominated”, Mem0 same-pin **12/30**.

### Live pins (2026-08-15 remasure)

| Suite | Brainy | Mem0 this cycle / published |
| --- | ---: | ---: |
| OpMem | **13/13** | **10/13** |
| Marketing | **17/17** | **4/17** empirical |
| LoCoMo 1×30 | **21/30** (MH **10/10**, OD **0/4**, temporal **11/16**) | **11/30** (MH 6, OD **3/4**, temporal 2) |
| LoCoMo full n=1540 | **175/1540 (11.4%)** product `/recall` | published **92.5%** (top-k 200, their harness) — not this pin |
| LME-20 | **4/20** product `/recall` | no fair pin on this harness |
| LME-500 / BEAM 1M/10M | **not run** (cost; LME-20 4/20 and BEAM 100K 8/20 already diagnose) | — |
| BEAM 100K | **8/20** search+harness | published 64.1 is BEAM **1M** |

1×30 did **not** drop vs R4h 20/30. Full LoCoMo **did** drop vs July **49.8%** because this remasure scored product `POST /recall` instead of search-hits + harness LLM answerer. Diagnosis: [locomo-full-recall-dip-why-20260817.md](../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md).

### Two stacked gaps

1. **Answer-path:** `/recall` cites slogans / enumerate / abstain (`firstStatementFromPacket` in `recall.go`). Directionally the first lever. **Modify:** 11.4% to ~50% is **not** a measured ceiling on current storage — July 49.8% was a different stack. Size it with a **current-SHA search+harness** diagnostic (stratified subset after R5A).
2. **Representation:** compiler coverage does not generalize (1×30 21/30 vs full SH 10.5%; LME multi-session 0/5). Even July search+harness was 49.8% vs Mem0 Platform 92.5% (n=1540, top-k 200). Canonical identity is the largest **structural** gap after the answer path (`entities.go` string keys; hops join mentions).

**Implication for next agent:** Current-SHA parity-gap review received 2026-08-17 ([verdict](./external-reviews/2026-08-17-parity-gap-verdict.md)). R0-R4 stay closed. Next product work is **R5A structured-first `/recall`** (retire `firstStatementFromPacket` as a normal factual strategy). OD 0/4 is a diagnostic, not the PR name. Then R5B typed packet, R6 coverage V2, R7-R9 identity/relations/hops, R10 dual-path freeze. Two published lanes: product `/recall` vs search+harness; do not mix with Mem0 92.5%. Do **not** claim beats-Mem0 / SOTA. Still reject: fusion retune, graph DB default, category dictionaries, LoCoMo-named rules, episode top-k to restore OD/SH, LME-500 as a quality claim, another full remasure this cycle, re-queue of the Wave 1 P0-P7 list, v2 DDL in R5A.

---

## Architecture verdict (context; live pins are CURRENT above)

**Approve the five-plane target. Do not treat the current implementation as SOTA-ready.**

The **2026-08-04 architect sequence (PR1–PR7) is complete for that pass.** Brainy remains a record-centric service mid-migration toward the five-plane target; OpMem/vertical greens still lean on `memory_records`. Do **not** re-litigate PR1–PR7 unless new evidence reopens a finding.

| Plane | Honest status |
| --- | --- |
| Source | Live (`IngestRequest` messages) |
| Evidence | **Raw Evidence Plane v2** (subject-safe dedupe) |
| Semantic | Text-first + typed extract v3 → atoms (sync + async) |
| Projection | Guarded `current_state` on **sync + async**; as_of / history reads |
| Recall | `/recall` consumes structured `evidence_packet` + plan; optional **hybrid LLM reader** (`BRAINY_RECALL_LLM`) |

### Accepted sequence — closed (2026-08-05)

| Step | Status |
| --- | --- |
| 1 Stage oracles + failure ledger | **Complete** |
| 2 Evidence Plane v2 (raw) | **Complete** |
| 3 Typed extract v3 + async atoms | **Complete** |
| 4 Temporal resolver + guarded current_state | **Complete** (async parity landed) |
| 5 Retrieval store (FTS rank + 768-d pgvector) | **Complete** (HNSW valid; corpus/FTS honesty patches) |
| 6 Typed query planner + evidence packets | **Complete** (reader synthesizes from packet; `tools_executed` honest) |
| 7 Executable packs v2 | **Complete** for architect finding (entities + state-machines loaded; support + marketing FSMs) |

Still **explicitly open** (not part of claiming PR1–PR7 done): pack authority / procedures / conflict packets; evidence-as-search-primary; hash/128 re-embed residue; finished LME-100 under contract; Phase-6 multi-seed SOTA gates. Fresh LoCoMo same-pin under recall-contract is recorded (Brainy 16/30 vs Mem0 11/30) — re-pin on staging after deploy.

### Historical measurements (pre-remasure; do not use as live)

| Pin | Result |
| --- | ---: |
| Staging Gate 0 1×30 (deploy `9bad898`) | **60% (18/30)**, MH **50%**, OD **25%** — [pin](../benchmarks/artifacts/locomo-staging-gate0-1x30-pin-20260811.md) |
| Staging Gate 0 3×90 | **35.6% (32/90)**, MH **19.4%**, OD **42.9%** — [pin](../benchmarks/artifacts/locomo-staging-gate0-3x90-pin-20260811.md) |
| Harden local 1×30 (PRs #93–#97) | **46.7% (14/30)**, MH **5/10**, OD **2/4** — [pin](../benchmarks/artifacts/locomo-harden-1x30-pin-20260811.md) |
| Post-cutover staging 1×30 (`1f2f26f`) | **50% (15/30)**, MH **50%**, OD **25%** — [pin](../benchmarks/artifacts/locomo-staging-postcutover-1x30-pin-20260811.md) |
| Post-cutover staging 3×90 (`1f2f26f`) | **36.7% (33/90)**, MH **22.2%**, OD **42.9%** — [pin](../benchmarks/artifacts/locomo-staging-postcutover-3x90-pin-20260811.md) |
| Post-cutover staging OpMem / marketing (`1f2f26f`) | **13/13** / **passed** |
| LoCoMo 1×30 V3 early `locomo-v3-early-20260810` | **53.3% (16/30)**, MH **50%**, hybrid **17/30** — [pin](../benchmarks/artifacts/locomo-v3-early-pin-20260810.md) |
| LoCoMo V3 3×90 | **34.4% (31/90)** — [pin](../benchmarks/artifacts/locomo-v3-multiconvo-pin-20260810.md) |
| OpMem / marketing (Gate 0 + harden) | **13/13** / **passed** |
| LME-20 product-recall | **Publishable** 0/20 `/recall`, jobs 4829=4829 failed=0 — [pin](../benchmarks/artifacts/lme20-product-recall-pr1-20260812-pin.md) |
| Local PR2 LoCoMo 1×30 (`24be5ab`) | **6/30**, MH 4/10 — [pin](../benchmarks/artifacts/locomo-pr2-dev-1x30-20260813.md) |
| Wave 1 local LoCoMo 1×30 (`a7a5184`) | **14/30**, MH **3/10**, temporal **9/16** — [pin](../benchmarks/artifacts/locomo-wave1-dev-1x30-20260813.md) |
| R1c local LoCoMo 1×30 (`21a632b`, PR #113) | **10/30**, MH **2/10**, OD **0/4**, temporal **8/16** — [pin](../benchmarks/artifacts/locomo-r1c-dev-1x30-20260814.md) (dip vs Wave 1; not a compiler win) |
| R1c local OpMem / marketing (`21a632b`) | **13/13** / **17/17** — [opmem](../benchmarks/artifacts/opmem-r1c-local-20260814.md) · [marketing](../benchmarks/artifacts/marketing-r1c-local-20260814.md) |

**Historical implication (2026-08-14, superseded):** that paragraph said next was R1b. R1b–R4 have since landed. Use **CURRENT (2026-08-17)** above.

---

## 0. One-paragraph product definition

Brainy is a **Go + Postgres memory service** for agents and products. It stores durable facts with **tenant/subject isolation**, **lifecycle** (suppress / supersede / correct), **vertical packs** (marketing, support), and **hybrid retrieval** (FTS + dense + entity hub + typed atoms). It competes on **operational correctness** and **governed vertical memory**, while pursuing **credible conversational recall** (LoCoMo / LongMemEval / BEAM) under an **anti-benchmax** rule: no benchmark surface forms in product code.

---

## 1. Assessment charter (what we want from you)

Please return structured feedback covering:

1. **Architecture fit** — Are the five planes still the right substrate? (Assume yes unless new evidence.)
2. **Diagnosis** — Two-gap split is directional; 49.8% is not a current-SHA ceiling. Keep representation-first. Reject re-queue of R0-R4.
3. **Published metric** — Two lanes: product `/recall` on the bake-off row; search+harness as industry-format, labeled separately.
4. **First product PR** — R5A structured-first `/recall`. OD 0/4 is a diagnostic, not the PR name.
5. **Fair Mem0 stack** — n=1540 component lane, top-k 200, LLM-over-search, Platform vs OSS labeled, after a freeze.
6. **Concrete next PRs** — R5A-R10 as in the parity-gap verdict. Not fusion, not graph DB, not LME-500-as-quality, not Wave 1 P0-P7, not v2 schema in R5A.

**Do not** propose LOCOMO-named regexes, held-out conversation prompt tuning, or “just add a graph database” without a measured gate.

---

## 2. Codebase graph (summary)

```text
Clients / evals
    │
    ▼
cmd/api ──────────────────────────────► cmd/worker (async extract)
    │                                         │
    ▼                                         ▼
internal/api ──► internal/memory.Service ◄── internal/jobs.Processor
                      │
        ┌─────────────┼──────────────┬──────────────┐
        ▼             ▼              ▼              ▼
   Extractors    Fusion V2      POST /recall    Packs YAML
   atoms/LLM     evidence-set   query_plan +    marketing/support
        │             │         evidence_packet     │
        └─────────────┴──────────────┴──────────────┘
                              │
                              ▼
                 internal/store/postgres
         memory_records (+ FTS) · embeddings (768 HNSW)
         atoms · entity_links · evidence · events · current_state · jobs
```

**Full graphs:** [codebase-graph.md](./codebase-graph.md)  
**JSON:** [codebase-graph.json](./codebase-graph.json)

### Hot path files (start here)

| Priority | Path | Why |
| ---: | --- | --- |
| 1 | `internal/memory/recall.go` + `planner.go` | Product synthesis; plan/packet already emitted |
| 2 | `internal/memory/service.go` | Ingest + `SearchOpt` (~2k LOC) |
| 3 | `internal/memory/fusion_v2.go` | Default-on fusion + semantic-only floor |
| 4 | `internal/memory/evidence_set.go` | List/multi-hop coverage selection |
| 5 | `internal/memory/temporal.go` | as_of / current_state / history |
| 6 | `internal/memory/provider_extractor.go` | Async LLM extract v3 |
| 7 | `internal/store/postgres/migrations.go` | Schema through mig 19 |
| 8 | `internal/pack/pack.go` + `packs/support/v2/` | Sidecars + ticket FSM |
| 9 | `evals/run_opmem.py` + `fixtures/opmem/` | Operational leadership suite |
| 10 | `evals/public/locomo/` + `stage_oracle.py` | Conversational smoke + ledger |

---

## 3. Architectural invariants (non-negotiable)

1. **Anti-benchmax** — denylist CI (`overfit_denylist_test.go`); no benchmark names/phrases in product prompts/rankers.  
2. **Evidence immutability** — raw source not silently rewritten; semantic layers may retire/supersede.  
3. **Tenant/subject enforcement below the LLM** — store queries always scoped.  
4. **Postgres-first** — graph DB only if measured (ADR-004).  
5. **Product `/recall`** — synthesis and abstention are product behavior, not harness-only.  
6. **Predicate-specific temporal policy** — not universal latest-wins (`predicate_policy.go`).

---

## 4. Current system capabilities

### 4.1 Operational memory

- Suppress / correct / supersede APIs  
- Domain events for batch invalidation  
- Auto-supersede for **stateful** predicates  
- OpMem fixture suite (correction, isolation, suppression, staleness, idempotency)

### 4.2 Vertical memory

- YAML packs with vocabulary, lifecycle rules, rank policy  
- v2 **loaded**: entities + state machines (support); brand/outcome extensions (marketing)  
- Support **ticket_status FSM** enforced on ingest  
- Fixtures under `fixtures/vertical/{marketing,support}`

### 4.3 Conversational retrieval (in progress toward SOTA)

- Hybrid: FTS (`ts_rank_cd` when available) + dense 768 ANN + entity hub + atom enumeration  
- Fusion V2 default-on (additive semantic/bm25/entity; semantic-only floor 0.78)  
- Evidence-set selection for list/multi-hop  
- Bounded subject corpus (`ListMemoriesLimited`)  
- Search/recall traces + intent labels + answer_status  
- `/recall` emits and **consumes** `explain.query_plan` + `explain.evidence_packet`

### 4.4 Mid-migration substrate

| Layer | Status |
| --- | --- |
| `memory_records` primary reads | **Live** |
| Typed `memory_atoms` + bitemporal cols | **Live** (async upsert path) |
| `memory_evidence` raw capture | **Live (v2)** |
| `memory_events` + participants | **MVP** |
| `memory_current_state` projection | **Guarded** (sync + async) |
| Planner / evidence packets | **Consumed by `/recall` reader** |
| Reader quality vs judge gold | **Weak** (deterministic extractive baseline 43.3%) |

---

## 5. Measured position (pin these when reviewing)

### Publish-stack baseline (2026-08-01)

| Axis | Signal |
| --- | ---: |
| LoCoMo full 3-seed mean | ≈ **49.8%** |
| LoCoMo multi-hop | ≈ **26%** |
| LongMemEval-S 100 | **4%** |
| BEAM-100K / 20 Q | **40%** |
| Search @ c=8 | p50 ≈ **2.4 s**, p95 ≈ **5.0 s** |

### Post-Fusion-V2 / main progress (2026-08-04)

| Axis | Signal |
| --- | ---: |
| LoCoMo smoke 2×30 Q | **50% (15/30)** |
| — temporal | **66.7%** |
| — multi-hop | **25%** |
| OpMem staging | **13/13** |
| Support | **4/4** |
| Marketing | **17/17** |
| Search smoke p50/p95 | ≈ **1.7 s / 3.0 s** |

### Post planner/packs smoke (2026-08-04 evening)

| Axis | Signal |
| --- | ---: |
| LoCoMo smoke 1×30 Q (LLM over search) | **60% (18/30)** |
| — temporal | **68.8%** |
| — multi-hop | **40%** |
| — open-domain | **75%** |
| Search p50/p95 | ≈ **807 / 1509 ms** |
| Failure ledger | **12/12 READER_MISS**; stage oracles all supported |
| HNSW `embedding_vec_768` | **valid** on staging (~1.4 GB) |

### Architect closeout — product `/recall` smoke (2026-08-05)

| Axis | Signal |
| --- | ---: |
| LoCoMo smoke 1×30 Q (`BRAINY_USE_RECALL=1`, sync) | **43.3% (13/30)** |
| — multi-hop | **50%** |
| — temporal | **43.8%** |
| — open-domain | **25%** |
| Search p50/p95 | ≈ **683 / 1108 ms** |
| Answer models | all `brainy-recall+*` |
| Architect PR1–PR7 | **Closed** |

**Interpretation for reviewers:** Operational + vertical suites remain green. Architect sequence is structurally complete. Conversational score under deterministic packet reader is lower than LLM-over-search (expected); next agent should improve **reader quality**, not reopen PR1–PR7.

Artifacts:  
- `docs/benchmarks/artifacts/locomo-smoke-recall-reader-20260805.md`  
- `docs/benchmarks/artifacts/locomo-smoke-planner-packs-20260804.md`  
- `docs/benchmarks/runs/locomo-smoke-f722342a.json` (+ `.manifest.json`)  
- `docs/benchmarks/artifacts/failure-ledger/locomo-recall-smoke.jsonl`  
- prior: `locomo-smoke-c223da3d.*`, publish-stack under `docs/benchmarks/artifacts/`

---

## 6. Failure taxonomy (use when diagnosing)

Primary labels (also in `internal/memory/trace.go`):

`SOURCE_MISS` · `WRITE_MISS` · `REPRESENTATION_MISS` · `ENTITY_LINK_MISS` · `RETRIEVAL_MISS` · `EVIDENCE_COVERAGE_MISS` · `TEMPORAL_RESOLUTION_MISS` · `CONFLICT_RESOLUTION_MISS` · `PLANNING_MISS` · `READER_MISS` · `ABSTENTION_MISS` · `JUDGE_MISS` · `HARNESS_ERROR`

Oracle helpers: `evals/public/oracle.py`, `evals/public/stage_oracle.py`, recall `oracle_mode` (must be operational or explicitly `oracle_unsupported`).

**Latest ledger skew:** LoCoMo smoke WRONGs are currently **READER_MISS-dominated** with supported upstream oracles — treat that as the default hypothesis until a new run says otherwise.

---

## 7. Competitor / research lenses (what to compare us against)

| System / paper | Borrow | Do not blindly copy |
| --- | --- | --- |
| Mem0 | Fusion pragmatism, hybrid budgets, harness pins | Managed headline scores as ground truth |
| Zep / Graphiti | Temporal filters, episode structure | Mandatory graph DB |
| AtomMem / TriMem | Atomic facts + profiles + raw evidence | Bench-specific extractors |
| APEX-MEM / Chronos | Events + structured ops | Unbounded tool-call agents |
| A-TMA / MemTrace | Stage-level evaluation | Treating QA score as sufficient |
| FunnelStory / Atlan-style | Canonical vertical schemas, authority | Forking a second memory engine per vertical |

Program detail: [sota-end-to-end-program.md](./sota-end-to-end-program.md) §24.

---

## 8. Proposed roadmap (compressed)

| Phase | Intent | Status |
| --- | --- | --- |
| 0 Baseline + traces + oracle | Diagnostic truth | **Oracle modes + failure ledger operational** |
| 1 Fusion / overfetch / evidence-set | Retrieval reliability | Useful; FTS rank plumbed when available |
| 2 Evidence + bitemporal | Immutable source + time | **Evidence Plane v2 landed**; atoms partial |
| 3 Events / procedures / profiles | Representation | MVP events; **typed extract v3 slots live** |
| 4 Planner + tools + abstention | Query controller | **Packet reader live** on `/recall`; quality gap remains |
| 5 Packs v2 | Vertical leadership | **Sidecars + support/marketing FSMs** (authority/procedures still open) |
| 6 Neutral proof | SOTA qualification | Smoke only; full multi-seed + LME pending |
| 7–8 Associative / learned policy | Research gates | Deferred |

**Active priority (competitive program):**  
1. PR1 — LME-20 measurement integrity — **done** (publishable 0/20)  
2. PR2 — Conversational append-only vs governed mutation — **code landed** (no quality pin yet)  
3. PR3 — Temporal features V1 + `temporal_score`  
4. PR4 — Retrieval V4 candidate/context/proof budgets  
5. PR5–PR8 — Context/proof split → entities → relations → hop V3  
6. PR9–PR10 — Assistant memories → frozen competitive qualification  

---

## 9. Reproduce commands (staging or local)

```bash
# OpMem
python3 evals/run_opmem.py --systems brainy --base-url "$BRAINY_BASE_URL"

# Vertical
python3 evals/run_vertical_eval.py --base-url "$BRAINY_BASE_URL" \
  --fixture-dir fixtures/vertical/support
python3 evals/run_vertical_eval.py --base-url "$BRAINY_BASE_URL" \
  --fixture-dir fixtures/vertical/marketing

# LoCoMo smoke + failure ledger
python3 evals/public/locomo/run_smoke.py --base-url "$BRAINY_BASE_URL" \
  --conversations 1 --questions 30 \
  --failure-ledger docs/benchmarks/artifacts/failure-ledger/locomo-smoke.jsonl

# Stratified LME-100
python3 -m public.longmemeval.run --base-url "$BRAINY_BASE_URL" \
  --limit 100 --seed 1 --async-timeout 1800

# Unit / integration
go test ./internal/memory/ ./internal/api/ ./internal/store/postgres/ ./internal/pack/ -count=1
```

Pins: Brainy commit SHA, dataset hash, answerer/judge models, top-k, Fusion flags — see manifests under `docs/benchmarks/runs/` and `docs/benchmarks/artifacts/`.

**Git lines:** `dev` = GitHub homepage / staging; `main` = production. Both currently **`8492ad3`**. Product SHA for the remasure pins is **`1b5ab3e`**. (Historical: Gate 0 staging `9bad898`; hardening cutover `308d3a1` / `1f2f26f`.)

---

## 10. Explicit open questions for external agents

**This pass (2026-08-17) — answered; remaining:**

1. Land **R5A** so `/recall` consumes typed values (not `firstStatementFromPacket`). OD 0/4 is a diagnostic inside that PR.
2. Keep two published lanes: product `/recall` vs search+harness (n=1540). Do not mix 11.4% with Mem0 92.5% as retrieval. Size the answer-path ceiling with current-SHA search+harness on a stratified subset (not a full remasure every PR).
3. Then R5B typed packet, R6 compiler coverage V2, R7-R9 canonical identity / relation V2 / hop ID joins.
4. Fair future Mem0 stack: top-k 200 + LLM-over-search on the *component* lane, labeled Platform vs OSS, after a freeze (R10).
5. LME-500 / BEAM 1M remain skipped until LME-20 quality and BEAM 100K justify the spend.

**Still open (lower priority this pass):**

6. Should **`memory_evidence` become the searchable primary** before more ranking work?
7. For vertical SOTA, is deepening **procedures/gotchas** higher ROI than a third pack?
8. Which latency cut (index, candidate generation, or fusion) is most likely to hit p50 &lt; 500 ms?
9. Are answer_status / abstention semantics complete enough for enterprise agents? (188 full-LoCoMo abstains; OD q14/q27 abstain with WRITE_MISS.)
10. What would you delete from `service.go` to reduce complexity without losing OpMem/vertical wins?

---

## 11. Handoff checklist for the receiving agent

- [ ] Read [2026-08-17 parity-gap verdict](./external-reviews/2026-08-17-parity-gap-verdict.md) (live next-work; R5A first)
- [ ] Skim [Wave 1 archaeology verdict](./external-reviews/2026-08-17-competitive-archaeology-verdict.md) only as history (do not re-queue P0-P7)
- [ ] Read [2026-08-17 full-recall self-review prompt](./external-reviews/2026-08-17-full-recall-self-review-prompt.md) for pin honesty
- [ ] Read [dip diagnosis](../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md) (49.8% is not a current-SHA ceiling)
- [ ] Read **CURRENT (2026-08-17)** at the top of this pack (ignore Gate 0 / R1b / LME 0/20 / Wave 1 P0-P7 as live)
- [ ] Read remasure pins: `locomo-fresh-full-20260815.md`, `locomo-fresh-1x30-20260815.md`, `lme20-fresh-20260815.md`
- [ ] Skim [codebase-graph.md](./codebase-graph.md) diagrams
- [ ] Inspect `recall.go` (`firstStatementFromPacket`), `hop_executor.go` (mention-as-ID, unscoped fallback), `entities.go`, `relations.go`
- [ ] Produce TEMPLATE verdict + findings table + next 3-7 PRs + kill list  

---

## 12. Appendix — directory cheat sheet

```text
cmd/api, cmd/worker          process entrypoints
internal/memory/             product brain (recall/planner/fusion/temporal)
internal/store/postgres/     schema + SQL (through mig 21)
internal/api/                HTTP
internal/jobs/               async worker
internal/pack/               pack registry + sidecars + FSM
packs/{marketing,support}/   vertical packs v1/v2
fixtures/{opmem,vertical}/   hermetic suites
evals/                       runners (OpMem, vertical, public benches)
docs/research/               strategy + this pack
docs/benchmarks/artifacts/   measured runs + failure ledger
docs/benchmarks/runs/        UnifiedResult JSON + manifests
docs/research/adr/           architecture decisions
```

---

*Prepared so an external agent can assess Brainy without tribal knowledge. Prefer citing SHAs and artifacts over vibes. Live conversational signal (2026-08-17): LoCoMo 1x30 **21/30** vs Mem0 **11/30** this freeze (trail OD); full `/recall` **11.4%** (dip vs July search+harness 49.8%; 49.8% is not a current-SHA ceiling; not vs Mem0 92.5% on the same path). Parity-gap review: keep R0-R4 closed; next is R5A structured-first `/recall`. Not publishable SOTA.*
