# Program execution status — recall contract (2026-08-11)

**External handoff:** [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)  
**Re-review brief:** [external-reviews/2026-08-10-v3-rereview-brief.md](./external-reviews/2026-08-10-v3-rereview-brief.md)  
**Accepted review:** [external-reviews/2026-08-07-recall-contract-verdict.md](./external-reviews/2026-08-07-recall-contract-verdict.md)  
**Tips:** `main` `6b307a4` (production) · `dev` `9bad898` (staging = Recall Contract V3 + re-review brief)

## Course correction (accepted)

Prior default (“reader quality over packets”) replaced by end-to-end **recall contract**. Architect PR1–PR7 remain **closed**. Multi-hop packet depth is on **production**. External re-review: **KEEP V3, harden** (ordered writes → semantic hops → truthful sufficiency).

## Implementation status

| Step | Status | Notes |
| --- | --- | --- |
| 1 Measurement | **Landed** | Judge retry + `JUDGE_MISS`; LoCoMo speaker roles; `/jobs` + `/jobs/status`; harness job wait; tighter oracle overlap |
| 2 Provenance | **Landed** | `observed_at` on raw evidence; `evidence_id` on records/events |
| 3 Contextual compile | **Landed** | `ContextualExtractor` injects recent + related memories |
| 4 Entity-scoped state | **Landed** | `entity::predicate` keys when subject entity known |
| 5 Hybrid reader | **Landed + early-qualified (local)** | Soft grounding + observability; local early pin `reader_source=hybrid_llm_packet` 17/30 |
| 6 Multi-hop packet depth | **Landed on `main` + `dev`** | Bridge/direct binding, second-pass, deterministic compose (PR #89 / `fc0fd93`) |
| 7 Proof pins | **Partial → V3 early** | 1×30 **16/30 MH 50%** ([pin](../benchmarks/artifacts/locomo-v3-early-pin-20260810.md)); 3×90 **31/90**; Mem0 same-pin 12/30; LME deferred |

## Post-merge execution

| Step | Status | Notes |
| --- | --- | --- |
| Staging enable + smoke | **Done** | `/healthz`, `/jobs/status`, `/recall` second_pass |
| Multi-hop → production | **Done** | `main` cutover `fc0fd93` |
| OpMem non-reg | **Done** | Staging **13/13** (Gate 0 reconfirm) |
| Marketing non-reg | **Done** | Staging passed (Gate 0 reconfirm) |
| LME-20 / LME-100 | **In progress** | Product-recall harness shipped (#95); isolated LME-20 running |
| Larger LoCoMo | **Partial** | Gate 0 staging 1×30 **18/30**; 3×90 running |
| PR hygiene | **Open** | Hardening PRs #93–#98 draft on `dev` |

## Recall Contract V3 (PR #92) — merged to `dev`

| Wave | Status | Notes |
| --- | --- | --- |
| A1–C2 | **Done** | See prior status; merged via #92 |
| D qualify | **Partial** | Local early pins; staging Gate 0 1×30 **18/30** |

## V3 hardening cycle (post re-review) — open PRs

| PR | Status | Notes |
| --- | --- | --- |
| Gate 0 staging re-pin | **Done** | SHA `9bad898`; OpMem/marketing green; 1×30 **18/30**; 3×90 **32/90** (MH 19.4%, OD 42.9%) |
| #93 subject-ordered claims | **Open** | Serialize extraction per `(tenant,subject)` |
| #94 authoritative ops | **Open** | Provider NONE/DELETE/UPDATE filter-before-merge |
| #95 LME product-recall | **Open** | `--product-recall` + publish pins; path proven `/recall` |
| #96 hop executor V2 | **Open** | Typed hops + `hop_join_proven` |
| #97 sufficiency status | **Open** | Truthful hybrid AnswerStatus + hop-chain prompt (stacked on #96) |
| #98 qualify umbrella | **Open** | Combined branch + Gate 0 + harden local pins |

### Post-hardening local (combined #93–#97)

| Pin | Result |
| --- | --- |
| OpMem / marketing | 13/13 / passed |
| LoCoMo 1×30 | **14/30** (MH 5/10, OD 2/4) — dip vs Gate 0; expected from stricter hop join |
| LME-20 | product-recall path proven; full publish score in progress |

## Still open (honest)

- Finish isolated LME-20 publish score; LME-100 only after that  
- Staging re-pin after merging #93–#97 to `dev`  
- Mem0 same-pin + multi-seed LoCoMo  
- Merge hardening PRs → `dev` (not `main` without ask)  
- Pack authority / procedures / conflict packets  
- Hash/128 re-embed residue  

## Claims discipline

- Allowed: Gate 0 staging 18/30 and 32/90; harden local 14/30 with honesty about the dip; OpMem/marketing non-reg; product-recall path proven.  
- Forbidden: unqualified “beats Mem0”; SOTA; “MH solved”; LME accuracy before full `--publish --product-recall` completes; calling harden 1×30 an improvement; calling 3×90 MH 50% (Gate 0 MH is 19.4%).
