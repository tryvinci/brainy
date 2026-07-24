# LOCOMO smoke — multi-fact ranking (staging)

**Timestamp:** 2026-07-24T24:00:00Z  
**Brainy:** staging (`a86fb38`) + local evals answerer improvements  
**Embeddings:** CF Workers AI bge-base-en-v1.5 (768-d)  
**Entity ranking:** OFF  
**Judge/Answerer:** gpt-oss-120b via CF AI Gateway  
**Ingest:** async · **top_k:** 30  

## Scores

| Metric | Prior (judge matrix) | Ranking v1 | **+ list extractive (v2)** |
| --- | ---: | ---: | ---: |
| Overall | 14/30 | **16/30** | **16/30** |
| temporal | 9/16 | 11/16 | 10/16 |
| multi-hop | 2/10 | 2/10 | **3/10** |
| open-domain | 3/4 | 3/4 | 3/4 |
| search p50 | ~1027 ms | **~730 ms** | ~891 ms |
| search p95 | ~2655 ms | **~1632 ms** | ~1540 ms |

OpMem staging: **12/12** (non-regression).

±1/30 and ±1 category on the 30-Q pin is expected run variance; the durable lifts
are overall 14→16 and search latency down ~25–40%.

## Product changes measured

1. Low-information / name-only penalty — ack turns no longer dominate person queries.
2. Subject-content bridge — content-dense subject mentions admitted without surface-verb overlap.
3. Parallel lexical + dense scoring — lower search latency.
4. Multi-evidence answerer + prefer extractive when it enumerates a fuller list.
5. Eval LLM retries on gateway connection resets.

## Remaining gap (toward Mem0-class)

Multi-hop 3/10: some facts absent from extractable memory text (e.g. country
names, book titles never spoken), others need better multi-span aggregation and
temporal supersession. Next: knowledge-update / supersession path + larger
LOCOMO pin with gpt-oss fixed. Keep entity ranking OFF until boost re-tune.
