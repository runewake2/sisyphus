---
title: "Remove empty subdirectories after an issue moves"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli]       # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/tracker-fixes-2 # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [tracker-fixes-2] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Remove empty subdirectories after an issue moves

## Summary

When an issue in a subdirectory changes state, its old subdirectory stays, empty, for example `issues/open/web/auth/`. Obsidian shows the empty folders, and they collect over time.

## Context

Found in a review of sisyphus 0.0.45 on 2026-10-11.

```
sisyphus new web/auth/move-test
sisyphus update move-test closed
find issues/open -type d        # issues/open/web and issues/open/web/auth stay
```

`saveIssue` in `cmd/sisyphus/issues.go` moves the file and leaves the directories. git and jj ignore empty
directories, but the working copy and Obsidian do not.

## Acceptance criteria

- [ ] After a move, sisyphus removes each directory between the old file and `issues/<state>/` that is now empty, and keeps `issues/<state>/` itself and any directory with other files.
- [ ] A test covers a nested move.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.56]]. After a move, sisyphus removes the empty directories between the old file and the state directory.

