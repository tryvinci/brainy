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

Then run:

```bash
python3 evals/run_eval.py --base-url http://127.0.0.1:8080
```

The default fixture directory is `fixtures/parity/`.
