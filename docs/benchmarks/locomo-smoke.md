# LOCOMO smoke — `locomo-smoke-content-bearing`

**Timestamp:** 2026-07-19T13:38:55Z  
**Brainy:** `http://127.0.0.1:8080` (commit `ca4d6a7cdb154af32db797488b46e245bcd0c8f0`)  
**Dataset SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Judge/Answerer:** `[REDACTED]`  
**Ingest:** async  

## Scores

| Metric | Value |
| --- | ---: |
| Overall | 0.267 (8/30) |
| Search p95 ms | 25.3 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.100 | 10 |
| open-domain | 0.500 | 4 |
| temporal | 0.312 | 16 |

## Ladder

| Run | Overall |
| --- | ---: |
| Pre ENG-171 | 2/30 |
| Post ENG-171 | 4/30 |
| Async + relative dates | 6–7/30 |
| **Content-bearing ranking** | **8/30 (26.7%)** |

Taxonomy: correct=8, retrieval_miss=9, answer_judge_miss=13.

L4 gate (≥12/30) still unmet. Dense embeddings still blocked (gateway 403).
