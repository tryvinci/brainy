# LOCOMO smoke — CF dense embeddings A/B (2026-07-23)

**Pins:** dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`,
async ingest, answerer/judge `gpt-oss-120b` via CF AI Gateway, top_k=15.
**Embedder:** `workers-ai/@cf/baai/bge-base-en-v1.5` (768-d) unless noted.

## Scores (1 conv / 30 Q)

| Config | Overall | Artifact |
| --- | ---: | --- |
| hash baseline (prior pin) | **13/30** | `locomo-smoke-entity-gated` |
| dense + entity ON (drained queue) | 11/30 | `locomo-smoke-cf-bge-base-drained` |
| dense + entity OFF | **13/30** | `locomo-smoke-cf-bge-base-entity-off` |

Do **not** cite the premature 14/30 (`locomo-smoke-cf-bge-base`) — QA started
before the async extract queue drained.

## Notes

- CF AI Gateway `/compat/embeddings` works with the `workers-ai/` model prefix.
- Dense embeddings alone are net-neutral under this judge; entity ranking still
  regresses. Default: entity ranking OFF (opt-in `BRAINY_ENTITY_RANKING=true`).
- Next for a Mem0-comparable claim: wire the same embed model on Render staging,
  then remeasure with a GPT-class judge (OpenAI wholesale credits currently 402
  on this gateway).
