---
title: "Stop the sisyphus command when an MCP call is cancelled or too slow"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [sisyphus-mcp, llm] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[agent-cooperation]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Stop the sisyphus command when an MCP call is cancelled or too slow

## Summary

sisyphus-mcp runs `sisyphus` with `exec.Command` and ignores the context of the call. A client that cancels a call, or a command that does not end, leaves the process to run, and the call never returns. Use the context of the call, and add a time limit.

## Context

`execSisyphus` in `cmd/sisyphus-mcp/main.go` uses `exec.Command`, not `exec.CommandContext`, and the handlers ignore their `context.Context`. The MCP client sends `notifications/cancelled` when a user stops a call, but the `sisyphus` process continues.

[[edit-without-terminal]] showed one way that a command waits for ever. Others: a `git mv` that waits for a lock of the git index, or a very large repo.

Proposed: pass the context of the call to `exec.CommandContext`, and add a default time limit (for example 60 seconds, changeable with a server flag or an environment variable). A call that reaches the limit returns an error that tells the limit and the command.

## Acceptance criteria

- [ ] A cancelled call stops its `sisyphus` process.
- [ ] A command that runs longer than the limit stops and returns a clear error.
- [ ] [[sisyphus-mcp]] documents the limit and how to change it.

## Out of scope

- Progress notifications for long commands.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.96]]. Each run uses the context of the call and a time limit of 60 seconds, which `SISYPHUS_MCP_TIMEOUT` changes. A cancelled or slow run stops with a clear error.

