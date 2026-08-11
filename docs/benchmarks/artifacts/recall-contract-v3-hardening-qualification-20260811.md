# Recall Contract V3 Hardening — qualification — 2026-08-11

**Umbrella branch:** `pr/v3-hardening-qualify-a6c7` · PR #98  
**Component PRs:** #93 subject-order · #94 authoritative ops · #95 LME product-recall · #96 hop executor · #97 sufficiency

## Gate 0 staging baseline (`9bad898`)

| Pin | Result |
| --- | --- |
| OpMem | 13/13 |
| Marketing | passed |
| LoCoMo 1×30 | **18/30** (MH 50%, OD 25%, temporal 75%) |
| LoCoMo 3×90 | **32/90** (MH 19.4%, OD 42.9%) |

## Post-hardening (local combined branch)

| Pin | Result |
| --- | --- |
| OpMem | 13/13 |
| Marketing | passed |
| LoCoMo 1×30 | **14/30** (MH 5/10, OD 2/4) |
| LME-20 product-recall | Path proven (`answer_path=/recall`); full score resume |

## Claims discipline

- Allowed: Gate 0 staging 18/30 and 32/90; product-recall path proven; OpMem/marketing non-reg green on harden.
- Forbidden: “beats Mem0”; SOTA; “MH solved”; publishable LME accuracy before full `--publish --product-recall` run completes; calling harden 1×30 an improvement (it is not).

## Follow-ups

1. Merge #93→#97 to `dev` (staging), re-pin staging.
2. Finish isolated LME-20; LME-100 only after that.
3. Mem0 same-pin + multi-seed LoCoMo after staging re-pin.
4. Do **not** merge to `main` without explicit ask.
