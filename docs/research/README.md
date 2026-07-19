# Brainy Research

Benchmark-backed claims for vertical memory. Goal: a public research surface in the spirit of [SuperMemory Research](https://supermemory.ai/research/) and Mem0’s LOCOMO writeups — **reproducible, cited, honest about gaps**.

---

## Published

| Piece | Status | Link |
| --- | --- | --- |
| OpMem v0 (operational memory) | Live results | [spec](./opmem-spec.md) · [staging vs Mem0](../benchmarks/staging-competitive-report.md) |
| Marketing vertical moat | Live results | [moat report](../benchmarks/marketing-moat-report.md) |
| Launch narrative | Draft | [launch narrative](../benchmarks/launch-narrative.md) |
| LOCOMO smoke (L3) | Live mid score | [locomo-smoke.md](../benchmarks/locomo-smoke.md) · [ladder](./public-bench-ladder.md) |
| Public-bench ladder (LOCOMO …) | L3 remeasured; L4 gated | [public-bench-ladder.md](./public-bench-ladder.md) |
| Proveable eval framework | Spec + L2/L3 harness | [proveable-eval-framework.md](./proveable-eval-framework.md) · [`evals/public/`](../../evals/public/) |

---

## Headline (today)

**Brainy 12/12 OpMem vs Mem0 9/12** on staging (parity 4/4), plus marketing vertical **16/16**.

**LOCOMO smoke (same pins, 1 convo / 30 Q):** **9/30 (30%)** after async provider-extract path + relative event-time enrichment — honest mid score, **not** a SOTA claim. See [locomo-smoke.md](../benchmarks/locomo-smoke.md) for category breakdown; multi-hop **0/10**. Full LOCOMO (L4) waits on ≥12/30 smoke gate. Details: [locomo-smoke.md](../benchmarks/locomo-smoke.md).

---

## External datasets we cite

Always outlink upstream when discussing public suites:

- **LOCOMO** — [snap-research/locomo](https://github.com/snap-research/locomo) · [ACL 2024](https://aclanthology.org/2024.acl-long.747/) · [project site](https://snap-research.github.io/locomo/)
- **Memory harness (runners)** — [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks) (LOCOMO / LongMemEval / BEAM runners + Mem0 adapters)
- **Industry research pages** — [SuperMemory research](https://supermemory.ai/research/) · [LongMemBench](https://supermemory.ai/research/longmembench/)

---

## Content format we are targeting

Posts should eventually look like Mem0’s LOCOMO-style note and SuperMemory’s research pages:

1. **One-line headline** with the primary metric (accuracy / pass rate)
2. **Systems compared** (Brainy, Mem0, …) + version pins
3. **Table**: accuracy · p50/p95 latency · tokens/query (when measured)
4. **Category breakdown** (temporal, multi-hop, suppression, …)
5. **Methodology** — dataset link, judge model, temperature, retrieval policy
6. **Reproduce** — exact CLI against staging or tagged release
7. **Limitations** — what we did *not* claim

First posts can ship with **empty / TBD cells** for latency and LOCOMO — preferable to inventing numbers.

---

## Next engineering step

See [public-bench-ladder.md](./public-bench-ladder.md): raise LOCOMO smoke past the **12/30** L4 unlock (working dense embeddings + multi-hop synthesis), then full LOCOMO + MarketingMem public track.
