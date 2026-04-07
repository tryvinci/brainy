# Mem0 Parity Matrix

## Reference Pin

- Repository: `https://github.com/mem0ai/mem0`
- Pinned commit: `a670333d67be1207b5be2fc73af60c3439444f48`

## Purpose

This file prevents the rebuild from drifting behind vague "Mem0-inspired" language.
It records which behaviors are treated as parity targets, which are approximate matches, and which are intentional deviations.

## Thin-Slice Targets

| Area | Status | Notes |
| --- | --- | --- |
| Ingest request shape | approximate-match | Thin slice accepts a simplified `messages[]` payload and `source_type`. |
| Search response shape | approximate-match | Response includes deterministic `explain` payloads from the baseline ranker. |
| Duplicate ingest behavior | exact-match target | Identical logical memories must be idempotent. |
| Corrections / suppression | exact-match target | Updated memory state must affect later retrieval deterministically. |
| Async workers | intentional deviation | Deferred until after the synchronous thin slice is green. |
| Embeddings / rerankers | intentional deviation | Deferred until deterministic baseline passes fixtures. |
| Provider-backed extraction | intentional deviation | Deferred until after local extraction and parity fixtures are stable. |

## Fixture Sources To Capture

- public Mem0 examples used for memory ingestion
- public search examples used for retrieval comparisons
- duplicate-ingest scenario
- memory correction or suppression scenario

## Open Follow-Ups

- capture concrete Mem0 example fixtures into `fixtures/parity/`
- record exact upstream files or docs used for each fixture
- update this matrix as the API surface gets closer to or farther from the reference
