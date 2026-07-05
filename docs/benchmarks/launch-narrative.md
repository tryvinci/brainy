# Brainy Launch Narrative — Benchmark-Led Public Launch

**Track B** · Published 2026-07-05  
**Audience:** Developers evaluating memory layers for marketing agents

---

## Headline

**Brainy passes 12/12 operational memory tasks and 16/16 marketing vertical fixtures** — with reproducible harnesses you can run locally in five minutes.

Generic memory APIs optimize for semantic search. Brainy optimizes for **governed marketing memory**: brand rules, campaign lifecycle, suppression durability, and outcome→belief ranking.

---

## What we measured

### OpMem v0 — operational correctness (12 tasks)

Structured memory must get **updates, suppressions, and isolation** right — not just embedding similarity.

| Dimension | Brainy | Verbatim RAG |
| --- | --- | --- |
| Suppression durability | **3/3** | 2/3 |
| Correction stickiness | **3/3** | 2/3 |
| Tenant/subject isolation | **3/3** | 3/3 |
| Staleness / preference updates | **2/2** | 2/2 |
| Idempotent ingest | **1/1** | 0/1 |
| **Total** | **12/12** | 9/12 |

Report: [opmem-baseline-report.md](./opmem-baseline-report.md)

### Marketing vertical moat (16 fixtures + Tier 4 seeds)

Capabilities Mem0 cannot model out of the box:

- Principle > preference hierarchy
- Brand rule extraction from voice conversations
- Campaign lifecycle suppression (`lc01`, `lc02`)
- Outcome → belief rank loop
- Scoped segment coexistence
- Hybrid paraphrase retrieval (deterministic embedder for CI)

Report: [marketing-moat-report.md](./marketing-moat-report.md) · [METHODOLOGY.md](./METHODOLOGY.md)

---

## Try it

```bash
git clone https://github.com/tryvinci/brainy.git && cd brainy
docker compose up --build -d
python3 evals/run_opmem.py --systems brainy,verbatim --base-url http://127.0.0.1:8080
python3 evals/run_vertical_eval.py --base-url http://127.0.0.1:8080
```

Tagged release: **`v0.1.0`** (developer preview) · Hosted beta: API key auth — see [commercial-beta-checklist.md](../commercial-beta-checklist.md)

---

## Honest limits

- Brainy uses a **deterministic local embedder** for CI reproducibility; Mem0 may win on provider-quality embeddings at scale until provider extraction ships.
- Finance vertical is **Gate M4 research** — not required for marketing launch.
- No “SOTA” claim without methodology — we publish fixtures, scores, and reproduction commands.

---

## Links

- Repo: https://github.com/tryvinci/brainy
- OpMem spec: [docs/research/opmem-spec.md](../research/opmem-spec.md)
- GTM roadmap: [docs/vertical/go-to-market-roadmap.md](../vertical/go-to-market-roadmap.md)
