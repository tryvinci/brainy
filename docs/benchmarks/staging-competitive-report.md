# Staging Competitive Report — Brainy vs Mem0

**Generated:** 2026-07-14  
**Brainy:** `https://brainy-api-staging.onrender.com` (`dev` deploy)  
**Mem0:** Platform API (`api.mem0.ai`) with async indexing poll (≥10s)

**Context:** This is an **operational + parity** head-to-head — not a LOCOMO / LongMemEval run. For the public-suite ladder (and outlinks), see [benchmarks README](./README.md) and [public-bench-ladder](../research/public-bench-ladder.md).

---

## Headline

| Suite | Brainy | Mem0 | Verbatim |
| --- | ---: | ---: | ---: |
| Parity (`fixtures/parity/`) | **4/4** | **4/4** | — |
| OpMem v0 (`fixtures/opmem/`) | **12/12** | **9/12** | **9/12** |

Brainy matches Mem0 on thin-slice parity and wins operational correctness (+3 OpMem tasks).

### What this is *not*

| Claim | Status |
| --- | --- |
| LOCOMO accuracy / multi-hop / temporal | **Not run** — dataset: [snap-research/locomo](https://github.com/snap-research/locomo) |
| LongMemEval | **Not run** — runner: [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks) |
| p95 latency + tokens/query (Mem0 blog style) | **Not measured** yet |

---

## OpMem failures (Mem0)

| Task | Failure |
| --- | --- |
| `cor02` correction stickiness | Top recall after revise does not contain corrected value (`ruby`) |
| `sup03` durable forget | Forgotten memory resurfaces after re-remember of same content |
| `upd01` stale fact | Newer fact (`June`) not ranked above older (`March`) |

Brainy passes all three (durable suppression + recency / correction ranking).

---

## Artifacts

- `docs/benchmarks/competitor-parity-staging.json`
- `docs/benchmarks/opmem-staging-vs-mem0.json`

## Reproduce

```bash
export BRAINY_BASE_URL=https://brainy-api-staging.onrender.com
export MEM0_API_KEY=...   # do not commit
python3 evals/run_competitor_benchmark.py --brainy-url "$BRAINY_BASE_URL" \
  --json-out docs/benchmarks/competitor-parity-staging.json
python3 evals/run_opmem.py --systems brainy,mem0,verbatim --base-url "$BRAINY_BASE_URL" \
  --json-out docs/benchmarks/opmem-staging-vs-mem0.json
```

## Notes

- Mem0 add is async (`PENDING`); adapters poll `GET /v1/memories` until indexed.
- Marketing vertical moat (16 fixtures) is Brainy-only by design — see [marketing-moat-report.md](./marketing-moat-report.md).
- Target narrative packaging: [SuperMemory research](https://supermemory.ai/research/) · harness: [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks).
