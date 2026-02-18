# Brainy

Brainy is a memory system designed to emulate how high-performing humans learn, form taste, and adapt beliefs over time.

It is built as a general cognitive-memory architecture, with marketing and branding intelligence as the first proving domain.

## Vision

Most memory systems optimize storage and retrieval quality. Brainy adds a higher-order layer:
- stable principles and identity priors
- ranked beliefs and conviction
- conflict reconciliation
- outcome-driven reflection and belief revision
- explainability and governance checkpoints

The goal is not only to remember more, but to reason better over time.

## Current Status

The implementation covers:
- full planning and architecture specs in `docs/brainy/`
- competitor reverse-engineering for 6 direct players
- core sprint implementation (A through F)
- benchmark harness with public and cognitive tracks
- rollout thresholds and quarterly moat-review artifacts

## System Architecture

Core engines:
- `IngestionEngine`: normalization and provenance capture
- `ConsolidationEngine`: episode/fact/principle/taste-signal extraction
- `BeliefGraphEngine`: hypothesis graph build, conflict detection, reconciliation triggers
- `RetrievalEngine`: taste-aware context assembly
- `ReflectionEngine`: conviction stop-loss and belief challenge transitions
- `GovernanceEngine`: checkpoints, rollback, audit, explainability

Public service contract (`BrainyService`):
- `ingest(...)`
- `consolidate(scope)`
- `retrieve(query)`
- `rank_hypotheses(context)`
- `record_outcome(...)`
- `run_reflection(job)`
- `explain(decision_id)`

## Repository Map

- `src/brainy/`: core system implementation
- `tests/`: scenario and contract tests
- `docs/brainy/00-context-capture.md`: captured problem framing
- `docs/brainy/01-program-plan.md`: phase plan and completion status
- `docs/brainy/architecture/`: cognition model, data contracts, review artifacts
- `docs/brainy/competitors/`: evidence packets, capability matrix, attack-surface map
- `docs/brainy/diagrams/`: system diagrams and architecture SVG
- `docs/brainy/benchmarks/`: latest benchmark reports and thresholds

## Competitor Scope (Direct Players)

- Mem0
- Supermemory
- Zep / Graphiti
- Letta
- Memobase
- Cognee

Competitor claims are explicitly tagged as `evidence` or `inference` in the docs.

## Quick Start

### 1. Install

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -e .[dev]
```

### 2. Run tests

```bash
python3 -m pytest -q
```

### 3. Run benchmarks

```bash
python3 scripts_run_benchmark.py
```

Benchmark outputs:
- `docs/brainy/benchmarks/latest-report.json`
- `docs/brainy/benchmarks/latest-report.md`

## Environment Variables

Populate API keys to enable competitor adapter execution in benchmark runs:
- `MEM0_API_KEY`
- `SUPERMEMORY_API_KEY`
- `ZEP_API_KEY`
- `LETTA_API_KEY`
- `MEMOBASE_API_KEY`
- `COGNEE_API_KEY`

Reference file:
- `.env.example`

## Testing Coverage

Implemented test coverage includes:
- ingestion normalization and provenance
- consolidation extraction behavior
- belief graph conflict handling
- retrieval and ranking behavior
- reflection stop-loss behavior
- governance checkpoint/rollback/explain flows
- end-to-end service contract
- benchmark runner output structure

## Program Artifacts

Planning and architecture artifacts are kept first-class in this repo:
- cognitive questions backlog for unresolved deep research threads
- hypothesis ledger schema and seed hypotheses with falsification links
- regression thresholds and rollout gates
- quarterly moat review template

## Notes on "State of the Art"

This repository intentionally avoids unsupported SOTA claims.
A SOTA claim should only be made when:
- competitor adapters are fully implemented
- reproducible side-by-side runs are archived
- benchmark methodology and caveats are published

## Next Implementation Priorities

1. Implement live competitor adapters (not stubs) for benchmark parity.
2. Add persistence backends behind repository interfaces.
3. Expand cognitive benchmarks for regime-shift and long-horizon identity coherence.
4. Add CI automation for benchmark regression gates.
