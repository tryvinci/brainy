# Mem0 parity gaps — where we win, where we fail, what to crush

**Updated:** 2026-07-25 · **Doctrine:** product gaps, not benchmax.  
**North star:** surpass Mem0 multi-axis ([path-to-sota.md](./path-to-sota.md)).

---

## Executive answer: are we at parity?

| Axis | Brainy | Mem0 (known) | Verdict |
| --- | --- | --- | --- |
| Thin-slice API parity | **4/4** | **4/4** | **Parity** |
| Operational correctness (OpMem) | **12/12** | **9/12** | **Ahead** |
| Vertical / governed memory | **16/16** marketing | No equivalent | **Ahead (category)** |
| Conversational recall (LOCOMO) | **19/30** smoke (gpt-oss, 1×30) | ~92 platform (GPT, top-200, full suite) | **Behind — not comparable pins** |
| Latency / tokens | p50 ~0.7–1.3s (staging) | Blog ~0.88s p50 @ top-200 | **Unknown fair pin** |
| Supersession / graph | v1 supersede API | Platform graph + multi-signal | **Partial** |

**Bottom line:** We are **at/above parity on operational + vertical** tracks. We are **not** at parity on the **conversational LOCOMO headline**. Closing that gap (fair pins) is the main engineering fight; publishing OpMem is how we already surpass them on production failure modes.

---

## Where we fail (conversational) — taxonomy from staging 19/30

Fail taxonomy on `locomo-staging-diversify-v1` (11 misses):

| Bucket | Count | Meaning |
| --- | ---: | --- |
| `retrieval_miss` | 8 | GT key tokens absent from top-15 (or never stored) |
| `synthesis_incomplete` | 2 | Evidence partially present; answer under-aggregates |
| `empty_answer` | 1 | Answerer returned don't-know / empty |

### Multi-hop misses (8/10 wrong)

| ID | Question | GT | Root cause (product) |
| --- | --- | --- | --- |
| q4 | Caroline identity | Transgender woman | Attribute never stated as searchable fact; only adjacent “trans*” talk |
| q7 | Relationship status | Single | Fact likely absent from extractable memories |
| q11 | Moved from | Sweden | Country never grounded as memory content |
| q15 | Melanie activities | pottery, camping, painting, swimming | Partial retrieval; incomplete list aggregation |
| q18 | Camped where | beach, mountains, forest | Locations not co-located with “camped” in top results |
| q19 | Kids like | dinosaurs, nature | Distinctive likes not surfaced |
| q23 | Books read | 2 titles | Title span missing / not extracted |
| q24 | Destress | Running, pottery | Partial (running); pottery not joined |

**Gap theme:** ingest does not emit **atomic attribute facts** (identity, status, origin, titled works, activity inventory); ranking + answerer cannot invent missing atoms.

---

## Capability gap vs Mem0 algorithm (product)

| Mem0 v3 technique | Brainy | Gap issue |
| --- | --- | --- |
| Single-pass ADD extract | Done (async provider + episodes) | Improve attribute atoms — **GAP-C1** |
| Multi-signal fusion (dense+BM25+entity) | Dense+token; IDF/entity gated OFF | Staging re-tune default-on — **GAP-C2** |
| Entity linking + boost | Extracted; ranking gated | Safe default-on after re-tune — **GAP-C2** |
| Temporal reasoning | `observed_at` + enrich | Stronger when-intent; relative plans — **GAP-C3** |
| Graph / multi-hop | Session + subject bridge + list MMR | Attribute graph + multi-span synth — **GAP-C4** |
| Top-200 retrieval budget | Eval top_k=30 | Raise product/eval budget with latency SLOs — **GAP-C5** |
| Fair GPT-class judge compare | gpt-oss only | Same-pin Mem0 LOCOMO re-run — **GAP-M1** |
| Supersession | ENG-86 v1 | Pack auto-rules / contradiction — **GAP-A1** (ahead on OpMem already) |

---

## Issue backlog (GitHub — Linear MCP needs re-auth to mirror)

| Gap | Issue | Status |
| --- | --- | --- |
| GAP-C1 Attribute atoms | [#50](https://github.com/tryvinci/brainy/issues/50) | In progress (deterministic atoms + provider prompt) |
| GAP-M1 Fair Mem0 LOCOMO | [#51](https://github.com/tryvinci/brainy/issues/51) | In progress (`--system mem0` backend) |
| GAP-C4 Multi-span synthesis | [#52](https://github.com/tryvinci/brainy/issues/52) | Open |
| GAP-C2 IDF/entity default-on | [#53](https://github.com/tryvinci/brainy/issues/53) | Open |
| GAP-C3 Temporal plans | [#54](https://github.com/tryvinci/brainy/issues/54) | Open |
| GAP-C5 Budget + latency SLO | [#55](https://github.com/tryvinci/brainy/issues/55) | Open |
| GAP-A1 Supersession v2 | [#56](https://github.com/tryvinci/brainy/issues/56) | Open (ENG-86 remainder) |
| GAP-P1 OpMem Paper 1 | [#57](https://github.com/tryvinci/brainy/issues/57) | Open |

**Linear:** authenticate the Linear MCP in Cursor to clone these as ENG-* and edit ENG-86/ENG-76.

---

## Measured same-pin attempt (2026-07-29)

| System | LOCOMO 1×30 (gpt-oss, top_k=30) |
| --- | ---: |
| Brainy (diversify peak) | **19/30** |
| Brainy (attribute atoms, solo) | 15/30 |
| Mem0 Platform (first attempt) | 1/30 (under-indexed — waiter fixed; re-run pending) |

Do **not** cite Mem0 1/30 as a capability claim. Cite Brainy OpMem 12/12 vs Mem0 9/12 for operational lead.

## Crush order (this cycle)

1. **GAP-C1** attribute extraction (code)  
2. **GAP-M1** fair Mem0 LOCOMO adapter (measure)  
3. **GAP-C4** multi-span synthesis  
4. Wire issues to Linear ENG-* when authenticated  

---

## Reproduce current Brainy pin

```bash
export BRAINY_BASE_URL=… BRAINY_API_KEY=… LLM_BASE_URL=… LLM_API_KEY=… LLM_MODEL=…
cd evals && python -m public.locomo.run_smoke \
  --conversations 1 --questions 30 --top-k 30 \
  --answerer-model "$LLM_MODEL" --judge-model "$LLM_MODEL"
python3 evals/run_opmem.py --systems brainy --base-url "$BRAINY_BASE_URL"
```
