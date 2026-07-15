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
    return gaps
