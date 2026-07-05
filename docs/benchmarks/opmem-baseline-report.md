# OpMem v0 Baseline Report

**Generated:** 2026-07-05  
**Spec:** [docs/research/opmem-spec.md](../research/opmem-spec.md)  
**Systems:** Brainy (deterministic + hybrid retrieval), verbatim baseline (naive raw-RAG stand-in).

## Summary

| Category | Brainy | Verbatim |
| --- | --- | --- |
| suppression | 3/3 | 2/3 |
| correction | 3/3 | 2/3 |
| isolation | 3/3 | 3/3 |
| staleness | 2/2 | 2/2 |
| idempotency | 1/1 | 0/1 |
| **overall** | **12/12** | **9/12** |

Brainy passes all OpMem v0 tasks after PR #17:

| Task | Fix |
| --- | --- |
| `sup03` durable forget | Suppressed records stay suppressed on dedupe-key re-ingest |
| `upd02` preference staleness | Relative recency boost breaks score ties (newer wins) |

Verbatim still fails where structured memory is expected to win (`cor02`, `dup01`, `sup03`).

## Reproduce

```bash
# CI-equivalent (embedded Postgres, spins its own server)
go test ./internal/api/ -run TestOpMemBenchmarkAgainstHTTPServer

# full run against a live Brainy API
docker compose up -d --build
python3 evals/run_opmem.py --systems verbatim,brainy --base-url http://127.0.0.1:8080

# harness only (verbatim, no server)
python3 evals/run_opmem.py --systems verbatim
```

Raw results: `docs/benchmarks/opmem-latest.json` (generated, not committed).

## See also

- [Marketing moat report](./marketing-moat-report.md)
- [Launch narrative](./launch-narrative.md)
- [METHODOLOGY.md](./METHODOLOGY.md)
