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
| W6 load `latency_load.py` (200 mem, 120 Q, c=8, top_k=30) 2026-07-31 | **2403 ms** | **4997 ms** |

Load artifact: `docs/benchmarks/runs/latency-load-20260731T065251Z.json`.
SLOs **not met under concurrency=8** while FTS GIN backfill is incomplete (`content_tsv` null on most historical rows; index present but cold). Idle/single-flight pins still closer to prior table.

## SLO targets (staging, conversational subject)

| Metric | Target | Status |
| --- | --- | --- |
| Search p50 | ≤ 1000 ms at top_k≤30 | **Met** on light pins; **miss** under c=8 load (2403 ms) |
| Search p95 | ≤ 2500 ms | **Met** light; **miss** under c=8 load (4997 ms) |
| OpMem full suite | wall < 120 s | **Met** |

## Knobs

- Reduce list-query subject admit (48 → lower) if p50 drifts
- Client `limit` / eval `--top-k`
- Expansion caps in `expandSubjectContentMemories`
