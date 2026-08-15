# AGENTS.md

Notes for automated coding agents working in this repository.

## Git identity

Maintainer and cloud agents author commits as:

```text
Siddhant Singh <s@siddhant.site>
```

```bash
git config user.name "Siddhant Singh"
git config user.email "s@siddhant.site"
```

Human contributors use their own locally configured identity — see
`CONTRIBUTING.md`. Do not rewrite someone else's commits as the maintainer
unless they asked you to author on their behalf.

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

## Public docs voice

The product README **may** include a **same-pin comparison summary** vs named
systems (today: Mem0 Platform; Graphiti/Zep as **no pin**), outlinking to
[docs/benchmarks/README.md](docs/benchmarks/README.md). Do **not** mix vendor
blog / [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks)
90%+ headlines into our n/N table. Do not write SOTA / beats-Mem0. 1×30 is
measurement. Trail axes (today: open-domain) must stay visible.

GTM, launch narrative, and the commercial checklist stay Brainy-product copy
(no bake-off tables). Quick Start stays how-to-run.

**Evals may name competitors** (`evals/`, including `evals/README.md` and
`evals/public/`). Name the system the harness actually calls (today: Mem0
Platform vs Mem0 OSS). Do not invent pins.

Contributor-facing layout follows [CONTRIBUTING.md](CONTRIBUTING.md),
[docs/README.md](docs/README.md), and [docs/api.md](docs/api.md) — do not dump
research notes into the README.

Detailed same-pin **why / next** belongs under
[docs/research/competitive/cycle-closeout.md](docs/research/competitive/cycle-closeout.md)
and historical `docs/benchmarks/artifacts/`. The README summary must outlink
the full [benchmarks](docs/benchmarks/README.md) page rather than paste
cycle-closeout.

## Benchmark cycle closeout (required)

Every remasure, merge, or “where we landed” cycle must report **in this order**, in **both** the user-facing summary and a new dated section of [docs/research/competitive/cycle-closeout.md](docs/research/competitive/cycle-closeout.md). Scores-only is incomplete. A Brainy pin without a detailed competitor compare **in the competitive folder** is not a cycle closeout. The **README** half is a same-pin **summary table** plus outlink to [docs/benchmarks/README.md](docs/benchmarks/README.md) (no SOTA; trail axes named). Evals may name competitors.

1. **Landed** — SHAs on `dev` / `main`, what product change shipped.
2. **Own pins** — OpMem, marketing, LoCoMo 1×30 **by category** (MH / OD / temporal), LME if run. Name dips as dips. 1×30 is measurement, not qualification.
3. **Competitor compare (detailed)** — required every cycle, **in** [cycle-closeout.md](docs/research/competitive/cycle-closeout.md). README gets a short same-pin summary that outlinks [docs/benchmarks/README.md](docs/benchmarks/README.md). Same-pin only for lead/trail. See [competitive/README.md](docs/research/competitive/README.md).
   - For each trailing axis: the product mechanism and the PoR step that closes it. For each leading axis: what we must not regress.
   - Unpublished vendors: write **no pin** unless we ran them. Published headlines are **context**, never scoreboard rows.
4. **Why the delta** — product mechanism (compiler coverage, provenance crowding, reader). Not vibes.
5. **Next** — one step on [docs/research/sota-representation-path.md](docs/research/sota-representation-path.md), mapped to the largest gap (today: OD 0/4 → R5 structured-first answer). Kill list: no fusion fishing, no graph DB default, no category dictionaries, no unbounded top-k, no LoCoMo/LME-named product rules, no SOTA claims.

`dev` is staging. `main` is production — only fast-forward `main` with explicit user approval.
