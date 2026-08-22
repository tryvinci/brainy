# Diagnostic MH-33 skip-ingest (PROOF-only) — 2026-08-21 (#135)

**Label:** `diagnostic-skip-ingest-not-integrity-s0-1`. Frozen tenant `diag-mh-135` on `brainy_mh`.
WRITE already done. This pass is **PROOF-only** vs the prior diagnostic **7/33**.
**Does not replace** the last integrity pin **2/33**.

| Field | Value |
| --- | --- |
| Tenant prefix | `diag-mh-135` (subjects conv-26, 41–44, 47–50) |
| Dataset SHA | `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` |
| Sample | stratified 180 seed 1, **multi-hop 33** |
| Product SHA | `33bdea4` on `pr/mh-list-join-proof-1e9e` (#135) |
| Lane | product `POST /recall` |
| Ingest | **skip-ingest** (same 21181 memories as the 7/33 diagnostic) |
| ANN | active, mixed_dimensions=false, signatures.match, embed/extract fallbacks 0 |
| Run id | `diag-mh-135-7765147e` |
| Raw | [locomo-mh-diag-135-skip-ingest-20260821.json](./locomo-mh-diag-135-skip-ingest-20260821.json) |

## Scores

| Lane | Multi-hop | Notes |
| --- | ---: | --- |
| Integrity skip-ingest (last pin) | **2/33** | [packet-proof 2026-08-20](./locomo-mh-packet-proof-20260820.md) |
| Diagnostic fresh ingest | **7/33 (0.212)** | WRITE+PROOF mixed. [artifact](./locomo-mh-diag-135-20260821.md) |
| This skip-ingest | **8/33 (0.242)** | PROOF-only vs 7/33. Do not mix with 2/33. |

An intermediate cap-only skip-ingest scored **5/33** (hard truncate cut gold in the tail). Ranking by query evidence, then cap, recovered those lists. Do not publish 5/33 as a pin.

Judge CORRECT this run:

- `conv-42-q56` Nate+Joanna animals → `Turtles, Nature` (same join class as the 2/33 attributed win)
- `conv-50-q59` count → `2` (was `3`; atom-refill no longer padded the typed set)
- `conv-41-q60` dogs' names → `Shadow, Coco` (was a possession dump)
- `conv-26-q52` pets' names → `Luna, Oliver, Bailey` (was a possession dump)
- `conv-41-q7` childhood items → `little doll, film camera (as a kid)` (was a 272-char dump)
- `conv-48-q77` diving spot → crowded list still containing `Phuket`
- `conv-43-q25` sports collectible → signed basketball present (still mixed with other possessions)
- `conv-48-q73` polar/skill dump shortened; judge CORRECT on gold `yes` (fragile crowded-accept)

Named dip vs 7/33: `conv-49-q15` unhealthy snacks. The long dump had contained soda/candy; ranking kept health slogans and dropped the gold tail.

Practice-place (`conv-48-q82`) no longer dumps unwind slogans (`Relaxing, Escaping, Spending`); answer is a short place list (`Park, Yoga In The Park, …`) but still WRONG (missed studio/beach/home).

Do not publish this as SOTA, as full LoCoMo, or as a frozen same-pin vs Mem0.
