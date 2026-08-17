# External reviews — intake process

Standing process for architecture / SOTA reviews from external agents or humans.

## SOP

1. **Receive** the review (paste or attachment).
2. **Adjudicate** against the current codebase (spot-check critical claims; do not accept vibes alone).
3. **Archive** under `docs/research/external-reviews/YYYY-MM-DD-short-title.md` using [TEMPLATE.md](./TEMPLATE.md).
4. **Refresh handoff surfaces** so the next agent sees truth:
   - `docs/research/external-agent-assessment-pack.md`
   - `docs/research/codebase-graph.md` + `codebase-graph.json`
   - `docs/research/program-execution-status.md`
5. **Queue work** in the accepted priority order (do not default to fusion retune).
6. **Link** the archive from `docs/research/README.md`.

## How to commission the next external pass

This pass is **received**: [2026-08-17-competitive-archaeology-verdict.md](./2026-08-17-competitive-archaeology-verdict.md).

For a later pass (after R5):

1. Write a **new** dedicated self-review prompt (do not reuse the 2026-08-11 hardening prompt or treat a `bd987fa`-pinned report as live).
2. Attach the dip diagnosis and this verdict.
3. Attach the assessment pack as architecture context (read **CURRENT 2026-08-17** first).
4. Require the [TEMPLATE.md](./TEMPLATE.md) return shape.
5. Adjudicate findings with code evidence before queueing PRs. Historical prompts: [2026-08-17-full-recall-self-review-prompt.md](./2026-08-17-full-recall-self-review-prompt.md) · [2026-08-11-hardening-self-review-prompt.md](./2026-08-11-hardening-self-review-prompt.md)

## Default priority after 2026-08-04 architecture verdict

> raw evidence → typed semantics → temporal truth → kill scan-heavy retrieval → plan evidence → executable vertical packs

**PR1–PR7 of that sequence are CLOSED** (2026-08-05 closeout). Do not re-queue them unless a new review reopens a finding.

## Default priority after V3 hardening cutover (2026-08-11)

**Accepted course:** KEEP V3, harden — then merge.  
**Landed:** PRs #93–#98 on `dev` (`1f2f26f`) and production `main` (`308d3a1`). Staging Render live on `1f2f26f`.

## Default priority after competitive architecture verdict (2026-08-11)

**Accepted:** [2026-08-11-competitive-architecture-verdict.md](./2026-08-11-competitive-architecture-verdict.md) — KEEP V3; **adjust** next program to competitive parity (Mem0 recall + Graphiti relations + Brainy governed truth).

**Competitive SOP:** [../competitive/README.md](../competitive/README.md) · [gap matrix](../competitive/competitive-gap-matrix.md)

**This-pass verdict (received):**  
[2026-08-17-competitive-archaeology-verdict.md](./2026-08-17-competitive-archaeology-verdict.md)

**This-pass self-review prompt:**  
[2026-08-17-full-recall-self-review-prompt.md](./2026-08-17-full-recall-self-review-prompt.md)

**Prior handoff / self-review prompt (historical):**  
[2026-08-11-hardening-self-review-prompt.md](./2026-08-11-hardening-self-review-prompt.md)

**Prior briefs (historical):**  
[2026-08-10-v3-rereview-brief.md](./2026-08-10-v3-rereview-brief.md) · [2026-08-10-rereview-brief.md](./2026-08-10-rereview-brief.md) · [2026-08-07-recall-contract-verdict.md](./2026-08-07-recall-contract-verdict.md)

**Representation-path amendment (2026-08-14):** [2026-08-14-representation-path-additions.md](./2026-08-14-representation-path-additions.md) — accepted. Execution: [sota-representation-path.md](../sota-representation-path.md).

**R0–R4 landed as measurement** (1×30 MH 10/10; OD still 0/4). Do not re-queue them. A 2026-08-17 archaeology report pinned to Wave 1 `bd987fa` asked for P0-P7 of that sequence — **rejected as live work**. Verdict: [2026-08-17-competitive-archaeology-verdict.md](./2026-08-17-competitive-archaeology-verdict.md).

**Next work (in order) — after 2026-08-17 adjudication:**

1. **R5 structured-first** — Make product `/recall` cite compiled facts (deterministic operators), not `firstStatementFromPacket` slogans / list-cue enumerate / abstain when the fact exists. This is the **11.4% → ~50%** lever **and** the 1×30 OD **0/4 vs Mem0 3/4** trail. Not a transcript reader.
2. **Remaining WRITE_MISS compiler coverage** — durable claims as facts. Lever: **~50 → Mem0-class** (July search+harness was already 49.8% vs Mem0 92.5% n=1540 top-k 200).
3. **R2 full entity IDs** when identity misses show up — not ahead of R5.
4. **R6 two-lane remasure** — 1×30 diagnostic → **3×90** → full `/recall` *and* search+harness, labeled separately → LME-20 quality → LME-500 only after that.
5. Keep OpMem 13/13 and marketing 17/17 green. Small parallel: latency histogram sum (`metrics.go`).

Do **not** spend the next cycle on another full remasure, on LME-500 as a quality claim, or on re-implementing R0-R4.

Do **not** default to fusion retune, graph DB, category dictionaries, hop-heuristic sprawl, or re-opening architect PR1–PR7.

## Pin honesty (binding for reviewers and agents)

| Pin family | Rule |
| --- | --- |
| **Live (2026-08-15 remasure, product `1b5ab3e`)** | OpMem **13/13**; marketing **17/17**; LoCoMo 1×30 **21/30** (MH 10/10, OD **0/4**, temporal 11/16) vs Mem0 **11/30**; full `/recall` **175/1540 (11.4%)**; LME-20 **4/20**; BEAM 100K **8/20**. `dev`=`main`=`8492ad3` |
| Full LoCoMo 11.4% | Named **dip** vs July search+harness **49.8%**. Not a harness glitch. Not current 49.8%. Not 70% as full LoCoMo. |
| Vendor 90+ | Mem0 92.5% is n=1540, **top-k 200**, LLM-over-search — **not** Brainy `/recall`. SuperMemory 95 LME is Recall@15. |
| LME-500 / BEAM 1M | **Not run** (cost). Do not invent scores. |
| Gate 0 staging (`9bad898`) | Historical. 1×30 **18/30**; 3×90 **32/90** with MH **19.4%** (not 50%) |
| Harden local (#93–#97) | Historical. 1×30 **14/30** — **dip**, not a win vs Gate 0 |
| Post-cutover staging (`1f2f26f`) | Historical. 1×30 **15/30**; 3×90 **33/90** with MH **22.2%** |
| LME integrity | 0/20 was integrity; quality pin is now **4/20** |
| Production | `main` = `dev` = `8492ad3` — still no conversational SOTA language |

## Rejected by default (unless new evidence)

- Hand-tuned fusion constants without real lexical ranks
- Expanding benchmark-shaped query-category dictionaries
- Graph DB before canonical entities + temporal reads + planner
- Treating `memory_current_state` as canonical truth
- Ungrounded rolling profiles
- Top-k inflation as architecture substitute
- Calling harden scores an improvement when they dipped
- Publishable LME / SOTA / “beats Mem0” without matching artifacts
- Publishing 70% as full LoCoMo; restoring 49.8% as current; mixing 92.5 vs 70
- Restoring OD/SH by stuffing episodes into top-k
- LME-500 or BEAM 1M as a quality claim while LME-20 is 4/20 and BEAM 100K is 8/20
