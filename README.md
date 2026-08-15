# Brainy

Brainy is a Go **vertical memory** service: an HTTP API (`cmd/api`) plus an async
worker (`cmd/worker`), Postgres persistence, and YAML vertical packs. Marketing
is the first wedge. Mem0 is a pinned behavioral reference, not a fork target.

## 5-Minute Quickstart

```bash
git clone https://github.com/tryvinci/brainy.git && cd brainy
docker compose up --build -d
# API + worker + Postgres. Wait for health (~30s).
curl -s http://localhost:8080/healthz
```

Local Compose sets `BRAINY_ENV=local` (no API key). The image includes
`tesseract-ocr` so `image_urls` on ingest can extract cover text.

Ingest, search, and recall:

```bash
curl -s -X POST http://localhost:8080/ingest \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":"demo","subject_id":"user-1","source_type":"conversation","messages":[{"role":"user","content":"We never use exclamation marks in brand copy."}]}'

curl -s 'http://localhost:8080/memories/search?tenant_id=demo&subject_id=user-1&q=brand+voice+rules'

curl -s -X POST http://localhost:8080/recall \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":"demo","subject_id":"user-1","q":"brand voice rules","mode":"enumerate"}'
```

Async ingest (worker already looping in Compose):

```bash
curl -s -X POST http://localhost:8080/ingest/async \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":"demo","subject_id":"user-1","source_type":"conversation","messages":[{"role":"user","content":"Ship the autumn campaign on 12 September."}]}'
```

Run evals against the live API:

```bash
python3 evals/run_vertical_eval.py --base-url http://localhost:8080
python3 evals/run_opmem.py --base-url http://localhost:8080
python3 evals/run_marketing_mvp_benchmark.py --base-url http://localhost:8080
```

**Self-host preview.** Docker Compose is the supported local path. Hosted
design-partner beta uses per-tenant API keys — see
[commercial-beta-checklist.md](docs/commercial-beta-checklist.md). ToS, privacy
policy, and PITR backups are still required before GA. This is not a SOTA claim.

Conversation ingest (including `image_urls`): [docs/conversation-ingest.md](docs/conversation-ingest.md).
Auth gotcha and cloud-agent notes: [AGENTS.md](AGENTS.md).

## Current pins (honest)

1×30 LoCoMo conv-26 is **measurement, not qualification**. Same dataset SHA and
judge as the frozen Mem0 Platform pin. Open-domain still trails. Do not read
these as SOTA or “beats Mem0.”

| Suite | Brainy | Mem0 same-pin | Report |
| --- | ---: | ---: | --- |
| LoCoMo 1×30 | **20/30** (MH **10/10**, OD **0/4**, temporal **10/16**) | 12/30 (MH 7/10, OD 3/4, temporal 2/16) | [R4h pin](docs/benchmarks/artifacts/locomo-mh-r4h-dev-1x30-20260815.md) · [Mem0 freeze](docs/benchmarks/artifacts/locomo-mem0-samepin-pr10-20260813.md) |
| OpMem | **13/13** | — | [R4h OpMem](docs/benchmarks/artifacts/opmem-mh-r4h-local-20260815.md) |
| Marketing vertical | **17/17** | — | [R4h marketing](docs/benchmarks/artifacts/marketing-mh-r4h-local-20260815.md) |

Cycle closeout: [docs/research/competitive/cycle-closeout.md](docs/research/competitive/cycle-closeout.md).
Next product step: [docs/research/sota-representation-path.md](docs/research/sota-representation-path.md) (R5 structured-first OD).

Docs: [verticalization model](docs/vertical/verticalization-model.md) · [research portal](docs/research/README.md) · [GTM roadmap](docs/vertical/go-to-market-roadmap.md) · [benchmarks](docs/benchmarks/README.md)

## What is shipped

- `POST /ingest` — deterministic sync (no LLM required)
- `POST /ingest/async` — worker extract when `BRAINY_PROVIDER_*` is set
- `GET /memories/search` and `POST /recall` (`context` / `enumerate` / `answer`)
- Postgres + local hash embeddings (optional provider embeddings / pgvector 768)
- Duplicate-ingest idempotency, correct / suppress / supersede
- Current-state projections (`view=current`), evidence plane (`BRAINY_EVIDENCE_STRICT`)
- Fenced worker leases (heartbeat + owner token)
- HTTP body cap (default 5 MiB) and server timeouts
- Optional `image_urls` OCR at WRITE (public http(s) only; Docker image has tesseract)

**Vertical strategy:** cognitive primitives + YAML packs — not per-vertical DB
kinds. First wedge: `packs/marketing/v1/pack.yaml`. See
[docs/vertical/verticalization-model.md](docs/vertical/verticalization-model.md).

## Repository Layout

- `archive/brainy-python-prototype/`: archived Python prototype and tests
- `cmd/api/`: Go API entrypoint
- `cmd/worker/`: Go worker entrypoint
- `internal/`: private Go application packages
- `packs/`: vertical pack definitions (YAML vocabulary, schemas, rank policy)
- `docs/`: rebuild docs, parity tracking, and cutover guidance
- `docs/vertical/`: verticalization model and marketing discovery
- `docs/research/`: representation path, cycle closeouts, competitive pins

## Mem0 Reference

The pinned Mem0 reference for this rebuild is tracked in [docs/mem0-parity-matrix.md](docs/mem0-parity-matrix.md).

Current pinned upstream commit:

- `a670333d67be1207b5be2fc73af60c3439444f48`

## Local Development

### Go tooling

Needs Postgres (`BRAINY_DATABASE_URL`; migrations apply on API startup).

```bash
go test ./...
go run ./cmd/api
BRAINY_WORKER_MODE=loop go run ./cmd/worker
```

When `BRAINY_API_KEYS` is set, unauthenticated requests return 401. For local
no-auth: `unset BRAINY_API_KEYS BRAINY_REQUIRE_API_KEY` and `export BRAINY_ENV=local`.
Otherwise send `Authorization: Bearer <key>` or `X-API-Key` (`tenant_id:key` pairs).

### Eval harness

Python evals under `evals/` are stdlib-only. They need a live API.

```bash
python3 evals/run_eval.py --base-url http://localhost:8080
python3 evals/run_vertical_eval.py --base-url http://localhost:8080
python3 evals/run_marketing_mvp_benchmark.py --base-url http://localhost:8080
```

Current parity fixtures live under `fixtures/parity/`. Marketing vertical fixtures: `fixtures/vertical/marketing/`.

For an operator-oriented local setup using an external Postgres instance, see [docs/external-postgres-runbook.md](docs/external-postgres-runbook.md).

**Vetting & GTM:** Marketing proof gates and paths to open source, published benchmarks, and commercial API — [docs/vertical/marketing-vetting-gate.md](docs/vertical/marketing-vetting-gate.md), [docs/vertical/go-to-market-roadmap.md](docs/vertical/go-to-market-roadmap.md), [docs/vertical/execution-plan.md](docs/vertical/execution-plan.md) (Linear ↔ GitHub sync).

### Docker (local / staging)

```bash
docker compose up --build
python3 evals/run_marketing_mvp_benchmark.py --base-url http://localhost:8080
```

### Environment

Copy or populate `.env.example` as needed.

## Execution Rules

- Preserve the archived prototype until the destructive-change gate passes.
- Keep Mem0 parity tracking explicit; do not use "Mem0-inspired" as a substitute for a pinned reference and documented deviations.
- Do not claim SOTA or beats-Mem0 without a frozen same-pin win on the axis you name. 1×30 is measurement.
- Human contributors: [CONTRIBUTING.md](CONTRIBUTING.md). Cloud agents: [AGENTS.md](AGENTS.md).
