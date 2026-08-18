"""LOCOMO smoke runner — proveable public-bench layer L3.

Usage:
  cd evals && python -m public.locomo.run_smoke \\
    --base-url https://brainy-api-staging.onrender.com \\
    --conversations 1 --questions 30

Env (Brainy):
  BRAINY_BASE_URL, BRAINY_API_KEY (optional)

Env (answerer + judge — any OpenAI-compatible host):
  LLM_BASE_URL   e.g. http://127.0.0.1:11434/v1  (Ollama)
  LLM_API_KEY    optional; falls back to OPENAI_API_KEY; local defaults to \"ollama\"
  LLM_MODEL      e.g. llama3.1 / qwen2.5 / mistral
  OPENAI_API_KEY still works for api.openai.com

Without an LLM endpoint the runner uses retrieval-concat + lexical judge
(labeled as such — not a publishable LOCOMO J-score). Open-weight models are
fine for publishable runs if the model id + base URL are pinned in the manifest.
"""
from __future__ import annotations

import argparse
import json
import os
import pathlib
import subprocess
import sys
import uuid

ROOT = pathlib.Path(__file__).resolve().parents[3]
EVALS = ROOT / "evals"
if str(EVALS) not in sys.path:
    sys.path.insert(0, str(EVALS))
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from public.backends.brainy import BrainyBackend, _probe_token  # noqa: E402
from public.judge import answer_from_memories, lexical_judge, llm_judge  # noqa: E402
from public.llm import resolve_config  # noqa: E402
from public.locomo.dataset import (  # noqa: E402
    LOCOMO_DATASET_URL,
    LOCOMO_PAPER,
    LOCOMO_UPSTREAM,
    ensure_dataset,
    iter_questions,
    iter_sessions,
    load_conversations,
)
from public.proveability import (  # noqa: E402
    RunManifest,
    default_lane_top_k,
    lane_answer_path,
    require_pins,
    resolve_eval_lane,
)
from public.stage_oracle import probe_failure_stages, write_failure_record  # noqa: E402
from public.schema import (  # noqa: E402
    CATEGORY_NAMES,
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


def git_commit() -> str:
    try:
        return (
            subprocess.check_output(
                ["git", "rev-parse", "HEAD"],
                cwd=ROOT,
                stderr=subprocess.DEVNULL,
            )
            .decode()
            .strip()
        )
    except Exception:
        return ""


def ingest_conversation(
    backend: BrainyBackend,
    user_id: str,
    sessions: list[dict],
    chunk: int = 8,
) -> int:
    """Ingest dialogue as atomic turns (one API message each).

    Batching multiple *messages* in one ingest call is fine — product extract
    atomizes per message. Pass session_id / observed_at as client metadata
    (same pattern product docs recommend for chat apps).

    Default public smoke uses ``/ingest/async`` so the worker provider extract
    path matches production clients. Per-batch wait is skipped during enqueue;
    a final readiness poll runs after all sessions are submitted.
    """
    remembered = 0
    probes: list[str] = []
    for session in sessions:
        meta: dict = {"session_id": session.get("session_id") or ""}
        observed = (session.get("observed_at") or "").strip()
        if observed:
            meta["observed_at"] = observed
            year = _probe_token([{"content": observed}])
            if year:
                probes.append(year)
        batch: list[dict] = []
        # Preserve speaker identity as stable user/assistant roles (not all-user).
        # Keep "Speaker: text" content so extract can attribute facts by name.
        speaker_roles: dict[str, str] = {}
        for turn in session.get("turns") or []:
            if not turn:
                continue
            speaker = turn[0]
            text = turn[1] if len(turn) > 1 else ""
            urls = list(turn[2]) if len(turn) > 2 and turn[2] else []
            sp = (speaker or "user").strip() or "user"
            if sp not in speaker_roles:
                speaker_roles[sp] = "user" if len(speaker_roles) % 2 == 0 else "assistant"
            msg = {"role": speaker_roles[sp], "content": f"{sp}: {text}"}
            if urls:
                msg["image_urls"] = urls
            batch.append(msg)
            if len(batch) >= chunk:
                backend.remember_messages(user_id, batch, metadata=meta, wait=False)
                probe = _probe_token(batch)
                if probe:
                    probes.append(probe)
                remembered += len(batch)
                batch = []
        if batch:
            backend.remember_messages(user_id, batch, metadata=meta, wait=False)
            probe = _probe_token(batch)
            if probe:
                probes.append(probe)
            remembered += len(batch)
    if backend.async_ingest and remembered:
        # Prefer job-completion barrier; fall back to capped search settle
        # only outside publish mode.
        try:
            backend.wait_until_jobs_done(user_id)
        except Exception:
            if getattr(backend, "publish_mode", False):
                raise
            uniq: list[str] = []
            for p in probes:
                if p and p not in uniq:
                    uniq.append(p)
                if len(uniq) >= 3:
                    break
            backend.wait_until_any_searchable(user_id, uniq or ["conversation"])
    return remembered


def run(args: argparse.Namespace) -> UnifiedResult:
    dataset_path, dataset_sha = ensure_dataset()
    conversations = load_conversations(dataset_path)
    if args.conversations:
        conversations = conversations[: args.conversations]

    base_url = (args.base_url or os.environ.get("BRAINY_BASE_URL") or "http://127.0.0.1:8080").rstrip("/")
    run_id = args.run_id or f"locomo-smoke-{uuid.uuid4().hex[:8]}"
    nonce = uuid.uuid4().hex[:8]
    async_ingest = not bool(getattr(args, "sync_ingest", False))
    system = (getattr(args, "system", None) or "brainy").strip().lower()
    lane_flag = str(getattr(args, "eval_lane", "") or "").strip()
    eval_lane = resolve_eval_lane(lane_flag, os.environ.get("BRAINY_USE_RECALL", ""))
    args.top_k = default_lane_top_k(
        eval_lane,
        explicit=getattr(args, "top_k", None),
        lane_flag_set=bool(lane_flag),
    )
    answer_path = lane_answer_path(eval_lane)
    if system != "mem0" and eval_lane == "product-recall":
        os.environ["BRAINY_USE_RECALL"] = "1"
        os.environ.setdefault("BRAINY_BASE_URL", base_url)
    elif lane_flag and eval_lane == "industry-search":
        os.environ.pop("BRAINY_USE_RECALL", None)
    if system == "mem0":
        from public.backends.mem0 import Mem0Backend  # local import — optional dependency path

        backend = Mem0Backend(async_timeout_s=float(getattr(args, "async_timeout", 180.0)))
        base_url = "https://api.mem0.ai"
        print("system=mem0 (Platform API; same-pin compare GAP-M1)", flush=True)
    else:
        backend = BrainyBackend(
            base_url,
            tenant_prefix=f"locomo-{nonce}",
            async_ingest=async_ingest,
            async_timeout_s=float(getattr(args, "async_timeout", 180.0)),
        )
        print(
            f"system=brainy ingest_mode={'async' if async_ingest else 'sync'} "
            f"eval_lane={eval_lane} answer_path={answer_path} top_k={args.top_k} "
            f"(provider extract only on async worker path)",
            flush=True,
        )

    llm = None if args.lexical_only else resolve_config(
        model=args.judge_model or args.answerer_model or "",
        base_url=args.llm_base_url,
        api_key=args.llm_api_key,
    )
    use_llm = llm is not None
    answerer_model = (args.answerer_model or (llm.model if llm else "")) if use_llm else ""
    judge_model = (args.judge_model or (llm.model if llm else "")) if use_llm else "lexical-overlap-v0"
    answerer_cfg = (
        resolve_config(model=answerer_model, base_url=args.llm_base_url, api_key=args.llm_api_key)
        if use_llm
        else None
    )
    judge_cfg = (
        resolve_config(model=judge_model, base_url=args.llm_base_url, api_key=args.llm_api_key)
        if use_llm
        else None
    )

    if use_llm:
        print(f"LLM answerer/judge: {judge_cfg.label if judge_cfg else llm.label}", flush=True)
    else:
        print("LLM unset — lexical judge + retrieval-concat (not publishable J-score)", flush=True)

    items: list[EvalItem] = []
    # When evaluating multiple conversations, distribute the question budget
    # across them. Otherwise a large --questions value is exhausted by conv[0]
    # and later conversations never run (breaks L4-style multi-convo runs).
    n_conv = max(1, len(conversations))
    if args.questions is None or args.questions <= 0:
        per_conv_budget = None
        global_budget = None
    elif n_conv > 1:
        per_conv_budget = max(1, args.questions // n_conv)
        global_budget = args.questions
    else:
        per_conv_budget = args.questions
        global_budget = args.questions

    for conv_idx, conversation in enumerate(conversations):
        sample_id = str(conversation.get("sample_id") or f"c{conv_idx}")
        user_id = f"{sample_id}"
        sessions = iter_sessions(conversation)
        n_turns = sum(len(s.get("turns") or []) for s in sessions)
        n_ingested = ingest_conversation(backend, user_id, sessions)
        print(f"[{sample_id}] ingested {n_ingested} turns ({n_turns} total, {len(sessions)} sessions)", flush=True)

        questions = iter_questions(conversation)
        used_this_conv = 0
        for qa in questions:
            if global_budget is not None and len(items) >= global_budget:
                break
            if per_conv_budget is not None and used_this_conv >= per_conv_budget:
                break
            cat_id = qa.get("category")
            try:
                cat_id_int = int(cat_id) if cat_id is not None else 0
            except (TypeError, ValueError):
                cat_id_int = 0
            group = CATEGORY_NAMES.get(cat_id_int, f"cat-{cat_id}")

            results, latency_ms = backend.recall(user_id, qa["question"], top_k=args.top_k)
            tenant_for_recall = ""
            if system == "brainy" and hasattr(backend, "_tenant"):
                tenant_for_recall = backend._tenant(user_id)
            answer, gen_model = answer_from_memories(
                qa["question"],
                results,
                model=answerer_model,
                config=answerer_cfg,
                tenant_id=tenant_for_recall,
                subject_id=user_id,
                require_product_recall=(system == "brainy" and eval_lane == "product-recall"),
            )
            if use_llm and judge_cfg is not None:
                judged = llm_judge(answer, qa["answer"], qa["question"], judge_cfg)
            else:
                judged = lexical_judge(answer, qa["answer"])

            items.append(
                EvalItem(
                    id=f"{sample_id}-{qa['id']}",
                    group=group,
                    question=qa["question"],
                    ground_truth=qa["answer"],
                    retrieval=RetrievalData(
                        search_query=qa["question"],
                        search_results=results,
                        search_latency_ms=latency_ms,
                        total_results=len(results),
                    ),
                    generation=GenerationData(generated_answer=answer, model=gen_model),
                    judgment=JudgmentData(
                        judgment=judged.judgment,
                        score=judged.score,
                        reason=judged.reason,
                        model=judged.model,
                    ),
                    extras={
                        "category_id": cat_id_int,
                        "sample_id": sample_id,
                        "evidence": qa.get("evidence") or [],
                    },
                )
            )
            if (
                system == "brainy"
                and judged.judgment != "CORRECT"
                and getattr(args, "failure_ledger", None)
            ):
                if judged.judgment == "JUDGE_MISS":
                    write_failure_record(
                        args.failure_ledger,
                        dataset="locomo-smoke",
                        question_id=f"{sample_id}-{qa['id']}",
                        question=qa["question"],
                        primary="JUDGE_MISS",
                        flags={
                            "group": group,
                            "category_id": cat_id_int,
                            "judgment": judged.judgment,
                            "ground_truth": qa["answer"],
                            "generated_answer": answer,
                        },
                        notes=judged.reason or "",
                    )
                else:
                    tenant_id = backend._tenant(user_id)
                    primary, flags = probe_failure_stages(
                        base_url,
                        tenant_id=tenant_id,
                        subject_id=user_id,
                        query=qa["question"],
                        answer_ok=False,
                        api_key=os.environ.get("BRAINY_API_KEY")
                        or (os.environ.get("BRAINY_API_KEYS") or "").split(",")[0].strip(),
                        gold=qa.get("answer") or "",
                    )
                    write_failure_record(
                        args.failure_ledger,
                        dataset="locomo-smoke",
                        question_id=f"{sample_id}-{qa['id']}",
                        question=qa["question"],
                        primary=primary or "WRITE_MISS",
                        flags={
                            **flags,
                            "group": group,
                            "category_id": cat_id_int,
                            "judgment": judged.judgment,
                            "ground_truth": qa["answer"],
                            "generated_answer": answer,
                        },
                        notes=judged.reason or "",
                    )
            used_this_conv += 1
            if len(items) % 10 == 0 or judged.judgment == "CORRECT":
                print(
                    f"  [{sample_id}] q={len(items)} {judged.judgment} "
                    f"lat={latency_ms:.0f}ms group={group}",
                    flush=True,
                )
        if global_budget is not None and len(items) >= global_budget:
            break

    metrics = compute_metrics(items, CATEGORIES_TO_SCORE)
    commit = git_commit()
    answerer_pin = (answerer_cfg.label if answerer_cfg else "") or "retrieval-concat-v0"
    judge_pin = (judge_cfg.label if judge_cfg else "") or "lexical-overlap-v0"
    result = UnifiedResult(
        metadata=Metadata(
            benchmark="locomo-smoke",
            project_name=system,
            run_id=run_id,
            timestamp=utc_now(),
            dataset_url=LOCOMO_DATASET_URL,
            dataset_sha256=dataset_sha,
            brainy_url=base_url,
            brainy_commit=commit if system == "brainy" else "",
            answerer_model=answerer_pin,
            judge_model=judge_pin,
            judge_temperature=0.0,
            top_k=args.top_k,
            config={
                "conversations": args.conversations,
                "questions": args.questions,
                "lexical_only": not use_llm,
                "llm_base_url": (llm.base_url if llm else ""),
                "ingest_mode": "async" if async_ingest else "sync",
                "locomo_upstream": LOCOMO_UPSTREAM,
                "locomo_paper": LOCOMO_PAPER,
                "eval_lane": eval_lane,
                "answer_path": answer_path,
            },
        ),
        metrics=metrics,
        evaluations=items,
    )

    manifest = RunManifest(
        benchmark="locomo-smoke",
        dataset_url=LOCOMO_DATASET_URL,
        dataset_sha256=dataset_sha,
        dataset_path=str(dataset_path),
        brainy_url=base_url,
        brainy_commit=commit,
        answerer_model=answerer_pin,
        judge_model=judge_pin,
        judge_temperature=0.0,
        top_k=args.top_k,
        conversation_limit=args.conversations,
        question_limit=args.questions,
        notes="Smoke run — not full LOCOMO. lexical judge is not publishable J-score. Dual-path: product-recall vs industry-search must not be mixed.",
        extras={
            "llm_base_url": (llm.base_url if llm else ""),
            "ingest_mode": "async" if async_ingest else "sync",
            "eval_lane": eval_lane,
            "answer_path": answer_path,
        },
    )
    gaps = require_pins(manifest)
    out_dir = pathlib.Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    result_path = out_dir / f"{run_id}.json"
    manifest_path = out_dir / f"{run_id}.manifest.json"
    result_path.write_text(json.dumps(result.to_dict(), indent=2) + "\n", encoding="utf-8")
    manifest.write(manifest_path)

    md_path = pathlib.Path(args.report) if args.report else out_dir / f"{run_id}.md"
    md_path.write_text(_render_report(result, gaps), encoding="utf-8")

    print(f"accuracy={metrics.overall_accuracy:.3f} ({metrics.correct}/{metrics.total})")
    print(f"wrote {result_path}")
    print(f"wrote {manifest_path}")
    print(f"wrote {md_path}")
    if gaps:
        print("proveability gaps:", "; ".join(gaps))
    if not use_llm:
        print("NOTE: lexical judge — do not cite as LOCOMO J-score in public posts")
    return result


def _render_report(result: UnifiedResult, gaps: list[str]) -> str:
    m = result.metrics
    meta = result.metadata
    lines = [
        f"# LOCOMO smoke — `{meta.run_id}`",
        "",
        f"**Timestamp:** {meta.timestamp}  ",
        f"**Brainy:** `{meta.brainy_url}` (commit `{meta.brainy_commit or 'unknown'}`)  ",
        f"**Dataset:** [{meta.dataset_url}]({meta.dataset_url})  ",
        f"**SHA256:** `{meta.dataset_sha256}`  ",
        f"**Answerer:** `{meta.answerer_model}`  ",
        f"**Judge:** `{meta.judge_model}` (temp={meta.judge_temperature})  ",
        f"**Schema:** {meta.schema_compatible}",
        "",
        "## Scores (categories 1–4; adversarial excluded)",
        "",
        f"| Metric | Value |",
        f"| --- | ---: |",
        f"| Overall | {m.overall_accuracy:.3f} ({m.correct}/{m.total}) |",
        f"| Search p50 ms | {m.latency_p50_ms or 0:.1f} |",
        f"| Search p95 ms | {m.latency_p95_ms or 0:.1f} |",
        "",
        "| Category | Acc | n |",
        "| --- | ---: | ---: |",
    ]
    for name, bucket in sorted(m.by_group.items()):
        lines.append(f"| {name} | {bucket.accuracy:.3f} | {bucket.total} |")
    lines += [
        "",
        "## Proveability",
        "",
    ]
    if gaps:
        lines.append("Gaps:")
        for g in gaps:
            lines.append(f"- {g}")
    else:
        lines.append("Pins present (dataset SHA, judge model, brainy URL/commit).")
    if meta.judge_model.startswith("lexical"):
        lines += [
            "",
            "> **Not publishable as LOCOMO J-score.** Re-run with a pinned LLM "
            "(Ollama / vLLM / OpenAI-compatible) via `LLM_BASE_URL` + `LLM_MODEL`.",
        ]
    lines += [
        "",
        "## Outlinks",
        "",
        f"- Dataset upstream: {LOCOMO_UPSTREAM}",
        f"- Paper: {LOCOMO_PAPER}",
        "- Harness peer: https://github.com/mem0ai/memory-benchmarks",
        "",
    ]
    return "\n".join(lines) + "\n"


def main() -> None:
    parser = argparse.ArgumentParser(description="Proveable LOCOMO smoke for Brainy")
    parser.add_argument("--base-url", default="")
    parser.add_argument(
        "--system",
        choices=("brainy", "mem0"),
        default="brainy",
        help="Backend for same-pin compares (mem0 requires MEM0_API_KEY)",
    )
    parser.add_argument("--conversations", type=int, default=1)
    parser.add_argument("--questions", type=int, default=30, help="Max questions across all convos")
    parser.add_argument(
        "--eval-lane",
        choices=("product-recall", "industry-search"),
        default="",
        help="R10 freeze label: product POST /recall vs search+shared-answerer+shared-judge. "
        "industry-search defaults --top-k 200 when --top-k is omitted.",
    )
    # Mem0 reports top_200; product /recall stays at 30 unless overridden.
    parser.add_argument("--top-k", type=int, default=None)
    parser.add_argument(
        "--answerer-model",
        default="",
        help="Override answerer model (default: LLM_MODEL or llama3.1/gpt-4o-mini)",
    )
    parser.add_argument(
        "--judge-model",
        default="",
        help="Override judge model (default: same as answerer / LLM_MODEL)",
    )
    parser.add_argument(
        "--llm-base-url",
        default="",
        help="OpenAI-compatible base URL (default: LLM_BASE_URL or api.openai.com/v1)",
    )
    parser.add_argument("--llm-api-key", default="", help="Overrides LLM_API_KEY / OPENAI_API_KEY")
    parser.add_argument("--lexical-only", action="store_true")
    parser.add_argument(
        "--sync-ingest",
        action="store_true",
        help="Use deterministic sync /ingest instead of async worker path (default: async)",
    )
    parser.add_argument(
        "--async-timeout",
        type=float,
        default=900.0,
        help="Seconds to wait for async extract to become searchable (LOCOMO-sized queues need ~10m)",
    )
    parser.add_argument("--run-id", default="")
    parser.add_argument(
        "--out-dir",
        default=str(ROOT / "docs" / "benchmarks" / "runs"),
    )
    parser.add_argument("--report", default="", help="Optional markdown report path")
    parser.add_argument(
        "--failure-ledger",
        default=str(ROOT / "docs/benchmarks/artifacts/failure-ledger/locomo-smoke.jsonl"),
        help="Append-only JSONL path for stage-oracle failure labels (WRONG/SKIPPED)",
    )
    args = parser.parse_args()
    run(args)


if __name__ == "__main__":
    main()
