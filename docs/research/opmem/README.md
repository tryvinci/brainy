# OpMem (operational memory correctness)

Paper-track benchmark for whether a memory system **behaves correctly** under forget, correct, isolate, and supersede — not whether it recalls facts on LoCoMo.

| Doc | Purpose |
| --- | --- |
| [opmem-spec.md](../opmem-spec.md) | Normative task schema, operations, scoring |
| [PUBLICATION_READINESS.md](./PUBLICATION_READINESS.md) | **Publication package** (protocol, claims, gaps, threats, artifacts, outline) |
| [paper/main.pdf](./paper/main.pdf) | Formatted manuscript (build: `make -C docs/research/opmem/paper pdf`) |
| [external-lanes-setup.md](./external-lanes-setup.md) | Phase 1 external lane adapters (dry-run / cost worksheet) |
| [posts/2026-07-opmem-v0.md](../posts/2026-07-opmem-v0.md) | July 2026 public draft (**12-task** numbers; superseded for pins) |
| [inventory-reproduction-plan-2026-09-18.md](../inventory-reproduction-plan-2026-09-18.md) | Cross-paper inventory — **PR [#188](https://github.com/tryvinci/brainy/pull/188)** until merged to `dev` |

**Code:** `fixtures/opmem/` (13 tasks) · `evals/run_opmem.py` · `evals/opmem_adapters.py` · CI `TestOpMemBenchmarkAgainstHTTPServer`

**Current merge-pin (embedded harness):** git SHA **`20533f2`** — Brainy **13/13**, verbatim **10/13**, Mem0 Platform **10/13** in [`opmem-pin-20260918.json`](../../benchmarks/artifacts/opmem-pin-20260918.json). Historical 2026-08-15 pin (SHA `1b5ab3e`): [PUBLICATION_READINESS.md](./PUBLICATION_READINESS.md#independent-evidence-review-132-vs-1013). Phase~2 external lanes: [PHASE2_SUMMARY.md](../../benchmarks/artifacts/opmem-external-lanes/PHASE2_SUMMARY.md) (pinned artifacts; no re-run required for doc edits).
