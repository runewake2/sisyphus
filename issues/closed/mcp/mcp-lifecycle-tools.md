---
title: "Add start, close, and abandon tools to the MCP server"
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
parent: "[[agents/agent-cooperation]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add start, close, and abandon tools to the MCP server

## Summary

The CLI has `start`, `close`, and `abandon`, but sisyphus-mcp has no tools for them. An agent that reads AGENTS.md or the docs knows the words "start" and "close", and must find out that it uses `sisyphus_update` with `state` and `resolution`. Add `sisyphus_start`, `sisyphus_close`, and `sisyphus_abandon`.

## Context

[[workflow-shortcut-commands]] added the three commands in [[0.0.73]]. [[sisyphus-mcp]] tells agents to use `sisyphus_update` for them. That is a mapping that each agent must learn. A small model also confuses `state: closed` with `resolution: abandoned` more easily than it confuses two tool names.

Each tool takes the same arguments as its command: `name`, plus `owner`, `bookmark`, and `workspace` for start, and `message` for close and abandon. Abandon requires `message`, so that the Resolution tells why.

Also see [[start-refuses-claimed-issue]]: `sisyphus_start` must pass `take`.

## Acceptance criteria

- [ ] `sisyphus_start`, `sisyphus_close`, and `sisyphus_abandon` exist, and run the commands of the same names.
- [ ] Tests cover each tool.
- [ ] [[sisyphus-mcp]] lists them and no longer tells agents to use `update` for them.

## Out of scope

- Removal of `state` from `sisyphus_update`.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.90]]. `sisyphus_start`, `sisyphus_close`, and `sisyphus_abandon` run the commands of the same names. Start takes `take`, and close and abandon need a `message`.

