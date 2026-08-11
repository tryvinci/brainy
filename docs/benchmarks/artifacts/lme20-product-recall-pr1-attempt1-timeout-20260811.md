# LME-20 product-recall — attempt 1 timeout — 2026-08-11

**Harness:** `--publish --product-recall --limit 20 --seed 1`  
**Run id:** `lme20-product-recall-pr1-20260811`  
**Build:** store-boundary UTF-8 sanitize (`4e56229`) + harness job accounting  
**Worker:** concurrency 4; `--async-timeout` default **1800s**

## Outcome

**Not publishable.** Fail-closed abort on item **17/20** (`c960da58`, single-session-user):

```text
jobs not done within timeout for subject=c960da58
failed=[] failures=[]
```

Tracked remaining jobs completed in DB ~1–3 minutes after the harness deadline (worker healthy). This is a **barrier timeout**, not a UTF-8 / extraction hard failure.

## Progress before abort (honest partial — not a score claim)

| # | QID | Type | Jobs | Result | path |
| ---: | --- | --- | --- | --- | --- |
| 1 | 7a87bd0c | knowledge-update | 255/255 failed=0 | WRONG | /recall |
| 2 | 00ca467f | multi-session | 243/243 failed=0 | WRONG | /recall |
| 3 | **4388e9dd** | single-session-assistant | **240/240 failed=0** | WRONG | /recall |
| 4 | 6b168ec8 | single-session-user | 241 exp / 240 done failed=0 | CORRECT | /recall |
| 5–16 | (see run log) | mixed | failed=0 each | mostly WRONG | /recall |
| 17 | c960da58 | single-session-user | timeout | ABORT | — |

Item **4388e9dd** is the prior UTF-8 abort subject (`SQLSTATE 22021` / `0x80`). It **completed cleanly** under the store write-boundary sanitize — UTF-8 gate cleared.

## Follow-up

Attempt 2: same seed/limit with `--async-timeout 3600` and worker concurrency **8** (`run-id=lme20-product-recall-pr1-20260811b`).
