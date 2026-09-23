# OpMem external lanes — execution summary

**Protocol:** 13 tasks / 47 steps (`fixtures/opmem/`).

## Full comparison

| Lane | Run ID | Score | Infra errors | SDK / infra |
| --- | --- | ---: | ---: | --- |
| mem0-oss | `opmem-phase2-full-mem0` | **9/13** | 0 | `mem0ai` 2.1.0, local Qdrant path |
| langmem | `opmem-phase2-full1` | **8/13** | 0 | Postgres + pgvector |
| letta | `opmem-phase2-letta-full` | **5/13** | 0 | `letta==0.11.7` |
| zep | `opmem-phase2-zep-full` | **3/9\*** | 0 | `zep-cloud==3.28.0`, free tier, $0 paid |

**\* Zep (publication line):** Score **3/9** counts only tasks where forget is in scope for the lane. **Not supported** (listed separately, not counted as failures): `iso03_forget_isolated`, `sup01_basic_forget`, `sup02_targeted_forget`, `sup03_durable_forget` — the `zep-cloud` 3.28 thread/graph adapter has no delete path for OpMem `forget`. **Raw harness** on the unchanged artifact `live-full-opmem-phase2-zep-full.json` remains **3/13**. See `zep-score-interpretation-opmem-phase2-zep-full.json`.

### Zep supported-task breakdown (`opmem-phase2-zep-full`)

| Status | Tasks |
| --- | --- |
| **Passed (3)** | `cor02_correction_stickiness`, `cor03_revised_retrievable`, `dup01_idempotent_remember` |
| **Failed, scored (6)** | `cor01_basic_revision`, `iso01_subject_isolation`, `iso02_tenant_isolation`, `upd01_stale_fact`, `upd02_preference_change`, `upd03_state_supersession` |
| **Not supported — forget (4)** | `iso03_forget_isolated`, `sup01_basic_forget`, `sup02_targeted_forget`, `sup03_durable_forget` |

## Spend

Payload worst-case OpenAI ~$0.15 vs $3 ceiling (`cost-worksheet-payload-budget.json`). Zep lane $0 paid.
