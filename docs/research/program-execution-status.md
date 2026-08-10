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
| 5 Hybrid reader | **Landed + early-qualified (local)** | Soft grounding + observability; local early pin `reader_source=hybrid_llm_packet` 17/30 |
| 6 Multi-hop packet depth | **Landed on `main` + `dev`** | Bridge/direct binding, second-pass, deterministic compose (PR #89 / `fc0fd93`) |
| 7 Proof pins | **Partial → V3 early** | 1×30 **16/30 MH 50%** ([pin](../benchmarks/artifacts/locomo-v3-early-pin-20260810.md)); 3×90 + LME-20 in flight; prior 3×90 27/90 |

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

## Recall Contract V3 (PR #92)

| Wave | Status | Notes |
| --- | --- | --- |
| A1 recency | **Done** | Newest-first contextual extract |
| A2 hybrid | **Done** | Soft grounding + explain reasons; fires locally |
| A3 early pin | **Done** | 16/30; hybrid 17/30; GO |
| B1 job barrier | **Done** | `wait_until_jobs_done` + publish fail-closed + chunking |
| B2 Mem0 ops | **Done** | ADD/UPDATE/DELETE/NONE → supersede/suppress |
| C1 typed hops | **Done** | resolve_entity → fetch_predicate second-pass |
| C2 concurrency | **Done** | `BRAINY_WORKER_CONCURRENCY` (staging default 4) |
| D qualify | **Partial** | OpMem 13/13 + marketing passed; 3×90 + LME-20 running |

## Still open (honest)

- Finish LME-20 under job barrier; then LME-100  
- Mem0 same-pin re-measure (needs `MEM0_API_KEY` / after 3×90)  
- Multi-seed full LoCoMo  
- Merge PR #92 → `dev` then staging re-pin  
- Pack authority / procedures / conflict packets  
- Hash/128 re-embed residue  
