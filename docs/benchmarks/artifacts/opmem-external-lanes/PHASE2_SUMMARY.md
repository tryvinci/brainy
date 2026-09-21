# OpMem external lanes — execution summary

**Protocol:** 13 tasks / 47 steps (`fixtures/opmem/`), harness `evals/run_opmem.py` / `run_opmem_external_live.py`.

## Full comparison (valid lanes)

| Lane | Run ID | Score | Infra errors | Notes |
| --- | --- | ---: | ---: | --- |
| mem0-oss | `opmem-phase2-full-mem0` | **9/13** | 0 | `mem0ai` 2.1.0, local Qdrant path |
| langmem | `opmem-phase2-full1` | **8/13** | 0 | Postgres + pgvector |
| letta | `opmem-phase2-letta-full` | **5/13** | 0 | `letta==0.11.7`, `opmem_facts` block mapping |
| zep | — | — | — | **blocked** — no `ZEP_API_KEY` on host (`live-zep-blocked-opmem-phase2-zep-smoke.json`) |

## Letta reproduction

```bash
pip install -r evals/opmem_external/requirements-external-lanes.txt
export OPENAI_API_KEY=...
scripts/opmem-start-letta-server.sh
export LETTA_BASE_URL=http://127.0.0.1:8283
python3 evals/run_opmem_external_live.py --full --lanes letta --run-id opmem-phase2-letta-full
```

## Zep reproduction (when key is injected)

```bash
export ZEP_API_KEY=...   # free tier; hard-stop on billing prompts
python3 evals/run_opmem_external_live.py --smoke --lanes zep
python3 evals/run_opmem_external_live.py --full --lanes zep --run-id opmem-phase2-zep-full
```

## Spend preflight

See `cost-worksheet-payload-budget.json` (worst-case OpenAI ~$0.15 vs $3 ceiling).
