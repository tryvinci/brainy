# Benchmax audit — leftover covering vs actually beating benchmarks

**Date:** 2026-08-25  
**SHA audited:** `f09c4f4` (P53 pin docs) / product `ae15e40`  
**Does not claim:** SOTA, beats-Mem0, that 137/180 is 90%, that leftover covering is worthless, or that we should delete the covering stack.

This is an honesty stop, not a remasure. No new LoCoMo / LME / Mem0 score.

## Verdict

The skip-ingest LoCoMo 180 leftover-covering campaign **is benchmaxxing-adjacent and must stop.** It is not how Brainy beats public benchmarks.

Each detector is written as generic English (query shape + leftover shape). The *queue of which shape to add next* is the remaining items on tenant `diag-mh-135`. That is Goodhart: we optimized the diagnostic 180, not the product that would win LoCoMo n=1540, LongMemEval, or a fair Mem0 same-pin.

**Do not ship P54** as a `peaceful moments` / nature / miss-about covering for `conv-44-q62`. That item is the example of saturation: the gold leftover is stored, but recovering it needs either a category dictionary or another one-item query-shape detector.

## What we actually score (do not mix)

| Pin | Path | n | Score | What it is |
| --- | --- | ---: | ---: | --- |
| Product `/recall` full | `POST /recall` | 1540 | **175/1540 = 11.4%** | Industry-format LoCoMo on this stack. SHA `1b5ab3e`. **Not re-run after leftover covering.** |
| Historical search+harness | search → LLM answerer | 1540 × 3 | **49.8%** mean | July, old stack. Not a current-SHA ceiling. |
| LoCoMo 1×30 | product `/recall` | 30 | **21/30 (70%)** | Diagnostic head. Not full LoCoMo. |
| This-VM skip-ingest 180, hybrid **off** | product `/recall` | 180 | **19/180** | Same store, no reader. |
| This-VM skip-ingest 180, hybrid **on** P53 | product `/recall` + leftover covering | 180 | **137/180 (76.1%)** | Same 10 convos, fail-closed skip-ingest, seed 1. |
| This-VM industry | search+harness, top-k 200 | 180 | **62/180 (34.4%)** | Closest *lane* to Mem0's published protocol. **Unchanged by leftover covering.** |
| Mem0 Platform published | their harness, top-k 200 | 1540 | **92.5%** | Context row. Not same-pin. |
| Mem0 Platform same-pin 1×30 | our harness, handicapped | 30 | **11/30** | Do not refresh lead/trail from 137 vs 11. |
| LME-20 | product `/recall` | 20 | **4/20** | Not re-run. |
| Fair Mem0 180 | v3, top-k 200 | 180 | **no pin** | HTTP 429 until **2026-09-01**. |

137/180 does **not** replace 11.4%, 70% 1×30, integrity 32/180, or industry 62/180. Publishing 137/180 as “we beat LoCoMo” would be false.

Why 70% and 11.4% coexist: [locomo-full-recall-dip-why-20260817.md](../../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md). Vendor percents: [published-claims.md](../../benchmarks/published-claims.md). Fair Mem0 knobs: [mem0-harness-audit-2026-08-22.md](./mem0-harness-audit-2026-08-22.md).

## Why leftover covering looks like a product win (and where that is true)

Leftover covering is a real reader bug class: hybrid `/recall` restates the question or a nearby compiler fact while a distinctive stored leftover is the answer. Early shapes (host, advice, how-long, how-often, listed activities) are generic English and would fire on any corpus.

That stopped being the growth path around P45–P53:

- **24** `looks*Query` detectors and **58** `leftoverCovering*Line` helpers in `internal/memory/recall.go`.
- About **4650 / 9228** lines of that file are leftover covering (~50%).
- P28 **113/180** → P53 **137/180** is **+24 in 25 isolated increments**, almost all +1 / −0 named recoveries from the remaining ledger.
- Cycle-closeout has said “isolated leftover covering is saturating” since P36. We kept going anyway.
- Industry lane **62/180** did not move. Covering is a product-hybrid reader, not retrieval quality.

Selection bias is the tell: we did not discover `what did … realize after` from product traffic. We discovered it because `conv-26-q83` was the next miss.

Anti-benchmax checklist from [public-bench-ladder.md](../public-bench-ladder.md):

| Check | This campaign |
| --- | --- |
| Same pins documented | Pass (dataset SHA, tenant, skip-ingest, hybrid on). |
| No dataset speaker/answer special-cases | **Fail in spirit.** Detectors are not named LoCoMo, but they are item-queued. Kill list already forbids LoCoMo-named rules and category dictionaries. |
| Score delta attributed to a product commit | Pass per increment; each commit is one LoCoMo item. |
| Cross-judge Mem0 numbers labeled incomparable | Pass in docs. Do not regress that honesty. |

## Remaining 43 on P53 (why 162/180 is not a covering problem)

Ledger: `docs/benchmarks/artifacts/failure-ledger/locomo-s0-diag-mh-135-p53-product-recall-s1-717510.jsonl`. **43 misses:** RETRIEVAL 17, PROOF 14, READER 7, WRITE 4, HARNESS 1. Groups: MH 15, SH 13, temporal 8, OD 7.

Covering can only recover **stored distinctive leftover vs restatement**. Optimistic remaining covering yield on this 180 is **0–4 items** if we keep fishing. 137+4 = **141**, still **21 short of 162**. The rest is a different product:

| Class | Examples | Mechanism (PoR) | Covering? |
| --- | --- | --- | --- |
| Gold not compiled | Xenoblade, summer-2022 game start, Witcher six months, Caroline body/friends, yoga 2020, Phuket diving, Voyageurs, Tim visualizing | **S1 compiler WRITE** (needs re-ingest) | No |
| Count / enumerate dump | Melanie children 7 vs 3; Andrew pets 13 vs 1/3; John ankle 38 vs 2 | **S2 entity-scoped enumerate** | No |
| Incomplete lists | LGBTQ participating, outdoor colleagues, dog items, charity beneficiaries, food suggestions, meals, who-told | **S2 enumerate + hops** | No |
| OD hypothetical / coincidence | LGBTQ member, move home, Christmas wedding, Good Sports, extra exercises, studied together | **S2b OD synthesis from facts** | No |
| Relative / window dates | Sunday before 25 May; weekend after 3 June; Oct 19–24 accident | Dated facts + relative calendar, **not** invent-Sunday | No |
| Steal-slot / identity | Dave’s wind-in-hair; Deborah connected-to-body; German vs Spanish; October→September pottery; John’s skateboard for James | Must **not** “fix” | No |
| Peaceful leftover | `conv-44-q62` gold leftover stored; pred restates the about-clause | Generic novelty-vs-restatement **maybe**; a peaceful/nature dictionary **no** | Refused as P54 |
| Camping peaceful | `conv-41-q145` gold not stored | Write, not covering | No |
| Harness | `conv-48-q116` timeout | Harness, not a LoCoMo rule | No |

MH is still **18/33** (held since P21). OD is still **4/11**. Temporal **30/38** is already high on this skip-ingest store; remaining temporal is write/relative-date, not another leftover shape. n=1540 MH is **7.4%**. Do not treat 18/33 as n=1540 MH.

## What “the product truly beats benchmarks” is allowed to mean

Per [sota-execution-plan.md](../sota-execution-plan.md) and the cycle-closeout SOP:

1. **Industry-format LoCoMo** is n=1540, cats 1–4, ingest → search → LLM answer → LLM judge, usually **top-k 200**. That is our **industry-search** lane, not leftover covering on `/recall`. Last industry pin on this 180 is **62/180**. Last full product `/recall` is **11.4%**. Mem0 Platform **92.5%** is that industry format on *their* harness.
2. **Same-pin lead/trail** requires the same dataset SHA, judge, answerer, and question set. The frozen Mem0 1×30 is **handicapped** (v2 search, chunk 8, no timestamps, top-k 30). Fair Mem0 180 waits on quota **2026-09-01**. Until that lands, **do not** say we beat Mem0 from 21/30, 137/180, or 70%.
3. **LME** last **4/20**. SuperMemory 95% is Recall@15, not an LLM-judge percent. LME-500 is not a quality claim while LME-20 is embarrassing.
4. **OpMem 13/13** and marketing **17/17** are real product leads. Keep them. They are not LoCoMo.

The accepted product bet is still [sota-representation-path.md](../sota-representation-path.md): compile durable facts, retrieve those, episodes as provenance. Leftover covering is a reader patch over thin compilation. It cannot close the representation gap (July 49.8% vs Mem0 92.5% on n=1540).

## Stop / continue

**Stop**

- New `looks*Query` leftover-covering detectors queued from this 180’s remaining ledger.
- Category dictionaries (peaceful/nature, charity/self-care, hard-work/goals, …).
- LoCoMo-named product rules.
- Treating 137/180 as 80% / 90% / n=1540 / Mem0 same-pin.
- P54 `conv-44-q62` covering.
- Invent-Sunday, steal-slots, German-vs-Spanish, October→September.

**Continue (generic product, measured)**

1. **S2 enumerate / counts** — entity-scoped how-many and list completion. Transfers off this 180. `filterChildCountItems` already exists and still answers 7 for Melanie; do not add a Melanie-only rule.
2. **S1 compiler WRITE** — re-ingest (skip-ingest 180 cannot measure write). Attacks Xenoblade / dated starts / OD facts.
3. **Industry lane on current SHA** — stratified, top-k 200, tokens reported. This is the lane that can be compared to Mem0’s protocol.
4. **Fair Mem0 180** after 2026-09-01, then same-pin language only.
5. Full n=1540 **only at S6**.

Existing leftover covering stays. Do not rip it out in this cycle. Do not grow it item-by-item.

## Evidence

- P53 ledger: 43 remaining misses, SH PROOF 8.
- `internal/memory/recall.go`: 24 query detectors, 58 covering-line helpers, ~4650 covering lines.
- Industry 62/180 vs product hybrid 137/180 on the same tenant.
- Full `/recall` 11.4% never remeasured on the covering SHA.
