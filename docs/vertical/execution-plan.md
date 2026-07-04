# Execution Plan — Marketing Vetting & GTM (Linear ↔ GitHub sync)

**Status:** Active (2026-06-30)  
**Repo:** [tryvinci/brainy](https://github.com/tryvinci/brainy) (`dev`)  
**Linear project:** [SoTA Vertical Memory (Brainy)](https://linear.app/engramhq/project/sota-vertical-memory-brainy-4efb2f9a793a)  
**Linear doc:** [Marketing Vetting & GTM Execution Plan](https://linear.app/engramhq/document/marketing-vetting-and-gtm-execution-plan-177a372fc2eb)  
**Policy:** [`marketing-vetting-gate.md`](./marketing-vetting-gate.md) · [`go-to-market-roadmap.md`](./go-to-market-roadmap.md)

---

## Gate status

| Gate | Name | Status | GitHub milestone | Linear milestone |
| --- | --- | --- | --- | --- |
| **M1** | Deterministic marketing MVP | **Done** | — | Gate M1: Marketing Deterministic MVP |
| **M2** | Publish & OSS preview | **Done** (PR #13 merged) | [#1](https://github.com/tryvinci/brainy/milestone/1) | Gate M2: Publish & OSS Preview |
| **M3** | Marketing technical proof | **Done** | [#2](https://github.com/tryvinci/brainy/milestone/2) | Gate M3: Marketing Technical Proof |
| **M4** | Finance / second vertical | **Unblocked** (research) | — | — |
| **M5** | Commercial API beta | Open | [#3](https://github.com/tryvinci/brainy/milestone/3) | Gate M5: Commercial API Beta |

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

### Gate M3 — Active

| Linear | GitHub | Title |
| --- | --- | --- |
| [ENG-87](https://linear.app/engramhq/issue/ENG-87) | [#5](https://github.com/tryvinci/brainy/issues/5) | pgvector + hybrid retrieval |
| [ENG-99](https://linear.app/engramhq/issue/ENG-99) | [#6](https://github.com/tryvinci/brainy/issues/6) | Close remaining eval seeds |
| — | [#7](https://github.com/tryvinci/brainy/issues/7) | Outcome → belief MVP-3 |
| — | [#8](https://github.com/tryvinci/brainy/issues/8) | Pack JSON Schema validation MVP-4 |
| — | [#10](https://github.com/tryvinci/brainy/issues/10) | Benchmark methodology + moat report |

### Gate M5 — Commercial

| Linear | GitHub | Title |
| --- | --- | --- |
| — | [#11](https://github.com/tryvinci/brainy/issues/11) | API key auth per tenant |
| — | [#12](https://github.com/tryvinci/brainy/issues/12) | Commercial beta checklist |

### Blocked

| Linear | Notes |
| --- | --- |
| [ENG-56](https://linear.app/engramhq/issue/ENG-56) | Finance epic — **blocked until Gate M3** |
| [ENG-76](https://linear.app/engramhq/issue/ENG-76) | Finance taxonomy — research only |
| [ENG-78](https://linear.app/engramhq/issue/ENG-78) | Finance eval fixtures — blocked at M4 |

---

## Operating rhythm

1. **Every PR:** `go test ./...` (Tiers 0–3).
2. **M2:** `docker compose up` + eval harness before staging deploy.
3. **Issue sync:** Update this file when creating/closing Linear or GitHub issues.
4. **No finance implementation** until M3 sign-off.

---

## References

- [`marketing-mvp-benchmark.md`](./marketing-mvp-benchmark.md)
- [`staging-deploy-runbook.md`](../staging-deploy-runbook.md)
- GitHub issues: https://github.com/tryvinci/brainy/issues
