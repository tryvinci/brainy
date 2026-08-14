"""Oracle / stage-diagnosis helpers for Brainy evals (program §16.3)."""

from __future__ import annotations

from typing import Any

# Mirror internal/memory/trace.go failure taxonomy.
FAILURE_LABELS = (
    "SOURCE_MISS",
    "WRITE_MISS",
    "REPRESENTATION_MISS",
    "ENTITY_LINK_MISS",
    "RELATION_MISS",
    "RETRIEVAL_MISS",
    "EVIDENCE_COVERAGE_MISS",
    "TEMPORAL_RESOLUTION_MISS",
    "CONFLICT_RESOLUTION_MISS",
    "PLANNING_MISS",
    "PROOF_MISS",
    "READER_MISS",
    "ABSTENTION_MISS",
    "JUDGE_MISS",
    "HARNESS_ERROR",
)


def _gold_text(gold: Any) -> str:
    if gold is None:
        return ""
    if isinstance(gold, (list, tuple)):
        return " ".join(str(x) for x in gold if x is not None).strip()
    return str(gold).strip()


def gold_in_texts(gold: Any, texts: list[str] | tuple[str, ...] | str | None) -> bool:
    needle = _gold_text(gold).lower()
    if not needle:
        return False
    if texts is None:
        return False
    if isinstance(texts, str):
        return needle in texts.lower()
    return any(needle in (t or "").lower() for t in texts)


def classify_failure(
    *,
    source_present: bool,
    semantic_present: bool | None,
    retrieved: bool,
    coverage_ok: bool | None,
    answer_ok: bool,
    abstained: bool,
    expected_abstain: bool = False,
    gold_in_facts: bool | None = None,
    gold_in_episodes: bool | None = None,
) -> str:
    """Return a primary failure label for a single question.

    READER_MISS only when the semantic representation needed to answer was
    present. Gold sitting in a chat turn is WRITE_MISS, not READER_MISS.
    """
    if not source_present:
        return "SOURCE_MISS"
    if gold_in_facts is False:
        return "WRITE_MISS"
    if semantic_present is False:
        return "WRITE_MISS"
    if not retrieved:
        return "RETRIEVAL_MISS"
    if coverage_ok is False:
        return "PROOF_MISS"
    if expected_abstain and not abstained:
        return "ABSTENTION_MISS"
    if not answer_ok:
        return "READER_MISS"
    return ""


def recall_body(
    tenant_id: str,
    subject_id: str,
    query: str,
    *,
    mode: str = "answer",
    top_k: int = 30,
    oracle_mode: str = "",
    include_historical: bool = False,
) -> dict[str, Any]:
    body: dict[str, Any] = {
        "tenant_id": tenant_id,
        "subject_id": subject_id,
        "q": query,
        "mode": mode,
        "top_k": top_k,
        "include_historical": include_historical,
    }
    if oracle_mode:
        body["oracle_mode"] = oracle_mode
    return body
