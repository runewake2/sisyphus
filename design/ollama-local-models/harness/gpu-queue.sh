#!/usr/bin/env bash
# Run the GPU benchmark queue one model at a time on the shared workstation.
# Models that were on the host before the queue started are kept; others are deleted after their run.
# Touch $SP/PAUSE to hold the queue (between tasks and between models); remove it to resume.
set -uo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
SP="${BENCH_WORK:-$HERE/../work}"; export BENCH_WORK="$SP"
HOST=http://host.docker.internal:11434
OUT="$SP/results-gpu"; mkdir -p "$OUT/logs"
[[ -f "$OUT/preexisting.json" ]] || curl -s "$HOST/api/tags" | jq '[.models[].name]' > "$OUT/preexisting.json"
for m in "$@"; do
  while [[ -f "$SP/PAUSE" ]]; do sleep 15; done
  if [[ -f "$OUT/logs/${m//[:\/]/_}.done" ]]; then echo "skip $m (done)"; continue; fi
  echo "=== $m $(date -Is)"
  "$HERE/run-model-gpu.sh" "$m" "$HOST"; rc=$?
  echo "$rc" > "$OUT/logs/${m//[:\/]/_}.done"
  if ! jq -e --arg m "$m" 'index($m)' "$OUT/preexisting.json" >/dev/null; then
    if curl -sf -X DELETE "$HOST/api/delete" -d "{\"model\":\"$m\"}" >/dev/null; then echo "deleted $m"; else echo "DELETE FAILED $m"; fi
  fi
  echo "=== end $m rc=$rc $(date -Is)"
done
echo "QUEUE DONE"
