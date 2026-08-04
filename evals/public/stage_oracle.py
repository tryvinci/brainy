"""Stage-oracle helpers and failure ledger writer (external review PR1)."""

from __future__ import annotations

import json
import time
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any

try:
    from oracle import FAILURE_LABELS, classify_failure, recall_body
except ImportError:  # package import (python -m public...)
    from public.oracle import FAILURE_LABELS, classify_failure, recall_body


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


def post_recall(base_url: str, body: dict[str, Any], *, api_key: str = "", timeout: float = 60.0) -> dict[str, Any]:
    """POST /recall; returns parsed JSON or {\"error\": ...}."""
    url = base_url.rstrip("/") + "/recall"
    data = json.dumps(body).encode("utf-8")
    headers = {"Content-Type": "application/json"}
    if api_key:
        headers["Authorization"] = f"Bearer {api_key}"
    req = urllib.request.Request(url, data=data, headers=headers, method="POST")
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        raw = e.read().decode("utf-8", errors="replace")
        try:
            return json.loads(raw)
        except json.JSONDecodeError:
            return {"error": {"status": e.code, "body": raw[:400]}}
    except Exception as e:  # noqa: BLE001 — harness must not crash on probe failure
        return {"error": {"message": str(e)}}


def label_from_oracle_response(oracle_mode: str, resp: dict[str, Any]) -> str:
    """Map a product oracle response into a coarse stage label."""
    if resp.get("error"):
        return "HARNESS_ERROR"
    explain = resp.get("explain") or {}
    if explain.get("oracle_unsupported"):
        return "HARNESS_ERROR"
    mode = (oracle_mode or "").lower()
    if mode == "evidence":
        n = int(explain.get("oracle_evidence_count") or 0)
        return "" if n > 0 else "SOURCE_MISS"
    if mode == "semantic":
        atoms = int(explain.get("oracle_atom_count") or 0)
        mems = int(explain.get("oracle_memory_count") or 0)
        if atoms > 0 or mems > 0:
            return ""
        return "REPRESENTATION_MISS"
    if mode == "retrieval":
        n = int(explain.get("oracle_memory_count") or 0)
        return "" if n > 0 else "RETRIEVAL_MISS"
    if mode == "coverage":
        n = int(explain.get("oracle_item_count") or 0)
        cov = resp.get("coverage") or {}
        if n > 0 or cov.get("satisfied"):
            return ""
        return "EVIDENCE_COVERAGE_MISS"
    status = (resp.get("answer_status") or "").lower()
    if status in {"not_found", "insufficient_evidence"}:
        return "RETRIEVAL_MISS"
    return ""


def probe_failure_stages(
    base_url: str,
    *,
    tenant_id: str,
    subject_id: str,
    query: str,
    answer_ok: bool,
    api_key: str = "",
) -> tuple[str, dict[str, Any]]:
    """
    Run evidence → semantic → retrieval → coverage probes and return
    (primary_label, flags). Empty primary means no diagnosed miss (reader/judge).
    """
    flags: dict[str, Any] = {"answer_ok": answer_ok}
    order = ("evidence", "semantic", "retrieval", "coverage")
    for mode in order:
        body = oracle_recall_request(tenant_id, subject_id, query, mode)
        resp = post_recall(base_url, body, api_key=api_key)
        label = label_from_oracle_response(mode, resp)
        flags[f"oracle_{mode}"] = {
            "label": label,
            "answer_status": resp.get("answer_status"),
            "explain": {
                k: (resp.get("explain") or {}).get(k)
                for k in (
                    "oracle_evidence_count",
                    "oracle_atom_count",
                    "oracle_memory_count",
                    "oracle_item_count",
                    "oracle_unsupported",
                )
                if (resp.get("explain") or {}).get(k) is not None
            },
        }
        if label:
            return label, flags
    if not answer_ok:
        return "READER_MISS", flags
    return "", flags


__all__ = [
    "FAILURE_LABELS",
    "classify_failure",
    "write_failure_record",
    "oracle_recall_request",
    "label_from_oracle_response",
    "post_recall",
    "probe_failure_stages",
]
