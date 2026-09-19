"""LangMem lane: local SDK + Postgres/pgvector (no InMemoryStore for measured runs)."""

from __future__ import annotations

import os

from opmem_external.pins import PINNED_CHAT_MODEL, PINNED_EMBED_MODEL
from opmem_external.protocol import OpMemLaneAdapter


class LangMemLaneAdapter(OpMemLaneAdapter):
    name = "langmem"
    write_settle_seconds = 0.5

    def __init__(self, *, dry_run: bool = True, run_id: str | None = None) -> None:
        super().__init__(dry_run=dry_run, run_id=run_id)
        self._database_url = os.environ.get("DATABASE_URL", "")
        if not dry_run and "InMemoryStore" in os.environ.get("LANGMEM_STORE", ""):
            self._record_model_constraint("InMemoryStore is forbidden for measured OpMem runs")
        self._assert_pinned_models(chat_ok=True, embed_ok=True)

    def lane_id(self) -> str:
        return "langmem_postgres_pgvector"

    def _reset_remote(self) -> None:
        raise NotImplementedError("live LangMem reset blocked until worksheet approval")

    def _remember_live(self, actor: tuple[str, str], content: str, mem_id: str) -> list[str]:
        raise NotImplementedError(
            f"live LangMem remember blocked; pin chat={PINNED_CHAT_MODEL} embed={PINNED_EMBED_MODEL}"
        )

    def _recall_live(self, actor: tuple[str, str], query: str) -> list[dict]:
        raise NotImplementedError("live LangMem recall blocked until worksheet approval")

    def _forget_live(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        raise NotImplementedError("live LangMem forget blocked until worksheet approval")

    def _revise_live(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        raise NotImplementedError("live LangMem revise blocked until worksheet approval")
