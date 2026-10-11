---
title: "Remove jj from the MCP argument descriptions"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [sisyphus-mcp, docs] # The components that the work touches, for example [widget-scheduler, plan]
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

# Remove jj from the MCP argument descriptions

## Summary

The `bookmark` and `workspace` arguments of `sisyphus_new` and `sisyphus_update` say "The jj bookmark" and "The jj workspace". The CLI flags no longer assume jj ([[vcs-neutral-defaults]]). An agent in a git repo can think that the arguments are not for it. Use the words of the CLI flags.

## Context

In `cmd/sisyphus-mcp/tools.go`, `newArgs` and `updateArgs`:

- `bookmark`: "The jj bookmark of the work on the issue."
- `workspace`: "The jj workspace where local work on the issue is happening."

[[0020-do-not-assume-jj]] says that sisyphus works the same with git alone. The CLI help says "A jj bookmark or git branch" and "jj workspace or git worktree". A test can compare each MCP argument description with the help text of its flag, so the two stay the same.

## Acceptance criteria

- [ ] No MCP argument description names only jj.
- [ ] A test checks that the descriptions of `bookmark` and `workspace` match the CLI help, or that no description has "jj" without "git".

## Out of scope

- The other argument descriptions: [[clarify-issue-name-descriptions]].

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.85]]. The `bookmark` and `workspace` descriptions of `sisyphus_new` and `sisyphus_update` name jj and git, as the CLI help does. A test checks every tool and argument description.

