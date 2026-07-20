"""Local OpenAI-compatible /v1/embeddings server (CPU, open-source model).

Purpose: give Brainy real dense embeddings for local evaluation when a hosted
embeddings endpoint is unavailable (e.g. a gateway that blocks /embeddings).
This is dev/eval tooling only — production should point BRAINY_EMBEDDING_BASE_URL
at a real managed embeddings provider.

Usage:
    pip install fastembed
    python evals/tools/local_embeddings_server.py --model BAAI/bge-base-en-v1.5 --port 8099

    # then run Brainy (API + worker) with:
    export BRAINY_EMBEDDING_BASE_URL=http://127.0.0.1:8099/v1
    export BRAINY_EMBEDDING_API_KEY=local
    export BRAINY_EMBEDDING_MODEL=bge-base-en-v1.5   # any non-empty id
    # entity ranking auto-enables when an embedding model is configured

Notes:
- Brainy stores provider-dim vectors in the float[] path; the pgvector(128)
  column is only used for the 128-d local hash embedder.
- Per-query similarity calibration (internal/memory/hybrid.go) makes ranking
  robust to a model's baseline cosine, so any embedding model works.
"""
from __future__ import annotations

import argparse
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


def build_handler(model):
    class Handler(BaseHTTPRequestHandler):
        def log_message(self, *_):
            pass

        def do_POST(self):
            if not self.path.rstrip("/").endswith("/embeddings"):
                self.send_response(404)
                self.end_headers()
                return
            length = int(self.headers.get("Content-Length", 0))
            body = json.loads(self.rfile.read(length) or b"{}")
            raw = body.get("input", "")
            texts = raw if isinstance(raw, list) else [raw]
            vecs = [list(map(float, v)) for v in model.embed([str(t) for t in texts])]
            data = [
                {"object": "embedding", "index": i, "embedding": v}
                for i, v in enumerate(vecs)
            ]
            payload = json.dumps({"object": "list", "data": data}).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(payload)))
            self.end_headers()
            self.wfile.write(payload)

    return Handler


def main() -> None:
    parser = argparse.ArgumentParser(description="Local OpenAI-compatible embeddings server")
    parser.add_argument("--model", default="BAAI/bge-base-en-v1.5")
    parser.add_argument("--port", type=int, default=8099)
    args = parser.parse_args()

    from fastembed import TextEmbedding

    model = TextEmbedding(model_name=args.model)
    server = ThreadingHTTPServer(("127.0.0.1", args.port), build_handler(model))
    print(f"embeddings server: {args.model} on http://127.0.0.1:{args.port}/v1/embeddings", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
