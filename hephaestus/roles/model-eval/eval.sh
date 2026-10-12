#!/bin/sh
# The command of the role model-eval: the benchmark harness for the model of the task, one at a time.
# The task message is the standard input, as JSON.
set -u
command -v jq >/dev/null 2>&1 || { echo "jq is not installed."; exit 1; }
message=$(cat)
model=$(printf '%s' "$message" | jq -r '.data.model // empty')
host=$(printf '%s' "$message" | jq -r '.data.host // "http://host.docker.internal:11434"')
if [ -z "$model" ]; then
	echo 'The task has no model: give it in the data, for example {"model": "qwen3:8b"}.'
	exit 1
fi
root=$(hephaestus files --root) || exit 1
harness="$root/design/ollama-local-models/harness"
work="${BENCH_WORK:-$harness/../work}"
mkdir -p "$work"
# One evaluation at a time: the machine is shared.
exec flock "$work/model-eval.lock" "$harness/run-model-gpu.sh" "$model" "$host"
