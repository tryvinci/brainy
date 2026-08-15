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
| **Brainy** | **49.8%** last full (n=1540, 2026-07-31) · **66.7%** 1×30 now | **4%** (n=100) · **0%** LME-20 | not run (40% on 100K/20q) | not run |
| **Mem0 Platform** | **92.5%** | **94.4%** | **64.1** | **48.6** |
| **Zep** | **75.14%** | 71.2% | — | — |
| **SuperMemory** | 77.1% | 95% Recall@15 | — | — |
| **Letta** | 74.0% | — | — | — |
| **Hindsight** | 92.0% | 94.6% | 73.9% | 64.1% |

Sources, n, metric type, and the Zep LoCoMo dispute:
[published-claims.md](./published-claims.md). Peer harness:
[mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks).

## Same-pin comparison

Pin date for Brainy: **R4h, 2026-08-15** (`f4ec4d7` product SHA; docs may be
later). Mem0 LoCoMo freeze: **2026-08-13**.

| Suite | Brainy | Mem0 Platform | Graphiti OSS / Zep Platform | Stand |
| --- | ---: | ---: | --- | --- |
| LoCoMo 1×30 overall | **20/30 (66.7%)** | **12/30 (40.0%)** frozen | **no same-pin** | Lead this freeze; **not** SOTA |
| Multi-hop | **10/10** | **7/10** | no same-pin | Lead this freeze |
| Open-domain | **0/4** | **3/4** | no same-pin | **Trail** |
| Temporal | **10/16** | **2/16** | no same-pin | Lead this freeze |
| OpMem | **13/13** | **9/12** (2026-07-14, **not re-run**) | no pin | Lead ops; Mem0 pin is stale |
| Marketing vertical | **17/17** | **4/16** empirical (2026-07-29, **not re-run**) | no pin | Lead governed vertical; Mem0 pin is stale |
| LongMemEval-20 | **0/20** integrity | no fair pin on this harness | no pin | Neither is a quality win |
| BEAM | **not re-run** this cycle | published elsewhere; not our pin | no pin | — |

Search p50 on the LoCoMo 1×30 pin: Brainy **125 ms** local vs Mem0 Platform
**471 ms**. That is a harness observation, **not** a production SLO.

Mem0 OSS was **not** re-measured. Do not treat Platform 12/30 as an
OSS-reproducible number.

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
| **LongMemEval** | Long-term extraction, temporal, multi-session | LongMemEval dataset | `evals/public/longmemeval/` — quality still **0/20** on `/recall` |
| **BEAM** | Retrieval across 100K–10M token chats | HuggingFace `Mohammadta/BEAM` | `evals/public/beam/` — **not re-run** this cycle (historical 8/20 on 100K/20q: [beam-100k-c0-async](./artifacts/beam-100k-c0-async.md)) |
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

## Current Brainy artifacts (R4h)

| Suite | Report |
| --- | --- |
| LoCoMo 1×30 | [locomo-mh-r4h-dev-1x30-20260815.md](./artifacts/locomo-mh-r4h-dev-1x30-20260815.md) |
| Mem0 LoCoMo freeze | [locomo-mem0-samepin-pr10-20260813.md](./artifacts/locomo-mem0-samepin-pr10-20260813.md) |
| OpMem | [opmem-mh-r4h-local-20260815.md](./artifacts/opmem-mh-r4h-local-20260815.md) |
| Marketing | [marketing-mh-r4h-local-20260815.md](./artifacts/marketing-mh-r4h-local-20260815.md) · [moat](./marketing-moat-report.md) |
| Cycle closeout (detailed why) | [cycle-closeout.md](../research/competitive/cycle-closeout.md) |
| BEAM 100K historical | [beam-100k-c0-async.md](./artifacts/beam-100k-c0-async.md) — do not mix with R4h |

Historical staging smoke (2026-07, do not mix with R4h): [locomo-smoke.md](./locomo-smoke.md).

Research notes: [docs/research](../research/README.md).
