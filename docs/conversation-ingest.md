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

`metadata.session_id` and `metadata.observed_at` are copied onto each memory (clients supply event time — see ENG-174 for search-time use).

## Anti-pattern

Do not rely on eval-only hacks (stuffing the full transcript at query time). Improve product extract/rank so live clients get the same recall quality.
