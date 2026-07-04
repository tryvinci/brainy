# Benchmark Methodology

**Status:** Published (Gate M3)  
**Mem0 reference:** commit `a670333d67be1207b5be2fc73af60c3439444f48`

## Purpose

Reproducible comparison of Brainy vs generic memory APIs on shared fixtures. Separate **parity** scenarios (both systems should behave similarly) from **vertical moat** scenarios (Brainy-only capabilities).

## Suites

| Suite | Directory | Brainy expectation | Mem0 expectation |
| --- | --- | --- | --- |
| Parity | `fixtures/parity/` | Pass all | Pass most (approximate) |
| Marketing vertical | `fixtures/vertical/marketing/` | Pass all | Fail or N/A on differentiation rows |

## Runners

```bash
# Brainy only (CI gate)
go test ./...
python3 evals/run_marketing_mvp_benchmark.py --base-url http://127.0.0.1:8080

# Side-by-side parity (requires MEM0_API_KEY)
python3 evals/run_competitor_benchmark.py --brainy-url http://127.0.0.1:8080
```

## Scoring

Score **per capability** in `evals/marketing_mvp_matrix.json`, not a single aggregate. Mem0 may still win on provider-quality embeddings at scale; Brainy documents deterministic hybrid retrieval in [marketing-moat-report.md](./marketing-moat-report.md).

## Artifacts

- `docs/vertical/marketing-mvp-benchmark.md` — Brainy Tier 3 report
- `docs/benchmarks/marketing-moat-report.md` — Tier 4 moat report (Gate M3)
- `docs/benchmarks/competitor-parity-latest.json` — optional Mem0 side-by-side
