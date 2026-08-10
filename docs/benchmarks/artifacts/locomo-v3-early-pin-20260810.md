# LoCoMo early pin — Recall Contract V3 Wave A — 2026-08-10

**Run:** `locomo-v3-early-20260810`  
**Commit:** `4909832` (recency + hybrid soft grounding + job barrier + Mem0-style ops + typed hops)  
**Flags:** `BRAINY_USE_RECALL=1`, `BRAINY_RECALL_LLM=1`, async provider extract  
**Host:** local API+worker (`BRAINY_WORKER_CONCURRENCY=4`)

## Scores

| Metric | Value |
| --- | ---: |
| Overall | **53.3% (16/30)** |
| multi-hop | **50% (5/10)** |
| temporal | 62.5% (10/16) |
| open-domain | 25% (1/4) |

Report: [locomo-v3-early-20260810.md](./locomo-v3-early-20260810.md)

## `reader_source` distribution (replay)

Replay of the same 30 questions against the ingested tenant (`locomo-*-conv-26`):

| Source | n |
| --- | ---: |
| `hybrid_llm_packet` | **17** |
| `evidence_packet` | 13 |

Multi-hop slice: hybrid **7/10**, packet 3/10.

Hybrid reasons observed: `ok` 13 · `freeform_accepted` 4 · `json_parse_error` 10 · `abstain` 3  
(JSON salvage follow-up landed after this pin.)

## Go / no-go vs prior pin

| Pin | Overall | MH | Hybrid firing |
| --- | ---: | ---: | --- |
| 2026-08-08 multi-hop | 14/30 (47%) | 50% | not confirmed |
| **2026-08-10 V3 early** | **16/30 (53%)** | 50% | **yes (17/30)** |
| 2026-08-07 recall-contract peak | 16/30 | 40% | n/a |
| 2026-08-07 Mem0 same-pin | 11/30 | 70% | n/a |

**Decision:** **GO** — overall not worse than 14/30; hybrid confirmed firing; OD no longer blanked (1/4 vs 0/4). Proceed Wave B already shipped; MH still trails Mem0 70% so typed hops remain justified; continue Wave D qualification.

Not a SOTA / “beats Mem0” claim.
