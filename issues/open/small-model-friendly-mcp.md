---
title: "Make sisyphus-mcp work well with small local models"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [sisyphus-mcp, llm] # The components that the work touches, for example [widget-scheduler, plan]
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

# Make sisyphus-mcp work well with small local models

## Summary

Change the sisyphus-mcp tools and the commands under them so that small local models can use them reliably. The spike found the same few causes behind most failures of every local model.

## Context

The spike [[spike-ollama-local-models]] tested five local Ollama models against `sisyphus-mcp` on 14 tasks (see [[ollama-local-models]]). Most failures came from the tools, not from the models:

- The optional `dir` argument gets junk values ([[mcp-hide-dir-argument]]).
- Search matches only one exact phrase ([[search-matches-all-words]]).
- An unknown name gives no hint about the correct name ([[suggest-names-on-missing-issue]]).
- The nested-name example in the `name` descriptions causes invented names ([[clarify-issue-name-descriptions]]).
- List and search rows have no `bookmark` ([[list-search-show-bookmark]]).
- The text output of `show` has no `resolution`. Another line of work fixes this: `show-prints-resolution` (done on bookmark `samw/ai/edit-issue-sections`).

The harness in `design/ollama-local-models/harness/` can measure each change. Run it again after the changes and compare with the results of the spike.

## Acceptance criteria

- [ ] All sub-issues are closed.
- [ ] The spike harness runs again on qwen3:8b and qwen3:4b, and the pass rate of config A (basic prompt, raw schema) goes up. Record the new results in [[ollama-local-models]].

## Out of scope

- Changes that help only one model.
- Fine-tuning a model.

## Notes

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
