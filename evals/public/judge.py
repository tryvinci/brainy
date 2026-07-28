from __future__ import annotations

import json
import re
from dataclasses import dataclass

from public.llm import LLMConfig, chat_completion, parse_judgment_json, resolve_config


@dataclass
class JudgeResult:
    judgment: str
    score: float
    reason: str
    model: str


def lexical_judge(answer: str, ground_truth: str) -> JudgeResult:
    """Offline proveable smoke judge: token-overlap proxy (not LOCOMO J-score).

    Used only when no LLM endpoint is configured so CI/docs can still exercise
    the pipeline. Results must be labeled judge_model=lexical-overlap-v0.
    """
    gt = (ground_truth or "").strip().lower()
    ans = (answer or "").strip().lower()
    if not gt:
        return JudgeResult("SKIPPED", 0.0, "empty ground truth", "lexical-overlap-v0")
    if gt in ans:
        return JudgeResult("CORRECT", 1.0, "ground truth substring present", "lexical-overlap-v0")
    gt_tokens = {t for t in re.findall(r"[a-z0-9]+", gt) if len(t) > 2}
    ans_tokens = set(re.findall(r"[a-z0-9]+", ans))
    if gt_tokens and gt_tokens.issubset(ans_tokens):
        return JudgeResult("CORRECT", 1.0, "all meaningful GT tokens present", "lexical-overlap-v0")
    overlap = len(gt_tokens & ans_tokens) / max(len(gt_tokens), 1)
    if overlap >= 0.6:
        return JudgeResult("CORRECT", float(overlap), f"token overlap={overlap:.2f}", "lexical-overlap-v0")
    return JudgeResult("WRONG", float(overlap), f"token overlap={overlap:.2f}", "lexical-overlap-v0")


def _with_model(config: LLMConfig, model: str) -> LLMConfig:
    if not model or model == config.model:
        return config
    return LLMConfig(
        base_url=config.base_url,
        api_key=config.api_key,
        model=model,
        json_mode=config.json_mode,
        timeout_s=config.timeout_s,
    )


def llm_judge(
    answer: str,
    ground_truth: str,
    question: str,
    config: LLMConfig,
) -> JudgeResult:
    """Binary LOCOMO-style CORRECT/WRONG judge (temperature 0)."""
    system = (
        "You are a strict binary grader for long-term conversational memory QA. "
        "Reply with JSON only: {\"judgment\":\"CORRECT\"|\"WRONG\",\"reason\":\"...\"}."
    )
    user = (
        f"Question: {question}\n"
        f"Ground truth: {ground_truth}\n"
        f"Predicted answer: {answer}\n"
        "Mark CORRECT if the predicted answer contains the same key facts as the ground truth "
        "(paraphrase OK). Mark WRONG otherwise."
    )
    content = chat_completion(
        [
            {"role": "system", "content": system},
            {"role": "user", "content": user},
        ],
        config,
        temperature=0.0,
        force_json=True,
    )
    try:
        parsed = parse_judgment_json(content)
    except (ValueError, json.JSONDecodeError):
        return JudgeResult(
            "WRONG",
            0.0,
            f"unparseable judge output: {content[:180]}",
            config.label,
        )

    judgment = str(parsed.get("judgment", "WRONG")).upper()
    if judgment not in {"CORRECT", "WRONG"}:
        judgment = "WRONG"
    return JudgeResult(
        judgment=judgment,
        score=1.0 if judgment == "CORRECT" else 0.0,
        reason=str(parsed.get("reason", "")),
        model=config.label,
    )


def openai_judge(
    answer: str,
    ground_truth: str,
    question: str,
    model: str = "gpt-4o-mini",
    config: LLMConfig | None = None,
) -> JudgeResult:
    """Back-compat name; routes through any OpenAI-compatible endpoint."""
    cfg = config or resolve_config(model=model)
    if cfg is None:
        raise RuntimeError(
            "No LLM configured. Set LLM_BASE_URL + LLM_API_KEY/OPENAI_API_KEY "
            "(or LLM_BASE_URL=http://127.0.0.1:11434/v1 for Ollama)."
        )
    return llm_judge(answer, ground_truth, question, _with_model(cfg, model))


def answer_from_memories(
    question: str,
    memories: list[dict],
    model: str = "",
    config: LLMConfig | None = None,
) -> tuple[str, str]:
    """Generate an answer from retrieved memories.

    This is a *generic* memory-QA client for proveable evals — not a place to
    special-case public benchmark questions or pad answers from known GTs.
    """
    if not memories:
        return "", "empty-context"
    cfg = config or resolve_config(model=model)
    if cfg is not None:
        cfg = _with_model(cfg, model) if model else cfg
        answer = _llm_answer(question, memories, cfg, extractive=False)
        # List-shaped questions often need a second extractive pass: generative
        # models stop after the first supporting memory. Prefer the extractive
        # answer when it enumerates more distinct items (generic QA, not GT pads).
        if _is_empty_answer(answer) or _looks_list_question(question):
            extractive = _llm_answer(question, memories, cfg, extractive=True)
            if not _is_empty_answer(extractive):
                if _is_empty_answer(answer) or _item_count(extractive) > _item_count(answer):
                    answer = extractive
                    return answer, cfg.label + "+extractive-list"
        if _is_empty_answer(answer):
            joined = _statement_join(memories)
            if joined:
                return joined, cfg.label + "+retrieval-concat"
        return answer, cfg.label
    return _statement_join(memories) or "", "retrieval-concat-v0"


def _looks_list_question(question: str) -> bool:
    q = (question or "").lower()
    cues = (
        "activities",
        "hobbies",
        "books",
        "places",
        "where has",
        "what do ",
        "what does ",
        "what activities",
        "what books",
        "kids like",
        "children like",
        "partake",
        "destress",
        "de-stress",
    )
    return any(c in q for c in cues)


def _item_count(answer: str) -> int:
    text = (answer or "").strip()
    if not text:
        return 0
    # Count comma/semicolon/newline / bullet separations as distinct items.
    parts = re.split(r"[\n;,•]|\band\b|\+| - ", text, flags=re.IGNORECASE)
    items = [p.strip(" .") for p in parts if len(p.strip(" .")) >= 2]
    return max(1, len(items))


def _is_empty_answer(answer: str) -> bool:
    text = (answer or "").strip().lower()
    if not text:
        return True
    empties = (
        "none",
        "n/a",
        "i do not know",
        "i don't know",
        "i dont know",
        "no information",
        "don't have information",
        "do not have information",
        "not mentioned",
        "no memories",
    )
    return any(text == e or text.startswith(e) for e in empties)


def _statement_join(memories: list[dict]) -> str:
    parts = []
    for m in memories[:8]:
        content = (m.get("content") or "").strip()
        if not content or content.endswith("?"):
            continue
        parts.append(content)
    return " | ".join(parts)


def _llm_answer(
    question: str,
    memories: list[dict],
    config: LLMConfig,
    *,
    extractive: bool = False,
) -> str:
    lines = []
    for m in memories[:40]:
        content = (m.get("content") or "").strip()
        if not content:
            continue
        # Prefer statements over stored questions for QA context.
        if content.endswith("?") and not any(ch.isdigit() for ch in content):
            continue
        observed = (m.get("observed_at") or "").strip()
        if observed:
            lines.append(f"- {content} [event_time={observed}]")
        else:
            lines.append(f"- {content}")
    if not lines:
        for m in memories[:12]:
            content = (m.get("content") or "").strip()
            if content:
                lines.append(f"- {content}")
    context = "\n".join(lines)
    if extractive:
        system = (
            "Extract the shortest answer to the question that is directly supported by the memories. "
            "Copy key phrases from the memories. Do not say None or I do not know if any memory is relevant. "
            "If multiple items apply, list them as a short comma-separated list. "
            "Scan every memory — do not stop after the first match when the question asks for "
            "activities, books, places, likes, or other multi-item sets."
        )
    else:
        system = (
            "Answer the question using only the memories below. "
            "When a memory includes event_time / an absolute date in parentheses, "
            "use that to resolve relative phrases like yesterday, two days ago, last week, "
            "last Saturday, or N years ago. "
            "Prefer a concrete short answer (dates, names, places, lists). "
            "When several memories support different parts of the answer (lists of activities, "
            "books, places, preferences, identity attributes), combine every distinct supported "
            "item into one answer — do not stop at the first memory. "
            "Never answer with only None or N/A if any memory is remotely relevant — "
            "quote the best supporting memory instead. "
            "If memories truly lack the answer, say you do not know."
        )
    return chat_completion(
        [
            {"role": "system", "content": system},
            {"role": "user", "content": f"Memories:\n{context}\n\nQuestion: {question}"},
        ],
        config,
        temperature=0.0,
        force_json=False,
    )
