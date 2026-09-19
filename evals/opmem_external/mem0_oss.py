"""Mem0 OSS lane: Python package + local Qdrant (no Platform API key)."""

from __future__ import annotations

import os
import shutil
from pathlib import Path
from typing import Any

from opmem_external.pins import PINNED_CHAT_MODEL, PINNED_EMBED_MODEL
from opmem_external.protocol import OpMemLaneAdapter
from opmem_external.telemetry import LaneTelemetry


class Mem0OSSLaneAdapter(OpMemLaneAdapter):
    name = "mem0-oss"
    write_settle_seconds = 2.0

    def __init__(self, *, dry_run: bool = True, run_id: str | None = None) -> None:
        super().__init__(dry_run=dry_run, run_id=run_id)
        self._qdrant_root = Path(os.environ.get("OPMEM_QDRANT_ROOT", "/tmp/opmem-qdrant"))
        self._memory: Any = None
        self._task_store_path: Path | None = None
        self.telemetry = LaneTelemetry(self.name)
        self._assert_pinned_models(chat_ok=True, embed_ok=True)

    def lane_id(self) -> str:
        return "mem0_oss_local_qdrant"

    def _config(self, store_path: Path, collection: str) -> dict:
        return {
            "vector_store": {
                "provider": "qdrant",
                "config": {
                    "path": str(store_path),
                    "collection_name": collection,
                    "embedding_model_dims": 1536,
                    "on_disk": True,
                },
            },
            "llm": {
                "provider": "openai",
                "config": {"model": PINNED_CHAT_MODEL, "temperature": 0},
            },
            "embedder": {
                "provider": "openai",
                "config": {"model": PINNED_EMBED_MODEL},
            },
        }

    def _load_memory(self) -> Any:
        if self._memory is not None:
            return self._memory
        from mem0 import Memory

        store_path = self._task_store_path or (self._qdrant_root / "default")
        store_path.mkdir(parents=True, exist_ok=True)
        collection = f"opmem_{self._task_name or 'task'}"
        self._memory = Memory.from_config(self._config(store_path, collection))
        return self._memory

    def _reset_remote(self) -> None:
        self._memory = None
        path = self._qdrant_root / self._run_id / self._task_name
        if path.exists():
            shutil.rmtree(path, ignore_errors=True)
        self._task_store_path = path
        self._task_store_path.mkdir(parents=True, exist_ok=True)

    def _user_id(self, actor: tuple[str, str]) -> str:
        return self.namespace(actor)

    def _remember_live(self, actor: tuple[str, str], content: str, mem_id: str) -> list[str]:
        mem = self._load_memory()
        user_id = self._user_id(actor)
        result = mem.add(
            [{"role": "user", "content": content}],
            user_id=user_id,
            infer=True,
        )
        self.telemetry.log_http("remember", 200, str(result)[:500])
        ids = []
        for item in (result or {}).get("results", []) if isinstance(result, dict) else []:
            if isinstance(item, dict) and item.get("id"):
                ids.append(item["id"])
        if not ids and isinstance(result, dict):
            for item in result.get("results", []):
                mid = item.get("memory_id") or item.get("id")
                if mid:
                    ids.append(mid)
        return ids[-1:] if ids else [mem_id]

    def _recall_live(self, actor: tuple[str, str], query: str) -> list[dict]:
        mem = self._load_memory()
        user_id = self._user_id(actor)
        raw = mem.search(query, filters={"user_id": user_id}, top_k=10)
        self.telemetry.log_http("recall", 200, str(raw)[:500])
        out = []
        for item in raw.get("results", []) if isinstance(raw, dict) else raw or []:
            if not isinstance(item, dict):
                continue
            raw = item.get("memory") or item.get("text") or item.get("content") or ""
            if isinstance(raw, dict):
                text = raw.get("data") or raw.get("text") or str(raw)
            else:
                text = str(raw)
            mid = item.get("id") or item.get("memory_id") or ""
            out.append({"id": mid, "content": text})
        return out

    def _forget_live(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        mem = self._load_memory()
        for mid in memory_ids:
            mem.delete(mid)
            self.telemetry.log_http("forget", 200, f"deleted {mid}")

    def _revise_live(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        mem = self._load_memory()
        if not memory_ids:
            return
        mem.update(memory_ids[0], data=content)
        self.telemetry.log_http("revise", 200, f"updated {memory_ids[0]}")
        if self.write_settle_seconds:
            import time

            time.sleep(self.write_settle_seconds)
