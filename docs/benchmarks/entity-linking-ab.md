# Entity linking / graph propagation — honest A/B

SOTA memory systems (Mem0, Zep/Graphiti, A-MEM, HippoRAG) rely on entity linking
and graph propagation. We implemented both generically and measured same-pin
LOCOMO smoke (1 conversation / 30 Q, async ingest, identical judge/answerer).

| Config | Overall |
| --- | ---: |
| Entity ranking OFF (default) | **13/30 (43.3%)** |
| Entity overlap boost ON | 8/30 |
| Entity + IDF weighting ON | 10/30 |
| Entity graph propagation (rerank-only) ON | 10/30 |

## Conclusion

Naive/graph entity reranking **regresses** conversational recall in our stack:
distinctive-entity mentions are frequently not the answer, and reshuffling displaces
strong lexical/temporal hits. HippoRAG/Mem0 gains depend on **dense semantic
embeddings** feeding the graph — which is currently blocked in this environment
(provider `/embeddings` returns 403).

Decision (product-first, anti-benchmax): keep entity **extraction + persistence**
(provenance + future graph layer) always on; keep entity **ranking** behind
`BRAINY_ENTITY_RANKING=true` until dense embeddings make it non-regressing.

Entities are not tuned to any benchmark — extraction is generic (proper nouns,
quoted spans, years).
