# Staging post-merge smoke — 2026-08-08

**Staging tip:** `db64d02` (multi-hop packet depth + `BRAINY_RECALL_LLM=1` in `render.yaml`)  
**API env:** `BRAINY_RECALL_LLM=1` + provider base/model/key set for hybrid reader  
**Worker:** provider extract enabled after env redeploy

## Checks

| Check | Result |
| --- | --- |
| `GET /healthz` | 200 |
| `GET /jobs/status?tenant_id=&subject_id=` | 200 (counts) |
| Sync `/ingest` + `POST /recall` multi-hop | 200; `second_pass` fired; coverage satisfied |
| Hybrid `reader_source=hybrid_llm_packet` | Pending env redeploy verification |

## Notes

- Async `/ingest/async` can lag when extraction_jobs backlog is large (prior LME-100 left pending rows). Prefer job wait via `/jobs/{id}` + `/jobs/status`.
- OpMem + marketing non-reg recorded separately under this date stamp.
