# LoCoMo S0 product `/recall` — P4 identity/garbage hybrid — 2026-08-22

Same frozen store as [locomo-s0-diag-mh-135-20260822.md](./locomo-s0-diag-mh-135-20260822.md): tenant `diag-mh-135` + conv-30, dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, stratified 180 seed 1, `--fail-closed --skip-ingest`. Product SHA `6f74024` (reject punctuation-only / low-letter hybrid answers; skip identity-only Structured dumps that miss leftover query tokens; keep only covering memories in that prompt; do not hop-ground hybrid answers onto those identity dumps). Hybrid **on** (`BRAINY_RECALL_LLM=1`). Where / polar stay locked. Count / dual-entity `mh_list` locks stay. Enumerated hop-ground skip from P2b stays. Distinctive-token admit from P3 stays.

P3 pair: [locomo-s0-diag-mh-135-p3-20260822.md](./locomo-s0-diag-mh-135-p3-20260822.md) (`5bc28ea`, **73/180**).

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
| industry search+harness (reader-off pin) | 62/180 (0.344) | 10/33 | 3/11 | 27/98 | 22/38 |

MH **held 16/33** (gain `conv-26-q52` pet names Oliver/Luna/Bailey; named loss `conv-43-q38` Tim most-visited country UK). SH **32→37**. OD **held 3/11** (gain `conv-47-q12` James girlfriend No; named loss `conv-26-q30` Melanie LGBTQ). Temporal **22→23** (recovers P3 named loss `conv-44-q38` wine-tasting weekend; named loss `conv-41-q53` John community center 2022). Product overall still leads this-VM industry 62/180 on the labeled product lane — still not a Mem0 same-pin.

Item flips vs P3: **+9 / −3 = net +6**.

## Failure ledger (101 misses)

| Primary | P3 | P4 |
| --- | ---: | ---: |
| PROOF_MISS | 34 | 32 |
| RETRIEVAL_MISS | 29 | 30 |
| READER_MISS | 34 | 30 |
| WRITE_MISS | 8 | 7 |
| HARNESS_ERROR | 2 | 2 |

Largest P4 cells: `single-hop:PROOF_MISS` 26 (was 28), `single-hop:READER_MISS` 19 (was 22), `multi-hop:RETRIEVAL_MISS` 12, `single-hop:RETRIEVAL_MISS` 12. WRITE 8→7 — do not merge #133. The two `HARNESS_ERROR` rows are oracle mislabels on `not in memory` (`conv-42-q146`, `conv-48-q116`), not harness crashes.

## What this says

1. Garbage hybrid (`!!!!`) was accepted because `isHybridGarbageAnswer` only matched a few sentinels. That overwrote typed lists (snakes, pet names). Rejecting no-letter / low-letter / single-rune spam recovers those lists.
2. Identity hops (`Maria is a inspiration`) crowded the hybrid prompt and then `groundToHopValues` replaced a covering hybrid answer (`Shadow`) with `Inspiration, Family, Team`. Skip is **identity-only** hops when leftover distinctive query tokens are covered in the packet — skill / dual-entity hops stay (instruments `clarinet, violin` held).
3. 79/180 is still far from 80% (would be 144/180) and is **not** n=1540. SH PROOF 26 + SH READER 19 remain the mass. Wheel of Time / nearby-wrong SH (German vs Spanish, pottery injury vs road-trip) stay misses.
4. MH 16/33 held with a named 1-for-1 swap (pets recovered, Tim-UK lost). Reader-off 19/180 remains the labeled no-LLM product pin.

Report: `locomo-s0-diag-mh-135-p4-product-recall-s1-bc8b5d` (summary JSON + failure ledger in this folder). Auto smoke JSON/md dumps are not committed (secret scanner).
