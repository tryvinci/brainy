# Brainy — Investor One-Pager

**Memory infrastructure for AI agents that stays correct over time.**

---

## What we do

- Brainy is a **memory layer for AI agents**: ingest conversations and events → extract lasting facts → retrieve the right context when the agent needs it.
- Built as a **Go service on Postgres** (self-host or hosted). Thin API: ingest, search, correct, suppress, supersede.
- First vertical: **marketing agents** — brand rules, campaign lifecycle, preference hierarchy, outcome→belief ranking via YAML “packs” (no per-vertical code forks).

## Thesis

The market is racing to score high on conversational recall (LoCoMo, LongMemEval). That race is noisy and contested.

**Enterprises buy something else:** memory that can **update, correct, forget, and govern** — and that speaks the language of their domain.

- Leaders (e.g. Mem0) are largely **ADD-only**: facts accumulate; stale truth still lives in the store.
- We bet on **operational correctness + vertical packs**: memory with a real update/forget contract, plus domain vocabulary and ranking out of the box.
- Tagline we own: **"Memory that stays true."**

## Where we stand (measured)

| Axis | Brainy | Mem0 (same fixtures / same pin) |
| --- | ---: | ---: |
| **OpMem** — update / correct / suppress / isolate | **12/12** | 9/12 |
| **Marketing vertical** — brand rules, campaigns, governance | **15–16/16** | 4/16 |
| **LOCOMO smoke** — 1 conversation × 30 Q (diagnostic pin) | **19/30** | 12/30 |
| Search latency (staging, best pins) | ~730–890ms p50 | ~0.9–1.1s p50 (their published) |

**Honest caveats (say these out loud):**

- Full industry suites (full LoCoMo ~1,540 Q, LongMemEval, BEAM) are **not yet run** on a publishable stack. Mem0/Hindsight headline ~90+ scores are on those suites — we do not claim parity there yet.
- Conversational multi-hop still trails on hard list questions; the fix is architectural (typed facts + structured recall), not more benchmark tuning.
- Product is **beta-ready** for design partners; GA legal/ops items (ToS, privacy, prod cutover) remain.

## Competitive frame

| Player | What they sell | Where we differ |
| --- | --- | --- |
| **Mem0** | Universal memory API; broad ecosystem; strong recall marketing | ADD-only; weak on correct/forget/vertical governance |
| **Zep** | Temporal knowledge graph; low-latency story | Graph complexity; not vertical packs |
| **Letta** | Stateful agent runtime + filesystem memory | Different layer (agent OS vs memory API) |
| **Others** (Supermemory, Cognee, Hindsight) | Cloud / graph / leaderboard plays | We win on **ops correctness + packs + cost/latency of a Go binary** |

## Pre-release program (what “done” looks like)

Program of record: `docs/research/master-plan.md`.

1. **De-overfit** — remove any LOCOMO-shaped shortcuts; honest re-baseline.
2. **Typed memory + structured recall** — facts as `(subject, predicate, value, time)`; enumeration queries that actually cover lists.
3. **Publishable benchmarks** — full LoCoMo, LongMemEval, BEAM, with multi-seed runs and counter-runs on competitors’ own harnesses.
4. **OpMem as a public benchmark** — make “can your memory update and forget?” the industry question we define.
5. **Second vertical pack** (support/CRM) — prove packs generalize beyond marketing.
6. **Release bar** — beat naive baselines (full-context / RAG / filesystem+grep) with error bars; hold OpMem + vertical lead; report accuracy **and** latency **and** tokens.

## Business posture

- **Stage:** Pre-release / design-partner beta. Apache-2.0 OSS + hosted path.
- **Wedge customer:** Teams shipping marketing (then support) agents who need brand/campaign truth to stay correct — not just “remember what was said.”
- **Moat shape:** Hard ops contract + pack model + reproducible evidence culture (fixtures and harnesses in-repo). Not a one-number leaderboard claim.

## Ask / next proof points for the board

- Design partners live on the marketing pack.
- OpMem paper + multi-vendor table (Brainy vs Mem0/Zep/Letta).
- One fair full-suite counter-run published with methodology.
- Second pack shipped; first paid pilot path clear.

---

# FAQ

**Q: What is Brainy in one sentence?**  
A: A memory API for AI agents that can remember, update, correct, and forget — with vertical packs so marketing (and next, support) agents get domain governance out of the box.

**Q: Why will this win against Mem0?**  
A: Mem0 wins on distribution and recall marketing. We win where buyers feel pain next: **stale and ungoverned memory**. Measured today: we beat them on operational memory (12/12 vs 9/12) and marketing governance fixtures (15–16/16 vs 4/16). We’re not claiming we’ve beaten their full conversational leaderboard numbers yet.

**Q: Are you SOTA on LoCoMo?**  
A: No. We’ve run a **diagnostic smoke pin** (19/30 vs Mem0 12/30 same pin). Full publishable suite runs are on the roadmap. The field’s LoCoMo numbers are also widely contested — we will publish with baselines, seeds, and judge prompts, or not at all.

**Q: What’s the product today?**  
A: Working Go API + worker, Postgres storage, marketing pack, async extract, correct/suppress/supersede, self-host via Docker, hosted staging path, CI + reproducible eval harnesses.

**Q: Who is the customer?**  
A: Builders of AI agents for **marketing and (next) support/CRM** — agencies, product teams, and platforms that need brand rules and customer state to stay accurate across sessions.

**Q: What’s the business model?**  
A: Open core (Apache-2.0) for adoption; hosted Brainy with tenancy/auth for design partners and GA. Vertical packs deepen switching costs.

**Q: What’s the biggest risk?**  
A: Conversational memory suites are the default buyer comparison. If we only lead on OpMem/verticals and lag badly on full LoCoMo/LongMemEval, sales cycles get harder. Mitigation: architectural recall work + fair counter-runs; never fake the number.

**Q: What’s the biggest opportunity?**  
A: Own the category narrative **“memory that stays true”** while the market is still fighting over noisy recall scores — and lock early vertical depth Mem0 can’t copy without becoming a different product.

**Q: What should we expect by the end of this pre-release cycle?**  
A: Publishable, multi-seed numbers on the industry suites; OpMem established as a public multi-vendor benchmark; second vertical pack; design partners in production-like use; a clear GA path — without claiming a single fake SOTA number.

**Q: Why trust your numbers?**  
A: Fixtures and runners are in the repo. We counter-run Mem0 on the same marketing and OpMem suites. We refuse single-seed “SOTA” claims. That discipline is intentional product strategy in a market full of unreproducible scores.
