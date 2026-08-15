# Go-to-Market Roadmap — Open Source, Benchmarks, Commercial

**Status:** Active (2026-07-04) — Gates M1–M3 done; launch tracks A→B→C sequential  
**Depends on:** [`marketing-vetting-gate.md`](./marketing-vetting-gate.md) · [`execution-plan.md`](./execution-plan.md)  
**Audience:** Founders, eng leads — what it takes from today to public repo, published benchmarks, and revenue.

---

## Where we are

| Asset | Status |
| --- | --- |
| Go API (ingest, search, correct, suppress, async worker) | Built, Docker Compose + CI |
| Marketing vertical pack + eval harness | **16/16 fixtures green** (Gate M3) |
| OpMem operational benchmark | **12/12 expected** (PR #16+#17 merged to `dev`) |
| CI (fixture regression) | Green on public `origin` |
| LICENSE (Apache-2.0) | **Done** |
| Docker Compose stack | **Done** |
| Live competitor adapter | **Done** (ENG-100) |
| Moat benchmark report + methodology | **Done** (Gate M3) |
| API auth / billing | **Beta-ready** — API keys (#11), checklist (#12) |
| Hosted production API | **Not started** (Track C) |

**Honest position:** Marketing technical proof is **done**. Next is a **sequential public launch**: OSS preview (A) → benchmark-led narrative (B) → hosted beta (C).

---

## Three launch tracks (sequential after M3)

Tracks run **in order**. Do not start B until A ships; do not start C until B publishes.

```
M1 ──► M2 ──► M3 (Done)
                  │
                  ▼
           Track A — OSS preview (v0.1.0)
                  │
                  ▼
           Track B — Benchmark-led launch
                  │
                  ▼
           Track C — Hosted API beta (M5)
```

| Track | Minimum gate to **start** | Minimum gate to **claim publicly** |
| --- | --- | --- |
| **A. Open source preview** | M3 done | `v0.1.0` tagged, README quickstart, Docker reproduces evals |
| **B. Published benchmarks** | Track A shipped | OpMem 12/12 + moat report public; launch blog/landing |
| **C. Hosted API beta** | Track B published | M5: auth (#11), commercial checklist (#12), 2–3 design partners |

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

### Checklist: developer preview (Track A — active)

Enough for **`v0.1.0` developer preview** on public `main`:

| Item | Status | Effort |
| --- | --- | --- |
| Apache-2.0 `LICENSE` | Done | — |
| Public repo + GitHub Actions | Done | — |
| `README.md` — 5-min quickstart | In progress | Small |
| `CONTRIBUTING.md`, `SECURITY.md` | Done | — |
| `.env.example` | Done | — |
| Docker Compose (API + Postgres + worker) | Done | — |
| Merge `dev` → `main` + tag `v0.1.0` | Pending | Small |

**Timeline:** ~1 week from M3 sign-off — mostly release ops, not new features.

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

## Track B — Published Brainy benchmarks

### What we have today (Tier 3)

- Fixture-driven **Brainy-only** runs
- Capability matrix (`evals/marketing_mvp_matrix.json`)
- Report: `docs/vertical/marketing-mvp-benchmark.md`

This supports internal claims and architecture docs. It is **not** sufficient for a public “we beat vendor X” blog post.

### Benchmark publication ladder

| Stage | What | Gate | Output |
| --- | --- | --- | --- |
| **B1 — Reproducible Brainy report** | Anyone clones repo, runs harness | M2 | CI badge + committed report on release tags |
| **B2 — Core parity on a live incumbent API** | Same `fixtures/parity/` against a hosted memory API | M2 | Internal pin under `docs/research/competitive/` |
| **B3 — Marketing vertical head-to-head** | Brainy vertical fixtures vs generic APIs on expressible scenarios; document what they cannot model | M3 | `docs/benchmarks/marketing-vertical-moat.md` |
| **B4 — Multi-vendor public track** | Shared public scenarios across vendors | M3+ | `docs/benchmarks/competitor-index-results.md` |

### What to build (engineering)

| Component | Purpose |
| --- | --- |
| `evals/competitors/` | Map fixture ingest/search to optional live APIs |
| `evals/competitors/base.py` | Shared runner interface, skip if API key missing |
| `evals/run_competitor_benchmark.py` | Side-by-side JSON + markdown report (internal) |
| `docs/benchmarks/METHODOLOGY.md` | Scoring rubric, fair-use, pinned versions |
| Nightly workflow (optional) | Competitor runs on schedule, not every PR (cost) |

### Scoring rubric (critical for credibility)

Do **not** use one aggregate score. Publish per capability:

| Dimension | Brainy expectation | Generic memory API |
| --- | --- | --- |
| Parity (preference, profile, fact) | Must match or explain delta | Baseline |
| Principle > preference | Win | Fail / N/A |
| Brand rule extraction | Win | Fail |
| Campaign lifecycle | Win | Fail |
| Suppression leak | Win | Approximate |
| Semantic paraphrase (post ENG-87) | Target win | May win today |

Generic APIs may win on raw embedding semantic search until ENG-87 ships — **say that explicitly** in published benchmarks, without naming vendors in user-facing copy.

### When to publish externally

| Channel | Minimum bar |
| --- | --- |
| README / docs site | B1 (M2) |
| Hacker News / blog “marketing memory benchmark” | B3 (M3) + methodology |
| Sales deck (capability, not vendor bake-off) | B3 + 1 design partner anecdote |
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

## Consolidated timeline (sequential tracks)

| Phase | Weeks (indic.) | Track | Unlocks |
| --- | --- | --- | --- |
| **M2 + M3** | Done | Engineering gates | Technical proof complete |
| **Track A** | 1 | OSS preview + `v0.1.0` | Public clone-and-run, contributor onboarding |
| **Track B** | 2–4 | OpMem 12/12 publish + launch content | Honest Brainy benchmark narrative |
| **Track C** | 6–10 | Auth, billing, design partners | First revenue (manual invoice) |
| **M4** | — | Finance research (parallel, non-blocking) | Second vertical discovery |

**Earliest honest dates (order of magnitude):**

- **`v0.1.0` developer preview:** ~1 week (Track A — now)
- **Public benchmark launch post:** ~3–5 weeks after Track A (Track B)
- **Paid API beta:** ~10–16 weeks from Track B start (Track C / M5)  

---

## Decision record (recommended)

| Decision | Recommendation |
| --- | --- |
| License | Apache-2.0 (runtime + packs) — **done** |
| Launch sequence | **Sequential:** A (OSS) → B (benchmarks) → C (hosted API) |
| First public artifact | `v0.1.0` on `main` + README quickstart (Track A) |
| Second public artifact | OpMem 12/12 + moat report launch post (Track B) |
| First revenue | Managed API for marketing agents — Track C / M5 |
| Competitor benchmarks | Adapter done; keep pins internal until Track B |
| Finance | Gate M4 research — not on launch critical path |
| Claim discipline | No “SOTA” until Track B published with methodology |

---

## References

- [`marketing-vetting-gate.md`](./marketing-vetting-gate.md) — M1–M4 gates
- [`marketing-mvp-benchmark.md`](./marketing-mvp-benchmark.md) — current Tier 3 report
- [`docs/brainy/03-competitor-index.md`](../brainy/03-competitor-index.md) — vendor list for B4 (internal)
- [`docs/external-postgres-runbook.md`](../external-postgres-runbook.md) — staging setup
