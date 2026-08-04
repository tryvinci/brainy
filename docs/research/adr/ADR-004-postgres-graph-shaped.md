# ADR-004 Postgres graph-shaped model

## Decision
Do not introduce a dedicated graph database in Phase 1–5. Model entities/relations/events in Postgres.

## Status
Accepted. Revisit only under measured traversal bottlenecks (program §8.2).
