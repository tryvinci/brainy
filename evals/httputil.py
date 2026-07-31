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
    request = urllib.request.Request(
        f"{base_url.rstrip('/')}{path}",
        data=json.dumps(payload).encode("utf-8"),
        headers=auth_headers({"Content-Type": "application/json"}),
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=timeout) as response:
        return json.loads(response.read().decode("utf-8"))


def get_json(base_url: str, path: str, params: dict[str, str], timeout: float = 30) -> dict:
    filtered = {key: value for key, value in params.items() if value}
    query = urllib.parse.urlencode(filtered)
    url = f"{base_url.rstrip('/')}{path}"
    if query:
        url = f"{url}?{query}"
    request = urllib.request.Request(url, headers=auth_headers(), method="GET")
    with urllib.request.urlopen(request, timeout=timeout) as response:
        return json.loads(response.read().decode("utf-8"))
