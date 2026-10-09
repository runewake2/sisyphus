---
title: "Document how to run sisyphus with a local Ollama model"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [docs, ollama, llm] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-09
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from: "[[spike-ollama-local-models]]" # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Document how to run sisyphus with a local Ollama model

## Summary

Write a guide that tells how to run sisyphus with a local Ollama model as a private task manager, with the model and settings that the spike recommends.

## Context

The spike ([[ollama-local-models]]) found that qwen3.5:9b with thinking on passes all 14 tasks at about 3.3 s per task on a 16 GB GPU, and that qwen3:8b with thinking off is the fastest good option. It also found that settings matter more than model size: a detailed system prompt and a hidden `dir` argument raised qwen3:14b from 39% to 86%.

The guide must tell how to: install Ollama, pull the model, connect an MCP client to `sisyphus-mcp`, set the system prompt (the "detailed" prompt in the spike harness is a starting point), and set `num_ctx` and thinking. Write it after [[small-model-friendly-mcp]], because those changes may make some settings unnecessary.

## Acceptance criteria

- [ ] A page in `docs/` tells how to run sisyphus with a local Ollama model.
- [ ] The page names a recommended model and settings, with a link to the spike results.
- [ ] README links to the page.

## Out of scope

- Support for model runners other than Ollama.

## Notes

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
