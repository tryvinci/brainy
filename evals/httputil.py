"""Shared HTTP helpers for Brainy eval runners."""
from __future__ import annotations

import json
import os
import urllib.parse
import urllib.request


def auth_headers(extra: dict[str, str] | None = None) -> dict[str, str]:
    # CF AI Gateway / staging WAF occasionally 403s bare Python urllib UA.
    headers = {"User-Agent": "curl/8.5.0"}
    headers.update(extra or {})
    key = os.environ.get("BRAINY_API_KEY", "").strip()
    if key:
        headers["Authorization"] = f"Bearer {key}"
    return headers


def post_json(base_url: str, path: str, payload: dict, timeout: float = 30) -> dict:
    import time
    import urllib.error

    body = json.dumps(payload).encode("utf-8")
    url = f"{base_url.rstrip('/')}{path}"
    last_err: Exception | None = None
    for attempt in range(5):
        request = urllib.request.Request(
            url,
            data=body,
            headers=auth_headers({"Content-Type": "application/json"}),
            method="POST",
        )
        try:
            with urllib.request.urlopen(request, timeout=timeout) as response:
                return json.loads(response.read().decode("utf-8"))
        except urllib.error.HTTPError as exc:
            last_err = exc
            # Staging WAF intermittently 403s large conversational payloads.
            if exc.code in {403, 429, 502, 503} and attempt < 4:
                time.sleep(1.5 * (attempt + 1))
                continue
            raise
        except (TimeoutError, urllib.error.URLError) as exc:
            last_err = exc
            if attempt < 4:
                time.sleep(1.5 * (attempt + 1))
                continue
            raise
    assert last_err is not None
    raise last_err

def get_json(base_url: str, path: str, params: dict[str, str], timeout: float = 30) -> dict:
    import time
    import urllib.error

    filtered = {key: value for key, value in params.items() if value}
    query = urllib.parse.urlencode(filtered)
    url = f"{base_url.rstrip('/')}{path}"
    if query:
        url = f"{url}?{query}"
    last_err: Exception | None = None
    for attempt in range(5):
        request = urllib.request.Request(url, headers=auth_headers(), method="GET")
        try:
            with urllib.request.urlopen(request, timeout=timeout) as response:
                return json.loads(response.read().decode("utf-8"))
        except urllib.error.HTTPError as exc:
            last_err = exc
            if exc.code in {403, 429, 502, 503} and attempt < 4:
                time.sleep(1.5 * (attempt + 1))
                continue
            raise
        except (TimeoutError, urllib.error.URLError, ConnectionResetError, BrokenPipeError, OSError) as exc:
            last_err = exc
            if attempt < 4:
                time.sleep(1.5 * (attempt + 1))
                continue
            raise
    assert last_err is not None
    raise last_err