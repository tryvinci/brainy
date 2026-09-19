"""Letta self-hosted App Server lane (not legacy Docker API or Letta Cloud paid)."""

from __future__ import annotations

import os

from opmem_external.pins import PINNED_CHAT_MODEL, PINNED_EMBED_MODEL
from opmem_external.protocol import OpMemLaneAdapter


class LettaAppServerLaneAdapter(OpMemLaneAdapter):
    name = "letta"
    write_settle_seconds = 1.0

    def __init__(self, *, dry_run: bool = True, run_id: str | None = None) -> None:
        super().__init__(dry_run=dry_run, run_id=run_id)
        self._base_url = os.environ.get("LETTA_BASE_URL", "http://127.0.0.1:8283").rstrip("/")
        self._assert_pinned_models(chat_ok=True, embed_ok=True)

    def lane_id(self) -> str:
        return "letta_app_server_self_hosted"

    def _reset_remote(self) -> None:
        raise NotImplementedError("live Letta reset blocked until worksheet approval")

    def _remember_live(self, actor: tuple[str, str], content: str, mem_id: str) -> list[str]:
        raise NotImplementedError(
            f"live Letta remember blocked; pin chat={PINNED_CHAT_MODEL} embed={PINNED_EMBED_MODEL}"
        )

    def _recall_live(self, actor: tuple[str, str], query: str) -> list[dict]:
        raise NotImplementedError("live Letta recall blocked until worksheet approval")

    def _forget_live(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        raise NotImplementedError("live Letta forget blocked until worksheet approval")

    def _revise_live(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        raise NotImplementedError("live Letta revise blocked until worksheet approval")
