# Mem0 — competitive notes (stub)

**Updated:** 2026-08-11  
**Classification guidance:** see [README.md](./README.md) · [implementation-borrow-log.md](./implementation-borrow-log.md)

## What to treat as true

- Mem0 OSS is a useful blueprint (Apache-2.0).
- Published benchmark scores often reflect the **managed platform**, including proprietary optimizations not in OSS.
- Documented strengths relevant to Brainy: ADD-only conversational extraction, hybrid multi-signal retrieval, entity matching in ranking, temporal features at construction + query-time ranking, broader candidate budgets with bounded final context.

## Brainy borrow stance

| Mechanism | Stance | Program PR |
| --- | --- | --- |
| ADD-only conversational facts + retain history | **ADAPT** (keep governed ops for operational/vertical) | PR2 |
| Temporal metadata + temporal ranking signal | **ADAPT** (reuse Brainy event/atom windows) | PR3 |
| Dense + BM25 + entity scoring | **ADAPT** (extend `fusion_v2`) | PR4 |
| Large candidate pool / fixed context tokens | **ADAPT** (explicit budgets; no blind top-k inflate) | PR4 |
| Abandon governed current_state | **REJECT** | — |

## Inspect before coding (checklist)

- [ ] Current OSS extract / ADD-only migration docs  
- [ ] Retrieval scoring (semantic + BM25 + entity)  
- [ ] Temporal feature fields and query-time use  
- [ ] Entity extraction → retrieval boost path  
- [ ] Same-pin harness notes for fair compares  

## Local Brainy counterparts

- Provider ops / merge: `internal/memory/provider_extractor.go`, `memory_ops.go`  
- Fusion: `internal/memory/fusion_v2.go`  
- Temporal resolver (as_of / current): `internal/memory/temporal.go`  
- Entity hub: `memory_entity_links` / `entity_hub.go`
