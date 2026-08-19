# SOTA execution plan — from landed substrate to a defensible claim

**Status:** accepted plan (2026-08-19). R0–R10 substrate is merged; every score below is a *gate*, not a promise.  
**Does not claim:** SOTA today, a LoCoMo/LME target score, or that any published vendor number is a same-pin.  
**PoR:** [sota-representation-path.md](./sota-representation-path.md) · [70–80% path](./locomo-full-70-80-path.md) · [dual-path freeze](./locomo-dual-path-freeze.md)

## What "SOTA" is allowed to mean here

Per the cycle-closeout SOP, competitive language requires a **frozen same-pin win**. The claim this plan works toward is:

> On the same dataset SHA, same judge (temp 0), same answerer, same question set, run through the same harness, Brainy leads Mem0 Platform on LoCoMo (both lanes labeled), holds a non-embarrassing LME-20, and keeps its existing OpMem / marketing-vertical leads.

Vendor headlines (Mem0 92.5%, SuperMemory 95 LME Recall@15) are context rows, never the scoreboard. Even after a same-pin win, product copy stays factual ("leads same-pin LoCoMo vs Mem0 Platform on our harness") — the kill list on the word "SOTA" in product copy stays until explicitly lifted.

## Current state vs goal (all pins pre-date the merged stack)

| Axis | Pin | SHA | Goal state |
| --- | ---: | --- | --- |
| Product `/recall` full n=1540 | **175/1540 = 11.4%** | `1b5ab3e` | Competitive band on current SHA, measured |
| — full single-hop | 88/841 = 10.5% | `1b5ab3e` | Mass lever: compiler + structured answers |
| — full multi-hop | 21/282 = 7.4% | `1b5ab3e` | ID hop joins measured at scale |
| 1×30 conv-26 head | 21/30 (MH 10/10, OD 0/4, temporal 11/16) | `1b5ab3e` | Diagnostic only; keep MH/temporal |
| Industry search+harness | 49.8% (July, old stack) | old | Current-SHA rerun at top-k 200, labeled |
| Mem0 Platform same-pin 1×30 | 11/30 | our harness | Re-freeze at n=1540 same-pin |
| LME-20 | 4/20 (multi-session 0/5, KU 0/3) | `1b5ab3e` | Non-embarrassing before LME-500 |
| BEAM 100K | 8/20 | old | Unchanged priority (after LoCoMo/LME) |
| OpMem / marketing | 13/13 · 17/17 | current | Must never regress (merge gate) |

**Critical fact:** every LoCoMo/LME pin above was measured **before** R5A structured-first, the copula clip, R6a named-subject, and the R5B–R10 stack. The current SHA has never been scored at scale. The first move is measurement, not more code.

## Gap analysis by failure stage

Using the earliest-failing-stage model (R0):

1. **WRITE_MISS (largest known mass).** Full SH 10.5% vs 1×30 70% means compiler coverage does not travel across 841 items. R6a + `she`/`he` coref close the named-subject class; the remainder is provider-extract quality on arbitrary dialogue (deterministic atoms are patterns; the provider path is the general compiler). LME multi-session 0/5 is the same stage.
2. **ANSWER_PATH residue.** R5A retired slogans, but at full scale: 188 abstains, enumerate quality, and OD hypothetical yes/no (0/4 diagnostic) remain unmeasured post-R5A.
3. **ENTITY/RELATION/HOP.** R7–R9 landed with unit/held-out proof; full-MH behavior (plan coverage: does the planner even emit hops for the 282 items?) is unmeasured. One known 1×30 MH miss is image-gold.
4. **TEMPORAL at scale.** 11/16 on the head; full-suite temporal unknown. `temporal_score` still partly rides episodes rather than dated facts (PoR POV 8).
5. **KNOWLEDGE UPDATE (LME KU 0/3).** Supersede-on-update correctness for provider facts.
6. **MEASUREMENT.** R10 lanes are wired but no current-SHA run exists on either lane.

## Increments (S0 → S6)

Each increment has an exit gate. OpMem 13/13, marketing 17/17, and the held-out compiler audits are non-negotiable merge gates on every one. Allocation between S1–S5 is decided by the S0 ledger, not by intuition.

### S0 — Baseline the merged stack (measurement only, no product code)

- Stratified 150–200 LoCoMo items (proportional SH/MH/temporal/OD), **both lanes** on the merged SHA: `--eval-lane product-recall` and `--eval-lane industry-search` (top-k 200).
- Stage-oracle failure ledger for every miss (earliest failing stage).
- Exit: fresh failure histogram by category and stage; product-lane delta vs the 11.4-era subset; industry-lane delta vs 49.8 July band. This histogram re-orders S1–S5 if it disagrees with the analysis above.

### S1 — Compiler Coverage V3 (attacks WRITE_MISS)

- S1a: semantic-coverage audit harness — held-out conversations scored as `% durable claims compiled / entity-bound / dated / evidenced` (the R0 report shape), runnable as a single command.
- S1b: provider-extract iteration measured **only** by that audit (ADD-only fact semantics, report-about-B binding, durable assistant facts). No benchmark surface-forms in prompts.
- S1c: temporal attributes on facts — dated provider facts carry `event_start`/`valid_from`; `temporal_score` scores dated semantic records, not transcripts.
- Exit: held-out coverage at an agreed bar (propose ≥85% durable-claim compilation on fresh held-out conversations) plus stratified SH/temporal lift vs S0.

### S2 — Structured answer completion (attacks ANSWER residue)

- S2a: enumerate answers scoped by `entity_id` (activities-of-X lists only X's atoms).
- S2b: OD hypothetical yes/no synthesized from compiled facts (career/possession/preference) — the 0/4 diagnostic.
- S2c: abstain calibration — abstain only after structured + context support is assessed (the 188-abstain class).
- Exit: stratified OD/enumerate/abstain deltas vs S0; no episode-top-k stuffing.

### S3 — Identity and hops at full scale (attacks ENTITY/HOP)

- S3a: hop-plan coverage metric on the full-MH slice (what fraction of MH questions yields typed hops at all); extend the planner where plans are absent — planner work, not looser proof rules.
- S3b: alias/nickname lifecycle — capture in-dialogue aliases into `memory_entities`; ranked resolution stays deterministic; ambiguous first names stay ambiguous.
- S3c: image-context facts — `[visible text: …]` already compiles; decide whether caption ingestion for `image_urls` is in scope (the known image-gold MH miss). If skipped, record it as an accepted miss class.
- Exit: hop-plan coverage + `hop_join_proven` rate on stratified MH; two-Johns audits stay green.

### S4 — LME lane (multi-session + knowledge update)

- S4a: multi-session continuity measured on LME-20 (0/5 today) — session-scoped recall over compiled facts, not longer transcripts.
- S4b: knowledge-update supersede correctness (0/3 today) — newest dated fact wins; prior remains historical, not deleted.
- Exit: LME-20 rerun on the frozen 20-item sample; LME-500 stays forbidden as a claim until LME-20 stops being embarrassing.

### S5 — Industry lane hardening

- S5a: current-SHA industry run on the S0 subset — answerer consumes retrieved **atoms** first (harvest path), top-k 200, retrieved tokens reported.
- S5b: close any answerer-side gaps the ledger shows (extractive/list merge quality), keeping the answerer generic (no benchmark special-casing).
- Exit: industry-lane stratified score ≥ the July 49.8 band on current SHA, with token/latency reported.

### S6 — Freeze, same-pin, and the claim

- S6a: 3×90 qualification slice, both lanes, same seeds/judge.
- S6b: full freeze — n=1540 product `/recall` and industry lane (3 seeds), LME-20, plus **Mem0 Platform same-pin n=1540** through the identical harness (top-k 200, their API, cost/latency recorded, Platform-vs-OSS labeled).
- S6c: publish the two-lane tables + same-pin table (README rules: published-% sourced, same-pin only for lead/trail). Only a frozen same-pin win permits competitive language; the word "SOTA" in product copy remains gated even then until explicitly lifted.

## Decision rules

- After S0, spend the next increment on the **largest earliest-stage bucket** in the ledger. Expected order is S1 → S2 → S3, but the histogram outranks the expectation.
- Any increment that cannot show a stratified delta gets one iteration, then is re-scoped — no polishing without measurement.
- Nothing in S1–S5 may cite LoCoMo/LME surface forms in product code or prompts (overfit denylist enforces part of this).
- Full n=1540 runs happen exactly once, at S6, per freeze. Stratified subsets are the iteration currency.

## Kill list (unchanged, restated)

No fusion fishing, no graph DB default, no category dictionaries, no unbounded top-k, no LoCoMo/LME-named product rules, no episode top-k stuffing to restore OD/SH, no mixing 11.4% with Mem0 92.5%, no publishing 1×30 as full LoCoMo, no LME-500/BEAM-1M as quality claims, no SOTA/beats-Mem0 in product copy without a frozen same-pin win and explicit approval, no reopening R0–R10 as if missing.

## Linear

- ENG-168 conversational long-memory epic (this plan)
- ENG-176 multi-hop synthesis (S3a planner coverage)
- ENG-69 Graphiti temporal fact model (S1c input)
