# External review — recall contract / SOTA diagnosis

**Date:** 2026-08-07  
**Source:** ChatGPT Plus (external technical review of assessment pack + codebase)  
**Adjudicator:** Cursor agent (spot-checked against code; accepted with compression)

## Verdict (1 paragraph)

Approve keeping the five-plane architecture and closed architect PR1–PR7. Reject the prior default next-step of “reader quality over packets alone.” The real bottleneck is the end-to-end **question → plan → evidence packet → sufficiency → reader → judge** contract, plus **write-time** memory construction (context-aware extract, provenance, entity-scoped state). Stage oracles that only check non-empty collections over-label `READER_MISS`. Mem0’s benchmark lead is partly harness/config, but also real: retrieve-before-extract, recent context, attribution, and link/update/contradict semantics that Brainy lacks.

## Findings table

| Finding | Accept / Modify / Reject | Code evidence | Action |
| --- | --- | --- | --- |
| Oracles over-label READER_MISS | **Accept** | `evals/public/stage_oracle.py` falls through to READER_MISS when stages non-empty | Oracle/judge hardening (Step 1) |
| `/recall` is packet compiler, not hybrid reader | **Accept** | `internal/memory/recall.go` `reader_source=evidence_packet` | Hybrid reader (Step 5) |
| Evidence write ignores errors; `occurredAt=nil` | **Accept** | `evidence_plane.go` `persistRawEvidence` | Provenance (Step 2) |
| Provider extract sees only new batch | **Accept** | `provider_extractor.go` `buildProviderUserPrompt` | Contextual compile (Step 3) |
| current_state keyed subject+predicate only | **Accept** | `memory_current_state` PK | Entity-scoped keys (Step 4) |
| LoCoMo flattens all roles to user | **Accept** | `locomo/run_smoke.py` ingest | Role preservation (Step 1) |
| Judge parse → WRONG | **Accept** | `judge.py` llm_judge | JUDGE_MISS + retry (Step 1) |
| Async wait = search settle only | **Accept** | `backends/brainy.py` | Job-status wait (Step 1) |
| Graph DB required | **Reject** | — | Stay on Postgres |
| Fusion retune / top-k inflation | **Reject** | — | Do not schedule |
| Full gold labels for every question first | **Modify** | — | Stratified sample + oracle tighten, not full-suite annotation before product work |
| 10-PR mega sequence as one cycle | **Modify** | — | Compress to 6 executed steps below |

## Accepted next sequence

1. Measurement honesty (judge, jobs wait, roles, oracles)
2. Evidence ↔ semantics provenance
3. Context-aware semantic compile
4. Entity-scoped state keys (minimal ER)
5. Plan → packet → sufficiency → hybrid reader
6. LME / LoCoMo / Mem0 proof pins

## Rejected / deferred

- Graph database migration
- Fusion constant retunes from 30-Q smokes
- Third shallow vertical before support/marketing depth
- Unrestricted query-time agent loops
- Benchmark-shaped regex extractors
- Claiming conversational SOTA before same-pin + multi-seed gates

## Artifact diffs required

- [x] assessment pack
- [x] codebase graph md/json (status notes)
- [x] program-execution-status
- [x] external-reviews README priority

## Linked PRs / commits

- https://github.com/tryvinci/brainy/pull/88 — recall-contract sequence (merged to `dev`)
- Related docs: #85 competitive positioning, #87 AGENTS.md identity note
