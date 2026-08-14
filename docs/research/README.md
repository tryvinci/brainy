# Brainy Research

**Strategy:** Surpass Mem0 and peers on **multiple axes** — operational correctness,
vertical governance, latency/cost, and (fairly measured) conversational recall — then
publish that evidence. Not LOCOMO parity alone. Not benchmax.

**Program of record (2026-08-11): [external-reviews/2026-08-11-competitive-architecture-verdict.md](./external-reviews/2026-08-11-competitive-architecture-verdict.md)** — competitive parity program (Mem0 recall + Graphiti relations + Brainy governed truth).  
**Competitive archaeology:** [competitive/README.md](./competitive/README.md) · [gap matrix](./competitive/competitive-gap-matrix.md) · **[cycle closeout (required)](./competitive/cycle-closeout.md)**  
**Prior PoR:** [sota-end-to-end-program.md](./sota-end-to-end-program.md) — still useful history; next sequence follows the competitive verdict.  
**External review intake:** [external-reviews/](./external-reviews/) · **hardening self-review prompt:** [2026-08-11-hardening-self-review-prompt.md](./external-reviews/2026-08-11-hardening-self-review-prompt.md)  
**External agent handoff (preferred): [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)** — architecture context; start from the competitive verdict.  
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
| **LOCOMO post-cutover staging** | **15/30** / **33/90** on `1f2f26f` (1×30 dipped vs Gate 0; 3×90 roughly flat, MH 22.2%) |
| **LOCOMO harden local** | **14/30** — honest dip vs Gate 0 after stricter hop-join |
| **LOCOMO Wave 1 local** | **14/30** on `a7a5184` (MH **3/10**, temporal **9/16**); not vs Gate 0; hybrid-reader confound vs local 6/30 |
| **LOCOMO R1c local** | **10/30** on `21a632b` (MH **2/10**, OD **0/4**, temporal **8/16**) — honest dip vs Wave 1; not a compiler coverage result |
| **LOCOMO compiler-quality local** | **11/30** on `d82f7d6` (MH **2/10**, OD **0/4**, temporal **9/16**) — q10 recovered; junk templates 45→6; still a dip vs Wave 1 |
| **Industry stand (same-pin)** | Trail Mem0 Platform LoCoMo **11/30 vs 12/30**; trail MH **2/10 vs 7/10**; trail OD **0/4 vs 3/4**; **lead** temporal **9/16 vs 2/16**. Lead OpMem 13/13 vs 9/12 and marketing 17/17 vs 4/16 (Mem0 ops/vertical pins not re-run this cycle). Graphiti/Zep: no pin. Full table: [cycle-closeout.md](./competitive/cycle-closeout.md) |
| **Next agent** | **R1b coverage**: durable claims still WRITE_MISS in transcripts (15/19 LoCoMo misses this run). Not ranking retune. [sota-representation-path.md](./sota-representation-path.md) |

Details: [competitive verdict](./external-reviews/2026-08-11-competitive-architecture-verdict.md) · [gap matrix](./competitive/competitive-gap-matrix.md) · [assessment pack](./external-agent-assessment-pack.md) · [program-execution-status.md](./program-execution-status.md)

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
| Competitive architecture verdict | **Accepted PoR (2026-08-11)** | [external-reviews/2026-08-11-competitive-architecture-verdict.md](./external-reviews/2026-08-11-competitive-architecture-verdict.md) |
| Representation path (execute now) | **Active** | [sota-representation-path.md](./sota-representation-path.md) |
| Representation-path review (2026-08-14) | **Accepted amendment** | [external-reviews/2026-08-14-representation-path-additions.md](./external-reviews/2026-08-14-representation-path-additions.md) |
| Competitive archaeology | Active | [competitive/README.md](./competitive/README.md) · [gap matrix](./competitive/competitive-gap-matrix.md) |
| Codebase graph (mermaid + JSON) | Active | [codebase-graph.md](./codebase-graph.md) · [codebase-graph.json](./codebase-graph.json) |
| SOTA end-to-end program (prior PoR) | Historical | [sota-end-to-end-program.md](./sota-end-to-end-program.md) |
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

1. **R1b** atomic compiler; held-out **semantic coverage** is the milestone (R0/R1a/R1c landed in #113).  
2. **R2–R5** canonical entities → relation **projection** → entity-ID hops → structured-first answers.  
3. Fair LoCoMo 3×90 + LME-20 quality under identical pins before any “beat Mem0” claim. Representation audit gates merge before those scores.  
4. Keep bounded episode fallback until R1b coverage is proven. Do not hard-drop episodes.  

Details: [sota-representation-path.md](./sota-representation-path.md)  
