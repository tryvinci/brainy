# Test Spec: Go-First Memory System Rebuild

## Objective

Define the proof required to replace the current Brainy prototype with a Go-first memory system and claim it is functionally usable end to end.

## Release Gates

The rebuild is not complete unless all of the following are true:

1. Core Go services compile and pass unit tests.
2. Database migrations apply cleanly to a fresh database and an upgraded database.
3. `v1-contracts-mem0-go-rebuild.md` is implemented without scope drift in the first thin slice.
4. Duplicate ingests are idempotent.
5. Feedback updates alter subsequent retrieval behavior correctly.
6. A reproducible eval run exists and is documented.
7. Archived prototype files remain accessible until explicit cleanup is approved.

## Minimum Credible Wedge Before Async Workers

Before the plan is allowed to expand into full worker-driven extraction, the following narrower proof must exist:

1. deterministic synchronous ingest -> persist -> retrieve passes against Postgres
2. duplicate ingests are idempotent on the synchronous path
3. retrieval returns stable explain/debug payloads without any network-backed provider
4. this path runs in CI without external model or embedding dependencies

The async worker pipeline is an expansion phase, not the first proof point.

## Test Layers

### Unit Tests

Cover:

- config loading
- contract validation
- normalization
- dedupe key generation
- ranking logic
- repository mapping
- feedback/update rules
- explain/debug payload generation
- provider interface behavior with fake embedding/extraction clients
- deterministic extractor behavior on fixed fixtures without network access

Commands target:

- `go test ./...`

### Integration Tests

Cover:

- Postgres migration apply/rollback
- repository CRUD against real database
- synchronous ingest API persists records and returns retrievable memories without workers
- enqueue/dequeue worker jobs
- ingest API persists raw source records and creates jobs
- worker execution materializes retrievable memories
- suppression/delete/update flows persist and affect reads
- compatibility JSON fixtures for ingest/retrieve/update surfaces
- provider-backed enrichment path can be disabled without breaking baseline ingest -> retrieve flow

Infrastructure:

- ephemeral Postgres container
- test seed fixtures

### End-to-End Tests

Cover:

- one thin-slice baseline where API ingest and retrieve works without the async worker path
- user ingesting profile/context data then retrieving relevant memories
- repeated ingest of equivalent content does not duplicate logical memory
- incorrect memory corrected via feedback path changes later retrieval
- multi-record session behaves deterministically under benchmark fixture
- archived prototype remains non-runnable from the new production path

Execution surface:

- API server + worker process + database running together

### Eval / Benchmark Tests

Cover:

- golden scenarios for profile memory, preference memory, and conversational memory
- retrieval precision/recall thresholds against fixture expectations
- response-shape compatibility where intentional
- regression report generation
- parity-matrix scenarios marked as exact-match, approximate-match, or intentional-deviation

Execution surface:

- Python harness calling the Go service via API

### Observability Tests

Cover:

- structured logs emitted for ingest, worker, retrieval, failure
- metrics increment for success/failure/latency
- trace or correlation ids flow through API and worker path
- error classes are stable enough for alerting
- replay/idempotency counters are exposed for duplicate-ingest and worker-retry paths

## Required Fixtures

- single-user profile memory fixture
- preference/mutation fixture
- duplicate-ingest fixture
- correction/suppression fixture
- noisy mixed-input fixture
- benchmark comparison fixture

## Negative Tests

- invalid payload rejected with stable error code
- worker retry on transient failure does not duplicate writes
- retrieval with missing embeddings degrades safely
- deleted/suppressed memory never leaks in results
- migration mismatch fails fast with actionable error
- provider timeout/failure does not corrupt stored raw ingest state
- compatibility shim, if added, cannot bypass core validation rules

## CI Shape

### Fast Path

- format/lint
- unit tests

### Standard Path

- unit + integration

### Release Path

- unit + integration + e2e + eval smoke + artifact generation

### Optional Hardening Path

- release path plus bounded load/replay test for ingest and worker idempotency

## Completion Evidence

Execution handoff is only considered verified when the final delivery includes:

- exact commands run
- pass/fail status
- generated benchmark/eval artifact locations
- known gaps, if any
