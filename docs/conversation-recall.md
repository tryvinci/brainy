# POST /recall — product synthesis surface (W4)

```http
POST /recall
Content-Type: application/json

{
  "tenant_id": "t1",
  "subject_id": "u1",
  "q": "What activities does Dana enjoy?",
  "mode": "enumerate",
  "top_k": 30,
  "budget_tokens": 4000
}
```

## Modes

| Mode | LLM? | Output |
| --- | --- | --- |
| `context` | no | `context_block` — deduped, token-budgeted memory lines |
| `enumerate` | no | `items[]` — distinct values + evidence memory IDs |
| `answer` | no* | Deterministic union/first-statement; `abstained` when empty |

\*Optional LLM answerer can be wired later via provider config; v1 is deterministic so every score is a product score.

## Eval thin client

Set `BRAINY_USE_RECALL=1` plus `BRAINY_RECALL_TENANT` / `BRAINY_RECALL_SUBJECT` to route
`evals/public/judge.py` through `/recall` instead of harness-side harvest.
