# Self-review prompt for external reviewer — full LoCoMo `/recall` dip (2026-08-17)

**Copy/paste this entire file to the external reviewer.**
It is self-contained. Prefer SHA + artifact citations over vibes.

**Canonical pack:** [../external-agent-assessment-pack.md](../external-agent-assessment-pack.md)
**Intake SOP:** [README.md](./README.md) · **Archive template:** [TEMPLATE.md](./TEMPLATE.md)
**Live status:** [../program-execution-status.md](../program-execution-status.md)
**Diagnosis (required):** [../../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md](../../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md)
**PoR:** [../sota-representation-path.md](../sota-representation-path.md)
**Cycle closeout:** [../competitive/cycle-closeout.md](../competitive/cycle-closeout.md) (section **2026-08-15/16**)
**Vendor percents (sourced, not same-pin):** [../../benchmarks/published-claims.md](../../benchmarks/published-claims.md)

**Repo tips at handoff:** `dev` = `main` = `8492ad3` (docs remasure + production FF 2026-08-17, explicit approval). Product SHA for those pins is still **`1b5ab3e`**. No product change in the remasure.

---

## Role

You are an independent architecture / measurement adjudicator for Brainy (Go + Postgres memory service).
Your job is a **fresh self-review of the 2026-08-15 full remasure and the 11.4% LoCoMo dip**. Challenge claims. Spot honesty gaps. Propose the next 3–7 reviewable PRs only.

Do **not** redesign the product from scratch. Do **not** reopen architect PR1–PR7 or R0–R4 without new contradictory evidence. Do **not** treat this as a request to run LME-500 or BEAM 1M.

---

## Assume true (do not re-litigate)

1. Five-plane target (source → evidence → semantic → projection → recall) remains the course.
2. Architect PR1–PR7 are **closed**. Representation path R0–R4 have **landed as measurement** (1×30 MH **10/10** on conv-26 head).
3. Accepted PoR: [sota-representation-path.md](../sota-representation-path.md) — compile interactions into facts/entities/relations; retrieve those; keep episodes as provenance + fallback. Not reader-first, not retrieval-tuning-first, not fusion-first.
4. Default rejects still stand: fusion retune, graph DB default, category dictionaries, LoCoMo/LME-named product rules, unbounded top-k, conversational SOTA / beats-Mem0 language, treating 1×30 as qualification, restoring OD/SH by stuffing episodes into top-k.
5. `dev` is GitHub homepage / default (staging). `main` is production. Both currently `8492ad3`.
6. OpMem **13/13** and marketing **17/17** must stay green. Those leads are real on this freeze (Mem0 10/13 and 4/17 empirical). They are not a substitute for conversational recall.

---

## What this cycle actually did

Measurement only. Product SHA `1b5ab3e`. Dedicated local API+worker, fresh DB, async ingest, `BRAINY_USE_RECALL=1`. Merged remasure docs: GitHub PR #126.

We re-pinned every in-tree suite at full (or max affordable) size, with a same-cycle Mem0 Platform counter-run on OpMem, marketing, and LoCoMo 1×30. We did **not** run LME-500 or BEAM 1M/10M (cost; see below).

---

## Measured pins (cite these; do not invent)

All 2026-08-15 remasure unless noted. Dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`.

| Suite | Brainy | Mem0 this cycle / published | Artifact |
| --- | ---: | ---: | --- |
| OpMem | **13/13** | **10/13** (fails `cor02`, `sup03`, `upd02`) | `opmem-fresh-local-20260815.md` · `opmem-mem0-fresh-20260815.md` |
| Marketing | **17/17** | **4/17** empirical (parity content 4/4) | `marketing-fresh-local-20260815.md` · `marketing-mem0-fresh-20260815.md` |
| LoCoMo 1×30 (conv-26 head) | **21/30** (MH **10/10**, OD **0/4**, temporal **11/16**) | **11/30** (MH 6/10, OD **3/4**, temporal 2/16) | `locomo-fresh-1x30-20260815.md` · `locomo-mem0-fresh-1x30-20260815.md` |
| Same 30 items inside full run | **20/30** | — | judge flake vs dedicated 1×30 |
| LoCoMo full cats 1–4 | **175/1540 (11.4%)** product `/recall`, 1 seed | published **92.5%** (1425/1540, **top-k 200**, their harness) — **not** this pin | `locomo-fresh-full-20260815.md` |
| LME-20 | **4/20** product `/recall` (jobs 4829=4829) | no fair pin on this harness; published 94.4% is 500 Q, top-k 200 | `lme20-fresh-20260815.md` |
| LME-500 | **not run** | — | skipped; LME-20 already 4/20, multi-session 0/5 |
| BEAM 100K conv-0 | **8/20** search+harness | published 64.1 is **BEAM 1M / 700 Q**, not 100K/20q | `beam-100k-fresh-20260815.md` |
| BEAM 1M / 10M | **not run** | — | skipped; 8/20 does not justify the spend |

Historical (different path, older stack; **not current**):

| Pin | n | Path | Overall |
| --- | ---: | --- | ---: |
| 2026-07-31 | 1540 × 3 seeds | search hits → harness LLM answerer + LLM judge | **49.8%** mean (seed-0 **49.4%**) |
| R4h 1×30 | 30 | product `/recall` | **20/30** (MH 10/10, OD 0/4, temporal 10/16) |

**Honesty rule for the dip:** 1×30 did **not** drop (20/30 → 21/30). Full LoCoMo **did** drop vs July because we scored **product `POST /recall`**, not search+harness. Do not restore 49.8% as current. Do not publish 70% as full LoCoMo. Do not call 11.4% a harness glitch.

### Full `/recall` vs July search+harness (same n, different path)

| Group | This `/recall` | July search+harness seed-0 |
| --- | ---: | ---: |
| overall | 175/1540 (**11.4%**) | 761/1540 (**49.4%**) |
| multi-hop | 21/282 (7.4%) | 71/282 (25.2%) |
| open-domain | 5/96 (5.2%) | 37/96 (38.5%) |
| single-hop | 88/841 (**10.5%**) | 477/841 (**56.7%**) |
| temporal | 61/321 (19.0%) | 176/321 (54.8%) |

Full LoCoMo is **841/1540 single-hop**. The 1×30 never scores those. Rest of conv-26 in the full run: **12/122 (9.8%)**. Full-suite MH **7.4%** is not a contradiction of 1×30 MH **10/10**: the other nine conversations are not the R4h hop graph.

Full run id: `locomo-fresh-full-20260815-s0-33161a`. All 1986 judged items via `/recall` (`brainy-recall+answer` 1641, `+enumerate` 157, `+abstain` 188).

---

## Two stacked gaps (challenge this; do not collapse)

Internal diagnosis, for you to accept / modify / reject:

1. **Answer-path gap.** Product `/recall` cites slogans / enumerate word-salad / `not in memory` instead of compiled facts (or even the gold substring sitting in an episode in the packet). Closing this is what could move **11.4% back toward ~50%** on *current* memory. This is product reader/synthesis, not “run more benches.”
2. **Representation gap.** Even July search+harness was **49.8% vs Mem0 Platform 92.5%** (n=1540, their harness, **top-k 200**; our published top-k was **30**). That is WRITE_MISS / thin compiled facts / they retrieve atomic memories at top-k 200. Closing this is the **~50 → Mem0-class** path. 1×30 OD is still **0/4 WRITE_MISS**.

### Smoking-gun examples (public LoCoMo; same conversation as the 1×30)

From the gitignored run JSON `docs/benchmarks/runs/locomo-fresh-full-20260815-s0-33161a.json`:

| qid | Gold (abbrev.) | `/recall` answer | Mode |
| --- | --- | --- | --- |
| `conv-26-q83` | self-care is important | We Can Really Accept Who We Are And Be Content | `brainy-recall+answer` |
| `conv-26-q84` | (self-care how) | Don't Forget To Prepare Emotionally, Since The Wait Can Be Hard | `+answer` |
| `conv-26-q85` | researching adoption agencies | Adoption Agencies, Back, Wicked, Love, … | `+enumerate` |
| `conv-26-q86` | LGBTQ+ individuals | Need Any Help | `+answer` |

Mechanism: `internal/memory/recall.go` — `answer` / `enumerate` / abstain. `firstStatementFromPacket` returns the first non-question `pkt.Contents` string (or `TemporalAnswer`). List-cue `looksListQuery` joins enumerate values. Industry harnesses sit an LLM on search hits. 188 items abstained, including 1×30 OD `q14` / `q27`.

Do not restore OD/SH by stuffing episodes into top-k. Do not add LoCoMo-named product rules.

---

## Are vendor percents the same kind of run?

**Internal claim: no.** Closest industry-format LoCoMo is Mem0 Platform **92.5% (1425/1540)**, cats 1–4, 10 convos, LLM-as-judge, **top-k 200**. Path is **ingest → search → LLM answer → LLM judge**, not Brainy product `/recall` with abstain/enumerate.

| System | LoCoMo number | What it actually is |
| --- | --- | --- |
| Mem0 Platform | 92.5% | n=1540, top-k 200, their harness |
| Mem0 paper 2025 | 68.5 / 66.9 | **not** the 2026 92.5; Letta compared against this |
| Zep | 75.14% J-score | disputed; Mem0 re-ran Zep at **58.44%** |
| SuperMemory | 77.1 LoCoMo; **95 LME is Recall@15** | not LLM-judge |
| Letta | 74% | agent + filesystem grep, vs then-published Mem0 68.5 |
| Hindsight | 92% | their AMB harness |
| Full-context baseline | ~73% | dump the whole convo; no memory system |
| Brainy July 2026 | 49.8% | search+harness, old stack, top-k 30 |
| Brainy now | 11.4% | product `/recall`, current stack, n=1540 |
| Brainy 1×30 | 70% | n=30 conv-26 head; **not** full LoCoMo |

**92.5 vs 11.4** is the honest industry-format compare **on this stack** (n=1540) but **not the same answer path / top-k**. **92.5 vs 49.8** is the old search+harness pin (still far behind). **92.5 vs 70** is invalid (n=1540 vs n=30).

Same-pin Mem0 this cycle (same SHA, judge temp 0, conv-26 1×30): Brainy **21/30** vs Mem0 **11/30**. Lead this freeze; **trail OD**. Not SOTA.

Mem0 OSS ≠ Mem0 Platform. Graphiti OSS ≠ Zep Platform.

---

## Why LME-500 and BEAM 1M/10M were not run

Cost/time, not forgotten. Challenge whether that was the right call:

- **LME-20** took ~7h (~250 extract jobs/item). 500 items = tens of hours + huge LLM extract spend. LME-20 landed **4/20** product `/recall` (lift vs 0/20 integrity, same seed/SHA). Multi-session still 0/5. Running 500 would not change the diagnosis.
- **BEAM 100K conv-0** = **8/20** search+harness (non-reg vs hist. 8/20). BEAM 1M is 700 Q on huge chats; Mem0’s 64.1 is that tier, **not** 100K/20q.

---

## Claims discipline (enforce)

**Allowed now**

- OpMem 13/13 vs Mem0 10/13 this freeze (lead ops).
- Marketing 17/17 vs Mem0 4/17 empirical this freeze (lead governed vertical).
- LoCoMo 1×30 **21/30** vs Mem0 **11/30** this freeze (lead overall / MH / temporal; trail OD 0/4 vs 3/4). Measurement, not qualification.
- Full LoCoMo **175/1540 (11.4%)** product `/recall` as a **named dip** vs July search+harness 49.8%.
- LME-20 **4/20** as lift vs own 0/20 integrity pin, not vs published 94.4%.
- BEAM 100K **8/20** as non-reg, not as BEAM 1M.

**Forbidden**

- Unqualified “beats Mem0” / SOTA / “MH solved”
- Publishing 70% as full LoCoMo
- Restoring 49.8% as current
- Calling 11.4% a harness glitch
- Mixing 92.5 vs 70 (n=1540 vs n=30)
- Comparing LME 4/20 to Mem0 94.4% or SuperMemory 95 Recall@15
- Comparing BEAM 8/20 to Mem0 64.1 (1M / 700 Q)
- Treating Gate 0 / harden / Wave 1 / R1c pins as live
- LoCoMo-named product rules; stuffing episodes to restore OD/SH
- Spending the next cycle on another full remasure or on LME-500 as a quality claim

---

## Required reading (in order)

1. This prompt
2. [locomo-full-recall-dip-why-20260817.md](../../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md)
3. [README.md](./README.md) (intake / current priority)
4. Cycle closeout **2026-08-15/16** in [cycle-closeout.md](../competitive/cycle-closeout.md)
5. Pins: `locomo-fresh-full-20260815.md`, `locomo-fresh-1x30-20260815.md`, `locomo-mem0-fresh-1x30-20260815.md`, `lme20-fresh-20260815.md`, `beam-100k-fresh-20260815.md`
6. [sota-representation-path.md](../sota-representation-path.md) (esp. R5 structured-first, kill list)
7. Hot path skim: `internal/memory/recall.go` (`firstStatementFromPacket`, enumerate, abstain), `reader_hybrid.go`, compiler/extract
8. [published-claims.md](../../benchmarks/published-claims.md) for vendor path labels
9. [external-agent-assessment-pack.md](../external-agent-assessment-pack.md) — architecture context; **ignore stale “next is R1b / LME 0/20 / READER_MISS-dominated” language** except as history
10. Prior briefs only if needed: [2026-08-14-representation-path-additions.md](./2026-08-14-representation-path-additions.md), [2026-08-11-competitive-architecture-verdict.md](./2026-08-11-competitive-architecture-verdict.md)

---

## Return format (mandatory)

Use [TEMPLATE.md](./TEMPLATE.md). Fill every section. In the verdict paragraph, answer **Keep representation-first course / adjust sequence / replace**.

Also answer these questions explicitly:

1. **Course** — Is the two-gap split (answer-path 11.4%→~50%, then representation ~50→Mem0-class) the right diagnosis, or is 11.4% mostly WRITE_MISS with slogans as a symptom?
2. **Published metric** — Should the **product** number stay `POST /recall` (honest product), or should we also pin search+harness as the industry-format LoCoMo number (n=1540, ideally top-k 200 + LLM-over-search)? Which number belongs on the README bake-off row?
3. **First product PR** — Given full single-hop **10.5%** is the mass and 1×30 OD is **0/4**, is **R5-on-OD** still the right first 1×30 step, or should “`/recall` cites compiled facts” land first? Name the failure class.
4. **Fair Mem0 stack** — What is a fair n=1540 compare (top-k, answerer, judge, seeds) we should eventually run, without calling it current?
5. **Skipped suites** — Was skipping LME-500 and BEAM 1M/10M correct? When would either change a product decision?
6. **Next 3–7 PRs** — Ordered, reviewable, failure-class tagged. Prefer: make `/recall` cite facts; R5 structured-first OD; remaining WRITE_MISS compiler coverage. Avoid architecture reopen, fusion fishing, LME-500-as-quality, another full remasure.
7. **Claims** — What may we say publicly after this remasure vs what remains forbidden?
8. **Kill list** — Confirm what not to do next.

### Findings table requirement

For each finding: **Accept / Modify / Reject**, code evidence (file + symbol), and a concrete action. Reject vibes-only findings.

---

## Explicit non-goals

- Do not invent LME-500 / BEAM 1M scores.
- Do not treat 1×30 70% as full LoCoMo or as qualification.
- Do not propose LOCOMO-named regexes, held-out prompt tuning, or episode top-k inflation to restore SH/OD.
- Do not default to “add a graph database.”
- Do not reopen architect PR1–PR7 or R0–R4 without new measured contradiction.
- Do not recommend spending the next cycle on another full remasure.
