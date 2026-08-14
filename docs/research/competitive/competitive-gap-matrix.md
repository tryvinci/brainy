# Competitive gap matrix

**Updated:** 2026-08-14  
**Source:** Competitive architecture review §23, adjudicated in [../external-reviews/2026-08-11-competitive-architecture-verdict.md](../external-reviews/2026-08-11-competitive-architecture-verdict.md)

Update continuously as PRs land. Prefer measured pins over vibes.

| Capability | Mem0 | Graphiti/Zep | Brainy now | Brainy target | Program PR |
| --- | --- | --- | --- | --- | --- |
| Durable conversational facts | strong | strong | **mixed** (facts + still-needed episode fallback) | **facts recall-primary; episodes provenance + fallback until compiler coverage** | R1a–R1c |
| ADD-only conversational history | yes | effectively | **policy split** (core append-only; verticals keep #94) | **yes (policy split)** | PR2 **landed** |
| Operational mutation semantics | limited | temporal invalidation | **strong** | **strong** | keep |
| Raw provenance | partial | strong | **strong** | **strong** | keep |
| Typed operational state | weak | relation state | **strong** | **strong** | keep |
| Temporal metadata | strong | strong | **partial+** (`memory_type` metadata + event end-by-id) | **strong** | PR3 **landed on `dev`** |
| Temporal retrieval scoring | strong | strong | **intent → IncludeHistorical + temporal_score** | **strong** | PR3 **landed on `dev`** |
| BM25 + dense | yes | yes | yes (`fusion_v2`) | yes | PR4 extend |
| Entity retrieval | strong | strong | partial (hub boosts) | **strong** | PR4/PR6 |
| Canonical entities | graph-backed | native | partial (`memory_entity_links` only) | **native Postgres identity (IDs, aliases, ranked resolution)** | R2 / PR6 after R1b |
| Relation memory | platform graph | **core** | weak | **projection of entity-valued atomic facts** | R3 / PR7 |
| Multi-hop traversal | entity graph | **native** | early (hop V2) | **entity-ID dependency joins** | R4 / PR8 after R3 |
| Governed answer sufficiency | limited | not core | **strong** | **strong** | keep |
| Evidence proof chain | limited | provenance | **context + proof split on `dev`** | **context + proof split** | PR5 **landed on `dev`** |
| Vertical packs | no equivalent | ontology | **strong** | **strong** | keep |
| Operational state machines | weak | weak | **strong** | **strong** | keep |
| Context token discipline | strong | strong | **MaxEvidenceTokens + pool 30/50/100/200 cap 200** | **strong** | PR4 **landed on `dev`** |
| LME publish integrity | n/a | n/a | path proven; aborted publish | **publishable LME-20 0/20** (jobs 4829=4829) | PR1 **done** |
| Assistant-generated memories | strong | strong | **skip phatic assistant episodes** | **qualified** | PR9 **landed on `dev`** |

## Current conversational pins (honesty)

| Pin | Result |
| --- | ---: |
| Gate 0 staging 1×30 (`9bad898`) | 18/30 · MH 50% · OD 25% |
| Post-cutover staging 1×30 (`1f2f26f`) | 15/30 · MH 50% · OD 25% |
| Post-cutover staging 3×90 (`1f2f26f`) | 33/90 · MH **22.2%** · OD 42.9% |
| OpMem / marketing post-cutover | 13/13 / passed |
| OpMem / marketing PR2 local (`10a31e3`) | 13/13 / passed — [opmem](../../benchmarks/artifacts/opmem-pr2-local-20260813.md) · [marketing](../../benchmarks/artifacts/marketing-pr2-local-20260813.md) |
| Local PR2 LoCoMo 1×30 (`24be5ab`) | **6/30** · MH 4/10 · temporal 1/16 — [pin](../../benchmarks/artifacts/locomo-pr2-dev-1x30-20260813.md) |
| Wave 1 local LoCoMo 1×30 (`a7a5184`) | **14/30** · MH **3/10** · OD 2/4 · temporal **9/16** — [pin](../../benchmarks/artifacts/locomo-wave1-dev-1x30-20260813.md) (hybrid-reader confound vs 6/30; not vs Gate 0) |
| Wave 1 local OpMem / marketing (`a7a5184`) | 13/13 / passed — [opmem](../../benchmarks/artifacts/opmem-wave1-local-20260813.md) · [marketing](../../benchmarks/artifacts/marketing-wave1-local-20260813.md) |
| LME-20 | **Publishable 0/20** `/recall` — [pin](../../benchmarks/artifacts/lme20-product-recall-pr1-20260812-pin.md) |
