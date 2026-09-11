#!/usr/bin/env python3
"""Staged LoCoMo runner (eval-protocol-v2).

preflight → enqueue → drain → verify_store → answer → judge → diagnose → finalize

Artifacts go under BRAINY_EVAL_ARTIFACT_ROOT (default /opt/cursor/artifacts/eval-runs),
never /tmp. Resume is by --run-id against that directory.
"""
from __future__ import annotations

import argparse
import os
import pathlib
import sys

ROOT = pathlib.Path(__file__).resolve().parents[3]
EVALS = ROOT / "evals"
if str(EVALS) not in sys.path:
    sys.path.insert(0, str(EVALS))

from public.backends.brainy import BrainyBackend  # noqa: E402
from public.client import BoundAPIClient  # noqa: E402
from public.judge import llm_judge  # noqa: E402
from public.llm import resolve_config  # noqa: E402
from public.locomo.dataset import (  # noqa: E402
    LOCOMO_DATASET_URL,
    ensure_dataset,
    iter_sessions,
    load_conversations,
    scored_question_pool,
)
from public.locomo.run_smoke import ingest_conversation  # noqa: E402
from public.protocol import PROTOCOL_V2  # noqa: E402
from public.staged_runner import (  # noqa: E402
    RunSpec,
    RunStore,
    StagedRunner,
    default_artifact_root,
    git_sha,
    product_answer_v2,
)


def _questions(conversations: list[dict]) -> list[dict]:
    pool = scored_question_pool(conversations)
    out = []
    for row in pool:
        out.append(
            {
                "id": f"{row['sample_id']}-{row['id']}",
                "sample_id": row["sample_id"],
                "question": row["question"],
                "ground_truth": row["answer"],
                "group": row["group"],
                "category_id": int(row.get("category") or 0),
                "subject_id": row["sample_id"],
            }
        )
    return out


def main() -> int:
    parser = argparse.ArgumentParser(description="Staged LoCoMo runner (eval-protocol-v2)")
    parser.add_argument("--base-url", default="", help="Resolved once; required")
    parser.add_argument("--run-id", required=True)
    parser.add_argument("--tenant-prefix", required=True)
    parser.add_argument("--conversations", type=int, default=10)
    parser.add_argument("--skip-ingest", action="store_true")
    parser.add_argument("--qualification-profile", action="store_true")
    parser.add_argument("--async-timeout", type=float, default=3600.0)
    parser.add_argument("--top-k", type=int, default=30)
    parser.add_argument("--judge-model", default="")
    parser.add_argument(
        "--artifact-root",
        default="",
        help="Persistent artifact directory (never /tmp)",
    )
    args = parser.parse_args()

    base = (args.base_url or os.environ.get("BRAINY_BASE_URL") or "").rstrip("/")
    if not base:
        print("base-url required", file=sys.stderr)
        return 2
    os.environ["BRAINY_BASE_URL"] = base
    os.environ["BRAINY_USE_RECALL"] = "1"
    os.environ["BRAINY_EVAL_PROTOCOL"] = PROTOCOL_V2

    artifact_root = pathlib.Path(args.artifact_root or default_artifact_root()) / args.run_id
    if "/tmp/" in str(artifact_root):
        print("refusing to store qualification artifacts under /tmp", file=sys.stderr)
        return 2

    dataset_path, dataset_sha = ensure_dataset()
    conversations = load_conversations(dataset_path)[: args.conversations]
    questions = _questions(conversations)
    subjects = sorted({str(q["subject_id"]) for q in questions})

    client = BoundAPIClient(base, recall_timeout_s=120.0)
    backend = BrainyBackend(
        base,
        tenant_prefix=args.tenant_prefix,
        async_ingest=True,
        async_timeout_s=float(args.async_timeout),
        publish_mode=True,
    )
    llm = resolve_config(model=args.judge_model)

    def enqueue() -> dict:
        job_ids: list[str] = []
        for conv in conversations:
            sample_id = str(conv.get("sample_id") or "")
            ingest_conversation(backend, sample_id, iter_sessions(conv), wait_jobs=False)
            job_ids.extend(backend._pending_jobs.get(sample_id) or [])
        return {"jobs_expected": len(job_ids), "job_ids": job_ids}

    def poll_jobs() -> dict:
        expected_ids = []
        enqueue_doc = (artifact_root / "enqueue.json")
        if enqueue_doc.exists():
            import json

            expected_ids = list((json.loads(enqueue_doc.read_text()) or {}).get("job_ids") or [])
        completed = failed = open_n = 0
        for job_id in expected_ids:
            info = client.job(job_id)
            status = str(info.get("status") or "").lower()
            if status == "completed":
                completed += 1
            elif status == "failed":
                failed += 1
            else:
                open_n += 1
        return {
            "jobs_expected": len(expected_ids),
            "jobs_completed": completed,
            "jobs_failed": failed,
            "open": open_n,
            "poll_s": 2.0,
        }

    def search(tenant_prefix: str, subject: str, query: str) -> list:
        tenant = f"{tenant_prefix}-{subject}"
        body = client.search(tenant, subject, query)
        return list(body.get("results") or [])

    def answer(question: dict) -> dict:
        subject = str(question["subject_id"])
        tenant = backend._tenant(subject)
        return product_answer_v2(
            client,
            question,
            tenant_id=tenant,
            subject_id=subject,
            top_k=args.top_k,
        )

    def judge(payload: dict) -> dict:
        if llm is None:
            raise SystemExit("eval-protocol-v2 requires a pinned judge LLM")
        q = payload["question"]
        ans = payload["answer"]
        result = llm_judge(
            str(ans.get("answer") or ""),
            str(q.get("ground_truth") or ""),
            str(q.get("question") or ""),
            llm,
            protocol=PROTOCOL_V2,
        )
        return {
            "id": q["id"],
            "judgment": result.judgment,
            "score": result.score,
            "reason": result.reason,
            "model": result.model,
        }

    spec = RunSpec(
        run_id=args.run_id,
        artifact_root=artifact_root,
        api_url=base,
        tenant_prefix=args.tenant_prefix,
        dataset_hash=dataset_sha,
        dataset_url=LOCOMO_DATASET_URL,
        protocol=PROTOCOL_V2,
        top_k=args.top_k,
        product_sha=git_sha(),
        harness_sha=git_sha(),
        judge_model=(llm.label if llm else ""),
        skip_ingest=bool(args.skip_ingest),
        qualification_profile=bool(args.qualification_profile),
        drain_timeout_s=float(args.async_timeout),
        subjects=subjects,
    )
    runner = StagedRunner(
        spec,
        client,
        RunStore(artifact_root),
        enqueue_fn=enqueue,
        job_poll_fn=None if args.skip_ingest else poll_jobs,
        search_fn=search,
        answer_fn=answer,
        judge_fn=judge,
    )
    result = runner.run(questions)
    metrics = result.get("metrics") or {}
    print(
        f"accuracy={metrics.get('overall_accuracy')} "
        f"({metrics.get('correct')}/{metrics.get('total')}) "
        f"wrote {artifact_root / 'result.json'}",
        flush=True,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
