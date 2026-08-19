# LoCoMo dual-path freeze (R10 wiring)

**Status:** harness labels landed. **Not a remasure.** Do not run n=1540 until a freeze is explicitly requested.  
**Does not claim:** SOTA, beats-Mem0, 70–80% on full LoCoMo, or 1×30 70% as n=1540.

## Two lanes (never mix)

| Lane | Flag | Answer path | Default top-k | What it measures |
| --- | --- | --- | ---: | --- |
| Product | `--eval-lane product-recall` (or `BRAINY_USE_RECALL=1`) | `POST /recall`, fail-closed | 30 | Structured product answers over compiled facts |
| Industry | `--eval-lane industry-search` | search → shared answerer → shared judge | **200** | Mem0-style retrieval+LLM; report retrieved tokens/chars, not only requested top-k |

Current pins stay labeled as they were measured:

- Product `/recall` full: **175/1540 = 11.4%** (SHA `1b5ab3e`)
- Industry search+harness (July, old stack): **49.8%** mean
- Mem0 Platform **92.5%** is n=1540, top-k 200, **their** harness — not a same-pin

## Freeze checklist (when a remasure is requested)

1. Held-out compiler audits green (named-subject / addressee / `she`/`he` last named person).
2. OpMem 13/13 and marketing 17/17 non-reg on the candidate SHA.
3. Dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`.
4. Judge temp 0. Same answerer/judge IDs on both lanes.
5. Stratified 100–200 **before** 3×90 **before** n=1540.
6. Run **both** lanes on the same SHA. Do not paste 11.4% next to 92.5% as one bake-off.
7. 1×30 is diagnostic only. Full MH 7.4% is the MH pin until a new full run.
8. No LoCoMo-named product rules. No episode top-k stuffing to restore OD/SH.

## Commands (do not run here)

```text
# Product lane
BRAINY_USE_RECALL=1 python -m public.locomo.run_smoke \
  --eval-lane product-recall --conversations 1 --questions 30

# Industry lane (Mem0-style depth)
python -m public.locomo.run_smoke \
  --eval-lane industry-search --conversations 1 --questions 30
# top-k defaults to 200 on this lane
```

Full n=1540 and LME-20 quality wait on an explicit freeze. LME-500 / BEAM 1M are not quality claims.
