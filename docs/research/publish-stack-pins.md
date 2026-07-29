# Publish-stack model pins (master-plan §5.1)

**Dev / CI stack (default):**
- Answerer/judge: CF Workers AI via AI Gateway (`LLM_MODEL`, typically gpt-oss-120b)
- Embeddings: staging `BRAINY_EMBEDDING_MODEL=workers-ai/@cf/baai/bge-base-en-v1.5`
- Purpose: iteration, OpMem, smoke

**Publish stack (phase gates / public claims):**
- Answerer/judge: GPT-class OpenAI-compatible model (set `LLM_MODEL` / `--judge-model` explicitly)
- Same Brainy staging commit + embedding pins
- Full LoCoMo via `python -m public.locomo.run_full --conversations 10 --seeds 1`
- Budget: one dry run ≈ material API spend; confirm before multi-seed

**Holdout:** tuning = LOCOMO convs 1–3; validation = 4–10 at most once per phase gate (`runs-log.md`).
