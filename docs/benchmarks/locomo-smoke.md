# LOCOMO smoke — CF judge/answerer model matrix

**Timestamp:** 2026-07-23T21:55:00Z  
**Brainy:** staging (`a8f8d8b`)  
**Embeddings:** `workers-ai/@cf/baai/bge-base-en-v1.5` (768-d) on staging API+worker  
**Entity ranking:** OFF (`BRAINY_ENTITY_RANKING=false`)  
**Ingest:** async  
**Pin:** 1 conversation / 30 questions  

Same Brainy retrieval + memories; only the CF AI Gateway chat model for
**answerer + judge** changes. Eval client uses `curl/8.5.0` UA + `max_tokens=2048`
so gpt-oss reasoning models do not leave `content` null (WAF 1010 / reasoning budget).

## Model matrix

| Judge + answerer (CF Workers AI) | Overall | temporal | multi-hop | open-domain |
| --- | ---: | ---: | ---: | ---: |
| gpt-oss-120b (`LLM_MODEL` / Workers AI) | **14/30 (0.467)** | 9/16 | 2/10 | 3/4 |
| mistral-small-3.1-24b-instruct | 13/30 (0.433) | 6/16 | 3/10 | 4/4 |
| llama-3.3-70b-instruct-fp8-fast | 11/30 (0.367) | 6/16 | 2/10 | 3/4 |

Full `workers-ai/@cf/...` model IDs are set via env (`LLM_MODEL`,
`--answerer-model`, `--judge-model`); see `evals/public/README.md`.

## Takeaways

1. **Models matter, modestly** (~±3/30 on this pin). gpt-oss-120b is the best
   of the three; Llama 3.3 70B is weakest as both answerer and judge here.
2. **Multi-hop stays ~2–3/10 across all three** — swapping judges does not
   unlock multi-hop. That points at retrieval / synthesis, not judge severity.
3. **Temporal is where gpt-oss pulls ahead** (9/16 vs 6/16). Open-domain is
   already near-saturated (3–4/4).
4. Prior same-pin staging dense run with gpt-oss was **13/30**; this remeasure
   landed **14/30** (run variance ±1 on the 30-Q pin).

## Earlier staging dense baseline (entity OFF)

| Config | Overall |
| --- | ---: |
| Local hash baseline | 13/30 |
| Staging CF dense + gpt-oss (prior) | 13/30 |
| **Staging CF dense + gpt-oss (this matrix)** | **14/30** |

OpMem on staging after dense wiring: **12/12**.

## How to reproduce

```bash
export BRAINY_BASE_URL=… BRAINY_API_KEY=…
export LLM_BASE_URL=… LLM_API_KEY=…
# Set LLM_MODEL (and --answerer-model / --judge-model) to each CF Workers AI
# chat id listed in evals/public/README.md (gpt-oss-120b, llama-3.3-70b,
# mistral-small-24b).

python -m public.locomo.run_smoke \
  --conversations 1 --questions 30 \
  --answerer-model "$LLM_MODEL" --judge-model "$LLM_MODEL" \
  --run-id "locomo-staging-judge-${LABEL}"
```
