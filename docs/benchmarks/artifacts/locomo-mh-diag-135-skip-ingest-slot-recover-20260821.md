# Diagnostic MH-33 skip-ingest (PROOF-only, slot-aligned recovery) — 2026-08-21 (#135)

**Label:** `diagnostic-skip-ingest-not-integrity-s0-1`. Frozen tenant `diag-mh-135` on `brainy_mh`.
WRITE already done. This pass is **PROOF-only** vs the kinship-dest skip-ingest **10/33**.
**Does not replace** the last integrity pin **2/33**.

| Field | Value |
| --- | --- |
| Tenant prefix | `diag-mh-135` (subjects conv-26, 41–44, 47–50) |
| Dataset SHA | `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` |
| Sample | stratified 180 seed 1, **multi-hop 33** |
| Product SHA | `2e84435` on `pr/mh-list-join-proof-1e9e` (#135) |
| Lane | product `POST /recall` |
| Ingest | **skip-ingest** (same 21181 memories as the 7/33 diagnostic) |
| ANN | active, mixed_dimensions=false, signatures.match, embed/extract fallbacks 0 |
| Run id | `diag-mh-135-51a94894` |
| Raw | [locomo-mh-diag-135-skip-ingest-slot-recover-20260821.json](./locomo-mh-diag-135-skip-ingest-slot-recover-20260821.json) |

## Scores

| Lane | Multi-hop | Notes |
| --- | ---: | --- |
| Integrity skip-ingest (last pin) | **2/33** | [packet-proof 2026-08-20](./locomo-mh-packet-proof-20260820.md) |
| Diagnostic fresh ingest | **7/33 (0.212)** | WRITE+PROOF mixed. [artifact](./locomo-mh-diag-135-20260821.md) |
| Crowding / on-prep / enumerate skip-ingest | **8/33 (0.242)** | Prior PROOF-only pins. |
| Community-join skip-ingest | **9/33 (0.273)** | [artifact](./locomo-mh-diag-135-skip-ingest-community-join-20260821.md) |
| Kinship-dest skip-ingest | **10/33 (0.303)** | [artifact](./locomo-mh-diag-135-skip-ingest-kinship-dest-20260821.md) |
| This skip-ingest | **12/33 (0.364)** | +2 vs kinship-dest 10/33. Do not mix with 2/33. |

An intermediate remasure on `94f119b` (before definite-NP / unwind-head tighten) was also **12/33**, with noisier place/unwind lists. Do not cite it as a separate pin.

## Composition vs kinship-dest 10/33

Held (10): `conv-42-q56`, `conv-50-q59`, `conv-41-q60`, `conv-26-q52`, `conv-41-q7`, `conv-48-q77`, `conv-43-q25`, `conv-49-q15`, `conv-48-q85`, `conv-48-q14`.

**Gained:**

- `conv-44-q26` besides+stressor → `work`. Mechanism: dest-subject facts that mention stress and work are prepended onto the activity hop when hops never attached health/occupation. No job lexicon. Hiking stays excluded by the existing `besides` filter.
- `conv-26-q60` instruments → `clarinet, violin`. Mechanism: dest-subject `plays` / `{noun} practice` objects are recovered when the activity index window has clarinet but not violin. No instrument lexicon. Rank boosts play/practice evidence and skips `play` as a drop-zero token.

Place list `conv-48-q82` now returns beach, the yoga studio, and park (previously Park-only). Judge still WRONG: mother's old home is WRITE-thin on this tenant, and extra locatives (Bali, living room) remain. Unwind `conv-26-q24` recovers running from `to *stress` evidence but still misses pottery. Pet tricks `conv-47-q40` recover sit/stay/paw/rollover from trick-mentioned content and miss swimming/frisbee/skateboard. Polar teach-console (`conv-48-q73`) and identity gold (`conv-26-q65`) are unchanged. Outdoor `conv-41-q32` is still a hike-only WRONG (mountaineering absent).

Do not publish this as SOTA, as full LoCoMo, or as a frozen same-pin vs Mem0.
