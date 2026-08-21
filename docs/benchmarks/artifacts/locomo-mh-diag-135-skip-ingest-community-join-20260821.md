# Diagnostic MH-33 skip-ingest (PROOF-only, community join) — 2026-08-21 (#135)

**Label:** `diagnostic-skip-ingest-not-integrity-s0-1`. Frozen tenant `diag-mh-135` on `brainy_mh`.
WRITE already done. This pass is **PROOF-only** vs the enumerate/outdoor skip-ingest **8/33**.
**Does not replace** the last integrity pin **2/33**.

| Field | Value |
| --- | --- |
| Tenant prefix | `diag-mh-135` (subjects conv-26, 41–44, 47–50) |
| Dataset SHA | `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` |
| Sample | stratified 180 seed 1, **multi-hop 33** |
| Product SHA | `c97cc0a` on `pr/mh-list-join-proof-1e9e` (#135) |
| Lane | product `POST /recall` |
| Ingest | **skip-ingest** (same 21181 memories as the 7/33 diagnostic) |
| ANN | active, mixed_dimensions=false, signatures.match, embed/extract fallbacks 0 |
| Run id | `diag-mh-135-74b14c7d` |
| Raw | [locomo-mh-diag-135-skip-ingest-community-join-20260821.json](./locomo-mh-diag-135-skip-ingest-community-join-20260821.json) |

## Scores

| Lane | Multi-hop | Notes |
| --- | ---: | --- |
| Integrity skip-ingest (last pin) | **2/33** | [packet-proof 2026-08-20](./locomo-mh-packet-proof-20260820.md) |
| Diagnostic fresh ingest | **7/33 (0.212)** | WRITE+PROOF mixed. [artifact](./locomo-mh-diag-135-20260821.md) |
| Crowding / on-prep / enumerate skip-ingest | **8/33 (0.242)** | Prior PROOF-only pins. |
| This skip-ingest | **9/33 (0.273)** | +1 vs enumerate 8/33. Do not mix with 2/33. |

An intermediate that applied containment to **all** multi-entity joins scored **7/33** (sports collectible collapsed to `Book`). Do not cite that 7/33 as a pin.

## Composition vs enumerate 8/33

Held (8): `conv-42-q56`, `conv-50-q59`, `conv-41-q60`, `conv-26-q52`, `conv-41-q7`, `conv-48-q77`, `conv-43-q25`, `conv-49-q15`.

**Gained:** `conv-48-q85` dual community → `yoga, deborah's running group`. Mechanism: community lists join by `organized`/`started`/`group` token-subset (yoga ∩ organized yoga) and keep a value that names the other hop entity (`deborah's running group` on Anna). Exact slot intersect was empty; content fallback had been unwind gerunds. Anna's activity hop is `search_fallback`; containment reads that fallback slot list. Judge accepted yoga + running. No activity lexicon.

`conv-41-q32` outdoor remains a single hike fact (still WRONG: mountaineering absent on this tenant). Polar teach-console (`conv-48-q73`) stays the named dip. Mother hobbies (`conv-48-q14`) still hop the source person.

Do not publish this as SOTA, as full LoCoMo, or as a frozen same-pin vs Mem0.
