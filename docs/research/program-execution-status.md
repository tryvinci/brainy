# Program execution status — recall contract (2026-08-07)

**External handoff:** [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)  
**Accepted review:** [external-reviews/2026-08-07-recall-contract-verdict.md](./external-reviews/2026-08-07-recall-contract-verdict.md)  
**Merged to `dev` (staging):** PR #88 (recall-contract) + related docs PRs

## Course correction (accepted)

Prior default (“reader quality over packets”) replaced by end-to-end **recall contract**. Architect PR1–PR7 remain **closed**.

## Implementation status (on `dev`)

| Step | Status | Notes |
| --- | --- | --- |
| 1 Measurement | **Landed on `dev`** | Judge retry + `JUDGE_MISS`; LoCoMo speaker roles; `/jobs` + `/jobs/status`; harness job wait; tighter oracle overlap |
| 2 Provenance | **Landed on `dev`** | `observed_at` on raw evidence; `evidence_id` on records/events |
| 3 Contextual compile | **Landed on `dev`** | `ContextualExtractor` injects recent + related memories |
| 4 Entity-scoped state | **Landed on `dev`** | `entity::predicate` keys when subject entity known |
| 5 Hybrid reader | **Landed on `dev`** | Packet items + coverage tighten; `BRAINY_RECALL_LLM=1` |
| 6 Proof pins | **Partial** | LoCoMo same-pin Brainy **16/30** vs Mem0 **11/30**; LME under contract still incomplete |

Proof: [recall-contract-proof-20260807.md](../benchmarks/artifacts/recall-contract-proof-20260807.md)

## Flags

| Env | Default | Purpose |
| --- | --- | --- |
| `BRAINY_RECALL_LLM` | off | Enable hybrid LLM reader on `/recall` |
| `BRAINY_USE_RECALL` | off | Eval harness uses product `/recall` |
| `BRAINY_EVIDENCE_STRICT` | off | Reserved for fail-closed evidence writes |

## Next execution order (post-merge)

1. **Staging enable** — set `BRAINY_RECALL_LLM=1` on staging API after deploy; confirm `/jobs/status` healthy  
2. **Finish LME** — stratified LME-20 then LME-100 with job-completion wait (scale worker if backlog)  
3. **Multi-hop packet depth** — bridge-chain binding / second retrieval pass (Mem0 still leads MH 70% vs 40%)  
4. **Larger LoCoMo** — fixed multi-conversation slice, then 3-seed full under same pins  
5. **OpMem + marketing non-reg** on staging after deploy  
6. **Deferred** — pack authority / procedures / conflicts; hash/128 re-embed; third vertical  

## Still open (honest)

- Pack authority / procedures / conflict packets  
- Hash/128 re-embed residue  
- Full multi-seed LoCoMo + finished LME-100 under new contract  
- Staging re-pin of Mem0 same-pin after deploy (local pin is directional)  
