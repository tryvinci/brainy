# OpMem external lanes — Phase 1 setup (no live vendor runs)

Phase 1 ships adapters, dry-run token instrumentation, and cost worksheets. **Do not** enable live comparisons until credentials exist and `cost-worksheet-dry-run.json` projects spend under owner ceilings.

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
| **letta** | `LETTA_APP_SERVER_TOKEN`; `OPENAI_API_KEY`; `LETTA_BASE_URL` (default `http://127.0.0.1:8283`) |
| **langmem** | `OPENAI_API_KEY`; `DATABASE_URL` (Postgres + pgvector); **do not** set `LANGMEM_STORE=InMemoryStore` for measured runs |

**Not used for mem0-oss lane:** `MEM0_API_KEY` (Platform).

Deterministic namespaces: set `OPMEM_RUN_ID` (e.g. `opmem-pin-20260919`) for reproducible IDs.

## Docker images (pin before live runs)

| Service | Image (pin in run manifest before live) | Port |
| --- | --- | ---: |
| Qdrant (mem0-oss) | `qdrant/qdrant:v1.12.5` | 6333 |
| Letta App Server | `letta/letta:latest` (replace with digest pin at live time) | 8283 |
| Postgres (langmem) | `pgvector/pgvector:pg17` | 5432 |

## Phase 1 commands (no paid vendor API traffic)

```bash
# Dry-run all four lanes over 13 fixtures + cost worksheet
python3 evals/run_opmem_cost_dry_run.py \
  --out-dir docs/benchmarks/artifacts/opmem-external-lanes

# Contract tests (mocked HTTP only)
python3 -m unittest evals/opmem_external/test_lane_contract.py -v
```

Live harness (blocked until worksheet + credentials):

```bash
export OPMEM_EXTERNAL_LIVE=1
python3 evals/run_opmem.py --systems mem0-oss,zep,letta,langmem ...
```

## Fairness rules

- Identical 13 tasks, step order, and substring scoring as `evals/run_opmem.py`.
- Memory mutations only in adapters; no answer-generation calls in OpMem scoring.
- `begin_task` resets lane namespace every task.
- Infrastructure errors remain **invalid/incomplete** runs (not comparable task scores).
