"""Letta self-hosted App Server lane (pip server 0.11.x + archival block mapping)."""

from __future__ import annotations

import os
import re
from typing import Any

from opmem_external.pins import PINNED_CHAT_MODEL, PINNED_EMBED_MODEL
from opmem_external.protocol import OpMemLaneAdapter, _overlap
from opmem_external.telemetry import LaneTelemetry

_BLOCK_LABEL = "opmem_facts"
_LINE_RE = re.compile(r"^\[(?P<id>[^\]]+)\]\s*(?P<body>.*)$")


class LettaAppServerLaneAdapter(OpMemLaneAdapter):
    name = "letta"
    write_settle_seconds = 1.0

    def __init__(self, *, dry_run: bool = True, run_id: str | None = None) -> None:
        super().__init__(dry_run=dry_run, run_id=run_id)
        self._base_url = os.environ.get("LETTA_BASE_URL", "http://127.0.0.1:8283").rstrip("/")
        self._client: Any = None
        self._agent_id: str | None = None
        self.telemetry = LaneTelemetry(self.name)
        self._assert_pinned_models(chat_ok=True, embed_ok=True)
        self._record_model_constraint(
            "Letta lane maps remember/revise/forget to agents.blocks.modify on "
            f"'{_BLOCK_LABEL}' (line-per-memory). Recall uses lexical overlap, not agent chat."
        )
        if not dry_run and not os.environ.get("LETTA_APP_SERVER_TOKEN", "").strip():
            # Local `letta server --no-secure` (pinned 0.11.7) does not require a token.
            if not self._local_insecure_server():
                self._missing_credentials.append("LETTA_APP_SERVER_TOKEN")

    def lane_id(self) -> str:
        return "letta_app_server_self_hosted"

    def _local_insecure_server(self) -> bool:
        host = self._base_url
        return host.startswith("http://127.0.0.1") or host.startswith("http://localhost")

    def _client_instance(self) -> Any:
        if self._client is not None:
            return self._client
        from letta_client import Letta

        token = os.environ.get("LETTA_APP_SERVER_TOKEN", "").strip()
        kwargs: dict[str, Any] = {"base_url": self._base_url}
        if token:
            kwargs["api_key"] = token
        self._client = Letta(**kwargs)
        return self._client

    def _model(self) -> str:
        return f"openai/{PINNED_CHAT_MODEL}"

    def _embedding(self) -> str:
        return f"openai/{PINNED_EMBED_MODEL}"

    def _reset_remote(self) -> None:
        client = self._client_instance()
        if self._agent_id:
            try:
                client.agents.delete(agent_id=self._agent_id)
            except Exception:
                pass
        self._agent_id = None
        ns = self.namespace(("t1", "u1"))
        agent = client.agents.create(
            name=f"opmem-{ns}",
            memory_blocks=[
                {"label": _BLOCK_LABEL, "value": ""},
                {"label": "persona", "value": "OpMem external lane; memory block only (no chat answers)."},
            ],
            model=self._model(),
            embedding=self._embedding(),
        )
        self._agent_id = agent.id
        self.telemetry.log_http("reset", 200, f"agent {self._agent_id}")

    def _read_lines(self) -> dict[str, str]:
        client = self._client_instance()
        if not self._agent_id:
            self._reset_remote()
        block = client.agents.blocks.retrieve(agent_id=self._agent_id, block_label=_BLOCK_LABEL)
        text = block.value or ""
        out: dict[str, str] = {}
        for line in text.splitlines():
            line = line.strip()
            if not line:
                continue
            m = _LINE_RE.match(line)
            if m:
                out[m.group("id")] = m.group("body")
        return out

    def _write_lines(self, lines: dict[str, str]) -> None:
        client = self._client_instance()
        if not self._agent_id:
            self._reset_remote()
        body = "\n".join(f"[{mid}] {txt}" for mid, txt in sorted(lines.items()))
        client.agents.blocks.modify(agent_id=self._agent_id, block_label=_BLOCK_LABEL, value=body)
        self.telemetry.log_http("blocks.modify", 200, f"{len(lines)} facts")

    def _remember_live(self, actor: tuple[str, str], content: str, mem_id: str) -> list[str]:
        lines = self._read_lines()
        lines[mem_id] = content
        self._write_lines(lines)
        return [mem_id]

    def _recall_live(self, actor: tuple[str, str], query: str) -> list[dict]:
        lines = self._read_lines()
        scored = sorted(lines.items(), key=lambda item: -_overlap(query, item[1]))
        out = [{"id": mid, "content": txt} for mid, txt in scored if _overlap(query, txt) > 0]
        self.telemetry.log_http("recall", 200, str(out)[:500])
        return out

    def _forget_live(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        lines = self._read_lines()
        for mid in memory_ids:
            lines.pop(mid, None)
        self._write_lines(lines)
        self.telemetry.log_http("forget", 200, f"removed {memory_ids}")

    def _revise_live(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        if not memory_ids:
            return
        lines = self._read_lines()
        lines[memory_ids[0]] = content
        self._write_lines(lines)
        self.telemetry.log_http("revise", 200, f"updated {memory_ids[0]}")
        if self.write_settle_seconds:
            import time

            time.sleep(self.write_settle_seconds)
