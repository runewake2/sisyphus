#!/usr/bin/env bash
# Benchmark one Ollama model against sisyphus-mcp with the standard configs.
# Usage: run-model.sh <model> <port> [threads]
# Starts a private ollama server on <port>, pulls the model, runs:
#   A baseline : prompt=basic    schema=raw   think=off (or default if no thinking)
#   B tuned    : prompt=detailed schema=nodir think=off (or default)
#   C thinking : prompt=detailed schema=nodir think=on   (only if the model can think)
# Results: results/<model-label>__*.json, logs in results/logs/.
set -uo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
SP="${BENCH_WORK:-$HERE/../work}"; export BENCH_WORK="$SP"
MODEL="$1"; PORT="$2"; THREADS="${3:-20}"
LBL="${MODEL//[:\/]/_}"
LOG="$SP/results/logs"; mkdir -p "$LOG"
export OLLAMA_MODELS="$SP/ollama-models"
HOST="http://127.0.0.1:$PORT"

"$HERE/opull.sh" "$MODEL" || { echo "PULL FAILED"; exit 2; }

OLLAMA_HOST="127.0.0.1:$PORT" OLLAMA_NUM_PARALLEL=1 OLLAMA_MAX_LOADED_MODELS=1 OLLAMA_KEEP_ALIVE=30m \
  "${OLLAMA_BIN:-ollama}" serve > "$LOG/$LBL.serve.log" 2>&1 &
SRV=$!
trap 'kill $SRV 2>/dev/null; wait $SRV 2>/dev/null' EXIT
for _ in $(seq 60); do curl -sf "$HOST/api/version" >/dev/null && break; timeout 1 tail -f /dev/null; done

CAPS="$(curl -s "$HOST/api/show" -d "{\"model\":\"$MODEL\"}" | jq -c '.capabilities')"
echo "capabilities: $CAPS"
if ! jq -e 'index("tools")' <<<"$CAPS" >/dev/null; then echo "NO TOOL SUPPORT"; exit 3; fi
THINK=default; jq -e 'index("thinking")' <<<"$CAPS" >/dev/null && THINK=off

H="python3 $HERE/harness.py --model $MODEL --host $HOST --num-thread $THREADS --repeats 2 --out $SP/results"
$H --think $THINK --prompt basic --schema raw      2>&1 | tee "$LOG/$LBL.A.log"
$H --think $THINK --prompt detailed --schema nodir 2>&1 | tee "$LOG/$LBL.B.log"
if [[ $THINK == off ]]; then
  $H --think on --prompt detailed --schema nodir   2>&1 | tee "$LOG/$LBL.C.log"
fi
echo "DONE $MODEL"
