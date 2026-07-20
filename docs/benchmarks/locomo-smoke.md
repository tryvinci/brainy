# LOCOMO smoke — `locomo-smoke-entity-gated`

**Timestamp:** 2026-07-20T02:15:13Z  
**Brainy commit:** `6e0d5d4ab25e4b8e71bf384b27e1942126c13620`  
**Judge/Answerer:** `[REDACTED]`  
**Ingest:** async  

## Scores

| Metric | Value |
| --- | ---: |
| Overall | **0.433 (13/30)** |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.200 | 10 |
| open-domain | 0.750 | 4 |
| temporal | 0.500 | 16 |

## Notes

SOTA-inspired **entity linking** (Mem0/Zep/A-MEM) added: entities are extracted and
persisted on every memory for provenance and the planned graph layer. The entity-overlap
**ranking boost** regressed same-pin smoke (distinctive-entity mentions are not answers),
so it is gated off by default (`WithEntityRanking`) — product-first, not benchmax.

Score matches the pre-entity product baseline (**13/30, 43.3%**); entity infra adds no regression.
