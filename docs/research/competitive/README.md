# Competitive archaeology — standing SOP

Standing process for inspecting Mem0 / Graphiti (and peers) before inventing Brainy memory mechanisms.

**Program of record (execution):** [sota-representation-path.md](./sota-representation-path.md) — compile interactions into facts/entities/relations; retrieve those; keep episodes as provenance + fallback. Wave 1 ranking PRs are not the SOTA bet. **Mem0 OSS ≠ Mem0 Platform; Graphiti ≠ Zep Platform.**  
**Gap matrix:** [competitive-gap-matrix.md](./competitive-gap-matrix.md)  
**Borrow log:** [implementation-borrow-log.md](./implementation-borrow-log.md)

## Principle

> Before inventing a new memory mechanism, inspect current Mem0 and Graphiti OSS implementations for the same problem. Reuse or adapt proven OSS ideas when compatible with Brainy's five-plane architecture. Deviate only when Brainy's evidence, operational, temporal, or vertical requirements provide a concrete reason.

Competitor repos are **reference implementations / architectural blueprints**, not exact copies of managed platforms that publish benchmark scores.

Track four surfaces separately:

| Track | Use |
| --- | --- |
| Mem0 OSS | Mechanisms we can inspect and reproduce |
| Mem0 Platform | Product quality we need to match (published numbers include proprietary opts) |
| Graphiti | Entity / relation / episode / validity architecture we can inspect |
| Zep Platform | What that architecture becomes in production |

## Per-gap workflow

1. **Identify** the competitor mechanism for the gap.
2. **Inspect** current code, tests, prompts, harnesses (link commits/paths in the borrow log).
3. **Classify:** `BORROW` | `ADAPT` | `REJECT`.
4. **Document why** in [implementation-borrow-log.md](./implementation-borrow-log.md).
5. **Implement** a Brainy-native version (Postgres-first; no graph DB default).
6. **Measure** with pins; update the gap matrix.

## Source notes

| System | Notes | Local doc |
| --- | --- | --- |
| Mem0 | OSS Apache-2.0; published scores often reflect **managed** platform | [mem0.md](./mem0.md) |
| Graphiti / Zep | Graphiti = open temporal graph engine; production Zep may use proprietary Context Graph Engine | [graphiti.md](./graphiti.md) |
| Letta | Optional later peer | [letta.md](./letta.md) |

## Do not copy

- Benchmark-specific prompts / category dictionaries  
- Undocumented constants “because the bench works”  
- Huge retrieval budgets without fixed context-token accounting  
- Graph DB migration without measured need  
- Abandoning Brainy governed current-state / OpMem / vertical packs  

## Related kill list

See [../external-reviews/2026-08-11-competitive-architecture-verdict.md](../external-reviews/2026-08-11-competitive-architecture-verdict.md) and [../external-reviews/README.md](../external-reviews/README.md).
