---
title: "Color the issues linked directly to the graph focus by state"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [feature, cli]   # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-08
closed: 2026-10-08     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/graph-directory-boxes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [graph-directory-boxes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {session-id: "9c5453b3-9fdb-4b87-ae3f-6d154e825978"} # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Color the issues linked directly to the graph focus by state

## Summary

In a colored `sisyphus graph`, draw each issue that an arrow links directly to the focus issue in the color of its state.

## Context

From the human, 2026-10-08, extending the 0.0.26 color experiment: color the issues that the focus issue's arrows link directly, by state: blue for open, yellow for in-progress, green for completed, and red for abandoned. Read as both directions of a direct arrow (parent, dependencies, sub-issues, dependents). A closed issue with no resolution counts as completed, the template's default.

## Acceptance criteria

- [x] A directly linked issue is blue if open, yellow if in-progress, green if completed, and red if abandoned
- [x] Issues that are not linked directly stay gray, and the focus and its arrows stay cyan
- [x] A legend line names the state colors that the drawing uses, only with color
- [x] `--json` nodes carry `resolution`, so a reader can tell completed from abandoned

## Out of scope

- State colors for issues that are not linked directly to the focus.

## Notes

2026-10-08: implemented. [[0008-color-linked-issues-by-state]] records it.

## Resolution

Completed. sisyphus [[0.0.27]] on bookmark samw/ai/graph-directory-boxes, on top of 0.0.26; it waits for a human merge.
