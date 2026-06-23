# Marketing MVP Benchmark Report

- **Benchmark:** `marketing-mvp-v1`
- **Generated:** 2026-06-23T00:36:38Z
- **Mem0 reference commit:** `a670333d67be1207b5be2fc73af60c3439444f48`

## Summary

| Suite | Brainy pass | Total | Mem0 expected |
| --- | ---: | ---: | --- |
| Parity | 4 | 4 | pass |
| Vertical (marketing) | 11 | 11 | fail |

**MVP ready:** yes

**Differentiation score:** 5/5 capabilities where Brainy passes and Mem0 lacks equivalent behavior.

## Capabilities vs Mem0

| Capability | Brainy | Mem0 | Differentiation |
| --- | --- | --- | --- |
| Principle ranks above soft preference | pass | no | yes |
| Suppressed taboo never resurfaces | pass | approx | no |
| Marketing vertical maps preferences to voice_profile | pass | no | yes |
| Editorial correction persists in retrieval | pass | approx | no |
| Core ingest unaffected by marketing pack weights | pass | yes | no |
| Subject isolation prevents cross-brand bleed | pass | approx | no |
| Duplicate marketing ingest dedupes | pass | yes | no |
| Never-sentences classify as brand_rule / principle | pass | no | yes |
| Response-style queries rank marketing preferences | pass | approx | no |
| Multi-message ingest extracts voice + rule | pass | no | yes |
| Archived campaign excluded from search | pass | no | yes |

## Fixture detail

### parity

- `dark_mode_vim_preference` — **pass**
- `factual_context` — **pass**
- `profile_lookup` — **pass**
- `response_style_preference` — **pass**

### vertical

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

## Interpretation

- **Parity suite** exercises Mem0-like ingest/search/dedupe behavior Brainy must not regress.
- **Vertical suite** exercises marketing pack capabilities Mem0 does not model.
- **Differentiation** = Brainy passes and `mem0_has: false` in the capability matrix.

Reproduce:

```bash
go run ./cmd/api  # with Postgres
python3 evals/run_marketing_mvp_benchmark.py --base-url http://127.0.0.1:8080
```
