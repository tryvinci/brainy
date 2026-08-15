# Benchmark Methodology

**Status:** Published (Gate M3) + public-suite ladder drafted

## Purpose

Reproducible measurement of Brainy on shared fixtures. Separate **parity**
scenarios (core ingest/search/correct) from **vertical moat** scenarios
(Brainy-only capabilities) from **public long-memory suites** (LOCOMO /
LongMemEval — cited, not yet a qualification pin).

## Suites

| Suite | Directory / upstream | Brainy expectation |
| --- | --- | --- |
| Parity | `fixtures/parity/` | Pass all |
| Marketing vertical | `fixtures/vertical/marketing/` | Pass all |
| OpMem | `fixtures/opmem/` · [spec](../research/opmem-spec.md) | Pass all |
| LOCOMO | [snap-research/locomo](https://github.com/snap-research/locomo) | 1×30 is measurement, not qualification |
| LongMemEval / BEAM | public long-memory datasets | TBD |

Index: [benchmarks README](./README.md) · ladder: [public-bench-ladder](../research/public-bench-ladder.md)

## Runners

```bash
# Brainy only (CI gate)
go test ./...
python3 evals/run_marketing_mvp_benchmark.py --base-url http://localhost:8080
python3 evals/run_opmem.py --systems brainy,verbatim --base-url http://localhost:8080
```

## Scoring

Score **per capability** in `evals/marketing_mvp_matrix.json` and OpMem
categories, not a single aggregate. Brainy documents deterministic hybrid
retrieval in [marketing-moat-report.md](./marketing-moat-report.md).

For LOCOMO-style posts, publish **accuracy × category + p95 latency +
tokens/query**, and always outlink the dataset + judge model. Use the
proveable harness in [`evals/public/`](../../evals/public/) and refuse to
publish if `require_pins()` returns gaps — see
[proveable-eval-framework.md](../research/proveable-eval-framework.md).

## Artifacts

- `docs/vertical/marketing-mvp-benchmark.md` — Brainy Tier 3 report
- `docs/benchmarks/marketing-moat-report.md` — Tier 4 moat report (Gate M3)
- `docs/research/README.md` — research portal
