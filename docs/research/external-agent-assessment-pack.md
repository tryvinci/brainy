# Brainy — External Agent Assessment Pack

**Status:** Canonical handoff artifact for external agents / reviewers  
**Date:** 2026-08-11 (architect PR1–PR7 **closed**; V3 hardening #93–#98 **on `main` + `dev`**)  
**How to use:** For the current external pass, prefer the dedicated self-review prompt first. Use this pack for architecture context + reproduce commands.

| Related doc | Role |
| --- | --- |
| [external-reviews/2026-08-11-hardening-self-review-prompt.md](./external-reviews/2026-08-11-hardening-self-review-prompt.md) | **Start here** — dedicated self-review prompt for the hardening cutover |
| [external-reviews/README.md](./external-reviews/README.md) | Standing intake SOP + current priority |
| [external-reviews/2026-08-10-v3-rereview-brief.md](./external-reviews/2026-08-10-v3-rereview-brief.md) | Prior V3 re-review brief (historical) |
| [external-reviews/2026-08-07-recall-contract-verdict.md](./external-reviews/2026-08-07-recall-contract-verdict.md) | Accepted recall-contract course correction |
| [external-reviews/2026-08-04-architecture-verdict.md](./external-reviews/2026-08-04-architecture-verdict.md) | Architecture course correction (accepted; **sequence closed**) |
| [recall-contract-v3-hardening-qualification-20260811.md](../benchmarks/artifacts/recall-contract-v3-hardening-qualification-20260811.md) | Hardening qualification + pins |
| [codebase-graph.md](./codebase-graph.md) | Visual/structural map |
| [codebase-graph.json](./codebase-graph.json) | Machine-readable graph |
| [sota-end-to-end-program.md](./sota-end-to-end-program.md) | Program of record |
| [program-execution-status.md](./program-execution-status.md) | Execution + measurement notes |

---

## Architecture verdict (read this first)

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

### Latest measurements

| Pin | Result |
| --- | ---: |
| Staging Gate 0 1×30 (deploy `9bad898`) | **60% (18/30)**, MH **50%**, OD **25%** — [pin](../benchmarks/artifacts/locomo-staging-gate0-1x30-pin-20260811.md) |
| Staging Gate 0 3×90 | **35.6% (32/90)**, MH **19.4%**, OD **42.9%** — [pin](../benchmarks/artifacts/locomo-staging-gate0-3x90-pin-20260811.md) |
| Harden local 1×30 (PRs #93–#97) | **46.7% (14/30)**, MH **5/10**, OD **2/4** — [pin](../benchmarks/artifacts/locomo-harden-1x30-pin-20260811.md) |
| Post-cutover staging 1×30 (`1f2f26f`) | **50% (15/30)**, MH **50%**, OD **25%** — [pin](../benchmarks/artifacts/locomo-staging-postcutover-1x30-pin-20260811.md) |
| Post-cutover staging OpMem / marketing (`1f2f26f`) | **13/13** / **passed** |
| LoCoMo 1×30 V3 early `locomo-v3-early-20260810` | **53.3% (16/30)**, MH **50%**, hybrid **17/30** — [pin](../benchmarks/artifacts/locomo-v3-early-pin-20260810.md) |
| LoCoMo V3 3×90 | **34.4% (31/90)** — [pin](../benchmarks/artifacts/locomo-v3-multiconvo-pin-20260810.md) |
| OpMem / marketing (Gate 0 + harden) | **13/13** / **passed** |
| LME-20 product-recall | Path `/recall` proven; publish run aborted on job failure — **not publishable** — [note](../benchmarks/artifacts/lme20-product-recall-partial-20260811.md) |

**Implication for next agent (2026-08-11):** Hardening #93–#98 is **merged** to `dev` (`1f2f26f`) and `main` (`308d3a1`). Gate 0 baseline remains the pre-harden staging pin. Harden local 1×30 **dipped** vs Gate 0 (expected stricter `hop_join_proven`) — not a win. Do **not** re-open fusion/graph/category dicts. Next: clean LME-20 publish, finish staging re-pin on `1f2f26f`, Mem0 same-pin, multi-seed LoCoMo, then external adjudication via the self-review prompt.

Start from [2026-08-11-hardening-self-review-prompt.md](./external-reviews/2026-08-11-hardening-self-review-prompt.md) + [recall-contract-v3-hardening-qualification-20260811.md](../benchmarks/artifacts/recall-contract-v3-hardening-qualification-20260811.md). Still reject: fusion retune, graph DB, category dictionaries, conversational SOTA claims.
---

## 0. One-paragraph product definition

Brainy is a **Go + Postgres memory service** for agents and products. It stores durable facts with **tenant/subject isolation**, **lifecycle** (suppress / supersede / correct), **vertical packs** (marketing, support), and **hybrid retrieval** (FTS + dense + entity hub + typed atoms). It competes on **operational correctness** and **governed vertical memory**, while pursuing **credible conversational recall** (LoCoMo / LongMemEval / BEAM) under an **anti-benchmax** rule: no benchmark surface forms in product code.

---

## 1. Assessment charter (what we want from you)

Please return structured feedback covering:

1. **Architecture fit** — Are the five planes (source → evidence → semantic → projection → recall) the right substrate for both conversational and vertical SOTA on Postgres?
2. **Diagnosis** — Given measured gaps (esp. multi-hop / LongMemEval) and the new **READER_MISS-dominant** ledger, is the failure taxonomy and Phase ordering still correct?
3. **Missing techniques** — What 2025–26 papers/systems should we borrow next for **reader/synthesis over packets** (without graph-DB default)?
4. **Risks** — Where is Brainy over-fitted, under-instrumented, or claiming more than evidence supports?
5. **Concrete next PRs** — Name 3–7 reviewable changes with expected failure-class impact (prefer `READER_MISS` / multi-hop composition).

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

**Active priority (post V3 hardening cutover):**  
1. Clean isolated LME-20 `--publish --product-recall` (then LME-100)  
2. Finish staging re-pin on deploy `1f2f26f` + Mem0 same-pin  
3. Multi-seed LoCoMo under recall-contract honesty  
4. Pack authority / procedures / conflict packets  
5. Evidence-as-search-primary only if ledger shows retrieval/coverage misses

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

**Git lines:** staging Render auto-deploys from **`dev`** (live `1f2f26f`); **`main`** is production (hardening cutover `308d3a1`).

---

## 10. Explicit open questions for external agents

1. Given **READER_MISS + supported oracles**, should the next PR be a **deterministic packet compiler** (enumeration/temporal/multi-hop) or an **LLM reader bound to packet IDs only**?  
2. Is multi-hop still partly a **representation** miss that oracles cannot see (atoms present but wrong shape)?  
3. Should **`memory_evidence` become the searchable primary** before more ranking work?  
4. For vertical SOTA, is deepening **procedures/gotchas** higher ROI than a third pack?  
5. What is a fair **Mem0 comparable stack** we should re-run under identical judge/budget?  
6. Which latency cut (index, candidate generation, or fusion) is most likely to hit p50 &lt; 500 ms now that smoke p50 ≈ 807 ms?  
7. Are answer_status / abstention semantics complete enough for enterprise agents?  
8. What would you delete from `service.go` to reduce complexity without losing OpMem/vertical wins?

---

## 11. Handoff checklist for the receiving agent

- [ ] Read [2026-08-11 hardening self-review prompt](./external-reviews/2026-08-11-hardening-self-review-prompt.md) first  
- [ ] Read **Architecture verdict** + latest measurement shift (§ top)  
- [ ] Read Gate 0 + harden pins under `docs/benchmarks/artifacts/*20260811*`  
- [ ] Skim [codebase-graph.md](./codebase-graph.md) diagrams  
- [ ] Inspect `hop_executor.go`, `reader_hybrid.go`, `provider_extractor.go`, claim serialization in `store.go`  
- [ ] Review latest artifacts in `docs/benchmarks/artifacts/` and `docs/benchmarks/runs/`  
- [ ] Produce TEMPLATE verdict + findings table + next 3–7 PRs + kill list  

---

## 12. Appendix — directory cheat sheet

```text
cmd/api, cmd/worker          process entrypoints
internal/memory/             product brain (recall/planner/fusion/temporal)
internal/store/postgres/     schema + SQL (through mig 19)
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

*Prepared so an external agent can assess Brainy without tribal knowledge. Prefer citing SHAs and artifacts over vibes. Latest conversational signal for this cutover: Gate 0 staging 18/30 + harden local 14/30 (honest dip) — not publishable SOTA.*
