# LOCOMO 1×30 — fresh Mem0 same-pin — `locomo-mem0-samepin-20260813`

**PR10 measurement (qualification not claimed).** Same harness as the Brainy PR2 remasure. Mem0 **leads**. Do not write beats-Mem0 / SOTA.

**Timestamp:** 2026-08-13T11:39:53Z  
**System:** Mem0 Platform API  
**Dataset SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Answerer / judge:** same pinned LLM as Brainy (temp=0.0)  
**Questions:** conv-26, 30 items, seed 1

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | **0.400 (12/30)** |
| Search p50 ms | 470.9 |
| Search p95 ms | 564.4 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.700 (7/10) | 10 |
| open-domain | 0.750 (3/4) | 4 |
| temporal | 0.125 (2/16) | 16 |

Replicates the Aug 10 same-pin Mem0 **12/30 / 70% MH** (old V3 Brainy pin). This run is on the **current dataset SHA**, not a blog number.

## Same-pin vs Brainy (this day)

| System | Overall | MH | Temporal | OD |
| --- | ---: | ---: | ---: | ---: |
| Mem0 Platform | **12/30** | **7/10** | 2/16 | 3/4 |
| Brainy local PR2 `24be5ab` | 6/30 | 4/10 | 1/16 | 1/4 |

Mem0 leads this pin. Brainy 3×90 MH 22.2% (post-cutover staging) is a different question set size — do not mix.

## Frozen harness (PR10)

- LoCoMo dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`
- LME dataset SHA `d6f21ea9d60a0d56f34a05b609c79c88a451d2ae03597821ea3d5a9678c3a442`
- Judge + answerer temp **0.0**
- Publishable LME-20 quality is still **0/20** ([integrity pin](./lme20-product-recall-pr1-20260812-pin.md)). LME-100 is not a quality run.

Claim allowed **only** if Brainy wins a later frozen pin under this harness. Today it does not.
