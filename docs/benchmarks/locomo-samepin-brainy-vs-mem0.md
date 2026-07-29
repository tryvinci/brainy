# LOCOMO same-pin: Brainy vs Mem0 (gpt-oss)

**Pins (identical):** 1 conversation / 30 Q · top_k=30 · gpt-oss-120b answerer+judge · temp=0  
**Dataset:** locomo10.json (same SHA as smoke harness)

| System | Overall | temporal | multi-hop | open-domain | Notes |
| --- | ---: | ---: | ---: | ---: | --- |
| **Brainy** (diversify peak) | **19/30** | 13/16 | 2/10 | 4/4 | staging dense emb |
| Brainy (recent atoms/synth) | 14–15/30 | ~9–10/16 | 2–3/10 | 3/4 | run variance |
| **Mem0 Platform** (hardened wait) | **12/30** | 2/16 | **6/10** | 4/4 | `locomo-mem0-samepin-v2` |

## Takeaways

1. Under **the same judge/budget**, Brainy’s best pin (**19/30**) beats Mem0 (**12/30**) overall.
2. Mem0 is **stronger on multi-hop (6/10 vs ~2–3/10)** — that is our primary product gap to crush (#50/#52).
3. Brainy is **stronger on temporal** in this pin (Mem0 2/16 — weak event-time grounding without our `observed_at` path).
4. Do **not** compare either number to Mem0’s blog ~92 (different judge, top_200, full suite).

## Reproduce

```bash
# Brainy
python -m public.locomo.run_smoke --system brainy --conversations 1 --questions 30 --top-k 30 \
  --answerer-model "$LLM_MODEL" --judge-model "$LLM_MODEL"

# Mem0 (MEM0_API_KEY)
python -m public.locomo.run_smoke --system mem0 --conversations 1 --questions 30 --top-k 30 \
  --async-timeout 2400 --answerer-model "$LLM_MODEL" --judge-model "$LLM_MODEL"
```
