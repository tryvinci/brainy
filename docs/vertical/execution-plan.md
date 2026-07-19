# Execution Plan — Marketing Vetting & GTM (Linear ↔ GitHub sync)

**Status:** Active (2026-07-04)  
**Repo:** [tryvinci/brainy](https://github.com/tryvinci/brainy) (`dev`)  
**Linear project:** [Memory System (Brainy)](https://linear.app/engramhq/project/memory-system-brainy-3c71cdab5cd1)  
**Linear doc:** [Marketing Vetting & GTM Execution Plan](https://linear.app/engramhq/document/marketing-vetting-and-gtm-execution-plan-177a372fc2eb)  
**Policy:** [`marketing-vetting-gate.md`](./marketing-vetting-gate.md) · [`go-to-market-roadmap.md`](./go-to-market-roadmap.md) · [`../research/proveable-eval-framework.md`](../research/proveable-eval-framework.md)

---

## Gate status

| Gate | Name | Status | GitHub milestone | Linear milestone |
| --- | --- | --- | --- | --- |
| **M1** | Deterministic marketing MVP | **Done** | — | Gate M1: Marketing Deterministic MVP |
| **M2** | Publish & OSS preview | **Done** (PR #13 merged) | [#1](https://github.com/tryvinci/brainy/milestone/1) | Gate M2: Publish & OSS Preview |
| **M3** | Marketing technical proof | **Done** (PR #14 merged) | [#2](https://github.com/tryvinci/brainy/milestone/2) | Gate M3: Marketing Technical Proof |
| **M4** | Finance / second vertical | **Unblocked** (research) | — | — |
| **M5** | Commercial API beta | **Done** (Track C) | [#3](https://github.com/tryvinci/brainy/milestone/3) | Gate M5: Commercial API Beta |

---

## Launch tracks (sequential — after M3)

Public launch proceeds as **three sequential tracks**, not parallel workstreams. Each track must complete before the next starts.

```
M3 Done ──► Track A (OSS preview) ──► Track B (benchmark launch) ──► Track C (hosted beta)
```

| Track | Name | Start when | Done when | Status |
| --- | --- | --- | --- | --- |
| **A** | OSS developer preview | M3 signed off | `v0.1.0` on `main`, README quickstart, Docker green | **Done** |
| **B** | Benchmark-led launch | Track A tagged | OpMem 12/12 published, moat report + methodology public, launch content | **Done** |
| **C** | Hosted API beta | Track B published | GH [#11](https://github.com/tryvinci/brainy/issues/11) auth + [#12](https://github.com/tryvinci/brainy/issues/12) commercial checklist | **Done** |

| Linear | Notes |
| --- | --- |
| [ENG-102](https://linear.app/engramhq/issue/ENG-102) | Track A — README quickstart + docs sync |
| [ENG-103](https://linear.app/engramhq/issue/ENG-103) | Track A — merge `dev` → `main`, tag `v0.1.0` |
| [ENG-104](https://linear.app/engramhq/issue/ENG-104) | Track B — OpMem 12/12 publish + launch narrative |

### Track B — after v0.1.0

| Step | Notes |
| --- | --- |
| Verify OpMem 12/12 on staging (`evals/run_opmem.py`) | PR #17 fixes `sup03`, `upd02` |
| Update `docs/benchmarks/opmem-baseline-report.md` | Publish 12/12 score |
| Public benchmark narrative (blog / landing) | Mem0/SuperMemory-style launch |

### Track C — after benchmark launch

| Linear / GitHub | Title |
| --- | --- |
| — | [#11](https://github.com/tryvinci/brainy/issues/11) API key auth per tenant |
| — | [#12](https://github.com/tryvinci/brainy/issues/12) Commercial beta checklist |

Finance (Gate M4) remains **research-only** and is not on the launch critical path.

---

## Track D — Public proveable eval (LOCOMO ladder)

Post-launch research track. **OpMem stays the CI merge gate**; LOCOMO is publication, not blocking merge.

| Layer | Deliverable | Status | Linear |
| --- | --- | --- | --- |
| L0–L1 | Own suites + research portal | Done | — |
| L2 | `evals/public/` UnifiedResult + Brainy adapter + pins | **Done** | [ENG-163](https://linear.app/engramhq/issue/ENG-163) |
| L3 | LOCOMO smoke (subset) + honest report | **Done** (7/30 remeasure 2026-07-19) | [ENG-164](https://linear.app/engramhq/issue/ENG-164) |
| L4 | Full LOCOMO + latency/tokens + blog | **Unblocked** (smoke 13/30 ≥12/30) | [ENG-165](https://linear.app/engramhq/issue/ENG-165) |
| L5 | LongMemEval / BEAM subset | Backlog | [ENG-166](https://linear.app/engramhq/issue/ENG-166) |
| L6 | MarketingMem public track | Backlog | [ENG-167](https://linear.app/engramhq/issue/ENG-167) |

Epic: [ENG-162](https://linear.app/engramhq/issue/ENG-162) · Milestone: [Public Proveable Eval Framework](https://linear.app/engramhq/project/memory-system-brainy-3c71cdab5cd1)

Docs: [`../research/public-bench-ladder.md`](../research/public-bench-ladder.md) · [`../research/proveable-eval-framework.md`](../research/proveable-eval-framework.md) · [`../../evals/public/`](../../evals/public/)

**Anti-benchmax:** LOCOMO fails drive **product** work ([ENG-168](https://linear.app/engramhq/issue/ENG-168)), not harness games. L3 smoke **13/30** (2026-07-19) after ranking/answerer loop; L4 unlock gate cleared. Continue full LOCOMO (ENG-165) + embeddings when provider available.

Reproduce:

```bash
cd evals && python -m public.locomo.run_smoke --conversations 1 --questions 30
```

---

## Sync index (Linear ↔ GitHub)

### Gate M1 — Done

| Linear | GitHub | Title |
| --- | --- | --- |
| [ENG-58](https://linear.app/engramhq/issue/ENG-58) | — | Marketing first vertical wedge |
| [ENG-61](https://linear.app/engramhq/issue/ENG-61) | — | Primitives + YAML packs |
| [ENG-80](https://linear.app/engramhq/issue/ENG-80) | — | Marketing use case map |
| [ENG-81](https://linear.app/engramhq/issue/ENG-81) | — | Brand voice spec + rank behavior |
| [ENG-82](https://linear.app/engramhq/issue/ENG-82) | — | Golden eval fixtures BV-01–10 |
| [ENG-83](https://linear.app/engramhq/issue/ENG-83) | — | Pack lifecycle engine |
| [ENG-85](https://linear.app/engramhq/issue/ENG-85) | — | Verticalization runtime skeleton |
| [ENG-90](https://linear.app/engramhq/issue/ENG-90) | — | Vertical eval CI |
| [ENG-93](https://linear.app/engramhq/issue/ENG-93) | — | Marketing MVP benchmark |

### Gate M2 — Done

| Linear | GitHub | Title | Status |
| --- | --- | --- | --- |
| [ENG-91](https://linear.app/engramhq/issue/ENG-91) | [#1](https://github.com/tryvinci/brainy/issues/1) / [PR #13](https://github.com/tryvinci/brainy/pull/13) | Open rebuild PR | **Merged** |
| [ENG-96](https://linear.app/engramhq/issue/ENG-96) | [#2](https://github.com/tryvinci/brainy/issues/2) | OSS legal files | Done |
| [ENG-97](https://linear.app/engramhq/issue/ENG-97) | [#3](https://github.com/tryvinci/brainy/issues/3) | Docker Compose stack | Done |
| [ENG-98](https://linear.app/engramhq/issue/ENG-98) | [#4](https://github.com/tryvinci/brainy/issues/4) | Staging + post-deploy eval | Done (docker-smoke CI) |
| [ENG-100](https://linear.app/engramhq/issue/ENG-100) | [#9](https://github.com/tryvinci/brainy/issues/9) | Mem0 live competitor adapter | Done |

### Gate M3 — Done

| Linear | GitHub | Title | Status |
| --- | --- | --- | --- |
| [ENG-87](https://linear.app/engramhq/issue/ENG-87) | [#5](https://github.com/tryvinci/brainy/issues/5) | pgvector + hybrid retrieval | Done |
| [ENG-99](https://linear.app/engramhq/issue/ENG-99) | [#6](https://github.com/tryvinci/brainy/issues/6) | Close remaining eval seeds | Done |
| [ENG-73](https://linear.app/engramhq/issue/ENG-73) | [#10](https://github.com/tryvinci/brainy/issues/10) | Benchmark methodology + moat report | Done |
| [ENG-63](https://linear.app/engramhq/issue/ENG-63) | — | Embedding strategy (pgvector phased) | Done |
| [ENG-83](https://linear.app/engramhq/issue/ENG-83) | — | Campaign lifecycle semantics | Done |
| [ENG-61](https://linear.app/engramhq/issue/ENG-61) | — | Primitives + YAML packs (PD) | Done |
| — | [#7](https://github.com/tryvinci/brainy/issues/7) | Outcome → belief MVP-3 | Done |
| — | [#8](https://github.com/tryvinci/brainy/issues/8) | Pack JSON Schema validation MVP-4 | Done |

| — | [PR #16](https://github.com/tryvinci/brainy/pull/16), [PR #17](https://github.com/tryvinci/brainy/pull/17) | OpMem benchmark + fixes | **Merged** (Track A prereq) |

### Gate M5 — Commercial (Track C)

| Linear | GitHub | Title |
| --- | --- | --- |
| — | [#11](https://github.com/tryvinci/brainy/issues/11) | API key auth per tenant |
| — | [#12](https://github.com/tryvinci/brainy/issues/12) | Commercial beta checklist |

### Unblocked (Gate M4 research)

| Linear | Notes |
| --- | --- |
| [ENG-56](https://linear.app/engramhq/issue/ENG-56) | Finance epic — **unblocked** after Gate M3; research only until M4 sign-off |
| [ENG-76](https://linear.app/engramhq/issue/ENG-76) | Finance taxonomy — research |
| [ENG-78](https://linear.app/engramhq/issue/ENG-78) | Finance eval fixtures — Gate M4 |

---

## Operating rhythm

1. **Every PR:** `go test ./...` (Tiers 0–3).
2. **M2:** `docker compose up` + eval harness before staging deploy.
3. **Issue sync:** Update this file when creating/closing Linear or GitHub issues.
4. **M3 signed off** — finance **research** may begin; no finance **pack merge** until Gate M4.

---

## References

- [`marketing-mvp-benchmark.md`](./marketing-mvp-benchmark.md)
- [`staging-deploy-runbook.md`](../staging-deploy-runbook.md)
- GitHub issues: https://github.com/tryvinci/brainy/issues
