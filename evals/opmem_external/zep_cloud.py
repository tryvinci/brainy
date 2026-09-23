"""Zep Cloud lane (free tier only; $0 paid ceiling). SDK 3.x thread + graph API."""

from __future__ import annotations

import os
import re

from opmem_external.protocol import OpMemLaneAdapter, _overlap
from opmem_external.telemetry import LaneTelemetry

_BILLING_HINT = re.compile(r"(upgrade|payment|billing|subscribe|credit|quota|limit exceeded)", re.I)


class ZepCloudLaneAdapter(OpMemLaneAdapter):
    name = "zep"
    write_settle_seconds = 1.5

    def __init__(self, *, dry_run: bool = True, run_id: str | None = None) -> None:
        super().__init__(dry_run=dry_run, run_id=run_id)
        self._api_key = os.environ.get("ZEP_API_KEY", "")
        self.telemetry = LaneTelemetry(self.name)
        self._thread_id: str | None = None
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
        try:
            client.user.list_ordered(page_size=1, page_number=1)
        except Exception as exc:  # noqa: BLE001
            self._abort_if_billing(str(exc))
            raise

    def _reset_remote(self) -> None:
        client = self._client()
        self.verify_free_tier()
        self._user_id = self.namespace(("t1", "u1"))
        self._thread_id = f"{self._user_id}-thread"
        if self._thread_id:
            try:
                client.thread.delete(self._thread_id)
            except Exception:
                pass
        try:
            client.user.add(user_id=self._user_id, first_name="OpMem", last_name="Lane")
        except Exception:
            pass
        client.thread.create(thread_id=self._thread_id, user_id=self._user_id)

    def _remember_live(self, actor: tuple[str, str], content: str, mem_id: str) -> list[str]:
        from zep_cloud.types import Message

        client = self._client()
        tid = self._thread_for(actor)
        client.thread.add_messages(
            thread_id=tid,
            messages=[Message(content=content, role="user")],
            return_context=False,
        )
        self.telemetry.log_http("remember", 200, "zep thread.add_messages")
        if self.write_settle_seconds:
            import time

            time.sleep(self.write_settle_seconds)
        return [mem_id]

    def _thread_for(self, actor: tuple[str, str]) -> str:
        if self._thread_id is None:
            self._reset_remote()
        return self._thread_id or self.namespace(actor)

    def _recall_live(self, actor: tuple[str, str], query: str) -> list[dict]:
        client = self._client()
        uid = self._user_id or self.namespace(actor)
        tid = self._thread_for(actor)
        out: list[dict] = []
        seen_content: set[str] = set()

        msg_list = client.thread.get(tid, lastn=100)
        for idx, msg in enumerate(msg_list.messages or []):
            text = (msg.content or "").strip()
            if not text or _overlap(query, text) <= 0:
                continue
            key = text.lower()
            if key in seen_content:
                continue
            seen_content.add(key)
            out.append({"id": f"zep-msg-{idx}", "content": text})

        try:
            results = client.graph.search(query=query, user_id=uid, limit=10)
            self._abort_if_billing(str(results))
            for ep in results.episodes or []:
                text = getattr(ep, "content", None) or getattr(ep, "name", None) or ""
                text = str(text).strip()
                if not text or text.lower() in seen_content:
                    continue
                if _overlap(query, text) <= 0:
                    continue
                seen_content.add(text.lower())
                eid = getattr(ep, "uuid_", None) or getattr(ep, "uuid", "zep-episode")
                out.append({"id": str(eid), "content": text})
            for edge in results.edges or []:
                text = getattr(edge, "fact", None) or getattr(edge, "name", None) or ""
                text = str(text).strip()
                if not text or text.lower() in seen_content:
                    continue
                if _overlap(query, text) <= 0:
                    continue
                seen_content.add(text.lower())
                eid = getattr(edge, "uuid_", None) or getattr(edge, "uuid", "zep-edge")
                out.append({"id": str(eid), "content": text})
        except Exception as exc:
            self._abort_if_billing(str(exc))

        out.sort(key=lambda row: -_overlap(query, row["content"]))
        self.telemetry.log_http("recall", 200, str(out)[:500])
        return out

    def _forget_live(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        self.telemetry.log_http("forget", 200, f"zep forget noop for {memory_ids}")

    def _revise_live(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        from zep_cloud.types import Message

        client = self._client()
        tid = self._thread_for(actor)
        client.thread.add_messages(
            thread_id=tid,
            messages=[Message(content=content, role="user")],
            return_context=False,
        )
        self.telemetry.log_http("revise", 200, "zep thread.add_messages revise")
        if self.write_settle_seconds:
            import time

            time.sleep(self.write_settle_seconds)

    def _simulate_remember(self, content: str) -> None:
        return

    def _simulate_recall(self, query: str) -> None:
        return

    def _simulate_revise(self, content: str) -> None:
        return
