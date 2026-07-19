# AGENTS.md

## Cursor Cloud specific instructions

Brainy is a Go-first HTTP "vertical memory" service. Standard dev/test/build commands
live in `CONTRIBUTING.md` and `README.md` (`go test ./...`, `go run ./cmd/api`,
`go run ./cmd/worker`, docker compose flow). Only the non-obvious cloud caveats are below.

### Services
- **API** (`cmd/api`, port `8080`) and **worker** (`cmd/worker`) — both connect to Postgres
  via `BRAINY_DATABASE_URL` and run migrations on boot.
- **PostgreSQL** — required for the API/worker to boot. Tests do NOT need it (they use an
  embedded Postgres binary downloaded on first `go test` run).

### Postgres in this environment (no Docker)
Docker is not available here, so we do not use `docker compose`. Instead a native Postgres 16
cluster is installed and seeded (role `brainy`/`brainy`, database `brainy`), matching the default
`BRAINY_DATABASE_URL` DSN in `.env.example` / `docker-compose.yml` (localhost:5432).

The cluster does not auto-start on VM boot — start it before running the API/worker:
```bash
sudo pg_ctlcluster 16 main start
```

Notes:
- `pgvector` is NOT installed. This is fine: migration `enable_pgvector_optional` catches the
  missing `vector` extension and no-ops, so search falls back to the hybrid/keyword path. Do not
  treat the absent `vector` extension as an error.
- `pg_trgm` (from `postgresql-contrib`) is required and already available; `brainy` owns the DB so
  it can create it.

### Running the app end-to-end
```bash
sudo pg_ctlcluster 16 main start   # if not already running
export BRAINY_DATABASE_URL="$(grep '^BRAINY_DATABASE_URL=' .env.example | cut -d= -f2-)"
go run ./cmd/api                   # API on :8080
BRAINY_WORKER_MODE=once go run ./cmd/worker   # drain async jobs once (default mode)
```
Async ingest (`POST /ingest/async`) only produces searchable memories after the worker runs.

### Evals (optional)
Python 3.12 stdlib-only harnesses under `evals/` run against a live API, e.g.
`python3 evals/run_marketing_mvp_benchmark.py --base-url http://127.0.0.1:8080`.
