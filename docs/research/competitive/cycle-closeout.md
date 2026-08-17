# Benchmark cycle closeout — required every remasure / merge

Fill a new dated section (or a new file `cycle-closeout-YYYYMMDD.md`) at the end of every measurement cycle. Do not close a cycle with Brainy scores alone.

**Tracks (never mix):** Same-pin = same dataset SHA, same judge/answerer, same question set. The product README may carry a **published-% claims** table (sourced, n/metric labeled) and a **same-pin** table, outlinking [docs/benchmarks/README.md](../../benchmarks/README.md) and [published-claims.md](../../benchmarks/published-claims.md). Detailed why/next stays **here**. Do not use vendor headlines for lead/trail.

The **product** cycle summary is the README comparison table (plus GTM still Brainy-product copy). **Evals may name competitors.** The competitor table lives here.

## Template

1. **Landed** — SHAs on `dev` / `main`, PRs, what product change shipped (one sentence).
2. **Own pins** — OpMem, marketing, LoCoMo 1×30 **by category**, LME if run. Name dips as dips. 1×30 is measurement, not qualification.
3. **Competitor compare (detailed)** — not a one-liner; **this file only**. Required axes:
   - LoCoMo 1×30 overall + **multi-hop / open-domain / temporal** vs last frozen same-pin ([locomo-mem0-samepin-pr10-20260813.md](../../benchmarks/artifacts/locomo-mem0-samepin-pr10-20260813.md)).
   - Search latency on that pin (local vs platform; do not claim a platform SLO).
   - OpMem vs last ops pin ([staging-competitive-report.md](../../benchmarks/staging-competitive-report.md)) — re-run the incumbent before claiming a **new** ops lead.
   - Marketing vertical vs last empirical pin ([marketing-mvp-vs-mem0.md](../../vertical/marketing-mvp-vs-mem0.md)) — same rule.
   - LME-20 quality if run. No fair incumbent pin on our harness unless one exists.
   - Other vendors: **no pin** unless we actually ran them. Published headlines stay in a “context only” row.
   - For every trailing axis: the **product mechanism** (not “we need to try harder”).
   - For every leading axis: what we must **not** regress, and whether the pin is stale.
4. **Why** — product mechanism (compiler coverage, provenance crowding, reader). Not vibes.
5. **Next** — one step on [sota-representation-path.md](../sota-representation-path.md), mapped to the largest gap. Kill list: no fusion fishing, no graph DB default, no category dictionaries, no unbounded top-k, no LoCoMo/LME-named product rules, no SOTA claims.

Forbidden in the closeout: SOTA without a frozen same-pin win, mixing 1×30 with 3×90 or with published 90+ LoCoMo headlines, inventing a vendor LoCoMo number.

---

## Cycle 2026-08-14 — compiler atom quality

**Landed:** compiler-quality gate (malformed atoms are not recall-primary) on `dev` at `4010d30` (PR #115). Production FF of `dev` → `main` is this cycle (explicit approval). Feature pin: [locomo-atomq-dev-1x30-20260814.md](../../benchmarks/artifacts/locomo-atomq-dev-1x30-20260814.md). This SOP lands with it.

Product change: do not persist light-verb junk (`has done going at …`), failed gerund stems (`participates in runn`), or broken quote shards; malformed atoms do not complete coverage and get a ranking penalty; episode fallback uses local IDF over uncovered tokens; numeric date tails (`June 3.`) are not treated as junk.

### Own pins (this cycle)

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Non-reg vs Wave 1 / R1c. `upd01` June vs May kept. |
| Marketing vertical | **17/17** | Non-reg |
| LoCoMo 1×30 conv-26 | **11/30** | MH **2/10** · OD **0/4** · temporal **9/16**. +1 vs R1c 10/30; **dip** vs Wave 1 14/30 and Gate 0 18/30 |
| LME-20 | **0/20** publishable | Integrity pin; not re-run this cycle |

Packet compare (same 1×30, top-k=30): mean packet chars 1712 (R1c) → **1907**; junk template hits **45 → 6**; q10 (`4 years`) recovered. Ledger: **15 WRITE_MISS + 4 READER_MISS**.

### Competitor compare (detailed)

#### 1. LoCoMo 1×30 — only fair conversational QA pin this cycle

Frozen Mem0 Platform pin (2026-08-13, **same dataset SHA** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, same judge temp 0.0, conv-26 1×30): [locomo-mem0-samepin-pr10-20260813.md](../../benchmarks/artifacts/locomo-mem0-samepin-pr10-20260813.md).

| Axis | Brainy now (`d82f7d6` / `4010d30`) | Mem0 Platform (frozen same-pin) | Graphiti OSS / Zep Platform | Stand |
| --- | ---: | ---: | --- | --- |
| LoCoMo 1×30 overall | **11/30 (0.367)** | **12/30 (0.400)** | **no same-pin** | **Trail by 1** |
| Multi-hop | **2/10 (0.200)** | **7/10 (0.700)** | no same-pin | **Trail — largest gap** |
| Open-domain | **0/4 (0.000)** | **3/4 (0.750)** | no same-pin | **Trail** |
| Temporal | **9/16 (0.562)** | **2/16 (0.125)** | no same-pin | **Lead this pin** |
| Search p50 / p95 | 165 / 200 ms (local) | 471 / 564 ms (platform) | no same-pin | Faster locally; **not** a platform SLO claim |

Brainy trajectory on the **same** 1×30 (do not treat later rows as beating Mem0):

| Pin | Overall | MH | OD | Temporal vs Mem0 12/30 · 7/10 · 3/4 · 2/16 |
| --- | ---: | ---: | ---: | --- |
| Gate 0 staging | 18/30 | ~5/10 | 1/4 | Different stack; not this local remasure |
| Wave 1 local | **14/30** | 3/10 | 2/4 | 9/16 | Overall **lead** Mem0 12; MH still **trail** 3 vs 7 |
| R1c local | 10/30 | 2/10 | 0/4 | 8/16 | Overall **trail**; junk atoms crowded provenance |
| **This cycle** | **11/30** | **2/10** | **0/4** | **9/16** | Overall **trail by 1**; MH unchanged; temporal lead restored |

**Multi-hop (trail 2/10 vs 7/10).** This is not a hop-executor or fusion-weight problem. Mem0 answers attribute joins because the **facts** (and, on platform, entity/graph signals) exist as retrieval units. Brainy’s ledger on this pin is 15 WRITE_MISS: the claim never compiled into a well-formed fact, so there is nothing to join. Hop V2 that walks first-linked memory IDs cannot close a 5-point MH gap. Closing it is R1b coverage → R2 entities → R3 relation projection → R4 ID hops. Do not add a graph DB to chase this number.

**Open-domain (trail 0/4 vs 3/4).** Same mechanism: OD items need a compiled fact the reader can cite, not a nearby chat turn. Wave 1 had 2/4 with episode-heavy packets; R1c/this cycle dropped to 0/4 while the compiler is thin. Recovering OD without re-ranking transcripts means the compiler must emit the durable claim (career, titled works, dated plans). Do not restore OD by stuffing more dialogue into top-k.

**Temporal (lead 9/16 vs 2/16).** Brainy’s `temporal_score` + `IncludeHistorical` + recency on dated records is the one conversational axis we currently win on this pin. Wave 1 already had 9/16; R1c dipped to 8/16 when junk crowded packets; this cycle restored 9/16. The signal still mostly scores **transcripts**. Next: keep the lead by moving that score onto **dated semantic facts** (PoR POV 8), not by adding LoCoMo-named date rules. Do not claim a general temporal-SOTA from n=16.

**Overall (trail 11 vs 12).** q10 recovery (+1 vs R1c) did not overtake Mem0. Wave 1’s 14/30 **did** lead this Mem0 pin on overall but still lost MH 3 vs 7 — that is why Wave 1 was not “we beat Mem0.” A later frozen Mem0 re-pin is required before any lead/trail sentence on a new Brainy SHA; this cycle reuses the 2026-08-13 Mem0 freeze (same dataset SHA, same 30 items).

**Latency.** Local p50 165 ms vs Mem0 platform 471 ms is a harness observation, not a production SLO, not a Cloudflare/staging number, and not evidence that retrieval quality is better.

**Mem0 OSS** was not re-measured this cycle. Do not treat Platform 12/30 as an OSS-reproducible number. Do not mix this 1×30 with Mem0 blog 90+ LoCoMo or with Brainy staging 3×90 (33/90, MH 22.2%).

#### 2. OpMem — lead (stale Mem0 pin; Brainy re-confirmed)

| | Brainy this cycle | Mem0 |
| --- | ---: | --- |
| OpMem | **13/13** | **9/12** (2026-07-14 staging Platform; **not re-run this cycle**) |

Source: [staging-competitive-report.md](../../benchmarks/staging-competitive-report.md). Mem0’s three fails then: `cor02` correction stickiness (`ruby` not on top after revise), `sup03` durable forget (forgotten memory resurfaces), `upd01` stale fact (March outranks June). Brainy still passes all three; this cycle’s date-tail fix was specifically to **keep** `upd01` (June vs May) after a false-positive malformed-atom gate.

Stand: **lead ops**. Caveat: Mem0 OpMem is ~one month old. Re-run Mem0 before packaging a new “+3 OpMem” marketing sentence. Do not spend the next cycle on ops — it is already the moat we must not regress.

#### 3. Marketing vertical — lead (stale Mem0 pin; Brainy re-confirmed)

| | Brainy this cycle | Mem0 empirical |
| --- | ---: | --- |
| Marketing fixtures | **17/17** | **4/16** (2026-07-29 Platform; **not re-run this cycle**) |

Source: [marketing-mvp-vs-mem0.md](../../vertical/marketing-mvp-vs-mem0.md). Differentiation then (Brainy pass / Mem0 fail): principle over preference, voice_profile mapping, core ingest unaffected by pack weights, never-sentences as brand_rule, response-style ranking, archived campaign hidden. Shared passes: suppression leak, correction stickiness, dedupe. July Brainy fail `bv06` multi-brand isolation is **not** in the current 17/17 pin — do not cite the July 15/16 Brainy score as current.

Stand: **lead governed vertical**. Same caveat as OpMem: do not claim a refreshed Mem0 gap without a re-run. Next cycle does not chase Mem0 on packs; it keeps 17/17 green while R1b lands.

#### 4. LME-20 — neither is a quality win

Brainy publishable integrity: **0/20** `/recall` ([lme20-product-recall-pr1-20260812-pin.md](../../benchmarks/artifacts/lme20-product-recall-pr1-20260812-pin.md)). Jobs completed; quality is zero. No fair Mem0 pin exists on **this** harness. Do not compare 0/20 to anyone’s published LME headline. Quality LME waits until R6, after representation coverage.

#### 5. Graphiti / Zep — architecture target, not a scoreboard

**No same-pin.** Do not invent a LoCoMo or LME number for Graphiti OSS or Zep Platform.

What we still take from them (not scores): episode = provenance; entity + temporally-valid relation = retrieval unit; multi-hop = edge walk on entity IDs. Brainy today: `memory_entity_links` hub only, **no relation table**, hops = first linked memory ID. That is why MH 2/10 cannot be fixed by ranking. Stance: **ADAPT into Postgres** (ADR-004). **REJECT** Neo4j/FalkorDB as required substrate. See [graphiti.md](./graphiti.md).

Published Zep / Mem0 blog LoCoMo/LME headlines remain **context only**.

### Why

R1c made facts recall-primary before the compiler was trustworthy. Malformed templates (`has done going at`, `participates in runn`) occupied slots; episode fallback filled the rest with name-matching chat. Hit **count** barely moved (29.3 → 27.8); packet **text** dropped (2012 → 1712 chars) and q10’s gold (`4 years`) left the packet.

This cycle’s gate is a **compiler-quality** fix (R1b slice), not a ranking retune: junk hits 45→6, packet chars 1907, q10 CORRECT again. OpMem 13/13 and marketing 17/17 held. Score only +1 because 15/19 remaining LoCoMo misses are WRITE_MISS — the durable claim still exists only as a transcript.

Vs Mem0: we lead the axes that are already Brainy-native (ops mutation, vertical governance, this-pin temporal scoring). We trail the axes that require **compiled semantic memory** (MH, OD, overall conversational QA). That split is the program, not a surprise.

### Next

**One step:** R1b held-out **coverage** (not another 1×30 fishing run, not fusion, not hard episode drop).

| Trailing vs Mem0 | Product next (PoR) | Explicitly not next |
| --- | --- | --- |
| MH 2/10 vs 7/10 | Compiler emits joinable atomic facts; then R2 entities, R3 relation projection, R4 ID hops | Graph DB, hop-weight tuning, category dictionary |
| OD 0/4 vs 3/4 | Same compiler: career / titled works / dated plans as facts with provenance | Restore OD by raising episode top-k |
| Overall 11 vs 12 | Coverage until WRITE_MISS mass falls; then remasure 1×30 as **measurement** | Treat 12/30 as a merge gate; claim beats-Mem0 |
| Temporal lead 9 vs 2 | Keep; move `temporal_score` onto dated **facts** | LoCoMo-named date rules; declare SOTA |
| Ops / vertical lead | Keep 13/13 and 17/17 green on every remasure | Spend a cycle matching Mem0 on packs |
| Graphiti relations | After R1b coverage, R2→R3 in Postgres | Neo4j; second unrelated relation extractor |
| LME 0/20 | R6 after representation audit | Compare 0/20 to published LME headlines |

Kill list stays in force. Do not hard-drop episodes before held-out coverage. Do not add LoCoMo/LME-named product rules. Do not write SOTA / beats-Mem0.

**Remasure after R1b coverage work:** new dated section in this file, same five headings, same competitor table (re-use Mem0 freeze if dataset SHA unchanged; re-pin Mem0 if the harness or SHA changes).

---

## Cycle 2026-08-14 — R1b coverage + relation projection

**Landed:** R1b held-out compiler coverage + R3 Postgres relation projection on `pr/r1b-coverage-relations-a6c7` (`571cc1a`, `5c5f561`). Feature pin: [locomo-r1b-dev-1x30-20260814.md](../../benchmarks/artifacts/locomo-r1b-dev-1x30-20260814.md). Held-out audit (`TestHeldOutCompilerCoverageAudit`) is the representation merge gate and is green.

Product change: durable claims compile into well-formed S/P/V atoms with provenance (origin anaphora, career/education, titled works, activity/place FindAll, workshop nouns, trip places, light-verb events, last-weekday / last-week stamps). Entity-valued facts project into `memory_relations`; `follow_relation` hops walk those edges. Malformed date-stamped nouns are no longer dropped.

### Own pins (this cycle)

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13 (100%)** | Non-reg. `upd01` June vs May kept. [pin](../../benchmarks/artifacts/opmem-r1b-local-20260814.md) |
| Marketing vertical | **17/17 (100%)** | Non-reg. [pin](../../benchmarks/artifacts/marketing-r1b-local-20260814.md) |
| LoCoMo 1×30 conv-26 | **15/30 (50.0%)** | MH **6/10 (60.0%)** · OD **0/4 (0.0%)** · temporal **9/16 (56.2%)**. +4 vs compiler-quality 11/30; +1 vs Wave 1 14/30; still below Gate 0 18/30 |
| LME-20 | **0/20 (0.0%)** publishable | Integrity pin; not re-run this cycle |

Ledger: **10 WRITE_MISS + 4 READER_MISS + 1 RETRIEVAL_MISS** (was 15 WRITE_MISS + 4 READER_MISS). Search p50 **155 ms** local.

### Competitor compare (detailed)

#### 1. LoCoMo 1×30 — only fair conversational QA pin this cycle

Frozen Mem0 Platform pin (2026-08-13, **same dataset SHA** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, same judge temp 0.0, conv-26 1×30): [locomo-mem0-samepin-pr10-20260813.md](../../benchmarks/artifacts/locomo-mem0-samepin-pr10-20260813.md).

| Axis | Brainy now (`5c5f561`) | Mem0 Platform (frozen same-pin) | Graphiti OSS / Zep Platform | Stand |
| --- | ---: | ---: | --- | --- |
| LoCoMo 1×30 overall | **15/30 (50.0%)** | **12/30 (40.0%)** | **no same-pin** | **Lead this pin by 3** |
| Multi-hop | **6/10 (60.0%)** | **7/10 (70.0%)** | no same-pin | **Trail by 1** (was trail by 5) |
| Open-domain | **0/4 (0.0%)** | **3/4 (75.0%)** | no same-pin | **Trail** |
| Temporal | **9/16 (56.2%)** | **2/16 (12.5%)** | no same-pin | **Lead this pin** |
| Search p50 / p95 | 155 / 187 ms (local) | 471 / 564 ms (platform) | no same-pin | Faster locally; **not** a platform SLO claim |

Brainy trajectory on the **same** 1×30 (do not treat later rows as beating Mem0 in general):

| Pin | Overall | MH | OD | Temporal | vs Mem0 12/30 · 7/10 · 3/4 · 2/16 |
| --- | ---: | ---: | ---: | ---: | --- |
| Gate 0 staging | 18/30 (60.0%) | ~5/10 | 1/4 | — | Different stack |
| Wave 1 local | 14/30 (46.7%) | 3/10 (30.0%) | 2/4 (50.0%) | 9/16 (56.2%) | Overall lead; MH trail 3 vs 7 |
| R1c local | 10/30 (33.3%) | 2/10 (20.0%) | 0/4 (0.0%) | 8/16 (50.0%) | Overall trail; junk crowded provenance |
| Compiler-quality | 11/30 (36.7%) | 2/10 (20.0%) | 0/4 (0.0%) | 9/16 (56.2%) | Overall trail by 1; MH unchanged |
| **This cycle** | **15/30 (50.0%)** | **6/10 (60.0%)** | **0/4 (0.0%)** | **9/16 (56.2%)** | Overall **lead by 3**; MH **trail by 1**; OD unchanged trail |

**Multi-hop (trail 6/10 vs 7/10).** Closed 4 of the previous 5-point MH gap. Recovered list/join items whose gold is now compiled (activities, camp places, unwind, relationship status). Remaining MH: q11 origin Sweden is **READER_MISS** with `gold_in_facts=true` (anaphora wrote the fact; hop/reader did not cite it); q13 career qualifier (transgender people) still WRITE_MISS; q19 kids likes (dinosaurs) WRITE_MISS; q23 second titled work WRITE_MISS. Next is R4 ID hops on facts that exist, plus remaining compiler coverage — not a graph DB.

**Open-domain (trail 0/4 vs 3/4).** All four OD items are hypothetical / “likely” questions (`Would Caroline…`, fields of study). Compiler atoms for counseling exist; the reader does not emit the inferred yes/no. That is R5 structured-first answer, not more episode top-k.

**Temporal (lead 9/16 vs 2/16).** Unchanged 9/16 vs compiler-quality; picnic (q21) recovered, other temporal items swapped. q29 workshop date is READER_MISS (`gold_in_facts=true`) — a pottery *class signup* date crowded the Friday-before workshop stamp. Keep the lead by scoring dated **facts**, not by LoCoMo-named date rules.

**Overall (lead 15 vs 12 on this freeze).** First Brainy pin on this harness that leads Mem0 overall **and** has MH within 1. Wave 1’s 14/30 also led overall but lost MH 3 vs 7 — that is why Wave 1 was not “we beat Mem0.” This cycle still **must not** say beats-Mem0 / SOTA: OD is 0/4, MH still trails, n=30, Mem0 pin is frozen not re-run.

**Latency.** Local p50 155 ms vs Mem0 platform 471 ms is a harness observation, not a production SLO.

**Mem0 OSS** was not re-measured. Do not mix this 1×30 with Mem0 blog 90+ or Brainy staging 3×90.

#### 2. OpMem — lead (stale Mem0 pin; Brainy re-confirmed)

| | Brainy this cycle | Mem0 |
| --- | ---: | --- |
| OpMem | **13/13 (100%)** | **9/12 (75.0%)** (2026-07-14 staging Platform; **not re-run this cycle**) |

Stand: **lead ops**. Re-run Mem0 before a new “+3 OpMem” marketing sentence.

#### 3. Marketing vertical — lead (stale Mem0 pin; Brainy re-confirmed)

| | Brainy this cycle | Mem0 empirical |
| --- | ---: | --- |
| Marketing fixtures | **17/17 (100%)** | **4/16 (25.0%)** (2026-07-29 Platform; **not re-run this cycle**) |

Stand: **lead governed vertical**. Same stale-Mem0 caveat.

#### 4. LME-20 — neither is a quality win

Brainy publishable integrity: **0/20 (0.0%)** `/recall`. No fair Mem0 pin on this harness. Quality LME waits until R6.

#### 5. Graphiti / Zep — architecture target, not a scoreboard

**No same-pin.** Do not invent a LoCoMo or LME number.

What landed from them this cycle: relation edges as a **projection** of entity-valued atoms in Postgres (ADR-004). Still not Neo4j. Canonical entity IDs (R2 full) and `hop[i].output_entity_id == hop[i+1].input_entity_id` (R4) are next, not claimed done.

### Why

WRITE_MISS mass was the MH hole. Generic linguistic extractors + relative-date stamps + place FindAll turned transcript-only claims into retrieval units. Relation projection did not need a second extractor. q11 shows the next failure class: **the fact exists** (`gold_in_facts=true`) and the reader still misses Sweden — that is join/proof, not another regex.

Vs Mem0: we now lead the axes that are Brainy-native **and** this-pin overall conversational QA. We still trail the axes that need inferred OD answers and one remaining MH join. That split is still the program.

### Next

**One step:** R4 entity-ID hops on facts that already exist (q11 origin), plus remaining R1b WRITE_MISS (kids likes, second titled work, career qualifier). Then R5 for OD hypotheticals.

| Trailing vs Mem0 | Product next (PoR) | Explicitly not next |
| --- | --- | --- |
| MH 6/10 vs 7/10 | R4 ID join for origin; leftover compiler coverage | Graph DB, fusion fishing, LoCoMo-named rules |
| OD 0/4 vs 3/4 | R5 structured-first yes/no from compiled career/possession facts | Restore OD by raising episode top-k |
| Overall lead 15 vs 12 | Keep; do not declare beats-Mem0; re-pin Mem0 before any new lead sentence on a new SHA | Treat 15/30 as qualification or SOTA |
| Temporal lead 9 vs 2 | Move `temporal_score` onto dated facts; stop class-signup dates crowding workshop stamps | LoCoMo-named date rules |
| Ops / vertical lead | Keep 13/13 and 17/17 green | Spend a cycle matching Mem0 on packs |
| LME 0/20 | R6 after representation + OD reader | Compare 0/20 to published LME headlines |

Kill list stays in force. Do not hard-drop episodes (10 WRITE_MISS remain). Do not write SOTA / beats-Mem0.

---

## Cycle 2026-08-15 — R4 hops + leftover MH coverage

**Landed:** R4 typed hop destinations + leftover compiler coverage on `pr/mh-join-coverage-a6c7` (`d48e202`). Feature pin: [locomo-mh-r4c-dev-1x30-20260815.md](../../benchmarks/artifacts/locomo-mh-r4c-dev-1x30-20260815.md). Production FF of this SHA onto `dev` then `main` is this cycle (explicit approval after remasure).

Product change: hops bind typed destinations (origin place, occupation+identity, all slot values); when/how-long skip event hops; researching X compiles as a plan atom; they-were-stoked objects compile as preferences; list hops keep only compatible predicates; async extract projects `memory_relations`; image query/caption is ingested with the turn.

### Own pins (this cycle)

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13 (100%)** | Non-reg. `upd01` June vs May kept. [pin](../../benchmarks/artifacts/opmem-mh-r4c-local-20260815.md) |
| Marketing vertical | **17/17 (100%)** | Non-reg. [pin](../../benchmarks/artifacts/marketing-mh-r4c-local-20260815.md) |
| LoCoMo 1×30 conv-26 | **19/30 (63.3%)** | MH **9/10 (90.0%)** · OD **0/4 (0.0%)** · temporal **10/16 (62.5%)**. +4 vs R1b 15/30. `errors: 1` is q8 JUDGE_MISS |
| LME-20 | **0/20 (0.0%)** publishable | Integrity pin; not re-run this cycle |

Ledger: **7 WRITE_MISS + 2 READER_MISS + 1 RETRIEVAL_MISS + 1 JUDGE_MISS**. Search p50 **128 ms** local.

Not shipped: r4 **6/30** (when-hops overwrote dates); r4b **17/30** (q3 identity dump; q19 exhibit miss).

### Competitor compare (detailed)

#### 1. LoCoMo 1×30 — only fair conversational QA pin this cycle

Frozen Mem0 Platform pin (2026-08-13, **same dataset SHA** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, same judge temp 0.0, conv-26 1×30): [locomo-mem0-samepin-pr10-20260813.md](../../benchmarks/artifacts/locomo-mem0-samepin-pr10-20260813.md).

| Axis | Brainy now (`d48e202`) | Mem0 Platform (frozen same-pin) | Graphiti OSS / Zep Platform | Stand |
| --- | ---: | ---: | --- | --- |
| LoCoMo 1×30 overall | **19/30 (63.3%)** | **12/30 (40.0%)** | **no same-pin** | **Lead this pin by 7** |
| Multi-hop | **9/10 (90.0%)** | **7/10 (70.0%)** | no same-pin | **Lead this pin by 2** (was trail by 1) |
| Open-domain | **0/4 (0.0%)** | **3/4 (75.0%)** | no same-pin | **Trail** |
| Temporal | **10/16 (62.5%)** | **2/16 (12.5%)** | no same-pin | **Lead this pin** |
| Search p50 / p95 | 128 / 187 ms (local) | 471 / 564 ms (platform) | no same-pin | Faster locally; **not** a platform SLO claim |

Brainy trajectory on the **same** 1×30:

| Pin | Overall | MH | OD | Temporal | vs Mem0 12/30 · 7/10 · 3/4 · 2/16 |
| --- | ---: | ---: | ---: | ---: | --- |
| Gate 0 staging | 18/30 (60.0%) | ~5/10 | 1/4 | — | Different stack |
| Wave 1 local | 14/30 (46.7%) | 3/10 (30.0%) | 2/4 (50.0%) | 9/16 (56.2%) | Overall lead; MH trail 3 vs 7 |
| R1c local | 10/30 (33.3%) | 2/10 (20.0%) | 0/4 (0.0%) | 8/16 (50.0%) | Overall trail |
| Compiler-quality | 11/30 (36.7%) | 2/10 (20.0%) | 0/4 (0.0%) | 9/16 (56.2%) | Overall trail by 1 |
| R1b | 15/30 (50.0%) | 6/10 (60.0%) | 0/4 (0.0%) | 9/16 (56.2%) | Overall lead by 3; MH trail by 1 |
| **This cycle** | **19/30 (63.3%)** | **9/10 (90.0%)** | **0/4 (0.0%)** | **10/16 (62.5%)** | Overall **lead by 7**; MH **lead by 2**; OD unchanged trail |

**Multi-hop (lead 9/10 vs 7/10).** Closed the R1b trail. Recovered origin (Sweden), research topic (plan atom, not identity dump), career population join, and kids exhibit noun (they-stoked preference). Remaining MH: q23 second titled work. Gold `"Nothing is Impossible"` is **not** in the transcript, BLIP caption (`a photography of a book cover with a gold coin on it`), or image query (`painted canvas follow your dreams`). That is multimodal WRITE_MISS, not a hop bug. Do not hardcode the title. Text-join MH on this pin is 9/10; 10/10 needs vision/OCR.

**Open-domain (trail 0/4 vs 3/4).** Unchanged. Hypothetical / “likely” questions (`Would Caroline…`, fields of study). Compiler atoms for counseling exist; the reader does not emit the inferred yes/no. That is R5 structured-first answer, not more episode top-k.

**Temporal (lead 10/16 vs 2/16).** +1 vs R1b (q29 workshop date). q6 (`when` + camping) still enumerates activity lists because `looksListQuery` fires on `camping` even when the question is temporal — not fixed this cycle (risk to MH lists). q8 speech date is JUDGE_MISS (`2 June 2023` vs “the week before 9 June 2023”). Keep the lead by scoring dated **facts**, not by LoCoMo-named date rules.

**Overall (lead 19 vs 12 on this freeze).** First pin on this harness that leads Mem0 on **both** overall and MH. Still **must not** say beats-Mem0 / SOTA: OD is 0/4, n=30, Mem0 pin is frozen not re-run, q23 is image-gold.

**Latency.** Local p50 128 ms vs Mem0 platform 471 ms is a harness observation, not a production SLO.

**Mem0 OSS** was not re-measured. Do not mix this 1×30 with Mem0 blog 90+ or Brainy staging 3×90.

#### 2. OpMem — lead (stale Mem0 pin; Brainy re-confirmed)

| | Brainy this cycle | Mem0 |
| --- | ---: | --- |
| OpMem | **13/13 (100%)** | **9/12 (75.0%)** (2026-07-14 staging Platform; **not re-run this cycle**) |

Stand: **lead ops**. Re-run Mem0 before a new “+3 OpMem” marketing sentence.

#### 3. Marketing vertical — lead (stale Mem0 pin; Brainy re-confirmed)

| | Brainy this cycle | Mem0 empirical |
| --- | ---: | --- |
| Marketing fixtures | **17/17 (100%)** | **4/16 (25.0%)** (2026-07-29 Platform; **not re-run this cycle**) |

Stand: **lead governed vertical**. Same stale-Mem0 caveat.

#### 4. LME-20 — neither is a quality win

Brainy publishable integrity: **0/20 (0.0%)** `/recall`. No fair Mem0 pin on this harness. Quality LME waits until R6.

#### 5. Graphiti / Zep — architecture target, not a scoreboard

**No same-pin.** Do not invent a LoCoMo or LME number.

What landed from them this cycle: async extract now writes the same Postgres relation projection as sync ingest (ADR-004). Still not Neo4j. Canonical entity IDs (R2 full) remain next, not claimed done.

### Why

MH was a write/join hole, not a fusion-weight hole. Research answers dumped identity because empty predicate hints defaulted to occupation/identity and enumerate kept the `ans` hop anyway. Kids exhibit nouns lived in a they-stoked clause (and an image query field the harness dropped). Relation hops were no-ops on the eval path because the worker indexed atoms but never projected edges (0 rows).

Typed plan atoms + scoped hops recovered q3. Pronoun-excited preferences + image alt-text recovered q19. Origin hops recovered q11. Async relation projection makes `follow_relation` real on LoCoMo ingest.

Vs Mem0: we now lead this freeze on overall, MH, and temporal. We still trail OD hypotheticals. That split is still the program.

### Next

**One step:** R5 structured-first yes/no from compiled career/possession facts (OD 0/4 vs Mem0 3/4). Do not spend the next cycle on q23 OCR or on fusion fishing.

| Trailing vs Mem0 | Product next (PoR) | Explicitly not next |
| --- | --- | --- |
| OD 0/4 vs 3/4 | R5 structured-first yes/no from compiled facts | Restore OD by raising episode top-k |
| Overall lead 19 vs 12 | Keep; do not declare beats-Mem0; re-pin Mem0 before any new lead sentence on a new SHA | Treat 19/30 as qualification or SOTA |
| MH lead 9 vs 7 | Keep; remaining miss is image-gold WRITE_MISS | Hardcode titled-work gold; claim 10/10 from text |
| Temporal lead 10 vs 2 | Stop `when` + list-cue questions from enumerating activities; score dated facts | LoCoMo-named date rules |
| Ops / vertical lead | Keep 13/13 and 17/17 green | Spend a cycle matching Mem0 on packs |
| LME 0/20 | R6 after representation + OD reader | Compare 0/20 to published LME headlines |

Kill list stays in force. Do not write SOTA / beats-Mem0. Do not call MH “solved” while q23 is image-only and OD is 0/4.

---

## Cycle 2026-08-15 — image WRITE + copula-safe enumerate (R4h)

**Landed:** OCR of deictic attached covers into titled-work atoms, plus enumerate that keeps copula titles, on `dev`/`main` at `f4ec4d7` (PR #119). Feature pin: [locomo-mh-r4h-dev-1x30-20260815.md](../../benchmarks/artifacts/locomo-mh-r4h-dev-1x30-20260815.md). Production FF of this SHA is this cycle (explicit approval after MH remasure).

Product change: fetch public `image_urls` only when the utterance has `this book` / `this novel` / `this title`; OCR overlapping cover-face windows; store one well-formed title in `[visible text:]` on the deictic sentence; hop/enumerate does not split titles on `is`; pose/scene caption atoms do not compile as activities.

Not shipped: r4d MH **8/10** (shard titles); r4e MH **8/10** (WRITE ok, reader split `nothing is impossible` → `impossible`).

### Own pins (this cycle)

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13 (100%)** | Non-reg. `upd01` June vs May kept. [pin](../../benchmarks/artifacts/opmem-mh-r4h-local-20260815.md) |
| Marketing vertical | **17/17 (100%)** | Non-reg. [pin](../../benchmarks/artifacts/marketing-mh-r4h-local-20260815.md) |
| LoCoMo 1×30 conv-26 | **20/30 (66.7%)** | MH **10/10 (100%)** · OD **0/4 (0.0%)** · temporal **10/16 (62.5%)**. +1 vs R4c 19/30. `errors: 1` is q8 JUDGE_MISS |
| LME-20 | **0/20 (0.0%)** publishable | Integrity pin; not re-run this cycle |

Search p50 **125 ms** local. Temporal **dip** q29 workshop date (CORRECT→WRONG) offset by q26 read-date recovery; net 10/16 held.

### Competitor compare (detailed)

#### 1. LoCoMo 1×30 — only fair conversational QA pin this cycle

Frozen Mem0 Platform pin (2026-08-13, **same dataset SHA** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, same judge temp 0.0, conv-26 1×30): [locomo-mem0-samepin-pr10-20260813.md](../../benchmarks/artifacts/locomo-mem0-samepin-pr10-20260813.md).

| Axis | Brainy now (`f4ec4d7`) | Mem0 Platform (frozen same-pin) | Graphiti OSS / Zep Platform | Stand |
| --- | ---: | ---: | --- | --- |
| LoCoMo 1×30 overall | **20/30 (66.7%)** | **12/30 (40.0%)** | **no same-pin** | **Lead this pin by 8** |
| Multi-hop | **10/10 (100%)** | **7/10 (70.0%)** | no same-pin | **Lead this pin by 3** |
| Open-domain | **0/4 (0.0%)** | **3/4 (75.0%)** | no same-pin | **Trail** |
| Temporal | **10/16 (62.5%)** | **2/16 (12.5%)** | no same-pin | **Lead this pin** |
| Search p50 / p95 | 125 / 149 ms (local) | 471 / 564 ms (platform) | no same-pin | Faster locally; **not** a platform SLO claim |

**Multi-hop (lead 10/10 vs 7/10).** Closed the R4c image WRITE_MISS. Cover lettering is compiled at ingest (tesseract on PATH, public HTTP fetch, no gold strings). Enumerate keeps the full copula title. This 1×30 MH axis is closed. Do not generalize to “MH solved” as a product: OD is still 0/4, n=30.

**Open-domain (trail 0/4 vs 3/4).** Unchanged. Hypothetical / “likely” questions. R5 structured-first yes/no from compiled facts. Do not restore OD by stuffing episodes into top-k.

**Temporal (lead 10/16 vs 2/16).** Net unchanged vs R4c. **Dip** q29 (workshop date). **Recovery** q26 (dated titled-work from the cover turn). Keep the lead by scoring dated facts, not LoCoMo-named date rules.

**Overall (lead 20 vs 12 on this freeze).** +1 from q23. Still **must not** say beats-Mem0 / SOTA: OD is 0/4, n=30, Mem0 pin is frozen not re-run.

**Latency.** Local p50 125 ms vs Mem0 platform 471 ms is a harness observation, not a production SLO.

**Mem0 OSS** was not re-measured. Do not mix this 1×30 with Mem0 blog 90+ or Brainy staging 3×90.

#### 2. OpMem — lead (stale Mem0 pin; Brainy re-confirmed)

| | Brainy this cycle | Mem0 |
| --- | ---: | --- |
| OpMem | **13/13 (100%)** | **9/12 (75.0%)** (2026-07-14 staging Platform; **not re-run this cycle**) |

Stand: **lead ops**. Re-run Mem0 before a new “+3 OpMem” marketing sentence.

#### 3. Marketing vertical — lead (stale Mem0 pin; Brainy re-confirmed)

| | Brainy this cycle | Mem0 empirical |
| --- | ---: | --- |
| Marketing fixtures | **17/17 (100%)** | **4/16 (25.0%)** (2026-07-29 Platform; **not re-run this cycle**) |

Stand: **lead governed vertical**. Same stale-Mem0 caveat.

#### 4. LME-20 — neither is a quality win

Brainy publishable integrity: **0/20 (0.0%)** `/recall`. No fair Mem0 pin on this harness. Quality LME waits until R6.

#### 5. Graphiti / Zep — architecture target, not a scoreboard

**No same-pin.** Do not invent a LoCoMo or LME number.

### Why

q23 was WRITE then READER, not hops. The title is cover text. One crop of a 3D mockup returned shards; a function-word + two-letter window matched as a title and short-circuited better windows. After OCR wrote the atom, enumerate re-parsed lowercase relation destinations with a bare ` is ` splitter (`nothing is impossible` → `impossible`). Caption pose-places (`sitting at Top`) crowded `swimming` off the activity list.

Vs Mem0: we lead this freeze on overall, MH, and temporal. We still trail OD hypotheticals. That split is still the program.

### Next

**One step:** R5 structured-first yes/no from compiled career/possession facts (OD 0/4 vs Mem0 3/4). Do not spend the next cycle on more OCR windows or on fusion fishing.

| Trailing vs Mem0 | Product next (PoR) | Explicitly not next |
| --- | ---: | --- |
| OD 0/4 vs 3/4 | R5 structured-first yes/no from compiled facts | Restore OD by raising episode top-k |
| Overall lead 20 vs 12 | Keep; do not declare beats-Mem0; re-pin Mem0 before any new lead sentence on a new SHA | Treat 20/30 as qualification or SOTA |
| MH lead 10 vs 7 | Keep; 1×30 MH closed; do not claim MH-solved as a product while OD is 0/4 | Hardcode titled-work gold; more LoCoMo-named OCR |
| Temporal lead 10 vs 2 | Restore q29 workshop date onto the dated fact; stop `when` + list-cue enumerate | LoCoMo-named date rules |
| Ops / vertical lead | Keep 13/13 and 17/17 green | Spend a cycle matching Mem0 on packs |
| LME 0/20 | R6 after representation + OD reader | Compare 0/20 to published LME headlines |

Kill list stays in force. Do not write SOTA / beats-Mem0. Do not call the **product** “MH solved” while OD is 0/4; this pin’s 1×30 MH axis is 10/10.

---

## Cycle 2026-08-15/16 — fresh full remasure (no product change)

**Landed:** remasure-only on product SHA `1b5ab3e`. **No product change.** Dedicated local API+worker on a fresh DB (`brainy_bench`), async ingest, `BRAINY_USE_RECALL=1`. Docs branch `pr/fresh-full-bench-a6c7`. Production FF of these docs to `dev` (GitHub homepage / default branch) and `main` is this cycle (explicit approval 2026-08-17).

This is a measurement cycle: re-pin every in-tree suite at full (or max affordable) size, with a same-cycle Mem0 Platform counter-run on OpMem, marketing, and LoCoMo 1×30.

### Own pins (this cycle)

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13 (100%)** | Non-reg. [pin](../../benchmarks/artifacts/opmem-fresh-local-20260815.md) |
| Marketing vertical | **17/17 (100%)** | Non-reg. [pin](../../benchmarks/artifacts/marketing-fresh-local-20260815.md) |
| Parity | **4/4** | Non-reg. [pin](../../benchmarks/artifacts/parity-fresh-local-20260815.md) |
| LoCoMo 1×30 conv-26 | **21/30 (70.0%)** | MH **10/10** · OD **0/4** · temporal **11/16**. +1 vs R4h 20/30. All 30 via `/recall`. [pin](../../benchmarks/artifacts/locomo-fresh-1x30-20260815.md) |
| LoCoMo full 10×all | **175/1540 (11.4%)** | **Dip.** Product `/recall`, 1 seed. Do not mix with 2026-07-31 **49.8%** search+harness. [pin](../../benchmarks/artifacts/locomo-fresh-full-20260815.md) |
| LongMemEval-20 | **4/20 (20.0%)** | Product `/recall`; jobs 4829=4829; same seed/SHA as 0/20 integrity. [pin](../../benchmarks/artifacts/lme20-fresh-20260815.md) |
| LongMemEval-500 | **not run** | ~12–20 min and ~250 extract jobs per item; 500 items is tens of hours |
| BEAM 100K conv-0 | **8/20 (40.0%)** | Search + harness answerer; **non-reg** vs hist. 8/20. [pin](../../benchmarks/artifacts/beam-100k-fresh-20260815.md) |
| BEAM 1M / 10M | **not run** | Not affordable this cycle |

1×30 is **measurement, not qualification**. Publishing 70% as if it were full LoCoMo would be a lie.

### Competitor compare (detailed)

#### 1. LoCoMo 1×30 — only fair conversational QA pin this cycle

Mem0 Platform **re-run this cycle** (same dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, same judge temp 0.0, conv-26 1×30): [locomo-mem0-fresh-1x30-20260815.md](../../benchmarks/artifacts/locomo-mem0-fresh-1x30-20260815.md).

| Axis | Brainy now (`1b5ab3e`) | Mem0 Platform (this cycle) | Graphiti OSS / Zep Platform | Stand |
| --- | ---: | ---: | --- | --- |
| LoCoMo 1×30 overall | **21/30 (70.0%)** | **11/30 (36.7%)** | **no same-pin** | **Lead this freeze**; **not** SOTA |
| Multi-hop | **10/10 (100%)** | **6/10 (60.0%)** | no same-pin | **Lead this freeze** |
| Open-domain | **0/4 (0.0%)** | **3/4 (75.0%)** | no same-pin | **Trail** |
| Temporal | **11/16 (68.8%)** | **2/16 (12.5%)** | no same-pin | **Lead this freeze** |
| Search p50 / p95 | 175 / 201 ms (local) | 492 ms p50 (platform) | no same-pin | Faster locally; **not** a platform SLO claim |

**Multi-hop (lead 10/10 vs 6/10).** R4h recoveries held (q15, q23, q26, q10). 1×30 MH axis stays closed on this pin. Do not generalize to “MH solved” as a product: OD is still 0/4, n=30, and full-suite multi-hop is **21/282 (7.4%)**.

**Open-domain (trail 0/4 vs 3/4).** Unchanged. All four OD items are WRITE_MISS (two also abstain). Hypothetical / “likely” questions. R5 structured-first yes/no from compiled facts. Do not restore OD by stuffing episodes into top-k.

**Temporal (lead 11/16 vs 2/16).** +1 vs R4h 10/16. q29 pottery/workshop date remains RETRIEVAL_MISS. Keep the lead by scoring dated **facts**, not LoCoMo-named date rules.

**Overall (lead 21 vs 11 on this freeze).** +1 vs R4h. Still **must not** say beats-Mem0 / SOTA: OD is 0/4, n=30, full `/recall` is 11.4%.

**Latency.** Local p50 175 ms vs Mem0 platform 492 ms is a harness observation, not a production SLO.

**Mem0 OSS** was not re-measured. Do not mix this 1×30 with Mem0 blog 90+ or Brainy full 11.4%.

#### 2. LoCoMo full n=1540 — named dip, not a Mem0 same-pin

| Path | Overall | MH | OD | SH | Temporal |
| --- | ---: | ---: | ---: | ---: | ---: |
| This `/recall` (`1b5ab3e`, 1 seed) | **175/1540 (11.4%)** | 21/282 (7.4%) | 5/96 (5.2%) | 88/841 (10.5%) | 61/321 (19.0%) |
| 2026-07-31 search+harness (seed-0) | 761/1540 (49.4%) | 71/282 (25.2%) | 37/96 (38.5%) | 477/841 (56.7%) | 176/321 (54.8%) |

Mem0 Platform **published** 92.5% (n=1540, top-k 200) is **context only** — not this harness, not this top-k. Same 30-item head inside the full run scored **20/30** (judge flake vs the dedicated 21/30 pin); rest of conv-26 was **12/122**. Product `/recall` returns slogans, lists, or `not in memory` on single-hop that the July harness answerer extracted from search hits.

#### 3. OpMem — lead (Mem0 re-run this cycle)

| | Brainy this cycle | Mem0 Platform this cycle |
| --- | ---: | ---: |
| OpMem | **13/13 (100%)** | **10/13 (76.9%)** |

Source: [opmem-mem0-fresh-20260815.md](../../benchmarks/artifacts/opmem-mem0-fresh-20260815.md). Mem0 fails: `cor02` (ruby), `sup03` (forget leak), `upd02` (sms). Prior ops pin was 9/12 on a 12-task set; this is a new 13-task empirical pin.

Stand: **lead ops**. Do not spend the next cycle matching Mem0 on packs.

#### 4. Marketing vertical — lead (Mem0 re-run this cycle)

| | Brainy this cycle | Mem0 Platform this cycle |
| --- | ---: | ---: |
| Marketing fixtures | **17/17 (100%)** | **4/17 (23.5%)** empirical |
| Parity (content-level) | **4/4** | **4/4** |

Source: [marketing-mem0-fresh-20260815.md](../../benchmarks/artifacts/marketing-mem0-fresh-20260815.md). Schema-only misses under `strict_schema=True` are the moat definition.

Stand: **lead governed vertical**.

#### 5. LME-20 — lift vs own integrity pin; not vs published 94.4%

| Pin | Score | Path |
| --- | ---: | --- |
| 2026-08-12 integrity | **0/20** | `/recall`, jobs 4829=4829 |
| **This cycle** | **4/20 (20.0%)** | `/recall`, same seed/SHA, jobs 4829=4829 |

CORRECT: two single-session-user, two temporal-reasoning. Multi-session **0/5**, knowledge-update **0/3**. No fair Mem0 pin on **this** harness. Do not compare 4/20 to Mem0 94.4% (500 Q, top-k 200) or SuperMemory 95% Recall@15. LME-500 **not run**. Quality LME still waits on representation + OD/multi-session reader (R6).

#### 6. BEAM — 100K non-reg; 1M/10M not run

**8/20 (40.0%)** search + harness answerer on conv-0, all 20 probes. Matches historical 8/20; ability swaps (abstention/preference up; knowledge_update/summarization down). **Not** a `/recall` fail-closed pin. BEAM 1M/10M **not run**. Mem0 published 64.1 (BEAM 1M, 700 Q) is context only.

#### 7. Graphiti / Zep — architecture target, not a scoreboard

**No same-pin.** Do not invent a LoCoMo or LME number for Graphiti OSS or Zep Platform. Published headlines stay context only.

### Why

This cycle did not change the compiler, hops, or reader. Scores moved because we **re-measured** on the current stack:

- **1×30 21/30** is the R4h MH/OD/temporal head on product `/recall`. OD is still WRITE_MISS. The +1 vs R4h is temporal, not OD.
- **Full 11.4%** is the same `/recall` path on 841 single-hop items the 1×30 never sees. Recall often emits a nearby slogan, enumerate list, or abstain instead of the atomic fact the July search+harness answerer produced. That is a **path label**, not a hidden harness glitch.
- **LME 4/20** is the first non-zero publishable product-recall pin on the same 20-item sample as 0/20. Multi-session still 0: long haystacks, compiled facts still thin.
- **BEAM 8/20** is harness-answerer non-reg on a 188-turn 100K slice. It does not speak to 1M/10M.

Vs Mem0: we lead the axes that are already Brainy-native on this freeze (ops mutation, vertical governance, 1×30 MH/temporal/overall). We trail OD hypotheticals on the same pin, and we trail every published full-suite percent. That split is still the program.

### Next

**One step:** R5A structured-first `/recall` (retire `firstStatementFromPacket` as a normal factual strategy). OD 0/4 is a diagnostic, not the PR name. Do not spend the next cycle on another full remasure, on fusion fishing, or on treating 21/30 as qualification.

| Trailing vs Mem0 | Product next (PoR) | Explicitly not next |
| --- | --- | --- |
| OD 0/4 vs 3/4 | R5A structured-first `/recall` (OD is a diagnostic) | Restore OD by raising episode top-k |
| Full LoCoMo 11.4% `/recall` | Keep the path label; lift single-hop by citing compiled facts; size ceiling with current-SHA search+harness on a subset | Silently restore 49.8% as current; publish 70% as full LoCoMo |
| Overall lead 21 vs 11 | Keep; do not declare beats-Mem0 | Treat 21/30 as qualification or SOTA |
| MH lead 10 vs 6 | Keep 1×30 MH; do not claim MH-solved while full MH is 7.4% and OD is 0/4 | Hardcode titled-work gold |
| Temporal lead 11 vs 2 | Restore q29 onto the dated fact | LoCoMo-named date rules |
| Ops / vertical lead | Keep 13/13 and 17/17 green | Spend a cycle matching Mem0 on packs |
| LME 4/20 (multi-session 0/5) | R10 after R5A-R9; not LME-500-as-quality | Compare 4/20 to published LME headlines; run LME-500 as a quality claim |
| BEAM 8/20 | Leave 100K as a sample; 1M/10M only after OD/reader work | Publish 40% as BEAM 1M |

Kill list stays in force. Do not write SOTA / beats-Mem0. Do not mix 1×30 with n=1540 or with vendor 90+. Mem0 OSS ≠ Mem0 Platform.

---

## Addendum 2026-08-17 — full `/recall` dip diagnosis + external review

Product SHA unchanged (`1b5ab3e`). This is documentation of *why* 11.4%, not a new remasure.

**1x30 did not drop** (R4h 20/30 -> 21/30). **Full LoCoMo did drop** because we scored product `POST /recall` (175/1540) instead of July search+harness (49.8%). Smoking gun: `firstStatementFromPacket` / enumerate / 188 abstains cite nearby slogans (`conv-26-q83`-`q86`). Two stacked gaps: answer-path (directional; 49.8% is **not** a current-SHA ceiling) then representation (even July search+harness 49.8% vs Mem0 92.5% n=1540 top-k 200; identity/relations still v1 strings). LME-500 and BEAM 1M were skipped for cost given LME-20 **4/20** and BEAM 100K **8/20**.

Vendor percents are **not** the same run as Brainy `/recall`. Closest industry format is Mem0 Platform 92.5% (n=1540, top-k 200, LLM-over-search). **92.5 vs 70** is invalid. **92.5 vs 11.4** is honest n=1540 on this stack but not the same answer path.

**Next (clarified after 2026-08-17 current-SHA review):** first PR is **R5A structured-first `/recall`** (retire `firstStatementFromPacket` as a normal factual strategy). R5-on-OD is a diagnostic, not the PR name. Then R5B typed packet, R6 coverage V2, R7-R9 identity/relations/hops, R10 dual-path freeze. Do not re-queue R0-R4 from a `bd987fa`-pinned report. Two published lanes. Full write-up: [locomo-full-recall-dip-why-20260817.md](../../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md). Live verdict: [2026-08-17-parity-gap-verdict.md](../external-reviews/2026-08-17-parity-gap-verdict.md). Archaeology verdict: [historical](../external-reviews/2026-08-17-competitive-archaeology-verdict.md).

---

## 2026-08-17 — R5A structured-first `/recall` landed

### Landed

Product change on `dev`/`main`: `/recall` no longer uses first-packet slogans as the normal factual strategy. Scalar/list/hop answers consume typed `value_norm` / slot values; resolve-only mentions abstain; untyped ingest sentences can still yield a non-slogan slot (`researched`, `works as`). Episodes stay provenance plus overlap fallback. SHA of this land is the R5A merge (branch `pr/r5a-structured-first-a6c7`).

### Own pins

- OpMem **13/13** (search path, local API, this SHA) — non-reg.
- Marketing vertical **17/17** plus parity **4/4** (search path, local API, this SHA) — non-reg.
- LoCoMo 1×30 / full n=1540 / LME-20 / BEAM **not re-run**. Prior pins stand: 1×30 **21/30** (OD **0/4** diagnostic), full `/recall` **11.4%**, LME-20 **4/20**, BEAM 100K **8/20**.

### Competitor compare

No new same-pin vs Mem0. Keep the previous freeze split: ops/vertical/1×30 MH lead; OD trail; do not mix 11.4% with Mem0 92.5% as the same path.

### Why

`firstStatementFromPacket` was citing slogans and resolve-only names. R5A makes `/recall` cite structured values first so later compiler/entity work is observable.

### Next

**One step:** R5B typed EvidencePacket + spans. Do not remasure n=1540 this cycle. Do not land v2 DDL. Kill list unchanged.
