# OpMem non-reg — R4 local — 2026-08-15

**Commit:** `d48e202` (R4 hops + leftover compiler coverage)
**Harness:** `python3 evals/run_opmem.py --systems brainy` (local API)
**API:** `/tmp/brainy-bin/api` rebuilt from this commit; `BRAINY_RECALL_LLM=1`

**Result:** **13/13 (100%) passed** (correction 3/3, isolation 3/3, suppression 3/3, staleness 3/3, idempotency 1/1)

`upd01` June vs May kept. Ops lead vs Mem0 **9/12 (75.0%)** (2026-07-14 Platform; **not re-run this cycle**) is unchanged and stale on the Mem0 side.
