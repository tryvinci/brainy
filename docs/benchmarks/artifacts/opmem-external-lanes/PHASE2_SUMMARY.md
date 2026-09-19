# OpMem external lanes — Phase 2 execution summary

**Run host:** cloud agent VM, 2026-09-19. **Protocol:** 13 tasks / 47 steps (`fixtures/opmem/`), identical harness (`evals/run_opmem.py`).

## Spend preflight (payload budget)

| Lane | Worst-case USD (1.10× retry margin) | Ceiling | Within? |
| --- | ---: | ---: | --- |
| mem0-oss | 0.114 | $1 | yes |
| letta | 0.014 | $1 | yes |
| langmem | 0.018 | $1 | yes |
| zep | 0 (Zep-native) | $0 | yes |
| **OpenAI total** | **0.145** | **$3** | yes |

Source: `cost-worksheet-payload-budget.json` (tiktoken on serialized bodies incl. mem0 `ADDITIVE_EXTRACTION_PROMPT`).  
Legacy `cost-worksheet-dry-run.json` mock meter (~409/66/191 tokens per lane) is **not** authoritative.

## Smoke (`opmem-phase2-smoke2`)

| Lane | Result |
| --- | --- |
| mem0-oss | pass `dup01_idempotent_remember` |
| langmem | pass `dup01_idempotent_remember` |
| zep | blocked — `ZEP_API_KEY` not in runtime |
| letta | blocked — `LETTA_APP_SERVER_TOKEN` not in runtime |

## Full comparison (one pass per valid lane)

| Lane | Run ID | Score | Infra errors | Notes |
| --- | --- | ---: | ---: | --- |
| mem0-oss | `opmem-phase2-full-mem0` | **9/13** | 0 | Local Qdrant path; `mem0ai` 2.1.0 |
| langmem | `opmem-phase2-full1` | **8/13** | 0 | `postgresql://…/opmem_langmem` + pgvector |
| zep | — | — | — | blocked |
| letta | — | — | — | blocked |

Invalid run (adapter bug, superseded for mem0-oss): `opmem-phase2-full1` mem0 lane had 3 infra errors on `revise` (`mem.update` API).

## Package pins (manifest)

See `run-manifest-opmem-phase2-full1.json` and `run-manifest-opmem-phase2-full-mem0.json`.

## Commands

```bash
python3 evals/run_opmem_cost_dry_run.py
export DATABASE_URL=postgresql://opmem:opmem@127.0.0.1:5432/opmem_langmem
export OPMEM_QDRANT_ROOT=/tmp/opmem-qdrant-live
python3 evals/run_opmem_external_live.py --smoke --lanes mem0-oss,langmem,zep,letta
python3 evals/run_opmem_external_live.py --full --lanes mem0-oss,langmem
```
