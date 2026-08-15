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
# Empirical Mem0 Platform counter-run (needs MEM0_API_KEY):
# python3 evals/run_marketing_mvp_benchmark.py --base-url http://localhost:8080 --systems brainy,mem0
```

Writes `docs/vertical/marketing-mvp-benchmark.json` and `.md`. Capability matrix: `evals/marketing_mvp_matrix.json` (includes declared Mem0 `mem0_has` cells; `--systems mem0` measures them).

Correction stickiness:

```bash
python3 evals/correction_stickiness_eval.py --base-url http://localhost:8080
```

OpMem (suppression leaks, correction stickiness, isolation, staleness,
idempotency — spec: `docs/research/opmem-spec.md`, fixtures: `fixtures/opmem/`):

```bash
python3 evals/run_opmem.py --systems verbatim,brainy --base-url http://localhost:8080
# python3 evals/run_opmem.py --systems verbatim,brainy,mem0 --base-url http://localhost:8080
```

Task failures are diagnostic (reported, exit 0); only infrastructure errors
fail the run. CI executes it via `TestOpMemBenchmarkAgainstHTTPServer`.

Fixture directories:

- `fixtures/parity/` — core parity
- `fixtures/vertical/marketing/` — marketing pack goldens (BV-01–BV-10, LC-01–LC-02)

CI runs parity, vertical, and MVP suites via `go test ./internal/api/...`.
Docker smoke: `.github/workflows/docker-smoke.yml`.

## Same-pin vs Mem0

Evals **name competitors**. The live adapter in this tree is **Mem0 Platform**
(`evals/competitors/mem0_adapter.py`, `https://api.mem0.ai`). That is not Mem0
OSS. There is no Graphiti / Zep runner here yet.

Set `MEM0_API_KEY`. Parity fixtures side-by-side:

```bash
python3 evals/run_competitor_benchmark.py --brainy-url http://localhost:8080
```

Writes `docs/benchmarks/competitor-parity-latest.json`. Marketing empirical
Mem0: `--systems brainy,mem0` on `run_marketing_mvp_benchmark.py`. OpMem:
`--systems …,mem0` on `run_opmem.py`.

LOCOMO same-pin (Mem0 Platform backend): see [public/README.md](public/README.md)
(`--system mem0`). Schema shape matches
[mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks)
`UnifiedResult` 1.0.

Same-pin score tables belong under `docs/research/competitive/` and
`docs/benchmarks/artifacts/`. Do not copy bake-off tables into the product
[README](../README.md). Do not claim SOTA or “beats Mem0.”

## Merge bar

`go test ./...` is the merge bar (includes the HTTP harnesses above). A second
vertical pack is not the next contributor task — talk to maintainers first.
See [docs/vertical/marketing-vetting-gate.md](../docs/vertical/marketing-vetting-gate.md)
if you are extending the marketing fixtures.
