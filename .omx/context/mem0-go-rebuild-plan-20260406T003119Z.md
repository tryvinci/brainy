Task statement

Plan an end-to-end replacement of the current Brainy repo into a production-oriented memory system inspired by Mem0, using Go for the core serving path while preserving room for Python-based evals and benchmark tooling.

Desired outcome

- Produce an approved implementation plan, not code changes yet.
- Define whether to fork/port/copy versus reimplement against Mem0 as a benchmark and reference.
- Specify a phased repo replacement strategy, architecture, delivery slices, migration path from the current prototype, and an end-to-end testing plan.
- Prepare artifacts suitable for a later execution handoff via ralph or team mode.

Known facts / evidence

- Current repo is a compact Python prototype with in-memory storage and rule-based engines under `src/brainy/`.
- Existing Brainy code is coherent as an executable design spike but not a production system.
- Current package metadata shows no runtime dependencies and only `pytest` as a dev dependency in `pyproject.toml`.
- Mem0's public repository currently advertises Apache 2.0 licensing.
- Apache 2.0 permits reproduction, modification, derivative works, and redistribution subject to attribution / notice / modification-marking requirements.
- User prefers a Go-first architecture for the core system and is open to repo replacement.

Constraints

- This planning work must not execute destructive repo replacement yet.
- Planning should assume current repo contents may be removed later, but only after execution approval.
- Need a defensible path that differentiates from a straight Mem0 clone.
- Final implementation should reach a point where the system is testable end to end.
- Current repo guidance prefers small, reviewable, reversible steps during execution.
- No assumption that Mem0 internals should be copied wholesale; legal permissibility is distinct from product strategy.

Unknowns / open questions

- Which exact Mem0 capabilities are in-scope for v1 parity versus intentionally omitted?
- Whether to keep any part of the current Brainy prototype as docs, fixtures, or benchmark harness.
- Whether evaluation and benchmark tooling should remain in this repo or split into a sibling workspace later.
- Which storage backend should be the first production backend: Postgres+pgvector, Qdrant, or another combination.
- What degree of API compatibility with Mem0 is desirable for easier benchmarking / migration.

Likely codebase touchpoints

- `src/brainy/` may be removed or archived during execution.
- `tests/` will likely be replaced with new Go and possibly Python eval suites.
- `docs/brainy/` may be archived, selectively retained, or superseded by new architecture docs.
- `README.md`, `pyproject.toml`, and benchmark scripts will need replacement or repositioning.
