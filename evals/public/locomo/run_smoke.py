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
from public.proveability import RunManifest, require_pins  # noqa: E402
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
        for speaker, text in session.get("turns") or []:
            batch.append({"role": "user", "content": f"{speaker}: {text}"})
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
        backend.wait_until_any_searchable(user_id, probes or ["conversation"])
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
    backend = BrainyBackend(
        base_url,
        tenant_prefix=f"locomo-{nonce}",
        async_ingest=async_ingest,
        async_timeout_s=float(getattr(args, "async_timeout", 180.0)),
    )
    print(
        f"ingest_mode={'async' if async_ingest else 'sync'} "
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
    question_budget = args.questions

    for conv_idx, conversation in enumerate(conversations):
        sample_id = str(conversation.get("sample_id") or f"c{conv_idx}")
        user_id = f"{sample_id}"
        sessions = iter_sessions(conversation)
        n_turns = sum(len(s.get("turns") or []) for s in sessions)
        n_ingested = ingest_conversation(backend, user_id, sessions)
        print(f"[{sample_id}] ingested {n_ingested} turns ({n_turns} total, {len(sessions)} sessions)", flush=True)

        questions = iter_questions(conversation)
        for qa in questions:
            if question_budget is not None and len(items) >= question_budget:
                break
            cat_id = qa.get("category")
            try:
                cat_id_int = int(cat_id) if cat_id is not None else 0
            except (TypeError, ValueError):
                cat_id_int = 0
            group = CATEGORY_NAMES.get(cat_id_int, f"cat-{cat_id}")

            results, latency_ms = backend.recall(user_id, qa["question"], top_k=args.top_k)
            answer, gen_model = answer_from_memories(
                qa["question"], results, model=answerer_model, config=answerer_cfg
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
        if question_budget is not None and len(items) >= question_budget:
            break

    metrics = compute_metrics(items, CATEGORIES_TO_SCORE)
    commit = git_commit()
    answerer_pin = (answerer_cfg.label if answerer_cfg else "") or "retrieval-concat-v0"
    judge_pin = (judge_cfg.label if judge_cfg else "") or "lexical-overlap-v0"
    result = UnifiedResult(
        metadata=Metadata(
            benchmark="locomo-smoke",
            project_name="brainy",
            run_id=run_id,
            timestamp=utc_now(),
            dataset_url=LOCOMO_DATASET_URL,
            dataset_sha256=dataset_sha,
            brainy_url=base_url,
            brainy_commit=commit,
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
        notes="Smoke run — not full LOCOMO. lexical judge is not publishable J-score.",
        extras={
            "llm_base_url": (llm.base_url if llm else ""),
            "ingest_mode": "async" if async_ingest else "sync",
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
    parser.add_argument("--conversations", type=int, default=1)
    parser.add_argument("--questions", type=int, default=30, help="Max questions across all convos")
    parser.add_argument("--top-k", type=int, default=15)
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
        default=180.0,
        help="Seconds to wait for async extract to become searchable",
    )
    parser.add_argument("--run-id", default="")
    parser.add_argument(
        "--out-dir",
        default=str(ROOT / "docs" / "benchmarks" / "runs"),
    )
    parser.add_argument("--report", default="", help="Optional markdown report path")
    args = parser.parse_args()
    run(args)


if __name__ == "__main__":
    main()
