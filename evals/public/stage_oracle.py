"""Stage-oracle helpers and failure ledger writer (external review PR1)."""

from __future__ import annotations

import json
import time
from pathlib import Path
from typing import Any

from oracle import FAILURE_LABELS, classify_failure, recall_body


def write_failure_record(
    path: str | Path,
    *,
    dataset: str,
    question_id: str,
    question: str,
    primary: str,
    secondary: str | None = None,
    flags: dict[str, Any] | None = None,
    notes: str = "",
) -> None:
    if primary and primary not in FAILURE_LABELS and primary != "":
        raise ValueError(f"unknown failure label: {primary}")
    rec = {
        "ts": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "dataset": dataset,
        "question_id": question_id,
        "question": question,
        "primary": primary,
        "secondary": secondary or "",
        "flags": flags or {},
        "notes": notes,
    }
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as f:
        f.write(json.dumps(rec, ensure_ascii=True) + "\n")


def oracle_recall_request(
    tenant_id: str,
    subject_id: str,
    query: str,
    oracle_mode: str,
    **kwargs: Any,
) -> dict[str, Any]:
    body = recall_body(tenant_id, subject_id, query, oracle_mode=oracle_mode, **kwargs)
    return body


def label_from_oracle_response(oracle_mode: str, resp: dict[str, Any]) -> str:
    """Map a product oracle response into a coarse stage label."""
    explain = resp.get("explain") or {}
    if explain.get("oracle_unsupported"):
        return "HARNESS_ERROR"
    if oracle_mode == "evidence":
        n = int(explain.get("oracle_evidence_count") or 0)
        return "" if n > 0 else "SOURCE_MISS"
    status = (resp.get("answer_status") or "").lower()
    if status in {"not_found", "insufficient_evidence"}:
        return "RETRIEVAL_MISS"
    return ""


__all__ = [
    "FAILURE_LABELS",
    "classify_failure",
    "write_failure_record",
    "oracle_recall_request",
    "label_from_oracle_response",
]
