# GAP-C5: retrieval budget + latency SLO

## API

- `GET /memories/search?...&limit=N` — cap returned results (product clients).
- Eval default `top_k=30` (Mem0 blog uses top_200 — different cost point).

## Observed staging p50 (search)

| Config | p50 | p95 |
| --- | ---: | ---: |
| Pre-diversify | ~1027 ms | ~2655 ms |
| Post content-dense + diversify | ~730–1270 ms | ~1540–1750 ms |
| Mem0 same-pin LOCOMO | ~419 ms | — |

## SLO targets (staging, conversational subject)

| Metric | Target | Status |
| --- | --- | --- |
| Search p50 | ≤ 1000 ms at top_k≤30 | **Met** on best pins; watch expansion |
| Search p95 | ≤ 2500 ms | **Met** recently |
| OpMem full suite | wall < 120 s | **Met** |

## Knobs

- Reduce list-query subject admit (48 → lower) if p50 drifts
- Client `limit` / eval `--top-k`
- Expansion caps in `expandSubjectContentMemories`
