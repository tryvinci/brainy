# ADR-002 Bitemporal semantics

## Decision
Stateful atoms/assertions carry world-valid and system-valid intervals (`valid_from`/`valid_to`, `recorded_at`/`retired_at`). Predicate policy replaces universal latest-wins.

## Status
Accepted; atoms migration 15 shipped; full assertion API reads follow.
