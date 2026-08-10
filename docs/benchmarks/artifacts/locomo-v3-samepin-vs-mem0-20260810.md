# Same-pin LoCoMo 1×30 — Brainy V3 vs Mem0 — 2026-08-10

Identical dataset pin (`locomo10.json`, conv budget 1×30, categories 1–4).

| System | Overall | temporal | multi-hop | open-domain | Artifact |
| --- | ---: | ---: | ---: | ---: | --- |
| **Brainy V3** (`4909832`, product `/recall`, hybrid on) | **16/30 (53.3%)** | 10/16 | 5/10 | 1/4 | [early pin](./locomo-v3-early-pin-20260810.md) |
| **Mem0 Platform** | **12/30 (40.0%)** | (see report) | **7/10 (70%)** | (see report) | [mem0 report](./locomo-mem0-samepin-v3-20260810.md) |

## Read

- Brainy wins **overall** on this pin (16 vs 12).
- Mem0 still leads **multi-hop** (70% vs 50%).
- Hybrid reader confirmed firing on Brainy (17/30 `hybrid_llm_packet`).
- **Not** a SOTA claim; MH gap remains the priority for typed-hop / update follow-through after merge to staging.
