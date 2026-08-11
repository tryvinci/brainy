# Program execution status — recall contract (2026-08-11)

**External handoff:** [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)  
**Self-review prompt (give to reviewer):** [external-reviews/2026-08-11-hardening-self-review-prompt.md](./external-reviews/2026-08-11-hardening-self-review-prompt.md)  
**Intake SOP:** [external-reviews/README.md](./external-reviews/README.md)  
**Accepted review:** [external-reviews/2026-08-07-recall-contract-verdict.md](./external-reviews/2026-08-07-recall-contract-verdict.md)  
**Tips:** `main` `308d3a1` (production = V3 hardening cutover) · `dev` `1f2f26f` (staging Render live)

## Course correction (accepted)

Prior default (“reader quality over packets”) replaced by end-to-end **recall contract**. Architect PR1–PR7 remain **closed**. Multi-hop packet depth is on **production**. External re-review: **KEEP V3, harden** (ordered writes → semantic hops → truthful sufficiency) — **executed and merged**.

## Implementation status

| Step | Status | Notes |
| --- | --- | --- |
| 1 Measurement | **Landed** | Judge retry + `JUDGE_MISS`; LoCoMo speaker roles; `/jobs` + `/jobs/status`; harness job wait; tighter oracle overlap |
| 2 Provenance | **Landed** | `observed_at` on raw evidence; `evidence_id` on records/events |
| 3 Contextual compile | **Landed** | `ContextualExtractor` injects recent + related memories |
| 4 Entity-scoped state | **Landed** | `entity::predicate` keys when subject entity known |
| 5 Hybrid reader | **Landed** | Soft grounding + truthful AnswerStatus + hop-chain prompt (#97) |
| 6 Multi-hop packet depth | **Landed on `main` + `dev`** | Plus hop executor V2 / `hop_join_proven` (#96) |
| 7 Proof pins | **Partial** | Gate 0 staging 18/30 + 32/90; harden local 14/30 (honest dip); LME path proven, full publish incomplete |

## Post-merge execution

| Step | Status | Notes |
| --- | --- | --- |
| Hardening PRs #93–#98 → `dev` | **Done** | Staging tip `1f2f26f` |
| Hardening → `main` (production) | **Done** | Explicit ask; tip `308d3a1` |
| Staging Render deploy | **Live** | API+worker on `1f2f26f` |
| OpMem non-reg | **Done** | Post-cutover staging `1f2f26f` **13/13** |
| Marketing non-reg | **Done** | Post-cutover staging **passed** |
| LME-20 / LME-100 | **Blocked / retry** | `--product-recall` path proven; publish run aborted on extraction job failure — not publishable |
| Larger LoCoMo | **Partial** | Post-cutover 1×30 **15/30**; 3×90 re-pin in progress |
| External self-review prompt | **Ready** | [2026-08-11-hardening-self-review-prompt.md](./external-reviews/2026-08-11-hardening-self-review-prompt.md) |

## Recall Contract V3 (PR #92) — merged

| Wave | Status | Notes |
| --- | --- | --- |
| A1–C2 | **Done** | Merged via #92 |
| D qualify | **Superseded by hardening qualify** | See Gate 0 + harden pins |

## V3 hardening cycle — closed (merged)

| PR | Status | Notes |
| --- | --- | --- |
| Gate 0 staging re-pin | **Done** | SHA `9bad898`; OpMem/marketing green; 1×30 **18/30**; 3×90 **32/90** (MH 19.4%, OD 42.9%) |
| #93 subject-ordered claims | **Merged** | Serialize extraction per `(tenant,subject)` |
| #94 authoritative ops | **Merged** | Provider NONE/DELETE/UPDATE filter-before-merge |
| #95 LME product-recall | **Merged** | `--product-recall` + publish pins; path proven `/recall` |
| #96 hop executor V2 | **Merged** | Typed hops + `hop_join_proven` |
| #97 sufficiency status | **Merged** | Truthful hybrid AnswerStatus + hop-chain prompt |
| #98 qualify umbrella | **Merged** | Gate 0 + harden docs |

### Post-hardening local (combined #93–#97)

| Pin | Result |
| --- | --- |
| OpMem / marketing | 13/13 / passed |
| LoCoMo 1×30 | **14/30** (MH 5/10, OD 2/4) — dip vs Gate 0; expected from stricter hop join |
| LME-20 | product-recall path proven; full publish aborted (job failure) — not an accuracy claim |

## Still open (honest)

- Clean isolated LME-20 `--publish --product-recall` score; LME-100 only after that  
- Finish / publish post-cutover staging re-pin on `1f2f26f`  
- Mem0 same-pin + multi-seed LoCoMo  
- Pack authority / procedures / conflict packets  
- Hash/128 re-embed residue  
- External adjudication of this hardening cutover via the self-review prompt  

## Claims discipline

- Allowed: Gate 0 staging 18/30 and 32/90; harden local 14/30 with honesty about the dip; OpMem/marketing non-reg; product-recall path proven; hardening on `dev`+`main`.  
- Forbidden: unqualified “beats Mem0”; SOTA; “MH solved”; LME accuracy before a clean full `--publish --product-recall` completes; calling harden 1×30 an improvement; calling 3×90 MH 50% (Gate 0 MH is 19.4%).
