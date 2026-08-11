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

1. Hand the reviewer the **dedicated self-review prompt** (not a chat dump):  
   **[2026-08-11-hardening-self-review-prompt.md](./2026-08-11-hardening-self-review-prompt.md)**
2. Attach the assessment pack only as architecture context:  
   [external-agent-assessment-pack.md](../external-agent-assessment-pack.md)
3. Require the [TEMPLATE.md](./TEMPLATE.md) return shape (verdict + findings table + next sequence + kill list).
4. Adjudicate findings with code evidence before queueing PRs.

## Default priority after 2026-08-04 architecture verdict

> raw evidence → typed semantics → temporal truth → kill scan-heavy retrieval → plan evidence → executable vertical packs

**PR1–PR7 of that sequence are CLOSED** (2026-08-05 closeout). Do not re-queue them unless a new review reopens a finding.

## Default priority after V3 hardening cutover (2026-08-11)

**Accepted course:** KEEP V3, harden — then merge.  
**Landed:** PRs #93–#98 on `dev` (`1f2f26f`) and production `main` (`308d3a1`). Staging Render live on `1f2f26f`.

**Current handoff / self-review prompt:**  
[2026-08-11-hardening-self-review-prompt.md](./2026-08-11-hardening-self-review-prompt.md)

**Prior briefs (historical):**  
[2026-08-10-v3-rereview-brief.md](./2026-08-10-v3-rereview-brief.md) · [2026-08-10-rereview-brief.md](./2026-08-10-rereview-brief.md) · [2026-08-07-recall-contract-verdict.md](./2026-08-07-recall-contract-verdict.md)

**Next work (in order):**

1. Finish isolated LME-20 `--publish --product-recall` (no publishable LME accuracy until a clean full run)
2. Complete / record staging re-pin on deploy `1f2f26f` (OpMem, marketing, LoCoMo 1×30 + 3×90)
3. Mem0 same-pin re-measure under identical budget
4. Multi-seed LoCoMo
5. Pack authority / procedures / conflict packets

Do **not** default to fusion retune, graph DB, category dictionaries, or re-opening architect PR1–PR7.

## Pin honesty (binding for reviewers and agents)

| Pin family | Rule |
| --- | --- |
| Gate 0 staging (`9bad898`) | 1×30 **18/30**; 3×90 **32/90** with MH **19.4%** (not 50%) |
| Harden local (#93–#97) | 1×30 **14/30** — **dip**, not a win vs Gate 0 |
| LME | Path `/recall` proven; aborted/partial runs are **not** accuracy claims |
| Production | `main` `308d3a1` holds the hardening cutover — still no conversational SOTA language |

## Rejected by default (unless new evidence)

- Hand-tuned fusion constants without real lexical ranks
- Expanding benchmark-shaped query-category dictionaries
- Graph DB before canonical entities + temporal reads + planner
- Treating `memory_current_state` as canonical truth
- Ungrounded rolling profiles
- Top-k inflation as architecture substitute
- Calling harden scores an improvement when they dipped
- Publishable LME / SOTA / “beats Mem0” without matching artifacts
