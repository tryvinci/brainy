"""Mem0 OSS lane: Python package + local Qdrant (no Platform API key)."""

from __future__ import annotations

import os

from opmem_external.pins import PINNED_CHAT_MODEL, PINNED_EMBED_MODEL
from opmem_external.protocol import OpMemLaneAdapter


class Mem0OSSLaneAdapter(OpMemLaneAdapter):
    name = "mem0-oss"
    write_settle_seconds = 2.0

    def __init__(self, *, dry_run: bool = True, run_id: str | None = None) -> None:
        super().__init__(dry_run=dry_run, run_id=run_id)
        self._qdrant_url = os.environ.get("QDRANT_URL", "http://127.0.0.1:6333")
        self._assert_pinned_models(chat_ok=True, embed_ok=True)

    def lane_id(self) -> str:
        return "mem0_oss_local_qdrant"

    def _reset_remote(self) -> None:
        raise NotImplementedError("live mem0-oss reset requires Phase 2 credentials check")

    def _remember_live(self, actor: tuple[str, str], content: str, mem_id: str) -> list[str]:
        raise NotImplementedError(
            "live mem0-oss remember blocked until dry-run worksheet approved; "
            f"use model={PINNED_CHAT_MODEL} embed={PINNED_EMBED_MODEL}"
        )

    def _recall_live(self, actor: tuple[str, str], query: str) -> list[dict]:
        raise NotImplementedError("live mem0-oss recall blocked until worksheet approval")

    def _forget_live(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        raise NotImplementedError("live mem0-oss forget blocked until worksheet approval")

    def _revise_live(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        raise NotImplementedError("live mem0-oss revise blocked until worksheet approval")
