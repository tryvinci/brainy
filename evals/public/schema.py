from __future__ import annotations

from dataclasses import asdict, dataclass, field
from datetime import datetime, timezone
from typing import Any


@dataclass
class RetrievalData:
    search_query: str = ""
    search_results: list[dict[str, Any]] = field(default_factory=list)
    search_latency_ms: float = 0.0
    total_results: int = 0


@dataclass
class GenerationData:
    generated_answer: str = ""
    model: str = ""
    prompt_tokens: int | None = None
    completion_tokens: int | None = None


@dataclass
class JudgmentData:
    judgment: str = ""  # CORRECT / WRONG / SKIPPED
    score: float = 0.0
    reason: str = ""
    model: str = ""


@dataclass
class EvalItem:
    id: str
    group: str = ""
    question: str = ""
    ground_truth: str = ""
    retrieval: RetrievalData | None = None
    generation: GenerationData | None = None
    judgment: JudgmentData | None = None
    extras: dict[str, Any] = field(default_factory=dict)


@dataclass
class GroupMetrics:
    group_name: str
    total: int = 0
    correct: int = 0
    accuracy: float = 0.0


@dataclass
class Metrics:
    overall_accuracy: float = 0.0
    total: int = 0
    correct: int = 0
    errors: int = 0
    by_group: dict[str, GroupMetrics] = field(default_factory=dict)
    latency_p50_ms: float | None = None
    latency_p95_ms: float | None = None


@dataclass
class Metadata:
    benchmark: str = ""
    project_name: str = ""
    run_id: str = ""
    timestamp: str = ""
    schema_compatible: str = "mem0ai/memory-benchmarks@UnifiedResult-1.0"
    dataset_url: str = ""
    dataset_sha256: str = ""
    brainy_url: str = ""
    brainy_commit: str = ""
    answerer_model: str = ""
    judge_model: str = ""
    judge_temperature: float = 0.0
    top_k: int = 10
    config: dict[str, Any] = field(default_factory=dict)


@dataclass
class UnifiedResult:
    schema_version: str = "1.0"
    metadata: Metadata = field(default_factory=Metadata)
    metrics: Metrics = field(default_factory=Metrics)
    evaluations: list[EvalItem] = field(default_factory=list)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


CATEGORY_NAMES = {
    1: "multi-hop",
    2: "temporal",
    3: "open-domain",
    4: "single-hop",
    5: "adversarial",
}

# Industry default: score categories 1-4 (exclude adversarial from overall).
CATEGORIES_TO_SCORE = {1, 2, 3, 4}


def utc_now() -> str:
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def compute_metrics(items: list[EvalItem], score_groups: set[int] | None = None) -> Metrics:
    score_groups = score_groups or CATEGORIES_TO_SCORE
    by_group: dict[str, GroupMetrics] = {}
    latencies: list[float] = []
    total = correct = errors = 0

    for item in items:
        cat_id = item.extras.get("category_id")
        if isinstance(cat_id, int) and cat_id not in score_groups:
            continue
        group = item.group or "unknown"
        bucket = by_group.setdefault(group, GroupMetrics(group_name=group))
        bucket.total += 1
        total += 1
        if item.judgment is None:
            errors += 1
            continue
        if item.judgment.judgment == "CORRECT":
            correct += 1
            bucket.correct += 1
        elif item.judgment.judgment not in {"WRONG", "CORRECT"}:
            errors += 1
        if item.retrieval:
            latencies.append(item.retrieval.search_latency_ms)

    for bucket in by_group.values():
        bucket.accuracy = (bucket.correct / bucket.total) if bucket.total else 0.0

    latency_p50 = latency_p95 = None
    if latencies:
        ordered = sorted(latencies)
        latency_p50 = _percentile(ordered, 50)
        latency_p95 = _percentile(ordered, 95)

    return Metrics(
        overall_accuracy=(correct / total) if total else 0.0,
        total=total,
        correct=correct,
        errors=errors,
        by_group=by_group,
        latency_p50_ms=latency_p50,
        latency_p95_ms=latency_p95,
    )


def _percentile(ordered: list[float], pct: float) -> float:
    if not ordered:
        return 0.0
    if len(ordered) == 1:
        return ordered[0]
    rank = (pct / 100.0) * (len(ordered) - 1)
    low = int(rank)
    high = min(low + 1, len(ordered) - 1)
    weight = rank - low
    return ordered[low] * (1 - weight) + ordered[high] * weight
