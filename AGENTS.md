# AGENTS.md

## Cursor Cloud specific instructions

Brainy is a Go vertical-memory service: an HTTP API (`cmd/api`) plus an async
extraction worker (`cmd/worker`), backed by Postgres. Standard commands live in
`README.md` and `CONTRIBUTING.md`; this section only records non-obvious
setup/run caveats for cloud agents.

### Tests / lint / build (no external services required)

- `go test ./...` is the authoritative suite (mirrors CI in
  `.github/workflows/test.yml`). Postgres-backed tests use
  `github.com/fergusstrange/embedded-postgres`, which downloads a self-contained
  PostgreSQL 17 on first use and caches it. The first `go test` after a cold
  cache needs network and is slow (~minutes); subsequent runs are fast.
- Lint/build: `go vet ./...`, `go build ./cmd/api ./cmd/worker`.

### Running the API / worker locally

The API and worker connect to an external Postgres via `BRAINY_DATABASE_URL`
(default `postgres://brainy:brainy@localhost:5432/brainy?sslmode=disable`).
Migrations auto-apply on API startup; they require the `pg_trgm` extension.
`pgvector` is optional — the vector migration self-skips when the extension is
absent and the service falls back to `REAL[]` hash embeddings + trigram search.

PostgreSQL 16 is preinstalled (not started at boot). Start it and provision the
`brainy` role/db once per fresh VM:

```bash
sudo service postgresql start
sudo -u postgres psql -c "CREATE ROLE brainy LOGIN PASSWORD 'brainy';" 2>/dev/null || true
sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname='brainy'" | grep -q 1 \
  || sudo -u postgres psql -c "CREATE DATABASE brainy OWNER brainy;"
```

Then `go run ./cmd/api` (health at `http://127.0.0.1:8080/healthz`). Docker is
NOT installed here, so the `docker compose up --build` quickstart in `README.md`
does not work; the local Postgres path above is the supported way to run the app
in this environment.

### Auth gotcha

This environment injects staging-style secrets, including `BRAINY_API_KEYS` and
`BRAINY_REQUIRE_API_KEY`. When those are set the API returns `401 unauthorized`
for unauthenticated requests. For the documented local no-auth dev experience,
unset them before starting the API:

```bash
unset BRAINY_API_KEYS BRAINY_REQUIRE_API_KEY
export BRAINY_ENV=local
```

Otherwise authenticate with a key from `BRAINY_API_KEYS` (format `tenant_id:key`,
comma-separated; `*` matches any tenant) via `Authorization: Bearer <key>` or
`X-API-Key`.

### Embeddings / provider

`LLM_*` and `BRAINY_EMBEDDING_*` secrets are injected, so the service boots with
an OpenAI-compatible provider embedder/extractor and search returns a semantic
signal. Clearing those env vars falls back to the offline local hash embedder —
sync `/ingest` and search work fully offline either way.

### Evals

Python eval harnesses under `evals/` are stdlib-only (no `pip install`) and run
against a live API, e.g.:

```bash
python3 evals/run_eval.py --base-url http://127.0.0.1:8080          # Tier 1 parity
python3 evals/run_vertical_eval.py --base-url http://127.0.0.1:8080 # Tier 2 marketing
```
