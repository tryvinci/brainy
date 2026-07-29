# LOCOMO smoke — multi-hop fix (speaker + intent + harvest)

**Staging:** `b3264b0` · gpt-oss · top_k=30 · async  

| Run | Overall | **multi-hop** | temporal | open | Notes |
| --- | ---: | ---: | ---: | ---: | --- |
| Diversify peak | 19/30 | 2/10 | 13/16 | 4/4 | |
| Entity hub v2 | 16/30 | 3/10 | 10/16 | 3/4 | identity pass |
| **Speaker+intent+harvest** | **19/30** | **5/10** | 11/16 | 3/4 | **single + Sweden pass** |
| Mem0 same-pin | 12/30 | **6/10** | 2/16 | 4/4 | |

OpMem: **12/12**.

## Multi-hop now passing

- research, identity, **relationship (single)**, **origin (Sweden)**, career

## Still failing (lists)

- activities / camped locations / kids likes / books / destress — partial retrieval or incomplete list vs GT

Gap to Mem0 MH: **1 question** (5/10 vs 6/10).
