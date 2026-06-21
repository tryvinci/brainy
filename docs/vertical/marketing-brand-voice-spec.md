# Marketing Brand Voice Memory Spec

**Status:** Approved for implementation (ENG-81)  
**Depends on:** ENG-85 verticalization skeleton, `packs/marketing/v1/pack.yaml`  
**Feeds:** ENG-82 eval fixtures  
**Last updated:** 2026-06-19

## Scope

Define how **brand voice** is represented in the general vertical runtime using two cognitive primitives:

| Primitive | Pack label | Role |
|---|---|---|
| **Principle** | `brand_rule` | Immutable taboos and hard constraints |
| **IdentityPrior** | `voice_profile` | Persistent tone, style, vocabulary |

Soft preferences without brand authority use label `preference` (also IdentityPrior primitive, lower precedence in conflicts).

---

## Override hierarchy (retrieval)

When multiple memories match a query, rank in this order:

```
1. Principle (brand_rule)     — weight 100 in pack
2. IdentityPrior (voice_profile) — weight 80
3. IdentityPrior (preference)    — weight 80, same primitive; disambiguate by label in explain
4. Belief, Outcome, Episode…     — per pack rank_policy
5. Suppressed / archived         — excluded
```

**Invariant:** A Principle must never rank below a soft preference on the same query when both match.

Implementation: `internal/memory/service.go` — `applyPrimitiveBonus()` uses `rank_policy.primitive_weights` from the active vertical pack; kind-based bonuses are skipped when `primitive` is `principle` or `identity_prior`.

---

## Classification (ingest → label + primitive)

Classification runs after deterministic extraction (`kind: profile|preference|fact`) and before persist.

### Current rules (MVP-1)

| Signal | Extractor rule | Pack mapping | Result |
|---|---|---|---|
| Sentence starts with "Never …" | `constraint_never` | `brand_rule` / `principle` | Principle |
| Contains "prefer(s)" | `preference_prefer` | `voice_profile` / `identity_prior` | IdentityPrior |
| Sentiment like/dislike | `preference_sentiment` | `voice_profile` / `identity_prior` | IdentityPrior |

Code: `internal/memory/extractor.go` + `internal/memory/vertical.go` (`ApplyVerticalPack`).

### Planned rules (MVP-1.1 — pack-driven)

Move heuristics into `packs/marketing/v1/pack.yaml` under `classification_rules`:

```yaml
classification_rules:
  - match: { prefix: "never " }
    label: brand_rule
    primitive: principle
  - match: { contains: "do not claim" }
    label: brand_rule
    primitive: principle
  - match: { contains: "our voice is" }
    label: voice_profile
    primitive: identity_prior
  - match: { kind: preference }
    label: voice_profile
    primitive: identity_prior
```

First matching rule wins. Keeps marketing logic in the pack, not Go.

---

## Storage shape

No separate tables. Each memory row:

```json
{
  "vertical": "marketing",
  "label": "brand_rule",
  "primitive": "principle",
  "kind": "fact",
  "content": "Never mention competitor X in any copy",
  "lifecycle_state": "active",
  "metadata": {}
}
```

Voice profile example:

```json
{
  "vertical": "marketing",
  "label": "voice_profile",
  "primitive": "identity_prior",
  "kind": "preference",
  "content": "Prefers warm, concise, second-person copy",
  "metadata": {
    "tone": ["warm", "concise"],
    "person": "second"
  }
}
```

Optional structured fields live in `metadata` (validated by pack schema when present).

---

## API usage

### Ingest

```http
POST /ingest
Content-Type: application/json

{
  "tenant_id": "agency1",
  "subject_id": "brand-acme",
  "vertical": "marketing",
  "source_type": "brand_guidelines",
  "messages": [
    { "role": "user", "content": "Our voice is warm but authoritative. Never use slang." }
  ]
}
```

Expect two memories: one `voice_profile`, one `brand_rule` (after sentence splitting).

### Search

```http
GET /memories/search?tenant_id=agency1&subject_id=brand-acme&vertical=marketing&q=voice%20tone
```

Response includes `explain.primitive` and `explain.primitive_bonus` when pack weights apply.

---

## Principle governance

Principles are **immutable by default** (Brainy architecture invariant).

| Action | Allowed | Mechanism |
|---|---|---|
| User corrects typo in brand rule | Yes | `POST /memories/{id}/correct` — lineage in `correction_history` |
| User softens a taboo | No — retire, don't edit | Suppress old + ingest new rule (future: governance endpoint) |
| Brand refresh | Explicit | Bulk suppress/archive old Principles; ingest new set with provenance |
| Accidental override by preference | Prevented | Rank hierarchy + eval tests |

**Phase 2:** `POST /memories/{id}/retire` with `reason_code` and audit event for Principle changes.

---

## Failure modes & mitigations

| Failure | Mitigation | Eval |
|---|---|---|
| Taboo leaks under paraphrase | Suppression + semantic search tests | ENG-82 #2 |
| Preference overrides brand rule | Primitive rank weights | ENG-82 #1 |
| Voice drift after rebrand | Supersession / archive old voice_profile | ENG-82 #4 |
| Multi-brand bleed | `tenant_id` + `subject_id` isolation | ENG-82 #7 |
| "Never" not extracted | `constraint_never` extractor rule | unit test |

---

## Eval scenarios (for ENG-82)

| ID | Scenario | Pass criteria |
|---|---|---|
| BV-01 | Ingest preference + brand_rule; search "competitor" | `brand_rule` ranks first |
| BV-02 | Search paraphrase of taboo topic | Suppressed rule never appears |
| BV-03 | Correct voice_profile; search paraphrase | Corrected content sticks |
| BV-04 | Ingest new voice after suppressing old | Only new voice surfaces |
| BV-05 | Generic `core` ingest unaffected | No primitive bonus without vertical |

---

## Implementation checklist

- [x] Pack vocabulary: `brand_rule`, `voice_profile`
- [x] Rank weights in pack YAML
- [x] `ApplyVerticalPack` on sync + async ingest
- [x] Search `vertical` param + primitive bonus
- [x] Extractor: `Never …` → fact → Principle mapping
- [x] Test: `TestVerticalPackPrincipleRanksAbovePreference`
- [ ] Pack-driven `classification_rules` (MVP-1.1)
- [ ] `voice_profile` metadata schema in pack YAML
- [ ] ENG-82 fixtures BV-01–BV-05

---

## References

- `docs/vertical/marketing-use-case-map.md` § Brand Voice Agent
- `docs/brainy/architecture/00-cognitive-primitives.md`
- `packs/marketing/v1/pack.yaml`
- `docs/vertical/verticalization-model.md`
