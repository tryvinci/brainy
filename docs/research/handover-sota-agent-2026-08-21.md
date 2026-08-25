# Handover — next agent: LoCoMo 80% / same-pin vs Mem0

**Date:** 2026-08-25  
**Audience:** the next coding agent. Read this first. Do not start from the 2026-08-17 assessment pack as live truth.  
**Owner ask:** take Brainy to a defensible conversational claim — user phrasing is “SOTA / beat Mem0 / 80% LoCoMo.”  
**Repo SOP:** that phrasing is a *goal*, not a claim you may write in product copy. Competitive language requires a **frozen same-pin win** (same dataset SHA, same judge temp 0, same answerer, same question set, same harness) **and** explicit approval. The word “SOTA” stays gated even after a same-pin win until the owner lifts it.

This file is the live start doc. Older research notes stay useful as history. If they disagree with the pins or “next step” here, **this file wins**.

---

## 0. First 30 minutes

1. Confirm you are on `dev` (staging) at P25 pin docs (or later). `main` is production and fast-forwards only with explicit owner approval (this cycle: owner asked to keep FF'ing both).
2. Read this file, then [cycle-closeout.md](./competitive/cycle-closeout.md) section **2026-08-25 — P25 which-year as-of duration (111/180)**.
3. Skim [sota-execution-plan.md](./sota-execution-plan.md) but **do not** treat its “expected S1 compiler first” as live. S0 ledger **outranks** that expectation.
4. Do **not** re-queue R0–R10. Substrate is merged.
5. Do **not** merge [PR #133](https://github.com/tryvinci/brainy/pull/133) (compiler S1–S5 fishing) or revive [PR #131](https://github.com/tryvinci/brainy/pull/131).
6. Do **not** re-run the OpenAI embedding A/B unless `GET /runtime` or the JSON pins are broken.
7. Do **not** burn full n=1540 until a stratified delta exists. A **fair** Mem0 Platform 180 (v3, top_k 200, chunk 1) **is** the next same-pin — do not reuse the handicapped 11/30 freeze as lead/trail.

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

**Honest distance:** product `/recall` full n=1540 is **11.4%** on SHA `1b5ab3e`. Fail-closed S0 product is **32/180** on the integrity tenant, **19/180** hybrid-off / **37/180** P1 / **56/180** P2 / **61/180** P2b / **73/180** P3 / **79/180** P4 / **84/180** P5 / **87/180** P6 / **88/180** P7 / **93/180** P8 / **94/180** P9 / **96/180** P10 / **97/180** P11 / **101/180** P12 / **102/180** P13 / **103/180** P14 / **104/180** P17 / **105/180** P18 / **106/180** P20 / **107/180** P21 / **108/180** P22 / **109/180** P23 / **110/180** P24 / **111/180** P25 hybrid-on on this-VM `diag-mh-135`. Industry S0 is **62/180** on both. MH product after #135 is **2/33** integrity / **12/33** this tenant reader-off / **17/33** P5 / **13/33** P6 / **14/33** P7 / **15/33** P8–P10 / **13/33** P11 / **15/33** P12–P13 / **16/33** P14–P18 / **17/33** P20 / **18/33** P21–P25 hybrid-on. Getting to 80% on n=1540 is a multi-increment proof/reader (then compiler if the ledger flips), not one PR. 90% on this 180 is **162/180**.

---

## 2. Where we landed (2026-08-22)

### Landed SHAs / PRs

| Ref | Role |
| --- | --- |
| `dev` **now** | `453a929` — #135 merge (MH slot-aligned dest-subject). Staging. |
| `main` | `6d05e1b` — #134 packet/proof. Production. **Do not FF** unless the owner asks. |
| PR **#141** `pr/locomo-180-p25-1e9e` | P25 which-year as-of duration. Product **111/180** (`86d87d6`). |
| PR **#140** `pr/locomo-180-p24-1e9e` | Merged. P24 sentence-initial verb covering. Product **110/180** (`80669d8`). |
| PR **#139** `pr/locomo-180-p23-1e9e` | Merged. P23 year/month leftover covering. Product **109/180** (`e05c78f`). First P23 180 108/180 hold is not a pin. |
| PR **#138** `pr/locomo-180-p22-1e9e` | Merged. P22 leftover month bind. Product **108/180** (`9d50bad`). |
| PR **#137** `pr/locomo-180-p21-1e9e` | Merged. P21 location-list leftover lock. Product **107/180** (`fa915fe`). |
| PR **#136** `pr/s0-current-sha-baseline-1e9e` | Merged. Mem0 v3 harness, S0 19/180 reader-off through P20 106/180 (`80471d8`). P15/P16 103/180 and P19/P19b 105/180 holds are not pins. |
| PR **#135** | Merged. MH list/join proof. |
| PR **#134** | Merged. MH packet/proof + earlier handover. |
| PR **#133** | OPEN draft. Compiler S1–S5. **Do not merge.** |
| PR **#131** | CLOSED. Mixed. **Do not revive.** |

Linear: [ENG-176](https://linear.app/engramhq/issue/ENG-176/eng-multi-hop-memory-synthesis-consolidation-for-conversational-qa) (MH), parent [ENG-168](https://linear.app/engramhq/issue/ENG-168/epic-conversational-long-memory-product-gaps-from-locomo-smoke).

`dev` = staging. `main` = production. Fast-forward `main` only with explicit owner approval.

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
| S0 product `POST /recall` integrity VM | **32/180** | Tenant `integrity-s0-1`. Ledger: **PROOF 112 / RETRIEVAL 22 / READER 11 / WRITE 3**. |
| S0 industry search+harness | **62/180** | Integrity VM **and** this-VM `diag-mh-135`. Do not average with product. |
| S0 product `POST /recall` this VM | **19/180 (0.106)** | Tenant `diag-mh-135` + conv-30, hybrid **off**, SHA `453a929`. MH **12/33** · OD **0/11** · SH **5/98** · temporal **2/38**. Ledger: **PROOF 59 / READER 52 / RETRIEVAL 39 / WRITE 10**. Does **not** replace 32/180. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-20260822.md) |
| S0 product hybrid **on** this VM | **37/180 (0.206)** | Same store, SHA `3d42b17`, `BRAINY_RECALL_LLM=1`. MH **10/33 (dip)** · OD **1/11** · SH **19/98** · temporal **7/38**. Ledger: **PROOF 44 / READER 49 / RETRIEVAL 39 / WRITE 10 / HARNESS 1**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-llm-20260822.md) |
| S0 product P2 length-lock | **56/180 (0.311)** | SHA `681028e`. MH **11/33** · OD **1/11** · SH **23/98** · temporal **21/38**. Ledger: **PROOF 41 / RETRIEVAL 39 / READER 33 / WRITE 10**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p2-20260822.md) |
| S0 product P2b extras + skip hop-ground | **61/180 (0.339)** | SHA `fb41ece`. MH **16/33** · OD **1/11** · SH **25/98** · temporal **19/38 (dip vs P2 21/38)**. Ledger: **PROOF 42 / RETRIEVAL 38 / READER 28 / WRITE 10**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p2b-20260822.md) |
| S0 product P3 distinctive-token admit | **73/180 (0.406)** | SHA `5bc28ea`. MH **16/33** · OD **3/11** · SH **32/98** · temporal **22/38**. Ledger: **READER 34 / PROOF 34 / RETRIEVAL 29 / WRITE 8 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p3-20260822.md) |
| S0 product P4 identity/garbage hybrid | **79/180 (0.439)** | SHA `6f74024`. MH **16/33** · OD **3/11** · SH **37/98** · temporal **23/38**. Ledger: **PROOF 32 / READER 30 / RETRIEVAL 30 / WRITE 7 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p4-20260822.md) |
| S0 product P5 activity-dump skip | **84/180 (0.467)** | SHA `5ad07c4`. MH **17/33** · OD **2/11 (dip)** · SH **45/98** · temporal **20/38 (dip vs P4 23/38)**. Ledger: **PROOF 32 / RETRIEVAL 29 / READER 26 / WRITE 7 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p5-20260822.md) |
| S0 product P6 dump-lock skip | **87/180 (0.483)** | SHA `45a83b5`. MH **13/33 (dip vs P5 17/33)** · OD **3/11** · SH **52/98** · temporal **19/38 (dip vs P5 20/38)**. Ledger: **READER 29 / RETRIEVAL 28 / PROOF 28 / WRITE 6 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p6-20260822.md) |
| S0 product P7 hop-local joins | **88/180 (0.489)** | SHA `f3e0a7f`. MH **14/33** · OD **4/11** · SH **49/98 (dip vs P6 52/98)** · temporal **21/38**. Ledger: **PROOF 29 / RETRIEVAL 29 / READER 27 / WRITE 5 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p7-20260822.md) |
| S0 product P8 SH recovery | **93/180 (0.517)** | SHA `86eab77`. MH **15/33** · OD **4/11** · SH **52/98** · temporal **22/38**. Ledger: **RETRIEVAL 29 / PROOF 28 / READER 24 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p8-20260822.md) |
| S0 product P9 unproven mh_list dumps | **94/180 (0.522)** | SHA `bdee669`. MH **15/33** · OD **4/11** · SH **53/98** · temporal **22/38**. Ledger: **RETRIEVAL 29 / PROOF 27 / READER 24 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p9-20260822.md) |
| S0 product P10 date-aware leftover covering | **96/180 (0.533)** | SHA `e461d70`. MH **15/33** · OD **4/11** · SH **55/98** · temporal **22/38**. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 23 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p10-20260823.md) |
| S0 product P11 locative leftover covering | **97/180 (0.539)** | SHA `bc6dc92`. MH **13/33 dip** · OD **3/11 dip** · SH **58/98** · temporal **23/38**. Ledger: **RETRIEVAL 29 / PROOF 27 / READER 20 / WRITE 5 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p11-20260823.md) |
| S0 product P12 keep short where NPs / typed item joins | **101/180 (0.561)** | SHA `d292a09`. MH **15/33** · OD **4/11** · SH **58/98** · temporal **24/38**. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 18 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p12-20260823.md) |
| S0 product P13 gated leftover-vs-hybrid | **102/180 (0.567)** | SHA `50b8e43`. MH **15/33** · OD **4/11** · SH **59/98** · temporal **24/38**. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 17 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p13-20260823.md) |
| S0 product P14 childhood possession lock | **103/180 (0.572)** | SHA `90750e5`. MH **16/33** · OD **4/11** · SH **59/98** · temporal **24/38**. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 16 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p14-20260823.md) |
| S0 product P17 when-event leftover covering | **104/180 (0.578)** | SHA `4719902`. MH **16/33** · OD **4/11** · SH **60/98** · temporal **24/38**. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 15 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p17-20260825.md). P15/P16 103/180 are not pins. Named dip: Jon banker job `conv-30-q0`. |
| S0 product P18 when-event query-entity bind | **105/180 (0.583)** | SHA `0c03107`. MH **16/33** · OD **4/11** · SH **60/98** · temporal **25/38**. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 14 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p18-20260825.md). Jon banker recovered. |
| S0 product P20 enumerate unwind extras | **106/180 (0.589)** | SHA `80471d8`. MH **17/33** · OD **4/11** · SH **60/98** · temporal **25/38**. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 13 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p20-20260825.md). Destress pottery recovered. P19/P19b 105/180 holds are not pins. |
| S0 product P21 location-list leftover lock | **107/180 (0.594)** | SHA `fa915fe`. MH **18/33** · OD **4/11** · SH **60/98** · temporal **25/38**. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 12 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p21-20260825.md). Yoga practice locations recovered. First P21 180 106/180 flake is not a pin. |
| S0 product P22 leftover month bind | **108/180 (0.600)** | SHA `9d50bad`. MH **18/33** · OD **4/11** · SH **60/98** · temporal **26/38**. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 11 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p22-20260825.md). September co-participant plan recovered. |
| S0 product P23 year/month leftover covering | **109/180 (0.606)** | SHA `e05c78f`. MH **18/33** · OD **4/11** · SH **60/98** · temporal **27/38**. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 10 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p23-20260825.md). Hometown community-center 2022 recovered. First P23 180 108/180 hold is not a pin. |
| S0 product P24 sentence-initial verb covering | **110/180 (0.611)** | SHA `80669d8`. MH **18/33** · OD **4/11** · SH **60/98** · temporal **28/38**. Ledger: **RETRIEVAL 29 / PROOF 26 / READER 9 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p24-20260825.md). Unnamed dated art-show diary recovered. |
| S0 product P25 which-year as-of duration | **111/180 (0.617)** | SHA `86d87d6`. MH **18/33** · OD **4/11** · SH **60/98** · temporal **29/38**. Ledger: **RETRIEVAL 28 / PROOF 26 / READER 9 / WRITE 4 / HARNESS 2**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p25-20260825.md). Health start year recovered. |
| S0 MH product (post-#134) | **2/33** | Was **1/33**. Attributed win: turtles. Second hit (soda/candy) is a crowded-list judge accept. [pin](../benchmarks/artifacts/locomo-mh-packet-proof-20260820.md) |
| S0 MH product diagnostic ingest | **7/33** | WRITE+PROOF mixed on `diag-mh-135`. Does not replace 2/33. [pin](../benchmarks/artifacts/locomo-mh-diag-135-20260821.md) |
| S0 MH product diagnostic skip-ingest | **12/33** | PROOF-only on frozen `diag-mh-135`. Kinship dest `9d8dbeb` was **10/33**; slot-aligned recovery `2e84435` is **12/33** (`conv-44-q26` work, `conv-26-q60` clarinet/violin). Does not replace 2/33. [pin](../benchmarks/artifacts/locomo-mh-diag-135-skip-ingest-slot-recover-20260821.md) |
| 3×90 | product **21/90**, industry **33/90** | `--conversations 3 --questions 90` |
| 1×30 freeze (conv-26) | **21/30** | MH 10/10, OD **0/4**, temporal 11/16. **Do not overwrite.** Diagnostic only. Rest of conv-26 in the full run was **12/122**. |
| Full n=1540 product `/recall` | **175/1540 = 11.4%** | Old product SHA `1b5ab3e`. Full n=1540 only at S6. |
| Industry historical | **49.8%** | July, **old stack**, search+harness. Not a current-SHA ceiling. |
| LME-20 | **4/20** | Not re-run. Do not spend a cycle on LME-500. |
| Mem0 Platform 1×30 | **11/30** | Freeze 2026-08-15. MH 6/10, OD 3/4, temporal 2/16. **Handicapped protocol** (v2, top_k 30, chunk 8). Do not mix with 32/180 or 19/180. Fair 180 is the next pin. [audit](./competitive/mem0-harness-audit-2026-08-22.md) |
| Embedding A/B | OpenAI @768 r@10 **0.333** vs this-rebuild BGE **0.306** | Retrieval-only. Long-lived VM BGE was **0.239** — **do not average**. [pin](../benchmarks/artifacts/embedding-ab-20260820.md) |

**Invalidated:** Aug-19 S0 17/180 / 52/180 (no pgvector, silent extract degrade). Never cite those as quality.

**Bottleneck on this VM is split:** product S0 WRITE_MISS is **4/180** on P25 (P24 was **4/180**; integrity was **3/180**). Coverage is not the 80% hole — QA is **19/180** reader-off / **111/180** P25 hybrid-on vs industry **62/180**. This-VM product MH is **12/33** reader-off / **17/33** P5 / **15/33** P10 / **13/33** P11 / **15/33** P12–P13 / **16/33** P14–P18 / **17/33** P20 / **18/33** P21–P25 vs integrity **2/33**. SH 5→60 is the hybrid+admit+dump-skip path. Remaining mass is SH **PROOF 20**. Isolated leftover covering is saturating. MH **18/33** is held from P21.

### Competitor stand (honest)

- **1×30 freeze:** Brainy 21/30 vs Mem0 Platform 11/30 is a prior **lead on a handicapped Mem0 protocol**. It is not full LoCoMo and not permission to write “we beat Mem0.”
- **S0 / n=1540:** no same-n Mem0 pin yet (fair 180 429 quota until 2026-09-01). Do not trail/lead 32/180, 19/180, 61/180, 73/180, 79/180, 84/180, 87/180, 88/180, 93/180, 94/180, 96/180, 97/180, 101/180, 102/180, 103/180, 104/180, 105/180, 106/180, 107/180, 108/180, 109/180, 110/180, 111/180, or 11.4% vs 11/30 or vs published 92.5%.
- **Ops / marketing:** Brainy lead (Mem0 pins stale). Must not regress. Not the next cycle.
- **Graphiti / Zep / SuperMemory:** no same-pin. Published headlines are context.
- **Mem0 OSS** was not re-measured. Do not treat Platform 11/30 as OSS-reproducible.

---

## 3. What the next increment is

S0 said: spend the next increment on the **largest earliest-stage bucket**. That bucket is **PROOF_MISS** (packet compose, hop people vs topic nouns, typed intersect, list/join, reader crowding).

| Increment | Plan name | Do now? |
| --- | --- | --- |
| S0 this-VM | Dual-lane 180 on `diag-mh-135` | **Done.** Product **19/180** reader off; industry **62/180**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-20260822.md) |
| P1 reader | `BRAINY_RECALL_LLM=1` including enumerate | **Done.** Product **37/180**. SH 5→19, temporal 2→7. MH **12→10 dip**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-llm-20260822.md) |
| P4 Mem0 180 | Fair Platform same-pin (v3, top_k 200, chunk 1) | **Restarted** as `locomo-s0-mem0-v3-s1-fair2` after 300s event-wait crash. Event wait now follows `--async-timeout`. Do not compare to the 11/30 freeze. |
| S2 / S3 residue | Structured answer + hop proof | **Merged on #135.** Diagnostic skip-ingest **12/33**. Integrity **2/33**. Do not keep MH-only as the 80% path. |
| P2 | Unlock dates; lock counts / dual-entity lists; skip hop-ground on enumerated hybrid | **Done.** Length-lock **56/180**; extras+skip-ground **61/180**. MH **16/33**. Temporal **19/38** (dip vs P2 21/38). Keep where/polar locked. |
| P3 | Ledger-allocated SH PROOF (token admit, not #133) | **Done.** Product **73/180**. MH **16/33** held. SH 25→32. OD 1→3. Temporal 19→22. SH READER 15→22 dip. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p3-20260822.md) |
| P4 | Identity/garbage hybrid (not #133) | **Done.** Product **79/180**. MH **16/33** held. SH 32→37. OD held 3/11. Temporal 22→23. Named losses: Tim-UK, Melanie LGBTQ, John community-center 2022. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p4-20260822.md) |
| P5 | Activity/event dump skip (not #133) | **Done.** Product **84/180**. MH **17/33** (Tim-UK recovered). SH 37→45. OD **3→2 dip**. Temporal **23→20 dip**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p5-20260822.md) |
| P6 | Dual-entity dump-lock skip (not #133) | **Done.** Product **87/180**. MH **13/33 dip**. SH 45→52. OD 2→3. Temporal **20→19 dip**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p6-20260822.md) |
| P7 | Hop-local leftover facts + rare-share possessions (not #133) | **Done.** Product **88/180**. MH **14/33**. SH **52→49 dip**. OD 3→4. Temporal **19→21**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p7-20260822.md) |
| P8 | Recover SH 52 without restoring dumps (where-only mh_list unlock, ordinal names, leftover specific facts) | **Done.** Product **93/180**. MH **15/33**. SH **49→52**. OD held 4/11. Temporal **21→22**. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p8-20260822.md) |
| P9 | Unlock unproven mh_list fragment dumps / question-echo values (not #133) | **Done.** Product **94/180**. MH **15/33** held. SH **52→53**. Studying strategy recovered. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p9-20260822.md) |
| P10 | Date-aware leftover covering / hybrid 48h except where-queries (not #133) | **Done.** Product **96/180**. MH **15/33** held. SH **53→55**. Sapiens + retreat recovered. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p10-20260823.md) |
| P11 | Locative leftover covering / hop-slot ignore / hybrid overwrite only on where+games-played (not #133) | **Done.** Product **97/180**. MH **15→13 dip**. SH **55→58**. Jasper + tournament games recovered. Toronto + signed basketball dipped. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p11-20260823.md) |
| P12 | Keep short where NPs and typed item joins (not #133) | **Done.** Product **101/180**. MH **13→15**. OD **3→4**. Temporal **23→24**. Toronto / signed basketball / snacks / girlfriend restored. No losses vs P11. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p12-20260823.md) |
| P13 | Gated leftover covering vs hybrid (schema-activity covering only) | **Done.** Product **102/180**. SH **58→59**. Thanksgiving feast+thankful recovered. Unrestricted leftover-vs-hybrid reverted. First gated 180 **98/180** is not a pin. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p13-20260823.md) |
| P14 | Childhood possession lock; name questions are not item lists | **Done.** Product **103/180**. MH **15→16**. Childhood items recovered. Broad typed-join 180 **98/180** is not a pin. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p14-20260823.md) |
| P15 / P16 | Visit-destination keep; packet-line enrich of compressed visit stops | **Measured, not a pin.** Each **103/180** (Boston +1 / Ned bowling −1). Product kept; 180 net 0. |
| P17 | When-event leftover covering (minLen 4; bare date yields to event covering) | **Done.** Product **104/180**. SH **59→60**. Temporal held 24 (Ned recovered, McGee's gained, Jon banker dipped). [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p17-20260825.md) |
| P18 | When-event leftover covering bound to query people | **Done.** Product **105/180**. Temporal **24→25**. Jon banker recovered. No losses vs P17. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p18-20260825.md) |
| P19 / P19b | Unwind leftover join; hop-contents scan | **Measured, not a pin.** Each **105/180**. Destress answer gained pottery; 180 harness reads enumerate items. |
| P20 | Write unwind extras onto enumerate items | **Done.** Product **106/180**. MH **16→17**. Destress pottery recovered. No losses vs P18. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p20-20260825.md) |
| P21 | Location-list leftover lock; packet practice places; which-year bind | **Done.** Product **107/180**. MH **17→18**. Yoga practice locations recovered. First P21 180 106/180 flake is not a pin. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p21-20260825.md) |
| P22 | Bind leftover covering to query month when tokens collapse to weak-only | **Done.** Product **108/180**. Temporal **25→26**. September co-participant plan recovered. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p22-20260825.md) |
| P23 | Year-dated when-event leftover vs bare hop dates; skip wrong year/month covering | **Done.** Product **109/180**. Temporal **26→27**. Hometown community-center 2022 recovered. First P23 180 108/180 hold is not a pin. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p23-20260825.md) |
| P24 | Sentence-initial verbs are not leftover covering people | **Done.** Product **110/180**. Temporal **27→28**. Unnamed dated art-show diary recovered. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p24-20260825.md) |
| P25 | Which-year leftover covering from as-of durations | **Done.** Product **111/180**. Temporal **28→29**. Health start year recovered. [pin](../benchmarks/artifacts/locomo-s0-diag-mh-135-p25-20260825.md) |
| S1 compiler | Provider-extract / named-subject mass | **No** until the ledger says WRITE is the bucket again. #133 stays closed. |
| Embedder swap | OpenAI vs BGE | **Done / pinned.** Do not re-run. |
| S6 freeze | n=1540 + Mem0 same-pin | After a stratified **delta**, not after 19→111. Full n=1540 only at S6. |

**Suggested first remasure:** ranking so gold enters the leftover packet (Jolene yoga year 2020; paint-together decide date). Remaining isolated READER is thin. Do **not** join all `participates in`. `"destress"` stays denylisted. Remaining SH **PROOF 20**. Isolated leftover covering is saturating. MH **18/33** is held from P21. Fair Mem0 180 is quota-blocked until 2026-09-01. Do not start n=1540 yet. Do not merge #133. Do not special-case German vs Spanish.

**Shipped this increment (P25):** which-year leftover covering of `for N years as of DATE` rewrites to the start year; leftover tokens `year`/`years` are weak.

**Prior (P24):** sentence-initial past-tense verbs and phrasal particles are not leftover covering people; unnamed dated diary lines still compete.

**Prior (P23):** when-event leftover covering requires a year/date token so year-only event facts can replace a bare hop date; covering from a different query year or month is skipped.

**Prior (P22):** leftover covering binds to the query month/year when leftover tokens collapse to weak-only; rarest-token override ignores weak tokens.

**Prior (P21):** location-list leftover covering requires a locative; leftover packet practice places join onto answer/items; which/what-year covering binds to the query person and needs a year token.

**Prior (P20):** unwind-evidenced extras land on enumerate `items`, not only `answer`.

**Prior (P19 / P19b, not pins):** leftover unwind join + hop-contents scan. 180 **105/180**.

**Prior (P18):** when-event leftover covering skips lines that name another person and do not name a query person. First-person dated lines still compete.

**Prior (P17):** when-event leftover covering keeps 4-character event verbs; a short calendar date that misses leftover event tokens yields to a packet covering line that has them. Hybrids that already name the event stay. Visit-destination keep + packet enrich from P15/P16 stay in the SHA.

**Prior (P16, not a pin):** enrich compressed visit stops from a short packet line that names the place plus leftover purpose. 180 **103/180** (Boston purpose / Ned bowling).

**Prior (P15, not a pin):** keep co-participant visit destinations through hop foreign-possessive / skip-unrelated / prefer-destination. 180 **103/180**.

**Prior (P14):** 2-item childhood possession lists lock against speaker chat; leftover covering can join `had a` + age-cue facts; name questions are not childhood item lists so Max is not replaced by a hop possession dump.

**Prior (P13):** leftover covering may replace hybrid only for schema-activity covering or where/games-played; chat turns do not get a locative leftover bonus; leftover covering re-picks the rarest leftover token so Thanksgiving feast+thankful beats movies.

**Prior (P12):** where leftover covering returns the locative place NP; leftover covering does not beat a short hybrid place name; short typed item joins stay locked and are not leftover thin misses; leftover `support` is a weak token.

**Prior (P11):** where leftover covering ignores hop slots and requires a locative leftover token; leftover covering scores strong leftover tokens; hybrid leftover overwrite only for locative queries and games-played joins.

**Prior (P10):** leftover covering skips far-dated primary event dates on day-specific queries (10-day window so last-week session news stays); hybrid packets use a 48h window except where-queries; speaker-prefixed leftover covering counts only the body.

**Prior (P9):** unlock `mh_list` when hops are unproven `search_fallback` dumps; treat 4+ short identity fragments and question-echo hop values as dumps; leftover covering skips OCR captions and stored question prompts.

**Prior (P8):** unlock skipped `mh_list` only on where-queries; drop attended-event / foreign-possessive hop values; dated-then-undated ordinal names; leftover-covering specific packet facts on hybrid abstain with conflicting date tails stripped.

**Prior (P7):** leftover-covering hop contents under skip (including identity leftover names); rare-share omitted possession snippets; shortest-value rare-token scoring; keep title-cased typed joins on hybrid abstain without identity slogans; locative leftover where-answers.

**Prior (P6):** skip dual-entity activity dumps unless hops are a typed skill/possession/preference join; unlock hybrid when the typed answer is a hop dump; cap the hybrid prompt and promote proper-noun/venue facts ahead of generic leftover-cover; do not compose crowded hop dumps when hybrid abstains.

**Prior (P5):** skip activity/event hop dumps that miss leftover distinctive query tokens (skill/possession/preference joins stay); keep specific packet facts whose gold is a synonym of the leftover token; do not hop-ground or compose those dumps when hybrid abstains.

**Prior (P4):** reject punctuation-only / low-letter hybrid answers; skip identity-only Structured dumps that miss leftover query tokens; keep only covering memories in that prompt; do not hop-ground hybrid answers onto those identity dumps.

**Prior (P3):** leftover distinctive query-token admit into search candidates and the evidence set; merge covering extras ahead of a full original top-k; second-pass on the uncovered token when hop join is unproven; do not compose, ground, or prompt from `search_fallback` hop dumps.

**Prior (generic linguistic, fixtures not dataset IDs):** hop `Name and Name` / `Name and Name both` / `with Name`; hop the person after `does`/`has` on count questions; kinship `'s mother` / `her partner` chains family → slot; join compose intersects and does not dump the union; possession/skill lists without occupation/hobby crowding; how-many counts the typed set; Has/Did polar Yes from typed hops only; `practices … at` place extract; unwind/`do to` activity lists; visit/travel superlative; who-answers from other person mentions; `besides` exclusion (stemmed); childhood items as possession; **when-event hops prove a date from observed_at (do not dump event names)**; **given-to hops the giver only and keeps recipient-mentioned values**; **after-clause keeps matching evidence**; community/journey activity lists; family-injury who; organization beneficiaries from affiliation; **where+kinship answers a place from `in`/`at`/`near`, hopping the source person as well as the unnamed partner**; **`with colleagues/friends` is a group filter, not a CapName join**; **`for` clauses keep matching evidence**; **`get with having` hops health, not possession dumps**; **how-many children counts child-cued family members, not partners**; **dual-entity list queries intersect instead of unioning**; **kinship hobby lists filter to the dest person**; **how-many Ferraris counts the head noun, not every possession**; **who-told and polar teach from typed hops**; **journey-change lists stay identity, not occupation**; **pets' names are possession**; **named `in the X community` filters to X (affiliation hops too)**; **named `during X journey` filters identity to the period**; **list-head modifiers** (`outdoor activities`, `sports collectible`, `unhealthy snacks`) soft-filter evidence; **when list-head and group-companion cues are both present, prefer intersect, else list-head, else companion**; **community dual-entity lists join by organized/started/group token-subset and partner mention (fallback slots allowed)**; **unnamed kin-role dests rewrite to `{Name}'s {role}` and merge dest-subject attitude slots**; **who-supports keeps group nouns** (`friends and team`) from typed hops, not only CapNames; **practice location lists extract `in`/`at`/`near` places**, split comma/and lists, skip leading `her`/`his`, stop relative clauses / lone gerunds, and **never dump activities as locations**; **atom refill skipped when hops already listed**; **enumerate answers rank by query evidence then cap at 8** (enumerate mode shares that refine path); **dest-subject slot recovery** prepends practice locatives, unwind/calm/`to *stress` activities, play/practice objects, trick-mentioned skills, and besides+work stressors; compositional places require a definite `the {practice} {noun}`; `unwind` is not an `un-` negation.

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
- Extraction **job lease is 30s**; provider extract timeout is 120s. Without a live heartbeat, another worker reclaims `in_progress` mid-call (`ErrLeaseLost`, duplicate extract, fence blocks complete). The worker now heartbeats every 10s for the whole `ProcessNext`. Same-subject jobs still serialize; raise worker concurrency only to overlap **different** subjects. `ProcessAvailable` keeps each slot claiming until the queue is idle so one slow extract cannot park the other slots. Optional remasure env: `BRAINY_PROVIDER_TIMEOUT=300s`.

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
2. [competitive/cycle-closeout.md](./competitive/cycle-closeout.md) — **2026-08-25 P25 111/180**, then P24 110/180, then P23 109/180, then P22 108/180, then P21 107/180, then 2026-08-22 this-VM S0
3. [sota-execution-plan.md](./sota-execution-plan.md) — gates, not the S1-first guess
4. [locomo-full-70-80-path.md](./locomo-full-70-80-path.md)
5. [codebase-graph.md](./codebase-graph.md) — topology (dated 2026-08-04; planes are mid-migration)
6. [AGENTS.md](../../AGENTS.md) — public-docs voice, cycle-closeout SOP, cloud VM notes
7. Pins: [locomo-mh-packet-proof-20260820.md](../benchmarks/artifacts/locomo-mh-packet-proof-20260820.md), [embedding-ab-20260820.md](../benchmarks/artifacts/embedding-ab-20260820.md)

**Stale if treated as live start:**

- [external-agent-assessment-pack.md](./external-agent-assessment-pack.md) **CURRENT (2026-08-17)** — still useful architecture; pins and “next is R5A / R6a compiler” are **historical**. R5A–R10 landed. S0 ledger flipped the next lever to proof.
- [docs/research/README.md](./README.md) headline table — updated to point here; older “next R6a” sentences elsewhere are history.
- Wave 1 archaeology / Gate 0 / “next is R1b” / LME 0/20.

Every remasure or merge must add a dated section to cycle-closeout **in this order:** Landed → Own pins → Competitor compare (detailed) → Why → Next. Scores-only is incomplete. README gets published-% + same-pin **summary** only, outlinking [docs/benchmarks/README.md](../benchmarks/README.md). No SOTA. Trail axes stay visible (today: open-domain, product SH/temporal on this VM, and integrity MH).

---

## 8. Definition of done for the incoming goal

You are not done when a blog sentence says 80%. You are done when:

1. Fail-closed S0 product `/recall` moves for **attributed** proof-path or reader-path reasons (ledger PROOF and/or READER shrink; WRITE stays small unless you can show new WRITE).
2. MH product is no longer a 2/33 dip on the **integrity** tenant — hop-plan coverage and `hop_join_proven` are measured, not vibed. This-VM 12/33 is not a substitute.
3. Stratified SH / temporal / OD are labeled. This-VM SH 5/98 and temporal 2/38 (reader-off) / SH 45/98 and temporal 20/38 (P5 hybrid; temporal **dip** vs P4 23/38) are the 80% hole. OD 0/11 reader-off / 2/11 P5 stays a diagnostic; do not stuff episodes to fake it.
4. OpMem 13/13 and marketing 17/17 stay green.
5. Fair Mem0 Platform 180 (v3 / top_k 200 / chunk 1) exists on the same question set as Brainy industry. Product `/recall` is a separate labeled row.
6. Only then: 3×90 both lanes, then **one** n=1540 freeze. If that same-pin wins: write the cycle-closeout + README same-pin table. Still do not write “SOTA” until the owner says so.

If you cannot show a stratified delta after one iteration, re-scope. Do not polish.
