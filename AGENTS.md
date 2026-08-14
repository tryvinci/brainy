# AGENTS.md

Notes for automated coding agents working in this repository.

## Git identity

Author and committer for all commits:

```text
Siddhant Singh <s@siddhant.site>
```

```bash
git config user.name "Siddhant Singh"
git config user.email "s@siddhant.site"
```

## Cloud / ephemeral VM notes

Brainy is a Go vertical-memory service: an HTTP API (`cmd/api`) plus an async
extraction worker (`cmd/worker`), backed by Postgres. Standard commands live in
`README.md` and `CONTRIBUTING.md`; this section records non-obvious setup/run
caveats for cloud agent VMs.

### Tests / lint / build (no external services required)

- `go test ./...` is the authoritative suite (mirrors CI in
  `.github/workflows/test.yml`). Postgres-backed tests use
  `github.com/fergusstrange/embedded-postgres`, which downloads a self-contained
  PostgreSQL 17 on first use and caches it. The first `go test` after a cold
  cache needs network and is slow (~minutes); subsequent runs are fast.
- Lint/build: `go vet ./...`, `go build ./cmd/api ./cmd/worker`.

### Running the API / worker locally

The API and worker connect to Postgres via `BRAINY_DATABASE_URL`. Migrations
auto-apply on API startup; they require the `pg_trgm` extension. `pgvector` is
optional — the vector migration self-skips when the extension is absent and the
service falls back to `REAL[]` hash embeddings + trigram search.

If PostgreSQL is preinstalled but not running:

```bash
sudo service postgresql start
sudo -u postgres psql -c "CREATE ROLE brainy LOGIN PASSWORD 'brainy';" 2>/dev/null || true
sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname='brainy'" | grep -q 1 \
  || sudo -u postgres psql -c "CREATE DATABASE brainy OWNER brainy;"
```

Then `go run ./cmd/api` (health at `http://127.0.0.1:8080/healthz`).

### Auth gotcha

When `BRAINY_API_KEYS` / `BRAINY_REQUIRE_API_KEY` are set, the API returns
`401 unauthorized` for unauthenticated requests. For local no-auth:

```bash
unset BRAINY_API_KEYS BRAINY_REQUIRE_API_KEY
export BRAINY_ENV=local
```

Otherwise authenticate with a key from `BRAINY_API_KEYS` (format `tenant_id:key`,
comma-separated; `*` matches any tenant) via `Authorization: Bearer <key>` or
`X-API-Key`.

### Embeddings / provider

`LLM_*` and `BRAINY_EMBEDDING_*` may be injected. Clearing those env vars falls
back to the offline local hash embedder — sync `/ingest` and search still work.

### Evals

Python eval harnesses under `evals/` are stdlib-only (no `pip install`) and run
against a live API, e.g.:

```bash
python3 evals/run_eval.py --base-url http://127.0.0.1:8080
python3 evals/run_vertical_eval.py --base-url http://127.0.0.1:8080
```

## Benchmark cycle closeout (required)

Every remasure, merge, or “where we landed” cycle must report **in this order**, in **both** the user-facing summary and a new dated section of [docs/research/competitive/cycle-closeout.md](docs/research/competitive/cycle-closeout.md). Scores-only is incomplete. A Brainy pin without a detailed competitor compare is not a cycle closeout.

1. **Landed** — SHAs on `dev` / `main`, what product change shipped.
2. **Own pins** — OpMem, marketing, LoCoMo 1×30 **by category** (MH / OD / temporal), LME if run. Name dips as dips. 1×30 is measurement, not qualification.
3. **Competitor compare (detailed)** — required every cycle, not optional color. Same-pin only for lead/trail:
   - Mem0 OSS ≠ Mem0 Platform; Graphiti ≠ Zep Platform.
   - Table: overall + MH + OD + temporal vs last frozen Mem0 LoCoMo same-pin [docs/benchmarks/artifacts/locomo-mem0-samepin-pr10-20260813.md](docs/benchmarks/artifacts/locomo-mem0-samepin-pr10-20260813.md) (12/30, MH 7/10, OD 3/4, temporal 2/16). Say trail/lead **per axis**.
   - OpMem / marketing vs Mem0: [docs/benchmarks/staging-competitive-report.md](docs/benchmarks/staging-competitive-report.md), [docs/vertical/marketing-mvp-vs-mem0.md](docs/vertical/marketing-mvp-vs-mem0.md). Mark the Mem0 pin date; re-run Mem0 before claiming a **new** lead.
   - For each trailing axis: the product mechanism and the PoR step that closes it. For each leading axis: what we must not regress.
   - Graphiti/Zep: write **no pin** unless we ran them. Published headlines are **context**, never scoreboard rows. Do not invent a Graphiti/Zep LoCoMo number.
4. **Why the delta** — product mechanism (compiler coverage, provenance crowding, reader). Not vibes.
5. **Next** — one step on [docs/research/sota-representation-path.md](docs/research/sota-representation-path.md), mapped to the largest competitor gap (today: MH 2/10 vs Mem0 7/10 → R1b coverage, then entities/relations). Kill list: no fusion fishing, no graph DB default, no category dictionaries, no unbounded top-k, no LoCoMo/LME-named product rules, no SOTA / beats-Mem0 claims.

`dev` is staging. `main` is production — only fast-forward `main` with explicit user approval.
