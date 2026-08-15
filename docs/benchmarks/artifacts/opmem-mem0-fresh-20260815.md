# OpMem — Mem0 Platform counter-run — 2026-08-15

**Harness:** `python3 evals/run_opmem.py --systems mem0`
**System:** Mem0 Platform API (not Mem0 OSS)
**Same 13 fixtures** as Brainy [opmem-fresh-local-20260815.md](./opmem-fresh-local-20260815.md)

**Result:** **10/13 (76.9%)**

| Category | Mem0 | Brainy this cycle |
| --- | ---: | ---: |
| correction | 2/3 | 3/3 |
| isolation | 3/3 | 3/3 |
| suppression | 2/3 | 3/3 |
| staleness | 2/3 | 3/3 |
| idempotency | 1/1 | 1/1 |
| **overall** | **10/13** | **13/13** |

Mem0 fails: `cor02` (top does not contain `ruby`), `sup03` (forget still returns 1), `upd02` (top does not contain `sms`).

Prior ops pin was **9/12 (75.0%)** on 2026-07-14 (12-task set). This is a **new** 13-task empirical pin. Brainy leads ops on this remasure. Do not treat Platform 10/13 as an OSS-reproducible number.
