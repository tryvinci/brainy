# LME-20 — Recall Contract V3 — deferred note — 2026-08-10

**Status:** **Not a publishable pin.**

## What shipped for LME honesty

- Eval harness `wait_until_jobs_done` + publish fail-closed ([`evals/public/backends/brainy.py`](../../../evals/public/backends/brainy.py))
- Oversized payload chunking (default ≤48KB / soft 4 msgs)
- Worker `BRAINY_WORKER_CONCURRENCY` (local 4; staging render.yaml default 4)

## Why deferred this run

Attempted async LME-20 (`lme20-v3-20260810`) while LoCoMo multi-convo was also enqueueing. Queue peaked at **~320 pending** extraction jobs; LME competed with LoCoMo for the single worker pool. Cleared non-`locomo-*` pending jobs to finish multi-convo; LME-20 aborted intentionally rather than publish a partial score.

## Next

1. Merge PR #92 → `dev` (empty staging queue)
2. Run LME-20 alone with `--publish` once barrier + concurrency proven
3. Only then LME-100

Do **not** cite incomplete LME numbers as contract scores.
