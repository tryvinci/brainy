# OpMem remasure — fresh local — 2026-08-15

**Commit:** `1b5ab3e` (current `origin/dev` / `origin/main` at remasure start)
**Harness:** `python3 evals/run_opmem.py --systems verbatim,brainy` (local API)
**API / worker:** `/tmp/brainy-bin/api-fresh` + `worker-fresh` rebuilt from this commit; dedicated DB `brainy_bench`; `BRAINY_RECALL_LLM=1`; `BRAINY_USE_RECALL=1`; async worker concurrency 8.

**Result:** Brainy **13/13 (100%) passed** (correction 3/3, isolation 3/3, suppression 3/3, staleness 3/3, idempotency 1/1). Verbatim baseline 10/13.

`upd01` June vs May kept. Ops lead vs Mem0 **10/13 (76.9%)** this cycle ([opmem-mem0-fresh-20260815.md](./opmem-mem0-fresh-20260815.md)).
