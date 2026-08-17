# Competitive gap matrix

**Updated:** 2026-08-17  
**Source:** Competitive architecture review §23, adjudicated in [../external-reviews/2026-08-11-competitive-architecture-verdict.md](../external-reviews/2026-08-11-competitive-architecture-verdict.md). Live pins: 2026-08-15 remasure. Dip diagnosis: [locomo-full-recall-dip-why-20260817.md](../../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md).

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
| Canonical entities | graph-backed | native | **string keys** (`entities.go` quoted/proper-noun/year + `memory_entity_links` hub; hops still `res.Value = mention`) | **native Postgres identity (IDs, aliases, ranked resolution)** | R2 first slice landed; **R7** Canonical Entity V2 |
| Relation memory | platform graph | **core** | **v1 string edges** (`memory_relations` `SrcEntity`/`DstEntity`; mig v20) | **canonical-ID edges with validity + spans** (Graphiti semantics, Postgres) | R3 first slice landed; **R8** Relation V2 |
| Multi-hop traversal | entity graph | **native** | **1x30 MH 10/10**; full-suite MH 21/282 (7.4%); unscoped current-state fallback; empty entity-filter keeps all predicate hits | **canonical-ID dependency joins**; unscoped/fuzzy cannot be `typed_exact` proof | R4 landed (measurement); **R9** Hop Executor V3; do not claim MH-solved |
| Governed answer sufficiency | limited | not core | **strong** | **strong** | keep |
| Evidence proof chain | limited | provenance | **context + proof split on `dev`** | **context + proof split** | PR5 **landed on `dev`** |
| Vertical packs | no equivalent | ontology | **strong** | **strong** | keep |
| Operational state machines | weak | weak | **strong** | **strong** | keep |
| Context token discipline | strong | strong | **MaxEvidenceTokens + pool 30/50/100/200 cap 200** | **strong** | PR4 **landed on `dev`** |
| LME publish integrity | n/a | n/a | **4/20** `/recall` (jobs 4829=4829); LME-500 not run | **quality LME after answer-path + representation** | PR1 integrity done; quality still open |
| Product `/recall` synthesis | LLM-over-search | LLM-over-search | **firstStatementFromPacket / enumerate / abstain** (full LoCoMo 11.4%) | **cite compiled facts; R5A structured-first** (OD 0/4 is a diagnostic) | R5A then R5B typed packet |
| Assistant-generated memories | strong | strong | **skip phatic assistant episodes** | **qualified** | PR9 **landed on `dev`** |

## Current conversational pins (honesty)

| Pin | Result |
| --- | ---: |
| **Live remasure (`1b5ab3e`, 2026-08-15)** | OpMem **13/13**; marketing **17/17**; LoCoMo 1×30 **21/30** (MH 10/10, OD **0/4**, temporal 11/16) vs Mem0 **11/30**; full `/recall` **175/1540 (11.4%)**; LME-20 **4/20**; BEAM 100K **8/20** |
| Full LoCoMo path label | 11.4% is product `/recall`. July **49.8%** was search+harness on an older stack (**not** a current-SHA ceiling). Mem0 **92.5%** is n=1540 top-k 200 LLM-over-search — not this pin. [why](../../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md) · [verdict](../external-reviews/2026-08-17-parity-gap-verdict.md) |
| Gate 0 staging 1×30 (`9bad898`) | historical 18/30 · MH 50% · OD 25% |
| Post-cutover staging 1×30 (`1f2f26f`) | historical 15/30 · MH 50% · OD 25% |
| Post-cutover staging 3×90 (`1f2f26f`) | historical 33/90 · MH **22.2%** · OD 42.9% |
| LME-20 integrity (2026-08-12) | **0/20** `/recall` — superseded as quality by **4/20** |
