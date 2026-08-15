# Evals

Fixture-driven HTTP harnesses. They call a **running** Brainy API and check
ingest, search, recall, and operational flows. Python here is stdlib-only
(no `pip install`).

Start the API first (`docker compose up --build` or `go run ./cmd/api`).

## Usage

Parity fixtures (ingest / search / dedupe / correct):

```bash
python3 evals/run_eval.py --base-url http://localhost:8080
```

Marketing vertical fixtures:

```bash
python3 evals/run_vertical_eval.py --base-url http://localhost:8080
```

Marketing MVP benchmark (parity + vertical suites):

```bash
python3 evals/run_marketing_mvp_benchmark.py --base-url http://localhost:8080
```

Writes `docs/vertical/marketing-mvp-benchmark.json` and `.md`. Capability matrix: `evals/marketing_mvp_matrix.json`.

Correction stickiness:

```bash
python3 evals/correction_stickiness_eval.py --base-url http://localhost:8080
```

OpMem (suppression leaks, correction stickiness, isolation, staleness,
idempotency — spec: `docs/research/opmem-spec.md`, fixtures: `fixtures/opmem/`):

```bash
python3 evals/run_opmem.py --systems verbatim,brainy --base-url http://localhost:8080
```

Task failures are diagnostic (reported, exit 0); only infrastructure errors
fail the run. CI executes it via `TestOpMemBenchmarkAgainstHTTPServer`.

Fixture directories:

- `fixtures/parity/` — core parity
- `fixtures/vertical/marketing/` — marketing pack goldens (BV-01–BV-10, LC-01–LC-02)

CI runs parity, vertical, and MVP suites via `go test ./internal/api/...`.
Docker smoke: `.github/workflows/docker-smoke.yml`.

Optional same-pin compare against another HTTP memory API (research; needs that
system’s key): `evals/run_competitor_benchmark.py`. Do not copy those numbers
into user-facing docs.

## Merge bar

`go test ./...` is the merge bar (includes the HTTP harnesses above). A second
vertical pack is not the next contributor task — talk to maintainers first.
See [docs/vertical/marketing-vetting-gate.md](../docs/vertical/marketing-vetting-gate.md)
if you are extending the marketing fixtures.
