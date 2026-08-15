# Marketing Moat Benchmark Report

**Status:** Published (Gate M3)
**Generated:** 2026-06-30
**Methodology:** [METHODOLOGY.md](./METHODOLOGY.md)

## Executive summary

Brainy passes **all Tier 1–4 marketing vetting suites** on the Go rebuild:

| Tier | Suite | Result |
| --- | --- | --- |
| 0 | `go test ./...` | **Pass** |
| 1 | Parity fixtures (`fixtures/parity/`) | **4/4** |
| 2 | Marketing vertical fixtures | **16/16** |
| 3 | MVP benchmark matrix | **5/5 differentiation** |
| 4 | Use-case eval seeds + hybrid retrieval | **10/10 seeds + hybrid01** |

**Gate M3 verdict:** Marketing technical proof achieved on deterministic + hybrid retrieval. Finance (Gate M4) may proceed to research/pack drafting after architecture sign-off.

## Vertical capabilities

| Capability | Status | Notes |
| --- | --- | --- |
| Principle > preference hierarchy | pass | `bv01` |
| Marketing voice_profile mapping | pass | pack vocabulary |
| Never-sentence → brand_rule | pass | `bv08` |
| Multi-message voice + rule extraction | pass | `bv10` |
| Campaign lifecycle suppression | pass | `lc01`, `lc02` |
| Outcome → belief rank loop | pass | `ob05` |
| Cross-campaign pattern retrieval | pass | `pt08` |
| TasteSignal style ranking | pass | `ts09` |
| Scoped segment coexistence | pass | `sg10` |
| Paraphrase hybrid retrieval | pass | `hybrid01` — local deterministic embedder |

Brainy uses a deterministic local embedder for CI reproducibility. Hosted embedding quality is an operator choice.

## Tier 4 seed coverage

| # | Scenario | Fixture | Status |
| --- | --- | --- | --- |
| 1 | Principle over preference | `bv01` | pass |
| 2 | Taboo suppression | `bv02` | pass |
| 3 | Active > completed campaign | `lc02` | pass |
| 4 | Archived campaign hidden | `lc01` | pass |
| 5 | Outcome → belief rank | `ob05` | pass |
| 6 | Correction stickiness | `bv04` | pass |
| 7 | Multi-brand isolation | `bv06` | pass |
| 8 | Cross-campaign pattern | `pt08` | pass |
| 9 | Style-matched creative | `ts09` | pass |
| 10 | Scoped segment prefs | `sg10` | pass |

## Hybrid retrieval (ENG-87)

- **Storage:** `memory_embeddings` table with `REAL[]`; pgvector `vector(768)` + HNSW for hosted pin (Docker: `pgvector/pgvector:pg17`); hash embedder stays 128-d float[] for tests
- **Embedder:** deterministic local hash embedder (`internal/embedding/local.go`) — no external API in CI
- **Search:** token ILIKE + cosine similarity blend; paraphrase recall when token score = 0 and similarity ≥ 0.15
- **Regression:** all parity and vertical fixtures unchanged

## Reproduce

```bash
go test ./...
python3 evals/run_hybrid_eval.py --base-url http://localhost:8080
python3 evals/run_marketing_mvp_benchmark.py --base-url http://localhost:8080
```

## References

- [marketing-mvp-benchmark.md](../vertical/marketing-mvp-benchmark.md) — Tier 3 live report
- [marketing-vetting-gate.md](../vertical/marketing-vetting-gate.md) — gate policy
- [execution-plan.md](../vertical/execution-plan.md) — Linear ↔ GitHub sync
