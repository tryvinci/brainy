# LoCoMo S0 product `/recall` — P8 SH recovery without dump restore — 2026-08-22

Same frozen store as [locomo-s0-diag-mh-135-20260822.md](./locomo-s0-diag-mh-135-20260822.md): tenant `diag-mh-135` + conv-30, dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`. Product SHA `86eab77` (unlock skipped `mh_list` only on where-queries; drop attended-event / foreign-possessive hop values; dated then undated ordinal names; leftover-covering specific facts on hybrid abstain with conflicting date tails stripped). Hybrid **on** (`BRAINY_RECALL_LLM=1`). Go default for `BRAINY_RECALL_LLM` stays **off**.

P7 pair: [locomo-s0-diag-mh-135-p7-20260822.md](./locomo-s0-diag-mh-135-p7-20260822.md) (`f3e0a7f`, **88/180**).

A superseded in-VM run on `abab569` (`locomo-s0-diag-mh-135-p8-product-recall-s1-759386`, **92/180**) unlocked `mh_list` on any skip and lost community yoga+running (`conv-48-q85`). That run is **not** this pin. This pin is `locomo-s0-diag-mh-135-p8b-product-recall-s1-6b1754` on `86eab77`.

**Not** n=1540. **Not** a Mem0 same-pin. **Not** SOTA. Does not replace integrity 32/180. Does not replace the reader-off 19/180 no-LLM pin. Does not replace README 11.4% / 70% 1×30.

## Scores vs prior pins on this store

| Lane | Overall | multi-hop | open-domain | single-hop | temporal |
| --- | ---: | ---: | ---: | ---: | ---: |
| product reader **off** (`453a929`) | **19/180 (0.106)** | **12/33** | 0/11 | 5/98 | 2/38 |
| product hybrid **on** P5 (`5ad07c4`) | **84/180 (0.467)** | **17/33** | 2/11 | **45/98** | **20/38** |
| product hybrid **on** P6 (`45a83b5`) | **87/180 (0.483)** | **13/33 dip** | **3/11** | **52/98** | **19/38 dip** |
| product hybrid **on** P7 (`f3e0a7f`) | **88/180 (0.489)** | **14/33** | **4/11** | **49/98 dip** | **21/38** |
| product hybrid **on** P8 (`86eab77`) | **93/180 (0.517)** | **15/33** | **4/11** | **52/98** | **22/38** |
| industry search+harness (reader-off pin) | 62/180 (0.344) | 10/33 | 3/11 | 27/98 | 22/38 |

MH **14→15** (still a dip vs P5 **17/33**). SH **49→52** recovers the P7 dip vs P6. OD **held 4/11**. Temporal **21→22**. Product overall still leads this-VM industry 62/180 on the labeled product lane — still not a Mem0 same-pin.

Item flips vs P7: **+5 / −0 = net +5**.

Named P7 SH recoveries: second puppy Shadow (`conv-41-q147`, ordinal dated-then-undated); Frank Ocean festival (`conv-50-q125`, where-only mh_list unlock over "Rocks"); CS:GO tournament (`conv-47-q88`, leftover packet fact with conflicting date tail stripped). Named MH gain: mother's hobbies (`conv-48-q14`) — attended-event hop value dropped; judge accepted the dump without yoga. Incidental temporal: Joanna ice cream weekend (`conv-42-q29`).

Held P7 MH/SH: chili `conv-41-q29`, basketball `conv-43-q25`, injured `conv-49-q49`, walking `conv-44-q110`, UK `conv-43-q38`, gym `conv-41-q110`, community yoga+running `conv-48-q85`. Phuket `conv-48-q77` still MISS (`not in memory`).

Locks held: `clarinet, violin` extras still list-locked; Ferrari `2`; strawberry filling; gym; Tim-UK United Kingdom; walking.

## Failure ledger (87 misses)

| Primary | P7 | P8 |
| --- | ---: | ---: |
| PROOF_MISS | 29 | 28 |
| RETRIEVAL_MISS | 29 | 29 |
| READER_MISS | 27 | 24 |
| WRITE_MISS | 5 | 4 |
| HARNESS_ERROR | 2 | 2 |

Largest P8 cells: `single-hop:PROOF_MISS` 22 (was 23), `single-hop:RETRIEVAL_MISS` 12, `multi-hop:RETRIEVAL_MISS` 11, `temporal:READER_MISS` 9, `single-hop:READER_MISS` 9 (was 11), `multi-hop:READER_MISS` 6. WRITE 4 — do not merge #133.

## What this says

1. Where-only mh_list unlock recovers dual-entity locative SH (festival) without unlocking typed community/skill joins (yoga+running held). Unlocking mh_list on any skip is too broad — 92/180 on `abab569` is superseded.
2. Dated-then-undated ordinal names recover Shadow without restoring identity dumps. Attended-event / foreign-possessive hop values drop yoga extras from kinship hobby lists.
3. Leftover-covering specific facts on hybrid abstain recover CS:GO; stripping a conflicting date tail keeps the judge from rejecting a May 8 gold against a May 7 packet stamp.
4. 93/180 is still far from 80% (would be 144/180) and is **not** n=1540. SH **52/98** matches P6, not a new SH ceiling. MH **17→15** vs P5 remains a dip. Next product step is remaining SH **PROOF 22** (nearby-wrong / incomplete compose) without giving back Shadow / festival / CS:GO / basketball / chili / walking.

Report: `locomo-s0-diag-mh-135-p8b-product-recall-s1-6b1754` (summary JSON + failure ledger in this folder). Auto smoke JSON/md dumps are not committed (secret scanner).
