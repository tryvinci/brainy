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
| BEAM 100K this cycle | [beam-100k-fresh-20260815.md](./artifacts/beam-100k-fresh-20260815.md) |
| BEAM 100K historical | [beam-100k-c0-async.md](./artifacts/beam-100k-c0-async.md) |

Historical staging smoke (2026-07, do not mix with R4h): [locomo-smoke.md](./locomo-smoke.md).

Research notes: [docs/research](../research/README.md).
