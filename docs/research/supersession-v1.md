# Temporal supersession v1 (ENG-86)

**Status:** Implemented on `dev` path (schema + API)  
**Doctrine:** Product knowledge-update correctness — not LOCOMO-specific.

## Model

| Field | Meaning |
| --- | --- |
| `supersedes_id` | On the **new** record: ID of the memory it replaces |
| `superseded_at` | On the **old** record: when lifecycle became `superseded` |
| `lifecycle_state=superseded` | Excluded from default search (with archived/suppressed) |

## APIs

```http
POST /memories/{id}/supersede?tenant_id=&subject_id=
{"content":"Door code is 2222","source_text":"..."}
```

Creates a new active memory with `supersedes_id={id}`, marks `{id}` superseded.
Returns the new memory (OpMem `revise` still uses in-place `/correct`).

```http
POST /events
{
  "tenant_id":"t1","subject_id":"u1",
  "event_type":"campaign_ended",
  "supersede_memory_ids":["mem_…","mem_…"],
  "match": {"label":"promo","metadata":{"season":"summer"}}
}
```

Batch invalidation: explicit IDs and/or **match** by label/kind/metadata (v2).
Pack YAML lifecycle rules still set archived/suppressed at ingest time.

```http
GET /memories/search?...&q=...&include_historical=1
```

Opt-in: include `lifecycle=superseded` rows (still excludes archived/suppressed).

### Ingest lineage

```json
{"metadata":{"supersedes_memory_id":"mem_old"}, "messages":[...]}
```

After upsert, the prior memory is marked superseded; new row stores `supersedes_id`.

## Migration

Postgres migration **v11** (`add_supersession_lineage`).

## Out of scope (later)

- Automatic contradiction detection / LLM merge
- Pack-triggered invalidation without explicit IDs
- Bi-temporal valid_from / valid_to (ENG-59 model decision)
