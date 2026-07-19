# Public proveable eval framework

In-repo harness for **LOCOMO / LongMemEval-style** runs against Brainy, with Mem0-compatible `UnifiedResult` JSON and hard proveability pins.

Peer reference: [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks)  
Dataset: [snap-research/locomo](https://github.com/snap-research/locomo) · [ACL 2024](https://aclanthology.org/2024.acl-long.747/)

## Design rules

1. **Pin** dataset URL + content SHA256, Brainy URL/commit, judge model, temperature=0  
2. **Trace** every question: retrieval → generation → judgment  
3. **Record** search latency; optional tokens when LLM answerer is used  
4. **Never invent scores** — missing OpenAI key forces lexical mode labeled `lexical-overlap-v0` (not a publishable J-score)  
5. Score **categories 1–4**; exclude adversarial (5) from overall (industry default)

`RunManifest.require_pins()` returns gaps; empty gaps = publishable pins present.

## Layout

```
evals/public/
  schema.py          # UnifiedResult / Metrics
  proveability.py    # RunManifest + sha256 + require_pins
  judge.py           # lexical + OpenAI binary judge
  backends/brainy.py # /ingest (+ async) + /memories/search
  locomo/
    dataset.py       # download locomo10.json
    run_smoke.py     # L3 smoke runner (default: async ingest)
```

## Quickstart (LOCOMO smoke)

```bash
# Dataset downloads to datasets/locomo/ (gitignored)
export BRAINY_BASE_URL=https://brainy-api-staging.onrender.com
# export BRAINY_API_KEY=...   # if staging auth on

# Cloudflare AI Gateway (Workers AI — recommended)
export LLM_BASE_URL="https://gateway.ai.cloudflare.com/v1/<account>/<gateway>/compat"
export LLM_API_KEY="$CF_AIG_TOKEN"
export LLM_MODEL="workers-ai/@cf/openai/gpt-oss-120b"
# Alternate: workers-ai/@cf/meta/llama-3.3-70b-instruct-fp8-fast

cd evals
# Default: POST /ingest/async so provider extract on the worker is measured.
python -m public.locomo.run_smoke \
  --conversations 1 \
  --questions 30 \
  --out-dir ../docs/benchmarks/runs

# Escape hatch: deterministic sync path only
# python -m public.locomo.run_smoke --sync-ingest --conversations 1 --questions 30
```

| Choice | Model ID | When |
| --- | --- | --- |
| **Primary** | `workers-ai/@cf/openai/gpt-oss-120b` | Best open answerer+judge on this gateway |
| **Alternate** | `workers-ai/@cf/meta/llama-3.3-70b-instruct-fp8-fast` | Faster / widely cited 70B |
| **Frontier MoE** | `workers-ai/@cf/zai-org/glm-5.2` or `.../kimi-k2.6` | Max-quality experiments |

Avoid for LOCOMO: `kimi-k2.7-code`, QwQ/R1-distill (CoT), image/TTS/embed models, ≤8B.

Lexical-only (CI / harness prove, not public J-score):

```bash
python -m public.locomo.run_smoke --lexical-only --conversations 1 --questions 5
```

**Proveability note:** pin `LLM_BASE_URL` + full model id in the manifest. Scores are not comparable to Mem0 GPT-judged blog numbers.

Outputs: `{run_id}.json` (UnifiedResult), `{run_id}.manifest.json`, markdown report.

## Ladder mapping

| Layer | Status | Entry |
| --- | --- | --- |
| L2 adapter | this package | `backends/brainy.py` |
| L3 smoke | `locomo/run_smoke.py` | subset Qs |
| L4 full | TODO | all 10 convos + blog |
| L5 LongMemEval | TODO | same schema |
| L6 MarketingMem | vertical fixtures | sibling track |

See [docs/research/proveable-eval-framework.md](../../docs/research/proveable-eval-framework.md).
