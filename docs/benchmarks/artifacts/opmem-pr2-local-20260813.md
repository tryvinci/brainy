# OpMem non-reg — PR2 local — 2026-08-13

**Branch:** `pr/competitive-program-adoption-a6c7` (`10a31e3`)  
**Harness:** `python3 evals/run_opmem.py --systems brainy` (local API)  
**API:** local `/tmp/brainy-bin/api` rebuilt from this commit  

**Result:** **13/13 passed** (correction 3/3, isolation 3/3, suppression 3/3, staleness 3/3, idempotency 1/1)

Explicit `/correct` / `/suppress` paths still govern OpMem; conversational append-only extract ops did not regress this suite.
