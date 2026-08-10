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

## Default priority after 2026-08-04 architecture verdict

> raw evidence → typed semantics → temporal truth → kill scan-heavy retrieval → plan evidence → executable vertical packs

**PR1–PR7 of that sequence are CLOSED** (2026-08-05 closeout). Do not re-queue them unless a new review reopens a finding.

## Default priority for the next external agent pass

**Accepted 2026-08-07:** [2026-08-07-recall-contract-verdict.md](./2026-08-07-recall-contract-verdict.md)  
**Re-review brief 2026-08-10:** [2026-08-10-rereview-brief.md](./2026-08-10-rereview-brief.md)  
**Landed on `main` + `dev`:** recall-contract + multi-hop packet depth. Staging: `BRAINY_RECALL_LLM=1`; OpMem 13/13 + marketing non-reg passed; LoCoMo MH 50% on 1×30 after multi-hop.

> **Next:** finish LME-20/100 under job barriers → staging Mem0 same-pin re-measure → multi-seed LoCoMo → pack authority/procedures/conflicts

Do **not** default to fusion retune, graph DB, or re-opening architect PR1–PR7.

## Rejected by default (unless new evidence)

- Hand-tuned fusion constants without real lexical ranks
- Expanding benchmark-shaped query-category dictionaries
- Graph DB before canonical entities + temporal reads + planner
- Treating `memory_current_state` as canonical truth
- Ungrounded rolling profiles
- Top-k inflation as architecture substitute
