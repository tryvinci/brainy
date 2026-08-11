#!/usr/bin/env python3
"""LongMemEval-S runner against Brainy staging (master-plan E2/E3).

Dataset: longmemeval_s_cleaned.json (HuggingFace xiaowu0162/longmemeval-cleaned).
Default path uses /memories/search + harness answerer. Pass --product-recall to
force product POST /recall (fail-closed) for publishable product measurements.
"""
from __future__ import annotations

import argparse
import json
import os
import pathlib
import random
import subprocess
import sys
import uuid
from collections import Counter, defaultdict
from datetime import datetime, timezone

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "evals"))

from public.backends.brainy import BrainyBackend  # noqa: E402
from public.judge import answer_from_memories, llm_judge, lexical_judge  # noqa: E402
from public.llm import resolve_config  # noqa: E402
from public.proveability import RunManifest, require_pins, sha256_file  # noqa: E402
from public.schema import (  # noqa: E402
    EvalItem,
    GenerationData,
    JudgmentData,
    Metadata,
    RetrievalData,
    UnifiedResult,
    compute_metrics,
    utc_now,
)

QUESTION_TYPES = [
    "temporal-reasoning",
    "multi-session",
    "knowledge-update",
    "single-session-user",
    "single-session-assistant",
    "single-session-preference",
]

DEFAULT_DATASET = pathlib.Path("/workspace/datasets/longmemeval/longmemeval_s_cleaned.json")
DATASET_URL = (
    "https://huggingface.co/datasets/xiaowu0162/longmemeval-cleaned/"
    "resolve/main/longmemeval_s_cleaned.json"
)


def git_commit() -> str:
    try:
        return (
            subprocess.check_output(["git", "rev-parse", "--short", "HEAD"], cwd=ROOT)
            .decode()
            .strip()
        )
    except Exception:
        return ""


def load_dataset(path: pathlib.Path) -> list[dict]:
    with path.open(encoding="utf-8") as f:
        data = json.load(f)
    if not isinstance(data, list):
        raise SystemExit(f"expected list, got {type(data)}")
    return data


def stratified_sample(questions: list[dict], n: int, seed: int) -> list[dict]:
    if n <= 0 or n >= len(questions):
        return list(questions)
    by_type: dict[str, list[dict]] = defaultdict(list)
    for q in questions:
        by_type[str(q.get("question_type") or "unknown")].append(q)
    rng = random.Random(seed)
    for rows in by_type.values():
        rng.shuffle(rows)
    total = len(questions)
    allot: dict[str, int] = {}
    remaining = n
    types = sorted(by_type.keys())
    for i, t in enumerate(types):
        if i == len(types) - 1:
            allot[t] = remaining
        else:
            k = max(1, round(n * len(by_type[t]) / total))
            k = min(k, len(by_type[t]), remaining - (len(types) - i - 1))
            allot[t] = k
            remaining -= k
    out: list[dict] = []
    for t in types:
        out.extend(by_type[t][: allot[t]])
    rng.shuffle(out)
    return out[:n]


def _parse_date(s: str | None) -> str | None:
    if not s:
        return None
    s = s.strip()
    for fmt in ("%Y-%m-%d", "%Y/%m/%d", "%B %d, %Y", "%b %d, %Y"):
        try:
            return datetime.strptime(s[:32], fmt).replace(tzinfo=timezone.utc).isoformat()
        except ValueError:
            continue
    return None


def assert_queue_idle_precheck() -> str:
    """Empty-queue precondition for isolated LME publish runs.

    Returns queue_precheck token for the manifest: idle | assumed_idle.
    """
    db = (os.environ.get("BRAINY_DATABASE_URL") or os.environ.get("PGURL") or "").strip()
    if db:
        try:
            out = subprocess.check_output(
                [
                    "psql",
                    db,
                    "-tAc",
                    "SELECT COUNT(*) FROM extraction_jobs "
                    "WHERE status IN ('pending','in_progress')",
                ],
                stderr=subprocess.STDOUT,
                text=True,
                timeout=30,
            ).strip()
            open_n = int(out or "0")
        except Exception as exc:
            raise SystemExit(f"queue precheck failed via psql: {exc}") from exc
        if open_n > 0:
            raise SystemExit(
                f"publish requires empty extraction queue; open jobs={open_n}. "
                "Drain or wait, then re-run isolated LME."
            )
        return "idle"
    if os.environ.get("BRAINY_ASSUME_QUEUE_IDLE", "").lower() in {"1", "true", "yes", "on"}:
        return "assumed_idle"
    raise SystemExit(
        "publish requires empty-queue precheck: set BRAINY_DATABASE_URL (psql) "
        "or BRAINY_ASSUME_QUEUE_IDLE=1 after verifying the worker queue is idle"
    )


def answer_path_for_model(gen_model: str) -> str:
    if (gen_model or "").startswith("brainy-recall"):
        return "/recall"
    return "/memories/search+harness"


def ingest_haystack(backend: BrainyBackend, user_id: str, question: dict, chunk: int = 2) -> int:
    sessions = question.get("haystack_sessions") or []
    dates = question.get("haystack_dates") or []
    session_ids = question.get("haystack_session_ids") or []
    n = 0
    probes: list[str] = []
    for idx, session in enumerate(sessions):
        if not session:
            continue
        sid = session_ids[idx] if idx < len(session_ids) else f"s{idx}"
        date_str = dates[idx] if idx < len(dates) else ""
        meta: dict = {"session_id": sid}
        observed = _parse_date(date_str)
        if observed:
            meta["observed_at"] = observed
        batch: list[dict] = []
        for turn in session:
            role = (turn.get("role") or "user").strip()
            content = (turn.get("content") or "").strip()
            if not content:
                continue
            if role not in ("user", "assistant", "system"):
                role = "user"
            batch.append({"role": role, "content": content})
            # Size-aware flush: BrainyBackend also splits, but keep LME batches small.
            if len(batch) >= chunk:
                backend.remember_messages(user_id, batch, metadata=meta, wait=False)
                n += len(batch)
                batch = []
        if batch:
            backend.remember_messages(user_id, batch, metadata=meta, wait=False)
            n += len(batch)
        if session:
            last = (session[-1].get("content") or "").split()
            if last:
                probes.append(last[0][:40])
    if n and backend.async_ingest:
        try:
            backend.wait_until_jobs_done(user_id)
        except Exception:
            if getattr(backend, "publish_mode", False):
                raise
            backend.wait_until_any_searchable(user_id, probes[:3] or ["conversation"], settle_polls=6)
    return n


def write_report(
    result: UnifiedResult,
    path: pathlib.Path,
    *,
    gaps: list[str] | None = None,
    stage_breakdown: dict[str, dict] | None = None,
) -> None:
    m = result.metrics
    lines = [
        f"# {result.metadata.benchmark} — {result.metadata.run_id}",
        "",
        f"- accuracy: **{m.overall_accuracy:.3f}** ({m.correct}/{m.total})",
        f"- errors: {m.errors}",
        f"- latency p50/p95: {m.latency_p50_ms} / {m.latency_p95_ms} ms",
        "",
        "## By type",
        "",
    ]
    for name, g in sorted(m.by_group.items()):
        lines.append(f"- {name}: {g.correct}/{g.total} ({g.accuracy:.3f})")
    if stage_breakdown:
        lines += ["", "## Stage / answer_path breakdown", ""]
        for name, row in sorted(stage_breakdown.items()):
            lines.append(
                f"- {name}: correct={row.get('correct', 0)}/{row.get('total', 0)} "
                f"answer_path={row.get('answer_paths')}"
            )
    if gaps:
        lines += ["", "## Proveability gaps", ""]
        for g in gaps:
            lines.append(f"- {g}")
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def run(args: argparse.Namespace) -> UnifiedResult:
    product_recall = bool(getattr(args, "product_recall", False))
    if product_recall:
        os.environ["BRAINY_USE_RECALL"] = "1"

    dataset_path = pathlib.Path(args.dataset)
    questions = load_dataset(dataset_path)
    selected = stratified_sample(questions, args.limit, args.seed)
    print(
        f"LongMemEval-S selected {len(selected)}/{len(questions)} "
        f"types={dict(Counter(q.get('question_type') for q in selected))}",
        flush=True,
    )

    base_url = (args.base_url or os.environ.get("BRAINY_BASE_URL") or "http://127.0.0.1:8080").rstrip("/")
    if product_recall:
        os.environ["BRAINY_BASE_URL"] = base_url

    publish = bool(args.publish)
    queue_precheck = ""
    if publish:
        queue_precheck = assert_queue_idle_precheck()
        print(f"queue_precheck={queue_precheck}", flush=True)

    run_id = args.run_id or f"lme-s-{uuid.uuid4().hex[:8]}"
    nonce = uuid.uuid4().hex[:8]
    backend = BrainyBackend(
        base_url,
        tenant_prefix=f"lme-{nonce}",
        async_ingest=not args.sync_ingest,
        async_timeout_s=float(args.async_timeout),
        publish_mode=publish,
    )
    llm = None if args.lexical_only else resolve_config(
        model=args.judge_model or args.answerer_model or "",
        base_url=args.llm_base_url,
        api_key=args.llm_api_key,
    )
    answerer_model = (args.answerer_model or (llm.model if llm else "")) if llm else ""
    judge_model = (args.judge_model or (llm.model if llm else "")) if llm else ""
    if llm:
        print(f"LLM: {llm.label}", flush=True)
    if product_recall:
        print("product-recall: BRAINY_USE_RECALL=1 (fail-closed on non-/recall answers)", flush=True)

    items: list[EvalItem] = []
    answer_paths: list[str] = []
    reader_sources: Counter[str] = Counter()
    jobs_expected = 0
    jobs_completed = 0
    jobs_failed = 0

    for i, q in enumerate(selected):
        qid = str(q.get("question_id") or f"q{i}")
        qtype = str(q.get("question_type") or "unknown")
        user_id = f"{qid}"
        question_text = str(q.get("question") or "")
        gt = str(q.get("answer") or "")
        print(f"[{i+1}/{len(selected)}] {qid} type={qtype} ingest…", flush=True)
        try:
            before_jobs = list(backend._pending_jobs.get(user_id, []))
            n_ing = ingest_haystack(backend, user_id, q)
            after_jobs = list(backend._pending_jobs.get(user_id, []))
            # Tracked ids are cleared after wait; count from ingest side-effect length.
            enqueued = max(0, len(after_jobs) - len(before_jobs))
            # Prefer counting all jobs ever tracked for this subject during ingest.
            # wait_until_jobs_done clears the list; approximate via n_ing batches.
            jobs_expected += max(enqueued, 1 if n_ing and backend.async_ingest else 0)
            print(f"  ingested ~{n_ing} msgs", flush=True)
            results, latency_ms = backend.recall(user_id, question_text, top_k=args.top_k)
            for r in results:
                src = str((r.get("explain") or {}).get("reader_source") or r.get("reader_source") or "")
                if src:
                    reader_sources[src] += 1
            if llm or product_recall:
                answer, gen_model = answer_from_memories(
                    question_text,
                    results,
                    model=answerer_model,
                    config=llm,
                    tenant_id=backend._tenant(user_id),
                    subject_id=user_id,
                    require_product_recall=product_recall,
                )
                path = answer_path_for_model(gen_model)
                answer_paths.append(path)
                if product_recall and path != "/recall":
                    raise RuntimeError(f"product-recall required answer_path=/recall, got {path} model={gen_model}")
                if llm:
                    jr = llm_judge(answer, gt, question_text, llm)
                else:
                    jr = lexical_judge(answer, gt)
            else:
                answer = " | ".join((r.get("content") or "") for r in results[:8])
                gen_model = "lexical"
                answer_paths.append(answer_path_for_model(gen_model))
                jr = lexical_judge(answer, gt)
            # Job outcome sampling via subject status after ingest wait.
            try:
                from httputil import get_json

                counts = get_json(
                    base_url,
                    "/jobs/status",
                    {"tenant_id": backend._tenant(user_id), "subject_id": user_id},
                    timeout=30,
                )
                jobs_completed += int(counts.get("completed") or 0)
                jobs_failed += int(counts.get("failed") or 0)
            except Exception:
                if publish:
                    raise
            items.append(
                EvalItem(
                    id=qid,
                    group=qtype,
                    question=question_text,
                    ground_truth=gt,
                    retrieval=RetrievalData(
                        search_query=question_text,
                        search_results=results,
                        search_latency_ms=latency_ms,
                        total_results=len(results),
                    ),
                    generation=GenerationData(generated_answer=answer, model=gen_model),
                    judgment=JudgmentData(
                        judgment=jr.judgment,
                        score=jr.score,
                        reason=jr.reason,
                        model=jr.model,
                    ),
                    extras={"answer_path": answer_paths[-1]},
                )
            )
            print(f"  {jr.judgment} lat={latency_ms:.0f}ms path={answer_paths[-1]}", flush=True)
        except Exception as exc:
            print(f"  ERROR {exc}", flush=True)
            items.append(
                EvalItem(
                    id=qid,
                    group=qtype,
                    question=question_text,
                    ground_truth=gt,
                    judgment=JudgmentData(
                        judgment="SKIPPED", score=0.0, reason=str(exc)[:300], model="error"
                    ),
                    extras={"error": str(exc)[:500], "answer_path": "error"},
                )
            )
            answer_paths.append("error")
            if publish and product_recall:
                # Fail-closed: do not continue publishing a partial product-recall run.
                raise SystemExit(f"publish product-recall failed on {qid}: {exc}") from exc

    # Score all LME types (no category_id filter).
    metrics = compute_metrics(items, score_groups=set())
    stage_breakdown: dict[str, dict] = {}
    for item in items:
        row = stage_breakdown.setdefault(
            item.group, {"total": 0, "correct": 0, "answer_paths": Counter()}
        )
        row["total"] += 1
        if item.judgment and item.judgment.judgment == "CORRECT":
            row["correct"] += 1
        path = str((item.extras or {}).get("answer_path") or "unknown")
        row["answer_paths"][path] += 1
    stage_breakdown_out = {
        k: {
            "total": v["total"],
            "correct": v["correct"],
            "answer_paths": dict(v["answer_paths"]),
        }
        for k, v in stage_breakdown.items()
    }

    dominant_path = ""
    if answer_paths:
        dominant_path = Counter(answer_paths).most_common(1)[0][0]

    result = UnifiedResult(
        metadata=Metadata(
            benchmark="longmemeval-s",
            project_name="brainy",
            run_id=run_id,
            timestamp=utc_now(),
            dataset_url=DATASET_URL,
            dataset_sha256=sha256_file(dataset_path) if dataset_path.exists() else "",
            brainy_url=base_url,
            brainy_commit=git_commit(),
            answerer_model=(
                "brainy-recall" if product_recall else (llm.label if llm else "lexical")
            ),
            judge_model=(llm.label if llm else "lexical-overlap-v0"),
            judge_temperature=0.0,
            top_k=args.top_k,
            config={
                "limit": args.limit,
                "seed": args.seed,
                "async_ingest": not args.sync_ingest,
                "dataset": str(dataset_path),
                "types": dict(Counter(i.group for i in items)),
                "product_recall": product_recall,
                "publish": publish,
            },
        ),
        metrics=metrics,
        evaluations=items,
    )

    extras: dict = {
        "ingest_mode": "async" if not args.sync_ingest else "sync",
        "product_recall": product_recall,
        "answer_path": "/recall" if product_recall and dominant_path == "/recall" else dominant_path,
        "answer_path_counts": dict(Counter(answer_paths)),
        "jobs_expected": jobs_expected,
        "jobs_completed": jobs_completed,
        "jobs_failed": jobs_failed,
        "reader_source": dict(reader_sources),
        "stage_breakdown": stage_breakdown_out,
        "queue_precheck": queue_precheck or "not_required",
    }
    if product_recall:
        extras["product_recall"] = True

    manifest = RunManifest(
        benchmark="longmemeval-s",
        dataset_url=DATASET_URL,
        dataset_sha256=result.metadata.dataset_sha256,
        dataset_path=str(dataset_path),
        brainy_url=base_url,
        brainy_commit=result.metadata.brainy_commit,
        answerer_model=result.metadata.answerer_model,
        judge_model=result.metadata.judge_model,
        judge_temperature=0.0,
        top_k=args.top_k,
        question_limit=args.limit,
        notes="LongMemEval-S; product-recall requires answer_path=/recall",
        extras=extras,
    )
    gaps = require_pins(manifest)
    if publish and gaps:
        raise SystemExit("proveability gaps: " + "; ".join(gaps))

    out_dir = pathlib.Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    (out_dir / f"{run_id}.json").write_text(json.dumps(result.to_dict(), indent=2) + "\n")
    manifest.write(out_dir / f"{run_id}.manifest.json")
    report_path = pathlib.Path(args.report) if getattr(args, "report", "") else out_dir / f"{run_id}.md"
    write_report(result, report_path, gaps=gaps, stage_breakdown=stage_breakdown_out)
    print(
        f"DONE accuracy={metrics.overall_accuracy:.3f} "
        f"correct={metrics.correct}/{metrics.total} "
        f"answer_path={extras.get('answer_path')}",
        flush=True,
    )
    if gaps:
        print("proveability gaps:", "; ".join(gaps), flush=True)
    return result


def main() -> int:
    p = argparse.ArgumentParser(description="LongMemEval-S vs Brainy")
    p.add_argument("--dataset", default=str(DEFAULT_DATASET))
    p.add_argument("--base-url", default="")
    p.add_argument("--limit", type=int, default=100, help="0 = all 500")
    p.add_argument("--seed", type=int, default=1)
    p.add_argument("--top-k", type=int, default=30)
    p.add_argument("--sync-ingest", action="store_true")
    p.add_argument("--publish", action="store_true", help="Fail closed on job barrier / incomplete runs")
    p.add_argument(
        "--product-recall",
        action="store_true",
        help="Force BRAINY_USE_RECALL=1; fail items that do not answer via POST /recall",
    )
    p.add_argument("--async-timeout", type=float, default=1800.0)
    p.add_argument("--lexical-only", action="store_true")
    p.add_argument("--answerer-model", default="")
    p.add_argument("--judge-model", default="")
    p.add_argument("--llm-base-url", default="")
    p.add_argument("--llm-api-key", default="")
    p.add_argument("--run-id", default="")
    p.add_argument("--report", default="", help="Optional markdown report path")
    p.add_argument("--out-dir", default=str(ROOT / "docs" / "benchmarks" / "runs"))
    args = p.parse_args()
    run(args)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
