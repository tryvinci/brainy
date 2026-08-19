"""Stage-oracle helpers and failure ledger writer (external review PR1 / R0)."""

from __future__ import annotations

import json
import time
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any

try:
    from oracle import FAILURE_LABELS, classify_failure, gold_in_texts, gold_semantically_in_texts, recall_body
except ImportError:  # package import (python -m public...)
    from public.oracle import FAILURE_LABELS, classify_failure, gold_in_texts, gold_semantically_in_texts, recall_body


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


def _query_token_overlap(query: str, texts: list[str]) -> float:
    import re

    q = {t for t in re.findall(r"[a-z0-9]+", (query or "").lower()) if len(t) > 2}
    if not q:
        return 1.0
    blob = " ".join(texts).lower()
    hit = sum(1 for t in q if t in blob)
    return hit / max(len(q), 1)


def _explain(resp: dict[str, Any]) -> dict[str, Any]:
    return resp.get("explain") or {}


def label_from_oracle_response(
    oracle_mode: str,
    resp: dict[str, Any],
    *,
    query: str = "",
    gold: Any = "",
) -> str:
    """Map a product oracle response into a stage label.

    Representation requires non-episode facts (or atoms). Gold in a chat
    transcript is WRITE_MISS, not a pass. READER_MISS is assigned only by
    probe_failure_stages after structured representation reached the packet.
    """
    if resp.get("error"):
        return "HARNESS_ERROR"
    explain = _explain(resp)
    if explain.get("oracle_unsupported"):
        return "HARNESS_ERROR"
    mode = (oracle_mode or "").lower()
    if mode == "evidence":
        n = int(explain.get("oracle_evidence_count") or 0)
        blob = str(resp.get("context_block") or "")
        # Rows present means source entered Brainy. Gold missing from a
        # truncated evidence dump is not SOURCE_MISS — later stages decide
        # WRITE vs RETRIEVAL vs READER.
        if n > 0 or gold_in_texts(gold, blob):
            return ""
        return "SOURCE_MISS"
    if mode in {"semantic", "representation"}:
        facts = int(explain.get("oracle_fact_count") or 0)
        atoms = int(explain.get("oracle_atom_count") or 0)
        fact_blob = str(explain.get("oracle_fact_blob") or "")
        episode_blob = str(explain.get("oracle_episode_blob") or "")
        if gold:
            if gold_in_texts(gold, fact_blob) or gold_semantically_in_texts(gold, fact_blob):
                return ""
            if gold_in_texts(gold, episode_blob) or (facts == 0 and atoms == 0):
                return "WRITE_MISS"
            if facts > 0 or atoms > 0:
                # Compiled facts exist; lexical gold miss is not yet a write
                # miss — retrieval may still surface a paraphrase.
                return ""
            return "WRITE_MISS"
        if facts > 0 or atoms > 0:
            return ""
        return "WRITE_MISS"
    if mode == "retrieval":
        n = int(explain.get("oracle_memory_count") or 0)
        facts = int(explain.get("oracle_fact_count") or 0)
        if n <= 0:
            return "RETRIEVAL_MISS"
        if facts <= 0 and explain.get("oracle_representation_status") != "empty":
            return "RETRIEVAL_MISS"
        pkt = explain.get("evidence_packet") or {}
        texts = list(pkt.get("contents") or [])
        if not texts and resp.get("context_block"):
            texts = [str(resp.get("context_block"))]
        if gold and texts and not gold_in_texts(gold, texts) and not gold_semantically_in_texts(gold, texts):
            return "RETRIEVAL_MISS"
        if query and texts and _query_token_overlap(query, texts) < 0.2:
            return "RETRIEVAL_MISS"
        return ""
    if mode == "coverage":
        n = int(explain.get("oracle_item_count") or 0)
        cov = resp.get("coverage") or {}
        pkt = explain.get("evidence_packet") or {}
        pkt_cov = pkt.get("coverage") or {}
        hop_proven = pkt_cov.get("hop_join_proven")
        if hop_proven is False:
            return "PROOF_MISS"
        if n <= 0 and not cov.get("satisfied"):
            return "PROOF_MISS"
        texts = list(pkt.get("contents") or [])
        if resp.get("answer"):
            texts.append(str(resp.get("answer")))
        if gold and texts and not gold_in_texts(gold, texts) and not gold_semantically_in_texts(gold, texts) and not cov.get("satisfied"):
            return "PROOF_MISS"
        if query and texts and _query_token_overlap(query, texts) < 0.15:
            return "PROOF_MISS"
        if cov.get("satisfied") is False:
            return "PROOF_MISS"
        return ""
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
    gold: Any = "",
) -> tuple[str, dict[str, Any]]:
    """
    Run evidence → retrieval → representation → coverage probes and return
    (primary_label, flags). Empty primary means no diagnosed miss.

    Retrieval is assessed before representation is blamed: a lexical gold miss
    in the fact dump is not WRITE_MISS when the claim was retrieved, and a
    store hit that never entered the packet is RETRIEVAL_MISS.

    READER_MISS only if structured facts that contain gold (when provided)
    reached the packet and synthesis still failed.
    """
    flags: dict[str, Any] = {"answer_ok": answer_ok, "gold": gold or ""}
    order = ("evidence", "retrieval", "representation", "coverage")
    pending_retrieval = ""
    for mode in order:
        body = oracle_recall_request(tenant_id, subject_id, query, mode)
        resp = post_recall(base_url, body, api_key=api_key)
        label = label_from_oracle_response(mode, resp, query=query, gold=gold)
        explain = _explain(resp)
        gold_facts = gold_in_texts(gold, str(explain.get("oracle_fact_blob") or ""))
        gold_facts_sem = gold_semantically_in_texts(gold, str(explain.get("oracle_fact_blob") or ""))
        gold_eps = gold_in_texts(gold, str(explain.get("oracle_episode_blob") or ""))
        flags[f"oracle_{mode}"] = {
            "label": label,
            "answer_status": resp.get("answer_status"),
            "explain": {
                k: explain.get(k)
                for k in (
                    "oracle_evidence_count",
                    "oracle_atom_count",
                    "oracle_fact_count",
                    "oracle_episode_count",
                    "oracle_memory_count",
                    "oracle_item_count",
                    "oracle_representation_status",
                    "oracle_unsupported",
                )
                if explain.get(k) is not None
            },
        }
        if gold:
            flags[f"oracle_{mode}"]["gold_in_facts"] = gold_facts
            flags[f"oracle_{mode}"]["gold_in_facts_semantic"] = gold_facts_sem
            flags[f"oracle_{mode}"]["gold_in_episodes"] = gold_eps
        if mode == "retrieval" and label == "RETRIEVAL_MISS":
            # Hold retrieval miss until we know whether the claim was written.
            pending_retrieval = "RETRIEVAL_MISS"
            continue
        if label:
            return label, flags
    if pending_retrieval:
        rep = flags.get("oracle_representation") or {}
        if gold and (rep.get("gold_in_facts") or rep.get("gold_in_facts_semantic")):
            return "RETRIEVAL_MISS", flags
        if gold:
            return "WRITE_MISS", flags
        return "RETRIEVAL_MISS", flags
    if not answer_ok:
        rep = flags.get("oracle_representation") or {}
        gold_in_facts = bool(rep.get("gold_in_facts") or rep.get("gold_in_facts_semantic"))
        if gold and not gold_in_facts:
            return "WRITE_MISS", flags
        fact_n = int((rep.get("explain") or {}).get("oracle_fact_count") or 0)
        if fact_n <= 0:
            return "WRITE_MISS", flags
        return "READER_MISS", flags
    return "", flags


__all__ = [
    "FAILURE_LABELS",
    "classify_failure",
    "gold_in_texts",
    "gold_semantically_in_texts",
    "write_failure_record",
    "oracle_recall_request",
    "label_from_oracle_response",
    "post_recall",
    "probe_failure_stages",
]
