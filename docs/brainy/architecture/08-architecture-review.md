# Architecture Review (Decision Complete)

## MVP Slices
1. Ingestion: normalize events and attach provenance.
2. Consolidation: produce episodes, patterns, and taste signals.
3. Belief Graph: maintain ranked hypotheses and conflict edges.
4. Retrieval Engine: compose taste-aware context bundles.
5. Reflection Engine: update beliefs from outcome deltas.
6. Evaluation Harness: run scenario and competitor-aligned benchmarks.

## Storage Strategy
- Default v1: hybrid in-memory model with persistence interfaces.
- Implementations remain storage-agnostic behind repository interfaces.

## Core Interfaces
- `ingest(event)`
- `consolidate(scope)`
- `retrieve(query)`
- `rank_hypotheses(context)`
- `record_outcome(outcome_event)`
- `run_reflection(job)`
- `explain(decision_id)`

## Non-Functional Requirements
- Every belief update auditable.
- Conflicts explicit and queryable.
- Deterministic benchmark replay support.

## Decision Completeness Check
- No unresolved TBD blockers in MVP scope.
