#!/usr/bin/env python3
from __future__ import annotations

import io
import json
import unittest
import urllib.error
from unittest import mock

from competitors.mem0_adapter import Mem0Adapter, observed_at_to_epoch


class _FakeResp:
    def __init__(self, payload, status: int = 200) -> None:
        self._payload = payload
        self.status = status

    def read(self) -> bytes:
        if isinstance(self._payload, bytes):
            return self._payload
        if self._payload is None:
            return b""
        return json.dumps(self._payload).encode("utf-8")

    def __enter__(self):
        return self

    def __exit__(self, *exc):
        return False


class ObservedAtTests(unittest.TestCase):
    def test_locomo_session_date(self) -> None:
        epoch = observed_at_to_epoch("1:56 pm on 8 May, 2023")
        self.assertIsNotNone(epoch)
        self.assertGreater(epoch, 1_600_000_000)

    def test_unix_and_empty(self) -> None:
        self.assertEqual(observed_at_to_epoch(1_700_000_000), 1_700_000_000)
        self.assertIsNone(observed_at_to_epoch(""))
        self.assertIsNone(observed_at_to_epoch(None))


class Mem0AdapterHTTPTests(unittest.TestCase):
    def test_search_uses_v3_then_falls_back_to_v2(self) -> None:
        calls: list[str] = []

        def fake_urlopen(req, timeout=60):
            calls.append(req.full_url)
            if req.full_url.endswith("/v3/memories/search/"):
                raise urllib.error.HTTPError(
                    req.full_url, 404, "gone", hdrs=None, fp=io.BytesIO(b"")
                )
            body = json.loads(req.data.decode("utf-8"))
            self.assertEqual(body["top_k"], 200)
            self.assertEqual(body["filters"], {"user_id": "u1"})
            return _FakeResp({"results": [{"id": "m1", "memory": "plays clarinet", "score": 0.9}]})

        adapter = Mem0Adapter(api_key="tok")
        with mock.patch("urllib.request.urlopen", fake_urlopen):
            hits = adapter.search("u1", "what instrument", top_k=200)
        self.assertEqual(len(hits), 1)
        self.assertIn("/v3/memories/search/", calls[0])
        self.assertIn("/v2/memories/search/", calls[1])
        self.assertEqual(adapter.search_path, "/v2/memories/search/")

    def test_add_passes_timestamp_and_waits_for_v3_event(self) -> None:
        calls: list[tuple[str, dict | None]] = []

        def fake_urlopen(req, timeout=60):
            payload = json.loads(req.data.decode("utf-8")) if req.data else None
            calls.append((req.get_method(), req.full_url, payload))
            if req.full_url.endswith("/v3/memories/add/") and req.get_method() == "POST":
                self.assertEqual(payload["timestamp"], 1_700_000_000)
                return _FakeResp({"event_id": "evt-1"})
            if req.full_url.endswith("/v1/event/evt-1/") and req.get_method() == "GET":
                return _FakeResp({"status": "SUCCEEDED", "results": [{"memory": "ok"}]})
            raise AssertionError(req.full_url)

        adapter = Mem0Adapter(api_key="tok")
        with mock.patch("urllib.request.urlopen", fake_urlopen):
            out = adapter.add_messages(
                "u1",
                [{"role": "user", "content": "hi"}],
                timestamp=1_700_000_000,
            )
        self.assertEqual(out["status"], "SUCCEEDED")
        self.assertTrue(calls[0][1].endswith("/v3/memories/add/"))

    def test_add_falls_back_to_v1_on_404(self) -> None:
        calls: list[str] = []

        def fake_urlopen(req, timeout=60):
            calls.append(req.full_url)
            if req.full_url.endswith("/v3/memories/add/"):
                raise urllib.error.HTTPError(
                    req.full_url, 404, "gone", hdrs=None, fp=io.BytesIO(b"")
                )
            self.assertTrue(req.full_url.endswith("/v1/memories/"))
            return _FakeResp({"results": [{"memory": "ok"}]})

        adapter = Mem0Adapter(api_key="tok")
        with mock.patch("urllib.request.urlopen", fake_urlopen):
            out = adapter.add_messages("u1", [{"role": "user", "content": "hi"}], wait_event=False)
        self.assertEqual(adapter.add_path, "/v1/memories/")
        self.assertIn("/v3/memories/add/", calls[0])
        self.assertIn("/v1/memories/", calls[1])
        self.assertIn("results", out)


if __name__ == "__main__":
    unittest.main()
