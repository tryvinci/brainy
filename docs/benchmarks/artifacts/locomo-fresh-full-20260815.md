# LoCoMo full 10×all — fresh remasure — 2026-08-15

**Honest pin.** Product `/recall` on current `1b5ab3e` is **175/1540 (11.4%)**. That is a **dip** vs the last full pin **49.8%** (3-seed mean, 2026-07-31, search + harness answerer, old stack). Do not mix the two. Do not hide the dip. Not SOTA.

**Timestamp:** 2026-08-15T20:00:29Z
**Run:** `locomo-fresh-full-20260815-s0-33161a` (1 seed)
**Stack:** dedicated local API+worker rebuilt from `1b5ab3e`; async ingest; `BRAINY_USE_RECALL=1`; `BRAINY_RECALL_LLM=1`
**Dataset SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`
**Questions:** 10 conversations, cats 1–4, n=1540 (adversarial excluded from overall; 1986 items judged)
**Answer path:** all 1986 items via product `/recall` (`brainy-recall+answer` 1641, `+enumerate` 157, `+abstain` 188)

## Scores (categories 1–4)

| Metric | This pin (`/recall`) | Last full (2026-07-31, search+harness) |
| --- | ---: | ---: |
| Overall | **0.114 (175/1540)** | 0.494 seed-0 / **0.498 mean** |
| Search p50 ms | 230 | 2017 (old stack) |
| multi-hop | **21/282 (7.4%)** | 71/282 (25.2%) seed-0 |
| open-domain | **5/96 (5.2%)** | 37/96 (38.5%) seed-0 |
| single-hop | **88/841 (10.5%)** | 477/841 (56.7%) seed-0 |
| temporal | **61/321 (19.0%)** | 176/321 (54.8%) seed-0 |

`metrics.errors: 1` (one `JUDGE_MISS` on single-hop).

## Why this is not a 1×30 contradiction

Same SHA, same `/recall` path:

| Slice | Score |
| --- | ---: |
| Dedicated 1×30 (conv-26 head, MH/OD/temporal only) | **21/30 (70.0%)** |
| Same 30 items inside this full run | **20/30** (judge flake vs the dedicated pin) |
| Rest of conv-26 scored (mostly single-hop) | **12/122 (9.8%)** |
| Full n=1540 | **175/1540 (11.4%)** |

The 1×30 head is the MH/OD/temporal slice R4h optimized. Full LoCoMo is **841/1540 single-hop**. Product `/recall` often returns a nearby slogan, list, or `not in memory` instead of the atomic fact the July harness answerer extracted from search hits.

## Path label (required)

| Pin | n | Path | Overall |
| --- | ---: | --- | ---: |
| 2026-07-31 full | 1540 × 3 seeds | search hits → harness LLM answerer/judge | **49.8% mean** |
| **This pin** | 1540 × 1 seed | product `POST /recall` → same judge family | **11.4%** |
| This cycle 1×30 | 30 | product `/recall` | **70.0%** |

Publishing 70% as if it were full LoCoMo would be a lie. Publishing 49.8% as the current stack would also be a lie. Both numbers stay labeled.

## Claims

Allowed: current-stack full `/recall` is **11.4%**; 1×30 remains **21/30** measurement; OD/MH/SH all trail the July full pin. Forbidden: replacing 49.8% silently; calling 11.4% a harness glitch without evidence; SOTA; mixing n=30 with n=1540 in one percent.
