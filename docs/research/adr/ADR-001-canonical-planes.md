# ADR-001 Canonical evidence and memory planes

## Decision
Adopt five planes (source → evidence → semantic → projection → recall) on Postgres.

## Status
Accepted (program of record 2026-08).

## Consequences
New tables land incrementally; `memory_records` remains compatible during shadow writes.
