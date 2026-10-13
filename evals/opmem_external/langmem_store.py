"""LangMem lane: local SDK + Postgres/pgvector (no InMemoryStore for measured runs)."""

from __future__ import annotations

import json
import os
import uuid
from typing import Any

from langchain_core.messages import HumanMessage

from opmem_external.pins import PINNED_CHAT_MODEL, PINNED_EMBED_MODEL
from opmem_external.protocol import OpMemLaneAdapter
from opmem_external.telemetry import LaneTelemetry


def _stringify_memory_value(val: Any) -> str:
    if val is None:
        return ""
    if isinstance(val, str):
        return val
    if isinstance(val, dict):
        for key in ("content", "text", "memory"):
            inner = val.get(key)
            if isinstance(inner, str):
                return inner
            if inner is not None:
                return _stringify_memory_value(inner)
        return json.dumps(val, ensure_ascii=False)
    content = getattr(val, "content", None)
    if isinstance(content, str):
        return content
    return str(val)


class LangMemLaneAdapter(OpMemLaneAdapter):
    name = "langmem"
    write_settle_seconds = 0.5

    def __init__(self, *, dry_run: bool = True, run_id: str | None = None) -> None:
        super().__init__(dry_run=dry_run, run_id=run_id)
        self._database_url = os.environ.get(
            "DATABASE_URL",
            "postgresql://opmem:opmem@127.0.0.1:5432/opmem_langmem",
        )
        if not dry_run and "InMemoryStore" in os.environ.get("LANGMEM_STORE", ""):
            self._record_model_constraint("InMemoryStore is forbidden for measured OpMem runs")
        self._store_ctx: Any = None
        self._store: Any = None
        self._manager: Any = None
        self.telemetry = LaneTelemetry(self.name)
        self._assert_pinned_models(chat_ok=True, embed_ok=True)

    def lane_id(self) -> str:
        return "langmem_postgres_pgvector"

    def _user_id(self, actor: tuple[str, str]) -> str:
        return self.namespace(actor)

    def _ensure_store(self) -> Any:
        if self._store is not None:
            return self._store
        from langgraph.store.postgres import PostgresStore

        self._store_ctx = PostgresStore.from_conn_string(
            self._database_url,
            index={
                "dims": 1536,
                "embed": f"openai:{PINNED_EMBED_MODEL}",
            },
        )
        self._store = self._store_ctx.__enter__()
        self._store.setup()
        return self._store

    def _ensure_manager(self) -> Any:
        if self._manager is not None:
            return self._manager
        from langmem import create_memory_store_manager

        store = self._ensure_store()
        self._manager = create_memory_store_manager(
            f"openai:{PINNED_CHAT_MODEL}",
            namespace=("opmem", "{langgraph_user_id}"),
            store=store,
            enable_deletes=True,
            instructions=(
                "Extract durable user memory facts only. Do not answer questions; "
                "memory operations for OpMem benchmark lane."
            ),
        )
        return self._manager

    def _config(self, actor: tuple[str, str]) -> dict:
        return {"configurable": {"langgraph_user_id": self._user_id(actor)}}

    def _reset_remote(self) -> None:
        self._manager = None
        if self._store is not None:
            user_prefix = self._user_id(("t1", "u1"))[:8]
            try:
                import psycopg

                with psycopg.connect(self._database_url) as conn:
                    with conn.cursor() as cur:
                        cur.execute(
                            "DELETE FROM store WHERE prefix LIKE %s",
                            (f"opmem.%",),
                        )
                    conn.commit()
            except Exception as exc:
                self.telemetry.log_http("reset", 500, f"reset partial: {exc}")
        self._memory_table.clear()

    def _remember_live(self, actor: tuple[str, str], content: str, mem_id: str) -> list[str]:
        manager = self._ensure_manager()
        cfg = self._config(actor)
        manager.invoke({"messages": [HumanMessage(content=content)]}, cfg)
        self.telemetry.log_http("remember", 200, "langmem manager.invoke")
        return [mem_id]

    def _recall_live(self, actor: tuple[str, str], query: str) -> list[dict]:
        store = self._ensure_store()
        user = self._user_id(actor)
        namespace = ("opmem", user)
        hits = store.search(namespace, query=query, limit=10)
        self.telemetry.log_http("recall", 200, str(hits)[:500])
        out: list[dict] = []
        for item in hits:
            val = getattr(item, "value", item)
            text = _stringify_memory_value(val)
            key = getattr(item, "key", None) or str(uuid.uuid4())
            out.append({"id": key, "content": text})
        return out

    def _forget_live(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        store = self._ensure_store()
        namespace = ("opmem", self._user_id(actor))
        for mid in memory_ids:
            try:
                store.delete(namespace, mid)
            except Exception:
                pass
        self.telemetry.log_http("forget", 200, f"deleted {memory_ids}")

    def _revise_live(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        manager = self._ensure_manager()
        cfg = self._config(actor)
        manager.invoke(
            {
                "messages": [
                    HumanMessage(
                        content=f"Update memory to reflect: {content}"
                    )
                ]
            },
            cfg,
        )
        self.telemetry.log_http("revise", 200, "langmem revise invoke")
        if self.write_settle_seconds:
            import time

            time.sleep(self.write_settle_seconds)
