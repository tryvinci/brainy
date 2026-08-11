from __future__ import annotations

import os
import pathlib
import re
import sys
import time

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))

from httputil import get_json, post_json  # noqa: E402


class BrainyBackend:
    """HTTP backend against a live Brainy API (local or staging)."""

    name = "brainy"

    def __init__(
        self,
        base_url: str,
        tenant_prefix: str = "public",
        *,
        async_ingest: bool = False,
        async_timeout_s: float = 120.0,
        async_poll_s: float = 1.0,
        publish_mode: bool | None = None,
        max_ingest_bytes: int = 48_000,
    ) -> None:
        self.base_url = base_url.rstrip("/")
        self.tenant_prefix = tenant_prefix
        self.async_ingest = async_ingest
        self.async_timeout_s = async_timeout_s
        self.async_poll_s = async_poll_s
        if publish_mode is None:
            publish_mode = os.environ.get("BRAINY_EVAL_PUBLISH", "").lower() in {
                "1",
                "true",
                "yes",
                "on",
            }
        self.publish_mode = bool(publish_mode)
        self.max_ingest_bytes = max(4_000, int(max_ingest_bytes))
        self._pending_jobs: dict[str, list[str]] = {}

    def _tenant(self, user_id: str) -> str:
        return f"{self.tenant_prefix}-{user_id}"

    def remember(self, user_id: str, content: str) -> list[str]:
        """Ingest a single text blob as one message (legacy helper)."""
        return self.remember_messages(user_id, [{"role": "user", "content": content}])

    def remember_messages(
        self,
        user_id: str,
        messages: list[dict],
        metadata: dict | None = None,
        *,
        wait: bool | None = None,
    ) -> list[str]:
        """Ingest one or more messages (atomic turns preferred per product docs).

        When ``async_ingest`` is enabled, posts to ``/ingest/async`` (worker /
        provider extract path) and optionally waits until jobs complete.
        Oversized batches are split to avoid HTTP 400 payload limits.
        """
        out: list[str] = []
        for batch in _iter_ingest_batches(messages, self.max_ingest_bytes):
            out.extend(self._remember_messages_once(user_id, batch, metadata, wait=wait))
        return out

    def _remember_messages_once(
        self,
        user_id: str,
        messages: list[dict],
        metadata: dict | None,
        *,
        wait: bool | None,
    ) -> list[str]:
        tenant = self._tenant(user_id)
        payload: dict = {
            "tenant_id": tenant,
            "subject_id": user_id,
            "source_type": "conversation",
            "messages": messages,
        }
        if metadata:
            payload["metadata"] = metadata

        if self.async_ingest:
            response = post_json(self.base_url, "/ingest/async", payload, timeout=60)
            should_wait = self.async_ingest if wait is None else wait
            job_id = str(response.get("job_id") or "")
            if job_id:
                self._pending_jobs.setdefault(user_id, []).append(job_id)
            if should_wait:
                if job_id:
                    self.wait_until_jobs_done(user_id, job_ids=[job_id])
                elif self.publish_mode:
                    raise RuntimeError(
                        "publish mode requires job_id from /ingest/async; got none"
                    )
                else:
                    probe = _probe_token(messages)
                    if probe:
                        self.wait_until_searchable(user_id, probe, min_results=1)
            return [job_id] if job_id else []

        response = post_json(self.base_url, "/ingest", payload, timeout=120)
        return [m["memory_id"] for m in response.get("memories", [])]

    def wait_until_jobs_done(
        self,
        user_id: str,
        job_ids: list[str] | None = None,
        *,
        timeout_s: float | None = None,
    ) -> None:
        """Block until extraction jobs complete (prefer over search settle).

        Polls ``GET /jobs/{id}`` for tracked ids and ``GET /jobs/status`` until
        open==0. In publish mode, missing job APIs or failed jobs raise
        (fail-closed) instead of silently falling back to search settle.
        """
        tenant = self._tenant(user_id)
        tracked = [j for j in (job_ids or list(self._pending_jobs.get(user_id, []))) if j]
        deadline = time.time() + (self.async_timeout_s if timeout_s is None else timeout_s)
        saw_status_api = False
        terminal_failed: list[str] = []

        while time.time() < deadline:
            unfinished: list[str] = []
            for job_id in tracked:
                try:
                    info = get_json(self.base_url, f"/jobs/{job_id}", {}, timeout=30)
                except Exception as exc:
                    if self.publish_mode:
                        raise RuntimeError(
                            f"publish mode: job status unavailable for {job_id}: {exc}"
                        ) from exc
                    unfinished.append(job_id)
                    continue
                status = str(info.get("status") or "").lower()
                if status in {"completed"}:
                    continue
                if status in {"failed"}:
                    terminal_failed.append(job_id)
                    continue
                unfinished.append(job_id)
            tracked = unfinished

            try:
                counts = get_json(
                    self.base_url,
                    "/jobs/status",
                    {"tenant_id": tenant, "subject_id": user_id},
                    timeout=30,
                )
                saw_status_api = True
            except Exception as exc:
                if self.publish_mode:
                    raise RuntimeError(
                        f"publish mode: /jobs/status unavailable: {exc}"
                    ) from exc
                counts = {}

            open_n = int(counts.get("open") or 0)
            if not tracked and (not saw_status_api or open_n == 0):
                if terminal_failed and self.publish_mode:
                    raise RuntimeError(
                        f"publish mode: extraction jobs failed: {terminal_failed}"
                    )
                # Clear completed tracked ids for this user.
                if user_id in self._pending_jobs:
                    done = set(job_ids or []) | set(terminal_failed)
                    if not job_ids:
                        self._pending_jobs[user_id] = []
                    else:
                        self._pending_jobs[user_id] = [
                            j for j in self._pending_jobs[user_id] if j not in done
                        ]
                return

            time.sleep(max(self.async_poll_s, 1.0))

        detail = f"tracked={tracked!r} failed={terminal_failed!r}"
        raise TimeoutError(
            f"jobs not done within timeout for subject={user_id} ({detail}). "
            "Is the worker running?"
        )

    def wait_until_searchable(
        self,
        user_id: str,
        query: str,
        *,
        min_results: int = 1,
        timeout_s: float | None = None,
    ) -> list[dict]:
        """Poll search until worker extract has made memories findable."""
        return self.wait_until_any_searchable(
            user_id,
            [query],
            min_results=min_results,
            timeout_s=timeout_s,
        )

    def wait_until_any_searchable(
        self,
        user_id: str,
        queries: list[str],
        *,
        min_results: int = 1,
        timeout_s: float | None = None,
        settle_polls: int = 6,
    ) -> list[dict]:
        """Poll until probe hits, then wait for result counts to stabilize.

        Async jobs are FIFO; the first searchable batch must not unblock QA
        while later batches are still pending. Caps probes and tolerates
        transient search timeouts under staging load.
        """
        seen: set[str] = set()
        probes: list[str] = []
        for q in queries:
            q = (q or "").strip()
            if not q or q in seen:
                continue
            seen.add(q)
            probes.append(q)
            if len(probes) >= 3:
                break
        if not probes:
            probes = ["conversation"]
        deadline = time.time() + (self.async_timeout_s if timeout_s is None else timeout_s)
        last: list[dict] = []
        last_query = probes[0]
        saw_hit = False
        stable = 0
        last_n = -1
        while time.time() < deadline:
            best: list[dict] = []
            for query in probes:
                last_query = query
                try:
                    last, _ = self.recall(user_id, query, top_k=max(min_results, 10), timeout=120)
                except Exception:
                    # Staging can time out under embed load; keep polling.
                    time.sleep(self.async_poll_s)
                    continue
                if len(last) > len(best):
                    best = last
                if len(last) >= min_results:
                    saw_hit = True
            n = len(best)
            if saw_hit:
                if n == last_n and n > 0:
                    stable += 1
                    if stable >= settle_polls:
                        return best or last
                else:
                    stable = 0
                last_n = n
            time.sleep(max(self.async_poll_s, 2.0))
        if saw_hit and last:
            return last
        raise TimeoutError(
            f"async extract not searchable within timeout "
            f"(probes={probes!r}, last_query={last_query!r}, got {len(last)} results). "
            "Is the worker running with provider/deterministic extract?"
        )

    def recall(
        self, user_id: str, query: str, top_k: int = 10, timeout: float = 120
    ) -> tuple[list[dict], float]:
        tenant = self._tenant(user_id)
        started = time.perf_counter()
        body = get_json(
            self.base_url,
            "/memories/search",
            {
                "tenant_id": tenant,
                "subject_id": user_id,
                "q": query,
            },
            timeout=timeout,
        )
        latency_ms = (time.perf_counter() - started) * 1000.0
        results = []
        for item in body.get("results", [])[:top_k]:
            row = {
                "id": item.get("memory_id", ""),
                "content": item.get("content", ""),
                "score": float(item.get("score") or 0),
            }
            if item.get("observed_at"):
                row["observed_at"] = item.get("observed_at")
            results.append(row)
        return results, latency_ms


_STOP = {
    "the",
    "a",
    "an",
    "and",
    "or",
    "to",
    "of",
    "in",
    "on",
    "for",
    "is",
    "it",
    "i",
    "you",
    "we",
    "they",
    "he",
    "she",
    "was",
    "were",
    "be",
    "with",
    "that",
    "this",
    "at",
    "as",
    "from",
    "by",
    "my",
    "me",
    "our",
    "your",
}


def _message_bytes(message: dict) -> int:
    content = str(message.get("content") or "")
    return len(content.encode("utf-8"))


def _iter_ingest_batches(messages: list[dict], max_bytes: int) -> list[list[dict]]:
    """Split messages into payload-safe batches (avoids HTTP 400 on huge haystacks)."""
    if not messages:
        return []
    batches: list[list[dict]] = []
    cur: list[dict] = []
    cur_bytes = 0
    soft_count = 4  # keep small even when messages are short
    for msg in messages:
        raw = dict(msg)
        content = str(raw.get("content") or "")
        encoded = content.encode("utf-8")
        if len(encoded) > max_bytes:
            # Hard-split a single oversized turn into contiguous slices.
            step = max(1, max_bytes - 64)
            for i in range(0, len(encoded), step):
                piece = encoded[i : i + step].decode("utf-8", errors="ignore").strip()
                if not piece:
                    continue
                part = dict(raw)
                part["content"] = piece
                batches.append([part])
            continue
        n = len(encoded)
        if cur and (cur_bytes + n > max_bytes or len(cur) >= soft_count):
            batches.append(cur)
            cur = []
            cur_bytes = 0
        cur.append(raw)
        cur_bytes += n
    if cur:
        batches.append(cur)
    return batches


def _probe_token(messages: list[dict]) -> str:
    """Pick a distinctive token from the last message for readiness polling."""
    for message in reversed(messages):
        content = str(message.get("content") or "")
        # Prefer year / date-like / longer tokens.
        candidates = re.findall(r"[A-Za-z0-9]{3,}", content)
        ranked = sorted(
            (c for c in candidates if c.lower() not in _STOP),
            key=lambda c: (c.isdigit(), len(c)),
            reverse=True,
        )
        if ranked:
            return ranked[0]
    return ""
