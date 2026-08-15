# Contributing to Brainy

Thanks for wanting to help. By participating you agree to the
[Code of Conduct](CODE_OF_CONDUCT.md).

Search [existing issues](https://github.com/tryvinci/brainy/issues) before
opening a new one. Support and docs pointers: [SUPPORT.md](SUPPORT.md).

## Ways to contribute

- Bug reports and reproductions
- Fixes and tests (`go test ./...` is the bar)
- Docs (especially [docs/api.md](docs/api.md) and the Quick Start in [README.md](README.md))
- Fixtures when you change product behavior (`fixtures/parity/`, `fixtures/vertical/marketing/`)
- Vertical pack YAML — not per-vertical database kinds

Security issues: **do not** open a public issue. See [SECURITY.md](SECURITY.md).

## Prerequisites

- Go **1.25+** (see `go.mod`)
- Docker (Compose path) **or** Postgres 17 with `pg_trgm` (`pgvector` optional)
- Python 3 (evals are stdlib-only; no `pip install`)

## Setup

```bash
git clone https://github.com/<your-fork>/brainy.git
cd brainy
git remote add upstream https://github.com/tryvinci/brainy.git
cp .env.example .env   # optional; Compose works without it
```

**Docker (easiest):**

```bash
docker compose up --build
curl -s http://localhost:8080/healthz
```

**From source:** set `BRAINY_DATABASE_URL`, then:

```bash
export BRAINY_ENV=local
unset BRAINY_API_KEYS BRAINY_REQUIRE_API_KEY
go run ./cmd/api
BRAINY_WORKER_MODE=loop go run ./cmd/worker
```

Auth: if `BRAINY_API_KEYS` is set, unauthenticated requests return 401. Local
no-auth is `BRAINY_ENV=local` with those vars unset. Keys are `tenant_id:key`
pairs (`*` matches any tenant).

Operator Postgres: [docs/external-postgres-runbook.md](docs/external-postgres-runbook.md).

## Tests

```bash
gofmt -l .
go vet ./...
go test ./...
```

`go test ./...` is what CI runs ([`.github/workflows/test.yml`](.github/workflows/test.yml)).
Postgres-backed tests use embedded PostgreSQL 17 (first run downloads a runtime
and is slow). Docker smoke: [`.github/workflows/docker-smoke.yml`](.github/workflows/docker-smoke.yml).

Evals need a live API:

```bash
python3 evals/run_eval.py --base-url http://localhost:8080
python3 evals/run_vertical_eval.py --base-url http://localhost:8080
python3 evals/run_opmem.py --base-url http://localhost:8080
```

## Pull requests

1. Fork and branch from **`dev`** (staging). `main` is production.
2. Keep the change focused. One logical change per PR.
3. Match surrounding Go style (`gofmt`). No new dependencies unless they remove
   more complexity than they add.
4. New behavior needs a fixture when it touches ingest, search, recall, or packs.
5. Update docs if you change HTTP contracts or pack format.
6. Make sure `go test ./...` is green.

Open the PR against `dev`. Use the PR template (summary, changes, test plan).

### Git identity

Use **your** `user.name` / `user.email`. Do not commit as the maintainer unless
they asked you to author on their behalf.

```bash
git config user.name "Your Name"
git config user.email "you@example.com"
```

Automated maintainer agents follow [AGENTS.md](AGENTS.md).

## Repository layout

- `cmd/api`, `cmd/worker` — binaries
- `internal/` — Go packages
- `packs/` — YAML vertical packs
- `fixtures/` — eval goldens
- `evals/` — HTTP eval harnesses (stdlib Python)
- `docs/` — human docs; start at [docs/README.md](docs/README.md)
- `docs/research/` — internal notes, not the getting-started path

## Packs

- Packs live in `packs/{vertical}/v1/pack.yaml`.
- Do not add per-vertical Postgres `kind` enums — primitives + labels.
- Talk to maintainers before starting a second vertical pack.

## License

Contributions are licensed under the [Apache License 2.0](LICENSE) (same as the
repository). Do not contribute code you cannot license that way. Cite the
project with [CITATION.cff](CITATION.cff).
