# OpMem non-reg — R1c local — 2026-08-14

**Commit:** `21a632b` (PR #113 merged to `dev` + `main`)
**Harness:** `python3 evals/run_opmem.py --systems brainy` (local API)
**API:** `/tmp/brainy-bin/api` rebuilt from this commit; `BRAINY_RECALL_LLM=1`

**Result:** **13/13 passed** (correction 3/3, isolation 3/3, suppression 3/3, staleness 3/3, idempotency 1/1)

Explicit `/correct` / `/suppress` paths still govern OpMem after coverage-gated fact-primary recall.
