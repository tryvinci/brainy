# LOCOMO smoke — multi-fact diversification (staging)

**Timestamp:** 2026-07-25  
**Brainy:** staging (`71f0709`)  
**Embeddings:** CF Workers AI bge-base-en-v1.5 (768-d)  
**Entity ranking:** OFF  
**Judge/Answerer:** gpt-oss-120b via CF AI Gateway  
**Ingest:** async · **top_k:** 30  

## Scores

| Metric | Prior (16/30 pin) | **Diversify list retrieval** |
| --- | ---: | ---: |
| Overall | 16/30 (0.533) | **19/30 (0.633)** |
| temporal | 10–11/16 | **13/16** |
| multi-hop | 2–3/10 | 2/10 |
| open-domain | 3/4 | **4/4** |
| search p50 | ~730–890 ms | ~1270 ms |
| search p95 | ~1540–1630 ms | ~1749 ms |

OpMem staging: **12/12** (non-regression).

## Product change

List-shaped queries get a larger MMR-selected subject-content admit set, richer
related-fact seeds, and MMR reorder of ranked results by novel content tokens.
Gated to `looksListQuery` only (not all multi-hop) so preference/recency ranking
stays intact.

## Takeaways

1. **Overall +3/30** vs prior same-pin — mainly temporal + open-domain.
2. **Multi-hop still 2/10** — several GT spans are incomplete or absent as
   extractable text (relationship status, country name, full book titles); list
   questions still under-aggregate even with diversified retrieval.
3. **Latency tradeoff** — larger candidate sets raise p50 (~+300–500 ms). Acceptable
   for accuracy; can cap expansion later if needed.

## Remaining gap

Multi-hop needs better fact extraction at ingest (named entities / attributes) and
stronger multi-span answer synthesis — not more LOCOMO cue lists.
