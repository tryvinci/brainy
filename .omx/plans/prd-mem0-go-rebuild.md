# PRD: Go-First Memory System Rebuild

## Status

Drafted via `ralplan` deliberate-mode consensus workflow.

## Problem

The current repository is a coherent Python prototype and architecture exploration, but it is not a production memory system. It lacks persistent storage, operational boundaries, credible benchmark parity, and a runtime architecture suited for long-lived service operation. The user wants to replace this repo with a serious memory product inspired by Mem0, while avoiding a blind line-by-line clone.

## Desired Outcome

Build a production-oriented memory system with:

- Go as the core service/runtime language
- Python retained only for evals, benchmark tooling, and experimental adapters where it materially accelerates iteration
- A clean architecture that can support end-to-end automated testing
- Explicit legal hygiene for any Apache 2.0 derivative use
- A repo transition strategy that safely retires the current prototype without losing useful artifacts

## RALPLAN-DR Summary

### Principles

1. Reimplement core behavior, do not perform a literal line-by-line port.
2. Keep the serving path operationally simple: one Go service, one primary database path, one queue/background-job model.
3. Preserve proof through tests and evals before widening scope.
4. Separate product-defining code from benchmark and research tooling.
5. Treat any copied upstream code as isolated, attributable, and replaceable.

### Decision Drivers

1. Long-term maintainability and product differentiation
2. Service reliability and operational simplicity
3. Ability to prove correctness with end-to-end tests and repeatable evals

### Viable Options

#### Option A: Fork Mem0 and incrementally replace internals

Pros:
- Fastest route to partial feature parity
- Reuses upstream design and test assumptions
- Easier short-term benchmark comparison

Cons:
- High architecture debt inheritance
- Harder to claim meaningful product differentiation
- Ongoing merge/upstream-sync burden
- Go migration becomes awkward and piecemeal

#### Option B: Go-first reimplementation using Mem0 as a behavioral and benchmark reference

Pros:
- Clean runtime boundaries
- Stronger ownership of architecture
- Easier to design around production concerns from day 1
- Lets us define explicit compatibility surfaces instead of inheriting all upstream assumptions

Cons:
- Slower initial build than a fork
- Requires deliberate scope control to avoid overbuilding
- More up-front product and API decisions

#### Option C: Continue evolving the current Brainy prototype

Pros:
- Lowest immediate disruption
- Preserves current docs and mental model
- Minimal migration effort

Cons:
- Current repo is too far from production shape
- Prototype abstractions are not a strong foundation for service hardening
- Likely leads to prolonged transitional architecture and dead-end refactors

### Chosen Option

Choose Option B.

### Why Alternatives Are Deprioritized

- Option A is valid only if speed to “something Mem0-like” matters more than architecture control. That is not the stated goal.
- Option C keeps too much prototype DNA alive and is likely to trap execution in endless incremental cleanup instead of a proper product build.

## Product Scope

### V1 In Scope

- Text memory ingestion API
- Normalization, deduplication, and source metadata capture
- Extraction pipeline for factual/profile/preference-style memories, with a deterministic local baseline and optional model-backed enrichments behind provider interfaces
- Persistent storage for memory records and retrieval metadata
- Search and retrieval APIs
- Feedback/update path for correcting or suppressing memories
- Background processing for extraction and consolidation
- Benchmark and eval harness to compare current behavior against target behavior
- Developer ergonomics: local dev, test fixtures, repeatable CI validation

### V1 Explicitly Out of Scope

- Recreating all Brainy “belief graph” / “taste system” concepts as first-class production features in the initial cut
- Multimodal memory
- Full workflow/orchestration cloning of every Mem0 feature
- Advanced multi-tenant admin console
- Large-scale distributed architecture before single-node correctness is proven

## Compatibility Strategy

The system should be Mem0-informed, not Mem0-bound.

- Match behavior where it helps benchmarking and migration.
- Do not preserve upstream internals, naming, or package layout unless doing so clearly reduces risk.
- Define an explicit compatibility surface:
  - ingest request shape
  - retrieval response shape
  - memory update/delete semantics
  - benchmark scenario definitions
- Treat all other internals as greenfield design territory.

Compatibility should be defined at two levels:

- Scenario compatibility: benchmark fixtures and golden outputs should let us compare behavior against a Mem0-like reference.
- Optional API shim compatibility: if migration pressure appears later, add a thin compatibility adapter layer; do not design the internal domain model around upstream endpoint quirks from day 1.

## Architecture Direction

### Core Stack

- Language: Go for core service and workers
- Primary storage: Postgres with `pgvector` as the default first backend
- Queue model: database-backed job table initially; external queue only if pressure proves it necessary
- Eval/benchmark tooling: Python in a separate `evals/` tree
- Provider boundary: model, embedding, and extraction provider code lives behind narrow interfaces so vendor churn does not leak into domain or storage packages

### Preferred First Wedge

Before implementing the full async worker pipeline, prove the narrower synchronous baseline:

- single-tenant or minimally scoped tenant model
- deterministic local extraction only
- ingest -> normalize -> dedupe -> persist -> retrieve in one request path
- no network dependency
- no provider calls required for correctness

This wedge exists to validate the data model, retrieval contract, and repo cutover with the smallest possible surface area. Async jobs, provider-backed enrichments, and broader multi-tenant behavior should be layered on only after this path passes end to end.

### Service Components

1. API layer
   - REST/JSON or gRPC-facing boundary
   - auth stub or simple API key boundary for v1
2. Ingestion service
   - validates requests
   - stores raw memory inputs
   - emits extraction jobs
3. Extraction/consolidation workers
   - normalize inputs
   - dedupe
   - classify/update memory entries
4. Retrieval service
   - filters by user/session/application scope
   - runs vector + metadata retrieval
   - merges and ranks results
5. Feedback/update service
   - corrections
   - deletion/suppression
   - memory freshness/override handling
6. Evals/benchmark harness
   - Python-based runners
   - fixture-driven comparisons
   - golden output assertions

### Persistence Sketch

Start with a small, explicit schema:

- `raw_ingests`
- `memory_records`
- `memory_embeddings`
- `feedback_events`
- `job_queue`

Avoid introducing more relational entities until retrieval correctness and update semantics are proven.

### Recommended Package Topology

- `cmd/api`
- `cmd/worker`
- `internal/api`
- `internal/app`
- `internal/memory`
- `internal/extract`
- `internal/retrieval`
- `internal/store`
- `internal/jobs`
- `internal/observability`
- `pkg/contracts`
- `evals/`
- `fixtures/`
- `docs/`

## Repository Transition Strategy

### Branch and Cutover Rule

- Perform replacement work on a dedicated branch.
- Archive before delete.
- Do not remove the archive directory in the same logical change that boots the new Go module.
- Do not allow the new production path to import archived Python code.

### Keep

- Planning artifacts in `docs/brainy/` that still describe useful product ideas, benchmarks, and open questions
- Any benchmark fixtures worth converting into eval datasets
- The current repo history

### Archive

- `src/brainy/`
- `tests/`
- `pyproject.toml`
- `scripts_run_benchmark.py`

Archive path recommendation:

- `archive/brainy-python-prototype/`

Do not delete these first. Move/archive them in the initial replacement branch so they remain inspectable during implementation.

### Destructive-Change Gate

The current root implementation should not be physically deleted until all of the following are true on the replacement branch:

1. new Go module builds successfully
2. migrations and repository tests pass
3. one ingest -> retrieval end-to-end test passes against persistent storage
4. archived prototype remains recoverable in-tree or by commit reference

### Replace First

1. Top-level `README.md`
2. language/tooling bootstrap
3. service contracts
4. storage schema
5. end-to-end test harness

### Replace Last

- architecture docs that should survive only after the new runtime boundaries are proven
- any transitional compatibility scripts

## Phased Plan

### Phase 0: Scope Lock and Repo Bootstrap

- Create new product brief and architecture overview
- Define v1 API and memory record contract
- Create `.omx/plans/v1-contracts-mem0-go-rebuild.md` and freeze the first thin slice before coding
- Define provider boundary for extraction, embeddings, and reranking
- Define deterministic baseline extraction behavior that can pass tests without network access
- Add Go module, build tooling, lint/test commands
- Add Python `evals/` workspace if retained in same repo
- Create third-party attribution files for any copied Apache 2.0 material
- Define exact “parity targets” versus “intentional deviations” from Mem0 behavior
- Add `docs/mem0-parity-matrix.md` covering scenario parity, response-shape parity, and intentional deviations
- Add `docs/cutover-plan.md` describing archive-first migration and destructive-change gate

Exit criteria:

- repo boots cleanly
- no production code copied yet without attribution plan
- API and storage contracts reviewed
- first-slice contracts explicitly frozen
- parity/deviation matrix committed
- provider boundary documented

### Phase 1: Archive Prototype and Establish Skeleton

- Move current Python prototype under `archive/brainy-python-prototype/`
- Add new Go service skeleton and worker skeleton
- Install CI commands for format, lint, unit test, integration test

Exit criteria:

- old prototype preserved
- new root commands compile
- CI green on empty/skeleton project

### Phase 2: Core Domain and Persistence

- Implement domain models: memory input, normalized memory, memory fact, source, feedback event
- Implement Postgres schema and repository layer
- Add job table for async extraction pipeline, but do not make the first vertical slice depend on it
- Implement provider interfaces with test doubles for embeddings/extraction
- Implement the synchronous deterministic ingest -> persist -> retrieve baseline path

Exit criteria:

- migrations apply cleanly
- repositories pass unit and integration tests
- raw ingest persists successfully
- synchronous baseline ingest -> retrieve flow passes against persistent storage
- provider-backed components are swappable in tests without network dependency

### Phase 3: Ingestion and Extraction Path

- Implement ingest API
- Implement normalization and dedupe rules
- Extend from the synchronous baseline into the worker-based extraction/consolidation path
- Add idempotency semantics
- Keep a deterministic extractor available for CI and local correctness tests; layer provider-backed enrichments on top only after baseline correctness exists

Exit criteria:

- duplicate ingests do not create duplicate logical memory records
- worker path is replay-safe
- fixture-based extraction tests pass

### Phase 4: Retrieval and Feedback Path

- Implement vector + metadata retrieval
- Implement ranking policy
- Implement update/delete/suppress feedback APIs
- Add retrieval trace/explain metadata

Exit criteria:

- retrieval correctness passes golden tests
- feedback updates change subsequent retrieval results
- explainability surface exists for debugging

### Phase 5: Eval and Benchmark Harness

- Build Python eval harness against the Go API
- Encode target scenarios drawn from current Brainy docs and Mem0-inspired use cases
- Add baseline comparison outputs and reproducible local runbook

Exit criteria:

- one-command local eval run
- golden fixtures committed
- benchmark reports reproducible in CI or documented local workflow

### Phase 6: Hardening and Cutover Cleanup

- Add observability, structured logs, metrics, tracing
- Add failure-mode tests
- Remove no-longer-needed transitional files from root
- Update docs and operating runbooks

Exit criteria:

- end-to-end test suite green
- observability dashboards/log fields defined
- repo root reflects only the new system and intentional archives

## Definition of Done

The rebuild is complete only when:

1. The new Go service handles ingest, retrieval, and feedback end to end against persistent storage.
2. The worker path is idempotent under replay and retry.
3. The eval harness can run reproducible benchmark fixtures against the Go API.
4. The repo contains an explicit attribution trail for any copied Apache 2.0 code or assets.
5. The old prototype is archived or removed intentionally, with no production path depending on it.

## Deliberate-Mode Pre-Mortem

### Scenario 1: “Go rebuild stalls because scope silently expands into a novel platform”

Failure mode:
- team tries to outperform Mem0 and preserve Brainy’s belief/taste concepts in v1

Mitigation:
- hold v1 to ingest/store/retrieve/update/eval
- explicitly defer belief-graph-style features until post-v1

### Scenario 2: “License-safe plan becomes code-copy sprawl”

Failure mode:
- upstream Mem0 code gets copied across multiple directories without attribution boundaries or replacement intent

Mitigation:
- require all copied code to live under isolated directories or marked files with retained notices
- maintain `THIRD_PARTY_NOTICES.md`
- prefer behavioral reimplementation over source copying

### Scenario 3: “System works locally but cannot be trusted operationally”

Failure mode:
- happy-path APIs pass, but worker retries, dedupe, and retrieval correctness break under load or replay

Mitigation:
- require idempotency tests, integration fixtures, and observability before declaring readiness

## Expanded Test Plan

### Unit

- request validation
- normalization rules
- dedupe key generation
- ranking policy
- repository serialization/deserialization
- feedback state transitions

### Integration

- migrations against ephemeral Postgres
- ingest -> job enqueue -> worker extraction -> retrieval
- idempotent replay of ingest and worker jobs
- delete/suppress/update propagation to retrieval

### End-to-End

- API ingest and retrieve for a realistic user session
- correction flow where a bad memory is fixed and retrieval changes
- batch ingest followed by deterministic benchmark/eval run

### Observability

- structured logs include request ids, tenant/user ids, and job ids
- metrics exist for ingest latency, worker success/failure, retrieval latency, dedupe hit rate
- failure paths emit machine-checkable error categories

## ADR

### Decision

Rebuild the memory system as a Go-first service, using Mem0 as a behavioral reference and benchmark baseline rather than a codebase to port wholesale.

### Drivers

- Need a production-ready runtime shape
- Need clearer ownership of architecture and IP
- Need a testable path from prototype to product

### Alternatives Considered

- Fork Mem0 and evolve it
- Continue evolving current Brainy prototype

### Why Chosen

This approach balances legal safety, differentiation, maintainability, and operational fitness better than either a fork or continued prototype evolution.

### Consequences

- Higher initial build cost than forking
- Cleaner long-term system boundaries
- Need discipline to prevent scope creep and over-abstraction

### Follow-Ups

- decide exact v1 API surface
- decide first extraction strategy and model/provider boundaries
- decide whether eval tooling stays in-repo or in a sibling repo
- decide whether a Mem0-compatible API shim is needed after v1 or can remain fixture-only

## Execution Handoff Guidance

### If Executing via `ralph`

Use a sequential single-owner build when:

- repo surgery and migration ordering matter more than parallel speed
- you want tight control over archive/move/delete timing

Recommended lane sequence:

1. `architect` high: lock contracts and directory layout
2. `executor` high: bootstrap Go project and archive current prototype
3. `test-engineer` medium: stand up integration/e2e scaffolding
4. `executor` high: implement storage/ingest/worker/retrieval slices
5. `verifier` high: run full validation and repo cleanup

### If Executing via `team`

Use parallel team execution when:

- the service, storage, and eval harness can progress independently after the initial contract lock

Suggested roster:

- `architect` high: contracts, storage, repo transition decisions
- `executor` high: API/service implementation
- `executor` medium: workers/background jobs
- `test-engineer` medium: integration/e2e scaffolding
- `writer` low: docs, migration notes, third-party notice hygiene
- `verifier` high: final evidence pass

Suggested staffing split:

- Lane A: repo bootstrap + archive + Go module
- Lane B: persistence + migrations + repositories
- Lane C: API + retrieval + feedback
- Lane D: eval harness + benchmark fixtures

Suggested launch hint:

- `$team execute the approved mem0-go rebuild plan with lanes for bootstrap, persistence, service path, and eval harness`

### Suggested `ralph` Launch Hint

- `$ralph execute the approved mem0-go rebuild plan sequentially, preserving the prototype until the destructive-change gate passes`

### Team Verification Path

Before completion, require:

1. `go test ./...`
2. integration tests against ephemeral Postgres
3. end-to-end fixture run through API + worker path
4. eval harness baseline report committed or generated reproducibly
5. manual smoke run with logs/metrics visible

## Critic Verdict

APPROVE

Rationale:

- options are real and fairly compared
- chosen path is bounded and testable
- deliberate-mode pre-mortem and expanded test strategy are present
- destructive repo replacement is now explicitly gated
- execution handoff guidance is concrete enough for ralph or team follow-through
