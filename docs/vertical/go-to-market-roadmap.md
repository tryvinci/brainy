# Go-to-Market Roadmap — Open Source, Benchmarks, Commercial

**Status:** Product planning (2026-06-23)  
**Depends on:** [`marketing-vetting-gate.md`](./marketing-vetting-gate.md)  
**Audience:** Founders, eng leads — what it takes from today to public repo, published benchmarks, and revenue.

---

## Where we are

| Asset | Status |
| --- | --- |
| Go API (ingest, search, correct, suppress, async worker) | Built, local/staging only |
| Marketing vertical pack + eval harness | Built, M1 passed locally |
| CI (fixture regression) | Built, not on public origin yet |
| LICENSE | **Missing** |
| API auth / billing | **Missing** |
| Live competitor benchmarks | **Not started** (documented Mem0 gap only) |
| Hosted production API | **Not started** |

**Honest position:** Strong **engineering prototype** with reproducible internal vetting. Not yet a **public product** or **credible external benchmark publication**.

---

## Three outward-facing tracks (parallel after M2)

These are independent timelines that share engineering foundations:

```
                    M1 (now)
                       │
                       ▼
              M2 — Publish & staging
                       │
       ┌───────────────┼───────────────┐
       ▼               ▼               ▼
  Track A          Track B          Track C
  Open source      Benchmarks       Commercial API
  (repo public)    (marketing moat)   (revenue)
       │               │               │
       └───────────────┴───────────────┘
                       │
              M3 — Marketing technical proof
                       │
       ┌───────────────┴───────────────┐
       ▼                               ▼
  OSS v1.0 claim                  Paid API beta
  + benchmark report              + design partners
```

| Track | Minimum gate to **start** | Minimum gate to **claim publicly** |
| --- | --- | --- |
| **A. Open source** | M2 | M2 for “developer preview”; M3 for “marketing memory v1” |
| **B. Published benchmarks** | M2 (staging + Mem0 API adapter) | M3 (full seed coverage + methodology doc) |
| **C. Sell (API or OSS support)** | M3 + commercial hardening | M3 + 2–3 design partners + billing/auth |

Finance and second verticals remain **Gate M4** — unrelated to first public launch.

---

## Track A — Open source

### What “open source” means here

Two viable models (pick one before public launch):

| Model | License | Revenue |
| --- | --- | --- |
| **Open core** | Apache-2.0 / MIT for runtime + packs | Hosted API, enterprise support, managed Postgres |
| **Source available** | BSL or similar with conversion date | Same, plus delay on competitive SaaS clones |

Recommendation for Brainy: **Apache-2.0** on runtime + marketing pack YAML. Keeps pack format adoptable; monetize hosted ops and support.

### Checklist: developer preview (Gate M2)

Enough for a **public GitHub repo** labeled “early / not production”:

| Item | Status | Effort |
| --- | --- | --- |
| Choose license + add `LICENSE` | Missing | Small |
| Push `dev` to public `origin`, enable GitHub Actions | Open (ENG-91) | Small |
| `README.md` — quickstart, architecture link, disclaimer | Partial | Small |
| `CONTRIBUTING.md`, `SECURITY.md` | Missing | Small |
| `.env.example` — no secrets, documented vars | Check | Small |
| Docker Compose (API + Postgres + worker) | Missing | Medium |
| Remove/archive internal-only paths (`.omx` policy?) | Partial | Small |
| Issue templates / basic CODEOWNERS | Missing | Small |

**Timeline:** 1–2 weeks after M1 if prioritized — mostly ops/docs, not new features.

### Checklist: credible OSS v1.0 (Gate M3)

What reviewers and adopters expect before calling it “real”:

| Item | Gate |
| --- | --- |
| M3 vetting green (Tier 4 eval seeds) | M3 |
| Self-host runbook tested by someone not on core team | M3 |
| Semantic retrieval (ENG-87) documented with fallback | M3 |
| Versioned releases + changelog | M3 |
| API stability note (v0 vs v1 contract) | M3 |
| Marketing pack documented as reference implementation | Done |

**Do not** open source primarily for distribution before M2 — you want CI and staging reproducible first.

---

## Track B — Published marketing benchmarks vs competitors

### What we have today (Tier 3)

- Fixture-driven **Brainy-only** runs
- **Documented Mem0 gap matrix** (`evals/marketing_mvp_matrix.json`) — not live Mem0 API calls
- Report: `docs/vertical/marketing-mvp-benchmark.md`

This supports internal claims and architecture docs. It is **not** sufficient for a public “we beat Mem0/Zep” blog post.

### Benchmark publication ladder

| Stage | What | Gate | Output |
| --- | --- | --- | --- |
| **B1 — Reproducible Brainy report** | Anyone clones repo, runs harness | M2 | CI badge + committed report on release tags |
| **B2 — Live Mem0 parity comparison** | Same `fixtures/parity/` against Mem0 Cloud API | M2 | `docs/benchmarks/marketing-vs-mem0-parity.md` |
| **B3 — Marketing vertical head-to-head** | Brainy vertical fixtures vs Mem0 on expressible scenarios; document where Mem0 cannot model | M3 | `docs/benchmarks/marketing-vertical-moat.md` |
| **B4 — Multi-vendor public track** | Mem0 + Zep (+ optional Supermemory) on shared public scenarios | M3+ | `docs/benchmarks/competitor-index-results.md` |

### What to build (engineering)

| Component | Purpose |
| --- | --- |
| `evals/competitors/mem0_adapter.py` | Map fixture ingest/search to Mem0 API |
| `evals/competitors/base.py` | Shared runner interface, skip if API key missing |
| `evals/run_competitor_benchmark.py` | Side-by-side JSON + markdown report |
| `docs/benchmarks/METHODOLOGY.md` | Scoring rubric, fair-use, pinned API versions |
| Nightly workflow (optional) | Competitor runs on schedule, not every PR (cost) |

### Scoring rubric (critical for credibility)

Do **not** use one aggregate score. Publish per capability:

| Dimension | Brainy expectation | Mem0 / generic |
| --- | --- | --- |
| Parity (preference, profile, fact) | Must match or explain delta | Baseline |
| Principle > preference | Win | Fail / N/A |
| Brand rule extraction | Win | Fail |
| Campaign lifecycle | Win | Fail |
| Suppression leak | Win | Approximate |
| Semantic paraphrase (post ENG-87) | Target win | May win today |

Mem0 may win on raw embedding semantic search until ENG-87 ships — **say that explicitly** in published benchmarks.

### When to publish externally

| Channel | Minimum bar |
| --- | --- |
| README / docs site | B1 (M2) |
| Hacker News / blog “marketing memory benchmark” | B3 (M3) + methodology |
| Sales deck “vs Mem0” | B3 + 1 design partner anecdote |
| Analyst / press | B4 + hosted API uptime story |

---

## Track C — Selling as API or open source business

### Path 1: Hosted API (SaaS)

**Start design partner conversations:** Gate M3  
**Take money:** M3 + commercial checklist below

| Layer | Required for beta | Required for GA |
| --- | --- | --- |
| **Product** | Stable ingest/search, marketing vertical | SLA, versioning, migration policy |
| **Auth** | API keys per tenant | OAuth / SSO for enterprise |
| **Multi-tenancy** | `tenant_id` isolation (exists) | Hard isolation audit, quotas |
| **Billing** | Manual / invoice | Stripe metered (ingest + search calls) |
| **Ops** | Staging + prod, backups | On-call, status page, RPO/RTO |
| **Legal** | Terms of service, privacy policy | DPA, data residency option |
| **Support** | Slack with 2–3 partners | Ticketing, docs site |

Estimated eng after M3: **6–10 weeks** for credentialed beta (auth, deploy, billing stub, docs).

### Path 2: Open source + commercial (self-host)

**Start:** OSS developer preview at M2  
**Sell:** Support contracts, managed hosting, enterprise pack governance

| Offering | Buyer | Needs |
| --- | --- | --- |
| **Free self-host** | Developers | Docker, docs, LICENSE |
| **Managed Brainy** | Teams without ops | Hosted API (Path 1) |
| **Enterprise** | Agency / brand platforms | SSO, audit logs, pack signing, SLA |

Revenue without hosted API is slower — plan on **managed hosting** as primary monetization even if core is OSS.

### Path 3: Embed in Vinci product (not standalone API)

If Brainy is memory for an existing Vinci marketing agent:

- Gate M3 for vertical proof
- No public API required initially
- Benchmarks still valuable for positioning
- Open source optional (could stay private monorepo)

Clarify which path Engram/Vinci wants — it changes priority of auth/billing vs agent integration.

---

## Testing & exposure (how vetting connects to GTM)

Three test modes — see also [`marketing-vetting-gate.md`](./marketing-vetting-gate.md).

| Mode | When | Purpose |
| --- | --- | --- |
| **Fixture / CI** | Now, every PR | Regression — never skip |
| **Staging API** | M2 | Dogfood, design partners, competitor adapters |
| **Live competitor benchmarks** | M2 start, M3 publish | External proof — not a merge gate |

### Exposure timeline

| Audience | Gate | Surface |
| --- | --- | --- |
| Core team | Now | Local `go run ./cmd/api` |
| Internal dogfood | M2 | Staging URL + runbook |
| Open source contributors | M2 | Public repo + Docker |
| Design partners | M3 | Staging API key + marketing vertical docs |
| Public benchmark readers | M3 | Committed reports + methodology |
| Paying customers | M3 + commercial checklist | Prod API + billing |

---

## Consolidated timeline (realistic)

Assuming part-time to one eng focus; adjust if dedicated.

| Phase | Weeks (indic.) | Milestone | Unlocks |
| --- | --- | --- | --- |
| **M2** | 1–2 | Push public repo, staging, Docker, LICENSE | OSS preview, B1 benchmarks, dogfood |
| **M3 eng** | 4–8 | lc02, MVP-3/4, ENG-87, remaining fixtures | Credible marketing proof |
| **M3 publish** | 1–2 | B2–B3 benchmark reports, docs site | Blog, sales narrative |
| **Commercial beta** | 6–10 | API keys, deploy, 2–3 partners | First revenue (manual) |
| **OSS v1.0 tag** | — | M3 + release process | Community adoption |
| **M4** | — | Finance research only after M3 sign-off | Second vertical |

**Earliest honest dates (order of magnitude):**

- **Public repo (preview):** ~2 weeks from M2 start  
- **Published marketing vs Mem0 benchmark:** ~6–10 weeks (needs M3 + adapter work)  
- **Paid API beta:** ~10–16 weeks from today (M3 + commercial layer)  
- **Finance vertical:** after M4 — not on critical path to launch  

---

## Decision record (recommended)

| Decision | Recommendation |
| --- | --- |
| License | Apache-2.0 (runtime + packs) |
| First public artifact | Repo + reproducible benchmark report (B1), not hosted API |
| First revenue | Managed API for marketing agents (design partners) |
| Competitor benchmarks | Mem0 first (pinned parity fixtures), Zep second (public track only) |
| Finance | Blocked until M4 — not required for launch |
| Claim discipline | No “SOTA” until B3 published with methodology |

---

## References

- [`marketing-vetting-gate.md`](./marketing-vetting-gate.md) — M1–M4 gates
- [`marketing-mvp-benchmark.md`](./marketing-mvp-benchmark.md) — current Tier 3 report
- [`docs/mem0-parity-matrix.md`](../mem0-parity-matrix.md) — Mem0 reference pin
- [`docs/brainy/03-competitor-index.md`](../brainy/03-competitor-index.md) — vendor list for B4
- [`docs/external-postgres-runbook.md`](../external-postgres-runbook.md) — staging setup
