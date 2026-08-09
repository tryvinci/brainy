# Program execution status — recall contract (2026-08-08)

**External handoff:** [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)  
**Accepted review:** [external-reviews/2026-08-07-recall-contract-verdict.md](./external-reviews/2026-08-07-recall-contract-verdict.md)  
**Merged to `dev` (staging):** PR #88 (recall-contract) + PR #89 (multi-hop packet depth)  
**Also on `main`:** `fc0fd93` — multi-hop packet depth + post-merge pins (production)

## Course correction (accepted)

Prior default (“reader quality over packets”) replaced by end-to-end **recall contract**. Architect PR1–PR7 remain **closed**.

## Implementation status (on `dev`)

| Step | Status | Notes |
| --- | --- | --- |
| 1 Measurement | **Landed** | Judge retry + `JUDGE_MISS`; LoCoMo speaker roles; `/jobs` + `/jobs/status`; harness job wait; tighter oracle overlap |
| 2 Provenance | **Landed** | `observed_at` on raw evidence; `evidence_id` on records/events |
| 3 Contextual compile | **Landed** | `ContextualExtractor` injects recent + related memories |
| 4 Entity-scoped state | **Landed** | `entity::predicate` keys when subject entity known |
| 5 Hybrid reader | **Landed + staging on** | Packet items + coverage tighten; staging API `BRAINY_RECALL_LLM=1` (+ provider env for hybrid) |
| 6 Multi-hop packet depth | **Landed on `dev`** | Bridge/direct binding, second-pass retrieval, deterministic chain compose (PR #89 / `db64d02`) |
| 7 Proof pins | **Partial** | LoCoMo after multi-hop: **14/30** overall, **MH 50%** ([pin](../benchmarks/artifacts/locomo-multihop-pin-20260808.md)); prior same-pin 16/30 MH 40%; LME-20 sync in flight |

Proof: [recall-contract-proof-20260807.md](../benchmarks/artifacts/recall-contract-proof-20260807.md) · staging smoke [staging-postmerge-smoke-20260808.md](../benchmarks/artifacts/staging-postmerge-smoke-20260808.md)

## Flags

| Env | Default | Purpose |
| --- | --- | --- |
| `BRAINY_RECALL_LLM` | off (staging **on**) | Enable hybrid LLM reader on `/recall` |
| `BRAINY_USE_RECALL` | off | Eval harness uses product `/recall` |
| `BRAINY_EVIDENCE_STRICT` | off | Reserved for fail-closed evidence writes |

## Post-merge execution (2026-08-08)

| Step | Status | Notes |
| --- | --- | --- |
| Staging enable + smoke | **Done** | `/healthz`, `/jobs/status`, sync `/ingest` + `/recall` second_pass |
| Multi-hop packet depth | **Done on `dev`** | PR #89 merged |
| OpMem non-reg | **Done** | Staging Brainy **13/13** — [artifact](../benchmarks/artifacts/opmem-staging-nonreg-20260808.md) |
| Marketing non-reg | **Done** | Staging passed — [artifact](../benchmarks/artifacts/marketing-staging-nonreg-20260808.md) |
| LME-20 / LME-100 | **Partial** | LME-20 sync in progress — [partial](../benchmarks/artifacts/lme20-partial-20260808.md); LME-100 deferred (needs worker capacity) |
| Larger LoCoMo | **Partial** | 1×30: 14/30 MH 50%; multi-convo 3×90: **27/90 (30%)** — [pin](../benchmarks/artifacts/locomo-multiconvo-pin-20260808.md) |
| Deferred | Open | Pack authority / procedures / conflicts; hash/128 re-embed; third vertical |

## Still open (honest)

- Pack authority / procedures / conflict packets  
- Hash/128 re-embed residue  
- Finished LME-100 under new contract + multi-seed LoCoMo  
- Staging Mem0 same-pin re-measure after hybrid reader confirmed live  
- ~~Merge multi-hop to `main`~~ **done** (`fc0fd93`)  
