# Brainy — External Agent Assessment Pack

**Status:** Canonical handoff artifact for external agents / reviewers  
**Date:** 2026-08-04 (updated after external architecture verdict)  
**How to use:** Pass this file (plus optional [codebase-graph.md](./codebase-graph.md) / [codebase-graph.json](./codebase-graph.json)) to an external coding or research agent.

| Related doc | Role |
| --- | --- |
| [external-reviews/2026-08-04-architecture-verdict.md](./external-reviews/2026-08-04-architecture-verdict.md) | **Latest course correction** (accepted) |
| [external-reviews/README.md](./external-reviews/README.md) | Standing intake SOP for future reviews |
| [codebase-graph.md](./codebase-graph.md) | Visual/structural map |
| [codebase-graph.json](./codebase-graph.json) | Machine-readable graph |
| [sota-end-to-end-program.md](./sota-end-to-end-program.md) | Program of record |
| [program-execution-status.md](./program-execution-status.md) | Execution + measurement notes |

---

## Architecture verdict (read this first)

**Approve the five-plane target. Do not treat the current implementation as having reached it.**

Brainy today is a **record-centric memory service mid-migration**. OpMem/vertical greens validate legacy `memory_records` more than the new planes.

| Plane | Honest status |
| --- | --- |
| Source | Live (`IngestRequest` messages) |
| Evidence | Moving shadow → **raw Evidence Plane v2** |
| Semantic | Text-first + partial atoms/events |
| Projection | `current_state` rebuildable MVP — **not** canonical truth |
| Recall | Synthesis over search; not yet a planner |

**Next sequence (do not fusion-retune first):** raw evidence → typed semantics → temporal truth → kill scan-heavy retrieval → plan evidence → executable packs.

Verified debts (updated): FTS `ts_rank_cd` plumbed when available; **pgvector ANN pinned via additive `embedding_vec_768` (mig 18/19; legacy `vector(128)` retained)**; multi-hop gate tightened but subject scans remain; pack v2 YAML scaffolds not loaded by registry; `/recall` still not a typed planner.

---

## 0. One-paragraph product definition


Brainy is a **Go + Postgres memory service** for agents and products. It stores durable facts with **tenant/subject isolation**, **lifecycle** (suppress / supersede / correct), **vertical packs** (marketing, support), and **hybrid retrieval** (FTS + dense + entity hub + typed atoms). It competes on **operational correctness** and **governed vertical memory**, while pursuing **credible conversational recall** (LoCoMo / LongMemEval / BEAM) under an **anti-benchmax** rule: no benchmark surface forms in product code.

---

## 1. Assessment charter (what we want from you)

Please return structured feedback covering:

1. **Architecture fit** — Are the five planes (source → evidence → semantic → projection → recall) the right substrate for both conversational and vertical SOTA on Postgres?
2. **Diagnosis** — Given measured gaps (esp. multi-hop / LongMemEval), is the failure taxonomy and Phase ordering correct?
3. **Missing techniques** — What 2025–26 papers/systems should we borrow next (without graph-DB default)?
4. **Risks** — Where is Brainy over-fitted, under-instrumented, or claiming more than evidence supports?
5. **Concrete next PRs** — Name 3–7 reviewable changes with expected failure-class impact.

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
   atoms/LLM     evidence-set   answer_status   marketing/support
        │             │              │              │
        └─────────────┴──────────────┴──────────────┘
                              │
                              ▼
                 internal/store/postgres
         memory_records (+ FTS) · embeddings · atoms
         entity_links · evidence · events · current_state · jobs
```

**Full graphs:** [codebase-graph.md](./codebase-graph.md)  
**JSON:** [codebase-graph.json](./codebase-graph.json)

### Hot path files (start here)

| Priority | Path | Why |
| ---: | --- | --- |
| 1 | `internal/memory/service.go` | Ingest + `SearchOpt` (~2k LOC) |
| 2 | `internal/memory/fusion_v2.go` | Default-on fusion + semantic-only floor |
| 3 | `internal/memory/recall.go` | Product synthesis / abstention |
| 4 | `internal/memory/attribute_atoms.go` | Deterministic typed atoms |
| 5 | `internal/memory/provider_extractor.go` | Async LLM extract |
| 6 | `internal/store/postgres/migrations.go` | Schema v1–v16 |
| 7 | `internal/store/postgres/events.go` | Evidence/events/current_state MVP |
| 8 | `packs/support/v2/` + `packs/marketing/v2/` | Vertical pack v2 scaffolds |
| 9 | `evals/run_opmem.py` + `fixtures/opmem/` | Operational leadership suite |
| 10 | `evals/public/locomo/` | Conversational publish smoke/full |

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
- v2 scaffolds: entities, state machines (support), brand/outcome extensions (marketing)  
- Fixtures under `fixtures/vertical/{marketing,support}`

### 4.3 Conversational retrieval (in progress toward SOTA)

- Hybrid: FTS + dense + entity hub + atom enumeration  
- Fusion V2 default-on (additive semantic/bm25/entity; semantic-only floor 0.78)  
- Evidence-set selection for list/multi-hop  
- Bounded subject corpus (`ListMemoriesLimited`)  
- Search/recall traces + intent labels + answer_status  

### 4.4 Mid-migration substrate

| Layer | Status |
| --- | --- |
| `memory_records` primary reads | **Live** |
| Typed `memory_atoms` + bitemporal cols | **Live** |
| `memory_evidence` shadow writes | **Live (shadow)** |
| `memory_events` + participants | **MVP** |
| `memory_current_state` projection | **MVP** |
| Full planner / typed tools / as-of API | **Partial** (intents + statuses; tools thin) |

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

**Interpretation for reviewers:** Operational + vertical suites are green. Conversational smoke improved vs an early Fusion V2 25% pin, but **multi-hop and LongMemEval remain the credibility gap**. Smoke ≠ full 3-seed SOTA proof.

Artifacts: `docs/benchmarks/artifacts/locomo-smoke-main-progress-20260804.md`, `main-progress-20260804.json`, prior publish-stack files under `docs/benchmarks/artifacts/`.

---

## 6. Failure taxonomy (use when diagnosing)

Primary labels (also in `internal/memory/trace.go`):

`SOURCE_MISS` · `WRITE_MISS` · `REPRESENTATION_MISS` · `ENTITY_LINK_MISS` · `RETRIEVAL_MISS` · `EVIDENCE_COVERAGE_MISS` · `TEMPORAL_RESOLUTION_MISS` · `CONFLICT_RESOLUTION_MISS` · `PLANNING_MISS` · `READER_MISS` · `ABSTENTION_MISS` · `JUDGE_MISS` · `HARNESS_ERROR`

Oracle helpers: `evals/public/oracle.py`, `evals/public/stage_oracle.py`, recall `oracle_mode` (must be operational or explicitly `oracle_unsupported`).

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
| 0 Baseline + traces + oracle | Diagnostic truth | Traces landed; **oracle/ledger hardening in progress** |
| 1 Fusion / overfetch / evidence-set | Retrieval reliability | Useful progress; **not true BM25 fusion yet** |
| 2 Evidence + bitemporal | Immutable source + time | **Evidence Plane v2 (raw) in progress**; atoms partial |
| 3 Events / procedures / profiles | Representation | MVP events; **typed extract still flat text** |
| 4 Planner + tools + abstention | Query controller | Intents cosmetic; **not a planner** |
| 5 Packs v2 | Vertical leadership | **Scaffolds only** (registry loads pack.yaml) |
| 6 Neutral proof | SOTA qualification | Smoke only; full multi-seed pending |
| 7–8 Associative / learned policy | Research gates | Deferred |

**Post–2026-08-04 review priority:** PR1 oracle ledger (**operational**) → PR2 raw evidence (**landed**) → PR3 typed extract (**v3 slots + async atoms**) → PR4 temporal resolver (**as_of / guarded current_state**) → PR5 retrieval store (**768-d pgvector landed**; traces/OR-FTS remain) → PR6 planner → PR7 executable packs.

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

# LoCoMo smoke
python3 evals/public/locomo/run_smoke.py --base-url "$BRAINY_BASE_URL" \
  --conversations 2 --questions 30 --async-timeout 720

# Unit / integration
go test ./internal/memory/ ./internal/api/ ./internal/store/postgres/ -count=1
```

Pins: Brainy commit SHA, dataset hash, answerer/judge models, top-k, Fusion flags — see smoke manifests under `docs/benchmarks/artifacts/`.

---

## 10. Explicit open questions for external agents

1. Is **multi-hop** primarily an extraction/representation miss or an evidence-set/planning miss given atoms + Fusion V2? What measurement would decide?  
2. Should **`memory_evidence` become the searchable primary** before investing more in ranking?  
3. For vertical SOTA, is deepening **procedures/gotchas** higher ROI than a third pack?  
4. What is a fair **Mem0 comparable stack** we should re-run under identical judge/budget?  
5. Which latency cut (index, candidate generation, or fusion) is most likely to hit p50 &lt; 500 ms?  
6. Are answer_status / abstention semantics complete enough for enterprise agents?  
7. What would you delete from `service.go` to reduce complexity without losing OpMem/vertical wins?

---

## 11. Handoff checklist for the receiving agent

- [ ] Read **Architecture verdict** + §§0–5  
- [ ] Read [2026-08-04 architecture verdict](./external-reviews/2026-08-04-architecture-verdict.md)  
- [ ] Skim [codebase-graph.md](./codebase-graph.md) diagrams  
- [ ] Load [codebase-graph.json](./codebase-graph.json) if using tools  
- [ ] Inspect `SearchOpt` + `ScoreAndRankV2` + `Recall`  
- [ ] Skim migrations v12–v16 and one pack v2 directory  
- [ ] Review latest artifacts in `docs/benchmarks/artifacts/`  
- [ ] Produce: architecture verdict, gap diagnosis, prioritized PR list, rejected ideas  

---

## 12. Appendix — directory cheat sheet

```text
cmd/api, cmd/worker          process entrypoints
internal/memory/             product brain
internal/store/postgres/     schema + SQL
internal/api/                HTTP
internal/jobs/               async worker
packs/{marketing,support}/   vertical packs v1/v2
fixtures/{opmem,vertical}/   hermetic suites
evals/                       runners (OpMem, vertical, public benches)
docs/research/               strategy + this pack
docs/benchmarks/artifacts/   measured runs (redacted)
docs/research/adr/           architecture decisions
```

---

*Prepared so an external agent can assess Brainy without tribal knowledge. Prefer citing SHAs and artifacts over vibes. Numbers above mix 2026-08-01 publish-stack and 2026-08-04 progress retests — treat smoke as directional.*
