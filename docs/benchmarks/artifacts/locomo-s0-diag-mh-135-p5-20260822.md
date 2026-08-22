# LoCoMo S0 product `/recall` — P5 activity-dump skip — 2026-08-22

Same frozen store as [locomo-s0-diag-mh-135-20260822.md](./locomo-s0-diag-mh-135-20260822.md): tenant `diag-mh-135` + conv-30, dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`. Product SHA `5ad07c4` (skip activity/event hop dumps that miss leftover distinctive query tokens; keep skill/possession/preference joins; keep specific packet facts whose gold is a synonym of the leftover token; do not hop-ground or compose those dumps when hybrid abstains). Hybrid **on** (`BRAINY_RECALL_LLM=1`). Where / polar stay locked. Count / dual-entity `mh_list` locks stay. Enumerated hop-ground skip from P2b stays. Distinctive-token admit from P3 stays. Identity/garbage skip from P4 stays.

P4 pair: [locomo-s0-diag-mh-135-p4-20260822.md](./locomo-s0-diag-mh-135-p4-20260822.md) (`6f74024`, **79/180**).

**Not** n=1540. **Not** a Mem0 same-pin. **Not** SOTA. Does not replace integrity 32/180. Does not replace the reader-off 19/180 no-LLM pin.

## Scores vs prior pins on this store

| Lane | Overall | multi-hop | open-domain | single-hop | temporal |
| --- | ---: | ---: | ---: | ---: | ---: |
| product reader **off** (`453a929`) | **19/180 (0.106)** | **12/33** | 0/11 | 5/98 | 2/38 |
| product hybrid **on** P1 (`3d42b17`) | **37/180 (0.206)** | **10/33** | 1/11 | **19/98** | **7/38** |
| product hybrid **on** P2 length-lock (`681028e`) | **56/180 (0.311)** | **11/33** | 1/11 | **23/98** | **21/38** |
| product hybrid **on** P2b (`fb41ece`) | **61/180 (0.339)** | **16/33** | 1/11 | **25/98** | **19/38** |
| product hybrid **on** P3 (`5bc28ea`) | **73/180 (0.406)** | **16/33** | **3/11** | **32/98** | **22/38** |
| product hybrid **on** P4 (`6f74024`) | **79/180 (0.439)** | **16/33** | **3/11** | **37/98** | **23/38** |
| product hybrid **on** P5 (`5ad07c4`) | **84/180 (0.467)** | **17/33** | **2/11** | **45/98** | **20/38** |
| industry search+harness (reader-off pin) | 62/180 (0.344) | 10/33 | 3/11 | 27/98 | 22/38 |

MH **16→17** (named gain: `conv-43-q38` Tim most-visited country United Kingdom, the P4 visa-dump loss). SH **37→45**. OD **3→2 dip** (named loss `conv-47-q12` James girlfriend No → not in memory). Temporal **23→20 dip** (named losses: `conv-26-q36` mentorship weekend extra clause, `conv-26-q67` biking date, `conv-41-q20` Pacific Northwest 2022 → not in memory). Product overall still leads this-VM industry 62/180 on the labeled product lane — still not a Mem0 same-pin.

Item flips vs P4: **+10 / −5 = net +5**. Locks held: `clarinet, violin`; Ferrari `2`; snacks; strawberry filling; gym; Coco/Shadow; Oliver/Luna/Bailey; Susie/Seraphim; Shadow.

## Failure ledger (96 misses)

| Primary | P4 | P5 |
| --- | ---: | ---: |
| PROOF_MISS | 32 | 32 |
| RETRIEVAL_MISS | 30 | 29 |
| READER_MISS | 30 | 26 |
| WRITE_MISS | 7 | 7 |
| HARNESS_ERROR | 2 | 2 |

Largest P5 cells: `single-hop:PROOF_MISS` 25 (was 26), `single-hop:RETRIEVAL_MISS` 12, `single-hop:READER_MISS` 12 (was 19), `multi-hop:RETRIEVAL_MISS` 11, `temporal:READER_MISS` 10 (was 7 — **dip**). WRITE 7 — do not merge #133.

## What this says

1. P4 identity-only skip left typed **activity/event** slogan lists in the hybrid prompt. `which country has Tim visited` retrieved `Tim has experiences in the United Kingdom` but Structured `visa requirements, …` crowded the reader into abstain, then `composeFromHopValues` dumped the slogans. Skipping those dumps unless hops are a skill/possession/preference join recovers Tim-UK and several SH identity dumps (Max, school funding, memorials, winter reading).
2. 84/180 is still far from 80% (would be 144/180) and is **not** n=1540. SH PROOF 25 remains the mass. Temporal 23→20 is a named dip — do not add LoCoMo-named date rules to chase it.
3. MH 17/33 is one attributed Tim-UK recovery, not n=1540 MH. Reader-off 19/180 remains the labeled no-LLM product pin.

Report: `locomo-s0-diag-mh-135-p5-product-recall-s1-77cd22` (summary JSON + failure ledger in this folder). Auto smoke JSON/md dumps are not committed (secret scanner).
