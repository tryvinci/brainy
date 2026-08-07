# Program execution status — recall contract (2026-08-07)

**External handoff:** [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)  
**Accepted review:** [external-reviews/2026-08-07-recall-contract-verdict.md](./external-reviews/2026-08-07-recall-contract-verdict.md)

## Course correction

Prior default (“reader quality over packets”) replaced by end-to-end **recall contract**:

1. Measurement honesty  
2. Evidence ↔ semantics provenance  
3. Context-aware semantic compile  
4. Entity-scoped state keys  
5. Plan → packet → sufficiency → hybrid reader  
6. LME / LoCoMo / Mem0 proof pins  

Architect PR1–PR7 remain **closed**.

## Implementation status (this branch)

| Step | Status | Notes |
| --- | --- | --- |
| 1 Measurement | **Landed** | Judge retry + `JUDGE_MISS`; LoCoMo speaker roles; `/jobs` + `/jobs/status`; harness job wait; tighter oracle overlap |
| 2 Provenance | **Landed** | `observed_at` on raw evidence; `evidence_id` / `raw_evidence_ids` on records; events carry evidence_id |
| 3 Contextual compile | **Landed** | `ContextualExtractor` injects recent + related memories; provider prompt link/update rules |
| 4 Entity-scoped state | **Landed** | `entity::predicate` keys in `current_state` when subject entity known |
| 5 Hybrid reader | **Landed** | Packet items + coverage tighten; `BRAINY_RECALL_LLM=1` bounded LLM over packet |
| 6 Proof pins | **Partial** | Fresh LoCoMo same-pin Brainy 16/30 vs Mem0 11/30; LME-20 still draining async backlog |

## Flags

| Env | Default | Purpose |
| --- | --- | --- |
| `BRAINY_RECALL_LLM` | off | Enable hybrid LLM reader on `/recall` |
| `BRAINY_USE_RECALL` | off | Eval harness uses product `/recall` |
| `BRAINY_EVIDENCE_STRICT` | off | Reserved for fail-closed evidence writes |

## Still open (honest)

- Pack authority / procedures / conflict packets  
- Hash/128 re-embed residue  
- Full multi-seed LoCoMo + finished LME-100 under new contract  
- Fresh Mem0 same-pin on post-contract stack  
