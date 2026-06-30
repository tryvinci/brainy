# Execution Plan — Marketing Vetting & GTM (Linear ↔ GitHub sync)

**Status:** Active (2026-06-23)  
**Repo:** [tryvinci/brainy](https://github.com/tryvinci/brainy)  
**Linear project:** [SoTA Vertical Memory (Brainy)](https://linear.app/engramhq/project/sota-vertical-memory-brainy-4efb2f9a793a)  
**Canonical policy:** [`marketing-vetting-gate.md`](./marketing-vetting-gate.md) · [`go-to-market-roadmap.md`](./go-to-market-roadmap.md)

---

## Gate status

| Gate | Name | Status | Target |
| --- | --- | --- | --- |
| **M1** | Deterministic marketing MVP | **Done** | 2026-06 |
| **M2** | Publish & OSS preview | **In progress** | 2026-07 |
| **M3** | Marketing technical proof | Open | 2026-08 |
| **M4** | Finance / second vertical | **Blocked** | After M3 |
| **M5** | Commercial API beta | Open | After M3 |

---

## Sync index (Linear ↔ GitHub)

Update this table when issues are created or closed. GitHub is the doc source of truth; Linear is the execution tracker.

### Done (close in Linear)

| Linear | GitHub | Title | Gate |
| --- | --- | --- | --- |
| ENG-58 | — | Marketing first vertical wedge | M1 |
| ENG-61 | — | Primitives + YAML packs | M1 |
| ENG-80 | — | Marketing use case map | M1 |
| ENG-81 | — | Brand voice (Principle + IdentityPrior) | M1 |
| ENG-82 | — | Marketing golden eval fixtures BV-01–10 | M1 |
| ENG-83 | — | Pack lifecycle engine | M1 |
| ENG-85 | — | Verticalization runtime skeleton | M1 |
| ENG-90 | — | Vertical eval CI integration | M1 |
| ENG-93 | — | Marketing MVP benchmark report | M1 |

### M2 — Publish & OSS preview (active)

| Linear | GitHub | Title | Owner |
| --- | --- | --- | --- |
| ENG-91 | GH-M2-1 | Push dev, open rebuild PR (CI done) | Eng |
| ENG-94 | GH-M2-2 | Apache-2.0 LICENSE + CONTRIBUTING + SECURITY | Eng |
| ENG-95 | GH-M2-3 | Docker Compose (API + Postgres + worker) | Eng |
| ENG-96 | GH-M2-4 | Staging deploy runbook + post-deploy eval hook | Eng |
| ENG-97 | GH-M2-5 | GitHub milestone + project board for vetting gates | Eng |

### M3 — Marketing technical proof

| Linear | GitHub | Title | Blocked by |
| --- | --- | --- | --- |
| ENG-87 | GH-M3-1 | pgvector + hybrid retrieval + regression fixtures | ENG-63 PD |
| ENG-98 | GH-M3-2 | Close eval seeds 3, 5, 8–10 (lc02, belief, patterns) | M2 |
| ENG-99 | GH-M3-3 | MVP-3 outcome → belief rank loop | M2 |
| ENG-100 | GH-M3-4 | MVP-4 pack JSON Schema validation on ingest | M2 |
| ENG-101 | GH-M3-5 | Pack-driven classification_rules (MVP-1.1) | M2 |

### Benchmarks (start M2, publish M3)

| Linear | GitHub | Title | Gate |
| --- | --- | --- | --- |
| ENG-102 | GH-B-1 | Mem0 live adapter + parity side-by-side benchmark | M2 |
| ENG-103 | GH-B-2 | Benchmark METHODOLOGY.md + published moat report | M3 |

### Commercial (after M3)

| Linear | GitHub | Title | Gate |
| --- | --- | --- | --- |
| ENG-104 | GH-C-1 | API key auth per tenant | M3 |
| ENG-105 | GH-C-2 | Commercial beta checklist (billing, ToS, prod) | M3 |

### Blocked / deferred

| Linear | Notes |
| --- | --- |
| ENG-56, ENG-76, ENG-78 | Finance — **Gate M4 only** |
| ENG-92 | Provider extraction — after M3 core |
| ENG-73 | Superseded by `evals/marketing_mvp_matrix.json` + METHODOLOGY (ENG-103) |

---

## Weekly operating rhythm

1. **Every PR:** `go test ./...` (Tiers 0–3).
2. **Every merge to `dev`:** regenerate `marketing-mvp-benchmark.md` on release tags.
3. **M2+:** staging URL runs same eval harness as CI.
4. **No finance pack PRs** until M3 sign-off in Linear (Gate M4 unlocked).

---

## References

- Linear document: *Marketing Vetting & GTM Execution Plan* (project doc)
- GitHub milestone: `Marketing Gate M2`
- [`marketing-mvp-benchmark.md`](./marketing-mvp-benchmark.md)
