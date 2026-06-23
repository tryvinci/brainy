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

Parity fixtures (Mem0 thin-slice):

```bash
python3 evals/run_eval.py --base-url http://127.0.0.1:8080
```

Marketing vertical fixtures:

```bash
python3 evals/run_vertical_eval.py --base-url http://127.0.0.1:8080
```

Marketing MVP benchmark (parity + vertical suites, Mem0 gap report):

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
- `fixtures/vertical/marketing/` — marketing pack golden scenarios (BV-01–BV-10, LC-01 lifecycle)

CI runs both parity and vertical suites via `go test ./internal/api/...` (`TestEvalHarnessAgainstHTTPServer`, `TestVerticalEvalHarnessAgainstHTTPServer`).
