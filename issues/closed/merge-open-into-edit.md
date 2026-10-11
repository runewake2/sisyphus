---
title: "Use edit, not open, to edit an issue in the editor"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [ux, cli]        # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-ux-fixes, ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[cli-and-docs-ux]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Use edit, not open, to edit an issue in the editor

## Summary

Remove `sisyphus open`. `sisyphus edit <name>` with no section opens the issue in the editor, and `edit <name> <section> <text>` works as before.

## Context

Reported by the user on 2026-10-11 after [[0.0.76]]: `sisyphus open` sounds like it reopens an issue (the open state), not like it opens the file in an editor. `edit` already edits one section; with only a name, it can open the editor.

## Acceptance criteria

- [ ] `sisyphus edit <name>` opens $VISUAL or $EDITOR and checks the file after.
- [ ] `open` is gone from the CLI, the completion, the help, and the docs.
- [ ] `--append` and `--json` without a section, and two arguments, are errors.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.84]]. `sisyphus edit <name>` opens the issue in the editor, `edit <name> <section> <text>` works as before, and `open` is removed from the CLI, completion, help, and docs.

