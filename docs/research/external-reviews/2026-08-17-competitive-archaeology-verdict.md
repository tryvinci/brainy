# External review — competitive archaeology (pinned to Wave 1)

**Status:** **historical.** Source pin was Wave 1 `bd987fa`. Keep: do not re-queue P0-P7. **Live next-work** is [2026-08-17-parity-gap-verdict.md](./2026-08-17-parity-gap-verdict.md) (current SHA `8492ad3` / product `1b5ab3e`).

**Date:** 2026-08-17
**Source:** uploaded deep-research report ("Brainy Competitive Code Archaeology and Technical Gap Plan")
**Adjudicator:** coding agent on `pr/full-recall-dip-review-a6c7`
**Prompt used:** report was written against **`bd987fa`** (Wave 1, LoCoMo 1x30 **14/30**), with the 2026-08-17 dip self-review treated only as corroboration. Adjudicated against **current** product SHA `1b5ab3e` / docs `8492ad3` using [2026-08-17-full-recall-self-review-prompt.md](./2026-08-17-full-recall-self-review-prompt.md).

`bd987fa` is an ancestor of HEAD (59 commits behind this branch). Do not treat the report's "Brainy now" column as live.

## Verdict (1 paragraph)

**Keep** the representation-first course and the five-plane architecture. **Adjust** the incoming P0-P7 sequence: it is the Wave 1 plan that [sota-representation-path.md](../sota-representation-path.md) already executed as R0-R4. Do **not** re-queue a representation oracle, atomic compiler, fact-primary recall, relation store, or hop-join executor as if they were missing. **Accept** the two-lane evaluation rule (product `/recall` vs search+harness with a shared answerer) and the kill list (no fusion fishing, no graph DB, no LoCoMo-named rules, no SOTA). The single most important next product move is **R5 structured-first**: make `/recall` cite compiled facts with deterministic operators, not `firstStatementFromPacket` slogans. That is the 11.4% to ~50% lever and the 1x30 OD 0/4 trail. It is **not** "a better reader over chat transcripts." Remaining WRITE_MISS coverage is the ~50 to Mem0-class lever. Fix the latency histogram sum as a small parallel bug, not the cycle.

## Answers to prompt questions

1. **Course:** Keep representation-first. The report's thesis (facts for recall, relations for reasoning, episodes for provenance, projections for truth) is already PoR. Modify the closing slogan "do not spend the next cycle making a better reader over chat transcripts": agreed if that means LLM-over-episodes as the SOTA bet. Product `/recall` still has to *cite compiled facts*. That is structured-first (R5), not a transcript reader.

2. **Published metric (`/recall` vs search+harness):** **Accept two lanes.** (a) Memory-component: retrieved units to the same answerer/judge as Mem0 (isolates construction/retrieval). (b) Product: Brainy `POST /recall`. Do not describe 11.4% vs Mem0 92.5% as a pure retrieval comparison. README bake-off row stays product `/recall` (honest shipped path) with the path label; a future industry-format pin is search+harness, n=1540, labeled separately. 49.8% is that lane on an old stack, not current.

3. **First product PR (cite-facts vs R5 OD):** **Same family, not two bets.** Report P6: "deterministic value answers bypass reader where possible." That is cite-facts. R5 OD yes/no is the same packet contract. Land structured-first `/recall` first (mass is full single-hop 10.5%; 1x30 OD 0/4 is the same-pin trail). Failure class: ANSWER_PRESENT / READER_MISS only after FACT_PRESENT. Do not restore OD/SH by stuffing episodes into top-k.

4. **Fair Mem0 stack:** Two lanes; label Mem0 OSS vs Platform; shared answerer/judge; candidate budget separate from context-token budget; their harness cutoffs 10/20/50/200 and ~7K retrieval tokens are comparison points, not Brainy defaults. Full n=1540 only after structured-first + a freeze. Top-k 200 on *our* `/recall` is not a substitute for citing facts.

5. **Skipped suites (LME-500 / BEAM 1M):** **Accept skip.** Report P7 starts at LME-20 product `/recall` (already **4/20**) then LoCoMo 1x30 / 3x90 / multi-seed. LME-100/full only after infra stable. BEAM 1M is not in their immediate gate. Do not run LME-500 as a quality claim while LME-20 is 4/20.

6. **Next 3-7 PRs:** See accepted sequence below. Reject re-opening P0-P6 as missing substrate.

7. **Claims allowed vs forbidden:** Accept their forbidden list (SOTA, beats-Mem0/Zep, MH solved, LME-qualified before fail-closed, treating 92.5/94.4 as same-pin, OSS reproduces Platform, Graphiti requires Neo4j). Allowed: ops/vertical leads this freeze; 1x30 21/30 vs Mem0 11/30 with OD trail named; 11.4% as a named `/recall` dip vs July search+harness 49.8%.

8. **Kill list confirmation:** Confirmed. Plus: do not re-queue R0-R4; do not create a second `memory_facts` substrate in parallel with compiler facts / `memory_atoms` / `memory_relations` without a measured gap; do not another full remasure this cycle.

## Findings table

| Finding | Accept / Modify / Reject | Code evidence | Action |
| --- | --- | --- | --- |
| Five-plane + representation-first thesis | **Accept** | Already PoR in `sota-representation-path.md` | Keep |
| Pin the review at `bd987fa` | **Accept as source pin; reject as live** | `bd987fa` = Wave 1 14/30; 59 commits behind HEAD; remasure product `1b5ab3e` | Re-read live pins before any PR |
| P0 representation oracle still missing | **Reject (landed)** | `evals/public/stage_oracle.py`, `evals/public/oracle.py` (gold in episode-only store is WRITE_MISS; READER_MISS last) | Do not re-queue R0 |
| Attribute atoms tagged `primitive=episode` | **Reject (fixed)** | `TestAttributeAtomsIdentityAndOrigin` fails if attribute primitive is episode | Keep R1a |
| No relation table / hops are lexical | **Reject (landed as measurement)** | mig v20 `memory_relations`; `hopJoinProven` requires `hop[i].output_entity_id == hop[i+1].input_entity_id`; 1x30 MH 10/10 | Do not re-queue R3/R4; do not claim MH-solved (full MH 7.4%) |
| ContextEvidence / ProofChain missing types | **Reject** (report already said landed at `bd987fa`) | `EvidencePacket` | Observability only if a measured gap appears |
| Canonical `memory_entities` / `memory_facts` tables missing | **Modify** | Hub is still `memory_entity_links`; facts live as records/atoms + `memory_relations` projection, not the proposed UUID schema | R2 full IDs remain open; do not dual-write a second fact table as the next PR |
| Fact-primary + episode fallback missing | **Reject (landed)** | `trace.go` `representation_status`, `episode_fallback` | Keep coverage-gated fallback |
| Product `/recall` vs Mem0 search+harness is not a retrieval compare | **Accept** | Full pin 11.4% is `brainy-recall+*`; July 49.8% was search+harness | Two lanes; path labels stay |
| Next cycle = better reader over transcripts | **Modify** | `recall.go` `firstStatementFromPacket` still emits slogans (`conv-26-q83`) | R5 structured-first / deterministic cite-facts, not LLM-over-episodes |
| P6 deterministic operators for scalar/date/entity/list | **Accept** (this is the product PR) | `firstStatementFromPacket`, list-cue enumerate, 188 abstains | Land as R5 |
| Two competitive lanes | **Accept** | Report eval section; dip why already named the path change | Pin both when we remasure; do not mix on README |
| Separate candidate budget from context tokens | **Accept** | PR4 already has MaxEvidenceTokens + pool cap 200 | Keep; 7K is a Mem0 compare point, not a default |
| Skip LME-500 / BEAM 1M now | **Accept** | LME-20 4/20; BEAM 100K 8/20 | Unchanged |
| `latencyHistogram.observe` adds float bit patterns | **Accept** (still true) | `internal/observability/metrics.go` `sum.Add(math.Float64bits(seconds))` | Small parallel fix; not the cycle |
| Competitive archaeology as standing SOP | **Accept** (mostly present) | `docs/research/competitive/` already has README, mem0.md, graphiti.md, borrow log, gap matrix | Optional later: `zep-platform.md`, yaml matrix, benchmark-config-matrix. Not blocking |
| Tenant/subject-scoped entity identity; suppression cascade | **Accept** | Existing isolation + evidence plane | Keep as constraints on any R2 IDs work |
| Proposed P7 freeze then LME-20 / 1x30 / 3x90 / full | **Modify** | Remasure already ran those (except 3x90 / LME-500) | Remasure *after* R5, with two lanes; not before |
| Copy Mem0 OSS scoring constants / Graphiti rerankers | **Reject** | Report itself says do not cargo-cult | Kill list |
| Graph DB / Neo4j | **Reject** | ADR-004; report agrees | Kill list |
| Recreate P0-P7 as the next engineering cycle | **Reject** | R0-R4 landed; 1x30 21/30 vs Wave 1 14/30 | Next is R5 + remaining WRITE_MISS |

## Accepted next sequence

1. **R5 structured-first / cite compiled facts** — ANSWER_PRESENT. `/recall` deterministic operators from structured values; source text is provenance. Exit: 1x30 OD moves without episode stuffing; full `/recall` single-hop no longer slogan-dominated on items where a fact exists. Artifact: 1x30 + sampled full `/recall` (not a new n=1540 quality claim).
2. **Remaining WRITE_MISS compiler coverage** — career/possession/dated plans as facts. Exit: representation audit on held-out OD/SH items. Lever: ~50 to Mem0-class on the *component* lane.
3. **R2 full canonical entity IDs** (aliases, ranked resolution) when hops/coverage show identity misses — ENTITY_LINK_MISS. Not ahead of R5.
4. **Latency histogram sum** — observability bug. Unit test that two observations add as floats.
5. **R6 two-lane remasure** after a freeze: product `/recall` *and* search+harness, same SHA/judge; LoCoMo 1x30 diagnostic then 3x90; LME-20 quality; full n=1540 labeled by path. Keep OpMem 13/13 and marketing 17/17 green.

## Rejected / deferred

- Re-queue P0 oracle, P1 compiler, P2 fact-primary, P5 relation store, P6 hop executor as missing
- New `memory_facts` / `memory_entities` UUID schema as the next PR (adapt later if R2/R5 prove the current tables insufficient)
- `competitive-gap-matrix.yaml` / `benchmark-config-matrix.md` / `zep-platform.md` as blocking docs
- Candidate-budget sweep / temporal-score grid as this cycle's work
- LME-500 or BEAM 1M as a quality claim
- Another full remasure before R5
- LLM-over-search as a silent restore of 49.8% as "current product"
- Fusion fishing, graph DB, category dictionaries, LoCoMo-named product rules, SOTA / beats-Mem0

## Claims discipline check

- [x] No invented LME / SOTA / BEAM 1M scores
- [x] Gate 0 vs harden vs remasure pins not blended (`bd987fa` labeled Wave 1)
- [x] 1x30 70% not published as full LoCoMo
- [x] 11.4% labeled product `/recall`; 49.8% labeled search+harness (not current)
- [x] Vendor 90+ labeled n / metric / top-k / path
- [x] Dip honesty preserved where scores fell

## Artifact diffs required

- [x] assessment pack
- [ ] codebase graph md/json (no structural change this pass)
- [x] program-execution-status
- [x] PoR adoption note (`sota-representation-path.md` tips + this archive)
- [x] external-reviews/README.md priority pointer

## Linked PRs / commits

- Adjudication lands on `pr/full-recall-dip-review-a6c7` (GitHub PR targeting `dev`)
- Report source pin: `bd987fa` (Wave 1)
- Live product pin: `1b5ab3e`
- ENG-168 remains the conversational epic; this does not reopen R0-R4
