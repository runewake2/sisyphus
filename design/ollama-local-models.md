# Local Ollama models as sisyphus task managers

- Date: 2026-10-09
- Issue: [[spike-ollama-local-models]]
- Bookmark: samw/ai/ollama-model-spike
- Data: `design/ollama-local-models/` (harness, raw results, and one analysis for each model in `notes/`)

## Summary

A local model can operate sisyphus as a task manager. **qwen3.5:9b with thinking on passed all 28 runs
(14 tasks, 2 repeats) at 3.3 s for each task**, on a 16 GB GPU, with 5.8 GB of VRAM. That is the same
pass rate as Claude Opus 5.5. The fastest good option is **qwen3:8b with thinking off: 82% at 0.9 s
for each task**.

The settings had a larger effect than the model size. A detailed system prompt and a hidden `dir`
argument raised qwen3:14b from 39% to 86%. Most failures of all models had the same few causes in
`sisyphus-mcp` and the sisyphus commands. [[small-model-friendly-mcp]] tracks the fixes.

## Recommendation

| Use | Model | Settings | Pass rate | Time for each task | VRAM |
|---|---|---|---|---|---|
| Default | qwen3.5:9b | thinking on, detailed prompt, `dir` hidden | 100% | 3.3 s | 5.8 GB |
| Fast | qwen3:8b | thinking off, detailed prompt, `dir` hidden | 82% | 0.9 s | 6.3 GB |
| Not recommended | ministral-3:8b | (no thinking mode) | 68% at best | 0.9 s | 6.3 GB |

For all models:

- Hide the optional `dir` argument from the tool schema (see [[mcp-hide-dir-argument]]).
- Use a system prompt that gives the valid values, the meaning of each state and resolution, and the
  rule "look up names, do not guess". The "detailed" prompt in `harness/harness.py` is a good start.
- Use `num_ctx` 8192. The tool schemas use about 3,300 tokens.

[[ollama-task-manager-guide]] will turn this into a user guide.

## Method

### Harness

`harness/harness.py` connects an Ollama model to `sisyphus-mcp` through the Ollama `/api/chat` tool
calls. For each task, it:

1. Creates a new scratch repo with `sisyphus init` and 6 seeded issues: an epic with 2 sub-issues, an
   in-progress issue with an owner and a bookmark, a critical issue, and a closed issue.
2. Starts `sisyphus-mcp` in the repo, gives the model all 13 tools, and runs the chat for up to 8
   turns.
3. Reads the issue files on disk and the final answer, and checks them. A check also fails when the
   model changes an issue that the task does not name.

`--oracle` runs the correct commands for each task and proves that every check can pass.

### Tasks

| Kind | Tasks |
|---|---|
| Read (t01–t05) | filter by priority, find by topic, owner and bookmark, sub-issues of an epic, a question with the answer "no" |
| Write (t06–t12) | create with fields, start, close, abandon, set parent, add dependency, change priority |
| Multi-step (t13–t14) | create + parent + dependency in one request; find the critical issue, start it, and give its title |

### Configs

| Config | System prompt | Tool schema | Thinking |
|---|---|---|---|
| A | basic (3 sentences) | as served by `sisyphus-mcp` | off |
| B | detailed (valid values, semantics, steps) | `dir` hidden | off |
| C | detailed | `dir` hidden | on (only for models that can think) |

All runs: temperature 0.2, fixed seed, `num_ctx` 8192, 2 repeats.

### Hardware and load limits

The runs used Ollama 0.40.2 on a Windows workstation with a 16 GB GPU. A person used the workstation
at the same time, so the queue kept the load low:

- One model at a time.
- A 25% duty cycle: after each task, the queue rested for 3 times as long as the task took.
- Only models that fit fully in VRAM (no CPU offload).
- Each model was deleted after its run.

The first attempt used the CPU of the sandbox (64 cores, no GPU). qwen3:4b took about 110 s for
each task, and the load was too high, so the spike moved to the GPU. The pass and fail results of
the first 9 CPU runs were the same as on the GPU.

## Results

### All configs

| Model | Config | Pass | Read | Write | Multi | Gen tok/s | Mean s | Median s | Median output tokens | VRAM GB |
|---|---|---|---|---|---|---|---|---|---|---|
| qwen3:4b | A | 82%¹ | 100% | 79% | 50% | 187 | 14.4 | 6.9 | 1317 | 3.9 |
| qwen3:4b | B | 86% | 90% | 79% | 100% | 187 | 16.7 | 8.7 | 1643 | 3.9 |
| qwen3:4b | C | 86% | 100% | 79% | 75% | 183 | 14.9 | 11.1 | 2065 | 3.9 |
| qwen3:8b | A | 64% | 60% | 57% | 100% | 91 | 1.2 | 0.8 | 85 | 6.3 |
| qwen3:8b | B | 82% | 60% | 93% | 100% | 119 | 0.9 | 0.8 | 80 | 6.3 |
| qwen3:8b | C | 86% | 100% | 71% | 100% | 123 | 11.3 | 5.9 | 676 | 6.3 |
| **qwen3.5:9b** | A | 82% | 80% | 93% | 50% | 115 | 1.6 | 1.4 | 100 | 5.8 |
| **qwen3.5:9b** | B | 86% | 80% | 86% | 100% | 114 | 1.5 | 1.2 | 86 | 5.8 |
| **qwen3.5:9b** | **C** | **100%** | 100% | 100% | 100% | 115 | 3.3 | 3.3 | 305 | 5.8 |
| ministral-3:8b | A | 61% | 80% | 64% | 0% | 116 | 0.9 | 0.6 | 56 | 6.3 |
| ministral-3:8b | B | 68% | 80% | 79% | 0% | 112 | 0.9 | 0.8 | 64 | 6.3 |
| qwen3:14b | A | 39% | 70% | 29% | 0% | 78 | 2.7 | 2.4 | 154 | 10.5 |
| qwen3:14b | B | 86% | 80% | 93% | 75% | 75 | 1.4 | 1.4 | 86 | 10.5 |
| qwen3:14b | C | 75%² | 80% | 86% | 25% | 77 | 9.0 | 7.8 | 572 | 10.5 |
| Claude Opus 5.5³ | – | 100% | 100% | 100% | 100% | – | 6.9 | – | – | – |

1. Corrected from the raw 79%. One t06 run was scored wrong by a harness bug, which is now fixed
   (see `notes/qwen3_4b.md`).
2. Ollama on the host stopped during repeat 1. The 8 lost runs ran again with the same seed and
   settings. `results/qwen3_14b-C-before-rerun.json` has the file before the merge.
3. A Claude Code subagent on Claude Opus 5.5, with the basic prompt. It used the `sisyphus` CLI
   through Bash, not the MCP tools, because `sisyphus-mcp` was not connected to it. The MCP tools wrap
   the same commands. Its time includes the start of the subagent and the network.

gpt-oss:20b was not tested. It needs 14–16 GB of VRAM, and that is too much for a 16 GB GPU that a
person also uses.

### Tasks, best config of each model

| Task | qwen3:4b C | qwen3:8b C | qwen3.5:9b C | ministral-3:8b B | qwen3:14b B | Claude Opus 5.5 |
|---|---|---|---|---|---|---|
| t01 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 |
| t02 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 |
| t03 | 2/2 | 2/2 | 2/2 | 0/2 | 0/2 | 2/2 |
| t04 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 |
| t05 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 |
| t06 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 |
| t07 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 |
| t08 | 0/2 | 0/2 | 2/2 | 0/2 | 1/2 | 2/2 |
| t09 | 2/2 | 1/2 | 2/2 | 2/2 | 2/2 | 2/2 |
| t10 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 |
| t11 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 | 2/2 |
| t12 | 1/2 | 1/2 | 2/2 | 1/2 | 2/2 | 2/2 |
| t13 | 2/2 | 2/2 | 2/2 | 0/2 | 1/2 | 2/2 |
| t14 | 1/2 | 2/2 | 2/2 | 0/2 | 2/2 | 2/2 |

## Findings

### Settings

- **The detailed prompt and the hidden `dir` (A → B) helped every model.** The effect was largest
  for qwen3:14b (39% → 86%) and qwen3:8b (64% → 82%).
- **Thinking (B → C) helped qwen3.5:9b (86% → 100%) and the read tasks of qwen3:8b (60% → 100%).** It
  made qwen3:14b worse (86% → 75%): with thinking on, it stopped early, gave 3 empty answers, and asked
  the user instead of looking up issues.
- **Thinking cost differs much between models.** qwen3.5:9b used a median of 305 output tokens
  (maximum 564). qwen3:8b had one run with 16,490 tokens (140 s).
- **qwen3:4b ignores `think: false`.** It wrote its reasoning and a `</think>` tag into the answer in
  almost every run. Thus thinking off saves no time for that model, and C is its best config.

### Failure modes

The same causes appeared in most models:

| Cause | Effect | Fix |
|---|---|---|
| The `dir` description says "Defaults to sisyphus-mcp's own working directory". | Models send `dir: "sisyphus-mcp"`, `"current"`, or `"docs"`, and the call fails with `chdir`. qwen3:14b: 57 errors in config A. | [[mcp-hide-dir-argument]] |
| Search matches only one exact phrase. | `search "docs site theme update"` returns `[]` for "Update the docs site theme". t08 failed in most runs of the qwen3 models. | [[search-matches-all-words]] |
| "No issue named X" has no hint. | Models that guessed a name could not recover. qwen3:4b: 23 errors, qwen3:8b: 22. ministral-3:8b recovered in 0 of 6 runs. | [[suggest-names-on-missing-issue]] |
| The `name` descriptions use the example `web/auth/fix-login-bug`. | Models guess nested names (`docs/theme-update`), create nested names (all 6 t06 runs of qwen3.5:9b), and put the parent into the name (`auth-epic/audit-session-tokens`). | [[clarify-issue-name-descriptions]] |
| List and search rows have no `bookmark`. | With thinking off, models answered t03 with a wrong bookmark. | [[list-search-show-bookmark]] |
| No command edits an issue body. | Claude reported 6 times that it could not fill in Summary or Resolution. | `edit-issue-body-command` (done on bookmark `samw/ai/edit-issue-sections`) |
| The text output of `show` has no `resolution`. | Claude could not confirm an abandon. | `show-prints-resolution` (done on bookmark `samw/ai/edit-issue-sections`) |

Some failures came from the models:

- **Claims without action.** ministral-3:8b said "I started migrate-database" without a call that
  started it (t14, all runs).
- **Giving up early.** qwen3:8b in config A stopped and asked the user after every tool error (10
  runs). The detailed prompt reduced this to 2.
- **Invented data.** With thinking off, qwen3:8b invented issue data after an empty search (3 runs in
  B). Thinking removed this.

### Claude Opus 5.5

Claude passed 28 of 28 with the basic prompt and no other help. It never guessed a name. It searched
with single keywords ("payment", "billing", "stripe"), and it checked its changes with `show`. It also
told the user when a request looked like a mistake (making the database migration a sub-issue of the
auth epic). Each run used about 34,000 tokens, most of them for the context of the Claude Code agent.

The planned comparison of claude-opus-5-5, claude-opus-5, and claude-sonnet-5 through the Claude API
did not run: the API key of the test sandbox returned HTTP 429 for every model. See
[[compare-cloud-claude-models]].

## Limits

- The suite is small (14 tasks) and synthetic. It does not test long conversations, many issues, or
  vague requests.
- 2 repeats for each task give a coarse pass rate. One run is 3.6 percentage points.
- One GPU, one Ollama version, and one quantization for each model (the default tag, mostly Q4_K_M).
- Claude used the CLI and its help text. The local models used the MCP tools and their schemas.

## Reproduce

```bash
cd design/ollama-local-models/harness
export SISYPHUS_BIN=<directory with sisyphus and sisyphus-mcp>
python3 harness.py --model x --oracle                          # check the checkers
./gpu-queue.sh qwen3.5:9b qwen3:8b                             # Ollama at host.docker.internal:11434
python3 harness.py --model qwen3.5:9b --host http://127.0.0.1:11434 \
  --think on --prompt detailed --schema nodir --repeats 2      # one config by hand
```

`gpu-queue.sh` takes `BENCH_WORK` (the work directory, default `../work`) and stops between tasks
while the file `$BENCH_WORK/PAUSE` exists. `run-model.sh` and `opull.sh` are the CPU-only variant
that the spike used first.
