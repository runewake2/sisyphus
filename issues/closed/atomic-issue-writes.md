---
title: "Write issue files atomically and lock against concurrent changes"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli, llm]  # The components that the work touches, for example [widget-scheduler, plan]
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

# Write issue files atomically and lock against concurrent changes

## Summary

sisyphus writes an issue file in place, so a command that reads at the same time can see a part of the file. Two commands that change the same issue at the same time can lose a change. Write each file to a temporary file and rename it, and lock the issue while a command changes it.

## Context

`document.save` in `cmd/sisyphus/document.go` calls `os.WriteFile`, which truncates the file and then writes it. A `list` or `lint` of another agent in the same checkout that reads the file between the two steps sees an empty or short file, and reports a frontmatter error. `saveIssue` in `cmd/sisyphus/issues.go` then moves the file in a second step.

Two `start` commands on the same issue each read the file, change it, and write it. The second write removes the change of the first. With [[start-refuses-claimed-issue]], both commands can still read the issue as open and both succeed.

Agents often share one checkout, for example sub-agents of one session. So this is not rare.

Proposed: write to a temporary file in the same directory, then rename it over the target. For a change, hold a lock file (for example `issues/.lock`, or one lock per issue) from the read to the write, and fail with a clear error if the lock is held for longer than some seconds.

## Acceptance criteria

- [ ] A reader never sees a partial issue file.
- [ ] Of two concurrent `start` commands on one open issue with different owners, one succeeds and one fails.
- [ ] A test runs concurrent changes of one issue and checks that no change is lost.
- [ ] The lock file is not tracked by the VCS and lint ignores it.

## Out of scope

- Claims across workspaces: [[claims-across-workspaces]].

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.88]]. Issue writes go through a temporary file and a rename, and commands that change issues hold one lock in the shared jj or git directory. Concurrent starts give one owner, and concurrent appends lose no text. See [[0023-lock-and-write-issues-atomically]].

