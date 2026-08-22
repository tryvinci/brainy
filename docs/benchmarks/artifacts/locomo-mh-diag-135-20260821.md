# Diagnostic MH-33 product `/recall` — 2026-08-21 (#135)

**Label:** `diagnostic-not-integrity-s0-1`. Fresh async ingest on `brainy_mh`.
WRITE+PROOF mixed. **Not** skip-ingest attribution of `integrity-s0-1`.
**Does not replace** the last integrity pin **2/33**.

| Field | Value |
| --- | --- |
| Tenant prefix | `diag-mh-135` (subjects conv-26, 41–44, 47–50) |
| Dataset SHA | `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` |
| Sample | stratified 180 seed 1, **multi-hop 33** |
| Product SHA | `a7cf465` on `pr/mh-list-join-proof-1e9e` (#135) |
| Lane | product `POST /recall` |
| Jobs | **1472 completed / 0 failed** (provider extract + 768-d embed) |
| Memories | 21181 |
| ANN | active, mixed_dimensions=false, signatures.match, embed/extract fallbacks 0 |
| Run id | `diag-mh-135-48d06824` |
| Raw | [locomo-mh-diag-135-20260821.json](./locomo-mh-diag-135-20260821.json) |

## Scores

| Lane | Multi-hop | Notes |
| --- | ---: | --- |
| Integrity skip-ingest (last pin) | **2/33** | [packet-proof 2026-08-20](./locomo-mh-packet-proof-20260820.md) |
| This diagnostic | **7/33 (0.212)** | Fresh WRITE. Do not mix with 2/33. |

Judge CORRECT this run (several are crowded lists the judge accepted):

- `conv-42-q56` Nate+Joanna animals → `Turtles, Nature` (same join class as the 2/33 attributed win)
- `conv-48-q77` diving spot → crowded list containing `Phuket`
- `conv-43-q25` sports collectible → signed basketball present in a crowded possession dump
- `conv-41-q60` Maria's dogs → `Shadow, Coco` present in a crowded pet dump
- `conv-26-q52` pets' names → `Luna, Oliver, Bailey` present in a crowded dump
- `conv-41-q7` childhood items → judge CORRECT on a crowded possession dump
- `conv-49-q15` unhealthy snacks → judge CORRECT on a crowded preference dump (same class as the 2/33 soda/candy hit)

WRONG (26/33) is still mostly **list crowding / wrong slot**, plus the known WRITE-miss identity gold (`conv-26-q65`). Practice-place (`conv-48-q82`) returned park/beach but missed yoga studio and dumped unwind slogans. List-head outdoor (`conv-41-q32`) mentioned hiking but not mountaineering and dumped indoor items.

Do not publish this as SOTA, as full LoCoMo, or as a frozen same-pin vs Mem0.
