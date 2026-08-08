# LoCoMo pin after multi-hop packet depth — 2026-08-08

**Run:** `locomo-smoke-86032312`  
**Commit:** `3f73f43` (dev tip with multi-hop)  
**Flags:** `BRAINY_USE_RECALL=1`, `BRAINY_RECALL_LLM=1`, async provider extract  
**Host:** local API+worker

## Scores

| Metric | Value |
| --- | ---: |
| Overall | **46.7% (14/30)** |
| multi-hop | **50% (5/10)** |
| temporal | 56.2% (9/16) |
| open-domain | 0% (0/4) |

Full report: [locomo-smoke-multihop-20260808.md](./locomo-smoke-multihop-20260808.md)

## vs prior recall-contract same-pin

| Pin | Overall | MH |
| --- | ---: | ---: |
| 2026-08-07 Brainy `ef197919` | 53.3% (16/30) | 40% |
| 2026-08-08 after multi-hop | 46.7% (14/30) | **50%** |
| 2026-08-07 Mem0 same-pin | 36.7% (11/30) | 70% |

**Read:** Multi-hop improved 40%→50% under the new binder/second-pass; overall dipped (open-domain 0/4). Still behind Mem0 on MH. Not a SOTA claim.
