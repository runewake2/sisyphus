#!/usr/bin/env bash
# Benchmark one model on a remote (GPU) Ollama server with the standard configs.
# Usage: run-model-gpu.sh <model> [host-url]   (default host: http://host.docker.internal:11434)
# Pulls through the server's own /api/pull, runs configs A/B/C like run-model.sh,
# then unloads the model so the next one gets the whole GPU.
set -uo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
SP="${BENCH_WORK:-$HERE/../work}"; export BENCH_WORK="$SP"
MODEL="$1"; HOST="${2:-http://host.docker.internal:11434}"
LBL="${MODEL//[:\/]/_}"; OUT="$SP/results-gpu"; LOG="$OUT/logs"; mkdir -p "$LOG"

curl -sf "$HOST/api/version" | grep -q version || { echo "SERVER UNREACHABLE $HOST"; exit 4; }
curl -s --max-time 7200 "$HOST/api/pull" -d "{\"model\":\"$MODEL\",\"stream\":false}" | tee "$LOG/$LBL.pull.log"
grep -q '"success"' "$LOG/$LBL.pull.log" || { echo "PULL FAILED"; exit 2; }

CAPS="$(curl -s "$HOST/api/show" -d "{\"model\":\"$MODEL\"}" | jq -c '.capabilities')"
echo "capabilities: $CAPS"
jq -e 'index("tools")' <<<"$CAPS" >/dev/null || { echo "NO TOOL SUPPORT"; exit 3; }
THINK=default; jq -e 'index("thinking")' <<<"$CAPS" >/dev/null && THINK=off

# The host is a shared workstation: only run models that fit fully in VRAM (no CPU offload).
curl -s --max-time 600 "$HOST/api/generate" -d "{\"model\":\"$MODEL\",\"prompt\":\"\",\"keep_alive\":\"10m\",\"options\":{\"num_ctx\":8192}}" >/dev/null
FIT="$(curl -s "$HOST/api/ps" | jq -c --arg m "$MODEL" '.models[] | select(.name==$m) | {size, size_vram}')"
echo "loaded: $FIT"
if ! jq -e '.size_vram >= .size' <<<"$FIT" >/dev/null; then
  curl -s "$HOST/api/generate" -d "{\"model\":\"$MODEL\",\"keep_alive\":0}" >/dev/null
  echo "$FIT" > "$LOG/$LBL.novram"; echo "DOES NOT FIT VRAM"; exit 5
fi

H="python3 $HERE/harness.py --model $MODEL --host $HOST --num-thread 0 --repeats 2 --duty 0.25 --out $OUT"
$H --think $THINK --prompt basic --schema raw      2>&1 | tee "$LOG/$LBL.A.log"
$H --think $THINK --prompt detailed --schema nodir 2>&1 | tee "$LOG/$LBL.B.log"
[[ $THINK == off ]] && $H --think on --prompt detailed --schema nodir 2>&1 | tee "$LOG/$LBL.C.log"

curl -s "$HOST/api/generate" -d "{\"model\":\"$MODEL\",\"keep_alive\":0}" >/dev/null   # unload
echo "DONE $MODEL"
