# Brainy codebase graph

**Audience:** External agents and human reviewers  
**Date:** 2026-08-04  
**Companion:** [external-agent-assessment-pack.md](./external-agent-assessment-pack.md) (single handoff artifact)  
**Machine-readable:** [codebase-graph.json](./codebase-graph.json)

This document is the structural map of the Brainy repository: runtime topology, data planes, package ownership, and eval surfaces. Prefer reading this before diving into `internal/memory/service.go` (~2k LOC).

---

## 1. Runtime topology

```mermaid
flowchart TB
  subgraph Clients
    Agent[Agent / product]
    Eval[Eval harnesses]
  end

  subgraph Processes
    API["cmd/api<br/>HTTP :8080"]
    Worker["cmd/worker<br/>async extract loop"]
  end

  subgraph Core["internal/memory"]
    Svc[Service]
    Ext[Extractor chain]
    Rec[Recall]
    Fus[Fusion V2 + evidence-set]
  end

  subgraph Store["internal/store/postgres"]
    Recs[(memory_records)]
    Emb[(memory_embeddings)]
    Atoms[(memory_atoms)]
    Hub[(memory_entity_links)]
    Ev[(memory_evidence)]
    Events[(memory_events)]
    State[(memory_current_state)]
    Jobs[(ingest / extraction jobs)]
  end

  subgraph Packs["packs/*/v{1,2}"]
    Mkt[marketing]
    Sup[support]
  end

  Agent --> API
  Eval --> API
  API --> Svc
  API --> Rec
  Worker --> Ext
  Worker --> Svc
  Ext --> Svc
  Svc --> Fus
  Svc --> Recs & Emb & Atoms & Hub & Ev & Events & State & Jobs
  Svc --> Packs
  Worker --> Jobs
```

---

## 2. Five memory planes (target architecture)

Product code is mid-migration: `memory_records` remains the primary read path; evidence/events/current_state are **shadow / MVP** (migrations 15–16).

```mermaid
flowchart LR
  P1[1 Source<br/>chat / CRM / tools] --> P2[2 Evidence<br/>immutable]
  P2 --> P3[3 Semantic<br/>atoms / events / assertions]
  P3 --> P4[4 Projection<br/>current_state / profiles]
  P4 --> P5[5 Recall<br/>plan → retrieve → resolve → answer]
```

| Plane | Today (code) | Tables / types |
| --- | --- | --- |
| Source | `IngestRequest` messages | raw ingest payload + jobs |
| Evidence | shadow write on upsert | `memory_evidence` |
| Semantic | records + atoms + events MVP | `memory_records`, `memory_atoms`, `memory_events` |
| Projection | current-state upsert for stateful preds | `memory_current_state` |
| Recall | `SearchOpt` + `POST /recall` | Fusion V2, evidence-set, answer_status |

---

## 3. Ingest paths

```mermaid
sequenceDiagram
  participant C as Client
  participant API as cmd/api
  participant S as memory.Service
  participant W as cmd/worker
  participant PG as Postgres

  Note over C,PG: Sync path — deterministic (+ labeled) extract
  C->>API: POST /ingest
  API->>S: Ingest
  S->>S: extractOrLabel / attribute atoms
  S->>PG: UpsertMemory + embedding + entity links
  S->>PG: Shadow evidence / event / current_state
  S->>PG: autoSupersedePriorState (stateful preds)
  API-->>C: created/updated memories

  Note over C,PG: Async path — provider LLM extract on worker
  C->>API: POST /ingest/async
  API->>PG: EnqueueIngestJob (idempotent)
  API-->>C: ingest_id, job_id
  W->>PG: ClaimNextExtractionJob
  W->>S: ProviderExtractor + BuildMemoryRecord
  W->>PG: Upsert + embed + links + evidence/events
  W->>PG: CompleteExtractionJob
```

**Entry points**

| Path | File |
| --- | --- |
| API boot | `cmd/api/main.go` |
| Worker boot | `cmd/worker/main.go` (+ `scripts/worker-respawn.sh`) |
| HTTP routes | `internal/api/router.go` |
| Sync ingest | `internal/memory/service.go` → `Ingest` |
| Async claim | `internal/jobs/processor.go` → `ProcessNext` |
| LLM extract | `internal/memory/provider_extractor.go` |
| Deterministic extract | `internal/memory/extractor.go` |
| Attribute atoms | `internal/memory/attribute_atoms.go` |

---

## 4. Retrieval / recall path

```mermaid
flowchart TB
  Q[Query] --> Intent[AnalyzeQueryIntents]
  Intent --> Lex[FTS / ILIKE lexical<br/>overfetch]
  Intent --> Dense[Dense embedding scores]
  Intent --> Hub[Entity hub boosts]
  Intent --> Atom[Atom predicate scan<br/>list queries]
  Lex & Dense & Hub & Atom --> Cand[Candidate map<br/>status + lifecycle filter]
  Cand --> Score[scoreMemoryIDF]
  Score --> Fus[ScoreAndRankV2<br/>semantic + bm25 + entity]
  Fus --> Boost[recency / conviction / pack weights]
  Boost --> Sel{list or multi-hop?}
  Sel -->|yes| ES[selectEvidenceSet]
  Sel -->|no| TopK[top-k truncate]
  ES --> Out[SearchResponse + Trace]
  TopK --> Out
  Out --> Recall[POST /recall<br/>context / enumerate / answer]
  Recall --> Status[answer_status + abstention]
```

**Critical flags**

| Env | Default | Effect |
| --- | --- | --- |
| `BRAINY_FUSION_V2` | **on** | Mem0-style additive fusion; semantic-only floor 0.78 |
| `BRAINY_ENTITY_RANKING` | off | Overlap rerank (prior A/B regressor) |
| `BRAINY_IDF_RANKING` | off | IDF-weighted lexical |
| `BRAINY_ENSURE_FTS_INDEX` | off (API) | Controlled GIN build |

---

## 5. Package ownership graph

```mermaid
flowchart TB
  api[internal/api] --> memory[internal/memory]
  api --> auth[internal/auth]
  api --> obs[internal/observability]
  jobs[internal/jobs] --> memory
  jobs --> emb[internal/embedding]
  memory --> emb
  memory --> pack[internal/pack]
  memory --> storeIface[Store interface]
  storeIface --> pg[internal/store/postgres]
  cfg[internal/config] --> api
  cfg --> jobs
  pack --> yaml["packs/**/pack.yaml"]
```

| Package | Responsibility | Hot files |
| --- | --- | --- |
| `internal/memory` | Domain logic: ingest, search, recall, lifecycle, fusion | `service.go`, `recall.go`, `fusion_v2.go`, `attribute_atoms.go` |
| `internal/store/postgres` | Persistence + migrations | `store.go`, `atoms.go`, `events.go`, `entity_hub.go`, `migrations.go` |
| `internal/api` | HTTP + auth wiring | `router.go` |
| `internal/jobs` | Async worker processor | `processor.go` |
| `internal/pack` | Vertical pack registry | `pack.go` |
| `internal/embedding` | Local hash + OpenAI-compatible provider | `local.go`, provider client |
| `evals/` | OpMem, vertical, public LoCoMo/LME/BEAM | `run_opmem.py`, `evals/public/*` |
| `packs/` | Domain models v1/v2 | `marketing`, `support` |
| `fixtures/` | Hermetic eval cases | `opmem`, `vertical/*` |
| `docs/research/` | Strategy + external briefings | this tree |

---

## 6. Data model (logical)

```mermaid
erDiagram
  memory_records ||--o{ memory_embeddings : has
  memory_records ||--o{ memory_atoms : indexes
  memory_records ||--o{ memory_entity_links : linked_via
  memory_records ||--o| memory_evidence : shadowed_by
  memory_records ||--o{ memory_events : may_emit
  memory_events ||--o{ memory_event_participants : has
  memory_atoms ||--o| memory_current_state : projects

  memory_records {
    text tenant_id
    text subject_id
    text memory_id PK
    text content
    text kind
    text lifecycle_state
    text status
    timestamptz observed_at
    jsonb metadata
  }
  memory_atoms {
    text predicate
    text value_norm
    timestamptz valid_from
    timestamptz valid_to
    timestamptz recorded_at
    timestamptz retired_at
  }
  memory_evidence {
    text content_hash
    text suppression_status
    timestamptz occurred_at
  }
  memory_events {
    text event_type
    timestamptz starts_at
    float confidence
  }
```

**Migrations:** v1–v16 in `internal/store/postgres/migrations.go`  
Notable: v12 entity links, v13 atoms, v14 FTS tsv, v15 bitemporal atom cols, v16 evidence/events/current_state.

---

## 7. API surface

| Method | Path | Handler area | Notes |
| --- | --- | --- | --- |
| GET | `/healthz` | api | liveness |
| GET | `/metrics` | api | process metrics |
| POST | `/ingest` | Service.Ingest | sync |
| POST | `/ingest/async` | Service.IngestAsync | queue |
| GET | `/memories/search` | Service.SearchOpt | hybrid + Trace |
| POST | `/recall` | Service.Recall | context/enumerate/answer |
| POST | `/events` | ApplyDomainEvent | batch invalidation |
| POST | `/memories/{id}/suppress` | Suppress | status=suppressed |
| POST | `/memories/{id}/correct` | Correct | rewrite + stickiness |
| POST | `/memories/{id}/supersede` | Supersede | lifecycle supersession |

---

## 8. Eval & benchmark graph

```mermaid
flowchart LR
  subgraph Product
    BrainyAPI[Brainy staging/local]
  end

  subgraph Operational
    OpMem[evals/run_opmem.py]
    FixO[fixtures/opmem]
  end

  subgraph Vertical
    Vert[evals/run_vertical_eval.py]
    MktMVP[evals/run_marketing_mvp_benchmark.py]
    FixV[fixtures/vertical/*]
  end

  subgraph Conversational
    LoCoMo[evals/public/locomo]
    LME[evals/public/longmemeval]
    BEAM[evals/public/beam]
  end

  FixO --> OpMem --> BrainyAPI
  FixV --> Vert --> BrainyAPI
  FixV --> MktMVP --> BrainyAPI
  LoCoMo --> BrainyAPI
  LME --> BrainyAPI
  BEAM --> BrainyAPI
```

**Anti-benchmax guard:** `internal/memory/overfit_denylist_test.go` blocks LOCOMO surface forms in product code.

---

## 9. Suggested reading order for external agents

1. [external-agent-assessment-pack.md](./external-agent-assessment-pack.md) — self-contained assessment brief  
2. This graph + [codebase-graph.json](./codebase-graph.json)  
3. [sota-end-to-end-program.md](./sota-end-to-end-program.md) — program of record  
4. [program-execution-status.md](./program-execution-status.md) — latest measured numbers  
5. `internal/memory/service.go` `SearchOpt` + `fusion_v2.go` + `recall.go`  
6. `internal/store/postgres/migrations.go` (v12–v18; evidence v2 + pgvector 768)  
7. One vertical pack: `packs/support/v2/` and `fixtures/vertical/support/`  

**Hazards (honest):** hosted ANN is `vector(768)` after mig 18; HNSW valid on staging; hash/128 residue still needs re-embed. Packs v2 sidecars + support/marketing FSMs load at registry time. `/recall` consumes structured evidence packets with optional hybrid LLM reader (`BRAINY_RECALL_LLM`). Architect PR1–PR7 closed 2026-08-05; recall-contract steps 1–5 landed on `dev` 2026-08-07. **Next:** atomic compiler (R1b) + held-out coverage audit → entities/relation projection. R0/R1a/R1c landed in PR #113 (`21a632b`); local LoCoMo remasure **10/30** is a dip. See [sota-representation-path.md](./sota-representation-path.md).
---

## 10. Non-goals / traps for reviewers

- Do **not** treat eval harness answer prompts as product behavior; prefer `POST /recall`.  
- Do **not** recommend LOCOMO-named regexes or held-out conv tuning (rejected).  
- Graph DB is **gated research**, not current roadmap (ADR-004).  
- Staging Render services track **`dev`**; `main` is the production git line.
