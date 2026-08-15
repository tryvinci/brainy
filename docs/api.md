# HTTP API

Brainy is a JSON HTTP service (`cmd/api`). Default listen address `:8080`.
Request bodies are capped at **5 MiB** unless `BRAINY_MAX_BODY_BYTES` is set.

Errors look like:

```json
{"error":{"code":"invalid_json","message":"invalid json body"}}
```

Common codes: `invalid_json`, `bad_request`, `method_not_allowed`, `not_found`,
`unauthorized`, `forbidden`, `conflict`, `request_too_large`.

## Auth

Local Compose (`BRAINY_ENV=local`) does not require a key.

When `BRAINY_API_KEYS` is set (or `BRAINY_ENV=production` /
`BRAINY_REQUIRE_API_KEY=true`), every route except `/healthz` needs:

- `Authorization: Bearer <key>`, or
- `X-API-Key: <key>`

Keys are `tenant_id:secret` pairs, comma-separated. `*` matches any tenant. A
request `tenant_id` that does not match the key’s tenant is rejected (`403`).

## Routes

| Method | Path | Body / query |
| --- | --- | --- |
| `GET` | `/healthz` | — (plain `ok`) |
| `GET` | `/metrics` | Prometheus text |
| `POST` | `/ingest` | [Ingest](#ingest) |
| `POST` | `/ingest/async` | Same body; `202` + `job_id` |
| `GET` | `/memories/search` | `tenant_id`, `subject_id`, `q`; optional `vertical`, `scope`, `limit`, `include_historical` |
| `POST` | `/recall` | [Recall](#recall) |
| `POST` | `/memories/{id}/correct?tenant_id=&subject_id=` | `{"content":"...", "source_text":"..."}` |
| `POST` | `/memories/{id}/suppress?tenant_id=&subject_id=` | empty |
| `POST` | `/memories/{id}/supersede?tenant_id=&subject_id=` | `{"content":"...", "source_text":"..."}` |
| `POST` | `/events` | Domain event (batch supersede / match) |
| `GET` | `/jobs/status` | Queue snapshot |
| `GET` | `/jobs/{id}` | One async job |

Conversation clients: [conversation-ingest.md](./conversation-ingest.md). Prefer
`POST /ingest/async` in production when a worker + provider is configured.

## Ingest

```json
{
  "tenant_id": "demo",
  "subject_id": "user-1",
  "source_type": "conversation",
  "vertical": "marketing",
  "metadata": {
    "session_id": "sess-1",
    "observed_at": "2026-08-15T12:00:00Z"
  },
  "messages": [
    {
      "role": "user",
      "content": "We never use exclamation marks in brand copy.",
      "image_urls": ["https://example.com/cover.jpg"]
    }
  ]
}
```

`image_urls` must be public `http(s)` (no loopback/private hosts). OCR runs when
`tesseract` is on `PATH` (included in the Docker image).

Sync `/ingest` is deterministic (no LLM). `/ingest/async` enqueues a job; poll
`/jobs/{id}` or run `cmd/worker` with `BRAINY_WORKER_MODE=loop`.

## Recall

```json
{
  "tenant_id": "demo",
  "subject_id": "user-1",
  "q": "brand voice rules",
  "mode": "enumerate",
  "view": "current",
  "top_k": 30
}
```

`mode`: `context` | `enumerate` | `answer`. `view`: `current` | `historical` | `all`.

## Environment

See [`.env.example`](../.env.example). Notable knobs:

| Variable | Default | Meaning |
| --- | --- | --- |
| `BRAINY_ENV` | — | `local` skips required API keys |
| `BRAINY_HTTP_ADDR` | `:8080` | Listen address |
| `BRAINY_DATABASE_URL` | local Postgres DSN | Migrations apply on API start |
| `BRAINY_API_KEYS` | empty | `tenant:key,...` |
| `BRAINY_WORKER_MODE` | `once` | `loop` for Compose worker |
| `BRAINY_PROVIDER_*` | empty | OpenAI-compatible extract |
| `BRAINY_EMBEDDING_*` | empty | OpenAI-compatible embeddings; else local hash |
| `BRAINY_MAX_BODY_BYTES` | 5 MiB | Request body cap |
| `BRAINY_EVIDENCE_STRICT` | `false` | Fail ingest if raw evidence cannot be written |
