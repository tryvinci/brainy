# OpMem v0 Baseline Report

**Generated:** 2026-07-03
**Spec:** [docs/research/opmem-spec.md](../research/opmem-spec.md)
**Systems:** Brainy (deterministic + hybrid retrieval, embedded Postgres), verbatim baseline (naive raw-RAG stand-in). Mem0 pending (`MEM0_API_KEY` run).

## Summary

| Category | Brainy | Verbatim |
| --- | --- | --- |
| suppression | 2/3 | 2/3 |
| correction | 3/3 | 2/3 |
| isolation | 3/3 | 3/3 |
| staleness | 1/2 | 2/2 |
| idempotency | 1/1 | 0/1 |
| **overall** | **10/12** | **9/12** |

The failure sets are disjoint except for `sup03`, which is exactly the
discrimination the benchmark is designed to produce: extraction-based systems
and verbatim stores fail in different places.

## Findings

### 1. No system passes durable forget (`sup03`)

Both systems resurrect an explicitly forgotten memory when the same content is
re-ingested. In Brainy this is a real mechanism, not a coincidence: the upsert
path finds the suppressed record by dedupe key and resets `status` to active
(`internal/store/postgres/store.go`, `UpsertMemory` update branch). An operator
suppression is therefore only durable until the original source text is seen
again — a governance hole for taboo/compliance suppressions (cf. `bv02`).

Candidate fix: on upsert conflict with a suppressed record, keep the record
suppressed (require an explicit reinstate path).

### 2. Brainy fails preference staleness (`upd02`)

"I prefer email updates" followed by "I prefer SMS updates" ranks the *stale*
email preference first: both records get identical token/kind scores and the
tie breaks on memory id, i.e. insertion order — oldest first. The matching fact
task (`upd01`) passes only by luck of the deterministic embedder's similarity
tie-break, not by design; recency plays no role in ranking
(`internal/memory/service.go`, `scoreMemory`).

Candidate fix: recency-aware tie-breaking, or supersede semantics when a new
preference/fact conflicts with an existing one of the same shape (ENG-59 / ENG-86
territory).

### 3. Verbatim baseline fails where structured memory wins

The naive store fails correction stickiness (`cor02`: re-ingested stale content
outranks the correction), and idempotency (`dup01`: duplicates pollute recall).
These are precisely the operational behaviors that justify a structured memory
layer over raw RAG — now demonstrated by a reproducible harness instead of
asserted.

## Reproduce

```bash
# harness + fixtures against the in-process baseline (no server needed)
python3 evals/run_opmem.py --systems verbatim

# full run against a live Brainy API
python3 evals/run_opmem.py --systems verbatim,brainy --base-url http://127.0.0.1:8080

# CI-equivalent run (embedded Postgres, spins its own server)
go test ./internal/api/ -run TestOpMemBenchmarkAgainstHTTPServer
```

Raw results are written to `docs/benchmarks/opmem-latest.json` (generated, not
committed).
