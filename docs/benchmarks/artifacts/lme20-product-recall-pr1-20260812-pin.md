# LME-20 product-recall — PR1 publishable pin — 2026-08-12

**Harness:** `--publish --product-recall --limit 20 --seed 1 --async-timeout 3600`  
**Run id:** `lme20-product-recall-pr1-20260812`  
**Commit:** `1225be5` (job-id dedupe + write-artifacts-before-fail-closed)  
**Worker:** local loop, concurrency 8; API+worker UTF-8 sanitize (`4e56229`)  
**Queue precheck:** idle  
**Dataset SHA:** `d6f21ea9d60a0d56f34a05b609c79c88a451d2ae03597821ea3d5a9678c3a442`

## Outcome — publishable (integrity)

Proveability **passed** (`LME_EXIT=0`):

| Gate | Value |
| --- | --- |
| `answer_path` | `/recall` on **20/20** |
| `ingest_mode` | async |
| `jobs_expected` | **4829** |
| `jobs_completed` | **4829** |
| `jobs_failed` | **0** |
| `queue_precheck` | idle |
| Judge temperature | 0.0 |

Item `6b168ec8` (attempt-2 off-by-one) now **240/240**. Item `4388e9dd` (UTF-8 abort) **240/240**. Item `c960da58` (attempt-1 timeout) **254/254**.

## Accuracy (honest — not a quality win)

**0/20 (0.000)** overall. All six question types 0. Attempt-2 log had 1/20 CORRECT on `6b168ec8` but was **not publishable**; this pin is the first proveable product-recall LME-20 and scores **0/20**. Do not treat 0/20 as a regression claim vs the blocked attempt-2 log.

Latency p50/p95 (answer step): ~214 / ~297 ms.

## Artifacts

- [report](./lme20-product-recall-pr1-20260812.md)
- Full json/manifest remain local under `docs/benchmarks/runs/` (gitignored; models redacted at write time).

## Follow-up

- LME-100 unblocked **as a measurement gate**; still do not run it as a quality claim until conversational recall lifts (PR2+).
- Competitive PR2: conversational append-only vs governed mutation.
