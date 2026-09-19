"""Shared OpMem external-lane adapter protocol."""

from __future__ import annotations

import hashlib
import json
import os
import time
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

from opmem_external.pins import LANE_ENV_REQUIRED, PINNED_CHAT_MODEL, PINNED_EMBED_MODEL
from opmem_external.token_estimate import LaneDryRunStats, OperationCounts, TokenLedger


@dataclass
class AdapterResponseLog:
    lane: str
    task: str
    op: str
    status: int
    body_preview: str
    redacted: bool = True


class OpMemLaneAdapter(ABC):
    """Maps neutral remember/recall/revise/forget onto one external memory lane."""

    name: str
    supports_pinned_models: bool = True
    write_settle_seconds: float = 0.0

    def __init__(self, *, dry_run: bool = True, run_id: str | None = None) -> None:
        self.dry_run = dry_run
        self._run_id = run_id or os.environ.get("OPMEM_RUN_ID", "opmem-pin-0001")
        self._task_name = ""
        self._stats = LaneDryRunStats(lane=self.name)
        self._memory_table: dict[str, dict[str, str]] = {}
        self._response_logs: list[AdapterResponseLog] = []
        self._missing_credentials = self._check_credentials()
        self._stats.missing_credentials = list(self._missing_credentials)
        self._stats.credentials_present = not self._missing_credentials

    @abstractmethod
    def lane_id(self) -> str:
        ...

    def _check_credentials(self) -> list[str]:
        missing = []
        for key in LANE_ENV_REQUIRED.get(self.name, ()):
            if not os.environ.get(key, "").strip():
                missing.append(key)
        return missing

    def available(self) -> bool:
        if self.dry_run:
            return True
        return not self._missing_credentials

    def begin_task(self, task_name: str) -> None:
        self._task_name = task_name
        self.reset_namespace()

    def namespace(self, actor: tuple[str, str]) -> str:
        tenant, subject = actor
        raw = f"{self._run_id}|{self.name}|{self._task_name}|{tenant}|{subject}"
        digest = hashlib.sha256(raw.encode("utf-8")).hexdigest()[:16]
        return f"opmem-{digest}"

    def reset_namespace(self) -> None:
        self._stats.ops.reset += 1
        keys = [k for k in self._memory_table if k.startswith(f"{self._task_name}:")]
        for k in keys:
            del self._memory_table[k]
        if not self.dry_run:
            self._reset_remote()

    @abstractmethod
    def _reset_remote(self) -> None:
        ...

    def remember(self, actor: tuple[str, str], content: str) -> list[str]:
        self._stats.ops.remember += 1
        ns = self._table_key(actor)
        mem_id = self._deterministic_id("remember", content)
        if self.dry_run:
            self._simulate_remember(content)
            self._memory_table.setdefault(ns, {})[mem_id] = content
            self._log_response("remember", 200, {"memory_id": mem_id, "dry_run": True})
            return [mem_id]
        return self._remember_live(actor, content, mem_id)

    def recall(self, actor: tuple[str, str], query: str) -> list[dict]:
        self._stats.ops.recall += 1
        if self.dry_run:
            self._simulate_recall(query)
            ns = self._table_key(actor)
            rows = self._memory_table.get(ns, {})
            scored = sorted(rows.items(), key=lambda item: -_overlap(query, item[1]))
            out = [{"id": mid, "content": txt} for mid, txt in scored if _overlap(query, txt) > 0]
            self._log_response("recall", 200, {"results": out, "dry_run": True})
            return out
        return self._recall_live(actor, query)

    def forget(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        self._stats.ops.forget += 1
        if self.dry_run:
            ns = self._table_key(actor)
            bucket = self._memory_table.get(ns, {})
            for mid in memory_ids:
                bucket.pop(mid, None)
            self._simulate_forget()
            self._log_response("forget", 200, {"deleted": memory_ids, "dry_run": True})
            return
        self._forget_live(actor, memory_ids)

    def revise(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        self._stats.ops.revise += 1
        if self.dry_run:
            ns = self._table_key(actor)
            bucket = self._memory_table.get(ns, {})
            if memory_ids:
                bucket[memory_ids[0]] = content
            self._simulate_revise(content)
            self._log_response("revise", 200, {"memory_id": memory_ids[:1], "dry_run": True})
            if self.write_settle_seconds:
                time.sleep(self.write_settle_seconds)
            return
        self._revise_live(actor, memory_ids, content)

    def _table_key(self, actor: tuple[str, str]) -> str:
        return f"{self._task_name}:{self.namespace(actor)}"

    def _deterministic_id(self, op: str, content: str) -> str:
        seed = f"{self._run_id}:{self._task_name}:{op}:{content}"
        return "m-" + hashlib.sha256(seed.encode("utf-8")).hexdigest()[:12]

    def _log_response(self, op: str, status: int, body: Any) -> None:
        preview = json.dumps(body, ensure_ascii=False)[:500]
        self._response_logs.append(
            AdapterResponseLog(
                lane=self.name,
                task=self._task_name,
                op=op,
                status=status,
                body_preview=preview,
            )
        )
        self._stats.ops.http_calls += 1

    def _record_model_constraint(self, reason: str) -> None:
        if reason not in self._stats.product_constraints:
            self._stats.product_constraints.append(reason)

    def _assert_pinned_models(self, chat_ok: bool, embed_ok: bool) -> None:
        if not chat_ok:
            self._record_model_constraint(
                f"lane {self.name} cannot pin chat model {PINNED_CHAT_MODEL}"
            )
        if not embed_ok:
            self._record_model_constraint(
                f"lane {self.name} cannot pin embed model {PINNED_EMBED_MODEL}"
            )

    def _simulate_remember(self, content: str) -> None:
        self._stats.tokens.add_chat(
            f"[extract] model={PINNED_CHAT_MODEL}\n{content}",
            '{"facts":[]}',
        )
        self._stats.tokens.add_embed(content)

    def _simulate_recall(self, query: str) -> None:
        self._stats.tokens.add_embed(query)

    def _simulate_revise(self, content: str) -> None:
        self._stats.tokens.add_chat(f"[update] model={PINNED_CHAT_MODEL}\n{content}", "{}")

    def _simulate_forget(self) -> None:
        pass

    @abstractmethod
    def _remember_live(self, actor: tuple[str, str], content: str, mem_id: str) -> list[str]:
        ...

    @abstractmethod
    def _recall_live(self, actor: tuple[str, str], query: str) -> list[dict]:
        ...

    @abstractmethod
    def _forget_live(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        ...

    @abstractmethod
    def _revise_live(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        ...

    def dry_run_stats(self) -> LaneDryRunStats:
        return self._stats

    def write_raw_log(self, path: Path) -> None:
        path.parent.mkdir(parents=True, exist_ok=True)
        rows = [
            {
                "lane": r.lane,
                "task": r.task,
                "op": r.op,
                "status": r.status,
                "body_preview": r.body_preview,
            }
            for r in self._response_logs
        ]
        path.write_text(json.dumps(rows, indent=2) + "\n", encoding="utf-8")


def _overlap(query: str, content: str) -> int:
    q = {t for t in query.lower().split() if t}
    c = {t for t in content.lower().split() if t}
    return len(q & c)
