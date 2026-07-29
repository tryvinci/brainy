# Brainy Research

**Strategy:** Surpass Mem0 and peers on **multiple axes** — operational correctness,
vertical governance, latency/cost, and (fairly measured) conversational recall — then
publish that evidence. Not LOCOMO parity alone. Not benchmax.

Master plan: [path-to-sota.md](./path-to-sota.md) · Papers: [paper-topics.md](./paper-topics.md) · Ladder: [public-bench-ladder.md](./public-bench-ladder.md)

Style targets: [SuperMemory Research](https://supermemory.ai/research/) and Mem0’s LOCOMO writeups — **reproducible, cited, honest about gaps**.

---

## Headline (today)

| Axis | Result |
| --- | --- |
| **OpMem** (operational) | Brainy **12/12** vs Mem0 **9/12** on staging |
| **Marketing vertical** | **16/16** (Mem0: no equivalent track) |
| **LOCOMO smoke** (1×30, gpt-oss, dense emb) | **16/30** — mid score, multi-hop **3/10**; not a SOTA claim |
| **Search latency** | p50 ~730–890 ms after content-dense ranking (↓ vs ~1.0s) |

Details: [locomo-smoke.md](../benchmarks/locomo-smoke.md) · [staging competitive](../benchmarks/staging-competitive-report.md) · [moat](../benchmarks/marketing-moat-report.md)

---

## Published / live artifacts

| Piece | Status | Link |
| --- | --- | --- |
| OpMem v0 | Live results | [spec](./opmem-spec.md) · [staging vs Mem0](../benchmarks/staging-competitive-report.md) |
| Marketing vertical moat | Live results | [moat report](../benchmarks/marketing-moat-report.md) |
| Launch narrative | Draft | [launch narrative](../benchmarks/launch-narrative.md) |
| LOCOMO smoke (L3) | Live mid score | [locomo-smoke.md](../benchmarks/locomo-smoke.md) |
| Public-bench ladder | L3 live; L4 gated | [public-bench-ladder.md](./public-bench-ladder.md) |
| Proveable eval framework | Spec + harness | [proveable-eval-framework.md](./proveable-eval-framework.md) · [`evals/public/`](../../evals/public/) |
| Surpass plan | Active | [path-to-sota.md](./path-to-sota.md) |
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
