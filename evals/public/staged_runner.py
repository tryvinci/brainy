"""Resumable staged evaluation runner.

Stages: preflight → enqueue → drain → verify_store → answer → judge → diagnose → finalize

Identity (API URL, store, dataset, models, protocol) is recorded before work
begins. Resume rejects a mismatched identity. Completed answers are reused when
only judging needs retry. Final JSON is written only after the expected
question set is present and no UNRESOLVED / TRANSPORT_FAIL rows remain.
"""
from __future__ import annotations

import hashlib
import json
import os
import pathlib
import subprocess
import time
from dataclasses import dataclass, field
from typing import Any, Callable

from public.client import BoundAPIClient
from public.protocol import (
    JUDGMENT_TRANSPORT,
    JUDGMENT_UNRESOLVED,
    PROTOCOL_V2,
    STAGES,
    resolve_eval_protocol,
)
from public.schema import (
    CATEGORIES_TO_SCORE,
    EvalItem,
    GenerationData,
    JudgmentData,
    Metadata,
    RetrievalData,
    UnifiedResult,
    compute_metrics,
    utc_now,
)

ROOT = pathlib.Path(__file__).resolve().parents[2]


class StagedRunError(RuntimeError):
    """Qualification / resume cannot continue."""


class IdentityMismatch(StagedRunError):
    """Persisted identity does not match this process."""


@dataclass
class RunSpec:
    run_id: str
    artifact_root: pathlib.Path
    api_url: str
    tenant_prefix: str
    dataset_hash: str
    dataset_url: str = ""
    eval_lane: str = "product-recall"
    protocol: str = PROTOCOL_V2
    top_k: int = 30
    product_sha: str = ""
    harness_sha: str = ""
    answerer_model: str = ""
    judge_model: str = ""
    skip_ingest: bool = False
    qualification_profile: bool = False
    drain_timeout_s: float = 3600.0
    recall_timeout_s: float = 120.0
    max_transport_attempts: int = 3
    subjects: list[str] = field(default_factory=list)


def git_sha(cwd: pathlib.Path | None = None) -> str:
    try:
        return (
            subprocess.check_output(
                ["git", "rev-parse", "HEAD"],
                cwd=str(cwd or ROOT),
                stderr=subprocess.DEVNULL,
            )
            .decode()
            .strip()
        )
    except Exception:
        return ""


def prompt_hashes() -> dict[str, str]:
    from public.judge import JUDGE_SYSTEM_PROMPT

    return {
        "judge_system": hashlib.sha256(JUDGE_SYSTEM_PROMPT.encode("utf-8")).hexdigest(),
        "product_recall_mode": hashlib.sha256(b"mode=answer").hexdigest(),
    }


def store_identity_from_runtime(runtime: dict[str, Any]) -> str:
    ann = runtime.get("ann") or {}
    sigs = runtime.get("signatures") or {}
    payload = {
        "database": runtime.get("database") or "",
        "api_signature": sigs.get("api") or "",
        "worker_signature": sigs.get("worker") or "",
        "ann_active": bool(ann.get("active")),
        "pgvector": bool(ann.get("has_pgvector") or ann.get("HasPGVector")),
        "dim_histogram": ann.get("dim_histogram") or ann.get("DimHistogram") or {},
        "model_histogram": ann.get("model_histogram") or ann.get("ModelHistogram") or {},
    }
    blob = json.dumps(payload, sort_keys=True, default=str)
    return hashlib.sha256(blob.encode("utf-8")).hexdigest()


def default_artifact_root() -> pathlib.Path:
    env = (os.environ.get("BRAINY_EVAL_ARTIFACT_ROOT") or "").strip()
    if env:
        return pathlib.Path(env)
    persistent = pathlib.Path("/opt/cursor/artifacts/eval-runs")
    if persistent.parent.is_dir():
        return persistent
    return ROOT / "docs" / "benchmarks" / "runs"


class RunStore:
    """Incremental JSONL + identity on disk (never /tmp)."""

    def __init__(self, root: pathlib.Path) -> None:
        self.root = pathlib.Path(root)
        self.root.mkdir(parents=True, exist_ok=True)

    def path(self, name: str) -> pathlib.Path:
        return self.root / name

    def write_json(self, name: str, payload: dict[str, Any]) -> None:
        path = self.path(name)
        tmp = path.with_suffix(path.suffix + ".tmp")
        tmp.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
        tmp.replace(path)

    def read_json(self, name: str) -> dict[str, Any] | None:
        path = self.path(name)
        if not path.exists():
            return None
        return json.loads(path.read_text(encoding="utf-8"))

    def append_jsonl(self, name: str, row: dict[str, Any]) -> None:
        path = self.path(name)
        with path.open("a", encoding="utf-8") as handle:
            handle.write(json.dumps(row, default=str) + "\n")
            handle.flush()
            os.fsync(handle.fileno())

    def load_jsonl(self, name: str) -> list[dict[str, Any]]:
        path = self.path(name)
        if not path.exists():
            return []
        rows: list[dict[str, Any]] = []
        for line in path.read_text(encoding="utf-8").splitlines():
            line = line.strip()
            if not line:
                continue
            rows.append(json.loads(line))
        return rows

    def latest_by_id(self, name: str) -> dict[str, dict[str, Any]]:
        out: dict[str, dict[str, Any]] = {}
        for row in self.load_jsonl(name):
            key = str(row.get("id") or "")
            if key:
                out[key] = row
        return out


def compare_identity(persisted: dict[str, Any], current: dict[str, Any]) -> list[str]:
    keys = (
        "protocol",
        "api_url",
        "store_identity",
        "dataset_hash",
        "eval_lane",
        "tenant_prefix",
        "product_sha",
        "embedding_signature",
        "extraction_version",
    )
    gaps: list[str] = []
    for key in keys:
        left = persisted.get(key)
        right = current.get(key)
        if left and right and left != right:
            gaps.append(f"{key}: persisted={left!r} current={right!r}")
    return gaps


@dataclass
class StagedRunner:
    spec: RunSpec
    client: BoundAPIClient
    store: RunStore
    now: Callable[[], str] = utc_now
    sleep: Callable[[float], None] = time.sleep
    runtime_loader: Callable[[], dict[str, Any]] | None = None
    enqueue_fn: Callable[[], dict[str, Any]] | None = None
    job_poll_fn: Callable[[], dict[str, Any]] | None = None
    search_fn: Callable[[str, str, str], list[dict[str, Any]]] | None = None
    answer_fn: Callable[[dict[str, Any]], dict[str, Any]] | None = None
    judge_fn: Callable[[dict[str, Any]], dict[str, Any]] | None = None
    diagnose_fn: Callable[[dict[str, Any]], dict[str, Any]] | None = None

    def run(self, questions: list[dict[str, Any]]) -> dict[str, Any]:
        identity = self.preflight(questions)
        self.enqueue(identity)
        self.drain()
        self.verify_store(identity)
        self.answer(questions, identity)
        self.judge(questions, identity)
        self.diagnose()
        return self.finalize(questions, identity)

    def preflight(self, questions: list[dict[str, Any]]) -> dict[str, Any]:
        self._set_stage("preflight")
        protocol = resolve_eval_protocol(self.spec.protocol)
        runtime = self._runtime()
        fallbacks = runtime.get("fallbacks") or {}
        identity = {
            "protocol": protocol,
            "run_id": self.spec.run_id,
            "api_url": self.client.base_url,
            "product_sha": self.spec.product_sha or git_sha(),
            "harness_sha": self.spec.harness_sha or git_sha(),
            "dataset_hash": self.spec.dataset_hash,
            "dataset_url": self.spec.dataset_url,
            "eval_lane": self.spec.eval_lane,
            "tenant_prefix": self.spec.tenant_prefix,
            "top_k": self.spec.top_k,
            "answerer_model": self.spec.answerer_model,
            "judge_model": self.spec.judge_model,
            "prompt_hashes": prompt_hashes(),
            "store_identity": store_identity_from_runtime(runtime),
            "embedding_signature": ((runtime.get("signatures") or {}).get("api") or ""),
            "extraction_version": (
                ((runtime.get("worker") or {}).get("extractor") or {}).get("version")
                or ((runtime.get("api") or {}).get("extractor") or {}).get("version")
                or ((runtime.get("worker") or {}).get("extractor_signature") or "")
                or ""
            ),
            "database": runtime.get("database") or "",
            "runtime_fallbacks": fallbacks,
            "ann_active": bool((runtime.get("ann") or {}).get("active")),
            "question_ids": [str(q["id"]) for q in questions],
            "subjects": list(self.spec.subjects),
            "recorded_at": self.now(),
        }
        existing = self.store.read_json("identity.json")
        if existing:
            gaps = compare_identity(existing, identity)
            if gaps:
                raise IdentityMismatch("resume rejected: " + "; ".join(gaps))
            return existing
        if self.spec.qualification_profile:
            self._reject_fallbacks(fallbacks, runtime)
        self.store.write_json("identity.json", identity)
        self.store.write_json("questions.json", {"questions": questions})
        return identity

    def enqueue(self, identity: dict[str, Any]) -> dict[str, Any]:
        self._set_stage("enqueue")
        if self.spec.skip_ingest:
            summary = {"skipped": True, "jobs_expected": 0, "job_ids": []}
            self.store.write_json("enqueue.json", summary)
            return summary
        if self.store.read_json("enqueue.json") and self.store.path("jobs.jsonl").exists():
            return self.store.read_json("enqueue.json") or {}
        if self.enqueue_fn is None:
            raise StagedRunError("enqueue_fn is required unless skip_ingest is set")
        summary = self.enqueue_fn()
        expected = int(summary.get("jobs_expected") or 0)
        if expected <= 0:
            raise StagedRunError("enqueue produced no jobs")
        for job_id in summary.get("job_ids") or []:
            self.store.append_jsonl(
                "jobs.jsonl",
                {"id": job_id, "status": "enqueued", "at": self.now()},
            )
        self.store.write_json("enqueue.json", summary)
        return summary

    def drain(self) -> dict[str, Any]:
        self._set_stage("drain")
        if self.spec.skip_ingest:
            summary = {"skipped": True}
            self.store.write_json("drain.json", summary)
            return summary
        if self.job_poll_fn is None:
            raise StagedRunError("job_poll_fn is required to drain extraction")
        deadline = time.time() + float(self.spec.drain_timeout_s)
        last: dict[str, Any] = {}
        while time.time() < deadline:
            last = self.job_poll_fn()
            failed = int(last.get("jobs_failed") or 0)
            if failed:
                raise StagedRunError(f"extraction jobs failed: {last}")
            expected = int(last.get("jobs_expected") or 0)
            completed = int(last.get("jobs_completed") or 0)
            open_n = int(last.get("open") or 0)
            if expected > 0 and completed == expected and open_n == 0 and failed == 0:
                self.store.write_json("drain.json", last)
                return last
            self.sleep(min(2.0, max(0.05, float(last.get("poll_s") or 0.5))))
        raise StagedRunError(
            "drain timed out before expected jobs completed "
            f"(last={last}). empty queue is not sufficient"
        )

    def verify_store(self, identity: dict[str, Any]) -> dict[str, Any]:
        self._set_stage("verify_store")
        runtime = self._runtime()
        fallbacks = runtime.get("fallbacks") or {}
        if self.spec.qualification_profile:
            self._reject_fallbacks(fallbacks, runtime)
        current_store = store_identity_from_runtime(runtime)
        persisted = identity.get("store_identity") or ""
        if persisted and current_store != persisted:
            raise IdentityMismatch(
                f"store identity changed during run: {persisted} -> {current_store}"
            )
        enqueue = self.store.read_json("enqueue.json") or {}
        expected = int(enqueue.get("jobs_expected") or 0)
        poll = {"jobs_expected": expected, "jobs_completed": expected, "jobs_failed": 0, "open": 0}
        if self.job_poll_fn is not None and not self.spec.skip_ingest:
            poll = self.job_poll_fn()
            if int(poll.get("jobs_failed") or 0) != 0:
                raise StagedRunError(f"failed extraction jobs: {poll}")
            if expected and int(poll.get("jobs_completed") or 0) != expected:
                raise StagedRunError(
                    "verify_store: completed jobs != expected "
                    f"(completed={poll.get('jobs_completed')} expected={expected}); "
                    "an empty queue is not sufficient"
                )
            if int(poll.get("open") or 0) != 0:
                raise StagedRunError(f"verify_store: jobs still open {poll}")
        missing_index: list[str] = []
        if self.search_fn is not None:
            for subject in self.spec.subjects:
                hits = self.search_fn(self.spec.tenant_prefix, subject, "conversation")
                if not hits:
                    missing_index.append(subject)
        if missing_index:
            raise StagedRunError(
                "verify_store: subjects not searchable after completed jobs: "
                + ",".join(missing_index)
            )
        report = {
            "ok": True,
            "jobs": poll,
            "runtime_fallbacks": fallbacks,
            "store_identity": current_store,
        }
        self.store.write_json("verify_store.json", report)
        return report

    def answer(self, questions: list[dict[str, Any]], identity: dict[str, Any]) -> None:
        self._set_stage("answer")
        if self.answer_fn is None:
            raise StagedRunError("answer_fn is required")
        done = self.store.latest_by_id("answers.jsonl")
        for question in questions:
            qid = str(question["id"])
            existing = done.get(qid)
            if existing and existing.get("status") not in {JUDGMENT_TRANSPORT, "error", ""}:
                if existing.get("answer") is not None or existing.get("status") == "ok":
                    continue
            row = self.answer_fn(question)
            row = dict(row)
            row.setdefault("id", qid)
            row.setdefault("at", self.now())
            self.store.append_jsonl("answers.jsonl", row)
            done[qid] = row

    def judge(self, questions: list[dict[str, Any]], identity: dict[str, Any]) -> None:
        self._set_stage("judge")
        if self.judge_fn is None:
            raise StagedRunError("judge_fn is required")
        answers = self.store.latest_by_id("answers.jsonl")
        judged = self.store.latest_by_id("judgments.jsonl")
        for question in questions:
            qid = str(question["id"])
            prior = judged.get(qid)
            if prior and prior.get("judgment") in {"CORRECT", "WRONG"}:
                continue
            answer_row = answers.get(qid) or {}
            if answer_row.get("status") == JUDGMENT_TRANSPORT:
                self.store.append_jsonl(
                    "judgments.jsonl",
                    {
                        "id": qid,
                        "judgment": JUDGMENT_TRANSPORT,
                        "score": 0.0,
                        "reason": answer_row.get("reason") or "transport failure",
                        "at": self.now(),
                    },
                )
                continue
            row = self.judge_fn({"question": question, "answer": answer_row})
            row = dict(row)
            row.setdefault("id", qid)
            row.setdefault("at", self.now())
            self.store.append_jsonl("judgments.jsonl", row)

    def diagnose(self) -> None:
        self._set_stage("diagnose")
        answers = self.store.latest_by_id("answers.jsonl")
        judged = self.store.latest_by_id("judgments.jsonl")
        for qid, judgment in judged.items():
            if judgment.get("judgment") == "CORRECT":
                continue
            row = {
                "id": qid,
                "judgment": judgment.get("judgment"),
                "answer_status": (answers.get(qid) or {}).get("status"),
                "primary": "UNRESOLVED",
                "at": self.now(),
            }
            if self.diagnose_fn is not None:
                extra = self.diagnose_fn({"id": qid, "judgment": judgment, "answer": answers.get(qid)})
                row.update(extra or {})
            if not row.get("primary"):
                row["primary"] = "UNRESOLVED"
            self.store.append_jsonl("diagnostics.jsonl", row)

    def finalize(self, questions: list[dict[str, Any]], identity: dict[str, Any]) -> dict[str, Any]:
        self._set_stage("finalize")
        expected_ids = [str(q["id"]) for q in questions]
        answers = self.store.latest_by_id("answers.jsonl")
        judged = self.store.latest_by_id("judgments.jsonl")
        missing = [qid for qid in expected_ids if qid not in judged]
        unresolved = [
            qid
            for qid, row in judged.items()
            if row.get("judgment") in {JUDGMENT_UNRESOLVED, JUDGMENT_TRANSPORT, "JUDGE_MISS"}
        ]
        if missing:
            raise StagedRunError(f"finalize blocked: missing judgments for {missing[:8]}")
        if unresolved:
            raise StagedRunError(
                f"finalize blocked: unresolved/transport judgments {unresolved[:8]}"
            )
        if len(judged) < len(expected_ids):
            raise StagedRunError("finalize blocked: incomplete question set")
        items: list[EvalItem] = []
        questions_by_id = {str(q["id"]): q for q in questions}
        for qid in expected_ids:
            q = questions_by_id[qid]
            ans = answers.get(qid) or {}
            j = judged.get(qid) or {}
            items.append(
                EvalItem(
                    id=qid,
                    group=str(q.get("group") or ""),
                    question=str(q.get("question") or ""),
                    ground_truth=str(q.get("ground_truth") or q.get("answer") or ""),
                    retrieval=RetrievalData(
                        search_query=str(q.get("question") or ""),
                        search_results=list(ans.get("search_results") or []),
                        search_latency_ms=float(ans.get("latency_ms") or 0.0),
                        total_results=int(ans.get("total_results") or 0),
                    ),
                    generation=GenerationData(
                        generated_answer=str(ans.get("answer") or ""),
                        model=str(ans.get("model") or ""),
                    ),
                    judgment=JudgmentData(
                        judgment=str(j.get("judgment") or ""),
                        score=float(j.get("score") or 0.0),
                        reason=str(j.get("reason") or ""),
                        model=str(j.get("model") or ""),
                    ),
                    extras={
                        "category_id": q.get("category_id"),
                        "sample_id": q.get("sample_id"),
                        "eval_protocol": identity.get("protocol"),
                        "answer_status": ans.get("status"),
                    },
                )
            )
        metrics = compute_metrics(items, CATEGORIES_TO_SCORE)
        result = UnifiedResult(
            metadata=Metadata(
                benchmark="locomo-staged",
                project_name="brainy",
                run_id=self.spec.run_id,
                timestamp=self.now(),
                dataset_url=self.spec.dataset_url,
                dataset_sha256=self.spec.dataset_hash,
                brainy_url=self.client.base_url,
                brainy_commit=str(identity.get("product_sha") or ""),
                answerer_model=self.spec.answerer_model,
                judge_model=self.spec.judge_model,
                judge_temperature=0.0,
                top_k=self.spec.top_k,
                config={
                    "eval_lane": self.spec.eval_lane,
                    "eval_protocol": identity.get("protocol"),
                    "store_identity": identity.get("store_identity"),
                    "tenant_prefix": self.spec.tenant_prefix,
                    "stages": list(STAGES),
                },
            ),
            metrics=metrics,
            evaluations=items,
        )
        payload = result.to_dict()
        self.store.write_json("result.json", payload)
        self._set_stage("done")
        return payload

    def _runtime(self) -> dict[str, Any]:
        if self.runtime_loader is not None:
            return self.runtime_loader() or {}
        return self.client.runtime()

    def _set_stage(self, name: str) -> None:
        self.store.write_json("stage.json", {"stage": name, "at": self.now()})

    def _reject_fallbacks(self, fallbacks: dict[str, Any], runtime: dict[str, Any]) -> None:
        for key in ("api_embedder", "api_extractor", "worker_total"):
            try:
                n = int(fallbacks.get(key) or 0)
            except (TypeError, ValueError):
                n = 1
            if n > 0:
                raise StagedRunError(
                    f"qualification forbids provider fallback {key}={n}"
                )
        if not (runtime.get("ann") or {}).get("active"):
            raise StagedRunError("qualification requires active pgvector ANN")


def product_answer_v2(
    client: BoundAPIClient,
    question: dict[str, Any],
    *,
    tenant_id: str,
    subject_id: str,
    top_k: int = 30,
    max_attempts: int = 3,
) -> dict[str, Any]:
    """POST /recall mode=answer. Transport failures stay transport failures."""
    last_err = ""
    started = time.perf_counter()
    for attempt in range(max(1, max_attempts)):
        try:
            body = client.recall(
                tenant_id,
                subject_id,
                str(question.get("question") or ""),
                mode="answer",
                top_k=top_k,
            )
        except Exception as exc:  # noqa: BLE001
            last_err = str(exc)
            continue
        latency_ms = (time.perf_counter() - started) * 1000.0
        if body.get("abstained"):
            return {
                "id": question["id"],
                "status": "abstain",
                "answer": body.get("answer") or "",
                "model": "brainy-recall+answer",
                "latency_ms": latency_ms,
                "raw_keys": sorted(body.keys()),
                "attempt": attempt + 1,
            }
        answer = str(body.get("answer") or "").strip()
        return {
            "id": question["id"],
            "status": "ok" if answer else "empty",
            "answer": answer,
            "model": "brainy-recall+answer",
            "latency_ms": latency_ms,
            "abstained": bool(body.get("abstained")),
            "attempt": attempt + 1,
        }
    return {
        "id": question["id"],
        "status": JUDGMENT_TRANSPORT,
        "answer": "",
        "model": "brainy-recall+transport",
        "reason": last_err,
        "latency_ms": (time.perf_counter() - started) * 1000.0,
        "attempt": max_attempts,
    }
