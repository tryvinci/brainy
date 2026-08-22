# Diagnostic MH-33 skip-ingest (PROOF-only, on-prep / polar / un-) — 2026-08-21 (#135)

**Label:** `diagnostic-skip-ingest-not-integrity-s0-1`. Frozen tenant `diag-mh-135` on `brainy_mh`.
WRITE already done. This pass is **PROOF-only** vs the crowding skip-ingest **8/33**.
**Does not replace** the last integrity pin **2/33**.

| Field | Value |
| --- | --- |
| Tenant prefix | `diag-mh-135` (subjects conv-26, 41–44, 47–50) |
| Dataset SHA | `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` |
| Sample | stratified 180 seed 1, **multi-hop 33** |
| Product SHA | `a521c47` on `pr/mh-list-join-proof-1e9e` (#135) |
| Lane | product `POST /recall` |
| Ingest | **skip-ingest** (same 21181 memories as the 7/33 diagnostic) |
| ANN | active, mixed_dimensions=false, signatures.match, embed/extract fallbacks 0 |
| Run id | `diag-mh-135-0980f34d` |
| Raw | [locomo-mh-diag-135-skip-ingest-onprep-20260821.json](./locomo-mh-diag-135-skip-ingest-onprep-20260821.json) |

## Scores

| Lane | Multi-hop | Notes |
| --- | ---: | --- |
| Integrity skip-ingest (last pin) | **2/33** | [packet-proof 2026-08-20](./locomo-mh-packet-proof-20260820.md) |
| Diagnostic fresh ingest | **7/33 (0.212)** | WRITE+PROOF mixed. [artifact](./locomo-mh-diag-135-20260821.md) |
| Crowding skip-ingest | **8/33 (0.242)** | Prior PROOF-only pin. [artifact](./locomo-mh-diag-135-skip-ingest-20260821.md) |
| This skip-ingest | **8/33 (0.242)** | Same count; **composition change**. Do not mix with 2/33. |

An intermediate on SHA `d7ee0ce` (on-prep + polar lock, before the lowercased person-clause / un- gold fix) scored **7/33**. Do not cite 7/33 as a pin.

## Composition vs crowding 8/33

Judge CORRECT this run (8):

- `conv-42-q56` Nate+Joanna animals → `Turtles, Nature` (held)
- `conv-50-q59` count → `2` (held)
- `conv-41-q60` dogs' names → `Shadow, Coco` (held)
- `conv-26-q52` pets' names → `Luna, Oliver, Bailey` (held)
- `conv-41-q7` childhood items → `little doll, film camera (as a kid)` (held)
- `conv-48-q77` diving spot → crowded list still containing `Phuket` (held)
- `conv-43-q25` sports collectible → signed basketball present (held)
- `conv-49-q15` unhealthy snacks → `soda and candy` present (**attributed** un- filter; still mixed with other preferences)

Named dip vs crowding 8/33: `conv-48-q73` polar teach-console. Prior judge CORRECT on a crowded skill dump; polar hop-compose is now locked, so the answer is `not in memory`. Gold is `yes`. Honest abstention, still WRONG. Do not restore activity dumps to recover that row.

Practice-place (`conv-48-q82`) is now `Park` only (person-clause junk gone). Still WRONG: hops on this tenant do not carry `yoga on the beach` / studio locatives, and mother-house sentences lack the practice-object token so the focus pass skips them. No place lexicon.

Polar surf (`conv-48-q79`) is `not in memory` instead of `Beach, Relaxing, Escaping` (still WRONG).

Do not publish this as SOTA, as full LoCoMo, or as a frozen same-pin vs Mem0.
