# LME-20 abort root cause — 2026-08-11

**Job:** `job_1786417826467002375`  
**Status:** failed after 3/3 attempts → dead-lettered  
**Subject item:** LME question `4388e9dd` (single-session-assistant) during `--publish --product-recall`

## Failure reason (DB)

```text
ERROR: invalid byte sequence for encoding "UTF8": 0x80 (SQLSTATE 22021)
```

Postgres rejected a TEXT insert during async extraction persist. The harness previously aborted with job IDs only (`publish mode: extraction jobs failed: ['job_…']`), hiding this reason.

## Fix (PR1)

1. Harness: `wait_until_jobs_done` surfaces `failure_reason` + returns `jobs_expected/completed/failed` accounting; proveability requires `jobs_failed==0` and `completed==expected`.
2. Product: sanitize invalid UTF-8 on ingest normalize + `BuildMemoryRecord` (+ worker claim path) via `strings.ToValidUTF8`.

## Follow-up

Isolated LME-20 `--publish --product-recall` rerun after rebuild; pin publishable score or explicit failure artifact.

**Update (same day):** store write-boundary sanitize + attempt 2 (`lme20-product-recall-pr1-20260811b`) cleared UTF-8 subject `4388e9dd` and completed 20/20 `/recall` scores, but proveability blocked on jobs_expected/completed off-by-one (duplicate async `job_id`). See [lme20-product-recall-pr1-attempt2-20260811.md](./lme20-product-recall-pr1-attempt2-20260811.md).
