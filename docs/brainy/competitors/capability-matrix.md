# Cross-Competitor Capability Matrix

Legend:
- `Y` = explicit evidence in sources
- `P` = partial/indirect evidence
- `I` = inference only

| Capability | Mem0 | Supermemory | Zep | Letta | Memobase | Cognee |
|---|---|---|---|---|---|---|
| Managed cloud offering | Y | Y | Y | Y | Y | Y |
| Self-hosted option | Y | I | Y (Graphiti + BYOC) | Y | Y | Y |
| Explicit memory APIs | Y | Y | Y | Y | Y | Y |
| Graph-native layer | P (Graph memory plan mention) | Y (graph endpoints) | Y | P | I | Y |
| Temporal or fact invalidation semantics | I | I | Y | P | I | I |
| Profile/user-centric memory | P | Y | Y | Y | Y | P |
| Conversation ingestion primitives | P | Y | Y | Y | P | P |
| Memory deletion/forget controls | Y | Y | P | Y | P | Y |
| Enterprise controls (RBAC/SSO/audit/SLA) | Y | Y | Y | Y | I | P |
| Benchmark/eval narrative visible | P | P | P (latency marketing) | P | P | I |

## Evidence Notes
- Mem0: <https://docs.mem0.ai/introduction>, <https://docs.mem0.ai/openmemory/overview>, <https://mem0.ai/pricing>
- Supermemory: <https://supermemory.ai/docs/introduction>, <https://supermemory.ai/docs/api-reference>, <https://supermemory.ai/pricing>
- Zep: <https://help.getzep.com/overview>, <https://help.getzep.com/graphiti/getting-started/welcome>, <https://www.getzep.com/>, <https://www.getzep.com/pricing/>
- Letta: <https://docs.letta.com/guides/core-concepts/stateful-agents/>, <https://docs.letta.com/guides/core-concepts/memory/memory-blocks/>, <https://www.letta.com/pricing>, <https://docs.letta.com/guides/docker/>
- Memobase: <https://docs.memobase.io/introduction>
- Cognee: <https://docs.cognee.ai/>, <https://docs.cognee.ai/cognee-cloud/overview>, <https://docs.cognee.ai/api-reference/introduction>

## Phase Gate Check
Every row value is explicitly marked by evidence (`Y`/`P`) or inference (`I`).
