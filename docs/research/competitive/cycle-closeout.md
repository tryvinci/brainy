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

