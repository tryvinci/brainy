# Brainy Research

**Strategy:** Surpass Mem0 and peers on **multiple axes** — operational correctness,
vertical governance, latency/cost, and (fairly measured) conversational recall — then
publish that evidence. Not LOCOMO parity alone. Not benchmax.

**Program of record (2026-08): [sota-end-to-end-program.md](./sota-end-to-end-program.md)** — evidence/event/bitemporal architecture, phased execution, SOTA qualification gates.  
**External review intake:** [external-reviews/](./external-reviews/) · **self-review prompt:** [2026-08-11-hardening-self-review-prompt.md](./external-reviews/2026-08-11-hardening-self-review-prompt.md)  
**External agent handoff (preferred): [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)** — architecture context; start the current pass from the self-review prompt.  
**Codebase graph: [codebase-graph.md](./codebase-graph.md)** · machine-readable [codebase-graph.json](./codebase-graph.json).  
**Prior master plan: [master-plan.md](./master-plan.md)** — still useful for W1–W7 history; new work follows the end-to-end program.  
**Earlier external briefing: [sota-assessment-and-action-plan.md](./sota-assessment-and-action-plan.md)** — seeded the program; prefer the assessment pack for new agents.  
History: [path-to-sota.md](./path-to-sota.md) · Papers: [paper-topics.md](./paper-topics.md) · Ladder: [public-bench-ladder.md](./public-bench-ladder.md)

Style targets: [SuperMemory Research](https://supermemory.ai/research/) and Mem0’s LOCOMO writeups — **reproducible, cited, honest about gaps**.

---

## Headline (today)

| Axis | Result |
| --- | --- |
| **OpMem** (operational) | Brainy **13/13** on staging (Gate 0 + post-cutover `1f2f26f`) |
| **Marketing / Support vertical** | Marketing **passed** post-cutover; support prior **4/4** |
| **LOCOMO Gate 0 staging** | **18/30** (MH 50%, OD 25%); 3×90 **32/90** (MH 19.4%) |
| **LOCOMO harden local** | **14/30** — honest dip vs Gate 0 after stricter hop-join |
| **Industry stand** | Lead ops + marketing; conversational improving but not SOTA — [competitive positioning](../benchmarks/competitive-positioning-20260806.md) |
| **Next agent** | Clean LME-20 product-recall publish → finish staging LoCoMo re-pin → Mem0 same-pin → external self-review |

Details: [2026-08-11 self-review prompt](./external-reviews/2026-08-11-hardening-self-review-prompt.md) · [assessment pack](./external-agent-assessment-pack.md) · [program-execution-status.md](./program-execution-status.md) · [competitive positioning](../benchmarks/competitive-positioning-20260806.md)

---

## Published / live artifacts

| Piece | Status | Link |
| --- | --- | --- |
| Recall-contract proof (LoCoMo same-pin) | Active | [recall-contract-proof-20260807.md](../benchmarks/artifacts/recall-contract-proof-20260807.md) |
| Competitive positioning (industry stand) | Active | [competitive-positioning-20260806.md](../benchmarks/competitive-positioning-20260806.md) |
| OpMem v0 | Live results | [spec](./opmem-spec.md) · [staging vs Mem0](../benchmarks/staging-competitive-report.md) |
| Marketing vertical moat | Live results | [moat report](../benchmarks/marketing-moat-report.md) |
| Launch narrative | Draft | [launch narrative](../benchmarks/launch-narrative.md) |
| LOCOMO smoke (L3) | Live mid score | [locomo-smoke.md](../benchmarks/locomo-smoke.md) |
| Public-bench ladder | L3 live; L4 gated | [public-bench-ladder.md](./public-bench-ladder.md) |
| Proveable eval framework | Spec + harness | [proveable-eval-framework.md](./proveable-eval-framework.md) · [`evals/public/`](../../evals/public/) |
| Surpass plan | Active | [path-to-sota.md](./path-to-sota.md) |
| External agent assessment pack | **Preferred handoff** | [external-agent-assessment-pack.md](./external-agent-assessment-pack.md) |
| Codebase graph (mermaid + JSON) | Active | [codebase-graph.md](./codebase-graph.md) · [codebase-graph.json](./codebase-graph.json) |
| SOTA end-to-end program (PoR) | Active | [sota-end-to-end-program.md](./sota-end-to-end-program.md) |
| SOTA assessment + action plan (earlier briefing) | Superseded by pack + PoR | [sota-assessment-and-action-plan.md](./sota-assessment-and-action-plan.md) |
| Mem0 parity gaps | Active | [mem0-parity-gaps.md](./mem0-parity-gaps.md) |
| Paper roadmap | Active | [paper-topics.md](./paper-topics.md) |
| OpMem Paper 1 draft | Draft | [posts/2026-07-opmem-v0.md](./posts/2026-07-opmem-v0.md) |
| Gap crush checklist | **100%** | [gap-crush-checklist.md](./gap-crush-checklist.md) |

---

## External datasets we cite

- **LOCOMO** — [snap-research/locomo](https://github.com/snap-research/locomo) · [ACL 2024](https://aclanthology.org/2024.acl-long.747/)
- **Memory harness** — [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks)
- **Industry research pages** — [SuperMemory research](https://supermemory.ai/research/)

---

## Content format for public posts

1. One-line headline with primary metric  
2. Systems + version pins  
3. Table: accuracy · p50/p95 · tokens/query  
4. Category breakdown  
5. Methodology (dataset, judge, temperature, retrieval policy)  
6. Reproduce CLI  
7. Limitations — what we did *not* claim  

Empty / TBD cells beat invented numbers.

---

## Next engineering step

1. ~~ENG-86 temporal supersession v1~~ — [supersession-v1.md](./supersession-v1.md)  
2. Multi-hop product path (graph re-tune off 30-Q smoke)  
3. Paper 1 (OpMem) multi-system matrix + public post  
4. Fair full LOCOMO under identical pins before any “beat Mem0 on LOCOMO” claim  
