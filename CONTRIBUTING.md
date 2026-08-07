# Contributing to Brainy

Brainy is an early-stage Go memory service with vertical pack support. Marketing is the first proof vertical — see [`docs/vertical/marketing-vetting-gate.md`](docs/vertical/marketing-vetting-gate.md).

## Development

```bash
go test ./...
go run ./cmd/api
go run ./cmd/worker
```

With Docker:

```bash
docker compose up --build
python3 evals/run_marketing_mvp_benchmark.py --base-url http://127.0.0.1:8080
```

## Git identity

Use this author for commits:

```text
Siddhant Singh <s@siddhant.site>
```

## Pull requests

1. Branch from `dev`.
2. Keep changes focused; match existing Go style.
3. **New behavior requires a fixture** under `fixtures/parity/` or `fixtures/vertical/marketing/` when applicable.
4. Ensure `go test ./...` passes (includes eval e2e harnesses).
5. Update docs if you change API contracts or pack format.

## Vertical packs

- Packs live in `packs/{vertical}/v1/pack.yaml`.
- Do not add per-vertical Postgres `kind` enums — use primitives + labels.
- Finance and other verticals are blocked until marketing Gate M3 — see [`docs/vertical/execution-plan.md`](docs/vertical/execution-plan.md).

## Security

Report vulnerabilities per [`SECURITY.md`](SECURITY.md).
