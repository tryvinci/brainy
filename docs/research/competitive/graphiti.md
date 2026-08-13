# Graphiti / Zep — competitive notes (stub)

**Updated:** 2026-08-11  
**Classification guidance:** see [README.md](./README.md) · [implementation-borrow-log.md](./implementation-borrow-log.md)

## What to treat as true

- Graphiti is the open temporal graph engine associated with Zep's public architecture story.
- Production Zep may use a proprietary Context Graph Engine for large-scale retrieval — do not assume Graphiti == managed Zep scores.
- Documented strengths: episodes as provenance, entities as nodes, facts/relations as edges, validity windows, relation-aware search recipes (BM25, vector, BFS, RRF, MMR, etc.).

## Brainy borrow stance

| Mechanism | Stance | Program PR |
| --- | --- | --- |
| Episode / immutable source + provenance | **ADAPT** (Brainy evidence plane already strong) | keep / deepen |
| Entity + alias canonicalization | **ADAPT** (Postgres) | PR6 |
| Relation edges with validity windows | **ADAPT** (Postgres `relations`) | PR7 |
| Relation traversal for multi-hop | **ADAPT** (hop executor V3) | PR8 |
| Neo4j / FalkorDB as required substrate | **REJECT** | — |

## Inspect before coding (checklist)

- [ ] Entity / edge / episode models in Graphiti OSS  
- [ ] Validity window handling  
- [ ] Search recipe composition (without copying DB)  
- [ ] How multi-hop questions map to edge walks  

## Local Brainy counterparts

- Evidence / raw ingest: `memory_evidence`, `raw_ingests`  
- Entity hub only today: `memory_entity_links`  
- Hop executor: `internal/memory/hop_executor.go`  
- Events: `memory_events` (+ participants)
