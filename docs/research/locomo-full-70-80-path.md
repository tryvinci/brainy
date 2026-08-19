# LoCoMo Full 70–80% — execution path

**Status:** R5B–R10 representation stack landed on this branch (typed packets, `she`/`he` coref, canonical entity/relation IDs, hop ID joins, dual-path freeze wiring). Not a score claim.  
**Does not claim:** 70–80% on n=1540 this freeze; SOTA; beats-Mem0; 1×30 70% as full LoCoMo.

## Why 70–80% on n=1540 is two lanes

| Lane | Current pin | 70–80% meaning |
| --- | --- | --- |
| Product `POST /recall` | **175/1540 = 11.4%** (`1b5ab3e`) | Structured answers over compiled facts. Above July search+harness **49.8%**, so it needs **answer-path + representation**. Copula clip + R5A are answer-path. This file is the representation mass step. |
| Industry search → shared answerer → shared judge | July **49.8%** mean (old stack, top-k style) | Same band as Mem0 **paper** ~68.5 and full-context ~73%. Mem0 Platform **92.5%** is n=1540, **top-k 200**, their harness — not this path. |

1×30 conv-26 head **21/30 (70%)** is a diagnostic. Rest of conv-26 in the full run was **12/122 (9.8%)**. Full SH is **88/841 (10.5%)**. Publishing 70% as full LoCoMo is invalid.

## What actually blocks full SH

First-person speaker binding already compiles `Name: I researched X` → `Name researched X`. Full-suite questions are often **reports about someone else**:

- `Riley: Casey researched wildfire recovery` used to compile **Riley** researched (wrong subject) or nothing typed.
- `Riley: You realized that rest is part of training` never bound the addressee.
- `Morgan: Dana lives in Portland and is a carpenter` never compiled Dana.

That is `WRITE_MISS` / `REPRESENTATION_MISS`, not fusion weights, not episode top-k, not LoCoMo-named regexes.

R5A made `/recall` cite structured values. If the atom is bound to the wrong person, structured-first still answers the wrong person. Compiler subject binding is the mass lever for 841 single-hop items.

## This slice (R6a)

General linguistic forms only (`internal/memory/clause_subject.go`, `attribute_atoms.go`):

- Clause subject: first-person → speaker; `you` → two-party addressee or most recent other speaker; named person before the verb → that person; `she`/`he` → skip (R7 coref).
- Factive templates reused on that subject: researched, works as, lives in, realized that, is a (non-title), origin/move, events.
- Provider prompt: report-about-B and two-party `you` belong to B / addressee, not the reporter.
- Merge gate: held-out audit (`TestHeldOutNamedSubjectAndAddresseeAudit`), not a LoCoMo bump. `BRAINY_LEGACY_LOCOMO_ATOMS` stays default off.

Copula clip (`realized that` before adjective `is`) ships on the same branch so belief clauses keep their tails.

## What 70–80% still requires after this stack

The stack in this pass is the **substrate** for an honest later claim. It is not the claim.

1. **Compiler mass on the rest of SH** — named-subject + addressee + `she`/`he` last-named-person are in. Provider extract quality still depends on the model (`ContextualExtractor` already injects recent/related memories). Held-out audit stays the merge gate.
2. **Identity joins on full MH** — hops now join `ent:` IDs; unscoped predicate hits are context only. Full MH is still **7.4%** until a freeze remasure. Do not treat 1×30 10/10 as MH-solved.
3. **R10 freeze remasure** — product `/recall` *and* current-SHA search+harness (`--eval-lane`), labeled separately. Stratified 100–200 then 3×90; full n=1540 only after freeze. See [locomo-dual-path-freeze.md](./locomo-dual-path-freeze.md).
4. **Open-domain** stays a diagnostic. Do not restore OD/SH by stuffing episodes into top-k.
5. Industry-format 70–80% likely still needs an LLM answerer on retrieved **atoms**, not slogans. Product `/recall` 70–80% without that lane is not promised. Mem0 Platform 92.5% remains a different path (top-k 200, their harness).

## Kill list (unchanged)

No fusion fishing, no graph DB, no LoCoMo-named product rules, no spaCy requirement, no mixing 11.4% with Mem0 92.5%, no publishing 1×30 70% as full LoCoMo, no LME-500/BEAM 1M as a quality claim. Additive `memory_entities` / relation ID columns are dual-write (ADR-004), not a graph DB.
