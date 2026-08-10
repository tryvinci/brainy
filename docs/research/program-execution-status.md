# Program execution status — recall contract (2026-08-10)

**External handoff:** [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)  
**Re-review brief:** [external-reviews/2026-08-10-rereview-brief.md](./external-reviews/2026-08-10-rereview-brief.md)  
**Accepted review:** [external-reviews/2026-08-07-recall-contract-verdict.md](./external-reviews/2026-08-07-recall-contract-verdict.md)  
**Tips:** `main` `1ac592f` (production) · `dev` `b885038` (staging, synced)

## Course correction (accepted)

Prior default (“reader quality over packets”) replaced by end-to-end **recall contract**. Architect PR1–PR7 remain **closed**. Multi-hop packet depth is on **production**.

## Implementation status

| Step | Status | Notes |
| --- | --- | --- |
| 1 Measurement | **Landed** | Judge retry + `JUDGE_MISS`; LoCoMo speaker roles; `/jobs` + `/jobs/status`; harness job wait; tighter oracle overlap |
| 2 Provenance | **Landed** | `observed_at` on raw evidence; `evidence_id` on records/events |
| 3 Contextual compile | **Landed** | `ContextualExtractor` injects recent + related memories |
| 4 Entity-scoped state | **Landed** | `entity::predicate` keys when subject entity known |
| 5 Hybrid reader | **Landed + staging on** | Packet items + coverage; staging `BRAINY_RECALL_LLM=1` |
| 6 Multi-hop packet depth | **Landed on `main` + `dev`** | Bridge/direct binding, second-pass, deterministic compose (PR #89 / `fc0fd93`) |
| 7 Proof pins | **Partial** | 1×30 **14/30 MH 50%**; 3×90 **27/90**; LME incomplete |

## Post-merge execution

| Step | Status | Notes |
| --- | --- | --- |
| Staging enable + smoke | **Done** | `/healthz`, `/jobs/status`, `/recall` second_pass |
| Multi-hop → production | **Done** | `main` cutover `fc0fd93` |
| OpMem non-reg | **Done** | Staging **13/13** |
| Marketing non-reg | **Done** | Staging passed |
| LME-20 / LME-100 | **Partial / deferred** | Sync LME-20 incomplete; LME-100 needs capacity |
| Larger LoCoMo | **Partial** | 3×90 done; 3-seed full still open |
| PR hygiene | **Done** | Stale draft PR #24 closed; agent PRs #85–#91 merged |

## Still open (honest)

- Finish LME under job barriers; then LME-100  
- Staging Mem0 same-pin re-measure (esp. multi-hop) after hybrid confirm  
- Multi-seed full LoCoMo  
- Pack authority / procedures / conflict packets  
- Hash/128 re-embed residue  
