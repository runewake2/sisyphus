---
role: model-eval
kind: tool
command: sh "$HEPHAESTUS_ROLE_DIR/eval.sh"
trigger: request
on-failure: needs-human
can-call-referendum: false
---

# Model evaluation

## Charter

Run the task suite of sisyphus-mcp with one local model, and record its results
([[ollama-local-models]]).

It runs the benchmark harness in `design/ollama-local-models/harness/` for the model that its task
names, on the GPU host. It runs one model at a time: a second task waits for the first. It is a tool
role: it uses no model of its own, but the model that it tests uses the GPU, the CPU, and the RAM of
the machine for a long time.

## Inputs

- A task message, with the model in its data, for example `{"model": "qwen3:8b"}`, and optionally
  `"host"`, the Ollama service (default `http://host.docker.internal:11434`).

## Outputs

- The results of each config in `design/ollama-local-models/work/results-gpu/`, and the summary
  lines of the harness.
- The issue of the task, closed with the summary, or stopped with the output in its Notes.

## Works with

- `ux-researcher`: asks for a run, as the small model method of its review.
- `economist/compute`: reads what each run uses of the GPU, the CPU, and the RAM.

## Tools

`eval.sh` in the directory of this role file: the harness `run-model-gpu.sh`, `jq`, and `flock`, so
that one evaluation runs at a time.

## Ask a human

- Before each task: an evaluation uses the GPU and the workstation for a long time, and other work
  shares them. Send it only when a human asks for it.
- When the model has no tool support, or the pull fails.
