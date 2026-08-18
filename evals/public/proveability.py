from __future__ import annotations

import hashlib
import json
import pathlib
from dataclasses import asdict, dataclass, field
from typing import Any


@dataclass
class RunManifest:
    """Everything required to reproduce a public-bench run."""

    framework: str = "brainy/evals/public"
    framework_version: str = "0.1.0"
    benchmark: str = ""
    dataset_url: str = ""
    dataset_sha256: str = ""
    dataset_path: str = ""
    brainy_url: str = ""
    brainy_commit: str = ""
    answerer_model: str = ""
    judge_model: str = ""
    judge_temperature: float = 0.0
    top_k: int = 10
    conversation_limit: int | None = None
    question_limit: int | None = None
    scored_categories: list[int] = field(default_factory=lambda: [1, 2, 3, 4])
    notes: str = ""
    extras: dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)

    def write(self, path: pathlib.Path) -> None:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps(self.to_dict(), indent=2) + "\n", encoding="utf-8")


PRODUCT_RECALL_LANE = "product-recall"
INDUSTRY_SEARCH_LANE = "industry-search"


def resolve_eval_lane(eval_lane: str = "", use_recall_env: str = "") -> str:
    """Label the two freeze paths. Do not mix them in one score.

    product-recall: POST /recall (fail-closed when the lane is selected).
    industry-search: search → shared answerer → shared judge (Mem0-style; top-k 200).
    """
    raw = (eval_lane or "").strip().lower()
    if raw in {PRODUCT_RECALL_LANE, "product", "recall", "/recall"}:
        return PRODUCT_RECALL_LANE
    if raw in {INDUSTRY_SEARCH_LANE, "industry", "search", "search-harness"}:
        return INDUSTRY_SEARCH_LANE
    if (use_recall_env or "").strip().lower() in {"1", "true", "yes", "on"}:
        return PRODUCT_RECALL_LANE
    return INDUSTRY_SEARCH_LANE


def default_lane_top_k(lane: str, *, explicit: int | None, lane_flag_set: bool) -> int:
    if explicit is not None and explicit > 0:
        return explicit
    if lane == INDUSTRY_SEARCH_LANE and lane_flag_set:
        return 200
    return 30


def lane_answer_path(lane: str) -> str:
    if lane == PRODUCT_RECALL_LANE:
        return "/recall"
    return "search+harness"


def sha256_file(path: pathlib.Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def require_pins(manifest: RunManifest) -> list[str]:
    """Return proveability gaps (empty list => publishable)."""
    gaps: list[str] = []
    if not manifest.dataset_url:
        gaps.append("dataset_url missing")
    if not manifest.dataset_sha256:
        gaps.append("dataset_sha256 missing — download via framework before claiming results")
    if not manifest.brainy_url and not manifest.brainy_commit:
        gaps.append("brainy_url or brainy_commit required")
    if not manifest.judge_model:
        gaps.append("judge_model missing")
    if manifest.judge_temperature != 0.0:
        gaps.append("judge_temperature must be 0.0 for proveable public claims")
    extras = manifest.extras or {}
    if extras.get("product_recall"):
        if extras.get("answer_path") != "/recall":
            gaps.append("answer_path must be /recall for product-recall publish")
        if extras.get("ingest_mode") != "async":
            gaps.append("ingest_mode must be async for product-recall publish")
        for key in ("jobs_expected", "jobs_completed", "jobs_failed", "reader_source"):
            if key not in extras:
                gaps.append(f"extras.{key} missing for product-recall publish")
        try:
            expected = int(extras.get("jobs_expected") or 0)
            completed = int(extras.get("jobs_completed") or 0)
            failed = int(extras.get("jobs_failed") or 0)
        except (TypeError, ValueError):
            gaps.append("extras jobs_* must be integers for product-recall publish")
        else:
            if expected <= 0:
                gaps.append("extras.jobs_expected must be > 0 for product-recall publish")
            if failed != 0:
                gaps.append(f"extras.jobs_failed must be 0 (got {failed})")
            if completed != expected:
                gaps.append(
                    f"extras.jobs_completed ({completed}) must equal jobs_expected ({expected})"
                )
        if extras.get("queue_precheck") not in {"idle", "assumed_idle"}:
            gaps.append("queue_precheck must be idle or assumed_idle for product-recall publish")
    return gaps
