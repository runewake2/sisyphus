---
title: "Add -C to run sisyphus on another repo"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
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

# Add -C to run sisyphus on another repo

## Summary

sisyphus always finds the repo from the current directory (or its own location). Add a global `-C <dir>`, as in git, so that a script can work on another repo without `cd`.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo. `init` has `--dir`, and every MCP tool has a `dir` argument, but the CLI commands do not. A global
flag makes `findRoot` start at the given directory. It also gives `sisyphus-mcp` one way to pass the directory
(today it sets the working directory of the child process).

## Acceptance criteria

- [ ] `sisyphus -C <dir> <command>` works for every command, and finds the repo from `<dir>`.
- [ ] `init --dir` keeps working.
- [ ] A test covers `-C` with a relative and an absolute path.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.68]]. `-C` works for every command, and `init` writes into it.

