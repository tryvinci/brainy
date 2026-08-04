# Conversational memory ingest

How chat / agent clients should use `/ingest` so Brainy retains long-dialogue facts.

## Recommended pattern

Send **one utterance per `messages[]` entry** (speaker label optional but helpful):

```json
{
  "tenant_id": "acme",
  "subject_id": "user-42",
  "source_type": "conversation",
  "metadata": {
    "session_id": "sess-2023-05-07",
    "observed_at": "2023-05-07T18:00:00Z"
  },
  "messages": [
    {"role": "user", "content": "Caroline: I went to the LGBTQ support group on 7 May 2023"},
    {"role": "user", "content": "Melanie: That sounds important — how was it?"}
  ]
}
```

You may batch many messages in one request for throughput. Prefer **not** gluing multiple turns into a single string with only `\n` separators if you can use `messages[]` instead — but newline-separated turns in one content blob are also atomized.

## What Brainy stores

1. Structured prefs / profile / facts when keyword rules match  
2. Otherwise a **conversation episode** (`kind=fact`, `primitive=episode`) so casual dialogue is still searchable  

`metadata.session_id` and `metadata.observed_at` are copied onto each memory. `observed_at` is also stored as a typed event-time column and used for search recency (falling back to `updated_at` when unset).

## Async provider extract

**Production conversational clients should prefer `POST /ingest/async`.** Sync `/ingest` stays deterministic (no network) for CI and offline. The worker enables OpenAI-compatible provider extract via:

- `BRAINY_PROVIDER_BASE_URL` (or `LLM_BASE_URL`)
- `BRAINY_PROVIDER_API_KEY` (or `LLM_API_KEY`)
- `BRAINY_PROVIDER_MODEL` (or `LLM_MODEL`)

When configured, the worker runs deterministic baseline extract first, then provider extract, and **merges** both (provider structured facts + conversational episodes that are not exact duplicates). Provider failures fail the extraction job **before** upserts; the raw ingest payload is never rewritten.

## Embeddings

Hybrid search uses an `Embedder`:

- Default / CI: deterministic local hash embedder
- Optional: OpenAI-compatible `BRAINY_EMBEDDING_MODEL` (+ base URL/key, defaulting to provider/LLM settings)

Provider embedding failures soft-degrade to the local embedder. pgvector `embedding_vec` is `vector(768)` for the pinned hosted model (`bge-base-en-v1.5`); the 128-d local hash path keeps float[] only.


Local Docker Compose passes provider vars through from the host environment. Staging Blueprint (`render.yaml`) declares the same keys as Dashboard secrets on `brainy-worker-staging`.

```bash
curl -s -X POST http://127.0.0.1:8080/ingest/async \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":"acme","subject_id":"user-42","source_type":"conversation","metadata":{"observed_at":"2023-05-07T18:00:00Z"},"messages":[{"role":"user","content":"Caroline: I went to the LGBTQ support group on 7 May 2023"}]}'
# then search once the worker completes the job
```

## Anti-pattern

Do not rely on eval-only hacks (stuffing the full transcript at query time). Improve product extract/rank so live clients get the same recall quality.
