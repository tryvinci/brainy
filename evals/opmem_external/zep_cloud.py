"""Zep Cloud lane (free tier only; $0 paid ceiling)."""

from __future__ import annotations

import os

from opmem_external.protocol import OpMemLaneAdapter


class ZepCloudLaneAdapter(OpMemLaneAdapter):
    name = "zep"
    write_settle_seconds = 1.5

    def __init__(self, *, dry_run: bool = True, run_id: str | None = None) -> None:
        super().__init__(dry_run=dry_run, run_id=run_id)
        self._base_url = os.environ.get("ZEP_API_URL", "https://api.getzep.com").rstrip("/")
        # Zep Cloud memory API; CE/self-hosted Graphiti is out of scope for this pin.
        self._assert_pinned_models(chat_ok=False, embed_ok=False)
        self._record_model_constraint(
            "Zep Cloud memory path does not accept OpenAI chat/embed model pins; "
            "stop before substituting — record scores only under Zep-native models."
        )

    def lane_id(self) -> str:
        return "zep_cloud_free_tier"

    def _simulate_remember(self, content: str) -> None:
        # Zep Cloud billing is not OpenAI token-metered; dry-run counts ops only.
        return

    def _simulate_recall(self, query: str) -> None:
        return

    def _simulate_revise(self, content: str) -> None:
        return

    def _reset_remote(self) -> None:
        raise NotImplementedError("live Zep blocked: $0 paid ceiling; verify free credits first")

    def _remember_live(self, actor: tuple[str, str], content: str, mem_id: str) -> list[str]:
        raise NotImplementedError("live Zep remember blocked until free-tier credits confirmed")

    def _recall_live(self, actor: tuple[str, str], query: str) -> list[dict]:
        raise NotImplementedError("live Zep recall blocked until free-tier credits confirmed")

    def _forget_live(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        raise NotImplementedError("live Zep forget blocked until free-tier credits confirmed")

    def _revise_live(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        raise NotImplementedError("live Zep revise blocked until free-tier credits confirmed")
