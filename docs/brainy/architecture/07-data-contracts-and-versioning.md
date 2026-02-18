# Data Contracts and Versioning

## Canonical Types
- `MemoryIngestEvent`
- `MemoryArtifact`
- `BeliefNode`
- `OutcomeEvent`
- `ReflectionJob`
- `ContextQuery`
- `ContextBundle`
- `HypothesisRecord`

## Contract Strategy
- Version every top-level contract with `schema_version`.
- Backward-compatible additions allowed in minor versions.
- Breaking field changes require major version increment.
- All events carry provenance metadata and timestamps.

## Compatibility Rules
- Unknown fields must be ignored by readers.
- Deprecated fields retain support for one major cycle.
- Contract test fixtures must be updated on each version change.
