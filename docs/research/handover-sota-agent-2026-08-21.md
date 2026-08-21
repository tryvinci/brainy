# Handover — next agent: LoCoMo 80% / same-pin vs Mem0

**Date:** 2026-08-21  
**Audience:** the next coding agent. Read this first. Do not start from the 2026-08-17 assessment pack as live truth.  
**Owner ask:** take Brainy to a defensible conversational claim — user phrasing is “SOTA / beat Mem0 / 80% LoCoMo.”  
**Repo SOP:** that phrasing is a *goal*, not a claim you may write in product copy. Competitive language requires a **frozen same-pin win** (same dataset SHA, same judge temp 0, same answerer, same question set, same harness) **and** explicit approval. The word “SOTA” stays gated even after a same-pin win until the owner lifts it.

This file is the live start doc. Older research notes stay useful as history. If they disagree with the pins or “next step” here, **this file wins**.

---

## 0. First 30 minutes

1. Confirm you are on `dev` (staging) at `f6638d4` (or later). `main` is production and was fast-forwarded to the same SHA on 2026-08-21. Do not push `main` unless the owner asks again.
2. Read this file, then [cycle-closeout.md](./competitive/cycle-closeout.md) section **2026-08-20 — MH packet/proof**.
3. Skim [sota-execution-plan.md](./sota-execution-plan.md) but **do not** treat its “expected S1 compiler first” as live. S0 ledger **outranks** that expectation.
4. Do **not** re-queue R0–R10. Substrate is merged.
5. Do **not** merge [PR #133](https://github.com/tryvinci/brainy/pull/133) (compiler S1–S5 fishing) or revive [PR #131](https://github.com/tryvinci/brainy/pull/131).
6. Do **not** re-run the OpenAI embedding A/B unless `GET /runtime` or the JSON pins are broken.
7. Do **not** burn full n=1540 or a Mem0 same-pin until product `/recall` proof actually moves.

---

## 1. Goal vs allowed claim

| Phrase | What it is | What you may write |
| --- | --- | --- |
| “80% LoCoMo” | Owner target on **full** LoCoMo (n=1540), product and/or industry lane labeled | A measured pin with n, lane, SHA, judge. Never “80%” from 1×30. |
| “beat Mem0” | Same-pin lead vs **Mem0 Platform** on our harness | Only after S6 same-pin. Mem0 published 92.5% n=1540 is **their** path (top-k 200, their harness) — context, not a scoreboard row. |
| “SOTA” | Marketing word | Forbidden in README / GTM until a frozen same-pin win **and** explicit approval. |

Path docs (do not invent a new program):

- [sota-execution-plan.md](./sota-execution-plan.md) — S0→S6 gates. Histogram outranks the written S1-first order.
- [locomo-full-70-80-path.md](./locomo-full-70-80-path.md) — why 1×30 70% is not full LoCoMo.
- [locomo-dual-path-freeze.md](./locomo-dual-path-freeze.md) — product `/recall` vs industry search+harness.
- [sota-representation-path.md](./sota-representation-path.md) — compile facts; episodes are provenance.

**Honest distance:** product `/recall` full n=1540 is **11.4%** on SHA `1b5ab3e`. Fail-closed S0 product is **32/180**. MH product after this merge is **2/33**. Industry S0 is **62/180**. Getting to 80% on n=1540 is a multi-increment proof/reader (then compiler if the ledger flips), not one PR.

---

## 2. Where we landed (2026-08-21)

### Landed SHAs / PRs

| Ref | Role |
| --- | --- |
| `dev` / `main` **now** | `f6638d4` — #134 merge (packet/proof + this handover). Fast-forwarded 2026-08-21. |
| Parent before #134 | `6b8ac5f` — fail-closed + OpenAI A/B (PR #132) |
| PR **#134** `pr/mh-packet-proof-3086` | **Merged** 2026-08-21. MH packet/proof. |
| PR **#132** | Merged. Fail-closed flags, `GET /runtime`, `cmd/reembed`, extraction actually hosted. |
| PR **#133** | OPEN draft. Compiler S1–S5. **Do not merge.** |
| PR **#131** | CLOSED. Mixed. **Do not revive.** |

Linear: [ENG-176](https://linear.app/engramhq/issue/ENG-176/eng-multi-hop-memory-synthesis-consolidation-for-conversational-qa) (MH), parent [ENG-168](https://linear.app/engramhq/issue/ENG-168/epic-conversational-long-memory-product-gaps-from-locomo-smoke).

`dev` = staging. `main` = production. Fast-forward `main` only with explicit owner approval (this handover job had that approval for #134).

### What #134 shipped (product)

Generic packet/proof — **no LoCoMo-named rules, no new compiler regex, no fusion weights:**

1. Query cues attach predicates to hops (`like`/`prefer` → Preference, `own`/`how many have` → Possession, location/health similarly) in `internal/memory/temporal.go`.
2. Planner hops **capitalized people**; `both X and Y` → two entities; stop list includes `animal`/`animals`/`both`/`they`/`them`/`their` (`planner.go`).
3. Hop executor ignores `search_fallback` slot dumps and **intersects** typed values across entities (`hop_executor.go`).
4. When hop slots are empty, compose from hop contents / ProofChain (`likes`/`loves` extracts) (`multihop_packet.go`, `recall.go`).
5. ProofChain before slogans; coverage counts structured proof.

**Live diagnosis** on tenant `integrity-s0-1` / `conv-42`: “What animal do both Nate and Joanna like?” used to resolve entity **`animal`** and answer “Watching Pets Play…”. After #134: hops Nate+Joanna+Preference, answer **`Turtles, Dairy-free Desserts`**.

### Own pins (do not mix rows)

Dataset SHA: `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`

| Suite | Pin | Notes |
| --- | ---: | --- |
| OpMem | **13/13** | Merge gate. Do not spend a cycle here. |
| Marketing vertical | **17/17** | Merge gate. |
| Extraction ceiling | det **139/180**, provider **161/180** | MH coverage **32/33**. Gold is written. |
| S0 product `POST /recall` | **32/180** | Long-lived integrity VM. Ledger: **PROOF 112 / RETRIEVAL 22 / READER 11 / WRITE 3**. |
| S0 industry search+harness | **62/180** | Same tenant. Different path. Do not average with 32/180. |
| S0 MH product (post-#134) | **2/33** | Was **1/33**. Attributed win: turtles. Second hit (soda/candy) is a crowded-list judge accept. [pin](../benchmarks/artifacts/locomo-mh-packet-proof-20260820.md) |
| 3×90 | product **21/90**, industry **33/90** | `--conversations 3 --questions 90` |
| 1×30 freeze (conv-26) | **21/30** | MH 10/10, OD **0/4**, temporal 11/16. **Do not overwrite.** Diagnostic only. Rest of conv-26 in the full run was **12/122**. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old product SHA `1b5ab3e`. Full n=1540 only at S6. |
| Industry historical | **49.8%** | July, **old stack**, search+harness. Not a current-SHA ceiling. |
| LME-20 | **4/20** | Not re-run. Do not spend a cycle on LME-500. |
| Mem0 Platform 1×30 | **11/30** | Freeze 2026-08-15. MH 6/10, OD 3/4, temporal 2/16. **Do not mix with 32/180.** |
| Embedding A/B | OpenAI @768 r@10 **0.333** vs this-rebuild BGE **0.306** | Retrieval-only. Long-lived VM BGE was **0.239** — **do not average**. [pin](../benchmarks/artifacts/embedding-ab-20260820.md) |

**Invalidated:** Aug-19 S0 17/180 / 52/180 (no pgvector, silent extract degrade). Never cite those as quality.

**Bottleneck is PROOF, not compiler WRITE and not embedder.** S0 WRITE_MISS is **3/180**. Coverage 161/180 vs QA 32/180. MH coverage 32/33 vs QA 2/33.

### Competitor stand (honest)

- **1×30 freeze:** Brainy 21/30 vs Mem0 Platform 11/30 is a prior **lead on that 30-item pin**. It is not full LoCoMo and not permission to write “we beat Mem0.”
- **S0 / n=1540:** no same-n Mem0 pin. Do not trail/lead 32/180 or 11.4% vs 11/30 or vs published 92.5%.
- **Ops / marketing:** Brainy lead (Mem0 pins stale). Must not regress. Not the next cycle.
- **Graphiti / Zep / SuperMemory:** no same-pin. Published headlines are context.
- **Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

---

## 3. What the next increment is

S0 said: spend the next increment on the **largest earliest-stage bucket**. That bucket is **PROOF_MISS** (packet compose, hop people vs topic nouns, typed intersect, list/join, reader crowding).

| Increment | Plan name | Do now? |
| --- | --- | --- |
| S2 / S3 residue | Structured answer + hop proof | **In flight on #135.** Dual-entity list intersect, kinship-dest hobbies, specific possession counts, items-for, who-told, and polar teach are now shipped on top of where/group/for/having/child-count. **Next: remasure MH 33** on `integrity-s0-1`. Residue after remasure is likely identity-surface lists (denylist) and whatever the 33-slice still misses. |
| S1 compiler | Provider-extract / named-subject mass | **No** until a fail-closed remasure says WRITE is the bucket again. #133 stays closed. |
| Embedder swap | OpenAI vs BGE | **Done / pinned.** Do not re-run. |
| S4 LME | multi-session + KU | After LoCoMo proof moves. LME-20 4/20 is not the lever. |
| S5 industry | atoms-first answerer | After product proof moves; label the lane. |
| S6 freeze | n=1540 + Mem0 same-pin | **Once**, after stratified deltas exist. |

**Suggested first remasure:** MH-only 33 on `integrity-s0-1` with `--fail-closed --skip-ingest`. Attribute every new CORRECT. Do not invent a new 180 pin unless a full fail-closed S0 actually finishes. Do not treat unit-test date/transfer/list fixtures as a 2/33 replacement.

**Shipped this increment (generic linguistic, fixtures not dataset IDs):** hop `Name and Name` / `Name and Name both` / `with Name`; hop the person after `does`/`has` on count questions; kinship `'s mother` / `her partner` chains family → slot; join compose intersects and does not dump the union; possession/skill lists without occupation/hobby crowding; how-many counts the typed set; Has/Did polar Yes from typed hops only; `practices … at` place extract; unwind/`do to` activity lists; visit/travel superlative; who-answers from other person mentions; `besides` exclusion (stemmed); childhood items as possession; **when-event hops prove a date from observed_at (do not dump event names)**; **given-to hops the giver only and keeps recipient-mentioned values**; **after-clause keeps matching evidence**; community/journey activity lists; family-injury who; organization beneficiaries from affiliation; **where+kinship answers a place from `in`/`at`/`near`, hopping the source person as well as the unnamed partner**; **`with colleagues/friends` is a group filter, not a CapName join**; **`for` clauses keep matching event/item evidence**; **`get with having` hops health, not possession dumps**; **how-many children counts child-cued family members, not partners**; **dual-entity list queries intersect instead of unioning**; **kinship hobby lists filter to the dest person**; **how-many Ferraris counts the head noun, not every possession**; **who-told and polar teach from typed hops**; **journey-change lists stay identity, not occupation**; **pets' names are possession**.

---

## 4. Code structure

Brainy is a Go vertical-memory service: HTTP API + async extraction worker + Postgres.

```text
cmd/api          HTTP :8080 (integrity stack often :18100)
cmd/worker       async extract loop
cmd/reembed      pages the whole DB (ListEmbeddingTargets is global — stop the worker)

internal/memory  compiler + planner + hops + recall  (~56 Go files)
internal/store/postgres   records, atoms, embeddings, entity hub, migrations
internal/embedding        local hash, hosted BGE, OpenAI, fail-closed stats
internal/api              router, auth, /runtime
internal/config           BRAINY_* flags
internal/pack             vertical YAML packs
internal/jobs             worker job processor
internal/auth             API keys (tenant_id:key)

evals/                    stdlib-only Python
evals/public/locomo/      S0 / smoke / full
evals/public/backends/    brainy + mem0
docs/research/            this folder
docs/benchmarks/          pins + artifacts (do not dump into README)
packs/                    marketing / support YAML
```

### Recall path (the product you are changing)

```text
POST /ingest  → extract (deterministic atoms + provider) → memory_records / atoms / embeddings
POST /recall  → PlanQuery → search + hops → EvidencePacket → structured answer / enumerate / hybrid reader
POST /memories/search → retrieval only (industry lane harvests this + harness LLM)
GET  /runtime → signatures, dim, ANN, fallbacks (no secrets)
```

Key files in `internal/memory/`:

| File | Job |
| --- | --- |
| `recall.go` | `Service.Recall` — structured-first answer, ProofChain, enumerate, episode fallback |
| `planner.go` | `PlanQuery`, hop entity extraction, packet coverage |
| `hop_executor.go` | Execute hops, ignore search-fallback dumps, intersect typed values |
| `multihop_packet.go` | Compose from hop slots / contents / ProofChain |
| `temporal.go` | Temporal score + generic predicate hints |
| `provider_extractor.go` | Hosted compiler (Cloudflare gpt-oss-120b via AI Gateway in integrity) |
| `extractor.go` / `attribute_atoms.go` / `clause_subject.go` | Deterministic compiler |
| `reader_hybrid.go` | Optional LLM reader (`BRAINY_RECALL_LLM`) |
| `fusion_v2.go` | Ranking. **Do not retune to fish scores.** |
| `overfit_denylist_test.go` | Blocks LoCoMo/LME-named product rules |
| `compiler_audit_test.go` | Held-out compiler merge gate |
| `entities.go` / `entity_id.go` / `relations.go` | Canonical IDs (R7–R9 substrate) |
| `service.go` | Ingest, upsert, search, jobs (~2.5k LOC) |

Store: `internal/store/postgres/store.go`, `embedding.go`, `runtime.go`, `atoms.go`, `entity_hub.go`, `migrations.go`. `pg_trgm` required. `pgvector` optional — **without it, ANN is dead and S0 is a lie.** Never point at `brainy_sota` (no pgvector).

### Dual eval lanes (never mix scores)

| Lane | What it scores | Harness |
| --- | --- | --- |
| `product-recall` | `POST /recall` answer | `evals/public/locomo/run_s0.py` / `run_smoke.py` |
| `industry-search` | search hits → shared LLM answerer → shared judge | same, `--eval-lane industry-search` |

```bash
# MH-only remasure on a frozen integrity tenant
python3 -m public.locomo.run_smoke \
  --base-url http://127.0.0.1:18100 \
  --eval-lane product-recall \
  --fail-closed --skip-ingest \
  --tenant-prefix integrity-s0-1 \
  --conversations 10 \
  # restrict to MH in the smoke flags you already used

# S0 both lanes (do not start unless /runtime is clean)
python3 evals/public/locomo/run_s0.py \
  --base-url http://127.0.0.1:18100 \
  --fail-closed --skip-ingest \
  --tenant-prefix integrity-s0-1

# 3×90
# --conversations 3 --questions 90
```

Merge gates (every increment): `go test ./...`, OpMem 13/13, marketing 17/17, held-out compiler audits.

Lint/build: `go vet ./...`, `go build ./cmd/api ./cmd/worker`.

---

## 5. Runtime / integrity stack (do not reinvent)

Fail-closed flags (merged in #132):

- `BRAINY_EXTRACTION_STRICT`
- `BRAINY_EMBEDDING_STRICT`
- `BRAINY_EMBEDDING_DIMENSIONS=768`
- `BRAINY_REQUIRE_ANN`

Integrity env historically: `/tmp/integrity-stack.env`, API `:18100`, DB `brainy_integrity`. **Never start the API without that env** — default `BRAINY_MAX_BODY_BYTES=64` → HTTP 413 on ingest.

Before any remasure:

```bash
curl -s http://127.0.0.1:18100/runtime
```

Require: signatures match, dim **768 only**, ANN active, embed/extract **fallbacks 0**. If a hung S0 accumulated embed failures, **restart the API** to clear counters (fallbacks stay 0 but failure counts confuse you).

Extractor actually used: Cloudflare **gpt-oss-120b** via AI Gateway — not gpt-4o-mini. Embedder on the live remasure: hosted **BGE 768**, ANN on. Two stores exist (long-lived ~22,509 mems vs a rebuild ~22,481). Do not average their retrieval numbers.

**Process hygiene:** kill **specific PIDs** only. Never `pkill -f`. Stop the worker during `cmd/reembed`.

**Do not commit:** `127.0.0.1:18100` as a default, raw keys, gateway URLs with account IDs, or literal `[REDACTED]`.

Provider gotchas already burned:

- CF `/compat/embeddings` + `text-embedding-3-*` → 400
- CF `/openai` + CF token → 402
- Direct `api.openai.com` + `OPENAI_API_KEY` worked for the A/B
- urllib needs a `User-Agent` or Cloudflare 1010
- Full S0 180 on a cloud VM **stalled twice** on 120s embed timeouts. MH-only 33 succeeded without oracle probes.

Auth: if `BRAINY_API_KEYS` / `BRAINY_REQUIRE_API_KEY` are set, unauthenticated calls are 401. Local no-auth: unset those, `BRAINY_ENV=local`.

---

## 6. Kill list (hard)

- No fusion fishing / fusion-weight sweeps
- No unbounded top-k / episode stuffing to restore OD or SH
- No LoCoMo- or LME-named product rules or prompts
- No category dictionaries
- No graph DB default (Neo4j / FalkorDB). Postgres graph-shaped is ADR-004
- No new regex batch “just in case”
- No mixing pins (1×30 vs S0 vs 3×90 vs n=1540 vs Mem0 92.5%)
- No publishing 1×30 70% as full LoCoMo
- No LME-500 / BEAM-1M as a quality claim
- No SOTA / beats-Mem0 in README or GTM without frozen same-pin + approval
- No reopening R0–R10 as if missing
- No merging #133 until the ledger says WRITE
- No re-running OpenAI A/B for sport
- No waiting on the long-lived integrity VM (`bc-37d87fa2-…`, branch `pr/provider-embed-integrity-a6c7`) — #132 already merged

---

## 7. Docs the next agent should trust (in order)

1. **This file**
2. [competitive/cycle-closeout.md](./competitive/cycle-closeout.md) — **2026-08-20** then **2026-08-19/20 integrity**
3. [sota-execution-plan.md](./sota-execution-plan.md) — gates, not the S1-first guess
4. [locomo-full-70-80-path.md](./locomo-full-70-80-path.md)
5. [codebase-graph.md](./codebase-graph.md) — topology (dated 2026-08-04; planes are mid-migration)
6. [AGENTS.md](../../AGENTS.md) — public-docs voice, cycle-closeout SOP, cloud VM notes
7. Pins: [locomo-mh-packet-proof-20260820.md](../benchmarks/artifacts/locomo-mh-packet-proof-20260820.md), [embedding-ab-20260820.md](../benchmarks/artifacts/embedding-ab-20260820.md)

**Stale if treated as live start:**

- [external-agent-assessment-pack.md](./external-agent-assessment-pack.md) **CURRENT (2026-08-17)** — still useful architecture; pins and “next is R5A / R6a compiler” are **historical**. R5A–R10 landed. S0 ledger flipped the next lever to proof.
- [docs/research/README.md](./README.md) headline table — updated to point here; older “next R6a” sentences elsewhere are history.
- Wave 1 archaeology / Gate 0 / “next is R1b” / LME 0/20.

Every remasure or merge must add a dated section to cycle-closeout **in this order:** Landed → Own pins → Competitor compare (detailed) → Why → Next. Scores-only is incomplete. README gets published-% + same-pin **summary** only, outlinking [docs/benchmarks/README.md](../benchmarks/README.md). No SOTA. Trail axes stay visible (today: open-domain, and product MH).

---

## 8. Definition of done for the incoming goal

You are not done when a blog sentence says 80%. You are done when:

1. Fail-closed S0 product `/recall` moves for **attributed** proof-path reasons (ledger PROOF shrinks; WRITE stays small unless you can show new WRITE).
2. MH product is no longer a 2/33 dip — hop-plan coverage and `hop_join_proven` are measured, not vibed.
3. Stratified SH / temporal / OD are labeled. OD 0/4 stays a diagnostic; do not stuff episodes to fake it.
4. OpMem 13/13 and marketing 17/17 stay green.
5. Only then: 3×90 both lanes, then **one** n=1540 freeze, then Mem0 Platform same-pin on the identical harness.
6. If that same-pin wins: write the cycle-closeout + README same-pin table. Still do not write “SOTA” until the owner says so.

If you cannot show a stratified delta after one iteration, re-scope. Do not polish.
