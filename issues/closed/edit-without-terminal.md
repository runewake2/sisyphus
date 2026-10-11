---
title: "Do not start an editor without a terminal"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli, sisyphus-mcp, llm] # The components that the work touches, for example [widget-scheduler, plan]
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

# Do not start an editor without a terminal

## Summary

`sisyphus edit <name>` starts `$VISUAL` or `$EDITOR` even when no terminal is attached. An agent that runs it through a shell tool, and the `sisyphus_edit` MCP tool with an empty `section` and `text`, then wait for an editor that no one can use. Refuse to start an editor without a terminal.

## Context

Since [[0.0.84]], `edit` with only a name opens the editor. The check is the number of arguments only. `editInEditor` in `cmd/sisyphus/editor.go` does not check for a terminal.

- An agent's shell tool has no terminal. With `EDITOR=vim`, vim prints "Input is not from a terminal" and waits. The tool call does not end until its timeout.
- `sisyphus_edit` marks `section` and `text` as required, but an empty string is valid JSON. `appendPositionals` drops empty values, so the tool runs `sisyphus edit -- <name>`. In a test, a `sisyphus_edit` call with `"section": "", "text": ""` and an editor that waits 20 seconds took 20 seconds, and sisyphus-mcp started the editor of the user from the MCP server process.

Proposed:

- `edit` with only a name fails when standard input is not a terminal: `No terminal to run your editor in. To change one section, use: sisyphus edit <name> <section> <text>`.
- `sisyphus_edit` fails at once when `section` is empty, with a clear message.

## Acceptance criteria

- [ ] `sisyphus edit <name> </dev/null` fails at once, with the message above.
- [ ] `sisyphus_edit` with an empty `section` fails and does not start an editor.
- [ ] Tests cover both.

## Out of scope

- Time limits of other tools: [[mcp-cancel-and-timeout]].

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.95]]. `edit <name>` fails at once without a terminal and names the section form, and `sisyphus_edit` refuses an empty `section` before it runs sisyphus.

