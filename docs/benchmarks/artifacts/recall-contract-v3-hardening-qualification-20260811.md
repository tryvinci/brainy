# Recall Contract V3 Hardening — qualification — 2026-08-11

**Umbrella:** PRs #93–#98 merged → `dev` `1f2f26f` → production `main` `308d3a1`  
**Component PRs:** #93 subject-order · #94 authoritative ops · #95 LME product-recall · #96 hop executor · #97 sufficiency · #98 qualify docs  
**External self-review prompt:** [../../research/external-reviews/2026-08-11-hardening-self-review-prompt.md](../../research/external-reviews/2026-08-11-hardening-self-review-prompt.md)

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
| LoCoMo 1×30 | **14/30** (MH 5/10, OD 2/4) — honest dip vs Gate 0 |
| LME-20 product-recall | Path proven (`answer_path=/recall`); later publish run aborted on extraction job failure — **not publishable** |

## Post-cutover staging (`1f2f26f`) full pass

| Pin | Result |
| --- | --- |
| OpMem | **13/13** — [pin](./opmem-staging-postcutover-20260811.md) |
| Marketing | **passed** — [pin](./marketing-staging-postcutover-20260811.md) |
| LoCoMo 1×30 | **15/30 (50%)** · MH **50%** · OD **25%** · temporal **56.2%** — [pin](./locomo-staging-postcutover-1x30-pin-20260811.md) |
| LoCoMo 3×90 | in progress / see dated pin when written |

## Claims discipline

- Allowed: Gate 0 staging 18/30 and 32/90; harden local 14/30 with dip honesty; OpMem/marketing non-reg; product-recall path proven; hardening on `dev`+`main`.
- Forbidden: “beats Mem0”; SOTA; “MH solved”; publishable LME accuracy before a clean full `--publish --product-recall` run; calling harden 1×30 an improvement; calling 3×90 MH 50%.

## Follow-ups

1. Clean isolated LME-20 `--publish --product-recall`.
2. Finish staging LoCoMo re-pin on `1f2f26f` and record pins.
3. Mem0 same-pin + multi-seed LoCoMo.
4. External adjudication via the self-review prompt.
