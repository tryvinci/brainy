# Proveable public eval framework

**Status:** L2 shipped in-repo (`evals/public/`); L3 smoke runnable  
**Peers:** [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks) · [snap-research/locomo](https://github.com/snap-research/locomo)

This document is the contract for any number we put on a research page or launch post.

---

## What “proveable” means

A published score is **proveable** only if a third party can:

1. Download the **exact dataset bytes** (URL + SHA256)
2. Point at the **exact system** (Brainy commit + base URL; competitor API date if any)
3. Re-run **retrieval → answer → judge** with pinned models and `temperature=0`
4. Diff against our **UnifiedResult** JSON (per-question traces)

If any pin is missing, `require_pins()` fails and we must not claim a public number.

---

## Anti-benchmax doctrine (product first)

Public suites (LOCOMO, LongMemEval, …) are **diagnostic customers**, not targets to game.

| Do | Do not |
| --- | --- |
| Treat every fail as a **Brainy product bug** (extract, rank, temporal, API) | Special-case LOCOMO speakers, categories, or dataset IDs in product code |
| Improve capabilities that help **all** conversational / long-memory apps | Eval-only transcript stuffing at query time while real search stays weak |
| Keep pin + judge settings stable across product iterations | Soften the judge, invent answers, or swap to lexical scoring for nicer posts |
| Change the harness only when it better mirrors **real client usage** of the public API | “Win LOCOMO” patches that do not ship to production clients |

**Smoke verdict (2026-07-15):** 2/30 was **28/30 retrieval miss of ground-truth span** — keyword extractor drops casual dialogue facts / detaches dates; topical hash-hybrid ranking surfaces neighbors; no event-time fields. Next work is product ([ENG-92](https://linear.app/engramhq/issue/ENG-92) extraction, ranking, temporal), not harness tweaks.

Re-run LOCOMO **after** product lands to measure lift — same pins, same subset — so score deltas are attributable to Brainy.

---

## UnifiedResult (schema compatible)

Shape mirrors Mem0’s harness `UnifiedResult` (schema_version `1.0`):

| Block | Purpose |
| --- | --- |
| `metadata` | benchmark id, dataset URL/SHA, brainy URL/commit, answerer/judge models |
| `metrics` | overall accuracy, by-group, search p50/p95 |
| `evaluations[]` | question, GT, retrieval results + latency, generated answer, judgment |

Scored categories: **1–4** (multi-hop, temporal, open-domain, single-hop). Category **5** (adversarial) is traced but excluded from overall — same convention as Mem0 blog reports.

---

## Modes

| Mode | When | Publishable as LOCOMO J-score? |
| --- | --- | --- |
| `lexical-overlap-v0` + `retrieval-concat-v0` | no LLM / `--lexical-only` | **No** — harness prove only |
| Open-weight or OpenAI via `LLM_BASE_URL` | pinned model + base URL, temp 0 | **Yes** (cite model; not comparable across judges) |
| Full LOCOMO (L4) | all convos + blog | **Yes**, with Mem0 re-run or explicit cite |

Never paste competitor blog % into our cells as if we measured them.

### Recommended local open setup (Ollama)

```bash
ollama pull llama3.1
export LLM_BASE_URL=http://127.0.0.1:11434/v1
export LLM_MODEL=llama3.1
```

Stronger judge (optional): `qwen2.5:14b` or `llama3.1:70b` if you have the VRAM. Same family for answerer + judge keeps the pin simple.


---

## Reproduce block (copy into posts)

```bash
git checkout <brainy_commit>
export BRAINY_BASE_URL=...
export BRAINY_API_KEY=...          # if required
export LLM_BASE_URL=http://127.0.0.1:11434/v1
export LLM_MODEL=llama3.1

cd evals
python -m public.locomo.run_smoke \
  --conversations 1 --questions 30 \
  --out-dir ../docs/benchmarks/runs
```

Attach the `.manifest.json` and `.json` artifacts. Dataset SHA must match the post.

---

## Research outlinks (always cite)

- LOCOMO paper: https://aclanthology.org/2024.acl-long.747/
- LOCOMO data: https://github.com/snap-research/locomo
- Mem0 runners: https://github.com/mem0ai/memory-benchmarks
- SuperMemory research packaging: https://supermemory.ai/research/

---

## Linear track

Milestone **Public Proveable Eval Framework** under project *Memory System (Brainy)* — issues L2→L6. OpMem remains the CI merge gate; this ladder is research publication, not a blocking merge check until we say otherwise.

Ladder plan: [public-bench-ladder.md](./public-bench-ladder.md)
