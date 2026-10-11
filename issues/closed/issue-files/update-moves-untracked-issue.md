---
title: "Move an issue that git does not track"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli]       # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-08
closed: 2026-10-08     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/jj-colocated-move # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-jj-colocated-move] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Move an issue that git does not track

## Summary

`sisyphus update` must move an issue that git does not track. In a colocated jj repo, the first state change of every new issue failed.

## Context

In a colocated jj repo, jj does not add a new file to the git index until it commits the file. sisyphus update ran git mv for every move when the repo root had .git, and git mv refuses a file that git does not track. So the first state change of a new issue failed. A plain git repo had the same failure for an issue that was not added yet. See [[update-uses-vcs-move]].

## Acceptance criteria

- [x] `sisyphus update` moves an issue that git does not track, in a repo that has `.git`.
- [x] `sisyphus update` still uses `git mv` for an issue that git tracks.
- [x] A test covers an issue that git does not track.

## Out of scope

None.

## Notes

2026-10-08: found while sisyphus was configured in a colocated jj vault.

## Resolution

Completed in [[0.0.17]]. `saveIssue` runs `git mv` only when the git index has the file.
