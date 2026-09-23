"""Conservative pre-call token budget from serialized request payloads (fixtures × lane)."""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Iterable

import tiktoken

from opmem_external.pins import PINNED_CHAT_MODEL, PINNED_EMBED_MODEL
from opmem_external.rates import (
    RATE_CHAT_INPUT_PER_M,
    RATE_CHAT_OUTPUT_PER_M,
    RATE_EMBED_PER_M,
    RATE_SOURCE_NOTE,
    RATE_SOURCE_URL,
)

_ENC = tiktoken.get_encoding("cl100k_base")

# Conservative max completion tokens per call type (upper bound, not typical).
_MAX_OUTPUT_BY_KIND = {
    "mem0_extract": 800,
    "mem0_update": 400,
    "letta_memory_step": 300,
    "langmem_manage": 400,
}

_RETRY_MARGIN = 1.10  # settling / transient retries


def _count(text: str) -> int:
    return len(_ENC.encode(text or ""))


def _count_json(obj: Any) -> int:
    return _count(json.dumps(obj, ensure_ascii=False))


@dataclass
class BudgetLine:
    lane: str
    op: str
    task: str
    kind: str
    input_tokens: int
    output_tokens: int
    embed_tokens: int

    def chat_usd(self) -> float:
        return (
            self.input_tokens * RATE_CHAT_INPUT_PER_M / 1_000_000
            + self.output_tokens * RATE_CHAT_OUTPUT_PER_M / 1_000_000
        )

    def embed_usd(self) -> float:
        return self.embed_tokens * RATE_EMBED_PER_M / 1_000_000

    def total_usd(self) -> float:
        return self.chat_usd() + self.embed_usd()


@dataclass
class LaneBudget:
    lane: str
    lines: list[BudgetLine] = field(default_factory=list)
    product_constraints: list[str] = field(default_factory=list)

    def worst_case_usd(self) -> float:
        base = sum(line.total_usd() for line in self.lines)
        if self.lane == "zep":
            return 0.0
        return base * _RETRY_MARGIN

    def token_totals(self) -> dict[str, int]:
        return {
            "chat_input_tokens": sum(l.input_tokens for l in self.lines),
            "chat_output_tokens": sum(l.output_tokens for l in self.lines),
            "embed_tokens": sum(l.embed_tokens for l in self.lines),
        }

    def to_dict(self) -> dict:
        return {
            "lane": self.lane,
            "worst_case_usd": round(self.worst_case_usd(), 6),
            "token_totals": self.token_totals(),
            "line_count": len(self.lines),
            "product_constraints": self.product_constraints,
            "retry_margin": _RETRY_MARGIN,
        }


def _mem0_system_prompt() -> str:
    from mem0.configs.prompts import ADDITIVE_EXTRACTION_PROMPT

    return ADDITIVE_EXTRACTION_PROMPT


def _mem0_extract_payload(content: str) -> list[dict]:
    system = _mem0_system_prompt()
    user = json.dumps(
        {
            "new_messages": [{"role": "user", "content": content}],
            "summary": "",
            "recently_extracted_memories": [],
            "existing_memories": [],
            "last_k_messages": [],
            "observation_date": "2026-09-19",
            "current_date": "2026-09-19",
        },
        ensure_ascii=False,
    )
    return [
        {"role": "system", "content": system},
        {"role": "user", "content": user},
    ]


def _letta_wrapped(content: str) -> dict:
    return {
        "model": PINNED_CHAT_MODEL,
        "messages": [
            {
                "role": "system",
                "content": "Extract durable user memory facts; memory-only lane (no answer generation).",
            },
            {"role": "user", "content": content},
        ],
        "max_tokens": _MAX_OUTPUT_BY_KIND["letta_memory_step"],
    }


def _langmem_manage_payload(content: str) -> dict:
    return {
        "model": PINNED_CHAT_MODEL,
        "instructions": "Manage memory store entries for user facts (OpMem lane; no final answer).",
        "input": content,
    }


def _zep_add_payload(content: str) -> dict:
    return {
        "messages": [
            {
                "role_type": "user",
                "role": "opmem-user",
                "content": content,
            }
        ],
        "return_context": False,
    }


def build_lane_budgets(tasks: Iterable[dict]) -> dict[str, LaneBudget]:
    budgets = {
        "mem0-oss": LaneBudget("mem0-oss"),
        "letta": LaneBudget("letta"),
        "langmem": LaneBudget("langmem"),
        "zep": LaneBudget("zep"),
    }
    budgets["zep"].product_constraints.append(
        "Zep Cloud native extraction; OpenAI pins not applied; $0 paid ceiling."
    )

    for task in tasks:
        name = task["name"]
        for step in task["steps"]:
            op = step["op"]
            content = step.get("content", "")
            query = step.get("query", "")

            if op == "remember":
                msgs = _mem0_extract_payload(content)
                inp = sum(_count(m["content"]) for m in msgs)
                budgets["mem0-oss"].lines.append(
                    BudgetLine(
                        "mem0-oss", op, name, "mem0_extract", inp,
                        _MAX_OUTPUT_BY_KIND["mem0_extract"],
                        _count(content),
                    )
                )
                letta_p = _letta_wrapped(content)
                budgets["letta"].lines.append(
                    BudgetLine(
                        "letta", op, name, "letta_memory", _count_json(letta_p),
                        _MAX_OUTPUT_BY_KIND["letta_memory_step"],
                        _count(content),
                    )
                )
                lm = _langmem_manage_payload(content)
                budgets["langmem"].lines.append(
                    BudgetLine(
                        "langmem", op, name, "langmem_manage", _count_json(lm),
                        _MAX_OUTPUT_BY_KIND["langmem_manage"],
                        _count(content),
                    )
                )
                budgets["zep"].lines.append(
                    BudgetLine("zep", op, name, "zep_add", _count_json(_zep_add_payload(content)), 0, 0)
                )
            elif op == "recall":
                q = query or ""
                for lane in ("mem0-oss", "letta", "langmem"):
                    budgets[lane].lines.append(
                        BudgetLine(lane, op, name, "embed_query", 0, 0, _count(q))
                    )
                # mem0 may re-embed on search path conservatively
                budgets["mem0-oss"].lines.append(
                    BudgetLine("mem0-oss", op, name, "embed_query_dup", 0, 0, _count(q))
                )
                budgets["zep"].lines.append(
                    BudgetLine("zep", op, name, "zep_get", _count_json({"query": q}), 0, 0)
                )
            elif op == "revise":
                msgs = _mem0_extract_payload(content)
                inp = sum(_count(m["content"]) for m in msgs)
                budgets["mem0-oss"].lines.append(
                    BudgetLine(
                        "mem0-oss", op, name, "mem0_update", inp,
                        _MAX_OUTPUT_BY_KIND["mem0_update"],
                        _count(content),
                    )
                )
                budgets["letta"].lines.append(
                    BudgetLine(
                        "letta", op, name, "letta_update", _count_json(_letta_wrapped(content)),
                        _MAX_OUTPUT_BY_KIND["letta_memory_step"],
                        _count(content),
                    )
                )
                budgets["langmem"].lines.append(
                    BudgetLine(
                        "langmem", op, name, "langmem_update", _count_json(_langmem_manage_payload(content)),
                        _MAX_OUTPUT_BY_KIND["langmem_manage"],
                        _count(content),
                    )
                )
                budgets["zep"].lines.append(
                    BudgetLine("zep", op, name, "zep_update", _count_json({"text": content}), 0, 0)
                )
            elif op == "forget":
                for lane in ("mem0-oss", "letta", "langmem"):
                    budgets[lane].lines.append(
                        BudgetLine(lane, op, name, "delete_api", 50, 0, 0)
                    )
                budgets["zep"].lines.append(
                    BudgetLine("zep", op, name, "zep_delete", 50, 0, 0)
                )

    return budgets


def build_worksheet_from_budgets(
    budgets: dict[str, LaneBudget],
    *,
    task_count: int,
    step_count: int,
) -> dict:
    from opmem_external.pins import SPEND_CEILING_USD

    lanes = []
    openai_total = 0.0
    for lane, budget in sorted(budgets.items()):
        usd = budget.worst_case_usd()
        openai_total += usd
        lanes.append(
            {
                **budget.to_dict(),
                "projected_worst_case_usd": round(usd, 6),
                "ceiling_usd": SPEND_CEILING_USD.get(lane, 0.0),
                "within_ceiling": usd <= SPEND_CEILING_USD.get(lane, 0.0),
            }
        )
    lanes.append(
        {
            "lane": "_total_openai_worst_case",
            "projected_worst_case_usd": round(openai_total, 6),
            "ceiling_usd": SPEND_CEILING_USD["total"],
            "within_ceiling": openai_total <= SPEND_CEILING_USD["total"],
        }
    )
    return {
        "mode": "payload_budget",
        "task_count": task_count,
        "step_count": step_count,
        "pinned_chat_model": PINNED_CHAT_MODEL,
        "pinned_embed_model": PINNED_EMBED_MODEL,
        "rate_source_url": RATE_SOURCE_URL,
        "rate_source_note": RATE_SOURCE_NOTE,
        "ceilings_usd": SPEND_CEILING_USD,
        "lanes": lanes,
        "methodology": (
            "tiktoken cl100k_base on serialized request bodies incl. mem0 ADDITIVE_EXTRACTION_PROMPT; "
            "conservative max output caps; 1.10 retry margin; Zep excluded from OpenAI totals."
        ),
    }


def load_tasks(fixture_dir: Path) -> list[dict]:
    return [json.loads(p.read_text(encoding="utf-8")) for p in sorted(fixture_dir.glob("*.json"))]
