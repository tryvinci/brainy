# Benchmark Methodology

**Status:** Draft (Gate M2)  
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

Score **per capability** in `evals/marketing_mvp_matrix.json`, not a single aggregate. Document Mem0 wins on semantic paraphrase until ENG-87 ships.

## Artifacts

- `docs/vertical/marketing-mvp-benchmark.md` — Brainy Tier 3 report
- `docs/benchmarks/competitor-parity-latest.json` — optional Mem0 side-by-side
