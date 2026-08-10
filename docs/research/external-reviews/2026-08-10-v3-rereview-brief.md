# External re-review brief — Recall Contract V3 (2026-08-10)

**Purpose:** Hand this to an external review agent after V3 landed on **staging `dev`**. Adjudicate whether the adjusted recall-contract course is working, what remains blocked, and the next 3–7 PRs.

**Canonical pack:** [external-agent-assessment-pack.md](../external-agent-assessment-pack.md)  
**Prior briefs / reviews:** [2026-08-10-rereview-brief.md](./2026-08-10-rereview-brief.md) · [2026-08-07-recall-contract-verdict.md](./2026-08-07-recall-contract-verdict.md)  
**Live status:** [program-execution-status.md](../program-execution-status.md)  
**Qualification summary:** [recall-contract-v3-qualification-20260810.md](../../benchmarks/artifacts/recall-contract-v3-qualification-20260810.md)  
**PR:** https://github.com/tryvinci/brainy/pull/92 (merged → `dev`)

---

## 1. What to assume is true (do not re-litigate)

1. **Architect PR1–PR7 are closed.** Do not reopen fusion / graph DB / category dictionaries.
2. **Recall-contract sequence + multi-hop packet depth** already on production before V3.
3. **V3 adjusts the course** (your prior feedback): make hybrid fire like the harness answerer; soft grounding; job barrier; Mem0-style UPDATE/DELETE; typed hops only while MH lags; concurrency for LME scale.
4. **Do not recommend:** fusion retune, graph DB, category dictionaries, conversational SOTA claims, or “reader-only” as the next default.

---

## 2. What V3 shipped (on `dev` / staging)

| Wave | Change | Evidence |
| --- | --- | --- |
| A1 | Contextual extract newest-first + `[memory_id]` prior lines | `contextual_extractor.go` + test |
| A2 | Hybrid soft grounding + explain (`hybrid_reader_reason`) + OD skip + JSON salvage | `reader_hybrid.go` / `recall.go` |
| B1 | `wait_until_jobs_done` + publish fail-closed + ingest chunking | `evals/public/backends/brainy.py` |
| B2 | ADD/UPDATE/DELETE/NONE → supersede / suppress (sync + async) | `provider_extractor.go`, `memory_ops.go`, worker |
| C1 | Typed hops `resolve_entity` → `fetch_predicate` + single-subgoal second pass | `planner.go`, `recall.go` |
| C2 | `BRAINY_WORKER_CONCURRENCY` (staging default 4) | `cmd/worker`, `render.yaml` |

---

## 3. Measured pins after V3 (local agent host; staging re-pin pending Render deploy)

| Pin | Result | Notes |
| --- | ---: | --- |
| LoCoMo 1×30 V3 early | **16/30 (53.3%)**, MH **50%** | Was 14/30 after multi-hop; hybrid **17/30** `reader_source=hybrid_llm_packet` |
| Mem0 same-pin 1×30 | **12/30 (40%)**, MH **70%** | Same dataset budget |
| LoCoMo 3×90 | **31/90 (34.4%)** | Was **27/90 (30%)** |
| OpMem | **13/13** | Non-reg intact |
| Marketing | **passed** | Non-reg intact |
| LME-20 / LME-100 | **Deferred** | Barrier+concurrency shipped; run aborted under queue contention — **not publishable** |

Same-pin compare: [locomo-v3-samepin-vs-mem0-20260810.md](../../benchmarks/artifacts/locomo-v3-samepin-vs-mem0-20260810.md)

**Safe claim today:** Ops + marketing lead hold. Conversational overall recovered to the prior 16/30 peak with hybrid confirmed firing. Multi-convo improved 27→31/90. **Multi-hop still trails Mem0 (50% vs 70%).** LME still unproven under contract.

---

## 4. Residual risks / honesty gaps

- Pins above are **local** on the V3 binary; **staging Render re-pin** still needed after `dev` deploy (`BRAINY_RECALL_LLM=1` already on staging).
- MH gap vs Mem0 remains the clearest conversational deficit.
- LME-20 must be re-run alone on an empty queue with `--publish` / job barrier — do not cite partial LME.
- `json_parse_error` still appeared in hybrid reason mix (salvage landed after the early pin).
- Hash/128 re-embed residue; pack authority / procedures / conflict packets still open.

---

## 5. What we want from this re-review

Return structured feedback only:

1. **Verdict** — Keep / adjust / replace the V3-adjusted recall-contract course given these pins.
2. **MH gap** — Given hybrid now fires and typed hops shipped, is the remaining Mem0 MH lead mostly retrieval binding, extract freshness, or answer composition?
3. **Next 3–7 PRs** — Ordered, reviewable, failure-class tagged. Prefer staging re-pin, isolated LME-20 under barrier, MH-closing changes over architecture reopen.
4. **Claims discipline** — What may we say publicly now (overall same-pin vs Mem0?) vs what remains forbidden?
5. **Kill list** — Explicitly confirm what not to do next.

---

## 6. Required reading (in order)

1. This brief  
2. [recall-contract-v3-qualification-20260810.md](../../benchmarks/artifacts/recall-contract-v3-qualification-20260810.md)  
3. [locomo-v3-early-pin-20260810.md](../../benchmarks/artifacts/locomo-v3-early-pin-20260810.md)  
4. [locomo-v3-samepin-vs-mem0-20260810.md](../../benchmarks/artifacts/locomo-v3-samepin-vs-mem0-20260810.md)  
5. [locomo-v3-multiconvo-pin-20260810.md](../../benchmarks/artifacts/locomo-v3-multiconvo-pin-20260810.md)  
6. [program-execution-status.md](../program-execution-status.md)  
7. [external-agent-assessment-pack.md](../external-agent-assessment-pack.md) (for architecture context only)

---

## 7. Explicit non-goals for the reviewer

- Do not design a new memory product from scratch.
- Do not reopen fusion / RRF retune / graph DB / category dictionaries.
- Do not invent LME scores.
- Do not treat local pins as production-staging pins until Render `dev` deploy is re-measured.
