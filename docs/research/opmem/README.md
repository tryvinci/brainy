# OpMem (operational memory correctness)

Paper-track benchmark for whether a memory system **behaves correctly** under forget, correct, isolate, and supersede — not whether it recalls facts on LoCoMo.

| Doc | Purpose |
| --- | --- |
| [opmem-spec.md](../opmem-spec.md) | Normative task schema, operations, scoring |
| [PUBLICATION_READINESS.md](./PUBLICATION_READINESS.md) | **Publication package** (protocol, claims, gaps, threats, artifacts, outline) |
| [paper/main.pdf](./paper/main.pdf) | Formatted manuscript (build: `make -C docs/research/opmem/paper pdf`) |
| [posts/2026-07-opmem-v0.md](../posts/2026-07-opmem-v0.md) | July 2026 public draft (**12-task** numbers; superseded for pins) |
| [inventory-reproduction-plan-2026-09-18.md](../inventory-reproduction-plan-2026-09-18.md) | Cross-paper inventory — **PR [#188](https://github.com/tryvinci/brainy/pull/188)** until merged to `dev` |

**Code:** `fixtures/opmem/` (13 tasks) · `evals/run_opmem.py` · `evals/opmem_adapters.py` · CI `TestOpMemBenchmarkAgainstHTTPServer`

**Current empirical pin (not re-run on `bbe55f7`):** Brainy **13/13** vs Mem0 Platform **10/13** (2026-08-15, SHA `1b5ab3e`). See [PUBLICATION_READINESS.md](./PUBLICATION_READINESS.md#independent-evidence-review-132-vs-1013).
