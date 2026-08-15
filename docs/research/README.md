# Brainy Research

**Strategy:** Ship a governed vertical memory service that is operationally
correct, fast enough to self-host, and honest about conversational recall.
Not a single public-suite number. Not benchmax.

**Program of record (2026-08-11): [external-reviews/2026-08-11-competitive-architecture-verdict.md](./external-reviews/2026-08-11-competitive-architecture-verdict.md)** — compile interactions into facts/entities/relations; retrieve those; keep episodes as provenance.
**Internal competitive archaeology:** [competitive/README.md](./competitive/README.md) · **[cycle closeout (required)](./competitive/cycle-closeout.md)**
**Prior PoR:** [sota-end-to-end-program.md](./sota-end-to-end-program.md) — still useful history; next sequence follows the competitive verdict.
**External review intake:** [external-reviews/](./external-reviews/) · **hardening self-review prompt:** [2026-08-11-hardening-self-review-prompt.md](./external-reviews/2026-08-11-hardening-self-review-prompt.md)
**External agent handoff (preferred): [external-agent-assessment-pack.md](./external-agent-assessment-pack.md)** — architecture context; start from the competitive verdict.
**Codebase graph: [codebase-graph.md](./codebase-graph.md)** · machine-readable [codebase-graph.json](./codebase-graph.json).
**Prior master plan: [master-plan.md](./master-plan.md)** — still useful for W1–W7 history; new work follows the end-to-end program.
**Earlier external briefing: [sota-assessment-and-action-plan.md](./sota-assessment-and-action-plan.md)** — seeded the program; prefer the assessment pack for new agents.
History: [path-to-sota.md](./path-to-sota.md) · Papers: [paper-topics.md](./paper-topics.md) · Ladder: [public-bench-ladder.md](./public-bench-ladder.md)

Style: **reproducible, cited, honest about gaps**. The product README may carry a **published-%** table and a same-pin **summary** vs named systems (outlink [benchmarks](../benchmarks/README.md) · [published-claims](../benchmarks/published-claims.md)); GTM/launch stay Brainy-product copy. **Evals may name competitors.** See [AGENTS.md](../../AGENTS.md).

---

## Headline (today)

| Axis | Result |
| --- | --- |
| **OpMem** (operational) | **13/13** |
| **Marketing vertical** | **17/17** |
| **LOCOMO 1×30** | **21/30** (MH **10/10**, OD **0/4**, temporal **11/16**) — measurement, not qualification |
| **LongMemEval-20** | **0/20** integrity pin (not re-run this cycle) |
| **Next** | **R5** structured-first OD. [sota-representation-path.md](./sota-representation-path.md) |

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
| Public-bench ladder | L3 live; L4 gated | [public-bench-ladder.md](./public-bench-ladder.md) |
| Proveable eval framework | Spec + harness | [proveable-eval-framework.md](./proveable-eval-framework.md) · [`evals/public/`](../../evals/public/) |
| External agent assessment pack | **Preferred handoff** | [external-agent-assessment-pack.md](./external-agent-assessment-pack.md) |
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

1. **R5** structured-first OD (1×30 OD is still 0/4).
2. **R2–R4 remainder** canonical entities → relation **projection** → entity-ID hops, as needed for OD.
3. Fair LoCoMo 3×90 + LME-20 quality under identical pins before any SOTA claim. Representation audit gates merge before those scores.
4. Keep bounded episode fallback until coverage is proven. Do not hard-drop episodes.

Details: [sota-representation-path.md](./sota-representation-path.md)
