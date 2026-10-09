---
title: "Stop models from filling the optional dir argument with junk"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [sisyphus-mcp, llm] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-09
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from: "[[spike-ollama-local-models]]" # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[small-model-friendly-mcp]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Stop models from filling the optional dir argument with junk

## Summary

Small models fill the optional `dir` argument of every sisyphus-mcp tool with junk values, and each such call fails. Change the tools so that a model does not fill `dir` when it has no reason to.

## Context

Every tool in `cmd/sisyphus-mcp/tools.go` has an optional `dir` argument with this description: "The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory."

In the spike ([[ollama-local-models]]), models copied words from this text into `dir`: `"sisyphus-mcp"`, `"current"`, `"docs"`, `"web"`. The command then fails with `chdir <value>: no such file or directory`. qwen3:14b sent `dir` in 70 of 76 calls in config A and got 57 such errors. That is the cause of every config A failure of qwen3:14b (39% pass rate). When the harness hid `dir` from the schema, qwen3:14b went to 86%.

Options: remove `dir` from the default tool schema and add a server flag (for example `--allow-dir`) for clients that need it; or keep it, but change the description so that it does not name a value (for example "Leave this empty.").

## Acceptance criteria

- [ ] A model that uses the default tool schema does not see a `dir` description that it can copy as a value.
- [ ] Clients that need to act on another repo can still do so.
- [ ] The spike harness with `--schema raw` on qwen3:14b shows no `chdir` errors.

## Out of scope

- Other argument descriptions (see [[clarify-issue-name-descriptions]]).

## Notes

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
