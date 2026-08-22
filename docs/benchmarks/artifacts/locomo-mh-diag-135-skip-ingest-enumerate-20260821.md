# Diagnostic MH-33 skip-ingest (PROOF-only, enumerate refine / outdoor∩group) — 2026-08-21 (#135)

**Label:** `diagnostic-skip-ingest-not-integrity-s0-1`. Frozen tenant `diag-mh-135` on `brainy_mh`.
WRITE already done. This pass is **PROOF-only** vs the on-prep skip-ingest **8/33**.
**Does not replace** the last integrity pin **2/33**.

| Field | Value |
| --- | --- |
| Tenant prefix | `diag-mh-135` (subjects conv-26, 41–44, 47–50) |
| Dataset SHA | `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` |
| Sample | stratified 180 seed 1, **multi-hop 33** |
| Product SHA | `5804072` on `pr/mh-list-join-proof-1e9e` (#135) |
| Lane | product `POST /recall` |
| Ingest | **skip-ingest** (same 21181 memories as the 7/33 diagnostic) |
| ANN | active, mixed_dimensions=false, signatures.match, embed/extract fallbacks 0 |
| Run id | `diag-mh-135-2918f624` |
| Raw | [locomo-mh-diag-135-skip-ingest-enumerate-20260821.json](./locomo-mh-diag-135-skip-ingest-enumerate-20260821.json) |

## Scores

| Lane | Multi-hop | Notes |
| --- | ---: | --- |
| Integrity skip-ingest (last pin) | **2/33** | [packet-proof 2026-08-20](./locomo-mh-packet-proof-20260820.md) |
| Diagnostic fresh ingest | **7/33 (0.212)** | WRITE+PROOF mixed. [artifact](./locomo-mh-diag-135-20260821.md) |
| Crowding skip-ingest | **8/33 (0.242)** | [artifact](./locomo-mh-diag-135-skip-ingest-20260821.md) |
| On-prep skip-ingest | **8/33 (0.242)** | Composition change. [artifact](./locomo-mh-diag-135-skip-ingest-onprep-20260821.md) |
| This skip-ingest | **8/33 (0.242)** | Same CORRECT set as on-prep. Outdoor answer changed. Do not mix with 2/33. |

## Composition vs on-prep 8/33

Judge CORRECT this run (same 8): `conv-42-q56`, `conv-50-q59`, `conv-41-q60`, `conv-26-q52`, `conv-41-q7`, `conv-48-q77`, `conv-43-q25`, `conv-49-q15`.

Attributed proof change that did **not** move the count: `conv-41-q32` outdoor-with-colleagues. On-prep enumerate dump was unbounded; a later enumerate-refine spot-check collapsed to colleague-only `convention attendance`. This SHA answers `being outdoors - going for hikes` (list-head preferred over companion-only). Still **WRONG**: gold is Hiking, mountaineering, and mountaineering is not in the hop slot set (WRITE-thin on this tenant). Do not add an outdoor-sport lexicon.

Dual-community (`conv-48-q85`) remains unwind slogans (`relaxing, escaping, spending, gardening`) because exact slot intersect is empty (`yoga` vs `organized yoga`) and content fallback keeps shared unwind gerunds. Mother hobbies (`conv-48-q14`) still hop the source person. Polar teach-console (`conv-48-q73`) stays the named dip.

Do not publish this as SOTA, as full LoCoMo, or as a frozen same-pin vs Mem0.
