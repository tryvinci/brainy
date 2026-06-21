# Brainy

Brainy is being rebuilt as a Go-first memory service with Mem0 used as a pinned behavioral reference, not as a codebase to fork wholesale.

## Current Status

The original Python prototype has been archived under `archive/brainy-python-prototype/`.

The active rebuild target is the thin-slice contract in:

- `.omx/plans/v1-contracts-mem0-go-rebuild.md`
- `.omx/plans/prd-mem0-go-rebuild.md`
- `.omx/plans/test-spec-mem0-go-rebuild.md`

This first execution slice is intentionally narrow:

- `POST /ingest`
- `POST /ingest/async`
- `GET /memories/search`
- deterministic local extraction
- Postgres persistence
- duplicate-ingest idempotency
- correction and suppression paths
- async worker pipeline (retries, DLQ)
- no external model, embedding, or reranker dependency (yet)

**Vertical strategy (approved):** Cognitive primitives + YAML vertical packs — not per-vertical DB kinds. See [docs/vertical/verticalization-model.md](docs/vertical/verticalization-model.md). First wedge: marketing (`packs/marketing/v1/pack.yaml`).

## Repository Layout

- `archive/brainy-python-prototype/`: archived Python prototype and tests
- `cmd/api/`: Go API entrypoint
- `cmd/worker/`: Go worker entrypoint
- `internal/`: private Go application packages
- `packs/`: vertical pack definitions (YAML vocabulary, schemas, rank policy)
- `docs/`: rebuild docs, parity tracking, and cutover guidance
- `docs/vertical/`: verticalization model and marketing discovery
- `.omx/plans/`: approved execution artifacts

## Mem0 Reference

The pinned Mem0 reference for this rebuild is tracked in [docs/mem0-parity-matrix.md](/Users/sid/Documents/Projects/vinci/code/brainy/docs/mem0-parity-matrix.md).

Current pinned upstream commit:

- `a670333d67be1207b5be2fc73af60c3439444f48`

## Local Development

### Go tooling

```bash
go test ./...
go run ./cmd/api
go run ./cmd/worker
```

For a long-running worker:

```bash
BRAINY_WORKER_MODE=loop go run ./cmd/worker
```

### Eval harness

```bash
python3 evals/run_eval.py --base-url http://127.0.0.1:8080
```

Current parity fixtures live under `fixtures/parity/`.

For an operator-oriented local setup using an external Postgres instance, see [docs/external-postgres-runbook.md](/Users/sid/Documents/Projects/vinci/code/brainy/docs/external-postgres-runbook.md).

### Environment

Copy or populate `.env.example` as needed.

## Execution Rules

- Preserve the archived prototype until the destructive-change gate passes.
- Do not widen scope beyond the thin-slice contract until deterministic ingest -> persist -> search is green.
- Keep Mem0 parity tracking explicit; do not use "Mem0-inspired" as a substitute for a pinned reference and documented deviations.
