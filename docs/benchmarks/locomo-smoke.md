# LOCOMO smoke — multi-fact ranking (staging)

**Timestamp:** 2026-07-24T23:55:00Z  
**Brainy:** staging (`a86fb38` + evals client retry)  
**Embeddings:** CF Workers AI bge-base-en-v1.5 (768-d)  
**Entity ranking:** OFF  
**Judge/Answerer:** gpt-oss-120b via CF AI Gateway  
**Ingest:** async · **top_k:** 30  

## Scores

| Metric | Prior (judge matrix) | **This run** |
| --- | ---: | ---: |
| Overall | 14/30 (0.467) | **16/30 (0.533)** |
| temporal | 9/16 | **11/16** |
| multi-hop | 2/10 | 2/10 |
| open-domain | 3/4 | 3/4 |
| search p50 | ~1027 ms | **~730 ms** |
| search p95 | ~2655 ms | **~1632 ms** |

OpMem staging: **12/12** (non-regression).

## Product changes measured

1. Low-information / name-only penalty — ack turns no longer dominate person queries.
2. Subject-content bridge — content-dense subject mentions admitted without surface-verb overlap.
3. Parallel lexical + dense scoring — lower search latency.
4. Multi-evidence answerer prompt + top_k 30.

## Remaining gap

Multi-hop still 2/10: retrieval now surfaces more supporting facts (e.g. pottery,
transgender journey) but list/completeness synthesis and some missing titles
(e.g. book names not present as extractable strings) still fail the judge.
Next product levers: list-aggregation at answer time from multi-memory spans,
temporal supersession, and (later) entity-rank re-tune off this smoke.
