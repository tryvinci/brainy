# Brainy Research

**Strategy:** Ship a governed vertical memory service that is operationally
correct, fast enough to self-host, and honest about conversational recall.
Not a single public-suite number. Not benchmax.

**Next-agent start (2026-08-21):** [handover-sota-agent-2026-08-21.md](./handover-sota-agent-2026-08-21.md) — live pins, code map, kill list, and the proof-path increment. The 2026-08-17 assessment pack is architecture context; its “next is R5A / R6a compiler” is **not** live.

**Program of record (2026-08-11): [external-reviews/2026-08-11-competitive-architecture-verdict.md](./external-reviews/2026-08-11-competitive-architecture-verdict.md)** — compile interactions into facts/entities/relations; retrieve those; keep episodes as provenance.
**Internal competitive archaeology:** [competitive/README.md](./competitive/README.md) · **[cycle closeout (required)](./competitive/cycle-closeout.md)**
**Prior PoR:** [sota-end-to-end-program.md](./sota-end-to-end-program.md) — still useful history; next sequence follows the competitive verdict.
**External review intake:** [external-reviews/](./external-reviews/) · **this-pass verdict (live):** [2026-08-17-parity-gap-verdict.md](./external-reviews/2026-08-17-parity-gap-verdict.md) · prompt: [2026-08-17-full-recall-self-review-prompt.md](./external-reviews/2026-08-17-full-recall-self-review-prompt.md) · historical Wave 1: [2026-08-17-competitive-archaeology-verdict.md](./external-reviews/2026-08-17-competitive-archaeology-verdict.md) · historical hardening prompt: [2026-08-11-hardening-self-review-prompt.md](./external-reviews/2026-08-11-hardening-self-review-prompt.md)
**External agent handoff (preferred): [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)** — start from **CURRENT 2026-08-17**; architecture context below that.
**Codebase graph: [codebase-graph.md](./codebase-graph.md)** · machine-readable [codebase-graph.json](./codebase-graph.json).
**Prior master plan: [master-plan.md](./master-plan.md)** — still useful for W1–W7 history; new work follows the end-to-end program.
**Earlier external briefing: [sota-assessment-and-action-plan.md](./sota-assessment-and-action-plan.md)** — seeded the program; prefer the assessment pack for new agents.
History: [path-to-sota.md](./path-to-sota.md) · Papers: [paper-topics.md](./paper-topics.md) · Ladder: [public-bench-ladder.md](./public-bench-ladder.md)

Style: **reproducible, cited, honest about gaps**. The product README may carry a **published-%** table and a same-pin **summary** vs named systems (outlink [benchmarks](../benchmarks/README.md) · [published-claims](../benchmarks/published-claims.md)); GTM/launch stay Brainy-product copy. **Evals may name competitors.** See [AGENTS.md](../../AGENTS.md).

---

## Headline (today)

Live start + pins: [handover-sota-agent-2026-08-21.md](./handover-sota-agent-2026-08-21.md). R5A–R10 substrate is merged. S0 ledger is **PROOF**, not WRITE.

| Axis | Result |
| --- | --- |
| **OpMem** (operational) | **13/13** vs Mem0 **10/13** this freeze |
| **Marketing vertical** | **17/17** vs Mem0 **4/17** empirical |
| **LOCOMO 1×30** | **21/30** (MH **10/10**, OD **0/4**, temporal **11/16**) vs Mem0 **11/30** — measurement, not qualification. Do not overwrite. |
| **LOCOMO S0 n=180** | product `/recall` **32/180** · industry **62/180** · ledger PROOF 112 / RETRIEVAL 22 / READER 11 / WRITE **3** |
| **LOCOMO S0 MH** | product **2/33** after packet-proof (was 1/33). Coverage 32/33. |
| **LOCOMO full n=1540** | **175/1540 (11.4%)** product `/recall` — **dip** vs July search+harness **49.8%**; not vs Mem0 **92.5%** on the same path. [why](../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md) |
| **LongMemEval-20** | **4/20** product `/recall` (lift vs 0/20 integrity; not vs published 94.4%). LME-500 not run |
| **BEAM 100K** | **8/20** search+harness (non-reg; 1M/10M not run) |
| **Next** | Remasure MH-only 33 on integrity (`#135` named-community / during-clause proof). Do not merge #133. Do not re-queue R0–R10. [handover](./handover-sota-agent-2026-08-21.md) · [cycle-closeout](./competitive/cycle-closeout.md) |

Details: [competitive verdict](./external-reviews/2026-08-11-competitive-architecture-verdict.md) · [assessment pack](./external-agent-assessment-pack.md) · [program-execution-status.md](./program-execution-status.md)

---

## Published / live artifacts

| Piece | Status | Link |
| --- | --- | --- |
| Recall-contract proof | Active | [recall-contract-proof-20260807.md](../benchmarks/artifacts/recall-contract-proof-20260807.md) |
| OpMem v0 | Live results | [spec](./opmem-spec.md) · [fresh pin](../benchmarks/artifacts/opmem-fresh-local-20260815.md) |
| Marketing vertical moat | Live results | [moat report](../benchmarks/marketing-moat-report.md) |
| Launch narrative | Draft | [launch narrative](../benchmarks/launch-narrative.md) |
| LOCOMO 1×30 | Live measurement | [fresh pin](../benchmarks/artifacts/locomo-fresh-1x30-20260815.md) |
| LOCOMO full n=1540 | Live measurement + dip why | [fresh full](../benchmarks/artifacts/locomo-fresh-full-20260815.md) · [why 11.4%](../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md) |
| Full-recall self-review prompt | Historical (answered) | [2026-08-17 prompt](./external-reviews/2026-08-17-full-recall-self-review-prompt.md) |
| Parity-gap verdict | **This pass received (live)** | [2026-08-17 verdict](./external-reviews/2026-08-17-parity-gap-verdict.md) · [source](./external-reviews/2026-08-17-parity-gap-review.md) |
| Archaeology verdict | Historical (Wave 1 `bd987fa`) | [2026-08-17 archaeology](./external-reviews/2026-08-17-competitive-archaeology-verdict.md) |
| LongMemEval-20 | Live measurement | [fresh pin](../benchmarks/artifacts/lme20-fresh-20260815.md) |
| BEAM 100K | Live measurement | [fresh pin](../benchmarks/artifacts/beam-100k-fresh-20260815.md) |
| Public-bench ladder | L3 live; L4 gated | [public-bench-ladder.md](./public-bench-ladder.md) |
| Proveable eval framework | Spec + harness | [proveable-eval-framework.md](./proveable-eval-framework.md) · [`evals/public/`](../../evals/public/) |
| Next-agent SOTA/80% handover | **Live start (2026-08-21)** | [handover-sota-agent-2026-08-21.md](./handover-sota-agent-2026-08-21.md) |
| External agent assessment pack | Architecture context (2026-08-17 pins) | [external-agent-assessment-pack.md](./external-agent-assessment-pack.md) |
| Competitive architecture verdict | **Accepted PoR (2026-08-11)** | [external-reviews/2026-08-11-competitive-architecture-verdict.md](./external-reviews/2026-08-11-competitive-architecture-verdict.md) |
| Representation path (execute now) | **Active** | [sota-representation-path.md](./sota-representation-path.md) |
| Representation-path review (2026-08-14) | **Accepted amendment** | [external-reviews/2026-08-14-representation-path-additions.md](./external-reviews/2026-08-14-representation-path-additions.md) |
| Competitive archaeology (internal) | Active | [competitive/README.md](./competitive/README.md) |
| Codebase graph (mermaid + JSON) | Active | [codebase-graph.md](./codebase-graph.md) · [codebase-graph.json](./codebase-graph.json) |
| SOTA end-to-end program (prior PoR) | Historical | [sota-end-to-end-program.md](./sota-end-to-end-program.md) |
| SOTA assessment + action plan (earlier briefing) | Superseded by pack + PoR | [sota-assessment-and-action-plan.md](./sota-assessment-and-action-plan.md) |
| Paper roadmap | Active | [paper-topics.md](./paper-topics.md) |
| OpMem Paper 1 draft | Draft | [posts/2026-07-opmem-v0.md](./posts/2026-07-opmem-v0.md) |
| Gap crush checklist | **100%** | [gap-crush-checklist.md](./gap-crush-checklist.md) |

---

## External datasets we cite

- **LOCOMO** — [snap-research/locomo](https://github.com/snap-research/locomo) · [ACL 2024](https://aclanthology.org/2024.acl-long.747/)
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

Empty / TBD cells beat invented numbers. Public posts describe Brainy; they do not name competitor products.

---

## Next engineering step

R5A–R10 are **merged**. S0 WRITE_MISS is 3/180. Do not start with compiler mass.

1. Remaining MH list/join **packet/proof** (shared facts that are not only a `both X and Y` cue; enumerate lists still crowd). [handover](./handover-sota-agent-2026-08-21.md).
2. Remeasure MH-only 33 fail-closed on a frozen integrity tenant before burning n=180 or n=1540.
3. Only after proof moves: S6 freeze + Mem0 same-pin. No SOTA / beats-Mem0 in product copy until that win and explicit approval.

Details: [sota-execution-plan.md](./sota-execution-plan.md) (histogram outranks the written S1-first order) · [locomo-full-70-80-path.md](./locomo-full-70-80-path.md)
