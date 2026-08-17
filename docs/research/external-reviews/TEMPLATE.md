# External review template

**Date:** YYYY-MM-DD  
**Source:** (agent / human / firm)  
**Adjudicator:**  
**Prompt used:** (link to self-review prompt / brief)

For the 2026-08-17 full-recall pass, use [2026-08-17-full-recall-self-review-prompt.md](./2026-08-17-full-recall-self-review-prompt.md) and answer **all eight** prompt questions (course, published metric, first product PR, fair Mem0 stack, skipped suites, next PRs, claims, kill list). The numbered labels below are examples; match the prompt you were given.

## Verdict (1 paragraph)

Keep / adjust / replace the current course. State the single most important next move.

## Answers to prompt questions

1. **Course:**  
2. **Published metric (`/recall` vs search+harness):**  
3. **First product PR (cite-facts vs R5 OD):**  
4. **Fair Mem0 stack:**  
5. **Skipped suites (LME-500 / BEAM 1M):**  
6. **Next 3–7 PRs:**  
7. **Claims allowed vs forbidden:**  
8. **Kill list confirmation:**  

## Findings table

| Finding | Accept / Modify / Reject | Code evidence | Action |
| --- | --- | --- | --- |

## Accepted next sequence

Ordered list. Each item: failure class + expected artifact/pin.

## Rejected / deferred

## Claims discipline check

- [ ] No invented LME / SOTA / BEAM 1M scores
- [ ] Gate 0 vs harden vs remasure pins not blended
- [ ] 1×30 70% not published as full LoCoMo
- [ ] 11.4% labeled product `/recall`; 49.8% labeled search+harness (not current)
- [ ] Vendor 90+ labeled n / metric / top-k / path
- [ ] Dip honesty preserved where scores fell

## Artifact diffs required

- [ ] assessment pack
- [ ] codebase graph md/json
- [ ] program-execution-status
- [ ] PoR adoption note
- [ ] external-reviews/README.md priority pointer

## Linked PRs / commits
