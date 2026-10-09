---
title: "Spike: compare local Ollama models as sisyphus task managers"
state: in-progress     # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [spike, ollama, sisyphus-mcp] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-09
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ollama-model-spike # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ollama-model-spike] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Spike: compare local Ollama models as sisyphus task managers

## Summary

Find out which local models that run in Ollama can operate sisyphus as a task manager, through the
`sisyphus-mcp` tools. Run each candidate model against the same set of tasks, measure the results,
and record which models and settings work best.

## Context

`sisyphus-mcp` gives an agent every `sisyphus` command as an MCP tool over stdio (see
[[sisyphus-mcp]] and [[commands]]). Cloud models use these tools well. We do not know if a small
local model can do the same. If one can, a user can run Ollama as a private, offline task manager
for a repo.

A model must do these things to be useful:

- Select the correct tool for a request, for example `sisyphus_list` to find open work.
- Supply valid arguments: issue names in kebab-case, valid states, priorities, and efforts.
- Do multi-step work, for example create an issue, then set its parent and a dependency.
- Read tool output and give a correct answer, without invented issues or fields.

The spike uses a small harness that connects an Ollama model to `sisyphus-mcp` through the Ollama
chat API with tools. The harness runs a fixed task suite against a scratch repo that `sisyphus init`
creates, and then checks the resulting files. Each model is tested by a separate subagent, and the
results go into one report.

Candidate models (only models that Ollama marks as tool-capable, and that fit on the test machine):
the `qwen3` family, `llama3.1` / `llama3.2`, `mistral` / `mistral-nemo`, `granite3.3`, `phi4-mini`,
`smollm2`, and others that the subagents find.

## Acceptance criteria

- [ ] A repeatable harness connects an Ollama model to `sisyphus-mcp` and runs a fixed task suite.
- [ ] The task suite covers create, update state, list, search, show, parent, depends-on, and a
      multi-step request, and checks each result against the files on disk.
- [ ] Each candidate model has a result: pass rate per task, tool-call validity, speed
      (tokens per second and wall time), and memory use.
- [ ] A report compares the models and recommends a default model and settings (for example the
      system prompt, context size, temperature, and the set of tools to expose).
- [ ] Follow-up work (for example changes to tool descriptions that help small models) is recorded
      as new issues.

## Out of scope

- Changes to `sisyphus` or `sisyphus-mcp` behavior. Record them as follow-up issues.
- Fine-tuning a model.
- Hosted model APIs.

## Notes

- 2026-10-09: Spike started. The test sandbox has 64 CPU cores, 31 GB RAM, and no GPU, so only
  CPU inference is possible there, and speed results are a lower bound.

## Resolution

