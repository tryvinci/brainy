# Diagnostic MH-33 skip-ingest (PROOF-only, kinship dest) — 2026-08-21 (#135)

**Label:** `diagnostic-skip-ingest-not-integrity-s0-1`. Frozen tenant `diag-mh-135` on `brainy_mh`.
WRITE already done. This pass is **PROOF-only** vs the community-join skip-ingest **9/33**.
**Does not replace** the last integrity pin **2/33**.

| Field | Value |
| --- | --- |
| Tenant prefix | `diag-mh-135` (subjects conv-26, 41–44, 47–50) |
| Dataset SHA | `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` |
| Sample | stratified 180 seed 1, **multi-hop 33** |
| Product SHA | `9d8dbeb` on `pr/mh-list-join-proof-1e9e` (#135) |
| Lane | product `POST /recall` |
| Ingest | **skip-ingest** (same 21181 memories as the 7/33 diagnostic) |
| ANN | active, mixed_dimensions=false, signatures.match, embed/extract fallbacks 0 |
| Run id | `diag-mh-135-e4d586d3` |
| Raw | [locomo-mh-diag-135-skip-ingest-kinship-dest-20260821.json](./locomo-mh-diag-135-skip-ingest-kinship-dest-20260821.json) |

## Scores

| Lane | Multi-hop | Notes |
| --- | ---: | --- |
| Integrity skip-ingest (last pin) | **2/33** | [packet-proof 2026-08-20](./locomo-mh-packet-proof-20260820.md) |
| Diagnostic fresh ingest | **7/33 (0.212)** | WRITE+PROOF mixed. [artifact](./locomo-mh-diag-135-20260821.md) |
| Crowding / on-prep / enumerate skip-ingest | **8/33 (0.242)** | Prior PROOF-only pins. |
| Community-join skip-ingest | **9/33 (0.273)** | [artifact](./locomo-mh-diag-135-skip-ingest-community-join-20260821.md) |
| This skip-ingest | **10/33 (0.303)** | +1 vs community-join 9/33. Do not mix with 2/33. |

An intermediate rewrite-only remasure on `f5b9ea8` stayed **9/33** (dest entity rewrote, activity index still missed unpredicated facts). Do not cite 9/33 as a regression.

## Composition vs community-join 9/33

Held (9): `conv-42-q56`, `conv-50-q59`, `conv-41-q60`, `conv-26-q52`, `conv-41-q7`, `conv-48-q77`, `conv-43-q25`, `conv-49-q15`, `conv-48-q85`.

**Gained:** `conv-48-q14` dest hobbies → `reading, travel, art, cooking meals, cooking, …`. Mechanism: family hop value `mother` (or copula gloss `her mom`) is rewritten to `{Name}'s mother` from family evidence; the source entity id does not ride onto the activity hop; dest-subject facts without an activity atom (`had reading as one of her hobbies`, `passionate about travel`, `interested in art`) are merged as attitude slots. Judge accepted the four gold hobbies despite leftover dest activity noise. No hobby lexicon.

`conv-41-q32` outdoor remains a single hike fact (still WRONG: mountaineering absent on this tenant). Polar teach-console (`conv-48-q73`) stays the named dip. Place list (`conv-48-q82`) is still Park-only.

Do not publish this as SOTA, as full LoCoMo, or as a frozen same-pin vs Mem0.
