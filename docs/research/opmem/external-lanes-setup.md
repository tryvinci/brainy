# OpMem external lanes — setup

Phase 1 ships adapters and dry-run operation counts. **Spend approval** uses the payload budget worksheet (`cost-worksheet-payload-budget.json`), not the legacy mock token meter in `cost-worksheet-dry-run.json`. Phase 2 live runs: `evals/run_opmem_external_live.py`.

## Spend authority (USD)

| Lane | Ceiling | Notes |
| --- | ---: | --- |
| `mem0-oss` | $1 | Local Qdrant + OpenAI for extract/embed |
| `letta` | $1 | Self-hosted App Server + OpenAI |
| `langmem` | $1 | Postgres/pgvector + OpenAI |
| `zep` | **$0 paid** | Zep Cloud free tier only; stop if credits exhausted |
| **Total** | **$3** | Sum of OpenAI-metered lanes |

Rates used in worksheets (OpenAI list, per 1M tokens): `gpt-4.1-mini-2025-04-14` input $0.40 / output $1.60; `text-embedding-3-small` $0.02.

## Pinned models

- Chat: `gpt-4.1-mini-2025-04-14`
- Embed: `text-embedding-3-small`

If a lane cannot accept these pins, record the product constraint in the worksheet and **do not** substitute silently (Zep is documented as Zep-native only).

## Environment variables (never commit values)

| Lane | Variables |
| --- | --- |
| **mem0-oss** | `OPENAI_API_KEY` (vault); `QDRANT_URL` (default `http://127.0.0.1:6333`); optional `MEM0_OSS_CONFIG_PATH` |
| **zep** | `ZEP_API_KEY`; optional `ZEP_API_URL` (default `https://api.getzep.com`) |
| **letta** | `OPENAI_API_KEY`; `LETTA_BASE_URL` (default `http://127.0.0.1:8283`); optional `LETTA_APP_SERVER_TOKEN` (required for non-local servers) |
| **langmem** | `OPENAI_API_KEY`; `DATABASE_URL` (Postgres + pgvector); **do not** set `LANGMEM_STORE=InMemoryStore` for measured runs |

**Not used for mem0-oss lane:** `MEM0_API_KEY` (Platform).

Deterministic namespaces: set `OPMEM_RUN_ID` (e.g. `opmem-pin-20260919`) for reproducible IDs.

## Pinned Python packages

Install once: `pip install -r evals/opmem_external/requirements-external-lanes.txt`

| Component | Pin |
| --- | --- |
| Letta server | `letta==0.11.7` + `sqlite-vec` |
| Letta client | `letta-client==0.1.307` |
| Zep SDK | `zep-cloud==3.28.0` |

Start local Letta (no Docker): `scripts/opmem-start-letta-server.sh`

## Docker images (optional alternative)

| Service | Image | Port |
| --- | --- | ---: |
| Qdrant (mem0-oss) | `qdrant/qdrant:v1.12.5` | 6333 |
| Letta | prefer pip pin above; or `letta/letta` digest at live time | 8283 |
| Postgres (langmem) | `pgvector/pgvector:pg17` | 5432 |

## Phase 1 commands (no paid vendor API traffic)

```bash
# Dry-run all four lanes over 13 fixtures + cost worksheet
python3 evals/run_opmem_cost_dry_run.py \
  --out-dir docs/benchmarks/artifacts/opmem-external-lanes

# Contract tests (mocked HTTP only)
python3 -m unittest evals/opmem_external/test_lane_contract.py -v
```

Live harness:

```bash
pip install -r evals/opmem_external/requirements-external-lanes.txt
scripts/opmem-start-letta-server.sh   # letta lane
export DATABASE_URL=postgresql://opmem:opmem@127.0.0.1:5432/opmem_langmem
export ZEP_API_KEY=...               # free tier only
python3 evals/run_opmem_external_live.py --smoke --lanes mem0-oss,zep,letta,langmem
python3 evals/run_opmem_external_live.py --full --lanes mem0-oss,zep,letta,langmem
```

## Fairness rules

- Identical 13 tasks, step order, and substring scoring as `evals/run_opmem.py`.
- Memory mutations only in adapters; no answer-generation calls in OpMem scoring.
- `begin_task` resets lane namespace every task.
- Infrastructure errors remain **invalid/incomplete** runs (not comparable task scores).
