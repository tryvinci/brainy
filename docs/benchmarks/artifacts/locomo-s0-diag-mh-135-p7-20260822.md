# LoCoMo S0 product `/recall` — P7 hop-local joins — 2026-08-22

Same frozen store as [locomo-s0-diag-mh-135-20260822.md](./locomo-s0-diag-mh-135-20260822.md): tenant `diag-mh-135` + conv-30, dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`. Product SHA `f3e0a7f` (keep leftover-covering hop contents when skipping dumps, including identity-only leftover names; rare-share dual-entity possessions from typed Values then omitted hop-content snippets; score rare tokens by shortest matching value; keep title-cased typed joins on hybrid abstain without restoring identity slogans; where-answers only from locative leftover-covering places). Hybrid **on** (`BRAINY_RECALL_LLM=1`). Go default for `BRAINY_RECALL_LLM` stays **off**.

P6 pair: [locomo-s0-diag-mh-135-p6-20260822.md](./locomo-s0-diag-mh-135-p6-20260822.md) (`45a83b5`, **87/180**).

**Not** n=1540. **Not** a Mem0 same-pin. **Not** SOTA. Does not replace integrity 32/180. Does not replace the reader-off 19/180 no-LLM pin. Does not replace README 11.4% / 70% 1×30.

## Scores vs prior pins on this store

| Lane | Overall | multi-hop | open-domain | single-hop | temporal |
| --- | ---: | ---: | ---: | ---: | ---: |
| product reader **off** (`453a929`) | **19/180 (0.106)** | **12/33** | 0/11 | 5/98 | 2/38 |
| product hybrid **on** P5 (`5ad07c4`) | **84/180 (0.467)** | **17/33** | 2/11 | **45/98** | **20/38** |
| product hybrid **on** P6 (`45a83b5`) | **87/180 (0.483)** | **13/33 dip** | **3/11** | **52/98** | **19/38 dip** |
| product hybrid **on** P7 (`f3e0a7f`) | **88/180 (0.489)** | **14/33** | **4/11** | **49/98 dip** | **21/38** |
| industry search+harness (reader-off pin) | 62/180 (0.344) | 10/33 | 3/11 | 27/98 | 22/38 |

MH **13→14** (still a dip vs P5 **17/33**). SH **52→49 dip**. OD **3→4**. Temporal **19→21** (recovers the P6 temporal dip vs P5). Product overall still leads this-VM industry 62/180 on the labeled product lane — still not a Mem0 same-pin.

Item flips vs P6: **+9 / −8 = net +1**.

Named P6 MH recoveries: chili cook-off + ring-toss (`conv-41-q29`); signed basketball (`conv-43-q25`, judge accepted trophy extra); family injured (`conv-49-q49`). Still missing: mother's hobbies (`conv-48-q14`, yoga extra vs gold reading/travel/art/cooking); Phuket diving (`conv-48-q77` → not in memory). Tim-UK **held**. Walking **held**. Gym **held**. Instruments **held** (list lock, extras). Ferrari **held**. Filling **held**.

Named SH losses vs P6: second puppy Shadow (`conv-41-q147` → not in memory); Frank Ocean festival (`conv-50-q125` → "Rocks"); CS:GO tournament (`conv-47-q88` → not in memory); Calvin/Dave goals incomplete (`conv-50-q118`).

Locks held: `clarinet, violin` extras still list-locked; Ferrari `2`; strawberry filling; gym; Tim-UK United Kingdom; walking.

## Failure ledger (92 misses)

| Primary | P6 | P7 |
| --- | ---: | ---: |
| PROOF_MISS | 28 | 29 |
| RETRIEVAL_MISS | 28 | 29 |
| READER_MISS | 29 | 27 |
| WRITE_MISS | 6 | 5 |
| HARNESS_ERROR | 2 | 2 |

Largest P7 cells: `single-hop:PROOF_MISS` 23 (was 22), `single-hop:RETRIEVAL_MISS` 12, `single-hop:READER_MISS` 11 (was 9), `multi-hop:RETRIEVAL_MISS` 11, `temporal:READER_MISS` 9 (was 11), `multi-hop:READER_MISS` 7 (was 9). WRITE 5 — do not merge #133.

## What this says

1. Rare-share on omitted possession contents + keeping title-cased typed joins on hybrid abstain recovers signed basketball without re-enabling dual-entity sushi dumps. Walking/UK/gym still hold.
2. Hop-local leftover facts recover chili cook-off and family injured. Mother's hobbies still fail the judge on a yoga extra. Phuket is a write split (meditation retreat vs diving spot) — hybrid abstains.
3. 88/180 is still far from 80% (would be 144/180) and is **not** n=1540. SH **52→49** is a named dip (Shadow abstain, festival "Rocks"). MH **17→14** vs P5 remains a dip. Next product step is recover SH 52 without giving back basketball/chili/walking.

Report: `locomo-s0-diag-mh-135-p7-product-recall-s1-fed44d` (summary JSON + failure ledger in this folder). Auto smoke JSON/md dumps are not committed (secret scanner).
