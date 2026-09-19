"""Zep Cloud lane (free tier only; $0 paid ceiling)."""

from __future__ import annotations

import os
import re

from opmem_external.protocol import OpMemLaneAdapter
from opmem_external.telemetry import LaneTelemetry

_BILLING_HINT = re.compile(r"(upgrade|payment|billing|subscribe|credit|quota|limit exceeded)", re.I)


class ZepCloudLaneAdapter(OpMemLaneAdapter):
    name = "zep"
    write_settle_seconds = 1.5

    def __init__(self, *, dry_run: bool = True, run_id: str | None = None) -> None:
        super().__init__(dry_run=dry_run, run_id=run_id)
        self._api_key = os.environ.get("ZEP_API_KEY", "")
        self.telemetry = LaneTelemetry(self.name)
        self._session_id: str | None = None
        self._user_id: str | None = None
        self._assert_pinned_models(chat_ok=False, embed_ok=False)
        self._record_model_constraint(
            "Zep Cloud memory path does not accept OpenAI chat/embed model pins; "
            "stop before substituting — record scores only under Zep-native models."
        )

    def lane_id(self) -> str:
        return "zep_cloud_free_tier"

    def _client(self):
        from zep_cloud import Zep

        return Zep(api_key=self._api_key)

    def _abort_if_billing(self, text: str) -> None:
        if _BILLING_HINT.search(text or ""):
            raise RuntimeError("Zep hard-stop: billing/upgrade/credit prompt detected")

    def verify_free_tier(self) -> None:
        client = self._client()
        # Lightweight call; abort if response hints at paid upgrade.
        try:
            client.user.list_ordered(page_size=1, page_number=1)
        except Exception as exc:  # noqa: BLE001
            self._abort_if_billing(str(exc))
            raise

    def _reset_remote(self) -> None:
        client = self._client()
        self.verify_free_tier()
        self._user_id = self.namespace(("t1", "u1"))
        self._session_id = f"{self._user_id}-session"
        try:
            client.user.add(user_id=self._user_id, first_name="OpMem", last_name="Lane")
        except Exception:
            pass
        try:
            client.memory.add_session(session_id=self._session_id, user_id=self._user_id)
        except Exception:
            pass

    def _remember_live(self, actor: tuple[str, str], content: str, mem_id: str) -> list[str]:
        from zep_cloud.types import Message

        client = self._client()
        sid = self._session_for(actor)
        client.memory.add(
            session_id=sid,
            messages=[Message(content=content, role_type="user", role="opmem-user")],
            return_context=False,
        )
        self.telemetry.log_http("remember", 200, "zep memory.add")
        return [mem_id]

    def _session_for(self, actor: tuple[str, str]) -> str:
        if self._session_id is None:
            self._reset_remote()
        return self._session_id or self.namespace(actor)

    def _recall_live(self, actor: tuple[str, str], query: str) -> list[dict]:
        client = self._client()
        sid = self._session_for(actor)
        ctx = client.memory.get(session_id=sid, lastn=5)
        text = getattr(ctx, "context", None) or str(ctx)
        self._abort_if_billing(text)
        self.telemetry.log_http("recall", 200, text[:500])
        if not text:
            return []
        return [{"id": "zep-context", "content": text}]

    def _forget_live(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        # Zep session graph: no direct id delete in OpMem mapping; best-effort no-op with log.
        self.telemetry.log_http("forget", 200, f"zep forget noop for {memory_ids}")

    def _revise_live(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        from zep_cloud.types import Message

        client = self._client()
        sid = self._session_for(actor)
        client.memory.add(
            session_id=sid,
            messages=[Message(content=content, role_type="user", role="opmem-user")],
            return_context=False,
        )
        self.telemetry.log_http("revise", 200, "zep memory.add revise")

    def _simulate_remember(self, content: str) -> None:
        return

    def _simulate_recall(self, query: str) -> None:
        return

    def _simulate_revise(self, content: str) -> None:
        return
