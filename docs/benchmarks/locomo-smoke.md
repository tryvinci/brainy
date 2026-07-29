# LOCOMO smoke — Mem0-style extract + entity hub

**Staging:** `a61fa6f` · gpt-oss · top_k=30 · async  

| Run | Overall | multi-hop | temporal | open |
| --- | ---: | ---: | ---: | ---: |
| Prior diversify peak | 19/30 | 2/10 | 13/16 | 4/4 |
| Entity hub v1 (broken quotes) | 16/30 | 2/10 | 10/16 | 4/4 |
| **Entity hub v2 (quote hotfix)** | **16/30** | **3/10** | 10/16 | 3/4 |
| Mem0 same-pin | 12/30 | **6/10** | 2/16 | 4/4 |

OpMem: **12/12**.

## Multi-hop movement

- **q4 identity now passes** (was fail) — transgender signal surfaces.
- Sweden atom appears (`moved from Sweden`) but still under-ranked / weak speaker attribution (`User` vs Caroline) → answerer abstains.
- Remaining MH gap vs Mem0 is still mostly **list completeness** + a few missing atoms.

See `docs/research/multihop-mem0-learn.md`.
