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
| Async workers | implemented | Sync + async ingest, worker loop, DLQ/retries on `dev`. |
| Embeddings / rerankers | intentional deviation | Deferred until deterministic baseline passes fixtures. |
| Provider-backed extraction | intentional deviation | Deferred until after local extraction and parity fixtures are stable. |
| Vertical memory packs | intentional deviation | Mem0 has no pack model. Brainy uses cognitive primitives + YAML packs (`packs/`). First pack: marketing v1. |
| Cognitive primitive ranking | intentional deviation | Principle > IdentityPrior > Belief precedence; not in Mem0 reference. |

## Current Local Fixtures

- `fixtures/parity/dark_mode_vim_preference.json`
- `fixtures/parity/response_style_preference.json`
- `fixtures/parity/profile_lookup.json`
- `fixtures/parity/factual_context.json`

## Fixture Provenance

| Fixture | Provenance | Notes |
| --- | --- | --- |
| `dark_mode_vim_preference.json` | Directly derived from the pinned Mem0 README CLI example `mem0 add "Prefers dark mode and vim keybindings" --user-id alice` at commit `a670333d67be1207b5be2fc73af60c3439444f48`. | Direct parity reference for CLI-style preference capture and search. |
| `response_style_preference.json` | Derived from the pinned Mem0 README examples at commit `a670333d67be1207b5be2fc73af60c3439444f48`: the CLI memory-add example `Prefers dark mode and vim keybindings` and the basic usage `memory.search(query=message, user_id=user_id)` flow. | This is the current direct parity reference fixture. |
| `profile_lookup.json` | Synthetic baseline fixture | Used to verify deterministic profile retrieval behavior; not claimed as a direct Mem0 example match. |
| `factual_context.json` | Synthetic baseline fixture | Used to verify deterministic factual retrieval behavior; not claimed as a direct Mem0 example match. |

## Open Follow-Ups

- capture additional public Mem0 examples into `fixtures/parity/` as the API surface widens
- update this matrix as the API surface gets closer to or farther from the reference
- track vertical pack evals separately under `fixtures/vertical/` (not Mem0 parity)
