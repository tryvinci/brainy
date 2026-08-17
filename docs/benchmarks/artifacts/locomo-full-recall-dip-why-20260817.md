# Why full LoCoMo is 11.4% on product `/recall` (2026-08-17)

**Status:** diagnosis of the 2026-08-15 remasure. Product SHA `1b5ab3e` (no product change in that remasure).  
**Does not claim:** SOTA, a restored 49.8%, or that 11.4% is a harness glitch.

## One paragraph

We did **not** drop the 1×30. Conv-26 head is **21/30** (MH **10/10**, OD **0/4**, temporal **11/16**) vs the R4h pin **20/30**. We **did** drop full LoCoMo because this remasure scored **product `POST /recall`** (175/1540 = **11.4%**) instead of the July **search-hits + harness LLM answerer** path (mean **49.8%**, seed-0 **49.4%**). Sampling the gitignored run JSON shows `/recall` often returning a nearby **slogan**, an enumerate **word-salad**, or **`not in memory`**, while the compiled fact (or at least the gold substring in an episode) sits in the packet. That is an **answer-path** failure on top of the still-real **representation** gap (July 49.8% vs Mem0 Platform **92.5%** on n=1540).

## What we measured this freeze

| Pin | n | Path | Overall |
| --- | ---: | --- | ---: |
| 2026-07-31 historical | 1540 × 3 seeds | search hits → harness LLM answerer + LLM judge, old stack | **49.8%** mean (seed-0 **49.4%**) |
| 2026-08-15 remasure | 1540 × 1 seed | **product `POST /recall`**, current stack | **175/1540 (11.4%)** |
| Same remasure, 1×30 slice | 30 | product `/recall`, conv-26 head | **21/30** (same 30 inside full run: 20/30, judge flake) |

Full run id: `locomo-fresh-full-20260815-s0-33161a`. Artifact: [locomo-fresh-full-20260815.md](./locomo-fresh-full-20260815.md). Dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`. All 1986 judged items went through `/recall` (`brainy-recall+answer` 1641, `+enumerate` 157, `+abstain` 188).

## Why 70% and 11.4% are not a contradiction

The 1×30 is conv-26 **head only** (MH / OD / temporal). Full LoCoMo cats 1–4 is **841/1540 single-hop**. The rest of conv-26 in the full run is **12/122 (9.8%)**. Publishing 70% as full LoCoMo is invalid.

## By group vs July seed-0 (same n, different path)

| Group | This `/recall` | July search+harness seed-0 |
| --- | ---: | ---: |
| overall | 175/1540 (**11.4%**) | 761/1540 (**49.4%**) |
| multi-hop | 21/282 (7.4%) | 71/282 (25.2%) |
| open-domain | 5/96 (5.2%) | 37/96 (38.5%) |
| single-hop | 88/841 (**10.5%**) | 477/841 (**56.7%**) |
| temporal | 61/321 (19.0%) | 176/321 (54.8%) |

The mass of the dip is **single-hop**. MH on the 1×30 is already 10/10; MH on all 10 convos is 7.4% because the other nine conversations are not the R4h hop graph.

## Smoking-gun examples (public LoCoMo; same conversation as the 1×30)

From `docs/benchmarks/runs/locomo-fresh-full-20260815-s0-33161a.json` (gitignored, ~17MB, still on disk for this diagnosis):

| qid | Question (abbrev.) | Gold | `/recall` answer | Mode |
| --- | --- | --- | --- | --- |
| `conv-26-q83` | What did Melanie realize after the charity race? | self-care is important | We Can Really Accept Who We Are And Be Content | `brainy-recall+answer` |
| `conv-26-q84` | How did she realize that? | (self-care how) | Don't Forget To Prepare Emotionally, Since The Wait Can Be Hard | `+answer` |
| `conv-26-q85` | Caroline summer plans | researching adoption agencies | Adoption Agencies, Back, Wicked, Love, … | `+enumerate` |
| `conv-26-q86` | (LGBTQ+ individuals) | LGBTQ+ individuals | Need Any Help | `+answer` |

Mechanism in `internal/memory/recall.go`: `answer` / `enumerate` / abstain. `firstStatementFromPacket` and list-cue enumerate pick chat slogans. Industry harnesses sit an LLM on search hits. 188 items abstained (`not in memory`), including 1×30 OD `q14` / `q27`.

Do not restore OD/SH by stuffing episodes into top-k. Do not add LoCoMo-named product rules.

## Two stacked gaps (do not collapse)

1. **Answer-path gap.** Product `/recall` cites slogans / enumerate / abstain instead of compiled facts. Closing this is what could move **11.4% back toward ~50%** on *current* memory. This is product reader/synthesis, not “run more benches.”
2. **Representation gap.** Even July search+harness was **49.8% vs Mem0 Platform 92.5%** (n=1540, their harness, **top-k 200**). That is WRITE_MISS / thin compiled facts / they retrieve atomic memories at top-k 200. Closing this is the **~50 → 90** path. 1×30 OD is still **0/4 WRITE_MISS**.

11.4% is not a harness glitch. 49.8% is not “current.” 70% is not full LoCoMo.

## Are vendor percents the same kind of run?

**No.** Closest industry-format LoCoMo is Mem0 Platform **92.5% (1425/1540)**, cats 1–4, 10 convos, LLM-as-judge, **top-k 200**. Path is **ingest → search → LLM answer → LLM judge**, not Brainy product `/recall` with abstain/enumerate. Our published top-k was **30**.

| System | LoCoMo number | What it actually is |
| --- | --- | --- |
| Mem0 Platform | 92.5% | n=1540, top-k 200, their harness |
| Mem0 paper 2025 | 68.5 / 66.9 | **not** the 2026 92.5; Letta compared against this |
| Zep | 75.14% J-score | disputed; Mem0 re-ran Zep at **58.44%** |
| SuperMemory | 77.1 LoCoMo; **95 LME is Recall@15** | not LLM-judge |
| Letta | 74% | agent + filesystem grep, vs then-published Mem0 68.5 |
| Hindsight | 92% | their AMB harness |
| Full-context baseline | ~73% | dump the whole convo; no memory system |

**92.5 vs 11.4** is the honest industry-format compare **on this stack** (n=1540) but **not the same answer path / top-k**. **92.5 vs 49.8** is the old search+harness pin (still far behind). **92.5 vs 70** is invalid (n=1540 vs n=30).

Same-pin Mem0 this cycle (same SHA, judge temp 0, conv-26 1×30): Brainy **21/30** vs Mem0 **11/30** (MH 10 vs 6, OD **0 vs 3**, temporal 11 vs 2). Lead this freeze; **trail OD**. Not SOTA.

Sources: [published-claims.md](../published-claims.md).

## Why LME-500 and BEAM 1M/10M were not run

Cost/time, not forgotten:

- **LME-20** took ~7h (~250 extract jobs/item). 500 items = tens of hours + huge LLM extract spend. Marker: skipped. LME-20 landed **4/20** product `/recall` (lift vs 0/20 integrity, same seed/SHA). Multi-session still 0/5. Running 500 would not change the diagnosis.
- **BEAM 100K conv-0** = **8/20** search+harness (non-reg vs hist. 8/20). BEAM 1M is 700 Q on huge chats; Mem0’s 64.1 is that tier, **not** 100K/20q. 1M/10M not affordable and not justified while 8/20.

## What it takes to move (PoR; do not substitute)

Accepted path: [sota-representation-path.md](../../research/sota-representation-path.md). Kill list unchanged.

Two product moves, in order of **mass**:

1. **Make `/recall` cite compiled facts** (not `firstStatementFromPacket` slogans, not list-cue enumerate, not abstain when the fact exists). That is the **11.4% → ~50%** lever. Related to **R5 structured-first** (POV 10: answer model consumes structured values; source text is provenance).
2. **Compiler coverage (WRITE_MISS)** so durable claims exist as facts: career/possession/dated plans, then remaining R2 canonical IDs, R3 projection, R4 hops as needed. That is the **~50% → Mem0-class** lever. 1×30 OD 0/4 is the visible WRITE_MISS slice; full single-hop 10.5% is the mass.

Then R6: 1×30 diagnostic → **3×90** → remasure full `/recall` → LME-20 quality → LME-500 only after that. Keep OpMem 13/13 and marketing 17/17 green. No SOTA / beats-Mem0.

Do **not** spend the next cycle on another full remasure or on LME-500 as a quality claim.

Archaeology review (2026-08-17): two published lanes (product `/recall` vs search+harness); cite-facts and R5 OD are the same structured-first family. [verdict](../../research/external-reviews/2026-08-17-competitive-archaeology-verdict.md).
