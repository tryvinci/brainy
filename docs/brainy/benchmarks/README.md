# Benchmarking Framework

## Tracks
1. Public memory track
- Measures baseline end-to-end memory retrieval behavior.
- Includes Brainy run and competitor adapter stubs.

2. Brainy cognitive track
- Measures belief revision behavior under conviction failure.
- Verifies challenge transitions and reflection updates.

## Reproducibility
- Command: `python3 scripts_run_benchmark.py`
- JSON report: `docs/brainy/benchmarks/latest-report.json`
- Markdown report: `docs/brainy/benchmarks/latest-report.md`

## Competitor Runs
- Adapters are API-key gated.
- Expected keys:
  - `MEM0_API_KEY`
  - `SUPERMEMORY_API_KEY`
  - `ZEP_API_KEY`
  - `LETTA_API_KEY`
  - `MEMOBASE_API_KEY`
  - `COGNEE_API_KEY`

## Caveats
- Competitor adapters are scaffolded but not fully implemented yet.
- No state-of-the-art claim should be made until all adapters are implemented and runs are archived.
