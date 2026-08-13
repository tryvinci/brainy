# Competitive gap matrix

**Updated:** 2026-08-11  
**Source:** Competitive architecture review §23, adjudicated in [../external-reviews/2026-08-11-competitive-architecture-verdict.md](../external-reviews/2026-08-11-competitive-architecture-verdict.md)

Update continuously as PRs land. Prefer measured pins over vibes.

| Capability | Mem0 | Graphiti/Zep | Brainy now | Brainy target | Program PR |
| --- | --- | --- | --- | --- | --- |
| Durable conversational facts | strong | strong | strong | strong | — |
| ADD-only conversational history | yes | effectively | mixed (#94 applies ops to all provider extracts) | **yes (policy split)** | PR2 |
| Operational mutation semantics | limited | temporal invalidation | **strong** | **strong** | keep |
| Raw provenance | partial | strong | **strong** | **strong** | keep |
| Typed operational state | weak | relation state | **strong** | **strong** | keep |
| Temporal metadata | strong | strong | partial (events/atoms windows; no memory_type) | **strong** | PR3 |
| Temporal retrieval scoring | strong | strong | weak | **strong** | PR3 |
| BM25 + dense | yes | yes | yes (`fusion_v2`) | yes | PR4 extend |
| Entity retrieval | strong | strong | partial (hub boosts) | **strong** | PR4/PR6 |
| Canonical entities | graph-backed | native | partial (`memory_entity_links` only) | **native Postgres** | PR6 |
| Relation memory | platform graph | **core** | weak | **first-class** | PR7 |
| Multi-hop traversal | entity graph | **native** | early (hop V2) | **typed relation hops** | PR8 |
| Governed answer sufficiency | limited | not core | **strong** | **strong** | keep |
| Evidence proof chain | limited | provenance | **strong** (but hops replace packet) | **context + proof split** | PR5 |
| Vertical packs | no equivalent | ontology | **strong** | **strong** | keep |
| Operational state machines | weak | weak | **strong** | **strong** | keep |
| Context token discipline | strong | strong | moderate (`BudgetTokens`; `MaxEvidenceTokens` unused) | **strong** | PR4 |
| LME publish integrity | n/a | n/a | path proven; aborted publish | **publishable LME-20 0/20** (jobs 4829=4829) | PR1 **done** |
| Assistant-generated memories | strong | strong | partial | **qualified** | PR9 |

## Current conversational pins (honesty)

| Pin | Result |
| --- | ---: |
| Gate 0 staging 1×30 (`9bad898`) | 18/30 · MH 50% · OD 25% |
| Post-cutover staging 1×30 (`1f2f26f`) | 15/30 · MH 50% · OD 25% |
| Post-cutover staging 3×90 (`1f2f26f`) | 33/90 · MH **22.2%** · OD 42.9% |
| OpMem / marketing post-cutover | 13/13 / passed |
| LME-20 | **Publishable 0/20** `/recall` — [pin](../../benchmarks/artifacts/lme20-product-recall-pr1-20260812-pin.md) |
