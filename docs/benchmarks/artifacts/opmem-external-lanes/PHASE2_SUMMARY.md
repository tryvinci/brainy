# OpMem external lanes — execution summary

**Protocol:** 13 tasks / 47 steps (`fixtures/opmem/`).

## Full comparison

| Lane | Run ID | Score | Infra errors | SDK / infra |
| --- | --- | ---: | ---: | --- |
| mem0-oss | `opmem-phase2-full-mem0` | **9/13** | 0 | `mem0ai` 2.1.0, local Qdrant path |
| langmem | `opmem-phase2-full1` | **8/13** | 0 | Postgres + pgvector |
| letta | `opmem-phase2-letta-full` | **5/13** | 0 | `letta==0.11.7` |
| zep | `opmem-phase2-zep-full` | **3/13** | 0 | `zep-cloud==3.28.0`, free tier, \$0 paid |

## Zep (2026-09-23)

- `ZEP_API_KEY` present in `CLOUD_AGENT_INJECTED_SECRET_NAMES`; free-tier preflight OK (`zep-env-check-2026-09-23.json`).
- Adapter: `thread.add_messages` + recall from thread messages + `graph.search` (SDK 3.x; no legacy `memory` API).
- Forget: documented noop (no id-level delete in mapping).

```bash
python3 evals/run_opmem_external_live.py --smoke --lanes zep
python3 evals/run_opmem_external_live.py --full --lanes zep --run-id opmem-phase2-zep-full
```

## Spend

Payload worst-case OpenAI ~\$0.15 vs \$3 ceiling (`cost-worksheet-payload-budget.json`). Zep lane \$0 paid.
