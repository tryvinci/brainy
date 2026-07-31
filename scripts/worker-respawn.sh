#!/bin/sh
# Staging starter workers have been observed to receive SIGTERM ~60s after
# boot and then stay down until the next deploy. Keep the extractor alive by
# respawning; ignore TERM so a single signal cannot permanently drain the queue.
# Fresh deploys still replace the container (Render eventually SIGKILLs).
trap 'true' TERM INT
while true; do
  /app/brainy-worker
  code=$?
  echo "brainy-worker exited code=${code}; restarting in 2s" >&2
  sleep 2
done
