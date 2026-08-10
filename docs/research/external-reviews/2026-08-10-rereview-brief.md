# External re-review brief — Brainy recall contract (2026-08-10)

**Purpose:** Hand this to an external review agent for a fresh adjudication of where Brainy stands **after** the accepted 2026-08-07 recall-contract course correction and the 2026-08-08 post-merge execution cutover.

**Canonical pack:** [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)  
**Prior accepted reviews:** [2026-08-07-recall-contract-verdict.md](./external-reviews/2026-08-07-recall-contract-verdict.md) · [2026-08-04-architecture-verdict.md](./external-reviews/2026-08-04-architecture-verdict.md)  
**Live status:** [program-execution-status.md](./program-execution-status.md)  
**Repo tips:** `main` = `1ac592f` (production) · `dev` = `b885038` (staging; synced with main)

---

## 1. What to assume is true (do not re-litigate)

1. **Architect PR1–PR7 are closed** (2026-08-05). Reopen only with new contradictory evidence.
2. **Recall-contract sequence (steps 1–5) landed** on `dev` then `main` (PR #88 / merge `175c4fa` lineage).
3. **Multi-hop packet depth landed** on `dev` and **production `main`** (PR #89 → cutover `fc0fd93`).
4. **Do not recommend:** fusion retune, graph DB, category dictionaries, conversational SOTA claims, or “reader-only” as the next default.

---

## 2. What shipped since the last accepted review

| Area | Status | Evidence |
| --- | --- | --- |
| Measurement contract | Landed | Judge retry/`JUDGE_MISS`; LoCoMo speaker roles; `/jobs` + `/jobs/status`; harness job wait; tighter oracles |
| Provenance | Landed | `observed_at` on raw evidence; `evidence_id` on records/events |
| Contextual extract | Landed | `ContextualExtractor` prior-context / update rules |
| Entity-scoped state | Landed | `entity::predicate` keys |
| Hybrid reader | Landed + staging on | Packet-bound LLM reader behind `BRAINY_RECALL_LLM=1` (staging API on; provider env set) |
| Multi-hop packet depth | Landed on **prod** | `multiHopTargets`, bridge/direct binding, second-pass retrieval, deterministic compose (`internal/memory/multihop_packet.go`) |
| Staging smoke | Done | `/healthz`, `/jobs/status`, `/recall` with `second_pass` observed |
| OpMem non-reg | Done | Staging Brainy **13/13** |
| Marketing non-reg | Done | Staging **passed** (parity 4/4; capabilities 11/11) |

---

## 3. Current measured pins (honest)

| Pin | Result | Notes |
| --- | ---: | --- |
| LoCoMo same-pin recall-contract (2026-08-07) | Brainy **16/30 (53%)** vs Mem0 **11/30 (37%)** | Directional; Brainy MH **40%**, Mem0 MH **70%** |
| LoCoMo 1×30 after multi-hop (2026-08-08) | **14/30 (47%)**, MH **50%** | MH up 40→50; overall down vs prior (open-domain 0/4) |
| LoCoMo multi-convo 3×90 | **27/90 (30%)** | Harder slice; not a smoke-equivalent |
| OpMem staging | **13/13** | Ops lead intact |
| Marketing staging | **passed** | Vertical non-reg intact |
| LME-20 | **Partial / incomplete** | Sync ingest slow; HTTP 400s on some haystacks; **not a publishable pin** |
| LME-100 | **Deferred** | Prior absolute ~4%; needs empty queue + worker capacity |

Safe claim today: **ops + marketing lead under same pins; conversational improving under recall contract; multi-hop improved but still not Mem0-class; LME still the hard gap.**

---

## 4. Known residual risks

- Hybrid `reader_source=hybrid_llm_packet` not yet consistently confirmed on staging smokes (deterministic packet + second_pass are live).
- Async extraction queues can starve under LME-scale backlog (observed; cleared operationally once).
- Sync LME path hits HTTP 400 on some large haystacks.
- Hash/128 re-embed residue remains.
- Pack authority / procedures / conflict packets still open.
- No fresh Mem0 same-pin re-measure after multi-hop + hybrid staging cutover.

---

## 5. What we want from this re-review

Return structured feedback only:

1. **Verdict** — Keep / adjust / replace the recall-contract course given the new multi-hop + staging pins.
2. **Failure taxonomy** — Is multi-hop still primarily READER_MISS / coverage / representation? What changed with bridge-chains?
3. **Next 3–7 PRs** — Ordered, reviewable, with expected failure-class impact. Prefer LME job-barrier completion, MH composition quality, and honest same-pin re-measure over architecture reopen.
4. **Claims discipline** — What may we say publicly now vs what remains forbidden?
5. **Kill list** — Explicitly confirm what not to do next.

---

## 6. Required reading (in order)

1. This brief  
2. [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)  
3. [program-execution-status.md](./program-execution-status.md)  
4. [2026-08-07-recall-contract-verdict.md](./external-reviews/2026-08-07-recall-contract-verdict.md)  
5. Pins: [recall-contract-proof-20260807.md](../benchmarks/artifacts/recall-contract-proof-20260807.md) · [locomo-multihop-pin-20260808.md](../benchmarks/artifacts/locomo-multihop-pin-20260808.md) · [locomo-multiconvo-pin-20260808.md](../benchmarks/artifacts/locomo-multiconvo-pin-20260808.md) · [opmem-staging-nonreg-20260808.md](../benchmarks/artifacts/opmem-staging-nonreg-20260808.md) · [lme20-partial-20260808.md](../benchmarks/artifacts/lme20-partial-20260808.md)

Optional map: [codebase-graph.md](./codebase-graph.md)

---

## 7. Git / PR hygiene (as of this brief)

- Production: `main` @ `1ac592f`
- Staging: `dev` @ `b885038` (synced with main)
- Stale open draft PR #24 closed (superseded by later provider/staging work)
- Agent feature PRs #85–#91 merged to `dev`; multi-hop + pins also on `main`
