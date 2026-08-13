# LME-20 product-recall — PR1 attempt 2 — 2026-08-11

**Harness:** `--publish --product-recall --limit 20 --seed 1 --async-timeout 3600`  
**Run id:** `lme20-product-recall-pr1-20260811b`  
**Build:** store UTF-8 sanitize (`4e56229`) + local worker concurrency 8  
**Queue precheck:** idle  

## Outcome

**Not publishable.** All **20/20** items scored with `answer_path=/recall` and `jobs_failed=0`, but proveability fail-closed:

```text
extras.jobs_completed (4829) must equal jobs_expected (4830)
```

Root cause: item `6b168ec8` reported `jobs_expected=241 completed=240` — idempotent `/ingest/async` returned a duplicate `job_id` that the harness counted twice in `jobs_expected`. Harness fix: dedupe pending job IDs in `BrainyBackend` (append + `wait_until_jobs_done`).

Observed (log-derived; **not a publishable accuracy claim**): **1/20 CORRECT** (only `6b168ec8`), all others WRONG, all `path=/recall`.

## Integrity gates cleared

| Gate | Result |
| --- | --- |
| UTF-8 abort subject `4388e9dd` | **240/240 failed=0** (prior SQLSTATE 22021) |
| Prior timeout subject `c960da58` | **254/254 failed=0** (with 3600s async timeout) |
| Product path | **20/20** `answer_path=/recall` |
| Extraction hard failures | **jobs_failed=0** across run |
| Proveability job equality | **Blocked** (off-by-one) — intentional fail-closed |

## Attempt 1 cross-ref

[lme20-product-recall-pr1-attempt1-timeout-20260811.md](./lme20-product-recall-pr1-attempt1-timeout-20260811.md) — aborted item 17 @ 1800s; UTF-8 already cleared.

## Follow-up

- Re-run isolated LME-20 `--publish --product-recall` after job-id dedupe for a **publishable** pin.  
- LME-100 remains gated until a clean publishable LME-20.
