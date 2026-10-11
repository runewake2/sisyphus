---
title: "Add start, close, and abandon shortcuts for common state changes"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [ux, cli, feature] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[ux/cli-and-docs-ux]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add start, close, and abandon shortcuts for common state changes

## Summary

The most common changes, start work and close an issue, need `sisyphus update <name> in-progress` and `sisyphus update <name> closed --resolution abandoned`. Add short commands for them.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo. A user tried `sisyphus update fix-login-bug start` and `... done` before finding the state names
`in-progress` and `closed`. Shortcuts:

```
sisyphus start <name> [--owner X] [--bookmark Y]   # = update <name> in-progress
sisyphus close <name>                              # = update <name> closed
sisyphus abandon <name> [--reason "..."]           # = update <name> closed --resolution abandoned
```

Each one is `update` with fewer words, so the rules and warnings stay in one place. `abandon --reason` could fill
the Resolution section, which `update closed` warns about.

## Acceptance criteria

- [ ] `start`, `close`, and `abandon` call the same code as `update`, with the same flags where they apply.
- [ ] `abandon --reason` (and `close --resolution-text`, or a similar flag) fills the Resolution section.
- [ ] The MCP server exposes them, or the `sisyphus_update` description names the state values clearly; decide which.
- [ ] Docs pages and tests for each command.

## Out of scope

- Removing `update`: it stays the general command.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.73]]. `start`, `close`, and `abandon` call `update`, and `-m` fills the Resolution section when an issue closes. `sisyphus_update` has a `message` argument.

