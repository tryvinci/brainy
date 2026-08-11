# External review verdict — Competitive architecture program (2026-08-11)

**Date:** 2026-08-11  
**Source:** External competitive architecture review (Mem0/Graphiti parity path)  
**Adjudicator:** Coding agent (codebase-verified)  
**Prompt used:** [2026-08-11-hardening-self-review-prompt.md](./2026-08-11-hardening-self-review-prompt.md) + competitive architecture review brief

## Verdict (1 paragraph)

**Keep V3 architecture and hardening direction; adjust the next program.** Stop framing work as incremental fixes to the typed multi-hop pipeline alone. Adopt the competitive program: reach feature parity with publicly inspectable Mem0/Graphiti mechanisms inside Brainy's five-plane architecture, then combine them with Brainy's stronger evidence, temporal truth, operational memory, and vertical governance. Principle: **recall broadly → represent explicitly → prove narrowly → answer truthfully.** Single most important next move: close LME-20 measurement integrity (PR1), then execute the accepted PR2–PR10 sequence.

## Answers to prompt questions

1. **Course:** KEEP V3 harden outcomes; **adjust** next sequence toward competitor-parity mechanisms (ADD-only conversational history, temporal ranking features, multi-signal candidates, context/proof split, canonical entities, relation memory, relation hops) rather than more hop heuristics.
2. **Primary gap (MH):** Residual larger-slice MH (~22.2% on post-cutover 3×90) is primarily **representation + retrieval coverage** (no relation-native memory; hop resolve = first linked memory ID; packet replaces broad evidence with hop-selected items), not reader-only. Temporal 1×30 dip (75% → 56%) also shows retrieval/ranking gaps.
3. **Next 3–7 PRs (opening):** PR1 LME integrity → PR2 conversational vs governed write policy → PR3 temporal features V1 → PR4 retrieval V4 budgets → PR5 context vs proof → then PR6–PR8 entity/relation/hop V3.
4. **Claims allowed vs forbidden:** Allowed: post-cutover 15/30 and 33/90 with honesty; OpMem/marketing non-reg; `/recall` path proven. Forbidden: beats Mem0; SOTA; MH solved; publishable LME before clean full `--publish --product-recall`; calling 1×30 an improvement vs Gate 0.
5. **Kill list confirmation:** Fusion retune, graph DB default, category dictionaries, top-k inflation without fixed context budget, reader-only primary, LoCoMo/LME-named product rules, SOTA language — all remain rejected.

## Findings table

| Finding | Accept / Modify / Reject | Code evidence | Action |
| --- | --- | --- | --- |
| KEEP V3 harden; do not restart architecture | **Accept** | PRs #93–#98 merged; tips `308d3a1` / `1f2f26f` | Preserve; do not reopen architect PR1–PR7 |
| Next program = competitive parity then exceed | **Accept** | Post-cutover MH 22.2% on 3×90; no relation tables | Adopt as PoR; create `docs/research/competitive/` |
| Recall broadly / prove narrowly | **Accept** | `bindPacketFromHopResults` replaces packet with hop evidence (`recall.go`) | PR5 context vs proof split |
| Conversational ADD-only vs governed ops split | **Accept** | `mergeProviderAndBaseline` applies NONE/DELETE/UPDATE to all provider extracts (`provider_extractor.go`) | PR2 memory-mode policy (not rollback of #94) |
| Temporal ranking features missing | **Modify** | `memory_events.starts_at/ends_at/time_precision` (mig 16); atoms `valid_from/valid_to`; no `temporal_score` fusion signal | PR3 reuse mig-16; add `memory_type` + query intent + ranking signal — do not duplicate schema |
| Multi-signal retrieval / candidate budgets | **Modify** | `fusion_v2.go` already fuses semantic+BM25+entity with `signal_*` explain; `CandidateOverfetch` cap 200; `BudgetTokens` exists; `MaxEvidenceTokens` unused | PR4 extend (explicit `candidate_limit`, temporal signal, qualify 30/50/100/200 at fixed context tokens) — not greenfield |
| Canonical entities missing | **Accept** | Only `memory_entity_links` hub; `hop_executor.resolveEntityHop` = first linked memory ID | PR6 entities + aliases (mig 20+) |
| Relation-native memory missing | **Accept** | No SPO edge table; atoms keyed by conversation `subject_id` | PR7 relations + relation_evidence |
| Hop executor needs relation traversal | **Accept** | `follow_relation` shares resolve path; no edge walk | PR8 hop V3 with `hop[i].output == hop[i+1].input` |
| LME-20 integrity before architecture | **Modify** | `extraction_jobs.failure_reason` + retries exist; harness error lists IDs only (`backends/brainy.py`) | PR1 surface failure_reason + job accounting; isolated rerun — no new `failure_stage` column initially |
| Graph DB required | **Reject** | Prior architect reject; Graphiti lesson = graph *semantics* | Implement relations in Postgres |
| Fusion retune / category dicts / top-k inflation | **Reject** | Standing kill list | Keep kill list |
| Reader-only as next default | **Reject** | Hardening adjudication | Prefer capture/retrieve/represent |

## Accepted next sequence

1. **PR1** `MEASUREMENT / WRITE_PIPELINE` — LME harness failure_reason + jobs_expected/completed/failed; isolated LME-20 `--publish --product-recall` pin  
2. **PR2** `REPRESENTATION_MISS` — conversational append-only vs governed mutation policy  
3. **PR3** `TEMPORAL_RETRIEVAL_MISS` — memory_type + temporal_score ranking (reuse mig-16)  
4. **PR4** `RETRIEVAL_MISS` — candidate/context/proof budgets; fixed-token candidate matrix  
5. **PR5** `EVIDENCE_COVERAGE_MISS` — EvidencePacket → ContextEvidence + ProofChain  
6. **PR6** `ENTITY_RESOLUTION_MISS` — canonical entity store V2  
7. **PR7** `MULTIHOP_REPRESENTATION_MISS` — relation memory V1  
8. **PR8** `MULTIHOP_PLANNING_MISS` — relation-aware hop executor V3  
9. **PR9** `EXTRACTION_COVERAGE_MISS` — assistant-generated memory qualification  
10. **PR10** frozen competitive qualification — Brainy vs fresh Mem0 same-pin + multi-seed LoCoMo + LME-20/100  

## Rejected / deferred

- Graph DB / Neo4j as default  
- Blind copy of Mem0 managed-platform tricks or Graphiti proprietary Context Graph Engine  
- Expanding hop heuristics before capture/retrieve/represent parity  
- LME-100 before clean LME-20  
- SOTA / beats-Mem0 claims  

## Claims discipline check

- [x] No invented LME / SOTA scores  
- [x] Gate 0 vs harden / post-cutover pins not blended  
- [x] MH% cited with matching pin (1×30 50% vs 3×90 22.2%)  
- [x] Dip honesty preserved (1×30 18→15; temporal 75→56.2)  

## Artifact diffs required

- [x] assessment pack (implication + priority)  
- [ ] codebase graph md/json (defer until PR6/PR7 schema land)  
- [x] program-execution-status  
- [x] PoR adoption note (`docs/research/competitive/`)  
- [x] external-reviews/README.md priority pointer  

## Linked PRs / commits

- Hardening closed: #93–#98 · production `308d3a1` · staging `1f2f26f`  
- Docs PR: #99 lineage / competitive adoption commits on this branch  
- Competitive program start: PR1 LME integrity (code follow-up)
