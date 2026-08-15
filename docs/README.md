# Documentation

This is the map for humans using or extending Brainy. The product Quick Start
lives in the root [README](../README.md).

## Start here

| Doc | What it covers |
| --- | --- |
| [README](../README.md) | What Brainy is, Compose quickstart, API table, benchmark summary |
| [api.md](./api.md) | HTTP routes, auth, request bodies |
| [conversation-ingest.md](./conversation-ingest.md) | How chat clients should call `/ingest` |
| [benchmarks/README.md](./benchmarks/README.md) | Full benchmarks: same-pin vs Mem0 / Graphiti / Zep, harness, reproduce |
| [../evals/README.md](../evals/README.md) | Fixture harnesses; Mem0 Platform runners |
| [external-postgres-runbook.md](./external-postgres-runbook.md) | Run API + worker on your own Postgres |

## Product

| Doc | What it covers |
| --- | --- |
| [vertical/verticalization-model.md](./vertical/verticalization-model.md) | Packs vs schemas; primitives |
| [packs/marketing/v1/pack.yaml](../packs/marketing/v1/pack.yaml) | First vertical pack |
| [commercial-beta-checklist.md](./commercial-beta-checklist.md) | Self-host vs hosted-beta gaps |

## Contribute

| Doc | What it covers |
| --- | --- |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Setup, tests, PRs against `dev` |
| [SUPPORT.md](../SUPPORT.md) | Where to ask for help |
| [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md) | Community standards |
| [SECURITY.md](../SECURITY.md) | Vulnerability reporting |
| [AGENTS.md](../AGENTS.md) | Notes for automated coding agents |

## Research notes

Long-form internal notes live under [research/](./research/README.md). They are
not the getting-started path.

## Layout

- `cmd/api`, `cmd/worker` — HTTP API and extract worker
- `internal/` — Go packages
- `packs/` — YAML vertical packs (first: marketing)
- `fixtures/` — eval goldens
- `evals/` — HTTP harnesses against a live API
