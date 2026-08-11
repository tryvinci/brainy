# LME-20 product-recall — partial / aborted — 2026-08-11

**Harness:** `--publish --product-recall` on hardening build (local API+worker)  
**Run id:** `lme20-product-recall-20260811`  
**Queue precheck:** idle at start

## Proof observed

| Item | Type | Result | answer_path |
| --- | --- | --- | --- |
| 7a87bd0c | knowledge-update | WRONG | **`/recall`** |
| 00ca467f | multi-session | WRONG | **`/recall`** |
| 4388e9dd | single-session-assistant | **ABORT** | publish mode: extraction job `job_1786417826467002375` failed |

Product `/recall` path is fail-closed and recorded on scored items. Run did **not** complete 20/20 under publish mode.

**Not a publishable LME accuracy claim.** Retry isolated on an idle queue after worker/error root-cause check.
