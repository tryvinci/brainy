# Marketing MVP Benchmark Report

- **Benchmark:** `marketing-mvp-v1`
- **Generated:** 2026-07-29T22:03:10Z
- **Mem0 mode:** `declared` (static matrix only)
- **Mem0 reference commit:** `a670333d67be1207b5be2fc73af60c3439444f48`

## Summary

| Suite | Brainy | Mem0 |
| --- | ---: | ---: |
| Parity | 4/4 | pass |
| Vertical (marketing) | 16/16 | fail |

**MVP ready (Brainy):** yes

**Differentiation score:** 5/5 capabilities where Brainy passes and Mem0 does not.

## Capabilities vs Mem0

| Capability | Brainy | Mem0 declared | Mem0 empirical | Differentiation |
| --- | --- | --- | --- | --- |
| Principle ranks above soft preference | pass | no | n/a | yes |
| Suppressed taboo never resurfaces | pass | approx | n/a | no |
| Marketing vertical maps preferences to voice_profile | pass | no | n/a | yes |
| Editorial correction persists in retrieval | pass | approx | n/a | no |
| Core ingest unaffected by marketing pack weights | pass | yes | n/a | no |
| Subject isolation prevents cross-brand bleed | pass | approx | n/a | no |
| Duplicate marketing ingest dedupes | pass | yes | n/a | no |
| Never-sentences classify as brand_rule / principle | pass | no | n/a | yes |
| Response-style queries rank marketing preferences | pass | approx | n/a | no |
| Multi-message ingest extracts voice + rule | pass | no | n/a | yes |
| Archived campaign excluded from search | pass | no | n/a | yes |

## Fixture detail

### parity — Brainy

- `dark_mode_vim_preference` — **pass**
- `factual_context` — **pass**
- `profile_lookup` — **pass**
- `response_style_preference` — **pass**

### vertical — Brainy

- `bv01_principle_over_preference` — **pass**
- `bv02_suppression_leak` — **pass**
- `bv03_voice_preference_ingest` — **pass**
- `bv04_correction_stickiness` — **pass**
- `bv05_core_vertical_unaffected` — **pass**
- `bv06_multi_brand_isolation` — **pass**
- `bv07_dedupe_marketing` — **pass**
- `bv08_brand_rule_extraction` — **pass**
- `bv09_response_style_marketing` — **pass**
- `bv10_dual_message_brand_voice` — **pass**
- `lc01_archived_campaign_hidden` — **pass**
- `lc02_active_campaign_ranks_above_completed` — **pass**
- `ob05_outcome_updates_belief_rank` — **pass**
- `pt08_cross_campaign_pattern` — **pass**
- `sg10_scoped_segment_prefs_coexist` — **pass**
- `ts09_style_matched_creative_ranks` — **pass**

## Interpretation

- **Parity suite** — Mem0-like ingest/search/dedupe; both systems should mostly pass.
- **Vertical suite** — marketing pack behavior (Principle > preference, lifecycle, etc.).
- **Declared** `mem0_has` comes from the capability matrix (design expectation).
- **Empirical** runs the *same fixture JSON* against Mem0 Platform when `--systems` includes `mem0`.
- Fixtures that require `explain.primitive` / pack labels will fail on Mem0 by design — that is the moat.

Reproduce:

```bash
# Brainy only (declared Mem0 gaps)
python3 evals/run_marketing_mvp_benchmark.py --base-url "$BRAINY_BASE_URL"

# True counter-run (requires MEM0_API_KEY)
python3 evals/run_marketing_mvp_benchmark.py --base-url "$BRAINY_BASE_URL" \
  --systems brainy,mem0
```
