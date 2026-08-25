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

---

## 2026-08-18 — R6a compiler coverage (named-subject / addressee)

### Landed

Product compiler binds clause subjects instead of always attributing to the dialogue speaker. Reports (`Casey researched …`), two-party `you …`, and `Name lives in / works as / realized that / is a` compile to that person. First-person speaker binding is unchanged. Provider extract prompt matches. Copula clip (`realized that` before adjective `is`) is included so belief tails are not clipped. Held-out audit is the merge gate. Path write-up: [locomo-full-70-80-path.md](../locomo-full-70-80-path.md).

### Own pins

- `go test ./...` green on this SHA, including `TestOpMemBenchmarkAgainstHTTPServer` and `TestMarketingMVPBenchmarkAgainstHTTPServer`.
- LoCoMo 1×30 / full n=1540 / LME-20 / BEAM **not re-run**. Prior pins stand: 1×30 **21/30** (OD **0/4**), full `/recall` **11.4%**, LME-20 **4/20**, BEAM 100K **8/20**.
- This is **not** a 70–80% full-LoCoMo claim.

### Competitor compare

No new same-pin vs Mem0. Keep the freeze split: ops/vertical/1×30 MH lead; OD trail; do not mix 11.4% with Mem0 92.5%. 70–80% on n=1540 remains R6 remainder + R7–R10, with industry search+harness labeled separately.

### Why

1×30 70% did not travel to full SH 10.5% because third-person reports were bound to the reporter or not compiled. Wrong-subject atoms make R5A structured-first cite the wrong person.

### Next

**One step:** R10 freeze remasure only when requested (product `/recall` and industry search+harness labeled separately). Stratified search+harness diagnostic before any n=1540 remasure. Kill list unchanged.

---

## 2026-08-18 — R5B–R10 representation stack

### Landed

Product change on this branch: typed EvidencePacket context items; `she`/`he` bind to the last named person; tenant/subject `ent:` IDs with aliases (two Johns coexist when labels differ); relation edges dual-write canonical IDs + evidence span; hops join on entity IDs and refuse unscoped `typed_exact`; LoCoMo harness `--eval-lane product-recall|industry-search` (industry top-k 200). Additive Postgres mig v22. Path: [locomo-full-70-80-path.md](../locomo-full-70-80-path.md) · [locomo-dual-path-freeze.md](../locomo-dual-path-freeze.md).

### Own pins

- `go test ./...` including OpMem and marketing HTTP harnesses on this SHA.
- LoCoMo 1×30 / full n=1540 / LME-20 / BEAM **not re-run**. Prior pins stand: 1×30 **21/30** (OD **0/4**), full `/recall` **11.4%**, LME-20 **4/20**, BEAM 100K **8/20**.
- This is **not** a 70–80% full-LoCoMo claim and **not** SOTA.

### Competitor compare

No new same-pin vs Mem0. Keep the freeze split: ops/vertical/1×30 MH lead; OD trail; full MH 7.4% until remasure; do not mix 11.4% with Mem0 92.5%. Industry 70–80% still needs freeze search+harness on atoms at top-k 200.

### Why

SOTA-class conversational memory needs compiled facts bound to the right person, durable identity so hops do not join the wrong John, relations with ID endpoints, and an answer path that cites those values. This pass completes that substrate. Score movement waits on a labeled dual-path freeze.

### Next

**One step:** freeze remasure when requested. Kill list unchanged.

---

## 2026-08-19/20 — fail-closed runtime integrity (prove which memory we ran)

### Landed

Product change: fail-closed extract/embed (`BRAINY_EXTRACTION_STRICT` / `BRAINY_EMBEDDING_STRICT`), ANN precondition when a 768-d embedder is configured, `GET /runtime` manifest, embedding provenance + `cmd/reembed`, oracle retrieval-before-WRITE_MISS + semantic gold. gpt-oss ingest holes that silent baseline substitution had hidden: `max_tokens=4096`, kind coerce, numeric `value`. **No new compiler rules.** PR #132. Compiler/answer/alias/KU from mixed PR #131 lives on #133 and is not this cycle.

Extractor actually used: Cloudflare gpt-oss-120b via AI Gateway (what `GET /runtime` reports). No OpenAI key; do not claim gpt-4o-mini ran.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate held on this stack. |
| Marketing vertical | **17/17** | Merge gate held. |
| LoCoMo S0 stratified 180 | product **32/180 (0.178)** · industry **62/180 (0.344)** | Invalidates Aug-19 17/180 / 52/180 (no pgvector; silent extract degrade). [pin](../../benchmarks/artifacts/locomo-integrity-s0-20260819.md) |
| LoCoMo 3×90 | product **21/90 (0.233)** · industry **33/90 (0.367)** | MH-heavy slice. Industry overall/MH matches 2026-08-11 post-cutover 33/90 / 22.2% on a **different** stack. [pin](../../benchmarks/artifacts/locomo-integrity-3x90-20260820.md) |
| Extraction ceiling (semantic gold, n=180) | det **139/180** · provider **161/180** | Same frozen questions. Provider recovers 22/41 det misses; MH 24→32/33. [pin](../../benchmarks/artifacts/extraction-ceiling-20260819.md) |
| Embedding A/B (retrieval, not QA) | BGE r@10 **0.239** · hash r@10 **0.211** · OpenAI large/small @768 r@10 **0.333** (2026-08-20) | BGE dense admission ~**4**/q vs hash cosine **152**/q. OpenAI arms: [addendum](../../benchmarks/artifacts/embedding-ab-20260820.md). [pin](../../benchmarks/artifacts/embedding-ab-20260819.md) |
| LoCoMo 1×30 / n=1540 / LME-20 / BEAM | **not re-run as freeze pins** | Prior freeze stands: 1×30 **21/30** (OD **0/4**), full `/recall` **11.4%**, LME-20 **4/20**, BEAM 100K **8/20**. LME-20 haystack for n=20 seed 1 is 9.8M chars (~11× this LoCoMo ingest). |

S0 product ledger (P3 order): PROOF_MISS **112**, RETRIEVAL **22**, READER **11**, WRITE **3** (was WRITE_MISS=120 on the invalid pin). Representation coverage 161/180 vs product QA 32/180.

This is **not** a 70–80% claim and **not** SOTA.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze ([locomo-mem0-fresh-1x30-20260815.md](../../benchmarks/artifacts/locomo-mem0-fresh-1x30-20260815.md)): Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix S0 32/180 or 62/180 or 3×90 21/90 with that 30-item freeze.

#### 1. LoCoMo conversational QA — no new same-pin; integrity remasure only

| Axis | This cycle (integrity stack) | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 n=180 industry | **62/180** | no same-n pin | Same: different n, different path |
| 3×90 industry | **33/90**, MH **8/36** | no 3×90 freeze | Same-n as our 2026-08-11 staging 33/90; **not** vs Mem0 |
| Search p50 | 168–194 ms local | 492 ms platform on the 1×30 freeze | Harness observation, **not** a SLO |

**Multi-hop.** S0 product MH is **1/33** while P4 coverage is **32/33**. The claim is written; `/recall` does not prove it (PROOF_MISS 26 of 32 MH misses). Industry MH **4/33** and 3×90 industry MH **8/36 (22.2%)** are still the largest conversational gap vs the Mem0 freeze MH **6/10** — but those denominators are not the same pin. Mechanism: packet/proof, not “nothing compiled.” Do not add a graph DB. Do not grow `providerSystemPrompt` from the leftover 19 dual-miss items.

**Open-domain.** S0 product **1/11**, industry **4/11**, ceiling **5/11**. Still thin. Same mechanism as the freeze OD **0/4** trail vs Mem0 **3/4**: durable career/works/plans as cited facts, not episode stuffing.

**Temporal.** S0 industry **21/38 (0.553)** and 3×90 industry **22/45 (0.489)** stay the stronger Brainy-native axis on search+harness. Product `/recall` temporal is weaker (S0 11/38, 3×90 11/45). Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**; do not declare it from n=180.

**Overall.** Invalidated Aug-19 S0 (17/180, 52/180) is replaced by 32/180 and 62/180 on a stack that actually ran ANN + hosted extract. That is an integrity correction, not a Mem0 overtake. Wave 1 / Gate 0 / 49.8% search+harness remain historical. Published Mem0 92.5% n=1540 stays **context only**.

#### 2. OpMem — lead (Mem0 pin stale; Brainy re-confirmed)

Brainy **13/13** this stack. Last Mem0 Platform ops pin is **10/13** (2026-08-15 freeze) / older 9/12 staging. **Lead ops.** Do not package a new “+3” sentence without re-running Mem0. Do not spend the next cycle on ops.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy re-confirmed)

Brainy **17/17**. Last Mem0 empirical **4/17**. **Lead governed vertical.** Same caveat: no refreshed Mem0 gap without a re-run.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20** product `/recall`. Not re-run. Seed-1 n=20 haystacks are 9.8M chars / 9,593 turns (~11× the LoCoMo 10-conv ingest that took ~2h50 on this extractor). No fair Mem0 pin on this harness. Do not compare 4/20 to published 94.4%.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. R5B–R10 substrate (typed packets, entity IDs, relation IDs, ID hops) is already merged; this cycle did not change it.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

The Aug-19 S0 numbers measured a **degraded runtime**, not memory quality: no pgvector so dense search saw the last 64–256 writes; extract could return regex baseline on provider error; the oracle blamed WRITE_MISS before retrieval. Fail-closed + parser fixes made the hosted extractor actually run (coverage 161/180 vs det 139/180). QA did not follow coverage: product 32/180 is still mostly PROOF_MISS. BGE ANN is live, but hybrid admits only ~4 dense neighbors per query at k=200 — hash cosine over the full store has higher r@200 because it scores more of the store, not because hash is a better embedder. OpenAI `text-embedding-3-large` / `-small` @768 (direct key, not gateway credits) lift r@10 on a rebuilt `integrity-s0-1` pin to **0.333** vs BGE **0.306** on the same stack — retrieval-only, not a QA or SOTA claim.

### Next

**One step:** packet/proof for MH (coverage 32/33 vs product QA 1/33). Do not grow compiler regex. Do not merge #133 until a labeled proof-path change is attributable. Do not burn n=1540 or Mem0 same-pin on this S0. Kill list unchanged.

---

## 2026-08-20 — MH packet/proof (hop people, not topic nouns)

### Landed

Product change merged 2026-08-21: `dev` = `main` = `f6638d4` (PR #134, parent `6b8ac5f`). `/recall` hops capitalized people on `both X and Y` joins, ignores search-fallback slot dumps, intersects typed hop values, and composes from `likes`/`loves` hop contents when slots are empty. Generic predicate hints (preference / possession / location / health). **No new compiler rules. No fusion weights.** OpenAI embedding A/B stays the 2026-08-20 pin and was not re-run. #133 / #131 stay unmerged. Incoming agent start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

Extractor actually used at ingest time: Cloudflare gpt-oss-120b via AI Gateway (unchanged store). Embedder on remasure: hosted BGE 768, ANN active, signatures match, fallbacks 0.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) | **2/33 (0.061)** | Was **1/33** on the fail-closed S0 pin. Same tenant `integrity-s0-1`, same dataset SHA, skip-ingest. [pin](../../benchmarks/artifacts/locomo-mh-packet-proof-20260820.md) |
| LoCoMo S0 n=180 / 3×90 | **not re-completed** | Full `--fail-closed` product S0 started twice and stalled on 120s embed timeouts (gateway). Do not invent a new 180-row QA pin. Prior S0 **32/180** / **62/180** and 3×90 **21/90** / **33/90** stand. |
| LoCoMo 1×30 | **not re-run** | Freeze **21/30** (OD **0/4**) stands. Do not replace it. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 rebuild pin stands (OpenAI @768 r@10 **0.333** vs this-rebuild BGE **0.306**). Do not average with the long-lived VM BGE 0.239. |

This is **not** a 70–80% claim and **not** SOTA. Name the MH dip: **2/33 is still a dip.**

Attributed CORRECT: `conv-42-q56` turtles (shared hop proof `Turtles, Dairy-free Desserts`). Second CORRECT (`conv-49-q15` soda/candy) is a crowded preference list the judge accepted — not the same mechanism.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze ([locomo-mem0-fresh-1x30-20260815.md](../../benchmarks/artifacts/locomo-mem0-fresh-1x30-20260815.md)): Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix 2/33 or 32/180 with that 30-item freeze.

#### 1. LoCoMo conversational QA — MH proof slice only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** (no new 180 pin) | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product | **2/33** (was 1/33) | no 33-item freeze | **Still the largest conversational gap**; +1 is the turtles join, not qualification |
| Search p50 | ~200–400 ms local on the 33 MH items | 492 ms platform on the 1×30 freeze | Harness observation, **not** a SLO |

**Multi-hop (still trail).** Mechanism on the recovered item: hops resolved the topic noun `animal`, search-fallback activity lists became the answer, and “Joanna likes turtles” was already in hop contents. Fix is packet/proof (person hops + typed intersect + content extract), not compiler coverage and not embedder swap. Remaining 31/33 misses are still mostly list/join/reader — gold is usually written (P4 MH coverage 32/33). Do not add a graph DB. Do not merge #133 to fish those 31.

**Open-domain.** Not re-run. Prior S0 product **1/11** / freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Same mechanism: durable career/works/plans as cited facts.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**.

**Overall.** OpenAI A/B is retrieval-only and already recorded. This cycle does not claim a Mem0 overtake.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last integrity stack **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** Same caveat.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. R5B–R10 substrate unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

Coverage 161/180 vs product QA 32/180, and MH coverage 32/33 vs QA 1/33, was not an embedder problem. On the integrity store the turtles gold was written and even present in hop contents; the planner treated `animal` as the entity and `composeFromHopValues` emitted search-fallback activity slogans (“Watching Pets Play”). Hopping Nate/Joanna, dropping search-fallback slot dumps, extracting `likes`/`loves`, and intersecting typed values made `/recall` answer `Turtles, Dairy-free Desserts` (judge CORRECT). That is one proof-path item. The other 31 MH misses are still packet/reader/list joins — not a license to grow compiler regex or merge #133.

### Next

**One step:** remaining MH list/join proof (shared facts that are not a two-name `both` cue; enumerate lists still crowd). Do not merge #133 until a remasure says compiler work is justified. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin on this slice. Kill list unchanged. Incoming agent start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-21 — MH coordinated join / list intersect

### Landed

Product change on `pr/mh-list-join-proof-1e9e` (parent `6d05e1b` / #134). `/recall` hops coordinated people (`Tim and John`, `Nate and Joanna both`, `enjoy with Casey`) without requiring a leading `both`; count questions hop the person after `does`/`has`, not the counted class; kinship `'s mother` / `her partner` chains family → slot. Join answers **intersect** typed values and hop contents and **do not fall back to the union**. Generic `owns`/`bought`/`participated in` slot extract. **No LoCoMo-named rules, no compiler regex batch, no fusion weights.** #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) | **not re-run** | Prior **2/33** stands. Integrity tenant was not on this VM. Mechanism proven with held-out fixtures (coordinated possession join; disjoint join does not dump union). |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. Name the MH dip: **2/33 is still the last measured product MH pin.**

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix an unremeasured 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product | prior **2/33** (not remasured) | no 33-item freeze | **Still the largest conversational gap**; this cycle ships the join-without-`both` proof path |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail until remasured).** #134 recovered one `both X and Y` item. Remaining misses include coordinated subjects without `both` (“Tim and John own”, “Deborah and Anna participated”), `with`-person joins, kinship (`X's mother's hobbies`, `her partner`), and crowded unions when hop contents mixed two people’s lists. Fix is still packet/proof: hop both people, intersect typed/content values, chain kinship dest → slot. Gold is usually written (P4 MH coverage 32/33). Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged: coordinated resolve + intersect, not Neo4j.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF (112) not WRITE (3). Dual-hop only on the word `both` left “Name and Name” joins on a single person, so intersection never ran and hop-content compose **unioned** preference/possession lists (turtles plus dairy-free desserts; jersey plus baseball). Coordinated/`with`/auxiliary person hops plus join-only intersect close that class generically. Kinship `'s mother` is the same proof idea (walk the relative, then the slot) without a graph DB.

### Next

**One step:** remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`) and attribute every new CORRECT. Then remaining **single-entity** list proof (pets’ names, instruments, enumerate crowding). Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until the 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-21 — MH list / count / polar proof

### Landed

Product change on `pr/mh-list-join-proof-1e9e` (HEAD `d8802ed`). `/recall` now (1) hops the person for possession/skill lists (pets’ names, instruments, tricks) and skips search-fallback dumps so occupation/hobby do not crowd the list; (2) answers **how-many** as a count of the typed set (evidence IDs for `times`, unique values otherwise) with injury counts on health not possession; (3) answers **Has/Did** polar questions **Yes** only from typed hop slots, never search-fallback; (4) extracts practice locations from `practices … at`; (5) enumerates unwind/`do to` activity; (6) hops **visit/travel** as activity and picks the typed value with most evidence for most-frequently; (7) answers **who** from other person mentions in typed hops, not the verb object; (8) drops `besides` exclusions (stemmed, so hike≈hiking); (9) treats childhood items as possession not family. **No LoCoMo-named rules, no compiler regex batch, no fusion weights.** #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned. Integrity tenant is not on this VM.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) | **not re-run** | Prior **2/33** stands. Integrity API `:18100` is not on this VM. Mechanism proven with held-out fixtures (names vs occupation; instruments/tricks vs hobby; count 2 cars not job; two ankle incidents; polar Yes from tried-activity; practice location vs occupation; unwind vs occupation; most-visited country; besides-excluded stressor; who-supports; childhood items vs family). |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. Name the MH dip: **2/33 is still the last measured product MH pin.** Unit tests are not a 33-slice replacement.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix an unremeasured 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product | prior **2/33** (not remasured) | no 33-item freeze | **Still the largest conversational gap**; this cycle ships remaining list/count/polar proof paths |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail until remasured).** Coordinated join plus list/count/polar/who/superlative/besides/unwind are shipped on this PR. Remaining live misses after remasure will likely be temporal MH dates (`when` still does not dump event hops), transfer crowding (“given to”), and identity/community lists. Gold is usually written (P4 MH coverage 32/33). Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Polar Yes from compiled facts is the R5 mechanism for that class; do not restore OD by stuffing episodes.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**. Injury **counts** are health hops, not a LoCoMo date rule.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF (112) not WRITE (3). After coordinated join, remaining MH misses were single-entity lists hopping slot nouns, how-many dumping the set, Has/Did never planning hops, unwind questions with too few bearing tokens, visit questions mapping to origin, who-answers taking the verb object, and besides-clauses leaving the excluded item as the superlative winner (`hiking` does not contain the substring `hike`). Those classes now have generic linguistic proof. Temporal `when` hops stay suppressed so dated event lists are not dumped.

### Next

**One step:** remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`) and attribute every new CORRECT. Then temporal-MH dates / transfer crowding the remasure still misses. Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until the 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-21 — MH date / transfer / after proof

### Landed

Product change on `pr/mh-list-join-proof-1e9e` (#135, HEAD `fb02403`). `/recall` now (1) **plans hops for `when` questions** but still **does not dump event/activity names as the answer** — the answer is the dated `observed_at` (year-filtered, focus-ranked) from typed hops; historical hops read the atom set, not only current-state; (2) **`given to Name` hops the giver only** (recipient is not a join) and keeps values whose evidence mentions the recipient; (3) **`after` clauses** keep matching evidence when any item hits the clause tokens; (4) `healthy` no longer maps to health — `health` is a token, meals/food/suggestions stay preference; (5) community / participating / journey / changes hop activity+identity; (6) who-injured-in-family uses kinship→health; (7) organization/beneficiary who-answers use affiliation hop values, with `value_norm` when slot extract would slogan-reject the sentence. **No LoCoMo-named rules, no compiler regex batch, no fusion weights.** #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned. Integrity tenant is not on this VM.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) | **not re-run** | Prior **2/33** stands. Integrity API `:18100` is not on this VM. Mechanism proven with held-out fixtures (dated ankle injury in 2023 vs wrist/occupation; given-to quinoa vs giver soda; after-clause meals vs candy; community garden vs nurse; family-injury who; affiliation beneficiary). |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. Name the MH dip: **2/33 is still the last measured product MH pin.** Unit tests are not a 33-slice replacement.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix an unremeasured 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product | prior **2/33** (not remasured) | no 33-item freeze | **Still the largest conversational gap**; this cycle ships dated-when / transfer / after proof paths |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail until remasured).** Date/transfer/after/community/family-who/org-beneficiary are shipped on this PR. Remaining live misses after remasure will likely be identity-surface lists (denylist blocks benchmark names) and any class the 33-slice still misses. Gold is usually written (P4 MH coverage 32/33). Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Do not restore OD by stuffing episodes.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**. MH `when` now has a date-from-hops path; that is not a 1×30 remasure.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF (112) not WRITE (3). `when` questions used to skip hops entirely so dated injuries had no proof path; enabling hops without `hopComposeAllowed` lets us read `observed_at` instead of dumping “ankle”. Transfer questions were hopping preference for the giver but dumping every like, including items never given to the recipient; `given`/`to Name` is a recipient filter, not a two-person intersect. `healthy` was substring-matching `health` and sending food lists down the injury path. After-clauses and community/journey lists were untyped dumps.

### Next

**One step:** remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`) and attribute every new CORRECT. Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until the 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-21 — MH place / group / consequence / child-count proof

### Landed

Product change on `pr/mh-list-join-proof-1e9e` (#135, HEAD `f5c2e1b`). `/recall` now (1) **answers `where` from `in`/`at`/`near` place phrases on typed hops** instead of dumping activity names, and **kinship where-questions fetch the source person as well as the unnamed partner**; (2) **`with colleagues/friends/coworkers` is a group-noun filter**, not a CapName join; (3) **`for` clauses** keep matching event/item evidence when any item hits the clause tokens; (4) **`get with having` hops health** (consequence) and does not enumerate possession; (5) **how-many children counts family members with child cues**, not partners. **No LoCoMo-named rules, no compiler regex batch, no fusion weights.** #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned. Integrity tenant is not on this VM.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) | **not re-run** | Prior **2/33** stands. Integrity API `:18100` is not on this VM. Mechanism proven with held-out fixtures (kinship diving-spot place vs activity/occupation; colleague hiking vs solo ceramics; shelter bake sale vs nurse; allergies vs dog name; child count 2 vs partner). |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. Name the MH dip: **2/33 is still the last measured product MH pin.** Unit tests are not a 33-slice replacement.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix an unremeasured 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product | prior **2/33** (not remasured) | no 33-item freeze | **Still the largest conversational gap**; this cycle ships where-place / group-with / for-clause / having-effect / child-count proof paths |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail until remasured).** Where-place, group-with, for-clause events, having-consequence, and child-count are shipped on this PR. Remaining live misses after remasure will likely be identity-surface lists (denylist blocks benchmark names) and any class the 33-slice still misses. Gold is usually written (P4 MH coverage 32/33). Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Do not restore OD by stuffing episodes.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**. MH `when` already has a date-from-hops path; this cycle does not touch it.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF (112) not WRITE (3). `where` + unnamed kin was hopping only the partner dest and composing activity values (`diving`) instead of the place in the sentence. `with colleagues` is a group noun, not a CapName, so join-entity logic never applied and solo hobbies crowded the list. `planning for X` dumped occupation because purpose clauses were not evidence filters. `get with having` matched the possession list cue (`dogs`) and enumerated names instead of the health effect. `how many children` counted every family_member, including partners. Those classes now have generic linguistic proof. Identity-surface lists remain denylist-blocked (`caroline`, `transgender woman`, `destress`).

### Next

**One step:** remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`) and attribute every new CORRECT. Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until the 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-21 — MH dual-list / kinship-dest / specific-count proof

### Landed

Product change on `pr/mh-list-join-proof-1e9e` (#135, HEAD `5171e9d`). `/recall` now (1) **intersects dual-entity list queries** instead of unioning hop values and refilling from the first person's atoms; (2) **kinship hobby lists filter atoms to the dest person** so the source's activities do not crowd; (3) **how-many Ferraris counts the head noun**, not every possession; (4) **items-for** keeps matching possessions; (5) **who-told** and **polar teach** answer from typed hops. Journey-change lists and pets' names are locked with fixtures (no denylist surface forms). **No LoCoMo-named rules, no compiler regex batch, no fusion weights.** #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned. Integrity tenant is not on this VM. GitHub CI on HEAD `3bbeff6` was green (`test` + `docker-smoke`).

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) | **not re-run** | Prior **2/33** stands. Integrity API `:18100` is not on this VM. Mechanism proven with held-out fixtures (shared community garden vs private ceramics; mother's pottery vs source hiking; Ferrari count 2 vs cottage; puzzle toy for dogs vs couch; who-told Dana vs nurse; polar teach console; journey voice changes vs hiking/nurse; pets' names vs nurse). |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. Name the MH dip: **2/33 is still the last measured product MH pin.** Unit tests are not a 33-slice replacement.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix an unremeasured 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product | prior **2/33** (not remasured) | no 33-item freeze | **Still the largest conversational gap**; this cycle ships dual-list intersect / kinship-dest / specific-count proof paths |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail until remasured).** Dual-entity list intersect, kinship-dest hobbies, specific possession counts, items-for, who-told, and polar teach are shipped on this PR. Remaining live misses after remasure will likely be identity-surface lists (denylist blocks benchmark names) and any class the 33-slice still misses. Gold is usually written (P4 MH coverage 32/33). Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Do not restore OD by stuffing episodes.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF (112) not WRITE (3). Dual-person list questions enumerated the union of hop values and then refilled from the first entity's atoms, so collectible-style intersect never applied to "activities have X and Y participated in". Kinship hobby lists scanned the source person's activities. Specific "how many Ferraris" already stemmed the head noun; the fixture locks that against counting every possession. Identity-surface lists remain denylist-blocked.

### Next

**One step:** remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`) and attribute every new CORRECT. Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until the 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-21 — MH named-community / during-clause proof

### Landed

Product change on `pr/mh-list-join-proof-1e9e` (#135, HEAD `e7708e1`). `/recall` now (1) **filters named `in the X community` lists** to evidence that mentions X, so unrelated hobbies do not crowd a named-group question; (2) **filters `during X journey` identity lists** to the named period; (3) **community/participating hops affiliation** as well as activity. Unnamed "in the community" / "during her journey" stay unfiltered. **No LoCoMo-named rules** (`civic` / `recovery` fixtures, not denylist surface forms), no compiler regex batch, no fusion weights. #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned. Integrity tenant is not on this VM.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) | **not re-run** | Prior **2/33** stands. Integrity API `:18100` is not on this VM. Mechanism proven with held-out fixtures (civic festival/coalition vs hiking/nurse; recovery voice changes vs Ohio/hiking). |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. Name the MH dip: **2/33 is still the last measured product MH pin.** Unit tests are not a 33-slice replacement.

Provider extraction ceiling covers conv-26-q39 (named-community class). conv-26-q65 (named journey) is still a provider WRITE miss — the during-clause path cannot invent missing gold. conv-26-q24 (do-to unwind) was already covered and gold-written.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix an unremeasured 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product | prior **2/33** (not remasured) | no 33-item freeze | **Still the largest conversational gap**; this cycle ships named-community / during-clause / affiliation hops |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail until remasured).** Named-community and during-clause filters are the remaining list-crowding proof class that does not require denylist terms. After remasure, leftover misses are likely destress-surface (already matched by generic `do to`) and WRITE-miss identity gold (q65). Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Do not restore OD by stuffing episodes.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF (112) not WRITE (3). Unnamed community lists already hopped activity; named "in the X community" still dumped every activity because X was not an evidence filter. Group membership is often an **affiliation** row, so community/participating now hops that predicate too. Named "during X journey" identity lists mixed in unrelated origin/identity facts. Those classes now have generic linguistic proof. Do not put denylist surface forms in product code.

### Next

**One step:** remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`) and attribute every new CORRECT. Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until the 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-21 — MH list-head / supporter-group / practice-place proof

### Landed

Product change on `pr/mh-list-join-proof-1e9e` (#135, HEAD `8676586`). `/recall` now (1) **soft-filters list-head modifiers** (`outdoor activities`, `sports collectible`) so indoor/unrelated items do not crowd when evidence mentions the modifier; (2) **who-supports keeps group nouns** from typed hop values (`friends and team`), not only CapNames, and falls back to hop slots when no person name is present; (3) **practice / location lists extract `in`/`at`/`near` places**, split comma/and lists, and skip leading `her`/`his` so an activity slot (`yoga`) does not replace the places. **No LoCoMo-named rules**, no compiler regex batch, no fusion weights. #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned. Integrity tenant `integrity-s0-1` is not on this VM — a labeled diagnostic ingest is the remasure path here, not skip-ingest attribution of #134's store.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) | **not re-run** | Prior **2/33** stands. Mechanism proven with held-out fixtures (outdoor hiking vs indoor pottery; friends/team vs nurse; mother's old home / park / beach vs yoga/nurse). |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. Name the MH dip: **2/33 is still the last measured product MH pin.** Unit tests are not a 33-slice replacement.

Provider extraction ceiling covers conv-26-q39 (named-community class). conv-26-q65 (named journey) is still a provider WRITE miss. conv-26-q24 (do-to unwind) was already covered and gold-written.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix an unremeasured 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product | prior **2/33** (not remasured) | no 33-item freeze | **Still the largest conversational gap**; this cycle ships list-head modifiers, supporter-group who-answers, and practice-place lists |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail until remasured).** Remaining likely misses after remasure: WRITE-miss identity gold (q65) and identity-surface lists that still need denylist terms. Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Do not restore OD by stuffing episodes.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF (112) not WRITE (3). Who-answers scanned hop contents for CapNames and dropped group-noun supporters. Practice/location lists enumerated the activity slot (`yoga`) and aborted place extract on leading `her`. List-head modifiers (`outdoor`, `sports`) were the remaining adjective filter on enumerate heads. Those classes now have generic linguistic proof.

### Next

**One step:** remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`) if that tenant exists; otherwise a **labeled diagnostic** async ingest on a clean fail-closed DB (WRITE+PROOF mixed — not skip-ingest attribution). Attribute every new CORRECT. Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until the 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-21 — MH diagnostic remasure + worker lease/drain

### Landed

Product + worker on `pr/mh-list-join-proof-1e9e` (#135, HEAD `a7cf465`). This increment (1) **heartbeats extraction-job leases every 10s** for the whole `ProcessNext` so provider extract longer than the 30s lease is not reclaimed; (2) **keeps each worker slot claiming until the queue is idle** so one slow extract cannot park the other slots; (3) embedded-postgres tests bind an ephemeral port. Same-subject jobs still serialize. **No LoCoMo-named rules**, no compiler regex batch, no fusion weights. #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned.

Fail-closed diagnostic ingest on `brainy_mh` (tenant prefix `diag-mh-135`): **1472 jobs completed / 0 failed**, 21181 memories, ANN active, mixed_dimensions=false, signatures.match, fallbacks 0. Then product `/recall` scored the MH-33 slice with skip-ingest on that fresh store.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) integrity skip-ingest | **2/33** | Last attributed pin. **Unchanged.** |
| LoCoMo S0 MH slice (product `/recall`) diagnostic fresh ingest | **7/33 (0.212)** | [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-20260821.md). WRITE+PROOF mixed. Do **not** overwrite 2/33. |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. Name the MH dip: **integrity 2/33 is still the last attributed product MH pin.** The diagnostic 7/33 is a different store and cannot be used as skip-ingest proof-path attribution.

Provider extraction ceiling still covers conv-26-q39. conv-26-q65 remains a WRITE miss on this diagnostic (identity gold not invented). conv-26-q24 (do-to unwind) was coverage-true and still WRONG on product crowding.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix diagnostic 7/33 or integrity 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product (integrity) | **2/33** | no 33-item freeze | **Still the attributed conversational gap** |
| S0 MH product (diagnostic) | **7/33** WRITE+PROOF mixed | no 33-item freeze | Context only; not a same-pin row |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail until an integrity remasure).** Diagnostic CORRECTs were mostly crowded lists the judge accepted, plus the Nate+Joanna turtles join surviving a fresh WRITE. Remaining WRONG is list crowding / wrong slot, plus WRITE-miss identity gold (q65). Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Do not restore OD by stuffing episodes.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF, not WRITE. The 30s job lease vs 120s provider timeout was reclaiming live extracts (`ErrLeaseLost`), and `ProcessAvailable` parked idle slots until the slowest of a batch of N finished. Heartbeat + looping drain let the diagnostic ingest complete (1472/1472, 0 failed). Product `/recall` on that store is **7/33**, but WRITE is mixed with proof, so the extra CORRECTs vs 2/33 cannot be attributed to #135's list/join mechanisms. Crowding still dominates WRONG.

### Next

**One step:** remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`) for attributed proof-path lift. Until that tenant exists, do not treat diagnostic 7/33 as the pin. Then attack **list crowding** (enumerate dumps) without fusion fishing or LoCoMo-named rules. Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until an attributed 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-21 — MH list-crowding skip-ingest (PROOF-only)

### Landed

Product on `pr/mh-list-join-proof-1e9e` (#135, HEAD `33bdea4`). List-crowding proof: skip atom-index refill once hops already listed a slot; location lists extract `in`/`at`/`near` places (practice-object soft filter, relative-clause / gerund stop) and **never fall back to the activity dump**; rank enumerate items by query / named / childhood evidence, drop zero-score crowding, then cap at 8. Hybrid hop-compose does not overwrite a locked list answer. **No LoCoMo-named rules**, no category dictionaries, no fusion weights. #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned.

Skip-ingest re-score of frozen tenant `diag-mh-135` (same WRITE as the 7/33 diagnostic). Fail-closed runtime: ANN active, mixed_dimensions=false, signatures.match, fallbacks 0.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) integrity skip-ingest | **2/33** | Last integrity pin. **Unchanged.** |
| LoCoMo S0 MH slice diagnostic fresh ingest | **7/33 (0.212)** | WRITE+PROOF mixed. [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-20260821.md) |
| LoCoMo S0 MH slice diagnostic skip-ingest (this cycle) | **8/33 (0.242)** | PROOF-only vs 7/33. [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-skip-ingest-20260821.md) |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. **Integrity 2/33 is still the last attributed integrity-tenant pin.** 8/33 is a labeled diagnostic PROOF-only delta on `diag-mh-135`, not a replacement for 2/33.

A cap-only intermediate on the same tenant scored **5/33** (truncate cut gold in the tail). Ranking-then-cap recovered the lists. Do not cite 5/33 as a pin.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix diagnostic 8/33 or integrity 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product (integrity) | **2/33** | no 33-item freeze | **Still the attributed integrity-tenant gap** |
| S0 MH product (diagnostic skip-ingest) | **8/33** PROOF-only vs 7/33 | no 33-item freeze | Context / mechanism pin only; not a same-pin row |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail on the integrity tenant).** Diagnostic skip-ingest moved **7→8/33** with cleaner lists (pets' names, childhood items, a typed count) and preserved the Nate+Joanna turtles join. Remaining WRONG is still mostly wrong slot / incomplete place extract, plus WRITE-miss identity gold (q65). Unhealthy-snack gold was a named dip (ranking dropped a dump tail the judge had accepted). Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Do not restore OD by stuffing episodes.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF, not WRITE. Atom-index refill was appending up to 40 extra values onto hops that had already listed a slot (counts became 3 instead of 2; location lists dumped every activity). Location lists now extract places and do not fall back to activity slogans. Enumerate answers rank by query evidence (named-assignment, childhood clause, list-head modifiers) so a bound of 8 does not truncate gold in the tail. Residual misses are incomplete place extract, healthy/unhealthy slot confusion, and the WRITE-miss identity gold.

### Next

**One step:** remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`) so the 8/33 diagnostic delta is either confirmed or rejected on the integrity store. Then keep attacking remaining list/place crowding without fusion fishing or LoCoMo-named rules. Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until an **integrity-tenant** 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-21 — MH on-prep / polar-lock / un- skip-ingest (PROOF-only)

### Landed

Product on `pr/mh-list-join-proof-1e9e` (#135, HEAD `a521c47`). Place lists take locative `on` (yoga **on** the beach), cut nested preps and person+verb clauses even when the evidence blob is lowercased, skip date tails after `on`, and reject a place that is exactly the practice object. Polar queries no longer hop-compose or fall through to structured/episode dumps. Un- list heads drop evidence that has the positive/comparative form (`healthy` / `healthier`) but not the `un-` form, and do not require the adjective on gold items that never repeat it. Snacks are an enumerate/preference cue. **No LoCoMo-named rules**, no place lexicon, no fusion weights. #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned.

Skip-ingest re-score of frozen tenant `diag-mh-135` (same WRITE as the 7/33 diagnostic). Fail-closed runtime: ANN active, mixed_dimensions=false, signatures.match, fallbacks 0.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) integrity skip-ingest | **2/33** | Last integrity pin. **Unchanged.** |
| LoCoMo S0 MH slice diagnostic fresh ingest | **7/33 (0.212)** | WRITE+PROOF mixed. [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-20260821.md) |
| LoCoMo S0 MH slice diagnostic skip-ingest (crowding) | **8/33 (0.242)** | Prior PROOF-only. [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-skip-ingest-20260821.md) |
| LoCoMo S0 MH slice diagnostic skip-ingest (this cycle) | **8/33 (0.242)** | Composition change, not a lift. [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-skip-ingest-onprep-20260821.md) |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. **Integrity 2/33 is still the last attributed integrity-tenant pin.** 8/33 remains a labeled diagnostic on `diag-mh-135`, not a replacement for 2/33.

An intermediate on SHA `d7ee0ce` scored **7/33** (polar lock dropped a fragile crowded-accept before the un- gold fix landed). Do not cite that 7/33 as a pin.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix diagnostic 8/33 or integrity 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product (integrity) | **2/33** | no 33-item freeze | **Still the attributed integrity-tenant gap** |
| S0 MH product (diagnostic skip-ingest) | **8/33** composition vs crowding 8/33 | no 33-item freeze | Context / mechanism pin only; not a same-pin row |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail on the integrity tenant).** Diagnostic skip-ingest stayed **8/33**. Attributed swap: `conv-49-q15` unhealthy snacks is now CORRECT (`soda and candy` present) via the un- filter; `conv-48-q73` polar teach-console is a named dip (`not in memory` after polar hop-compose lock; prior CORRECT was a crowded dump). Practice-place is cleaner (`Park` only) but still incomplete — hops on this tenant never listed `yoga on the beach` / studio locatives. Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Do not restore OD by stuffing episodes.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF, not WRITE. Locative `on` plus nested-prep / person-verb cuts stop `yoga in the park` and `park Name met Name` from becoming places; evidence blobs are lowercased so the person cut must title-case first. Polar misses were hop-composing activity lists; locking compose is honest abstention, not a Yes. Un- list heads were ranking `Healthy Snacks` / `Healthier Snack Ideas` over soda/candy because the gold never repeats `unhealthy`; dropping positive/comparative evidence (and not requiring the `un-` token on gold) recovers that slot. Residual misses are still incomplete place extract (WRITE-thin locatives on this tenant), outdoor-with-colleagues dumps, and the WRITE-miss identity gold.

### Next

**One step:** remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`) so the diagnostic 8/33 composition is either confirmed or rejected on the integrity store. Remaining high-value PROOF classes on this tenant: outdoor-with-colleagues, dual-community intersect, polar Yes when typed hops actually carry the claim, practice places when locatives exist in hops. Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until an **integrity-tenant** 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-21 — MH enumerate refine / outdoor∩group skip-ingest (PROOF-only)

### Landed

Product on `pr/mh-list-join-proof-1e9e` (#135, HEAD `5804072`). Enumerate mode now shares `refineEnumeratedItems` with answer-mode lists (besides / hop-evidence, query rank, cap 8). When a query has **both** a list-head modifier and a group companion, the reader prefers items that hit both, else list-head, else companion — so outdoor family hiking is not replaced by a colleague indoor singleton. **No LoCoMo-named rules**, no outdoor-sport lexicon, no fusion weights. #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned.

Skip-ingest re-score of frozen tenant `diag-mh-135` (same WRITE as the 7/33 diagnostic). Fail-closed runtime: ANN active, mixed_dimensions=false, signatures.match, fallbacks 0.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) integrity skip-ingest | **2/33** | Last integrity pin. **Unchanged.** |
| LoCoMo S0 MH slice diagnostic fresh ingest | **7/33 (0.212)** | WRITE+PROOF mixed. [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-20260821.md) |
| LoCoMo S0 MH slice diagnostic skip-ingest (crowding) | **8/33 (0.242)** | [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-skip-ingest-20260821.md) |
| LoCoMo S0 MH slice diagnostic skip-ingest (on-prep) | **8/33 (0.242)** | [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-skip-ingest-onprep-20260821.md) |
| LoCoMo S0 MH slice diagnostic skip-ingest (this cycle) | **8/33 (0.242)** | Same CORRECT set as on-prep. [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-skip-ingest-enumerate-20260821.md) |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. **Integrity 2/33 is still the last attributed integrity-tenant pin.** 8/33 remains a labeled diagnostic on `diag-mh-135`, not a replacement for 2/33.

`go test ./...` was green on `5804072` before this remasure.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix diagnostic 8/33 or integrity 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product (integrity) | **2/33** | no 33-item freeze | **Still the attributed integrity-tenant gap** |
| S0 MH product (diagnostic skip-ingest) | **8/33** same CORRECT set as on-prep | no 33-item freeze | Context / mechanism pin only; not a same-pin row |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail on the integrity tenant).** Diagnostic skip-ingest stayed **8/33**. Outdoor-with-colleagues is now a single outdoor hike fact instead of a colleague indoor singleton or an unbounded dump; still WRONG because mountaineering is not in the hop slots. Dual-community still content-falls-back to unwind gerunds (`yoga` vs `organized yoga` is not an exact intersect). Polar teach-console remains the named dip. Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Do not restore OD by stuffing episodes.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF, not WRITE. Eval list questions hit `mode=enumerate`, which used to skip the answer-mode refine path. Sharing that path caps dumps. Sequential colleague-then-outdoor filtering then kept a single indoor colleague hit; preferring list-head when no item hits both restores outdoor hiking without a sport dictionary. Residual misses are dual-community exact-slot joins, kinship dest hobbies, incomplete place extract, and the WRITE-miss identity gold.

### Next

**One step:** keep attacking remaining list/join PROOF on this tenant — dual-community slot intersect must match `yoga`/`organized yoga` and partner-mentioned running groups, not fall back to unwind slogans. Then remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`). Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until an **integrity-tenant** 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-21 — MH community-join skip-ingest (PROOF-only)

### Landed

Product on `pr/mh-list-join-proof-1e9e` (#135, HEAD `c97cc0a`). Community dual-entity lists join by token-subset when the longer value has `organized` / `started` / `group`, and keep values that name the other hop entity. Containment may read `search_fallback` slot lists. Generic hops stay exact-intersect so sports-collectible / preference joins do not collapse. Enumerate refine + outdoor∩group from `5804072` remains. **No LoCoMo-named rules**, no activity lexicon, no fusion weights. #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned.

Skip-ingest re-score of frozen tenant `diag-mh-135` (same WRITE as the 7/33 diagnostic). Fail-closed runtime: ANN active, mixed_dimensions=false, signatures.match, fallbacks 0.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) integrity skip-ingest | **2/33** | Last integrity pin. **Unchanged.** |
| LoCoMo S0 MH slice diagnostic fresh ingest | **7/33 (0.212)** | WRITE+PROOF mixed. [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-20260821.md) |
| LoCoMo S0 MH slice diagnostic skip-ingest (enumerate / outdoor) | **8/33 (0.242)** | [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-skip-ingest-enumerate-20260821.md) |
| LoCoMo S0 MH slice diagnostic skip-ingest (this cycle) | **9/33 (0.273)** | +1 vs 8/33. [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-skip-ingest-community-join-20260821.md) |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. **Integrity 2/33 is still the last attributed integrity-tenant pin.** 9/33 is a labeled diagnostic on `diag-mh-135`, not a replacement for 2/33.

`go test ./...` was green on `c97cc0a` before this remasure. An unscopeed-containment intermediate scored **7/33**; do not cite it.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix diagnostic 9/33 or integrity 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product (integrity) | **2/33** | no 33-item freeze | **Still the attributed integrity-tenant gap** |
| S0 MH product (diagnostic skip-ingest) | **9/33** PROOF-only vs 8/33 | no 33-item freeze | Context / mechanism pin only; not a same-pin row |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail on the integrity tenant).** Diagnostic skip-ingest moved **8→9/33**. Attributed win: dual community `yoga` + `deborah's running group` via containment against a fallback hop plus partner mention. Sports collectible and Nate+Joanna turtles held because generic joins stayed exact. Outdoor-with-colleagues is still a hike-only WRONG (mountaineering WRITE-thin). Polar teach-console remains the named dip. Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Do not restore OD by stuffing episodes.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF, not WRITE. Dual community exact-intersect was empty (`yoga` vs `organized yoga`) and one entity's activity hop is search_fallback, so the reader fell back to shared unwind gerunds. Community lists now prove the shorter token-subset when the longer value carries `organized`/`started`/`group`, and keep a slot that names the other person. Scoping that to community lists avoided collapsing sports collectibles to `Book`. Residual misses are kinship dest hobbies, incomplete place extract, outdoor mountaineering WRITE, and identity gold.

### Next

**One step:** remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`) so the 9/33 diagnostic delta is either confirmed or rejected on the integrity store. Remaining high-value PROOF on this tenant: kinship dest hobbies, practice places when locatives exist, polar Yes when typed hops carry the claim. Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until an **integrity-tenant** 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-21 — MH kinship-dest skip-ingest (PROOF-only)

### Landed

Product on `pr/mh-list-join-proof-1e9e` (#135, HEAD `9d8dbeb`). Unnamed kinship dest hops rewrite a role word (`mother`) or copula gloss (`her mom`) to `{Name}'s {role}` from family-hop contents, drop the source entity id, and merge dest-subject attitude slots (`had X as a hobby`, `passionate about`, `interested in`) ahead of the typed activity list. List-class nouns (`hobbies` / `activities`) are not used as item-rank evidence. Community-join + enumerate refine + outdoor∩group remain. **No LoCoMo-named rules**, no hobby lexicon, no fusion weights. #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned.

Skip-ingest re-score of frozen tenant `diag-mh-135` (same WRITE as the 7/33 diagnostic). Fail-closed runtime: ANN active, mixed_dimensions=false, signatures.match, fallbacks 0.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) integrity skip-ingest | **2/33** | Last integrity pin. **Unchanged.** |
| LoCoMo S0 MH slice diagnostic fresh ingest | **7/33 (0.212)** | WRITE+PROOF mixed. [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-20260821.md) |
| LoCoMo S0 MH slice diagnostic skip-ingest (community join) | **9/33 (0.273)** | [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-skip-ingest-community-join-20260821.md) |
| LoCoMo S0 MH slice diagnostic skip-ingest (this cycle) | **10/33 (0.303)** | +1 vs 9/33. [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-skip-ingest-kinship-dest-20260821.md) |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. **Integrity 2/33 is still the last attributed integrity-tenant pin.** 10/33 is a labeled diagnostic on `diag-mh-135`, not a replacement for 2/33.

`go test ./...` was green on `9d8dbeb` before this remasure. A rewrite-only intermediate on `f5b9ea8` stayed **9/33**; do not cite it as a dip.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix diagnostic 10/33 or integrity 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product (integrity) | **2/33** | no 33-item freeze | **Still the attributed integrity-tenant gap** |
| S0 MH product (diagnostic skip-ingest) | **10/33** PROOF-only vs 9/33 | no 33-item freeze | Context / mechanism pin only; not a same-pin row |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail on the integrity tenant).** Diagnostic skip-ingest moved **9→10/33**. Attributed win: unnamed kin dest hobbies (`reading, travel, art, cooking`) via dest-mention rewrite plus dest-subject attitude slots. Dual community and turtles held. Outdoor-with-colleagues is still a hike-only WRONG (mountaineering WRITE-thin). Polar teach-console remains the named dip. Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Do not restore OD by stuffing episodes.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF, not WRITE. Family hops often store a role word, not a person name, so the next hop matched every source memory that mentioned `mother`. Rewriting dest to `{Name}'s mother` scoped the activity hop. Reading / travel / art on this tenant are dest-subject facts **without** an activity atom; merging attitude copulas (`had X as a hobby`, `passionate about`, `interested in`) recovered them without a hobby list. Residual misses are incomplete place extract, outdoor mountaineering WRITE, polar Yes, unwind rank, and identity gold.

### Next

**One step:** remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`) so the 10/33 diagnostic delta is either confirmed or rejected on the integrity store. Remaining high-value PROOF on this tenant: practice places when locatives exist, polar Yes when typed hops carry the claim, unwind rank for `do to`. Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until an **integrity-tenant** 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-21 — MH slot-aligned dest-subject skip-ingest (PROOF-only)

### Landed

Product on `pr/mh-list-join-proof-1e9e` (#135, HEAD `2e84435`). Dest-subject facts outside the typed atom window are recovered into hop slots: practice locatives (plus compositional `the {practice} {noun}`), unwind/calm/`to *stress` activities, `plays`/`{noun} practice` objects, trick-mentioned skills, and besides+stressor work facts. `unwind` is not treated as an `un-` negation. Kinship dest + community-join + enumerate refine remain. **No LoCoMo-named rules**, no park/beach/studio/instrument lexicon, no fusion weights. #133 / #131 stay unmerged. OpenAI A/B not re-run. n=1540 / Mem0 same-pin not burned.

Skip-ingest re-score of frozen tenant `diag-mh-135` (same WRITE as the 7/33 diagnostic). Fail-closed runtime: ANN active, mixed_dimensions=false, signatures.match, fallbacks 0.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; prior integrity gate stands. |
| Marketing vertical | **17/17** | Not re-run; prior integrity gate stands. |
| LoCoMo S0 MH slice (product `/recall`) integrity skip-ingest | **2/33** | Last integrity pin. **Unchanged.** |
| LoCoMo S0 MH slice diagnostic fresh ingest | **7/33 (0.212)** | WRITE+PROOF mixed. [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-20260821.md) |
| LoCoMo S0 MH slice diagnostic skip-ingest (kinship dest) | **10/33 (0.303)** | [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-skip-ingest-kinship-dest-20260821.md) |
| LoCoMo S0 MH slice diagnostic skip-ingest (this cycle) | **12/33 (0.364)** | +2 vs 10/33. [artifact](../../benchmarks/artifacts/locomo-mh-diag-135-skip-ingest-slot-recover-20260821.md) |
| LoCoMo S0 n=180 / 3×90 / 1×30 | **not re-run** | Prior pins stand. 1×30 **21/30** is still diagnostic. |
| LME-20 / n=1540 / BEAM | **not re-run** | Prior pins stand (LME-20 **4/20**, full `/recall` **11.4%**). |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** a 70–80% claim and **not** SOTA. **Integrity 2/33 is still the last attributed integrity-tenant pin.** 12/33 is a labeled diagnostic on `diag-mh-135`, not a replacement for 2/33.

`go test ./internal/... ./cmd/...` was green on this SHA before the remasure. An intermediate on `94f119b` was also **12/33** with noisier place/unwind lists; do not cite it as a separate pin.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep same-pin this cycle. Reuse the 2026-08-15 freeze: Mem0 Platform 1×30 **11/30** (MH 6/10, OD 3/4, temporal 2/16). Do **not** mix diagnostic 12/33 or integrity 2/33 with that 30-item freeze.

#### 1. LoCoMo conversational QA — proof mechanism only

| Axis | This cycle | Mem0 Platform freeze | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | **11/30** | Prior freeze **lead**; this cycle does not refresh it |
| S0 n=180 product | prior **32/180** | no same-n pin | **Do not trail/lead vs 11/30** |
| S0 MH product (integrity) | **2/33** | no 33-item freeze | **Still the attributed integrity-tenant gap** |
| S0 MH product (diagnostic skip-ingest) | **12/33** PROOF-only vs 10/33 | no 33-item freeze | Context / mechanism pin only; not a same-pin row |
| Search p50 | n/a | 492 ms platform on the 1×30 freeze | No new latency pin |

**Multi-hop (still trail on the integrity tenant).** Diagnostic skip-ingest moved **10→12/33**. Attributed wins: besides+stressor `work`; instrument list `clarinet, violin`. Place list now has beach/studio/park but still misses mother's old home (WRITE-thin). Unwind recovers running and still misses pottery. Pet tricks recover sit/stay/paw/rollover and miss swimming/frisbee/skateboard. Polar teach-console remains the named dip. Do not add a graph DB. Do not merge #133.

**Open-domain.** Not re-run. Prior freeze OD **0/4** vs Mem0 **3/4** still stands as a trail axis. Do not restore OD by stuffing episodes.

**Temporal.** Not re-run. Keep the freeze temporal lead (11/16 vs Mem0 2/16) as **stale until 1×30 is re-run**.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.** YAML packs remain the verticalisation layer; this cycle does not touch packs.

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context. Postgres graph-shaped hops (ADR-004) unchanged.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

S0 ledger is still PROOF, not WRITE. Typed hops miss dest-subject locatives, unwind/`to *stress` facts, play/practice objects, and trick-mentioned skills that sit outside the atom top-k window or carry no matching predicate. Recovering those slots — and ranking them by unwind/play/trick evidence — moved two MH items without category dictionaries. `unwind` was incorrectly treated as `un-`+`wind`, which disabled enumerate drop-zero and kept camping dumps. Residual misses are mother's-home WRITE, pottery without unwind cues, incomplete trick coverage, polar Yes, mountaineering WRITE, and identity gold.

### Next

**One step:** remasure MH-only 33 on `integrity-s0-1` (`--fail-closed --skip-ingest`) so the 12/33 diagnostic delta is either confirmed or rejected on the integrity store. Remaining high-value PROOF on this tenant: practice-home locatives when WRITE exists, unwind objects that lack unwind cues, polar Yes when typed hops carry the claim. Do not merge #133. Do not re-run OpenAI A/B. Do not burn n=1540 or Mem0 same-pin until an **integrity-tenant** 33-slice moves. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-22 — current-SHA S0 this-VM (`diag-mh-135` + conv-30)

### Landed

Product `/recall` on `origin/dev` **`453a929`** (merge of #135). This VM's harness SHA is **`98d5db8`** (PR #136): Mem0 Platform adapter aligned to v3 search/add + event wait, chunk 1, session timestamps, default top_k 200; product-recall S0 records a hung `/recall` as an item miss instead of aborting the 180. **No product `/recall` behavior change in those harness commits.** Hybrid reader **off** (`BRAINY_RECALL_LLM` unset). `#133` / `#131` stay unmerged. `main` stays at the packet/proof SHA; this cycle does **not** fast-forward production.

Store: local `:18100` / DB `brainy_mh`, tenant `diag-mh-135`, 10/10 LoCoMo conversations after ingesting `conv-30` (369 turns; 98/99 extract jobs completed; one `session_17` batch of four turns failed on a non-JSON provider body). ANN active, signatures match, embedder/extractor fallbacks 0, 22475 memories at 768-d. Dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`. Sample: stratified 180 seed 1.

Pin: [locomo-s0-diag-mh-135-20260822.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-20260822.md). Mem0 faithfulness audit (not a score): [mem0-harness-audit-2026-08-22.md](./mem0-harness-audit-2026-08-22.md).

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; merge gate stands. |
| Marketing vertical | **17/17** | Not re-run; merge gate stands. |
| LoCoMo S0 product `/recall` **this VM** (`diag-mh-135`, reader off) | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. Search p50 168 ms. Ledger: **PROOF 59 / READER 52 / RETRIEVAL 39 / WRITE 10 / HARNESS 1**. |
| LoCoMo S0 industry search+harness **this VM** | **62/180 (0.344)** | MH 10/33 · OD 3/11 · SH 27/98 · temporal 22/38. Search p50 142 ms. Same overall as integrity-VM industry. |
| LoCoMo S0 product **integrity VM** (`integrity-s0-1`) | **32/180** | Different tenant. Ledger PROOF 112 / RETRIEVAL 22 / READER 11 / WRITE 3. **Do not mix with 19/180.** |
| LoCoMo S0 MH slice integrity skip-ingest | **2/33** | Last attributed integrity-tenant pin. **Unchanged.** |
| LoCoMo S0 MH slice diagnostic skip-ingest | **12/33** | Matches this-VM S0 MH 12/33. Does not replace 2/33. |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. OD **0/4**. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old product SHA `1b5ab3e`. Not re-run. |
| Industry historical | **49.8%** | July, old stack. **Not** this SHA. This-VM industry is **34.4%**. |
| LME-20 / BEAM | **not re-run** | 4/20 and 8/20 stand. |
| Embedding A/B | **not re-run** | 2026-08-20 pin stands. |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 19/180 does not replace integrity 32/180. 12/33 MH on this tenant is the #135 slot-recovery pin at S0 n, not a path by itself to 80%.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. The 2026-08-15 Mem0 Platform 1×30 freeze remains **11/30** (MH 6/10, OD 3/4, temporal 2/16) — and the audit now shows that freeze is **not** their published protocol. A Brainy lead over that freeze is not a fair win.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | ---: | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol **handicapped** (v2, top_k 30, chunk 8, no timestamps) | Do **not** refresh lead/trail from 21 vs 11. Fair row is the in-flight stratified 180. |
| S0 n=180 product this VM | **19/180** reader off | **no same-n pin yet** | **Do not trail/lead vs 11/30 or vs published 92.5%** |
| S0 n=180 industry this VM | **62/180** | **no same-n pin yet** | Same-pin lane vs Mem0 once the fair 180 lands (top_k 200, shared judge) |
| S0 MH product (integrity) | **2/33** | no 33-item freeze | Still the attributed integrity-tenant gap |
| S0 MH product (this tenant) | **12/33** | no 33-item freeze | Product MH **leads** this-VM industry MH (10/33). Do not regress it. |
| Search p50 | product 168 ms / industry 142 ms local | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop.** This-VM product MH **12/33** vs industry **10/33**. Integrity **2/33** is still the other store. Do not add a graph DB. Do not merge #133. Closing 80% is not an MH-only program: even perfect MH+OD on this sample still needs most of SH 98 + temporal 38.

**Open-domain.** This-VM product **0/11**, industry **3/11**. 1×30 freeze OD **0/4** vs Mem0 **3/4** remains a trail axis on that freeze. Do not restore OD by stuffing episodes.

**Single-hop / temporal (the mass).** Product SH **5/98** and temporal **2/38** vs industry **27/98** and **22/38**. Largest product cells: `single-hop:PROOF_MISS` 48, `temporal:READER_MISS` 24, `single-hop:READER_MISS` 22, `single-hop:RETRIEVAL_MISS` 19. Reader-off `/recall` is the labeled gap vs the industry answerer; PROOF is still the largest bucket. P1 (hybrid reader, including enumerate mode) is the next product measurement. P2 (unlock date/where/polar / rebuild answer path) waits on whether P1 moves SH/temporal.

**Published Mem0 92.5% (n=1540, top_k 200, their harness)** stays a context row, never a scoreboard row.

#### 2. OpMem — lead (Mem0 pin stale; Brainy not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Do not package a new “+3” sentence.

#### 3. Marketing vertical — lead (Mem0 pin stale; Brainy not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.** Published headlines stay context.

**Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

### Why

Two stores, two product numbers: integrity 32/180 (PROOF-heavy, reader mostly unused) vs this-VM 19/180 (reader off, SH/temporal collapse). Industry matched at 62/180, so search+harness on current SHA is not the 49.8% historical stack. Product MH held at 12/33 after #135; the 80% hole is SH/temporal proof and reading, not another MH regex. WRITE 10/180 is still not the mass — do not merge compiler-fishing #133. The Mem0 11/30 freeze cannot decide lead/trail because the harness was not their protocol.

### Next

**One step:** P1 reader A/B on this frozen store (`BRAINY_RECALL_LLM=1`, skip-ingest product 180, hybrid reachable in **enumerate** as well as `answer`; keep date/where/polar locks). In parallel: fair Mem0 Platform 180 (`--system mem0 --top-k 200`, chunk 1, v3). P2 only if P1 moves SH/temporal. P3 follows the ledger (PROOF 48 on SH). n=1540 only after a stratified delta. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-22 — P1 hybrid reader A/B (this-VM `diag-mh-135`)

### Landed

Product `/recall` hybrid reader reachable in **enumerate** as well as `answer` (`d86dcf1`), plus `max_tokens: 2048` so hosted gpt-oss can emit JSON (`3d42b17`). Staging may set `BRAINY_RECALL_LLM=1`; the Go **default stays off** until an owner yes. Same frozen store as the reader-off pin: tenant `diag-mh-135` + conv-30, dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`. `#133` / `#131` stay unmerged. `main` is not fast-forwarded.

Pin: [locomo-s0-diag-mh-135-llm-20260822.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-llm-20260822.md). Reader-off pair: [locomo-s0-diag-mh-135-20260822.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-20260822.md).

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; merge gate stands. |
| Marketing vertical | **17/17** | Not re-run; merge gate stands. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product `/recall` this VM **hybrid on** | **37/180 (0.206)** | MH **10/33 (dip)** · OD **1/11** · SH **19/98** · temporal **7/38**. SHA `3d42b17`. Ledger: **PROOF 44 / READER 49 / RETRIEVAL 39 / WRITE 10 / HARNESS 1**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 37/180 does not replace integrity 32/180. MH 12→10 is a **dip** (lost clarinet/violin dump, dual collectible, community join, snack filter, Ferrari count; gained polar-teach, family-injury who, who-supports).

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`--system mem0 --top-k 200`, chunk 1, v3 add `/v3/memories/add/` + event wait) is still ingesting. The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19/180** off → **37/180** on | **no same-n pin yet** | Product doubled vs itself. Still trails this-VM industry **62/180**. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin yet** | Same-pin lane vs Mem0 once the fair 180 lands. |
| S0 MH product (this tenant) | reader-off **12/33** → hybrid **10/33** | no 33-item freeze | **Dip.** Do not treat hybrid-on MH as the #135 recovery pin. |
| Search p50 | hybrid-on product 130 ms local | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (dip 12→10).** Hybrid overwrote typed lists/counts (Melanie instruments + pottery/beach; Ferrari 2 → hop dump). Dual-entity joins that were nonempty also lost. Reader-off MH **12/33** still leads this-VM industry MH **10/33**; hybrid-on **ties** industry. Closing the dip is a lock, not more LLM.

**Open-domain.** 0→1/11. Industry **3/11**. Still a trail axis vs published Mem0 OD on the 1×30 freeze. Do not restore OD by stuffing episodes.

**Single-hop / temporal (the P1 move).** SH **5→19/98**, temporal **2→7/38**. Industry remains **27/98** and **22/38**. Largest hybrid-on cells: `single-hop:PROOF_MISS` 37, `temporal:READER_MISS` 22 (locked calendar dates vs gold “Sunday before …”), `single-hop:READER_MISS` 19, `single-hop:RETRIEVAL_MISS` 19. P1 **justifies P2** for SH/temporal. Where/polar stay locked.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.**

#### 3. Marketing vertical — lead (not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured.

### Why

Enumerate-mode hybrid is the mechanism for the SH/temporal lift: list questions now reach the packet reader instead of a locked dump. The MH dip is the same mechanism overwriting typed joins/counts that were already correct. PROOF 59→44 is SH dump cleanup, not new compiler coverage. WRITE stays 10 — do not merge #133. Reader-off 19/180 remains the labeled no-LLM product pin; 37/180 is staging-hybrid only.

### Next

**One step:** P2-narrow on this frozen store — unlock when-event dates so hybrid can keep weekday-relative phrasing; lock typed counts, dual-entity enumerated lists, and short typed lists that hybrid only expands. Remasure skip-ingest product 180 (`locomo-s0-diag-mh-135-p2`). Keep where/polar locked. Fair Mem0 180 continues in parallel. P3 follows the ledger (SH PROOF 37 is still the mass). n=1540 only after a stratified **delta** that is not just 19→37. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-22 — P2-narrow (length-lock 56/180, then extras/hop-ground 61/180)

### Landed

P2-narrow on the same frozen `diag-mh-135` + conv-30 store (dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`, `BRAINY_RECALL_LLM=1`). Go **default for hybrid stays off**. `#133` / `#131` stay unmerged. `main` is not fast-forwarded. Branch `pr/s0-current-sha-baseline-1e9e` (PR #136).

1. Length-lock (`681028e`): unlock when-event dates; lock typed `count_answer`; lock dual-entity joins even in `mode=answer`; lock enumerated lists unless hybrid shortens them. Pin: [locomo-s0-diag-mh-135-p2-20260822.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p2-20260822.md) — **56/180**.
2. Extras coverage + skip hop-ground (`fac229f`, `fb41ece`): prefer comma-split typed answers (not long evidence `Items` + token overlap); lock multi-item extras / short-list expansion; **do not** `groundToHopValues` enumerated hybrid lists. Pin: [locomo-s0-diag-mh-135-p2b-20260822.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p2b-20260822.md) — **61/180**.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Not re-run; merge gate stands. |
| Marketing vertical | **17/17** | Not re-run; merge gate stands. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P1 | **37/180 (0.206)** | MH **10/33 (dip)** · OD **1/11** · SH **19/98** · temporal **7/38**. SHA `3d42b17`. |
| LoCoMo S0 product hybrid **on** P2 length-lock | **56/180 (0.311)** | MH **11/33** · OD **1/11** · SH **23/98** · temporal **21/38**. SHA `681028e`. Ledger: **PROOF 41 / RETRIEVAL 39 / READER 33 / WRITE 10 / HARNESS 1**. |
| LoCoMo S0 product hybrid **on** P2b | **61/180 (0.339)** | MH **16/33** · OD **1/11** · SH **25/98** · temporal **19/38 (dip vs P2 21/38)**. SHA `fb41ece`. Ledger: **PROOF 42 / RETRIEVAL 38 / READER 28 / WRITE 10 / HARNESS 1**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 61/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. MH 12→16 on this tenant **closes the P1 dip**. Temporal 21→19 vs length-lock is a **dip**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`--system mem0 --top-k 200`, chunk 1, v3 add `/v3/memories/add/` + event wait) is still running (long per-conversation ingest). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **37** P1 → **56** P2 → **61** P2b | **no same-n pin yet** | Product 19→61 vs itself. Still trails this-VM industry **62/180** by 1. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin yet** | Same-pin lane vs Mem0 once the fair 180 lands. |
| S0 MH product (this tenant) | reader-off **12/33** → P1 **10/33** → P2 **11/33** → P2b **16/33** | no 33-item freeze | **Dip closed.** Product MH now leads this-VM industry MH **10/33**. Not a Mem0 same-pin. |
| Search p50 | P2 product 137 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (dip closed 12→16).** P1 hybrid overwrote typed lists with hop-slot dumps. P2 length-lock was not enough: extras vs typed was 0 when hybrid returned a subset, then `groundToHopValues` re-expanded `clarinet, violin` into pottery/beach. Skipping hop-ground on enumerated hybrid answers recovered instruments, snacks, and the Tim+John typed join (judge-accepted; extra Harry Potter values remain in the typed string). Two further MH gains (Maria fundraiser events; Tim most-visited country). Do not treat 16/33 as n=1540 MH.

**Open-domain.** Stuck at **1/11**. Industry **3/11**. Still a trail axis vs published Mem0 OD on the 1×30 freeze. Do not restore OD by stuffing episodes.

**Single-hop.** **5→19→23→25/98**. Industry **27/98**. Net +2 vs P2 (gained snake names / Tokyo location / Dave car restoration / Nate movies; lost Nike-Gatorade deals and Boston visit). SH **PROOF_MISS 36** is still the largest cell.

**Temporal (dip 21→19 vs P2).** P2 length-lock was the real calendar-unlock move (7→21). P2b lost mentorship (10 July vs weekend of 15–16 July) and an art-show day. Industry **22/38**. Name the dip. Do not add LoCoMo-named date rules to chase the two items.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (not re-run)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.**

#### 3. Marketing vertical — lead (not re-run)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured.

### Why

The P2b MH lift is **not** new compiler coverage. Hybrid already produced the short list; hop-slot grounding was throwing it away. Temporal 2→21 was the date unlock + hybrid relative phrasing (P2). P2b did not improve temporal and dipped 21→19. PROOF stays ~42 and WRITE stays 10 — do not merge #133. Reader-off 19/180 remains the labeled no-LLM product pin; 61/180 is staging-hybrid only.

### Next

**One step:** P3 on this frozen store, allocated by the P2b ledger — SH **PROOF_MISS 36** is the mass. Keep where/polar locked unless a later explicit step. Fair Mem0 180 continues in parallel (do not call it handicapped until org/project IDs are tried if the 180 is weak). n=1540 only after a stratified **delta** that is not just 19→37 or 19→61 **and** S6. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-22 — P3 distinctive query-token admit (73/180)

### Landed

P3 on the same frozen `diag-mh-135` + conv-30 store (dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`, `BRAINY_RECALL_LLM=1`). SHA `5bc28ea` on branch `pr/s0-current-sha-baseline-1e9e` (PR #136). Go **default for hybrid stays off**. `#133` / `#131` stay unmerged. `main` is not fast-forwarded.

Product change: admit leftover distinctive query tokens into the candidate pool and evidence set; prepend extras that cover those tokens instead of original-first top-k; second-pass on the uncovered token when hop join is unproven; do not compose, ground, or inject `search_fallback` hop dumps into the hybrid prompt. No LoCoMo-named rules. Pin: [locomo-s0-diag-mh-135-p3-20260822.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p3-20260822.md) — **73/180**.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P1 | **37/180 (0.206)** | MH **10/33 (dip)** · OD **1/11** · SH **19/98** · temporal **7/38**. SHA `3d42b17`. |
| LoCoMo S0 product hybrid **on** P2 length-lock | **56/180 (0.311)** | MH **11/33** · OD **1/11** · SH **23/98** · temporal **21/38**. SHA `681028e`. |
| LoCoMo S0 product hybrid **on** P2b | **61/180 (0.339)** | MH **16/33** · OD **1/11** · SH **25/98** · temporal **19/38 (dip vs P2 21/38)**. SHA `fb41ece`. |
| LoCoMo S0 product hybrid **on** P3 | **73/180 (0.406)** | MH **16/33** · OD **3/11** · SH **32/98** · temporal **22/38**. SHA `5bc28ea`. Ledger: **READER 34 / PROOF 34 / RETRIEVAL 29 / WRITE 8 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 73/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. MH 16/33 **held** (1-for-1 swap). SH READER 15→22 is a **dip** in that cell even as overall SH rose.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`--system mem0 --top-k 200`, chunk 1, v3 add `/v3/memories/add/` + event wait) is still running (reached conv-30 after a long conv-26 ingest). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **37** P1 → **56** P2 → **61** P2b → **73** P3 | **no same-n pin yet** | Product 19→73 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin yet** | Same-pin lane vs Mem0 once the fair 180 lands. |
| S0 MH product (this tenant) | reader-off **12/33** → P1 **10/33** → P2 **11/33** → P2b/P3 **16/33** | no 33-item freeze | **Held** at 16. Product MH still leads this-VM industry MH **10/33**. Named P3 MH loss: `conv-26-q52` pet names. |
| Search p50 | P3 product ~141 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (held 16/33).** P2b skip-hop-ground recovered instruments/snacks/Tim+John. P3 did not retune MH joins. One gain (`conv-48-q73` polar yes) and one loss (`conv-26-q52` Oliver/Luna/Bailey). Do not treat 16/33 as n=1540 MH. Must not regress clarinet/violin, snacks, Ferrari count=2.

**Open-domain.** **1→3/11**. Industry **3/11**. Still a trail axis vs published Mem0 OD on the 1×30 freeze. Gains include a titled-work item. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **5→19→23→25→32/98**. Industry **27/98**. Mechanism: distinctive-token admit surfaced compiled facts (strawberry filling, joined a gym, chili cook-off, cozy/comfortable). SH **PROOF 36→28** and SH **RETRIEVAL 19→12**; SH **READER 15→22 is a dip** (more items now fail at the reader after retrieval improved). Wheel of Time and self-care stay misses.

**Temporal.** **19→22/38**, matching this-VM industry **22/38**. Recovers P2b mentorship (`conv-26-q36` weekend before 17 July). New loss `conv-44-q38` (weekend before 24 Oct). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `5bc28ea` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured.

### Why

Compiled SH facts were in `memory_records` but starved from top-k: `plainto_tsquery` ANDs every query term, and ILIKE fallback is recency-capped on common person names. Distinctive leftover tokens (`filling`, `gym`) never entered the 30, so hybrid abstained or a nearby chocolate-cake neighbor won. Unproven hop dumps then crowded the prompt. Token admit + unproven-hop skip is retrieval/prompt, not compiler fishing. WRITE 10→8 is incidental; do not merge #133. Reader-off 19/180 remains the labeled no-LLM product pin; 73/180 is staging-hybrid only.

### Next

**One step:** remaining mass is SH **PROOF 28 + READER 22**. Fair Mem0 180 is still the same-pin (do not call it handicapped until org/project IDs are tried if the 180 is weak). n=1540 only at S6 — 19→73 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-22 — P4 identity/garbage hybrid (79/180)

### Landed

P4 on the same frozen `diag-mh-135` + conv-30 store (dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`, `BRAINY_RECALL_LLM=1`). SHA `6f74024` on branch `pr/s0-current-sha-baseline-1e9e` (PR #136). Go **default for hybrid stays off**. `#133` / `#131` stay unmerged. `main` is not fast-forwarded.

Product change: reject punctuation-only / low-letter / single-rune hybrid answers; skip **identity-only** Structured dumps that miss leftover distinctive query tokens; keep only covering memories in that prompt; do not hop-ground hybrid answers onto those identity dumps. Skill / possession / dual-entity hops stay. No LoCoMo-named rules. Pin: [locomo-s0-diag-mh-135-p4-20260822.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p4-20260822.md) — **79/180**.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P1 | **37/180 (0.206)** | MH **10/33 (dip)** · OD **1/11** · SH **19/98** · temporal **7/38**. SHA `3d42b17`. |
| LoCoMo S0 product hybrid **on** P2 length-lock | **56/180 (0.311)** | MH **11/33** · OD **1/11** · SH **23/98** · temporal **21/38**. SHA `681028e`. |
| LoCoMo S0 product hybrid **on** P2b | **61/180 (0.339)** | MH **16/33** · OD **1/11** · SH **25/98** · temporal **19/38 (dip vs P2 21/38)**. SHA `fb41ece`. |
| LoCoMo S0 product hybrid **on** P3 | **73/180 (0.406)** | MH **16/33** · OD **3/11** · SH **32/98** · temporal **22/38**. SHA `5bc28ea`. |
| LoCoMo S0 product hybrid **on** P4 | **79/180 (0.439)** | MH **16/33** · OD **3/11** · SH **37/98** · temporal **23/38**. SHA `6f74024`. Ledger: **PROOF 32 / READER 30 / RETRIEVAL 30 / WRITE 7 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 79/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. MH 16/33 **held** with a named 1-for-1 swap. Item flips vs P3: **+9 / −3 = net +6**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`--system mem0 --top-k 200`, chunk 1, v3 add `/v3/memories/add/` + event wait) **crashed** after scoring conv-26 and part of conv-30: one `POST /v3/memories/add/` event stayed non-SUCCEEDED past the adapter's hardcoded **300s** wait (`TimeoutError` on event `66f277d4-…`). That is a harness timeout, not a product `/recall` miss. The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 79 vs 11, or 79 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **37** P1 → **56** P2 → **61** P2b → **73** P3 → **79** P4 | **no same-n pin yet** (fair 180 crashed mid-ingest) | Product 19→79 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin yet** | Same-pin lane vs Mem0 once the fair 180 lands. |
| S0 MH product (this tenant) | reader-off **12/33** → P1 **10/33** → P2 **11/33** → P2b/P3/P4 **16/33** | no 33-item freeze | **Held** at 16. Product MH still leads this-VM industry MH **10/33**. Named P4 MH gain: `conv-26-q52` Oliver/Luna/Bailey. Named P4 MH loss: `conv-43-q38` Tim UK. |
| Search p50 | P4 product ~156.5 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (held 16/33).** P4 recovered the P3 named MH loss (`conv-26-q52` pet names) by rejecting `!!!!` garbage. Named new loss: `conv-43-q38` Tim most-visited country (United Kingdom → visa/travel identity dump). Do not treat 16/33 as n=1540 MH. Must not regress clarinet/violin, snacks, Ferrari count=2.

**Open-domain.** **Held 3/11**. Industry **3/11**. Gain `conv-47-q12` (James girlfriend → No). Named loss `conv-26-q30` (Melanie LGBTQ No → not in memory). Still a trail axis vs published Mem0 OD on the 1×30 freeze. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **5→19→23→25→32→37/98**. Industry **27/98**. Mechanism: garbage reject + identity-slot skip. Gains include Shadow, snakes Susie/Seraphim, piano, CS:GO tournament, shelter girl (judge-adjacent). SH **PROOF 28→26** and SH **READER 22→19**. Wheel of Time and nearby-wrong language (German vs Spanish) stay misses.

**Temporal.** **22→23/38**, still matching/leading this-VM industry **22/38**. Recovers P3 wine-tasting weekend (`conv-44-q38`). Named loss `conv-41-q53` (2022 → 17 July 2023). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `6f74024` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured.

### Why

Hybrid freeform accepted `!!!!` because garbage detection was sentinel-only; that overwrote typed possession lists. Independently, identity slogans (`Maria is a inspiration`) stayed in the packet after P3's skip of unproven hops, and `groundToHopValues` replaced a covering hybrid name (`Shadow`) because name questions are `mode=answer` (P2b only skipped hop-ground on **enumerated** lists). P4 is reader/prompt/grounding, not compiler fishing. WRITE 8→7 is incidental; do not merge #133. Reader-off 19/180 remains the labeled no-LLM product pin; 79/180 is staging-hybrid only.

### Next

**One step:** remaining mass is SH **PROOF 26 + READER 19**. Restart the fair Mem0 180 with the event-wait timeout wired to `--async-timeout` (do not call it handicapped until org/project IDs are tried if the 180 is weak). n=1540 only at S6 — 19→79 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-22 — P5 activity-dump skip (84/180)

### Landed

P5 on the same frozen `diag-mh-135` + conv-30 store (dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`, `BRAINY_RECALL_LLM=1`). SHA `5ad07c4` on branch `pr/s0-current-sha-baseline-1e9e` (PR #136). Go **default for hybrid stays off**. `#133` / `#131` stay unmerged. `main` is not fast-forwarded.

Product change: skip activity/event hop dumps that miss leftover distinctive query tokens (skill/possession/preference joins stay); keep specific packet facts whose gold is a synonym of the leftover token (United Kingdom / country); do not hop-ground or compose those dumps when hybrid abstains. No LoCoMo-named rules. Pin: [locomo-s0-diag-mh-135-p5-20260822.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p5-20260822.md) — **84/180**.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P1 | **37/180 (0.206)** | MH **10/33 (dip)** · OD **1/11** · SH **19/98** · temporal **7/38**. SHA `3d42b17`. |
| LoCoMo S0 product hybrid **on** P2 length-lock | **56/180 (0.311)** | MH **11/33** · OD **1/11** · SH **23/98** · temporal **21/38**. SHA `681028e`. |
| LoCoMo S0 product hybrid **on** P2b | **61/180 (0.339)** | MH **16/33** · OD **1/11** · SH **25/98** · temporal **19/38 (dip vs P2 21/38)**. SHA `fb41ece`. |
| LoCoMo S0 product hybrid **on** P3 | **73/180 (0.406)** | MH **16/33** · OD **3/11** · SH **32/98** · temporal **22/38**. SHA `5bc28ea`. |
| LoCoMo S0 product hybrid **on** P4 | **79/180 (0.439)** | MH **16/33** · OD **3/11** · SH **37/98** · temporal **23/38**. SHA `6f74024`. |
| LoCoMo S0 product hybrid **on** P5 | **84/180 (0.467)** | MH **17/33** · OD **2/11 (dip)** · SH **45/98** · temporal **20/38 (dip vs P4 23/38)**. SHA `5ad07c4`. Ledger: **PROOF 32 / RETRIEVAL 29 / READER 26 / WRITE 7 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 84/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. Temporal **23→20** and OD **3→2** are named dips. Item flips vs P4: **+10 / −5 = net +5**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`--system mem0 --top-k 200`, chunk 1, v3 add `/v3/memories/add/` + event wait wired to `--async-timeout 1800`) was restarted as `locomo-s0-mem0-v3-s1-fair2` after the prior run crashed on a hardcoded 300s event wait. The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 84 vs 11, or 84 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **37** P1 → **56** P2 → **61** P2b → **73** P3 → **79** P4 → **84** P5 | **no same-n pin yet** (fair 180 restarted) | Product 19→84 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin yet** | Same-pin lane vs Mem0 once the fair 180 lands. |
| S0 MH product (this tenant) | reader-off **12/33** → P1 **10/33** → P2 **11/33** → P2b/P3/P4 **16/33** → P5 **17/33** | no 33-item freeze | Tim-UK recovered. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P5 product ~147 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (17/33).** P5 recovered the P4 named MH loss (`conv-43-q38` Tim UK). Locks held: clarinet/violin, snacks, Ferrari count=2, pet names. Do not treat 17/33 as n=1540 MH.

**Open-domain.** **3→2/11 dip**. Named loss `conv-47-q12` (James girlfriend No → not in memory). Industry **3/11**. Do not restore OD by stuffing episodes.

**Single-hop.** **5→19→23→25→32→37→45/98**. Industry **27/98**. Mechanism: skip activity/event slogan dumps. Gains include Max, school funding, memorials, winter reading, seltzer/chocolate, dream book. SH **READER 19→12**. SH **PROOF 26→25**.

**Temporal.** **23→20/38 dip** vs P4 and vs this-VM industry **22/38**. Named losses: mentorship weekend (`conv-26-q36`, extra clause the judge rejected), biking date (`conv-26-q67`), Pacific Northwest 2022 (`conv-41-q20` → not in memory). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `5ad07c4` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured.

### Why

P4 skipped identity-only hops. Typed **activity** hops (`visa requirements, explore nature, traveling`) are not identity, so they still led the hybrid prompt. The covering UK fact was in the packet; the reader abstained; `composeFromHopValues` dumped the slogans. P5 skips those dumps unless hops are a skill/possession/preference join, and keeps non-crowded packet facts whose gold is a synonym of the leftover token. WRITE stays 7; do not merge #133. Reader-off 19/180 remains the labeled no-LLM product pin; 84/180 is staging-hybrid only.

### Next

**One step:** remaining mass is SH **PROOF 25**. Name the temporal 23→20 dip; do not add date rules. Fair Mem0 180 (`fair2`) is the same-pin (still running). n=1540 only at S6 — 19→84 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-22 — P6 dump-lock skip (87/180)

**Landed:** product SHA `45a83b5` on `pr/s0-current-sha-baseline-1e9e` (draft PR #136). Staging `dev` remains `453a929`. **Not** merged to `main`. Skip-ingest pin [locomo-s0-diag-mh-135-p6-20260822.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p6-20260822.md) (`locomo-s0-diag-mh-135-p6-product-recall-s1-b22074`).

Product change: two-name single-hop questions were planned as multi-hop, so activity slogan dumps stayed in the hybrid prompt and `mh_list` / `where` locks kept them over covering answers. Skip dual-entity **activity** dumps unless hops are a typed skill/possession/preference join; unlock hybrid when the typed answer is a hop dump; cap the hybrid prompt; promote proper-noun/venue facts ahead of generic leftover-cover; do not compose crowded hop dumps when hybrid abstains.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P1 | **37/180 (0.206)** | MH **10/33 (dip)** · OD **1/11** · SH **19/98** · temporal **7/38**. SHA `3d42b17`. |
| LoCoMo S0 product hybrid **on** P2 length-lock | **56/180 (0.311)** | MH **11/33** · OD **1/11** · SH **23/98** · temporal **21/38**. SHA `681028e`. |
| LoCoMo S0 product hybrid **on** P2b | **61/180 (0.339)** | MH **16/33** · OD **1/11** · SH **25/98** · temporal **19/38 (dip vs P2 21/38)**. SHA `fb41ece`. |
| LoCoMo S0 product hybrid **on** P3 | **73/180 (0.406)** | MH **16/33** · OD **3/11** · SH **32/98** · temporal **22/38**. SHA `5bc28ea`. |
| LoCoMo S0 product hybrid **on** P4 | **79/180 (0.439)** | MH **16/33** · OD **3/11** · SH **37/98** · temporal **23/38**. SHA `6f74024`. |
| LoCoMo S0 product hybrid **on** P5 | **84/180 (0.467)** | MH **17/33** · OD **2/11 (dip)** · SH **45/98** · temporal **20/38 (dip vs P4 23/38)**. SHA `5ad07c4`. |
| LoCoMo S0 product hybrid **on** P6 | **87/180 (0.483)** | MH **13/33 (dip vs P5 17/33)** · OD **3/11** · SH **52/98** · temporal **19/38 (dip vs P5 20/38)**. SHA `45a83b5`. Ledger: **READER 29 / RETRIEVAL 28 / PROOF 28 / WRITE 6 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 87/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. MH **17→13** and temporal **20→19** are named dips. Item flips vs P5: **+12 / −9 = net +3**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died during conv-26 ingest on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 87 vs 11, or 87 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **37** P1 → **56** P2 → **61** P2b → **73** P3 → **79** P4 → **84** P5 → **87** P6 | **no same-n pin** (fair 180 429) | Product 19→87 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P5 **17/33** → P6 **13/33 dip** | no 33-item freeze | Named MH dip. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P6 product ~150 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (13/33 dip).** Named losses: fundraiser events incomplete (`conv-41-q29`); Tim/John signed basketball not in memory (`conv-43-q25`); mother's hobbies incomplete (`conv-48-q14`); Phuket diving not in memory (`conv-48-q77`); family injured incomplete (`conv-49-q49`). Tim-UK **held**. Dual-entity activity skip recovered SH walking but over-trimmed some typed MH joins. Do not treat 13/33 as n=1540 MH.

**Open-domain.** **2→3/11**. Named gain `conv-26-q30` Melanie LGBTQ (P5 abstain). Industry **3/11**. Do not restore OD by stuffing episodes.

**Single-hop.** **5→19→23→25→32→37→45→52/98**. Industry **27/98**. Mechanism: dump-lock skip + place-fact ranking. Named gains: walking, Vancouver, Yoga, pregnant, festival. SH **PROOF 25→22**.

**Temporal.** **20→19/38 dip** vs P5 and vs this-VM industry **22/38**. Named losses: Gina internship date (`conv-30-q19` → not in memory), wine-tasting extra clause (`conv-44-q38`), James adopted Ned (`conv-47-q10` 17 March vs first week of April). Named recoveries vs P5: mentorship weekend (`conv-26-q36`), Pacific Northwest 2022 (`conv-41-q20`). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `45a83b5` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

P5 skipped activity dumps for one entity. Two-name SH questions still planned as MH, so typed sushi/title-case dumps locked out covering hybrid answers (walking, Vancouver). P6 skips those dual-entity activity dumps, unlocks hybrid when the typed answer is itself a dump, and promotes proper-noun/venue lines so leftover-cover cannot fill the prompt cap with visa "countries he wants to visit" and drop United Kingdom. The MH dip is the same mechanism over-firing on typed dual-entity joins (signed basketball, Phuket, fundraiser lists). WRITE 7→6; do not merge #133. Reader-off 19/180 remains the labeled no-LLM product pin; 87/180 is staging-hybrid only.

### Next

**One step:** recover MH **17/33** without giving back SH **52/98** (keep Tim-UK, walking, Vancouver, Yoga, pregnant). Remaining mass is SH **PROOF 22**. Name the MH 17→13 dip; do not add date rules. Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→87 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-22 — P7 hop-local joins (88/180)

**Landed:** product SHA `f3e0a7f` on `pr/s0-current-sha-baseline-1e9e` (draft PR #136). Staging `dev` remains `453a929`. **Not** merged to `main`. Skip-ingest pin [locomo-s0-diag-mh-135-p7-20260822.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p7-20260822.md) (`locomo-s0-diag-mh-135-p7-product-recall-s1-fed44d`).

Product change: keep leftover-covering hop contents when skipping Structured dumps (including identity-only leftover names); rare-share dual-entity possessions from typed Values then omitted hop-content snippets; score rare tokens by shortest matching value so extra court-photo df cannot beat basketball; keep title-cased typed joins when hybrid abstains without restoring identity slogans; where-answers only from locative leftover-covering places.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P5 | **84/180 (0.467)** | MH **17/33** · OD **2/11 (dip)** · SH **45/98** · temporal **20/38**. SHA `5ad07c4`. |
| LoCoMo S0 product hybrid **on** P6 | **87/180 (0.483)** | MH **13/33 (dip vs P5 17/33)** · OD **3/11** · SH **52/98** · temporal **19/38 (dip vs P5 20/38)**. SHA `45a83b5`. |
| LoCoMo S0 product hybrid **on** P7 | **88/180 (0.489)** | MH **14/33** · OD **4/11** · SH **49/98 (dip vs P6 52/98)** · temporal **21/38**. SHA `f3e0a7f`. Ledger: **PROOF 29 / RETRIEVAL 29 / READER 27 / WRITE 5 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 88/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. SH **52→49** is a named dip. MH **17→14** vs P5 remains a dip. Item flips vs P6: **+9 / −8 = net +1**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 88 vs 11, or 88 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **87** P6 → **88** P7 | **no same-n pin** (fair 180 429) | Product 19→88 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P5 **17/33** → P6 **13/33** → P7 **14/33** | no 33-item freeze | Still a dip vs P5 17. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P7 product ~150–230 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (14/33).** Named recoveries vs P6: chili cook-off + ring-toss (`conv-41-q29`); signed basketball (`conv-43-q25`); family injured (`conv-49-q49`). Still missing: mother's hobbies yoga extra (`conv-48-q14`); Phuket diving not in memory (`conv-48-q77`). Tim-UK **held**. Dual-entity activity skip still holds walking. Do not treat 14/33 as n=1540 MH.

**Open-domain.** **3→4/11**. Named gain `conv-42-q4` (pets without fur). Named loss `conv-26-q30` Melanie LGBTQ (P6 hit). Industry **3/11**. Do not restore OD by stuffing episodes.

**Single-hop.** **5→…→52→49/98 dip**. Industry **27/98**. Named losses: Shadow second-puppy abstain (`conv-41-q147`); Frank Ocean festival → "Rocks" (`conv-50-q125`); CS:GO tournament not in memory (`conv-47-q88`). Walking **held**. SH **PROOF 22→23**.

**Temporal.** **19→21/38** recovers the P6 dip vs P5 20. Named recoveries: Gina internship date (`conv-30-q19`), wine-tasting weekend (`conv-44-q38`), James in Toronto (`conv-47-q32`). Named loss: Joanna ice cream weekend (`conv-42-q29` → not in memory). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `f3e0a7f` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

P6 skipped dual-entity activity dumps and recovered SH walking, but rare-share only looked at typed Values (signed basketball lived in hop contents) and hybrid abstain treated title-cased possession joins as slogan dumps. P7 keeps leftover-covering hop contents under skip, rare-shares omitted possession snippets, scores by shortest matching value so court-photo df cannot beat basketball, and keeps those joins on hybrid abstain without restoring identity slogans. The SH dip is identity leftover skip dropping Shadow plus festival/place compose ("Rocks"). WRITE 6→5; do not merge #133. Reader-off 19/180 remains the labeled no-LLM product pin; 88/180 is staging-hybrid only.

### Next

**One step:** recover SH **52/98** (Shadow, festival, CS:GO) without giving back basketball / chili / walking / UK / gym. Remaining mass is SH **PROOF 23**. MH **17→14** vs P5 is still a named dip (hobbies yoga extra, Phuket write split). Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→88 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-22 — P8 SH recovery without dump restore (93/180)

**Landed:** product SHA `86eab77` on `pr/s0-current-sha-baseline-1e9e` (draft PR #136). Staging `dev` remains `453a929`. **Not** merged to `main`. Skip-ingest pin [locomo-s0-diag-mh-135-p8-20260822.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p8-20260822.md) (`locomo-s0-diag-mh-135-p8b-product-recall-s1-6b1754`). Product commits on this SHA: `f86ee94` leftover specific facts + attended-event drop; `abab569` dated ordinal names; `86eab77` where-only mh_list unlock. A superseded **92/180** on `abab569` (`…-p8-product-recall-s1-759386`) unlocked mh_list on any skip and lost community yoga+running — **not** this pin.

Product change: unlock skipped `mh_list` only when the query is locative (`where`); drop hop-slot values that are attended-events or foreign possessives; parse dated pet/name lines and lock the ordinal (first/second/third) by dated-then-undated order; on hybrid abstain, keep a leftover-covering specific packet fact (skip chat-turn lines) and strip a trailing date that conflicts with the query date.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P5 | **84/180 (0.467)** | MH **17/33** · OD **2/11 (dip)** · SH **45/98** · temporal **20/38**. SHA `5ad07c4`. |
| LoCoMo S0 product hybrid **on** P6 | **87/180 (0.483)** | MH **13/33 (dip vs P5 17/33)** · OD **3/11** · SH **52/98** · temporal **19/38 (dip vs P5 20/38)**. SHA `45a83b5`. |
| LoCoMo S0 product hybrid **on** P7 | **88/180 (0.489)** | MH **14/33** · OD **4/11** · SH **49/98 (dip vs P6 52/98)** · temporal **21/38**. SHA `f3e0a7f`. |
| LoCoMo S0 product hybrid **on** P8 | **93/180 (0.517)** | MH **15/33** · OD **4/11** · SH **52/98** · temporal **22/38**. SHA `86eab77`. Ledger: **RETRIEVAL 29 / PROOF 28 / READER 24 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 93/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. SH **49→52** recovers the P7 dip (matches P6, not a new SH ceiling). MH **17→15** vs P5 remains a dip. Item flips vs P7: **+5 / −0 = net +5**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 93 vs 11, or 93 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **88** P7 → **93** P8 | **no same-n pin** (fair 180 429) | Product 19→93 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P5 **17/33** → P6 **13/33** → P7 **14/33** → P8 **15/33** | no 33-item freeze | Still a dip vs P5 17. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P8 product ~150–230 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (15/33).** Named recovery vs P7: mother's hobbies (`conv-48-q14`) after dropping attended-event hop values. Chili / basketball / injured / community yoga+running **held**. Still missing: Phuket diving (`conv-48-q77` → not in memory) — write split, do not re-enable two-name skip exemption (restores sushi). Tim-UK **held**. Dual-entity activity skip still holds walking. Do not treat 15/33 as n=1540 MH.

**Open-domain.** **Held 4/11.** Industry **3/11**. Do not restore OD by stuffing episodes.

**Single-hop.** **49→52/98** recovers the P7 dip vs P6 52. Industry **27/98**. Named recoveries: Shadow second-puppy (`conv-41-q147`); Frank Ocean festival (`conv-50-q125`); CS:GO tournament (`conv-47-q88`). Walking **held**. Remaining mass is SH **PROOF 22**.

**Temporal.** **21→22/38**. Named incidental: Joanna ice cream weekend (`conv-42-q29`). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `86eab77` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

P7 kept leftover hop contents and rare-share possessions, which recovered chili/basketball/injured but locked a short unrelated `mh_list` slot ("Rocks") over a locative hybrid, abstained on ordinal second-puppy names, and left CS:GO as a leftover packet fact with a conflicting date tail. P8 unlocks skipped `mh_list` **only** on where-queries, orders named pets dated-then-undated, drops attended-event hop values from kinship lists, and on hybrid abstain keeps a leftover-covering specific fact with the conflicting date stripped. WRITE 5→4; do not merge #133. Reader-off 19/180 remains the labeled no-LLM product pin; 93/180 is staging-hybrid only.

### Next

**One step:** remaining SH **PROOF 22** (nearby-wrong / incomplete compose: wrong event, wrong language, identity dump crowding a titled work) without giving back Shadow / festival / CS:GO / basketball / chili / walking / UK / gym / community yoga+running. MH **17→15** vs P5 is still a named dip (Phuket write split). Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→93 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-22 — P9 unproven mh_list dumps (94/180)

**Landed:** product SHA `bdee669` on `pr/s0-current-sha-baseline-1e9e` (draft PR #136). Staging `dev` remains `453a929`. **Not** merged to `main`. Skip-ingest pin [locomo-s0-diag-mh-135-p9-20260822.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p9-20260822.md) (`locomo-s0-diag-mh-135-p9-product-recall-s1-abbfef`).

Product change: do not mh_list-lock when hops are unproven `search_fallback` dumps; treat 4+ short identity fragment lists and question-echo hop values (`any tips` / trailing `?`) as dumps; leftover covering skips OCR `[a photo …]` captions and stored question prompts. Typed 2-item community/skill joins stay locked.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P8 | **93/180 (0.517)** | MH **15/33** · OD **4/11** · SH **52/98** · temporal **22/38**. SHA `86eab77`. |
| LoCoMo S0 product hybrid **on** P9 | **94/180 (0.522)** | MH **15/33** · OD **4/11** · SH **53/98** · temporal **22/38**. SHA `bdee669`. Ledger: **RETRIEVAL 29 / PROOF 27 / READER 24 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 94/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. MH **17→15** vs P5 remains a dip. Item flips vs P8: **+1 / −0 = net +1**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 94 vs 11, or 94 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **93** P8 → **94** P9 | **no same-n pin** (fair 180 429) | Product 19→94 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P5 **17/33** → P8 **15/33** → P9 **15/33** | no 33-item freeze | Still a dip vs P5 17. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P9 product ~130–230 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (15/33 held).** Chili / basketball / injured / community yoga+running / mother's hobbies **held**. Still missing: Phuket diving (`conv-48-q77` → not in memory) — write split, do not re-enable two-name skip exemption. Tim-UK **held**. Walking **held**. Do not treat 15/33 as n=1540 MH.

**Open-domain.** **Held 4/11.** Industry **3/11**. Do not restore OD by stuffing episodes.

**Single-hop.** **52→53/98**. Industry **27/98**. Named recovery: studying/time-management strategy (`conv-48-q120`). Calvin/Dave goals still incomplete. Shadow / festival / CS:GO **held**. Remaining mass is SH **PROOF 21**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter, pottery hurt, self-care) are not this increment.

**Temporal.** **Held 22/38**. Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `bdee669` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

P8 recovered Shadow / festival / CS:GO but still mh_list-locked unproven search_fallback fragment dumps and stored question-echo hop values over covering hybrid facts. P9 unlocks those unproven dumps and skips OCR captions / question prompts in leftover covering. Typed 2-item joins stay locked, so community yoga+running and basketball held. WRITE held at 4; do not merge #133. Reader-off 19/180 remains the labeled no-LLM product pin; 94/180 is staging-hybrid only.

### Next

**One step:** remaining SH **PROOF 20** (nearby-wrong hybrid, incomplete dual-entity compose) without giving back studying / Shadow / festival / CS:GO / basketball / chili / walking / UK / gym / community yoga+running / Sapiens / retreat. MH **17→15** vs P5 is still a named dip (Phuket write split). Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→96 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-23 — P10 date-aware leftover covering (96/180)

**Landed:** product SHA `e461d70` on `pr/s0-current-sha-baseline-1e9e` (draft PR #136). Staging `dev` remains `453a929`. **Not** merged to `main`. Skip-ingest pin [locomo-s0-diag-mh-135-p10-20260823.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p10-20260823.md) (`locomo-s0-diag-mh-135-p10c-product-recall-s1-8d5416`). An earlier P10 run on `24c226b` was **94/180 churn** (temporal −2) and is not this pin.

Product change: leftover covering skips day-specific queries whose line primary event date is more than 10 days away, so session-relative tails cannot make a January fact match a February question; last-week session news stays. Hybrid packets use a 48h window except where-queries. Speaker-prefixed leftover covering counts only when the body covers leftover tokens.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P9 | **94/180 (0.522)** | MH **15/33** · OD **4/11** · SH **53/98** · temporal **22/38**. SHA `bdee669`. |
| LoCoMo S0 product hybrid **on** P10 | **96/180 (0.533)** | MH **15/33** · OD **4/11** · SH **55/98** · temporal **22/38**. SHA `e461d70`. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 23 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 96/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. MH **17→15** vs P5 remains a dip. Item flips vs P9: **+2 / −0 = net +2**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 96 vs 11, or 96 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **94** P9 → **96** P10 | **no same-n pin** (fair 180 429) | Product 19→96 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P5 **17/33** → P9 **15/33** → P10 **15/33** | no 33-item freeze | Still a dip vs P5 17. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P10 product ~50–180 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (15/33 held).** Chili / basketball / injured / community yoga+running **held**. Still missing: Phuket diving (`conv-48-q77` → not in memory) — write split, do not re-enable two-name skip exemption. Tim-UK **held**. Walking **held**. Do not treat 15/33 as n=1540 MH.

**Open-domain.** **Held 4/11.** Industry **3/11**. Do not restore OD by stuffing episodes.

**Single-hop.** **53→55/98**. Industry **27/98**. Named recoveries: Sapiens (`conv-48-q106`), retreat neat solutions (`conv-48-q109`). Shadow / festival / CS:GO / gym / studying **held**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **Held 22/38** (ice-cream weekend and Toronto held after the churn revision). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `e461d70` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

P9 leftover covering matched session-relative tails (`the week before 4 February`) so a January Stephenson fact beat covering February facts already in the packet (Sapiens chat turn; retreat accomplishment). P10 scores leftover covering by primary event date (10-day window keeps last-week gym news) and keeps speaker-prefixed lines only when the body covers leftover tokens. Hybrid packets use a tighter 48h window on when/what questions so same-month distractors do not crowd; where-queries are exempt so adjacent-day location facts (Toronto July 11 vs July 12) stay. Typed 2-item joins stay locked. WRITE held at 4; do not merge #133.

### Next

**One step:** remaining SH **PROOF 20** (nearby-wrong hybrid, incomplete dual-entity compose) without giving back Sapiens / retreat / studying / Shadow / festival / CS:GO / basketball / chili / walking / UK / gym / community yoga+running. MH **17→15** vs P5 is still a named dip (Phuket write split). Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→96 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-23 — P11 locative leftover covering (97/180)

**Landed:** product SHA `bc6dc92` on `pr/s0-current-sha-baseline-1e9e` (draft PR #136). Staging `dev` remains `453a929`. **Not** merged to `main`. Skip-ingest pin [locomo-s0-diag-mh-135-p11-20260823.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p11-20260823.md) (`locomo-s0-diag-mh-135-p11-product-recall-s1-253ab1`).

Product change: where leftover covering ignores hop slots so locative packet facts are not starved by activity dumps; leftover covering scores strong leftover tokens; hybrid leftover overwrite is limited to locative queries and games-played joins; thin leftover-miss slogans unlock `mh_list`.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P10 | **96/180 (0.533)** | MH **15/33** · OD **4/11** · SH **55/98** · temporal **22/38**. SHA `e461d70`. |
| LoCoMo S0 product hybrid **on** P11 | **97/180 (0.539)** | MH **13/33 dip** · OD **3/11 dip** · SH **58/98** · temporal **23/38**. SHA `bc6dc92`. Ledger: **RETRIEVAL 29 / PROOF 27 / READER 20 / WRITE 5 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 97/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. MH **17→13** vs P5 is a named dip. Item flips vs P10: **+5 / −4 = net +1**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 97 vs 11, or 97 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **96** P10 → **97** P11 | **no same-n pin** (fair 180 429) | Product 19→97 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P5 **17/33** → P10 **15/33** → P11 **13/33 dip** | no 33-item freeze | Dip vs P10 and vs P5 17. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P11 product ~50–180 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (13/33 dip).** Chili / walking / UK / community yoga+running **held**. Named losses vs P10: signed basketball (`conv-43-q25`), Sam snacks (`conv-49-q15`). Still missing: Phuket diving (`conv-48-q77` → not in memory) — write split, do not re-enable two-name skip exemption. Do not treat 13/33 as n=1540 MH.

**Open-domain.** **4→3/11 dip** (James girlfriend April 2022 → not in memory). Industry **3/11**. Do not restore OD by stuffing episodes.

**Single-hop.** **55→58/98**. Industry **27/98**. Named recoveries: Jasper (`conv-49-q87`), tournament games (`conv-47-q141`), Jolene video games + Susie (`conv-48-q145`). Sapiens / retreat / Shadow / festival / CS:GO / gym / studying **held**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **22→23/38**. Named recoveries: Caroline biking last-weekend, James adopted Ned. Named dip: Toronto July 12 (`conv-47-q32`) — leftover covering `James will depart for Toronto` failed the judge; P10 hybrid `Toronto` passed. Ice-cream weekend **held**. Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `bc6dc92` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

P10 leftover covering still starved locative packet facts when hop dumps mentioned leftover tokens (`family` / `road` / `trip`), so Jasper abstained. P11 ignores hop slots on where leftover covering and requires a locative leftover token so a retreat-in-Place line cannot stand in for a missing diving spot. Games-played leftover covering joins Fortnite/Overwatch/Apex. Hybrid leftover overwrite is limited to where-queries and those joins so Sapiens / chili / studying hybrid holds are not replaced. The Toronto dip is leftover covering replacing a short hybrid place name with a depart-for packet line the judge rejected. WRITE 4→5; do not merge #133.

### Next

**One step:** restore Toronto short hybrid and signed-basketball collectible without giving back Jasper / tournament games / Sapiens / retreat / studying / Shadow / festival / CS:GO / chili / walking / UK / gym / community yoga+running. Then remaining SH **PROOF 20**. MH **17→13** vs P5 is a named dip. Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→97 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

---

## 2026-08-23 — P12 keep short where NPs and typed item joins (101/180)

**Landed:** product SHA `d292a09` on `pr/s0-current-sha-baseline-1e9e` (draft PR #136). Staging `dev` remains `453a929`. **Not** merged to `main`. Skip-ingest pin [locomo-s0-diag-mh-135-p12-20260823.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p12-20260823.md) (`locomo-s0-diag-mh-135-p12-product-recall-s1-c7b010`).

Product change: where leftover covering returns the locative place NP; leftover covering does not beat a short hybrid place name that the covering already contains; short typed item joins (comma or `and`, including title-case possession) stay locked and are not leftover thin misses; leftover covering treats query-verb `support` as a weak token so emotional-support chat cannot stand in for a team fact.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P11 | **97/180 (0.539)** | MH **13/33 dip** · OD **3/11 dip** · SH **58/98** · temporal **23/38**. SHA `bc6dc92`. |
| LoCoMo S0 product hybrid **on** P12 | **101/180 (0.561)** | MH **15/33** · OD **4/11** · SH **58/98** · temporal **24/38**. SHA `d292a09`. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 18 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 101/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. MH **17→15** vs P5 remains a dip. Item flips vs P11: **+4 / −0 = net +4**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 101 vs 11, or 101 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **97** P11 → **101** P12 | **no same-n pin** (fair 180 429) | Product 19→101 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P5 **17/33** → P11 **13/33 dip** → P12 **15/33** | no 33-item freeze | Restored vs P11; still a dip vs P5 17. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P12 product ~110–180 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (15/33).** Chili / walking / UK / community yoga+running **held**. Named recoveries vs P11: signed basketball (`conv-43-q25`), Sam snacks (`conv-49-q15`). Still missing: Phuket diving (`conv-48-q77` → not in memory) — write split, do not re-enable two-name skip exemption. Do not treat 15/33 as n=1540 MH.

**Open-domain.** **3→4/11** restored (James girlfriend April 2022). Industry **3/11**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **Held 58/98**. Industry **27/98**. Jasper / tournament games / Jolene video games + Susie / Sapiens / retreat / Shadow / festival / CS:GO / gym / studying **held**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **23→24/38**. Named recovery: Toronto July 12 (`conv-47-q32`, hybrid `Toronto`). Caroline biking last-weekend, James adopted Ned, ice-cream weekend **held**. Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `d292a09` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

P11 leftover covering replaced a short hybrid place (`Toronto`) with a longer depart-for packet line the judge rejected, and treated title-case possession joins plus short snack joins as slogan dumps / thin misses so hybrid could overwrite them with nearby chat. P12 returns the locative place NP from where covering, keeps short typed item joins locked, and weakens leftover `support` so emotional-support chat cannot stand in for a team fact. WRITE 5→4; do not merge #133.

### Next

**One step:** remaining SH **PROOF 20** (nearby-wrong hybrid, incomplete dual-entity compose) without giving back Toronto / signed basketball / snacks / Jasper / tournament games / Sapiens / retreat / studying / Shadow / festival / CS:GO / chili / walking / UK / gym / community yoga+running. Thanksgiving feast still movies. Boston garage still incomplete. MH **17→15** vs P5 is still a named dip (Phuket write split). Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→101 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-23 — P13 gated leftover-vs-hybrid (102/180)

**Landed:** product SHA `50b8e43` on `pr/s0-current-sha-baseline-1e9e` (draft PR #136). Staging `dev` remains `453a929`. **Not** merged to `main`. Skip-ingest pin [locomo-s0-diag-mh-135-p13-20260823.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p13-20260823.md) (`locomo-s0-diag-mh-135-p13c-product-recall-s1-e1b10f`). Unrestricted leftover-vs-hybrid `d179211` was reverted (`f917c30`). A first gated 180 (`6907e6`, **98/180**) is not a pin.

Product change: leftover covering may replace a hybrid answer only when the covering is a schema-activity fact (`enjoys` / `likes` / `loves` / `participates in`) or the query is where / games-played; leftover covering does not locative-boost chat turns; leftover covering re-picks a line that covers the rarest leftover token so a Thanksgiving tradition covering keeps feast and thankful without the movies paraphrase.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P12 | **101/180 (0.561)** | MH **15/33** · OD **4/11** · SH **58/98** · temporal **24/38**. SHA `d292a09`. |
| LoCoMo S0 product hybrid **on** P13 | **102/180 (0.567)** | MH **15/33** · OD **4/11** · SH **59/98** · temporal **24/38**. SHA `50b8e43`. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 17 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 102/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. MH **17→15** vs P5 remains a dip. Item flips vs P12: **+1 / −0 = net +1**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 102 vs 11, or 102 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **101** P12 → **102** P13 | **no same-n pin** (fair 180 429) | Product 19→102 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P5 **17/33** → P12 **15/33** → P13 **15/33** | no 33-item freeze | Held vs P12; still a dip vs P5 17. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P13 product ~140 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (15/33).** Chili / walking / UK / community yoga+running / signed basketball / snacks **held**. Still missing: Phuket diving (`conv-48-q77` → not in memory) — write split. Do not treat 15/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **58→59/98**. Industry **27/98**. Named recovery: Thanksgiving feast+thankful (`conv-43-q99`). Jasper / tournament games / Jolene video games + Susie / Sapiens / retreat / Shadow / festival / CS:GO / gym / studying **held**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **Held 24/38**. Toronto July 12, Caroline biking last-weekend, James adopted Ned, ice-cream weekend **held**. Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `50b8e43` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

P12 leftover covering could not overwrite a movies hybrid because leftover-vs-hybrid was gated to where/games-played, and a speaker chat turn (`Tim: Thanksgiving's always special for us`) outscored the feast fact via a false locative bonus (Tim: + Thanksgiving's). Unrestricted leftover-vs-hybrid replaced Sapiens / chili / studying / UK / Jolene dumps. A leftover-token-only gate recovered Thanksgiving but lost horseback, letter-feeling, and first-console hybrids (98/180). Restricting leftover-vs-hybrid to schema-activity covering keeps Thanksgiving and those hybrids. WRITE held 4; do not merge #133.

### Next

**One step:** remaining SH **PROOF 20** (nearby-wrong hybrid, incomplete dual-entity compose) without giving back Thanksgiving / Toronto / signed basketball / snacks / Jasper / tournament games / Sapiens / retreat / studying / Shadow / festival / CS:GO / chili / walking / UK / gym / community yoga+running / horseback / first-console. Boston garage still incomplete. MH **17→15** vs P5 is still a named dip (Phuket write split). Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→102 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-23 — P14 childhood possession lock (103/180)

**Landed:** product SHA `90750e5` on `pr/s0-current-sha-baseline-1e9e` (draft PR #136). Staging `dev` remains `453a929`. **Not** merged to `main`. Skip-ingest pin [locomo-s0-diag-mh-135-p14-20260823.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p14-20260823.md) (`locomo-s0-diag-mh-135-p14d-product-recall-s1-21c5ea`). Broad typed-join / hop-slot leftover covering (`ffb36b8` / `0a2011e`, 180 `fd9dbf`, **98/180**) is not a pin.

Product change: lock a 2-item hybrid only for childhood *item* lists (`leftoverCoveringLockChildhoodPossessions`); leftover covering can join `had a` + age-cue packet lines; queries with name/named/names do not take that lock, so a compact hybrid name (`Max`) is not replaced by a hop possession dump.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P13 | **102/180 (0.567)** | MH **15/33** · OD **4/11** · SH **59/98** · temporal **24/38**. SHA `50b8e43`. |
| LoCoMo S0 product hybrid **on** P14 | **103/180 (0.572)** | MH **16/33** · OD **4/11** · SH **59/98** · temporal **24/38**. SHA `90750e5`. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 16 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 103/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. MH **17→16** vs P5 remains a dip. Item flips vs P13: **+1 / −0 = net +1**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 103 vs 11, or 103 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **102** P13 → **103** P14 | **no same-n pin** (fair 180 429) | Product 19→103 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P5 **17/33** → P13 **15/33** → P14 **16/33** | no 33-item freeze | +1 vs P13; still a dip vs P5 17. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P14 product ~140 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (16/33).** Named recovery: childhood items `conv-41-q7` (doll + film camera). Chili / walking / UK / community yoga+running / signed basketball / snacks **held**. Still missing: Phuket diving (`conv-48-q77` → not in memory) — write split. Do not treat 16/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **Held 59/98**. Industry **27/98**. Max (`conv-44-q82`) **held** after the name-query gate. Thanksgiving / Jasper / tournament games / Jolene video games + Susie / Sapiens / retreat / Shadow / festival / CS:GO / gym / studying **held**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **Held 24/38**. Toronto July 12, Caroline biking last-weekend, James adopted Ned, ice-cream weekend **held**. Broad P14 moved two dates and is not a pin. Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `90750e5` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

P13 leftover covering could not keep a 2-item childhood possession list: list-lock required 3 items, and speaker leftover covering treated "having a blast" as covering `having`, so pizza chat replaced doll + film camera. A broad typed-join lock recovered that list but also locked name questions and unrelated possession hops (Max dump, bank, school funding) and hop-slot leftover covering moved two temporal dates (**98/180**). Restricting the lock to childhood *item* queries, and excluding name/named/names, recovers the list without those dumps. WRITE held 4; do not merge #133.

### Next

**One step:** remaining SH **PROOF 20** (nearby-wrong hybrid, incomplete dual-entity compose) without giving back childhood items / Max / Thanksgiving / Toronto / signed basketball / snacks / Jasper / tournament games / Sapiens / retreat / studying / Shadow / festival / CS:GO / chili / walking / UK / gym / community yoga+running / horseback / first-console. Boston garage still incomplete. MH **17→16** vs P5 is still a named dip (Phuket write split). Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→103 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-25 — P17 when-event leftover covering (104/180)

**Landed:** product SHA `4719902` on `pr/s0-current-sha-baseline-1e9e` (draft PR #136). Staging `dev` remains `453a929`. **Not** merged to `main`. Skip-ingest pin [locomo-s0-diag-mh-135-p17-20260825.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p17-20260825.md) (`locomo-s0-diag-mh-135-p17-product-recall-s1-ac10f0`). P15 visit-destination (`3aa9313` … `6d40b93`) and P16 packet-line enrich (`a492922`) each measured **103/180** (Boston +1 / Ned −1) and are **not** pins.

Product change: when-event leftover covering lowers the leftover-token minLen to 4 (so 5-character event verbs such as `adopt` survive) and replaces a short calendar date that misses those tokens with a packet covering line that has them. Hybrids that already name the event stay. Visit-destination keep + packet-line enrich from P15/P16 stay in the SHA (Boston purpose clause).

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P14 | **103/180 (0.572)** | MH **16/33** · OD **4/11** · SH **59/98** · temporal **24/38**. SHA `90750e5`. |
| LoCoMo S0 product hybrid **on** P17 | **104/180 (0.578)** | MH **16/33** · OD **4/11** · SH **60/98** · temporal **24/38**. SHA `4719902`. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 15 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** 90%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 104/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. MH **17→16** vs P5 remains a dip. Item flips vs P14: **+2 / −1 = net +1**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 104 vs 11, or 104 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **103** P14 → **104** P17 | **no same-n pin** (fair 180 429) | Product 19→104 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P5 **17/33** → P14 **16/33** → P17 **16/33** | no 33-item freeze | Held vs P14; still a dip vs P5 17. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P17 product ~185 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (16/33).** Held vs P14. Childhood items `conv-41-q7` **held**. Chili / walking / UK / community yoga+running / signed basketball / snacks **held**. Still missing: Phuket diving (`conv-48-q77` → not in memory) — write split. Do not treat 16/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **59→60/98**. Industry **27/98**. Named recovery: Boston garage+purpose `conv-50-q110`. Max / Thanksgiving / Jasper / tournament games / Jolene video games + Susie / Sapiens / retreat / Shadow / festival / CS:GO / gym / studying **held**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **Held 24/38**. Named recovery: McGee's bar `conv-47-q46`. Ned adoption `2022-04-05` **held** (P16 had lost it to bowling). Named dip: Jon banker job `conv-30-q0` (Gina DoorDash leftover). Toronto July 12, Caroline biking last-weekend, ice-cream weekend, Gina internship 10 May, first-console **held**. Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `4719902` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

P16 leftover covering dropped event verbs shorter than 6 characters unless the query named a calendar day, so `adopt` never scored against bowling `17 March 2022` and Ned's hybrid/leftover became the hop activity date. Lowering when-event minLen to 4 and replacing a short calendar date that misses leftover event tokens recovers Ned and McGee's (a meet-at event whose covering line has the date). The same minLen also admits `lost`/`job`, so leftover covering can pick another person's dated job-loss (Gina DoorDash) over Jon banker — named dip, not a pin-blocker because overall is net +1. Visit-destination keep + packet enrich recover Boston without the P15 dump. WRITE held 4; do not merge #133.

### Next

**One step:** bind when-event leftover covering to a query entity so Jon banker recovers without giving back Ned / Boston / McGee's / Gina internship 10 May. Then leftover unwind-evidence join for destress pottery (packet already has "finds making pottery calming"; do **not** join all `participates in` — camping dump). Remaining SH **PROOF 20**. MH **17→16** vs P5 is still a named dip (Phuket write split). Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→104 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-25 — P18 when-event query-entity bind (105/180)

**Landed:** product SHA `0c03107` on `pr/s0-current-sha-baseline-1e9e` (draft PR #136). Staging `dev` remains `453a929`. **Not** merged to `main`. Skip-ingest pin [locomo-s0-diag-mh-135-p18-20260825.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p18-20260825.md) (`locomo-s0-diag-mh-135-p18-product-recall-s1-4e9e04`).

Product change: when-event leftover covering skips packet lines that name another person and do not name a query person. First-person and unnamed dated lines still compete. Covering that names the query person can still replace a bare date.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P17 | **104/180 (0.578)** | MH **16/33** · OD **4/11** · SH **60/98** · temporal **24/38**. SHA `4719902`. |
| LoCoMo S0 product hybrid **on** P18 | **105/180 (0.583)** | MH **16/33** · OD **4/11** · SH **60/98** · temporal **25/38**. SHA `0c03107`. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 14 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** 90%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 105/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. MH **17→16** vs P5 remains a dip. Item flips vs P17: **+1 / −0 = net +1**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 105 vs 11, or 105 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **104** P17 → **105** P18 | **no same-n pin** (fair 180 429) | Product 19→105 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P5 **17/33** → P17 **16/33** → P18 **16/33** | no 33-item freeze | Held vs P17; still a dip vs P5 17. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P18 product ~159 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (16/33).** Held vs P17. Childhood items `conv-41-q7` **held**. Chili / walking / UK / community yoga+running / signed basketball / snacks **held**. Still missing: Phuket diving (`conv-48-q77` → not in memory) — write split. Do not treat 16/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **Held 60/98**. Industry **27/98**. Boston garage+purpose **held**. Max / Thanksgiving / Jasper / chili+ring-toss / studying **held**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **24→25/38**. Named recovery: Jon banker job `conv-30-q0`. Ned adoption `2022-04-05`, McGee's bar, Toronto July 12, Caroline biking, Gina internship 10 May, first-console, ice-cream weekend **held**. Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `0c03107` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

P17 leftover covering scored short event verbs (`lost`) without binding the covering line to a query person, so Gina's DoorDash job-loss replaced Jon's banker date. Skipping covering lines that name a different person (and do not name a query person) recovers Jon. First-person dated plans without a name still compete, so ice-cream weekend does not fall back to a named pep-talk. Ned / McGee's / Boston covering lines name the query people and stay. WRITE held 4; do not merge #133.

### Next

**One step:** leftover unwind-evidence join for destress pottery (packet already has "finds making pottery calming"; do **not** join all `participates in` — camping dump; `"destress"` stays on the overfit denylist). Remaining SH **PROOF 20**. MH **17→16** vs P5 is still a named dip (Phuket write split). Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→105 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-25 — P20 enumerate unwind extras (106/180)

**Landed:** product SHA `80471d8` on `pr/s0-current-sha-baseline-1e9e` (draft PR #136). Staging `dev` remains `453a929`. **Not** merged to `main`. Skip-ingest pin [locomo-s0-diag-mh-135-p20-20260825.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p20-20260825.md) (`locomo-s0-diag-mh-135-p20-product-recall-s1-36d1d3`). P19 (`dd1fbdd`) and P19b (`de53ca7`) each held 105/180 — not pins.

Product change: unwind-evidenced packet/hop activity slots already joined into `answer`; list-mode `/recall` now also appends those extras onto enumerate `items`. Plain `participates in` without unwind evidence stays out.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; re-run on this SHA before merge. |
| Marketing vertical | **17/17** | Merge gate; re-run on this SHA before merge. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P18 | **105/180 (0.583)** | MH **16/33** · OD **4/11** · SH **60/98** · temporal **25/38**. SHA `0c03107`. |
| LoCoMo S0 product hybrid **on** P20 | **106/180 (0.589)** | MH **17/33** · OD **4/11** · SH **60/98** · temporal **25/38**. SHA `80471d8`. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 13 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** 90%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 106/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. MH **17/33** matches P5 on this axis. Item flips vs P18: **+1 / −0 = net +1**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 106 vs 11, or 106 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **105** P18 → **106** P20 | **no same-n pin** (fair 180 429) | Product 19→106 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P5 **17/33** → P18 **16/33** → P20 **17/33** | no 33-item freeze | Recovers the P5 MH count on this axis. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P20 product ~150 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (17/33).** Recovers P5 **17/33** on this axis. Named recovery: destress `conv-26-q24`. Childhood items `conv-41-q7` **held**. Chili / walking / UK / community yoga+running / signed basketball / snacks **held**. Still missing: Phuket diving (`conv-48-q77`) — write split. Do not treat 17/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **Held 60/98**. Industry **27/98**. Boston garage+purpose **held**. Max / Thanksgiving / Jasper / chili+ring-toss / studying **held**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **Held 25/38**. Jon banker, Ned `2022-04-05`, McGee's bar, Toronto July 12, Caroline biking, Gina internship 10 May, first-console, ice-cream weekend **held**. Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (re-run before merge)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Re-confirm on `80471d8` before merge.

#### 3. Marketing vertical — lead (re-run before merge)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

P19 joined unwind-evidenced calming slots into `answer`. The 180 product lane asks enumerate (`what does `) and the harness serializes `items`, so pottery never reached the judge. Writing the same extras onto enumerate `items` recovers destress without joining plain `participates in` (camping stays out). WRITE held 4; do not merge #133.

### Next

**One step:** remaining READER list joins that already have packet evidence (yoga practice locations, Jolene balance habits, Gina business advice) without category dictionaries. Remaining mass is SH **PROOF 20**. Phuket remains a write split. Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→106 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-25 — P21 location-list leftover lock (107/180)

**Landed:** product SHA `fa915fe` on `pr/locomo-180-p21-1e9e` (draft PR #137). Skip-ingest pin [locomo-s0-diag-mh-135-p21-20260825.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p21-20260825.md) (`locomo-s0-diag-mh-135-p21b-product-recall-s1-7785cc`). First P21 180 (`…-p21-product-recall-s1-f3add8`) was **106/180** with an OD hybrid abstain flake — not a pin.

Product change: location-list leftover covering no longer treats a short locative as a thin miss, covering lines must contain a place, leftover packet practice places join onto answer and items, and which/what-year covering binds to the query person and requires a year token. Does not expand `looksWhenEventQuery` (avoids the hop-date path).

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; last pin. Not re-run this increment. |
| Marketing vertical | **17/17** | Merge gate; last pin. Not re-run this increment. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P20 | **106/180 (0.589)** | MH **17/33** · OD **4/11** · SH **60/98** · temporal **25/38**. SHA `80471d8`. |
| LoCoMo S0 product hybrid **on** P21 | **107/180 (0.594)** | MH **18/33** · OD **4/11** · SH **60/98** · temporal **25/38**. SHA `fa915fe`. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 12 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** 90%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 107/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. Item flips vs P20: **+1 / −0 = net +1**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 107 vs 11, or 107 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **106** P20 → **107** P21 | **no same-n pin** (fair 180 429) | Product 19→107 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P5 **17/33** → P20 **17/33** → P21 **18/33** | no 33-item freeze | New high on this axis. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P21 product ~160 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (18/33).** Named recovery: practice locations `conv-48-q82`. Destress pottery **held**. Childhood items **held**. Still missing: Phuket diving (`conv-48-q77`) — write split. Do not treat 18/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. First P21 180 flake-abstained James girlfriend; live `/recall` and p21b 180 held the P20 no-record answer. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **Held 60/98**. Industry **27/98**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **Held 25/38**. Jon banker, Ned, McGee's, Toronto, Caroline biking, Gina internship, first-console **held**. Jolene yoga year still MISS (2020 not in the leftover packet). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (stale pin)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Not re-run this increment.

#### 3. Marketing vertical — lead (stale pin)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

A short locative list (`the park`) was classified as a leftover thin miss, so yoga-token purchase leftover replaced it. Requiring leftover covering lines to contain a place, and joining leftover packet locatives (park/beach/studio/mother) onto answer and items, recovers the location list. Clothing/room complements stay out. Which-year covering now skips foreign-person and undated lines; that stops Deborah leftover on Jolene's start year but cannot invent 2020 if the year fact is outside the packet. WRITE held 4; do not merge #133.

### Next

**One step:** remaining READER joins with packet evidence (Jolene balance habits, Gina business advice, relative-Sunday charity race, Melanie children count). Remaining mass is SH **PROOF 20** + RETRIEVAL 29 — ranking/compiler so gold enters the top packet. Phuket remains a write split. Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→107 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-25 — P22 leftover month bind (108/180)

**Landed:** product SHA `9d50bad` on `pr/locomo-180-p22-1e9e` (draft PR #138). Skip-ingest pin [locomo-s0-diag-mh-135-p22-20260825.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p22-20260825.md) (`locomo-s0-diag-mh-135-p22-product-recall-s1-72d637`). P21 is already on dest and main (`36f39c2` / `fa915fe`).

Product change: when leftover covering tokens collapse to weak-only after entity/calendar stripping, bind covering to the query month (or year if there is no month). Rarest-token override ignores weak tokens such as `activity`. Does not expand `looksWhenEventQuery`. Does not name LoCoMo.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; last pin. Not re-run this increment. |
| Marketing vertical | **17/17** | Merge gate; last pin. Not re-run this increment. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P21 | **107/180 (0.594)** | MH **18/33** · OD **4/11** · SH **60/98** · temporal **25/38**. SHA `fa915fe`. |
| LoCoMo S0 product hybrid **on** P22 | **108/180 (0.600)** | MH **18/33** · OD **4/11** · SH **60/98** · temporal **26/38**. SHA `9d50bad`. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 11 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** 90%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 108/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. Item flips vs P21: **+1 / −0 = net +1**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 108 vs 11, or 108 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **107** P21 → **108** P22 | **no same-n pin** (fair 180 429) | Product 19→108 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P21 **18/33** → P22 **18/33** | no 33-item freeze | Held P21 high. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P22 product ~150 ms local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (18/33).** Held P21 high. Destress pottery **held**. Yoga locations **held**. Childhood items **held**. Still missing: Phuket diving (`conv-48-q77`) — write split. Do not treat 18/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. James girlfriend April 2022 **held**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **Held 60/98**. Industry **27/98**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **25→26/38**. Named recovery: September co-participant plan `conv-49-q52`. Jon banker, Ned, McGee's, Toronto, Caroline biking, Gina internship, first-console **held**. Jolene yoga year still MISS (2020 not in the leftover packet). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (stale pin)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Not re-run this increment.

#### 3. Marketing vertical — lead (stale pin)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

Activity questions that name people and a month were left with only weak covering tokens (`activity` / `together` / `plan`) after entity and calendar stripping. Rarest-token override then picked generic pep-talk that uniquely said "activity" over a dated co-participant plan already in the leftover packet. Binding covering to the query month in that collapse recovers the dated plan. Weak tokens cannot rarest-override. WRITE held 4; do not merge #133.

### Next

**One step:** remaining READER joins with packet evidence (Jolene balance habits, Gina business advice, relative-Sunday charity race, Melanie children count, community-center 2022). Remaining mass is SH **PROOF 20** + RETRIEVAL 29 — ranking/compiler so gold enters the top packet. Phuket remains a write split. Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→108 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-25 — P23 year/month leftover covering (109/180)

**Landed:** product SHA `e05c78f` on `pr/locomo-180-p23-1e9e` (PR #139). Skip-ingest pin [locomo-s0-diag-mh-135-p23-20260825.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p23-20260825.md) (`locomo-s0-diag-mh-135-p23c-product-recall-s1-f3f3be`). P22 is already on dest and main (`2287726` / `9d50bad`). First P23 180 (`…-p23-…-f9b076`, 108/180 hold) is **not** a pin.

Product change: when-event leftover covering requires a year/date token so year-only event facts can replace a bare hop date; covering from a different query year or month is skipped. Does not expand `looksWhenEventQuery`. Does not name LoCoMo.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; last pin. Not re-run this increment. |
| Marketing vertical | **17/17** | Merge gate; last pin. Not re-run this increment. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P22 | **108/180 (0.600)** | MH **18/33** · OD **4/11** · SH **60/98** · temporal **26/38**. SHA `9d50bad`. |
| LoCoMo S0 product hybrid **on** P23 | **109/180 (0.606)** | MH **18/33** · OD **4/11** · SH **60/98** · temporal **27/38**. SHA `e05c78f`. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 10 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** 90%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 109/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. Item flips vs P22: **+1 / −0 = net +1**. First P23 180 was +1/−1 (community-center / August teammates) and is not a pin.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 109 vs 11, or 109 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **108** P22 → **109** P23 | **no same-n pin** (fair 180 429) | Product 19→109 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P22 **18/33** → P23 **18/33** | no 33-item freeze | Held P21 high. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P23 product local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (18/33).** Held P21 high. Destress pottery **held**. Yoga locations **held**. Childhood items **held**. Still missing: Phuket diving (`conv-48-q77`) — write split. Do not treat 18/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. James girlfriend April 2022 **held**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **Held 60/98**. Industry **27/98**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **26→27/38**. Named recovery: hometown community-center year `conv-41-q53`. August teammates `conv-43-q91` **held** (first P23 180 lost it). Jon banker, Ned, McGee's, Toronto, Caroline biking, Gina internship, first-console, September paint **held**. Jolene yoga year still MISS (2020 not in the leftover packet). Deborah art show still MISS (sentence-initial verb treated as a person). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (stale pin)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Not re-run this increment.

#### 3. Marketing vertical — lead (stale pin)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

Year-only event facts in leftover covering were blocked because the covering date check required `parseDateFromText`, which fails on `in 2022`. Requiring a year token on when-event leftover covering lets those facts replace a bare hop date. Without a year/month match, a 2022 basketball leftover can answer an August 2023 teammates query because `contentCoversQueryToken` stems `teammates`~`team`. Skip covering from a different query year or month. WRITE held 4; do not merge #133.

### Next

**One step:** remaining READER joins with packet evidence (Deborah art-show sentence-initial verb, Jolene balance habits, Gina business advice, relative-Sunday charity race, Melanie children count). Remaining mass is SH **PROOF 20** + RETRIEVAL 29 — ranking/compiler so gold enters the top packet. Phuket remains a write split. Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→109 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-25 — P24 sentence-initial verb covering (110/180)

**Landed:** product SHA `80669d8` on `pr/locomo-180-p24-1e9e` (PR #140). Skip-ingest pin [locomo-s0-diag-mh-135-p24-20260825.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p24-20260825.md) (`locomo-s0-diag-mh-135-p24-product-recall-s1-139aa5`). P23 is already on dest and main (`1d86b66` / `e05c78f`).

Product change: sentence-initial past-tense verbs and phrasal particles are not leftover covering people, so unnamed dated diary lines still compete. Named other people still drop covering. Does not name LoCoMo.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; last pin. Not re-run this increment. |
| Marketing vertical | **17/17** | Merge gate; last pin. Not re-run this increment. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P23 | **109/180 (0.606)** | MH **18/33** · OD **4/11** · SH **60/98** · temporal **27/38**. SHA `e05c78f`. |
| LoCoMo S0 product hybrid **on** P24 | **110/180 (0.611)** | MH **18/33** · OD **4/11** · SH **60/98** · temporal **28/38**. SHA `80669d8`. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 9 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** 90%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 110/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. Item flips vs P23: **+1 / −0 = net +1**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 110 vs 11, or 110 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **109** P23 → **110** P24 | **no same-n pin** (fair 180 429) | Product 19→110 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P23 **18/33** → P24 **18/33** | no 33-item freeze | Held P21 high. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P24 product local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (18/33).** Held P21 high. Destress pottery **held**. Yoga locations **held**. Childhood items **held**. Still missing: Phuket diving (`conv-48-q77`) — write split. Do not treat 18/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. James girlfriend April 2022 **held**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **Held 60/98**. Industry **27/98**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **27→28/38**. Named recovery: unnamed dated art-show diary `conv-48-q44`. Community-center, August teammates, Jon banker, Ned, McGee's, Toronto, Caroline biking, Gina internship, first-console, September paint **held**. Jolene yoga year still MISS (2020 not in the leftover packet). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (stale pin)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Not re-run this increment.

#### 3. Marketing vertical — lead (stale pin)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

Unnamed dated diary lines that start with a capitalized verb were dropped as foreign-person covering. Sentence-initial past-tense morphology and phrasal particles are orthography, not names. Named other people still drop covering. Short names stay people. WRITE held 4; do not merge #133.

### Next

**One step:** remaining READER joins with packet evidence (Jolene balance habits, Gina business advice, relative-Sunday charity race, Melanie children count, John ankle date vs pep-talk, paint-together decide date). Remaining mass is SH **PROOF 20** + RETRIEVAL 29 — ranking/compiler so gold enters the top packet. Phuket remains a write split. Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→110 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-25 — P25 which-year as-of duration (111/180)

**Landed:** product SHA `86d87d6` on `pr/locomo-180-p25-1e9e` (PR #141). Skip-ingest pin [locomo-s0-diag-mh-135-p25-20260825.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p25-20260825.md) (`locomo-s0-diag-mh-135-p25-product-recall-s1-732148`). P24 is already on dest and main (`6743a88` / `80669d8`).

Product change: which-year leftover covering of the form `for N years as of DATE` rewrites to the start year `asOf.Year()-N`. Leftover tokens `year`/`years` are weak so the question word does not bind covering over the event. Does not name LoCoMo.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; last pin. Not re-run this increment. |
| Marketing vertical | **17/17** | Merge gate; last pin. Not re-run this increment. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P24 | **110/180 (0.611)** | MH **18/33** · OD **4/11** · SH **60/98** · temporal **28/38**. SHA `80669d8`. |
| LoCoMo S0 product hybrid **on** P25 | **111/180 (0.617)** | MH **18/33** · OD **4/11** · SH **60/98** · temporal **29/38**. SHA `86d87d6`. Ledger: **RETRIEVAL 28 / PROOF 26 / READER 9 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** 90%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 111/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. Item flips vs P24: **+1 / −0 = net +1**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 111 vs 11, or 111 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **110** P24 → **111** P25 | **no same-n pin** (fair 180 429) | Product 19→111 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P24 **18/33** → P25 **18/33** | no 33-item freeze | Held P21 high. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P25 product local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (18/33).** Held P21 high. Destress pottery **held**. Yoga locations **held**. Childhood items **held**. Still missing: Phuket diving (`conv-48-q77`) — write split. Do not treat 18/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. James girlfriend April 2022 **held**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **Held 60/98**. Industry **27/98**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **28→29/38**. Named recovery: which-year as-of duration `conv-49-q25`. Community-center, August teammates, art-show April, Jon banker, Ned, McGee's, Toronto, Caroline biking, Gina internship, first-console, September paint **held**. Jolene yoga year still MISS (2020 not in the leftover packet). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (stale pin)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Not re-run this increment.

#### 3. Marketing vertical — lead (stale pin)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

Which-year leftover covering returned the as-of calendar year from `for N years as of DATE`. The as-of year is not the year the event started. Rewrite that form to `asOf.Year()-N`. Weak leftover tokens `year`/`years` so the question word does not bind covering over the event (a packet with both a yoga-2020 line and a health duration otherwise lets the duration steal yoga which-year). WRITE held 4; do not merge #133.

### Next

**One step:** ranking so gold enters the leftover packet (Jolene yoga year 2020 is not in covering; paint-together decide date is not in top-30 for that query form). Remaining isolated READER is thin (Jolene balance habits, Gina business advice, relative-Sunday charity race, Melanie children count, John ankle vs pep-talk). Remaining mass is SH **PROOF 20** + RETRIEVAL 28. Isolated leftover covering is saturating. Phuket remains a write split. Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→111 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-25 — P26 when-event speech-act flood (112/180)

**Landed:** product SHA `e5ac883` on `pr/locomo-180-p26-1e9e` (PR #142). Skip-ingest pin [locomo-s0-diag-mh-135-p26-20260825.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p26-20260825.md) (`locomo-s0-diag-mh-135-p26-product-recall-s1-80549a`). P25 is already on dest and main (`e179ce6` / `86d87d6`).

Product change: when/which-year lexical search drops leftover-weak tokens (`decide`/`year`) when another event noun remains, so dated dual-entity plans enter the packet. Leftover covering treats `decide*` as weak and scores lines that name both query people higher than a solo-painter dated dump. Does not require both people. Does not name LoCoMo.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; last pin. Not re-run this increment. |
| Marketing vertical | **17/17** | Merge gate; last pin. Not re-run this increment. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P25 | **111/180 (0.617)** | MH **18/33** · OD **4/11** · SH **60/98** · temporal **29/38**. SHA `86d87d6`. |
| LoCoMo S0 product hybrid **on** P26 | **112/180 (0.622)** | MH **18/33** · OD **4/11** · SH **60/98** · temporal **30/38**. SHA `e5ac883`. Ledger: **RETRIEVAL 28 / PROOF 26 / READER 8 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** 90%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 112/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. Item flips vs P25: **+1 / −0 = net +1**.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 112 vs 11, or 112 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **111** P25 → **112** P26 | **no same-n pin** (fair 180 429) | Product 19→112 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P25 **18/33** → P26 **18/33** | no 33-item freeze | Held P21 high. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P26 product local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (18/33).** Held P21 high. Destress pottery **held**. Yoga locations **held**. Childhood items **held**. Still missing: Phuket diving (`conv-48-q77`) — write split. Do not treat 18/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. James girlfriend April 2022 **held**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **Held 60/98**. Industry **27/98**. Remaining mass is SH **PROOF 20**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **29→30/38**. Named recovery: dual-entity dated paint plan `conv-49-q53`. Health start year, community-center, August teammates, art-show April, Jon banker, Ned, McGee's, Toronto, Caroline biking, Gina internship, first-console, September activity **held**. Jolene yoga year still MISS (2020 start year is not a stored fact). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (stale pin)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Not re-run this increment.

#### 3. Marketing vertical — lead (stale pin)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

`decide` in a when-event query flooded lexical search, so the dated dual-entity plan never entered top-30. Drop leftover-weak tokens from when/which-year lexical search when another event noun remains. Dual-entity leftover covering beats a solo-painter dated dump. Do not require both people (McGee's names John only). WRITE held 4; do not merge #133.

### Next

**One step:** remaining gold often is not a stored fact (Jolene yoga 2020, Phuket diving, Wolves, Wheel of Time) or is a count dump / invent-Sunday / steal-slot reader. Isolated leftover covering is saturating. Ranking still helps where gold is in the subject corpus but not top-30 (Gina advice, veterans party, Voyageurs). Remaining mass is SH **PROOF 20** + RETRIEVAL 28. Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→112 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-25 — P28 instrument leftover covering + enumerate items (113/180)

**Landed:** product SHA `454fbb3` on `pr/locomo-180-p28-1e9e` (PR #144). Skip-ingest pin [locomo-s0-diag-mh-135-p28-20260825.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p28-20260825.md) (`locomo-s0-diag-mh-135-p28-product-recall-s1-545e82`). P26 is already on dest and main (`9905c79` / `e5ac883`). P27 (`7e3583a`, PR #143) is **not** a pin.

Product change: leftover covering for `what does the X help … with` prefers purpose lines (`uses` / `monitor` / `reminder` / `progress`) over ownership and `help` floods, then copies that covering sentence into enumerate `items`. Does not drop `help` from all instrument-purpose lexical search. Does not classify exercise-feel as instrument-purpose. Does not name LoCoMo.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; last pin. Not re-run this increment. |
| Marketing vertical | **17/17** | Merge gate; last pin. Not re-run this increment. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P26 | **112/180 (0.622)** | MH **18/33** · OD **4/11** · SH **60/98** · temporal **30/38**. SHA `e5ac883`. |
| LoCoMo S0 product hybrid **on** P28 | **113/180 (0.628)** | MH **18/33** · OD **4/11** · SH **61/98** · temporal **30/38**. SHA `454fbb3`. Ledger: **RETRIEVAL 28 / PROOF 25 / READER 8 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** 90%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 113/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. Item flips vs P26: **+1 / −0 = net +1**. P27 180 was **112/180** (+0/−0) because covering rewrote `answer` while the harness returned enumerate `items`.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 113 vs 11, or 113 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **112** P26 → **113** P28 | **no same-n pin** (fair 180 429) | Product 19→113 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P26 **18/33** → P28 **18/33** | no 33-item freeze | Held P21 high. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P28 product local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (18/33).** Held P21 high. Destress pottery **held**. Yoga locations **held**. Childhood items **held**. Still missing: Phuket diving (`conv-48-q77`) — write split. Do not treat 18/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. James girlfriend April 2022 **held**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **60→61/98**. Industry **27/98**. Named recovery: smartwatch reminder. Remaining mass is SH **PROOF 19**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **Held 30/38**. Paint Saturday, health start year, community-center, August teammates, art-show April, Jon banker, Ned, McGee's, Toronto, Caroline biking, Gina internship, first-console, September activity **held**. Jolene yoga year still MISS (2020 start year is not a stored fact). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (stale pin)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Not re-run this increment.

#### 3. Marketing vertical — lead (stale pin)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

Leftover covering already selected the purpose fact (`uses` / `monitor` / `reminder`). The 180 harness classifies `"what does "` as enumerate and prefers `items` over `answer`, so a generic hybrid paraphrase stayed visible. Copy the covering sentence into enumerate items for instrument-purpose queries only. Do not collapse unwind lists. Do not drop `help` from all instrument-purpose lexical search (exercise-feel hang). WRITE held 4; do not merge #133.

### Next

**One step:** remaining gold often is not a stored fact (Jolene yoga 2020, Phuket diving, Wolves, Wheel of Time) or is a count dump / invent-Sunday / steal-slot reader. Isolated leftover covering is saturating. Ranking still helps where gold is in the subject corpus but not top-30 (Gina advice — do not drop `advice` from FTS AND; veterans party; Voyageurs). Remaining mass is SH **PROOF 19** + RETRIEVAL 28. Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→113 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-25 — P30 what-made leftover covering + lexical admit (114/180)

**Landed:** product SHA `dbc33d1` on `pr/locomo-180-p30-1e9e` (PR #146). Skip-ingest pin [locomo-s0-diag-mh-135-p30-20260825.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p30-20260825.md) (`locomo-s0-diag-mh-135-p30-product-recall-s1-1f2c6b`). P28 is already on dest and main (`0f417d8` / `454fbb3`). P27 (`7e3583a`, PR #143) and P29 (`747ab1d`, PR #145) are **not** pins.

Product change: what-made leftover covering prefers off-query evidence (`push` / `remind`) over enjoy/participate restatement. Lexical search drops structure tokens (`made`/`part`), the queried person, and short reason verbs (`stay`/`easy`) when other event tokens remain, so first-person cause lines enter the recall packet. Filtered token-admit and evidence covering are gated to what-made queries so which-year still covers `year`. Does not add `made`/`part` to global leftover-weak. Does not switch FTS to OR-all-terms. Does not name LoCoMo.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; last pin. Not re-run this increment. |
| Marketing vertical | **17/17** | Merge gate; last pin. Not re-run this increment. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P28 | **113/180 (0.628)** | MH **18/33** · OD **4/11** · SH **61/98** · temporal **30/38**. SHA `454fbb3`. |
| LoCoMo S0 product hybrid **on** P30 | **114/180 (0.633)** | MH **18/33** · OD **4/11** · SH **62/98** · temporal **30/38**. SHA `dbc33d1`. Ledger: **RETRIEVAL 28 / PROOF 24 / READER 8 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** 90%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 114/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. Item flips vs P28: **+1 / −0 = net +1**. P29 180 was **111/180** (−2) — do not revive attended-echo ranking.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 114 vs 11, or 114 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **113** P28 → **114** P30 | **no same-n pin** (fair 180 429) | Product 19→114 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P28 **18/33** → P30 **18/33** | no 33-item freeze | Held P21 high. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P30 product local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (18/33).** Held P21 high. Destress pottery **held**. Yoga locations **held**. Childhood items **held**. Still missing: Phuket diving (`conv-48-q77`) — write split. Do not treat 18/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. James girlfriend April 2022 **held**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **61→62/98**. Industry **27/98**. Named recovery: running-group push (`conv-48-q138`). Remaining mass is SH **PROOF 18**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment.

**Temporal.** **Held 30/38**. Paint Saturday, health start year, community-center, August teammates, art-show April, Jon banker, Ned, McGee's, Toronto, Caroline biking, Gina internship **held**. Jolene yoga year still MISS (2020 start year is not a stored fact). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (stale pin)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Not re-run this increment.

#### 3. Marketing vertical — lead (stale pin)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

The gold cause line was stored (`We help and push each other during our runs…`) but never entered the recall packet. FTS `plainto_tsquery` ANDs `made`/`part`/`Deborah`; ILIKE-OR recency overfetch is ~700 rows (gold at ~360 vs cap 120); evidence-set covering then stuffed high-DF `stay` lines into top-30. Dropping what-made structure/person/short-reason tokens shrinks the ILIKE pool to the working `running group motivated` set so the leftover is rank 1 and hybrid can cite it. Do not apply that covering filter to which-year (strips `year`, returns attended-health echoes). WRITE held 4; do not merge #133.

### Next

**One step:** remaining gold often is not a stored fact (Jolene yoga 2020, Phuket diving, Wolves, Wheel of Time) or is a count dump / invent-Sunday / steal-slot reader. Isolated leftover covering is saturating. Ranking still helps where gold is in the subject corpus but not top-30 (Gina advice — do not drop `advice` from FTS AND; veterans party; Voyageurs). Remaining mass is SH **PROOF 18** + RETRIEVAL 28. Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→114 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not revive P29. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-25 — P31 how-describe lexical admit (115/180)

**Landed:** product SHA `98d8931` on `pr/locomo-180-p31-1e9e` (PR #147). Skip-ingest pin [locomo-s0-diag-mh-135-p31-20260825.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p31-20260825.md) (`locomo-s0-diag-mh-135-p31b-product-recall-s1-909042`). P30 is already on dest and main (`0ffbd6f` / `dbc33d1`). First P31 180 (`a6117e`, **114/180**) is **not** a pin (James girlfriend hybrid flake). P27 (`7e3583a`, PR #143) and P29 (`747ab1d`, PR #145) are **not** pins.

Product change: how-does/did/do-X-describe-Y lexical search drops `describe*` so compiler `X describes Y` stamps do not flood, drops capitalized person tokens when other object tokens remain (first-person leftover omits the name), and drops `got` after person filtering so acquisition ILIKE does not crowd the object pool. Filtered token-admit and evidence covering are gated to what-made **or** how-describe so which-year still covers `year`. Destress and instrument-purpose are not how-describe. Does not drop Joanna from all search. Does not name LoCoMo.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | ---: |
| OpMem | **13/13** | Merge gate; last pin. Not re-run this increment. |
| Marketing vertical | **17/17** | Merge gate; last pin. Not re-run this increment. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P30 | **114/180 (0.633)** | MH **18/33** · OD **4/11** · SH **62/98** · temporal **30/38**. SHA `dbc33d1`. |
| LoCoMo S0 product hybrid **on** P31 | **115/180 (0.639)** | MH **18/33** · OD **4/11** · SH **63/98** · temporal **30/38**. SHA `98d8931`. Ledger: **RETRIEVAL 28 / PROOF 23 / READER 8 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** 90%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 115/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. Item flips vs P30: **+1 / −0 = net +1**. First P31 180 was **114/180** (stuffed-animal +1 / girlfriend flake −1) — not a pin.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 115 vs 11, or 115 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **114** P30 → **115** P31 | **no same-n pin** (fair 180 429) | Product 19→115 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P30 **18/33** → P31 **18/33** | no 33-item freeze | Held P21 high. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P31 product local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (18/33).** Held P21 high. Destress pottery **held**. Yoga locations **held**. Childhood items **held**. Still missing: Phuket diving (`conv-48-q77`) — write split. Do not treat 18/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. James girlfriend April 2022 **held** on the pin run (first P31 180 hybrid-abstained; live `/recall` still answers no). Do not restore remaining OD by stuffing episodes.

**Single-hop.** **62→63/98**. Industry **27/98**. Named recovery: stuffed-animal good vibes (`conv-42-q123`). Remaining mass is SH **PROOF 17**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment. Turtle-care and camping-peaceful still miss (no turtle/care tokens; peaceful not stored).

**Temporal.** **Held 30/38**. Paint Saturday, health start year, community-center, August teammates, art-show April, Jon banker, Ned, McGee's, Toronto, Caroline biking, Gina internship **held**. Jolene yoga year still MISS (2020 start year is not a stored fact). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (stale pin)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Not re-run this increment.

#### 3. Marketing vertical — lead (stale pin)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

The gold leftover (`It's a stuffed animal to remind you of the good vibes`) is stored and active, recency 12 in the 22-hit stuffed|animal pool, but live `/recall` never admitted it. Compiler `Nate describes his gaming space` plus Joanna/Nate person ILIKE flooded the pool; keeping `got` then grew the ILIKE-OR set to 76 so covering/cap dropped gold. Dropping how-describe structure/person/`got` shrinks the pool to stuffed+animal so gold is in top-k and hybrid cites good vibes. Do not apply that covering filter to which-year. WRITE held 4; do not merge #133.

### Next

**One step:** remaining gold often is not a stored fact (Jolene yoga 2020, Phuket diving, Wolves, Wheel of Time, camping peaceful) or is a count dump / invent-Sunday / steal-slot reader. Isolated leftover covering is saturating. Ranking still helps where gold is in the subject corpus but not top-30 (Gina advice — do not drop `advice` from FTS AND; veterans party; Voyageurs; Joanna writing). Remaining mass is SH **PROOF 17** + RETRIEVAL 28. Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→115 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not revive P29. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

## 2026-08-25 — P32 host leftover covering + session admit (116/180)

**Landed:** product SHA `cd77a74` on `pr/locomo-180-p32-1e9e` (PR #148). Skip-ingest pin [locomo-s0-diag-mh-135-p32-20260825.md](../../benchmarks/artifacts/locomo-s0-diag-mh-135-p32-20260825.md) (`locomo-s0-diag-mh-135-p32b-product-recall-s1-ecb9c9`). P31 is already on dest and main (`86d2a65` / `98d8931`). First P32 180 (`7fa994`, **115/180**) is **not** a pin (party stamp without share-stories). P27 (`7e3583a`, PR #143) and P29 (`747ab1d`, PR #145) are **not** pins.

Product change: what-did/does/has-X-host leftover covering prefers leftover hosted-event lines (` party ` / ` dinner ` / ` gathering ` / ` celebration ` / ` reception `) over realize speech and joins up to two hosted-event covering lines. Search admits leftover hosted-event neighbors from leftover-covering candidate sessions via a bounded session-id fetch (not FTS head, not a raised global neighbor cap), then ranks those lines into host-query top-k. Host queries keep leftover hosted-event episode primitives even when fact-primary coverage looks complete. Destress and what-made are not host queries. Does not add a host→party dictionary. Does not drop May. Does not name LoCoMo.

### Own pins

| Suite | Brainy | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate; last pin. Not re-run this increment. |
| Marketing vertical | **17/17** | Merge gate; last pin. Not re-run this increment. |
| LoCoMo S0 product `/recall` this VM **reader off** | **19/180 (0.106)** | MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. SHA `453a929`. |
| LoCoMo S0 product hybrid **on** P31 | **115/180 (0.639)** | MH **18/33** · OD **4/11** · SH **63/98** · temporal **30/38**. SHA `98d8931`. |
| LoCoMo S0 product hybrid **on** P32 | **116/180 (0.644)** | MH **18/33** · OD **4/11** · SH **64/98** · temporal **30/38**. SHA `cd77a74`. Ledger: **RETRIEVAL 28 / PROOF 22 / READER 8 / WRITE 4 / HARNESS 2**. |
| LoCoMo S0 industry search+harness this VM | **62/180 (0.344)** | Unchanged vs reader-off pin. |
| LoCoMo S0 product integrity VM | **32/180** | Different tenant. **Do not mix.** |
| 1×30 conv-26 | **21/30** | Diagnostic; not overwritten. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old SHA `1b5ab3e`. Not re-run. |
| LME-20 / BEAM | **not re-run** | |

This is **not** 80%, **not** 90%, **not** n=1540, **not** a Mem0 same-pin, and **not** SOTA. 116/180 does not replace integrity 32/180 or the no-LLM 19/180 pin. Item flips vs P31: **+1 / −0 = net +1**. First P32 180 was **115/180** (party stamp without share-stories) — not a pin.

### Competitor compare (detailed)

No new Mem0 / Graphiti / Zep **score** this cycle. Fair Mem0 Platform 180 (`locomo-s0-mem0-v3-s1-fair2`) died on **HTTP 429 usage quota** (SEARCH 1000/1000, reset **2026-09-01**). The 2026-08-15 Mem0 1×30 freeze remains **11/30** and **handicapped** — do not refresh lead/trail from 21 vs 11, 116 vs 11, or 116 vs unpublished Platform 180.

#### 1. LoCoMo conversational QA

| Axis | This cycle | Mem0 Platform | Stand |
| --- | ---: | --- | --- |
| 1×30 overall | **not re-run** (prior Brainy **21/30**) | freeze **11/30**, protocol handicapped | Do **not** refresh lead/trail. |
| S0 n=180 product this VM | **19** off → **115** P31 → **116** P32 | **no same-n pin** (fair 180 429) | Product 19→116 vs itself. Leads this-VM industry **62/180** on the **product** lane. Not a Mem0 same-pin. |
| S0 n=180 industry this VM | **62/180** | **no same-n pin** | Same-pin lane vs Mem0 after quota reset. |
| S0 MH product (this tenant) | reader-off **12/33** → P31 **18/33** → P32 **18/33** | no 33-item freeze | Held P21 high. Product MH still leads this-VM industry MH **10/33**. |
| Search p50 | P32 product local (search; hybrid recall is separate) | freeze 492 ms platform | Harness observation, not a SLO |

**Multi-hop (18/33).** Held P21 high. Destress pottery **held**. Yoga locations **held**. Childhood items **held**. Still missing: Phuket diving (`conv-48-q77`) — write split. Do not treat 18/33 as n=1540 MH.

**Open-domain.** **Held 4/11**. Industry **3/11**. James girlfriend April 2022 **held**. Do not restore remaining OD by stuffing episodes.

**Single-hop.** **63→64/98**. Industry **27/98**. Named recovery: veterans party + share-stories (`conv-41-q98`). Remaining mass is SH **PROOF 16**. Write-missing golds (Wolves, Wheel of Time, Monster Hunter) are not this increment. Turtle-care and camping-peaceful still miss (no turtle/care tokens; peaceful not stored). Gina advice still miss (gold has no advice/business tokens; do not drop `advice`).

**Temporal.** **Held 30/38**. Paint Saturday, health start year, community-center, August teammates, art-show April, Jon banker, Ned, McGee's, Toronto, Caroline biking, Gina internship **held**. Jolene yoga year still MISS (2020 start year is not a stored fact). Do not add LoCoMo-named date rules.

**Published Mem0 92.5%** stays context, never a scoreboard row.

#### 2. OpMem — lead (stale pin)

Last **13/13**. Last Mem0 Platform ops pin **10/13**. **Lead ops.** Not re-run this increment.

#### 3. Marketing vertical — lead (stale pin)

Last **17/17** vs Mem0 empirical **4/17**. **Lead governed vertical.**

#### 4. LME-20 — no pin this cycle

Last Brainy pin **4/20**. Not re-run.

#### 5. Graphiti / Zep — no pin

**No same-pin.**

**Mem0 OSS** was not re-measured. Platform fair 180 is **quota-blocked** until 2026-09-01.

### Why

The gold leftover (`John organized a small party for veterans` plus `We had a great time throwing a small party and inviting some veterans to share their stories`) is stored and active in session_15, but live `/recall` never admitted it. Generic neighbor cap 16 walks store order (session_15 has 69 rows; party is neighbor ~50). `ListMemoriesLimited` is a 400-row recency window (party list rank 1224 / 2571). FTS overfetch never includes that session; realize/photograph enter later via dense/related and date-token ranking buries the undated party past default top-k 30. Host leftover covering then cannot prefer a line that is not in the packet. Fetch bounded rows by leftover-covering candidate `session_id`s (not FTS head), boost leftover hosted-event lines on host queries, and keep leftover hosted-event episode primitives so share-stories joins the party stamp. Do not add `party` as a query token. WRITE held 4; do not merge #133.

### Next

**One step:** ranking so gold enters the leftover packet without dropping FTS structure tokens (Gina advice — do not drop `advice`; turtle-care; Joanna writing). Remaining gold often is not a stored fact (Jolene yoga 2020, Phuket diving, Wolves, Wheel of Time, camping peaceful) or is a count dump / invent-Sunday / steal-slot reader. Isolated leftover covering is saturating. Remaining mass is SH **PROOF 16** + RETRIEVAL 28. Fair Mem0 180 waits on quota reset 2026-09-01. n=1540 only at S6 — 19→116 is a stratified delta but not permission to burn full LoCoMo yet. Do not merge #133. Do not revive P29. Do not add a host→party dictionary. Do not write SOTA. Kill list unchanged. Start: [handover-sota-agent-2026-08-21.md](../handover-sota-agent-2026-08-21.md).

