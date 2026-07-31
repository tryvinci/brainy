#!/usr/bin/env python3
"""BEAM sample runner against Brainy (master-plan E3).

Downloads BEAM conversation buckets from HuggingFace (Mohammadta/BEAM) and
runs ingest→search→answer→judge using the same CF OpenAI-compatible stack.
Defaults to a small 100K sample for an affordable first report.
"""
from __future__ import annotations

import argparse
import json
import os
import pathlib
import subprocess
import sys
import uuid

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "evals"))

from public.backends.brainy import BrainyBackend  # noqa: E402
from public.judge import answer_from_memories, lexical_judge, llm_judge  # noqa: E402
from public.llm import resolve_config  # noqa: E402
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

DEFAULT_CACHE = pathlib.Path("/workspace/datasets/beam")


def git_commit() -> str:
    try:
        return (
            subprocess.check_output(["git", "rev-parse", "--short", "HEAD"], cwd=ROOT)
            .decode()
            .strip()
        )
    except Exception:
        return ""


def download_beam(chat_size: str, cache_dir: pathlib.Path) -> list[dict]:
    cache_dir.mkdir(parents=True, exist_ok=True)
    out = cache_dir / f"beam_{chat_size}.json"
    if out.exists():
        print(f"Using cached {out}", flush=True)
        return json.loads(out.read_text(encoding="utf-8"))

    try:
        from datasets import load_dataset
    except ImportError as exc:
        raise SystemExit(
            "datasets package required for BEAM download: pip install datasets\n" + str(exc)
        )

    split_map = {"100K": "100K", "500K": "500K", "1M": "1M", "10M": "10M"}
    if chat_size not in split_map:
        raise SystemExit(f"unsupported chat size {chat_size}; choose from {sorted(split_map)}")

    name = "Mohammadta/BEAM-10M" if chat_size == "10M" else "Mohammadta/BEAM"
    print(f"Downloading {name} split={split_map[chat_size]} …", flush=True)
    ds = load_dataset(name, split=split_map[chat_size])
    rows = [dict(ds[i]) for i in range(len(ds))]
    out.write_text(json.dumps(rows), encoding="utf-8")
    print(f"Cached {len(rows)} conversations → {out}", flush=True)
    return rows


def _messages_from_conversation(conv: dict):
    """Yield (session_date, messages) from a BEAM conversation."""
    if "sessions" in conv:
        for sess in conv["sessions"] or []:
            date = sess.get("date") or sess.get("timestamp") or sess.get("session_date")
            msgs = sess.get("messages") or sess.get("dialog") or sess.get("turns") or []
            cleaned = []
            for m in msgs:
                if isinstance(m, dict):
                    role = (m.get("role") or m.get("speaker") or "user").strip()
                    content = (m.get("content") or m.get("text") or "").strip()
                elif isinstance(m, (list, tuple)) and len(m) >= 2:
                    role, content = str(m[0]), str(m[1])
                else:
                    continue
                if not content:
                    continue
                if role not in ("user", "assistant", "system"):
                    role = "user"
                cleaned.append({"role": role, "content": content})
            if cleaned:
                yield date, cleaned
        return

    # Flat BEAM "chat" may be list-of-batches (list[list[turn]]) or list[turn]
    chat = conv.get("chat")
    if isinstance(chat, list) and chat:
        # list of batches
        if isinstance(chat[0], list):
            for batch in chat:
                cleaned = []
                for m in batch:
                    if not isinstance(m, dict):
                        continue
                    role = (m.get("role") or "user").strip()
                    content = (m.get("content") or "").strip()
                    if not content:
                        continue
                    if role not in ("user", "assistant", "system"):
                        role = "user"
                    cleaned.append({"role": role, "content": content})
                if cleaned:
                    yield None, cleaned
            return
        msgs = chat
    else:
        msgs = conv.get("messages") or conv.get("conversation") or []
    cleaned = []
    for m in msgs:
        if not isinstance(m, dict):
            continue
        role = (m.get("role") or "user").strip()
        content = (m.get("content") or "").strip()
        if not content:
            continue
        if role not in ("user", "assistant", "system"):
            role = "user"
        cleaned.append({"role": role, "content": content})
    if cleaned:
        yield conv.get("date"), cleaned


def ingest_conversation(backend: BrainyBackend, user_id: str, conv: dict, chunk: int = 8) -> int:
    n = 0
    probes: list[str] = []
    for date, messages in _messages_from_conversation(conv):
        meta: dict = {}
        if date:
            meta["observed_at"] = str(date)
        for i in range(0, len(messages), chunk):
            batch = messages[i : i + chunk]
            try:
                backend.remember_messages(user_id, batch, metadata=meta or None, wait=False)
            except Exception:
                # WAF can 403 multi-turn batches; fall back to single-message posts.
                for msg in batch:
                    backend.remember_messages(user_id, [msg], metadata=meta or None, wait=False)
            n += len(batch)
            tok = (batch[-1].get("content") or "").split()
            if tok:
                probes.append(tok[0][:40])
    if n and backend.async_ingest:
        backend.wait_until_any_searchable(user_id, probes[:3] or ["conversation"], settle_polls=6)
    return n


def iter_questions(conv: dict) -> list[dict]:
    """Normalize BEAM probing_questions (dict-by-type) or flat qa lists."""
    out = []
    pq = conv.get("probing_questions")
    if isinstance(pq, str):
        import ast
        try:
            pq = ast.literal_eval(pq)
        except Exception:
            try:
                pq = json.loads(pq)
            except Exception:
                pq = {}
    if isinstance(pq, dict) and pq:
        for q_type, type_questions in pq.items():
            rows = type_questions if isinstance(type_questions, list) else [type_questions]
            for i, q in enumerate(rows):
                if isinstance(q, str):
                    out.append({"id": f"{q_type}-{i}", "question": q, "answer": "", "type": str(q_type)})
                elif isinstance(q, dict):
                    # BEAM rubrics: use nugget descriptions joined as GT proxy when no answer
                    rubric = q.get("rubric") or {}
                    nuggets = []
                    if isinstance(rubric, dict):
                        for n in rubric.get("nuggets") or []:
                            if isinstance(n, dict):
                                nuggets.append(str(n.get("description") or n.get("nugget") or ""))
                            else:
                                nuggets.append(str(n))
                    gt = str(q.get("answer") or q.get("ground_truth") or q.get("gt") or " | ".join(x for x in nuggets if x))
                    out.append({
                        "id": str(q.get("id") or q.get("question_id") or f"{q_type}-{i}"),
                        "question": str(q.get("question") or q.get("question_text") or q.get("text") or ""),
                        "answer": gt,
                        "type": str(q.get("question_type") or q_type),
                    })
        return out

    qs = conv.get("questions") or conv.get("qa") or conv.get("evals") or []
    for i, q in enumerate(qs):
        if not isinstance(q, dict):
            continue
        out.append(
            {
                "id": str(q.get("id") or q.get("question_id") or i),
                "question": str(q.get("question") or q.get("question_text") or q.get("text") or ""),
                "answer": str(q.get("answer") or q.get("ground_truth") or q.get("gt") or ""),
                "type": str(q.get("type") or q.get("question_type") or q.get("ability") or "unknown"),
            }
        )
    return out


def write_report(result: UnifiedResult, path: pathlib.Path) -> None:
    m = result.metrics
    lines = [
        f"# {result.metadata.benchmark} — {result.metadata.run_id}",
        "",
        f"- accuracy: **{m.overall_accuracy:.3f}** ({m.correct}/{m.total})",
        f"- errors: {m.errors}",
        f"- latency p50/p95: {m.latency_p50_ms} / {m.latency_p95_ms} ms",
        "",
        "## By ability",
        "",
    ]
    for name, g in sorted(m.by_group.items()):
        lines.append(f"- {name}: {g.correct}/{g.total} ({g.accuracy:.3f})")
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def run(args: argparse.Namespace) -> UnifiedResult:
    cache = pathlib.Path(args.cache_dir)
    conversations = download_beam(args.chat_size, cache)
    start_s, _, end_s = args.conversations.partition("-")
    start_i = int(start_s or 0)
    end_i = int(end_s or start_s or 0)
    selected = conversations[start_i : end_i + 1]
    print(
        f"BEAM {args.chat_size}: using conversations [{start_i},{end_i}] "
        f"({len(selected)} of {len(conversations)})",
        flush=True,
    )
    if selected:
        print(f"  sample keys: {sorted(selected[0].keys())}", flush=True)

    base_url = (
        args.base_url or os.environ.get("BRAINY_BASE_URL") or "http://127.0.0.1:8080"
    ).rstrip("/")
    run_id = args.run_id or f"beam-{args.chat_size}-{uuid.uuid4().hex[:8]}"
    nonce = uuid.uuid4().hex[:8]
    backend = BrainyBackend(
        base_url,
        tenant_prefix=f"beam-{nonce}",
        async_ingest=not args.sync_ingest,
        async_timeout_s=float(args.async_timeout),
    )
    llm = None if args.lexical_only else resolve_config(
        model=args.judge_model or args.answerer_model or "",
        base_url=args.llm_base_url,
        api_key=args.llm_api_key,
    )
    answerer_model = (args.answerer_model or (llm.model if llm else "")) if llm else ""
    if llm:
        print(f"LLM: {llm.label}", flush=True)

    items: list[EvalItem] = []
    for ci, conv in enumerate(selected):
        abs_i = start_i + ci
        user_id = f"{args.chat_size}_{abs_i}"
        print(f"[{args.chat_size}/{abs_i}] ingest…", flush=True)
        try:
            n_ing = ingest_conversation(backend, user_id, conv)
            print(f"  ingested ~{n_ing} msgs", flush=True)
        except Exception as exc:
            print(f"  INGEST ERROR {exc}", flush=True)
            continue

        for qi, qa in enumerate(iter_questions(conv)):
            if args.questions and qi >= args.questions:
                break
            qtext, gt, qtype = qa["question"], qa["answer"], qa["type"]
            if not qtext:
                continue
            try:
                results, latency_ms = backend.recall(user_id, qtext, top_k=args.top_k)
                if llm:
                    answer, gen_model = answer_from_memories(
                        qtext, results, model=answerer_model, config=llm
                    )
                    jr = llm_judge(answer, gt, qtext, llm)
                else:
                    answer = " | ".join((r.get("content") or "") for r in results[:8])
                    gen_model = "lexical"
                    jr = lexical_judge(answer, gt)
                items.append(
                    EvalItem(
                        id=qa["id"],
                        group=qtype,
                        question=qtext,
                        ground_truth=gt,
                        retrieval=RetrievalData(
                            search_query=qtext,
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
                        extras={"chat_size": args.chat_size, "conv_idx": abs_i},
                    )
                )
                print(f"  q{qi} {qtype} {jr.judgment} lat={latency_ms:.0f}ms", flush=True)
            except Exception as exc:
                print(f"  q{qi} ERROR {exc}", flush=True)
                items.append(
                    EvalItem(
                        id=qa["id"],
                        group=qtype,
                        question=qtext,
                        ground_truth=gt,
                        judgment=JudgmentData(
                            judgment="SKIPPED", score=0.0, reason=str(exc)[:300], model="error"
                        ),
                        extras={"error": str(exc)[:500]},
                    )
                )

    metrics = compute_metrics(items, score_groups=set())
    result = UnifiedResult(
        metadata=Metadata(
            benchmark=f"beam-{args.chat_size}",
            project_name="brainy",
            run_id=run_id,
            timestamp=utc_now(),
            dataset_url="https://huggingface.co/datasets/Mohammadta/BEAM",
            brainy_url=base_url,
            brainy_commit=git_commit(),
            answerer_model=(llm.label if llm else "lexical"),
            judge_model=(llm.label if llm else "lexical-overlap-v0"),
            top_k=args.top_k,
            config={
                "chat_size": args.chat_size,
                "conversations": args.conversations,
                "questions_per_conv": args.questions,
                "async_ingest": not args.sync_ingest,
            },
        ),
        metrics=metrics,
        evaluations=items,
    )
    out_dir = pathlib.Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    (out_dir / f"{run_id}.json").write_text(json.dumps(result.to_dict(), indent=2) + "\n")
    write_report(result, out_dir / f"{run_id}.md")
    print(
        f"DONE accuracy={metrics.overall_accuracy:.3f} correct={metrics.correct}/{metrics.total}",
        flush=True,
    )
    return result


def main() -> int:
    p = argparse.ArgumentParser(description="BEAM sample runner vs Brainy")
    p.add_argument("--base-url", default="")
    p.add_argument("--chat-size", default="100K", choices=("100K", "500K", "1M", "10M"))
    p.add_argument("--conversations", default="0-0", help="Inclusive range, e.g. 0-2")
    p.add_argument("--questions", type=int, default=0, help="0 = all questions per conv")
    p.add_argument("--top-k", type=int, default=30)
    p.add_argument("--sync-ingest", action="store_true")
    p.add_argument("--async-timeout", type=float, default=3600.0)
    p.add_argument("--lexical-only", action="store_true")
    p.add_argument("--answerer-model", default="")
    p.add_argument("--judge-model", default="")
    p.add_argument("--llm-base-url", default="")
    p.add_argument("--llm-api-key", default="")
    p.add_argument("--cache-dir", default=str(DEFAULT_CACHE))
    p.add_argument("--run-id", default="")
    p.add_argument("--out-dir", default=str(ROOT / "docs" / "benchmarks" / "runs"))
    args = p.parse_args()
    run(args)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
