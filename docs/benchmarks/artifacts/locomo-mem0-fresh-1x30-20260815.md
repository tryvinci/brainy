# LOCOMO 1×30 — fresh Mem0 same-pin — `locomo-mem0-fresh-1x30-20260815`

**This-cycle measurement.** Same harness, dataset SHA, question set (conv-26, 30 items, cats 1–4), judge temp 0.0 as Brainy [locomo-fresh-1x30-20260815.md](./locomo-fresh-1x30-20260815.md). Mem0 Platform **trails overall**. Do not write SOTA / beats-Mem0.

**Timestamp:** 2026-08-15T15:35:39Z
**System:** Mem0 Platform API (not Mem0 OSS)
**Dataset SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`
**Ingest:** 419 turns / 19 sessions, conv-26

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | **0.367 (11/30)** |
| Search p50 ms | 491.5 |
| Search p95 ms | 614.6 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.600 (**6/10**) | 10 |
| open-domain | 0.750 (**3/4**) | 4 |
| temporal | 0.125 (**2/16**) | 16 |

Prior freeze 2026-08-13 was **12/30 / MH 7/10 / OD 3/4 / temporal 2/16**. This re-run is **−1 overall, −1 MH**; OD and temporal unchanged. Use **this** pin for lead/trail this cycle.

## Same-pin vs Brainy (this day)

| System | Overall | MH | OD | Temporal | Search p50 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Brainy local `1b5ab3e` | **21/30 (70.0%)** | **10/10** | **0/4** | **11/16** | 175 ms (local) |
| Mem0 Platform | **11/30 (36.7%)** | **6/10** | **3/4** | **2/16** | 492 ms (platform) |

Brainy leads overall / MH / temporal on this freeze. Brainy **trails open-domain 0/4 vs 3/4**. Local vs platform latency is a harness observation, **not** an SLO.

Mem0 OSS was not measured. Do not mix this 1×30 with Mem0 blog 90+ LoCoMo.
