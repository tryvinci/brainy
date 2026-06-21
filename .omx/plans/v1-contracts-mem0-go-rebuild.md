# V1 Contracts: Mem0-Informed Go Rebuild

## Purpose

Freeze the first thin execution slice so implementation starts from explicit contracts instead of drifting from a broad PRD.

This artifact is intentionally narrower than the full product plan. It defines the minimum credible path that must work before async workers, network-backed providers, or wider feature parity are allowed to expand scope.

## First-Slice Scope

The first executable slice includes only:

- `POST /ingest`
- `GET /memories/search`
- deterministic local extraction
- Postgres persistence
- duplicate-ingest idempotency
- correction or suppression for one memory class
- no external model, embedding, or reranker dependency
- no async worker requirement for correctness

Out of scope for this slice:

- async extraction workers
- provider-backed extraction
- provider-backed embeddings
- multimodal inputs
- broad multi-tenant admin concerns
- full Mem0 API compatibility

## Canonical Memory Taxonomy

Only these memory kinds are valid in the first slice:

1. `profile`
   - stable user facts or descriptors
   - example: `User works at Acme`
2. `preference`
   - likes, dislikes, style choices, or recurring preferences
   - example: `User prefers concise answers`
3. `fact`
   - contextual factual data that may be useful for later retrieval
   - example: `Launch date is May 12`

The implementation must reject or downgrade any extraction that does not fit one of these classes.

## Canonical Memory Record Schema

Required fields:

- `memory_id`
- `tenant_id`
- `subject_id`
- `kind`
- `content`
- `source_text`
- `source_type`
- `dedupe_key`
- `status`
- `created_at`
- `updated_at`

Optional first-slice fields:

- `confidence`
- `metadata`
- `extraction_version`
- `explain`

Disallowed in the first slice:

- provider-specific fields in the core domain model
- vector-only records without a canonical text memory row

## Ingest Contract

### Request

```json
{
  "tenant_id": "t_123",
  "subject_id": "u_123",
  "messages": [
    {
      "role": "user",
      "content": "I prefer concise, direct answers."
    }
  ],
  "source_type": "conversation"
}
```

### Response

```json
{
  "ingest_id": "ing_123",
  "accepted": true,
  "created": 1,
  "updated": 0,
  "deduped": 0,
  "memories": [
    {
      "memory_id": "mem_123",
      "kind": "preference",
      "content": "Prefers concise, direct answers",
      "status": "active"
    }
  ]
}
```

### Rules

- identical logical memory ingests must be idempotent
- ingest must succeed without network access
- local deterministic extraction must be sufficient for correctness

## Search Contract

### Request

`GET /memories/search?tenant_id=t_123&subject_id=u_123&q=How should I respond?`

### Response

```json
{
  "results": [
    {
      "memory_id": "mem_123",
      "kind": "preference",
      "content": "Prefers concise, direct answers",
      "score": 0.92,
      "explain": {
        "matched_terms": ["respond"],
        "ranking_basis": "deterministic_baseline"
      }
    }
  ]
}
```

### Rules

- response ordering must be deterministic for fixed fixtures
- retrieval must return explain/debug payloads in the first slice
- first slice may use metadata/full-text ranking before embeddings are introduced

## Correction / Suppression Semantics

Allowed first-slice mutation:

- mark one memory as suppressed
- replace one memory with corrected content

Required guarantee:

- a suppressed or corrected memory must affect subsequent search results deterministically

## First-Slice Storage Decision

Postgres is mandatory for the first slice.

`pgvector` is optional in the first slice and may be deferred if:

- deterministic full-text or metadata ranking satisfies the golden fixtures
- the API contract does not promise semantic-vector retrieval yet

If embeddings are deferred, the plan must record that as an intentional deviation in the parity matrix.

## Mem0 Reference Pinning

Before execution starts:

- pin one specific Mem0 upstream commit for reference
- capture the exact public examples or fixtures being used for parity comparison
- record intentional deviations from that reference in `docs/mem0-parity-matrix.md`

## Exit Criteria For This Contract

This contract is satisfied only when:

1. deterministic ingest -> persist -> search passes in CI
2. duplicate ingest is idempotent
3. one correction or suppression flow updates later search results
4. no network dependency is required for correctness
5. explain/debug payloads are returned for search results

## Post-Thin-Slice Extension (Vertical Packs)

The frozen first slice uses `kind: profile|preference|fact` only. Vertical expansion adds cognitive **primitives** and **YAML packs** — not new `kind` enum values per domain.

- Architecture: `docs/vertical/verticalization-model.md`
- First pack: `packs/marketing/v1/pack.yaml`
- v1 `kind` remains for Mem0-compat; packs map labels → `primitive` + legacy `kind`

This is an intentional deviation from Mem0; see `docs/mem0-parity-matrix.md`.
