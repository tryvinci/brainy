# Brainy

[![Test](https://github.com/tryvinci/brainy/actions/workflows/test.yml/badge.svg?branch=dev)](https://github.com/tryvinci/brainy/actions/workflows/test.yml)
[![Docker smoke](https://github.com/tryvinci/brainy/actions/workflows/docker-smoke.yml/badge.svg?branch=dev)](https://github.com/tryvinci/brainy/actions/workflows/docker-smoke.yml)
[![Go](https://img.shields.io/github/go-mod/go-version/tryvinci/brainy)](https://github.com/tryvinci/brainy/blob/dev/go.mod)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

Self-hosted **memory for agents**: ingest conversations, persist facts, search,
and recall — with current-state, corrections, and YAML vertical packs.

Marketing is the first pack. The runtime is domain-agnostic: verticals are
configuration, not forked schemas.

## Why Brainy

Most memory layers stop at “embed it and hope.” Brainy is an HTTP service with
an async worker and Postgres, built so memories can be **governed**:

- **Write** — sync ingest (offline, deterministic) or async extract (optional LLM)
- **Read** — lexical + hybrid search, plus `/recall` (`context` / `enumerate` / `answer`)
- **Govern** — correct, suppress, supersede; current-state views; tenant isolation
- **Specialize** — packs under `packs/` (vocabulary, rank policy, fixtures)

## Quick start

```bash
git clone https://github.com/tryvinci/brainy.git && cd brainy
docker compose up --build -d
curl -s http://localhost:8080/healthz
```

Compose starts Postgres, the API, and the worker. `BRAINY_ENV=local` means no
API key. The image includes `tesseract-ocr` for optional `image_urls` on ingest.

```bash
curl -s -X POST http://localhost:8080/ingest \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":"demo","subject_id":"user-1","source_type":"conversation","messages":[{"role":"user","content":"We never use exclamation marks in brand copy."}]}'

curl -s 'http://localhost:8080/memories/search?tenant_id=demo&subject_id=user-1&q=brand+voice+rules'

curl -s -X POST http://localhost:8080/recall \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":"demo","subject_id":"user-1","q":"brand voice rules","mode":"enumerate"}'
```

Async ingest (worker is already looping):

```bash
curl -s -X POST http://localhost:8080/ingest/async \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":"demo","subject_id":"user-1","source_type":"conversation","messages":[{"role":"user","content":"Ship the autumn campaign on 12 September."}]}'
```

## HTTP API

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/healthz` | Liveness |
| `POST` | `/ingest` | Sync ingest (no LLM required) |
| `POST` | `/ingest/async` | Queue extract for the worker |
| `GET` | `/memories/search` | Search |
| `POST` | `/recall` | Recall (`context` / `enumerate` / `answer`) |
| `POST` | `/memories/{id}/correct` | In-place correction |
| `POST` | `/memories/{id}/suppress` | Hide from default search |
| `POST` | `/memories/{id}/supersede` | Replace with lineage |
| `POST` | `/events` | Batch domain events |
| `GET` | `/jobs/status`, `/jobs/{id}` | Async job status |
| `GET` | `/metrics` | Prometheus text |

Full request shapes: [docs/api.md](docs/api.md). Conversation pattern (including
`image_urls`): [docs/conversation-ingest.md](docs/conversation-ingest.md).

When `BRAINY_API_KEYS` is set, send `Authorization: Bearer <key>` or `X-API-Key`.
Local Compose does not require a key.

## Documentation

Start at **[docs/README.md](docs/README.md)**.

| I want to… | Go here |
| --- | --- |
| Run against my own Postgres | [docs/external-postgres-runbook.md](docs/external-postgres-runbook.md) |
| Write a vertical pack | [docs/vertical/verticalization-model.md](docs/vertical/verticalization-model.md) |
| Benchmarks vs other memory systems | [docs/benchmarks/README.md](docs/benchmarks/README.md) |
| Run evals | [evals/README.md](evals/README.md) |
| Contribute or get support | [CONTRIBUTING.md](CONTRIBUTING.md) · [SUPPORT.md](SUPPORT.md) |
| Report a vulnerability | [SECURITY.md](SECURITY.md) |

## Benchmarks

**As of 2026-08-16 remasure** on product SHA `1b5ab3e`. Vendors publish a
**percent per suite** (LoCoMo / LongMemEval / BEAM). That is the industry
scoreboard. **Not SOTA.** Cells below are **sourced claims**, not a same-pin
bake-off (n, top-k, judge, and sometimes the metric differ).

Full sourced table:
**[docs/benchmarks/published-claims.md](docs/benchmarks/published-claims.md)**.
Harness peer: [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks).

| | LoCoMo | LongMemEval | BEAM 1M | BEAM 10M |
| --- | ---: | ---: | ---: | ---: |
| **Brainy** | **11.4%** full `/recall` (n=1540) · 70.0% 1×30 | **20%** LME-20 `/recall` (4/20) · **4%** (n=100 hist.) | not run (40% on 100K/20q this cycle) | not run |
| **Mem0 Platform** | **92.5%** | **94.4%** | **64.1** | **48.6** |
| **Zep** | **75.14%** | 71.2% | — | — |
| **SuperMemory** | 77.1% | 95% Recall@15 | — | — |
| **Letta** | 74.0% | — | — | — |
| **Hindsight** | 92.0% | 94.6% | 73.9% | 64.1% |

Brainy’s **11.4%** is this cycle’s **full** LoCoMo on product `/recall` (n=1540).
The 2026-07-31 **49.8%** was search + harness answerer on an older stack — do not
mix. Vendor 90+ percents are also n=1540-class **LLM-over-search** (Mem0 top-k 200),
not Brainy `/recall`. Why the dip: [docs/benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md](docs/benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md).
SuperMemory’s 95 is **Recall@15**, not LLM-judge. Graphiti OSS has no
published %. Zep LoCoMo was disputed (Mem0 re-ran Zep at 58.44%; Zep publishes
75.14%).

### Same-pin

Same dataset SHA, judge temperature, and question set. We **trail
open-domain** on this pin.

| Suite | Brainy (`1b5ab3e`, 2026-08-15) | Mem0 Platform | Graphiti OSS / Zep |
| --- | ---: | ---: | --- |
| LoCoMo 1×30 | **70.0%** (21/30; MH 10/10, OD **0/4**, temporal 11/16) | **36.7%** (11/30; MH 6/10, OD 3/4, temporal 2/16) this cycle | no same-pin |
| OpMem | **100%** (13/13) | **76.9%** (10/13) this cycle | no pin |
| Marketing vertical | **100%** (17/17) | **23.5%** (4/17 empirical) this cycle | no pin |
| LongMemEval-20 | **20%** (4/20) product `/recall` | no fair pin on this harness | no pin |
| BEAM 100K | **40%** (8/20) search+harness | see published % above | no pin |

Skip-ingest stratified 180 on this VM (not public LoCoMo, not this 1×30 pin):
**151/180** product `/recall` (MH 28/33, OD **4/11** trail, SH 87/98,
temporal 32/38). Not 90% (162/180 on that sample). Details:
[docs/benchmarks/README.md](docs/benchmarks/README.md).

Mem0 **OSS** was not re-measured. Reproduce and artifacts:
**[docs/benchmarks](docs/benchmarks/README.md)**. Pins:
[LoCoMo](docs/benchmarks/artifacts/locomo-fresh-1x30-20260815.md) ·
[Mem0 same-pin](docs/benchmarks/artifacts/locomo-mem0-fresh-1x30-20260815.md) ·
[OpMem](docs/benchmarks/artifacts/opmem-fresh-local-20260815.md) ·
[marketing](docs/benchmarks/artifacts/marketing-fresh-local-20260815.md) ·
[LME-20](docs/benchmarks/artifacts/lme20-fresh-20260815.md) ·
[BEAM 100K](docs/benchmarks/artifacts/beam-100k-fresh-20260815.md).

## Status

Developer preview (`v0.1.0`). Self-host with Docker Compose. Hosted beta needs
API keys and still lacks ToS / privacy / PITR before GA.

## Development

Requires Go 1.25+ and Postgres (`BRAINY_DATABASE_URL`). Migrations apply on API
startup (`pg_trgm` required; `pgvector` optional).

```bash
go test ./...
go run ./cmd/api
BRAINY_WORKER_MODE=loop go run ./cmd/worker
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for tests, branch policy, and PR
expectations. [AGENTS.md](AGENTS.md) is for automated coding agents.

## Community

- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Contributing](CONTRIBUTING.md)
- [Support](SUPPORT.md)
- [Security policy](SECURITY.md)
- [Citation](CITATION.cff)

Bugs and features go through GitHub issue templates. Do not file public issues
for undisclosed vulnerabilities.

## License

[Apache License 2.0](LICENSE).
