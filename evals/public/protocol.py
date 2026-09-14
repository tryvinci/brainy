"""Versioned evaluation protocol.

v1 (legacy smoke / in-flight n=1540): evaluator may pick recall mode,
HTTP failures become \"not in memory\", malformed judge JSON can fall back
to substring CORRECT.

v2 (qualification): product scoring is the same mode:\"answer\" request a
reference client sends; transport failures stay transport failures;
unresolved judgments block finalization.
"""
from __future__ import annotations

PROTOCOL_V1 = "eval-protocol-v1"
PROTOCOL_V2 = "eval-protocol-v2"
CURRENT_PROTOCOL = PROTOCOL_V2

JUDGMENT_UNRESOLVED = "UNRESOLVED"
JUDGMENT_TRANSPORT = "TRANSPORT_FAIL"

STAGES = (
    "preflight",
    "enqueue",
    "drain",
    "verify_store",
    "answer",
    "judge",
    "diagnose",
    "finalize",
)


def resolve_eval_protocol(raw: str = "") -> str:
    text = (raw or "").strip().lower()
    if text in {PROTOCOL_V1, "v1", "legacy"}:
        return PROTOCOL_V1
    if text in {PROTOCOL_V2, "v2", "qualification", ""}:
        return PROTOCOL_V2
    if "v1" in text:
        return PROTOCOL_V1
    return PROTOCOL_V2


def is_v2(protocol: str) -> bool:
    return resolve_eval_protocol(protocol) == PROTOCOL_V2
