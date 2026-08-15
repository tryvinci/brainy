# Evals

This directory contains the first fixture-driven eval harness for the Go rebuild.

The harness calls the running API over HTTP and verifies the deterministic thin slice:

- ingest
- search
- duplicate-ingest idempotency
- optional suppression flow

## Usage

Start the API server first:

```bash
go run ./cmd/api
```

Parity fixtures (ingest / search / dedupe / correct):

```bash
python3 evals/run_eval.py --base-url http://127.0.0.1:8080
```

Marketing vertical fixtures:

```bash
python3 evals/run_vertical_eval.py --base-url http://127.0.0.1:8080
```

Marketing MVP benchmark (parity + vertical suites):

```bash
python3 evals/run_marketing_mvp_benchmark.py --base-url http://127.0.0.1:8080
```

Writes `docs/vertical/marketing-mvp-benchmark.json` and `.md`. Capability matrix: `evals/marketing_mvp_matrix.json`.

Correction stickiness:

```bash
python3 evals/correction_stickiness_eval.py --base-url http://127.0.0.1:8080
```

Default fixture directories:
- `fixtures/parity/` — core parity
- `fixtures/vertical/marketing/` — marketing pack golden scenarios (BV-01–BV-10, LC-01–LC-02)

CI runs parity, vertical, and MVP benchmark suites via `go test ./internal/api/...`. Docker smoke: `.github/workflows/docker-smoke.yml`.

Competitor benchmark (optional, requires a competitor API key):

```bash
python3 evals/run_competitor_benchmark.py --brainy-url http://127.0.0.1:8080
```

OpMem operational-correctness benchmark (suppression leaks, correction
stickiness, isolation, staleness, idempotency — spec:
`docs/research/opmem-spec.md`, fixtures: `fixtures/opmem/`):

```bash
python3 evals/run_opmem.py --systems verbatim,brainy --base-url http://127.0.0.1:8080
```

Task failures are diagnostic (reported, exit 0); only infrastructure errors
fail the run. CI executes it via `TestOpMemBenchmarkAgainstHTTPServer`.

## Vetting gates

Marketing must prove technical capabilities before finance or a second vertical. See [`docs/vertical/marketing-vetting-gate.md`](../docs/vertical/marketing-vetting-gate.md).

| Tier | What | Command / CI |
| --- | --- | --- |
| 0 | Go tests | `go test ./...` |
| 1 | Core parity | `evals/run_eval.py` / `TestEvalHarnessAgainstHTTPServer` |
| 2 | Marketing golden | `evals/run_vertical_eval.py` / `TestVerticalEvalHarnessAgainstHTTPServer` |
| 3 | MVP benchmark | `evals/run_marketing_mvp_benchmark.py` / `TestMarketingMVPBenchmarkAgainstHTTPServer` |
| 4 | Full use-case seed coverage + semantic | Gate **M3** — not complete |

**Gate M1** (Tiers 0–3) is the current bar for merge. **Gate M3** is required before finance.
