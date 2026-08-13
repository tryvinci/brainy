# OpMem non-reg — Wave 1 local — 2026-08-13

**Commit:** `a7a5184` (Wave 1 merged to `dev`)  
**Harness:** `python3 evals/run_opmem.py --systems brainy` (local API)  
**API:** `/tmp/brainy-bin/api` rebuilt from this commit; `BRAINY_RECALL_LLM=1`

**Result:** **13/13 passed** (correction 3/3, isolation 3/3, suppression 3/3, staleness 3/3, idempotency 1/1)

Explicit `/correct` / `/suppress` paths still govern OpMem after PR2 append-only conversational extract and Wave 1 retrieval/extract changes.
