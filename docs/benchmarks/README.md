# Benchmarks

How Brainy is measured, and how that compares to other memory systems.

The README carries two tables: **published %** (industry scoreboard) and
**same-pin** (the only fair lead/trail). This page is the full section:
sourced claims, caveats, artifacts, and reproduce commands. Internal cycle
write-ups live in
[research/competitive/cycle-closeout.md](../research/competitive/cycle-closeout.md).

**Honest rules**

- We do **not** invent scores. Suites without a run are **not run** / **no pin**.
- Same-pin = same dataset SHA, same judge/answerer, same question set.
- Published vendor percents live in [published-claims.md](./published-claims.md). Label metric and n. Do not treat them as same-pin lead/trail.
- 1×30 LoCoMo is **measurement, not qualification**. Not SOTA.
- **Mem0 OSS ≠ Mem0 Platform.** **Graphiti OSS ≠ Zep Platform.**

## Published %

Industry format: one percent per public suite. **Not same-pin.**

| | LoCoMo | LongMemEval | BEAM 1M | BEAM 10M |
| --- | ---: | ---: | ---: | ---: |
| **Brainy** | **11.4%** full `/recall` (n=1540) · **70.0%** 1×30 | **20%** LME-20 `/recall` (4/20) · **4%** (n=100 hist.) | not run (40% on 100K/20q this cycle) | not run |
| **Mem0 Platform** | **92.5%** | **94.4%** | **64.1** | **48.6** |
| **Zep** | **75.14%** | 71.2% | — | — |
| **SuperMemory** | 77.1% | 95% Recall@15 | — | — |
| **Letta** | 74.0% | — | — | — |
| **Hindsight** | 92.0% | 94.6% | 73.9% | 64.1% |

Sources, n, metric type, and the Zep LoCoMo dispute:
[published-claims.md](./published-claims.md). Peer harness:
[mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks).

## Same-pin comparison

Pin date for Brainy: **fresh remasure, 2026-08-15** (`1b5ab3e` product SHA).
Mem0 LoCoMo / OpMem / marketing freeze: **this cycle**. Full LoCoMo n=1540
landed at **11.4%** product `/recall`. LME-20 landed at **4/20**. BEAM 100K
landed at **8/20**. LME-500 and BEAM 1M/10M not run.

| Suite | Brainy | Mem0 Platform | Graphiti OSS / Zep Platform | Stand |
| --- | ---: | ---: | ---: | --- |
| LoCoMo 1×30 overall | **21/30 (70.0%)** | **11/30 (36.7%)** | **no same-pin** | Lead this freeze; **not** SOTA |
| Multi-hop | **10/10** | **6/10** | no same-pin | Lead this freeze |
| Open-domain | **0/4** | **3/4** | no same-pin | **Trail** |
| Temporal | **11/16** | **2/16** | no same-pin | Lead this freeze |
| OpMem | **13/13** | **10/13** | no pin | Lead ops |
| Marketing vertical | **17/17** | **4/17** empirical | no pin | Lead governed vertical |
| LongMemEval-20 | **4/20 (20.0%)** product `/recall` | no fair pin on this harness | no pin | Lift vs 0/20 integrity; **not** vs published 94.4% |
| BEAM 100K | **8/20 (40.0%)** search+harness | published elsewhere; not our pin | no pin | Non-reg vs hist. 8/20; 1M/10M not run |

Search p50 on the LoCoMo 1×30 pin: Brainy **175 ms** local vs Mem0 Platform
**492 ms**. That is a harness observation, **not** a production SLO.

Mem0 OSS was **not** re-measured. Do not treat Platform 11/30 as an
OSS-reproducible number.

The 2026-08-15 Mem0 1×30 freeze used v2 search, v1 add, chunk 8, no session
timestamps, and top_k 30. That is **not** their published LoCoMo protocol
(v3 search/add, chunk 1, unix `timestamp`, top_k 200). Do **not** write a
new lead/trail sentence from 21/30 vs 11/30 until the fair stratified 180
lands. Audit: [mem0-harness-audit-2026-08-22.md](../research/competitive/mem0-harness-audit-2026-08-22.md).

Do not paste the published-% table into this same-pin n/N table. Those
percents (and the Zep 58–84 dispute) are catalogued in
[published-claims.md](./published-claims.md).

## Public suites

The three suites in [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks)
are first-class here. We run our own adapters (`UnifiedResult` JSON); we do not
vendor their repo.

| Benchmark | What it tests | Upstream | Our runner |
| --- | --- | --- | --- |
| **LOCOMO** | Long multi-session dialogues; factual / multi-hop / temporal QA | [snap-research/locomo](https://github.com/snap-research/locomo) · [ACL 2024](https://aclanthology.org/2024.acl-long.747/) | `evals/public/locomo/` (`--system brainy` or `mem0`) |
| **LongMemEval** | Long-term extraction, temporal, multi-session | LongMemEval dataset | `evals/public/longmemeval/` — this cycle **4/20** on `/recall` ([pin](./artifacts/lme20-fresh-20260815.md)); LME-500 not run |
| **BEAM** | Retrieval across 100K–10M token chats | HuggingFace `Mohammadta/BEAM` | `evals/public/beam/` — this cycle **8/20** on 100K conv-0 ([pin](./artifacts/beam-100k-fresh-20260815.md)); 1M/10M not run |
| Harness peer | Ingest → search → LLM answer/judge; `UnifiedResult` JSON | **[mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks)** | `evals/public/schema.py` · Brainy drop-in: `evals/public/backends/memory_benchmarks_brainy.py` |

Also outlinked (not in the comparison table): [LongMemBench](https://supermemory.ai/research/longmembench/).

Our LoCoMo smoke defaults to **one conversation × 30 questions**, categories
1–4 (adversarial excluded), judge temperature **0**. Dataset SHA for the frozen
compare: `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`.

## Own suites

Reproducible in this repo against a live Brainy API (stdlib Python, no `pip`).

| Suite | Spec / fixtures | Runner |
| --- | --- | --- |
| OpMem | [opmem-spec](../research/opmem-spec.md) · `fixtures/opmem/` | `evals/run_opmem.py` |
| Parity | `fixtures/parity/` | `evals/run_eval.py` |
| Marketing vertical | `fixtures/vertical/marketing/` | `evals/run_vertical_eval.py` |
| Marketing MVP / moat | [METHODOLOGY](./METHODOLOGY.md) | `evals/run_marketing_mvp_benchmark.py` |

CI (`go test ./...`) runs the HTTP harnesses. Same-pin vs Mem0 Platform needs
`MEM0_API_KEY` and is **not** a merge gate.

## Reproduce

Start the API (`docker compose up --build` or `go run ./cmd/api`), then:

```bash
export BRAINY_BASE_URL=http://localhost:8080

python3 evals/run_opmem.py --systems brainy,verbatim --base-url "$BRAINY_BASE_URL"
python3 evals/run_vertical_eval.py --base-url "$BRAINY_BASE_URL"
python3 evals/run_marketing_mvp_benchmark.py --base-url "$BRAINY_BASE_URL"

# Mem0 Platform counter-run (optional)
# python3 evals/run_opmem.py --systems verbatim,brainy,mem0 --base-url "$BRAINY_BASE_URL"
# python3 evals/run_marketing_mvp_benchmark.py --base-url "$BRAINY_BASE_URL" --systems brainy,mem0
# python3 evals/run_competitor_benchmark.py --brainy-url "$BRAINY_BASE_URL"
```

Public suites (LoCoMo 1×30 Brainy, then Mem0 Platform; LME / BEAM optional):

```bash
cd evals
python -m public.locomo.run_smoke --system brainy --conversations 1 --questions 30
# python -m public.locomo.run_smoke --system mem0 --conversations 1 --questions 30
# python -m public.longmemeval.run --limit 20 --product-recall
# python -m public.beam.run --chat-size 100K --conversations 0-0
```

More harness detail: [evals/README.md](../../evals/README.md) ·
[evals/public/README.md](../../evals/public/README.md).
Ladder: [research/public-bench-ladder.md](../research/public-bench-ladder.md).

## Current Brainy artifacts (fresh remasure 2026-08-15)

| Suite | Report |
| --- | --- |
| LoCoMo 1×30 | [locomo-fresh-1x30-20260815.md](./artifacts/locomo-fresh-1x30-20260815.md) |
| LoCoMo full n=1540 | [locomo-fresh-full-20260815.md](./artifacts/locomo-fresh-full-20260815.md) · [why 11.4%](./artifacts/locomo-full-recall-dip-why-20260817.md) |
| LongMemEval-20 | [lme20-fresh-20260815.md](./artifacts/lme20-fresh-20260815.md) |
| Mem0 LoCoMo freeze | [locomo-mem0-fresh-1x30-20260815.md](./artifacts/locomo-mem0-fresh-1x30-20260815.md) |
| OpMem | [opmem-fresh-local-20260815.md](./artifacts/opmem-fresh-local-20260815.md) · [Mem0](./artifacts/opmem-mem0-fresh-20260815.md) |
| Marketing | [marketing-fresh-local-20260815.md](./artifacts/marketing-fresh-local-20260815.md) · [Mem0](./artifacts/marketing-mem0-fresh-20260815.md) · [moat](./marketing-moat-report.md) |
| Cycle closeout (detailed why) | [cycle-closeout.md](../research/competitive/cycle-closeout.md) |
| Fail-closed integrity remasure (2026-08-19/20) | [S0 dual-lane](./artifacts/locomo-integrity-s0-20260819.md) · [3×90](./artifacts/locomo-integrity-3x90-20260820.md) · [extraction ceiling](./artifacts/extraction-ceiling-20260819.md) · [embedding A/B](./artifacts/embedding-ab-20260819.md). Invalidates Aug-19 17/180 / 52/180. **Not** a replacement for the 1×30 freeze above. |
| Current-SHA S0 this-VM (2026-08-22) | Tenant `diag-mh-135` + conv-30, hybrid **off**: product **19/180**, industry **62/180**. [pin](./artifacts/locomo-s0-diag-mh-135-20260822.md). **Not** a replacement for integrity 32/180 or the 1×30 freeze. |
| P1 hybrid-on S0 this-VM (2026-08-22) | Same store, `BRAINY_RECALL_LLM=1`: product **37/180**. MH **10/33 dip** vs reader-off 12/33. SH 19/98, temporal 7/38. [pin](./artifacts/locomo-s0-diag-mh-135-llm-20260822.md). **Not** n=1540, **not** a Mem0 same-pin. |
| P2-narrow S0 this-VM (2026-08-22) | Same store, hybrid on. Length-lock **56/180** (MH 11/33, temporal **21/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p2-20260822.md). Extras-lock + skip hop-ground **61/180** (MH **16/33**, temporal **19/38 dip** vs 21/38). [pin](./artifacts/locomo-s0-diag-mh-135-p2b-20260822.md). **Not** n=1540, **not** a Mem0 same-pin. |
| P3 distinctive-token S0 this-VM (2026-08-22) | Same store, hybrid on. Admit leftover query tokens; skip unproven hop dumps. **73/180** (MH **16/33** held, OD **3/11**, SH **32/98**, temporal **22/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p3-20260822.md). **Not** n=1540, **not** a Mem0 same-pin. |
| P4 identity/garbage S0 this-VM (2026-08-22) | Same store, hybrid on. Reject garbage hybrid; skip identity-only hop dumps; do not hop-ground those dumps. **79/180** (MH **16/33** held, OD **3/11**, SH **37/98**, temporal **23/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p4-20260822.md). **Not** n=1540, **not** a Mem0 same-pin. |
| P5 activity-dump S0 this-VM (2026-08-22) | Same store, hybrid on. Skip activity/event hop dumps that miss leftover query tokens. **84/180** (MH **17/33**, OD **2/11 dip**, SH **45/98**, temporal **20/38 dip**). [pin](./artifacts/locomo-s0-diag-mh-135-p5-20260822.md). **Not** n=1540, **not** a Mem0 same-pin. |
| P6 dump-lock S0 this-VM (2026-08-22) | Same store, hybrid on. Skip dual-entity activity dumps; unlock hop-dump locks; place-fact ranking. **87/180** (MH **13/33 dip**, OD **3/11**, SH **52/98**, temporal **19/38 dip**). [pin](./artifacts/locomo-s0-diag-mh-135-p6-20260822.md). **Not** n=1540, **not** a Mem0 same-pin. |
| P7 hop-local joins S0 this-VM (2026-08-22) | Same store, hybrid on. Leftover hop contents under skip; rare-share omitted possessions; keep typed joins on hybrid abstain. **88/180** (MH **14/33**, OD **4/11**, SH **49/98 dip**, temporal **21/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p7-20260822.md). **Not** n=1540, **not** a Mem0 same-pin. |
| P8 SH recovery S0 this-VM (2026-08-22) | Same store, hybrid on. Where-only mh_list unlock; dated ordinal names; attended-event hop drop; leftover specific facts with conflicting date tails stripped. **93/180** (MH **15/33**, OD **4/11**, SH **52/98**, temporal **22/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p8-20260822.md). **Not** n=1540, **not** a Mem0 same-pin. |
| P9 unproven mh_list dumps S0 this-VM (2026-08-22) | Same store, hybrid on. Unlock unproven search_fallback mh_list dumps; skip OCR captions and question-echo hop values. **94/180** (MH **15/33**, OD **4/11**, SH **53/98**, temporal **22/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p9-20260822.md). **Not** n=1540, **not** a Mem0 same-pin. |
| P10 date-aware leftover covering S0 this-VM (2026-08-23) | Same store, hybrid on. Far-dated leftover covering skip (10-day window); hybrid 48h window except where-queries; speaker-prefix leftover covering uses the body. **96/180** (MH **15/33**, OD **4/11**, SH **55/98**, temporal **22/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p10-20260823.md). **Not** n=1540, **not** a Mem0 same-pin. |
| P11 locative leftover covering S0 this-VM (2026-08-23) | Same store, hybrid on. Where leftover covering ignores hop slots; strong leftover tokens; hybrid leftover overwrite only on where/games-played joins. **97/180** (MH **13/33 dip**, OD **3/11 dip**, SH **58/98**, temporal **23/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p11-20260823.md). **Not** n=1540, **not** a Mem0 same-pin. |
| P12 keep short where NPs / typed item joins S0 this-VM (2026-08-23) | Same store, hybrid on. Where covering returns the locative place NP; short typed item joins stay locked; leftover `support` is a weak token. **101/180** (MH **15/33**, OD **4/11**, SH **58/98**, temporal **24/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p12-20260823.md). **Not** n=1540, **not** a Mem0 same-pin. |
| P13 gated leftover-vs-hybrid S0 this-VM (2026-08-23) | Same store, hybrid on. Leftover covering may replace hybrid only for schema-activity covering (enjoys/likes/loves/participates in) or where/games-played; chat turns do not get a locative leftover bonus. **102/180** (MH **15/33**, OD **4/11**, SH **59/98**, temporal **24/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p13-20260823.md). **Not** n=1540, **not** a Mem0 same-pin. |
| P14 childhood possession lock S0 this-VM (2026-08-23) | Same store, hybrid on. 2-item childhood possession lists lock against speaker chat; leftover covering can join `had a` + age-cue facts; name questions are not childhood item lists. **103/180** (MH **16/33**, OD **4/11**, SH **59/98**, temporal **24/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p14-20260823.md). Broad typed-join 180 **98/180** is not a pin. **Not** n=1540, **not** a Mem0 same-pin. |
| P17 when-event leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. When-event leftover covering keeps 4-character event verbs; a short calendar date that misses those tokens yields to a packet covering line that has them. **104/180** (MH **16/33**, OD **4/11**, SH **60/98**, temporal **24/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p17-20260825.md). P15/P16 **103/180** net-zero 180s are not pins. Named dip: Jon banker job. **Not** n=1540, **not** a Mem0 same-pin. |
| P18 when-event query-entity bind S0 this-VM (2026-08-25) | Same store, hybrid on. When-event leftover covering skips lines that name another person and do not name a query person; first-person dated lines still compete. **105/180** (MH **16/33**, OD **4/11**, SH **60/98**, temporal **25/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p18-20260825.md). Jon banker recovered. **Not** n=1540, **not** a Mem0 same-pin. |
| P20 enumerate unwind extras S0 this-VM (2026-08-25) | Same store, hybrid on. Unwind-evidenced packet/hop activity slots join onto enumerate `items` (not only `answer`). **106/180** (MH **17/33**, OD **4/11**, SH **60/98**, temporal **25/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p20-20260825.md). Destress pottery recovered. P19/P19b **105/180** holds are not pins. **Not** n=1540, **not** a Mem0 same-pin. |
| P21 location-list leftover lock S0 this-VM (2026-08-25) | Same store, hybrid on. Location-list leftover covering requires a locative; leftover packet practice places join onto answer/items; which/what-year covering binds to the query person and needs a year token. **107/180** (MH **18/33**, OD **4/11**, SH **60/98**, temporal **25/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p21-20260825.md). Yoga practice locations recovered. First P21 180 **106/180** flake is not a pin. **Not** n=1540, **not** a Mem0 same-pin. |
| P22 leftover month bind S0 this-VM (2026-08-25) | Same store, hybrid on. When leftover covering tokens collapse to weak-only, bind covering to the query month/year; rarest-token override ignores weak tokens. **108/180** (MH **18/33**, OD **4/11**, SH **60/98**, temporal **26/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p22-20260825.md). September co-participant plan recovered. **Not** n=1540, **not** a Mem0 same-pin. |
| P23 year/month leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. When-event leftover covering requires a year/date token so year-only event facts can replace a bare hop date; covering from a different query year or month is skipped. **109/180** (MH **18/33**, OD **4/11**, SH **60/98**, temporal **27/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p23-20260825.md). Hometown community-center 2022 recovered. First P23 180 **108/180** hold is not a pin. **Not** n=1540, **not** a Mem0 same-pin. |
| P24 sentence-initial verb covering S0 this-VM (2026-08-25) | Same store, hybrid on. Sentence-initial past-tense verbs and phrasal particles are not leftover covering people, so unnamed dated diary lines still compete. **110/180** (MH **18/33**, OD **4/11**, SH **60/98**, temporal **28/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p24-20260825.md). Unnamed dated art-show diary recovered. **Not** n=1540, **not** a Mem0 same-pin. |
| P25 which-year as-of duration S0 this-VM (2026-08-25) | Same store, hybrid on. Which-year leftover covering of the form `for N years as of DATE` rewrites to the start year; leftover tokens `year`/`years` are weak. **111/180** (MH **18/33**, OD **4/11**, SH **60/98**, temporal **29/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p25-20260825.md). Health start year recovered. **Not** n=1540, **not** a Mem0 same-pin. |
| P26 when-event speech-act flood S0 this-VM (2026-08-25) | Same store, hybrid on. When/which-year lexical search drops leftover-weak tokens (`decide`/`year`) when another event noun remains; leftover covering prefers lines that name both query people. **112/180** (MH **18/33**, OD **4/11**, SH **60/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p26-20260825.md). Dual-entity dated paint plan recovered. **Not** n=1540, **not** a Mem0 same-pin. |
| P28 instrument leftover covering + enumerate items S0 this-VM (2026-08-25) | Same store, hybrid on. Instrument-purpose leftover covering prefers purpose lines over ownership/`help` floods; enumerate `items` copy that covering sentence. **113/180** (MH **18/33**, OD **4/11**, SH **61/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p28-20260825.md). Smartwatch reminder recovered. P27 112/180 hold is not a pin. **Not** n=1540, **not** a Mem0 same-pin. |
| P30 what-made leftover covering + lexical admit S0 this-VM (2026-08-25) | Same store, hybrid on. What-made leftover covering prefers off-query evidence over enjoy/participate restatement; lexical search drops `made`/`part`, the queried person, and short reason verbs so first-person cause lines enter the packet. **114/180** (MH **18/33**, OD **4/11**, SH **62/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p30-20260825.md). Running-group push recovered. P29 111/180 dip is not a pin. **Not** n=1540, **not** a Mem0 same-pin. |
| P31 how-describe lexical admit S0 this-VM (2026-08-25) | Same store, hybrid on. How-does/did/do-X-describe-Y lexical search drops `describe*`, capitalized person tokens when other object tokens remain, and `got` after person filtering so first-person leftover lines enter the packet. **115/180** (MH **18/33**, OD **4/11**, SH **63/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p31-20260825.md). Stuffed-animal good vibes recovered. First P31 180 114/180 flake is not a pin. **Not** n=1540, **not** a Mem0 same-pin. |
| P32 host leftover covering + session admit S0 this-VM (2026-08-25) | Same store, hybrid on. What-did/does/has-X-host leftover covering prefers hosted-event lines; bounded session-id fetch for leftover-covering candidate sessions; hosted-event rank boost; host queries keep leftover hosted-event episodes and join two covering lines. **116/180** (MH **18/33**, OD **4/11**, SH **64/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p32-20260825.md). Veterans party + share-stories recovered. First P32 180 115/180 hold is not a pin. **Not** n=1540, **not** a Mem0 same-pin. |
| P33 advice leftover covering + session admit S0 this-VM (2026-08-25) | Same store, hybrid on. What-advice leftover covering prefers hortative / first-person-gerund directive lines; bounded session-id fetch for advice-echo sessions; search floor for leftover with no query tokens; advice queries keep leftover hortative episodes and join up to three directive lines. **117/180** (MH **18/33**, OD **4/11**, SH **65/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p33-20260825.md). Gina advice recovered. **Not** n=1540, **not** a Mem0 same-pin. |
| P34 what-kind like-list leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. What-kind leftover covering prefers like-A,-B,-and-C leftover; bounded session-id fetch ignoring spread/kind restatement tokens; search floor for leftover with no query tokens; what-kind queries keep leftover like-list episodes; crowded hop-dump skip excepts kind-list leftover **only on what-kind queries**. **118/180** (MH **18/33**, OD **4/11**, SH **66/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p34-20260825.md). Dinner-spread like-list recovered. First P34 180 116/180 dip is not a pin. **Not** n=1540, **not** a Mem0 same-pin. |
| P35 how-describe-process prefix hortative leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. How-describe-process leftover covering prefers prefix hortative leftover (`just keep`) over companion slogans; sentence-initial `just` is hortative; process covering empty unless hortative and not process-restatement; hortative leftover sessions seed neighbors first. **119/180** (MH **18/33**, OD **4/11**, SH **67/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p35-20260825.md). Turtle-care recovered. Does not drop `turtles`/`care`. Does not steal Calvin electronic. **Not** n=1540, **not** a Mem0 same-pin. |
| P36 what-motivates first-person object-cause leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. What-motivates leftover covering prefers first-person object-cause leftover over turtle/have-faith/occupation companions; lexical search drops `motivate*`/`keep`/`even` only on `what motivates` queries; skip hops and hybrid when the search packet already has a cause leftover. **120/180** (MH **18/33**, OD **4/11**, SH **68/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p36-20260825.md). Joanna writing recovered. Does not drop `motivate` globally. Does not treat stay-motivated as what-motivates. **Not** n=1540, **not** a Mem0 same-pin. |
| P37 what-say-about they-evaluative leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. What-say-about leftover covering prefers short they-evaluative leftover over dance-photo flood; lexical search drops `say*`/`about` only on `what does/did/do … say about` queries; leftover-covering session fetch is 200 rows; enumerate `items` copy that covering sentence. **121/180** (MH **18/33**, OD **4/11**, SH **69/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p37-20260825.md). Dancers graceful recovered. First P37 180 120/180 hold is not a pin. Does not drop `say` globally. Does not add a dance dictionary. **Not** n=1540, **not** a Mem0 same-pin. |
| P38 what-say-about first-person got leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. What-say-about leftover covering admits first-person possessive/abundance leftover (`it's got` + quantifier) when another packet line covers the object tokens; they-evaluative leftovers still win; lexical search drops question-frame participles and treats all-caps tokens as non-people. **122/180** (MH **18/33**, OD **4/11**, SH **70/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p38-20260825.md). NYC `It's got so much to check out` recovered. Does not treat `it's gotta` as `it's got`. Does not steal Tim injury doctor-said. Does not add an NYC dictionary. **Not** n=1540, **not** a Mem0 same-pin. |
| P39 dated reported-speech leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. What-say-about leftover covering admits dated reported-speech leftover (`the <role> said` + copula) when the query names a calendar date; they-evaluative leftovers still win; undated queries still prefer first-person got leftover. **123/180** (MH **18/33**, OD **4/11**, SH **71/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p39-20260825.md). Tim injury `The doctor said it's not too serious` recovered. Does not add a doctor/injury dictionary. Does not reuse they-copula. Does not steal NYC `It's got`. **Not** n=1540, **not** a Mem0 same-pin. |
| P40 how-react leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. How-react leftover covering admits object-linked they-were observation leftover on `how do/does/did … react/respond to`; covering requires the observation line itself to name a query object; session expand seeds from FTS hits; list evidence-set cap re-inserts those observations. **124/180** (MH **18/33**, OD **4/11**, SH **72/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p40-20260825.md). Dogs/snow `they were confused` recovered. Does not drop `react` globally. Does not add a dog/snow dictionary. **Not** n=1540, **not** a Mem0 same-pin. |
| P41 what-did-purpose leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. What-did-purpose leftover covering admits adjacent purpose-pair action leftover on `what did/does/has … do … to …`; covering requires first-person or named query actor; lexical search drops month/year/comparatives only on that query shape. **125/180** (MH **18/33**, OD **4/11**, SH **73/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p41-20260825.md). Dog-owners group recovered. Does not add a dog/group dictionary. Does not drop November globally. **Not** n=1540, **not** a Mem0 same-pin. |
| P43 how-long-been leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. How-long-been leftover covering admits continuing-years leftover on `how long` + `been`; covering prefers `duration is N years` / `N years already` / `for N years` over copula status; lexical search drops `long` only on that query shape; session rank prefers duration leftover sessions over recency chatter. **127/180** (MH **18/33**, OD **4/11**, SH **75/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p43-20260825.md). Married 5-year duration recovered. Does not add a marriage dictionary. Does not drop `long` globally. **Not** n=1540, **not** a Mem0 same-pin. |
| P44 how-often leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. How-often leftover covering admits cadence leftover on `how often` + `does`/`did`/`do`; covering requires ≥2 remaining object tokens so road-trip months lose to meetup `once a week`; lexical search drops `often` and trailing `for` adjuncts only on that query shape; session rank prefers cadence leftover sessions over recency park-playdates. **128/180** (MH **18/33**, OD **4/11**, SH **76/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p44-20260825.md). Meetup once-a-week recovered. Does not add a how-often/playdate dictionary. Does not drop `often` globally. **Not** n=1540, **not** a Mem0 same-pin. |
| P45 currently-working leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. What-project leftover covering admits currently-working leftover on `what` + `project` + `working`/`work`; covering prefers `currently working` / `working on a new` over childhood desire or creating-own restatement; lexical search drops trailing `in … course` adjuncts and `project` only on that query shape; session rank prefers currently-working leftover sessions over recency comic-sketch chatter. **129/180** (MH **18/33**, OD **4/11**, SH **77/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p45-20260825.md). Football-simulator currently-working leftover recovered. Does not add a football/comic-sketch dictionary. Does not drop `game` globally. **Not** n=1540, **not** a Mem0 same-pin. |
| P46 become-interested leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. Dated what-new-hobby leftover covering admits first-person become-interested leftover on `what` + `hobby` + `interested`/`become` plus a calendar date; covering prefers `become interested` over a foreign-person `new hobby` restatement; lexical search drops `hobby` only on that query shape; session rank prefers become-interested leftover sessions over recency metal-detecting chatter. **130/180** (MH **18/33**, OD **4/11**, SH **78/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p46-20260825.md). James extreme-sports leftover recovered. Does not add an extreme-sports dictionary. Does not steal John's metal detecting. **Not** n=1540, **not** a Mem0 same-pin. |
| P47 how-plan-dream leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. How-plan-dream leftover covering admits first-person gathering/watching/guide leftover on `how does/did/do/will` + plan/pursue + dream + learn; covering prefers `gathering information` / `watching videos` / `beginners' guide` over a foreign-person `learning their stories` restatement; lexical search drops `plan`/`pursue`/`dream`/`learning` only on that query shape; session rank prefers prep-plan leftover sessions over recency exploring chatter. **131/180** (MH **18/33**, OD **4/11**, SH **79/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p47-20260825.md). Jolene surf-plan leftover recovered. Does not add a surf dictionary. Does not steal Deborah exploring. **Not** n=1540, **not** a Mem0 same-pin. |
| P48 focusing-besides leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. Focusing-besides leftover covering admits first-person focusing-on leftover that covers the besides-object and joins a possessive conjunct on `what has/have/is/was` + focus/focusing + besides/except/aside; covering prefers `focusing on` + ` and my ` / ` and our ` over occupation leftover that also says `focusing on`; lexical search drops `besides`/`except`/`aside`/`lately` only on that query shape; session rank prefers possessive-join leftover sessions over recency occupation chatter. **132/180** (MH **18/33**, OD **4/11**, SH **80/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p48-20260825.md). Jolene relationship leftover recovered. Does not add a relationship dictionary. Does not steal engineering leftover. **Not** n=1540, **not** a Mem0 same-pin. |
| P49 titled-show leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. What-new-series leftover covering admits quoted titled-show leftover on `what` + `new` + `series`/`show`; covering prefers a quoted watch/show/titled leftover over generic `excited about this new journey` leftover and quoted novels; lexical search drops `new`/`fantasy`/`tv`/`series`/`excited`/`about` only on that query shape; subject corpus lists 400 rows so titled leftover at recency 314 can seed session expand. **133/180** (MH **18/33**, OD **4/11**, SH **81/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p49-20260825.md). Wheel of Time leftover recovered. Does not add a Wheel of Time / fantasy dictionary. Does not steal Name of the Wind or Game of Thrones. **Not** n=1540, **not** a Mem0 same-pin. |
| P50 locative-purpose leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. Recently-at leftover covering admits first-person locative leftover that names an extra `for my`/`for our` purpose object on `what did` + `do` + recently/lately + locative `at`/`in`; covering prefers that leftover over thin dated locative compiler facts; lexical search drops `recently`/`lately` only on that query shape; fact-primary recall keeps locative-purpose episode leftover. **134/180** (MH **18/33**, OD **4/11**, SH **82/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p50-20260825.md). Japanese-house album leftover recovered. Does not add an album dictionary. Does not steal Dave congratulations leftover. **Not** n=1540, **not** a Mem0 same-pin. |
| P51 experiencing-feeling leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. How-feel leftover covering admits first-person experiencing leftover that names a new level of feeling on `how` + `about` + feel/felt; covering prefers that leftover over process restatements of practicing mindfulness and thin experiencing compiler facts; lexical search drops `feel`/`about`/`progress` only on that query shape and keeps `mindfulness`/`gratitude`; fact-primary recall keeps experiencing-feeling episode leftover. **135/180** (MH **18/33**, OD **4/11**, SH **83/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p51-20260825.md). Mindfulness-joy leftover recovered. Does not add a joy/happiness/mindfulness dictionary. Does not steal Deborah mix-of-happiness leftover. **Not** n=1540, **not** a Mem0 same-pin. |
| P52 coordinated-use leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. What-do leftover covering admits first-person-plural leftover that names hard work and determination on `what do`/`does`/`did` + two named hop entities + `use`; covering prefers `hard work` + `determination` + we/us/our over dedication restatements, Calvin-only compiler facts, and car-restoration determination leftover; lexical search drops `use`/`reach`/`goal` only on that query shape and keeps person tokens; fact-primary recall keeps work-determination episode leftover. **136/180** (MH **18/33**, OD **4/11**, SH **84/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p52-20260825.md). Hard-work-and-determination leftover recovered. Does not add a hard-work / goals dictionary. Does not steal dedication leftover as Calvin-only. **Not** n=1540, **not** a Mem0 same-pin. |
| P53 self-directed realize leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. What-did leftover covering admits first-person leftover that names a self-directed realize on `what did`/`does`/`has` + realize + `after`; covering prefers realize + `self-`/`myself` + actor over others-directed support leftover, thin believes-self-care compiler facts, and charity-race attendance leftover; lexical search drops `realize`/`after` only on that query shape and keeps event tokens; fact-primary recall keeps self-directed realize episode leftover. **137/180** (MH **18/33**, OD **4/11**, SH **85/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p53-20260825.md). Charity-race self-care leftover recovered. Does not add a charity-race or self-care dictionary. Does not invent Sunday. Does not drop `realize` globally. **Not** n=1540, **not** a Mem0 same-pin. |
| Honesty stop (2026-08-25) | Isolated leftover covering on this skip-ingest 180 is **saturating**. No more covering from this 180's leftover ledger. Do not treat 137/180 as 90%, n=1540, or a Mem0 same-pin. Industry on this tenant is still **62/180**. Full product `/recall` is still **11.4%**. [audit](../research/competitive/benchmax-audit-2026-08-25.md). |
| P54 entity-scoped how-many S0 this-VM (2026-08-26) | Same store, hybrid on. S2 product: hop the counted predicate; entity-scope how-many; collapse bare class labels; keep quantified/modified class phrases; specific heads intersect the typed set. **140/180** (MH **20/33**, OD **4/11**, SH **85/98**, temporal **31/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p54-20260826.md). Children 7→3, ankle 38→2, Sep pets 13→1. Unique losses none. Dec pets still 4 vs gold 3 (Scout stored as Andrew). **Not** leftover covering. **Not** n=1540, **not** a Mem0 same-pin. Not 90% (162/180 on this sample; n=1540 for public LoCoMo). |
| P58 historical typed-set lists S0 this-VM (2026-08-26) | Same store, hybrid on. S2 product: scan past current-state for item/skill/place-set/`activities`+`done` questions; singular `location` stays a point fact; leftover how/why does not widen hop scan. **141/180** (MH **21/33**, OD **4/11**, SH **85/98**, temporal **31/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p58-20260826.md). John outdoor hiking+mountaineering. Unique losses none vs P54. Collars/tricks still miss. **Not** leftover covering. **Not** n=1540, **not** a Mem0 same-pin. Not 90% (162/180 on this sample; n=1540 for public LoCoMo). |
| P59 dest-class lists S0 this-VM (2026-08-26) | Same store, hybrid on. S2 product: dest-being skills without requiring `trick`; transfer-cue item objects (`buy`/`for`); possessed-class dest names (`dog named`, not generic `named`). **143/180** (MH **23/33**, OD **4/11**, SH **85/98**, temporal **31/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p59-20260826.md). Audrey collars/tags/toys/beds; James swim/frisbee/skateboard. Unique losses none vs P58. p59/p59b 142/180 are not pins. **Not** leftover covering. **Not** n=1540, **not** a Mem0 same-pin. Not 90% (162/180 on this sample; n=1540 for public LoCoMo). |
| P61 beneficiary org sets S0 this-VM (2026-08-26) | Same store, hybrid on. S2 product: who/which beneficiary questions recover raise/for-cue affiliation objects (shelter, homeless, hospital) instead of leftover tournament slogans. **144/180** (MH **24/33**, OD **4/11**, SH **85/98**, temporal **31/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p61-20260826.md). Unique losses none vs P59. p60/p60b 143/180 food-set 180s are not pins. **Not** leftover covering. **Not** n=1540, **not** a Mem0 same-pin. Not 90% (162/180 on this sample; n=1540 for public LoCoMo). |
| P42 how-did-start leftover covering S0 this-VM (2026-08-25) | Same store, hybrid on. How-did-start leftover covering admits duration-matched inception leftover on `how did … start … years ago`; covering prefers a multi-stem inception pair over a walking-only duration fact or a gym transformation/journey restatement; lexical search drops start/journey wrappers only on that query shape. **126/180** (MH **18/33**, OD **4/11**, SH **74/98**, temporal **30/38**). [pin](./artifacts/locomo-s0-diag-mh-135-p42-20260825.md). Diet+walking pair leftover recovered. Does not add a diet/walking or gym dictionary. **Not** n=1540, **not** a Mem0 same-pin. |
| BEAM 100K this cycle | [beam-100k-fresh-20260815.md](./artifacts/beam-100k-fresh-20260815.md) |
| BEAM 100K historical | [beam-100k-c0-async.md](./artifacts/beam-100k-c0-async.md) |

Historical staging smoke (2026-07, do not mix with R4h): [locomo-smoke.md](./locomo-smoke.md).

Research notes: [docs/research](../research/README.md).
