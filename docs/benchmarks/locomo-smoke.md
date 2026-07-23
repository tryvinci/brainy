# LOCOMO smoke — staging dense embeddings (entity OFF)

**Timestamp:** 2026-07-23T21:29:07Z  
**Brainy:** staging (`b8bd3f00b098`)  
**Judge/Answerer:** `[REDACTED]` (gpt-oss-120b via CF AI Gateway)  
**Embeddings:** `workers-ai/@cf/baai/bge-base-en-v1.5` (768-d) on staging API+worker  
**Entity ranking:** OFF (`BRAINY_ENTITY_RANKING=false`)  
**Ingest:** async  

## Scores

| Metric | Value |
| --- | ---: |
| Overall | **0.433 (13/30)** |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.200 | 10 |
| open-domain | 0.500 | 4 |
| temporal | 0.562 | 16 |

## Comparison

| Config | Overall |
| --- | ---: |
| Local hash baseline | 13/30 |
| Local dense (bge-base) entity OFF | 12–13/30 |
| **Staging CF dense entity OFF** | **13/30** |

Taxonomy: {'correct': 13, 'answer_judge_miss': 11, 'retrieval_miss': 6}.

OpMem on staging after dense wiring: **12/12**.

Dense embeddings on staging are live and non-regressing with entity ranking off.
