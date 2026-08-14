# Implementation borrow log

Track BORROW / ADAPT / REJECT decisions before inventing Brainy mechanisms.

| Date | Gap | Competitor | Class | Why | Brainy PR / artifact |
| --- | --- | --- | --- | --- | --- |
| 2026-08-11 | Program frame | Mem0 + Graphiti public docs/OSS | **ADAPT** | Combine high-recall conversational capture + relation memory with Brainy evidence/ops/verticals; do not clone either product | Verdict + this program |
| 2026-08-11 | ADD-only conversational history | Mem0 migration / OSS extract | **ADAPT** | Keep #94 ops for governed predicates; split conversational append-only | PR2 (queued) |
| 2026-08-12 | Conversational vs governed write policy | Mem0 ADD-only + Brainy #94 | **ADAPT** | `WriteMutationModeOf`: core/empty → append-only merge/persist; non-core vertical → #94 NONE/DELETE/UPDATE | PR2 (code on #100) |
| 2026-08-11 | Temporal ranking features | Mem0 temporal reasoning stage | **ADAPT** | Reuse mig-16 event windows; add ranking signal — not duplicate schema | PR3 (queued) |
| 2026-08-11 | Broad multi-signal candidates | Mem0 multi-signal + top-200 docs | **ADAPT** | Extend fusion_v2; fixed context token budget; no bare top-k inflation | PR4 (queued) |
| 2026-08-11 | Relation-native multi-hop | Graphiti entities/edges/BFS recipes | **ADAPT** | Postgres relations + hop V3; **REJECT** Neo4j default | PR7–PR8 (queued) |
| 2026-08-14 | Fact-primary vs transcript index | Mem0 ADD facts + Graphiti episode/entity split | **ADAPT** | Default search drops `conversation_episode` when facts exist; episodes remain provenance. Do not copy Mem0 prompts. | R1 / this PR |
| 2026-08-14 | Attribute atoms tagged episode | (bug) | **REJECT** as design | Atoms and dated provider facts are recall-primary facts | R1 |

## How to append

1. Fill one row per mechanism before coding.
2. Link competitor paths/commits inspected.
3. Link Brainy PR once opened.
4. Update [competitive-gap-matrix.md](./competitive-gap-matrix.md) when status changes.
