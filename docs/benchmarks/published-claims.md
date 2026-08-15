# Published % claims (sourced)

Vendors publish a **percent per public suite**. This page is that scoreboard:
what they claim, on which metric, with a source. It is **not** a same-pin bake-off.

Lead/trail vs Mem0 Platform uses [same-pin](./README.md#same-pin-comparison) only.
Do not write SOTA / beats-Mem0 from this table.

**As-of:** 2026-08-15. Re-check primary URLs before quoting in a new cycle.

## Headline percents

| System | LoCoMo | LongMemEval | BEAM 1M | BEAM 10M |
| --- | ---: | ---: | ---: | ---: |
| **Brainy** (our last pin on that suite) | **49.8%** full (mean 3-seed, 767/1540, 2026-07-31) · **70.0%** current 1×30 | **4%** (4/100, 2026-08-01) · **0%** LME-20 `/recall` | not run (40% on 100K/20q) | not run |
| **Mem0 Platform** | **92.5%** (1425/1540, top-k 200) | **94.4%** (472/500, top-k 200) | **64.1** mean (70.1% pass) | **48.6** mean (50.5% pass) |
| **Mem0 OSS** | no current published full LoCoMo % | **91.0%** (GPT-5 extract; other extractors 88.6–89.8) | — | — |
| **Zep** | **75.14%** ±0.17 | **71.2%** (in SuperMemory’s GPT-4o table, not a Zep paper) | — | — |
| **SuperMemory** | **77.1** (GPT-4o, 2026-04) | **95%** Recall@15 (GPT-4o) | — | — |
| **Letta** | **74.0%** (filesystem + grep, GPT-4o mini) | — | — | — |
| **Hindsight** | **92.0%** | **94.6%** | **73.9%** | **64.1%** |
| Full-context baseline (often cited) | ~73% | **60.2%** (GPT-4o) | — | — |

**Graphiti OSS** has no separate published LoCoMo / LME / BEAM %. Zep is the product claim (Graphiti is the engine).

## What the percents actually are

| System | Suite | Number | Metric | n / setup | Source |
| --- | --- | ---: | --- | --- | --- |
| Mem0 Platform | LoCoMo | 92.5 | LLM-as-judge, cats 1–4 | 1540 Q, 10 convos, top-k 200 | [docs](https://docs.mem0.ai/core-concepts/memory-evaluation) · [memory-benchmarks](https://github.com/mem0ai/memory-benchmarks) |
| Mem0 Platform | LongMemEval | 94.4 | LLM-as-judge | 500 Q, top-k 200 | same |
| Mem0 Platform | BEAM 1M | 64.1 / 70.1% | mean score / pass@top-200 | 700 Q | docs = mean; harness README = pass (491/700) |
| Mem0 Platform | BEAM 10M | 48.6 / 50.5% | mean score / pass@top-200 | 200 Q | docs = mean; harness README = pass (101/200) |
| Mem0 OSS | LongMemEval | 91.0 | LLM-as-judge | 500 Q; GPT-5 extract, Qwen embedder | [memory-benchmarks OSS table](https://github.com/mem0ai/memory-benchmarks) |
| Mem0 paper (2025) | LoCoMo | 68.5 / 66.9 | LLM-as-judge | Mem0g / Mem0; **not** the 2026 92.5 | Letta and others cite the paper vs 74.0 filesystem |
| Zep | LoCoMo | 75.14 ±0.17 | J-score, cats 1–4 | full suite | [zep-papers](https://github.com/getzep/zep-papers/blob/main/kg_architecture_agent_memory/locomo_eval/README.md) · [blog](https://blog.getzep.com/lies-damn-lies-statistics-is-mem0-really-sota-in-agent-memory/) |
| Zep (Mem0 re-run) | LoCoMo | 58.44 ±0.20 | Mem0’s harness of Zep | disputed | same Zep blog thread |
| Zep (earlier claim) | LoCoMo | ~84 | withdrawn / corrected | arithmetic dispute with Mem0 | same |
| Zep (via SuperMemory) | LongMemEval | 71.2 | GPT-4o, SuperMemory’s table | LongMemEval_s | [LongMemBench](https://supermemory.ai/research/longmembench/) |
| SuperMemory | LoCoMo | 77.1 | GPT-4o | not on their README (they print “#1”) | [issue comment](https://github.com/supermemoryai/supermemory/issues/795#issuecomment-4238031186) |
| SuperMemory | LongMemEval | 95 | **Recall@15** with aggregation, GPT-4o | not LLM-judge | [LongMemBench](https://supermemory.ai/research/longmembench/) |
| SuperMemory | LongMemEval | 85.2 | later Gemini-3 row on the same page | different from 95 R@15 | same |
| Letta | LoCoMo | 74.0 | agent + filesystem tools, GPT-4o mini | vs then-published Mem0 68.5 | [Letta blog](https://www.letta.com/blog/benchmarking-ai-agent-memory) |
| Hindsight | LoCoMo | 92.0 | LLM-judge, single-query, v0.4.19 | AMB manifesto | [AMB](https://hindsight.vectorize.io/blog/2026/03/23/agent-memory-benchmark) |
| Hindsight | LongMemEval | 94.6 | same | same | same |
| Hindsight | BEAM 1M / 10M | 73.9 / 64.1 | published BEAM tiers | [BEAM compare](https://hindsight.vectorize.io/guides/2026/04/21/comparison-agent-memory-benchmark-hindsight-vs-alternatives) |
| Brainy | LoCoMo full | 49.8 mean | LLM-judge, cats 1–4, gpt-oss | 3 seeds × 1540; old stack | [summary](./artifacts/locomo-full-publish-summary.json) |
| Brainy | LoCoMo 1×30 | 70.0 | LLM-judge | 30 Q, remasure 2026-08-15 | [fresh](./artifacts/locomo-fresh-1x30-20260815.md) |
| Brainy | LongMemEval-S | 4.0 | LLM-judge | 100 Q, 2026-08-01 | [lme-s-100](./artifacts/lme-s-100.md) |
| Brainy | LongMemEval-20 | 0.0 | `/recall` integrity | 20 Q | cycle closeout |
| Brainy | BEAM 100K | 40.0 | 20 probing Q, 1 convo | not 1M/10M | [beam-100k](./artifacts/beam-100k-c0-async.md) |

## LoCoMo category % (published full-suite only)

Do not mix these with Brainy 1×30 category n/N.

| System | Overall | Single-hop | Multi-hop | Open-domain | Temporal |
| --- | ---: | ---: | ---: | ---: | ---: |
| Mem0 Platform (avg top-10…200) | 92.5 | 91.2 | 91.3 | **72.7** | 92.0 |
| Zep (corrected) | 75.14 | 79.79 | 74.11 | 66.04 | 67.71 |
| Brainy full seed-0 (2026-07-31) | 49.4 | 56.7 | **25.2** | 38.5 | 54.8 |

Mem0’s own weak axis on the published full suite is **open-domain** (72.7). Ours on the last full pin is **multi-hop** (25.2). Current 1×30 is a different n (MH 10/10, OD 0/4).

## How to read this vs same-pin

- **92.5 vs 49.8** is the industry-format LoCoMo compare (full n≈1540). A fresh n=1540 remasure is in flight.
- **92.5 vs 70.0** is **not** that compare (n=1540 vs n=30).
- **94.4 vs 4** is the industry-format LongMemEval compare (hundreds of Q vs our 100-Q pin). **94.4 vs 0/20** is worse: integrity pin, not quality.
- **95% Recall@15** is not an LLM-judge percent. Do not rank SuperMemory 95 against Mem0 94.4 as if the same metric.
- **64.1** on BEAM is Mem0’s mean score; their harness also prints **70.1% pass**. Hindsight’s 73.9 / 64.1 are a different published BEAM run.

Same-pin Mem0 Platform LoCoMo 1×30 this cycle is **36.7% (11/30)** vs Brainy **70.0% (21/30)**.
