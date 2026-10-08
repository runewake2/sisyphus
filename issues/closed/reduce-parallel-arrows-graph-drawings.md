---
title: "Reduce parallel arrows in sisyphus graph drawings"
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
deferred-from: "[[group-graph-nodes-directory-boxes]]" # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Reduce parallel arrows in sisyphus graph drawings

## Summary

Lay out `sisyphus graph` so that arrows do not turn back or run side by side, and each arrow is easy to follow.

## Context

From the human, 2026-10-08, about 0.0.21: the render is messy, with many parallel lines. Example: `sisyphus graph plugin-protocol` in the sisyphus repo.

Cause: an issue goes in the column after the first issue that points to it, so all sub-issues of an epic share one column. Each dependency between two of them then leaves on the right, runs back along a blank line, and enters on the left: two vertical lines and one long horizontal line per dependency, side by side.

## Acceptance criteria

- [x] In a graph without a cycle, each arrow points right and has one turn at most
- [x] Arrows that skip columns run straight along the row of the issue they enter, not along a shared line
- [x] Directory boxes still hold exactly their own issues
- [x] The examples in the graph docs are drawn by the new layout

## Out of scope

- A top-to-bottom drawing for wide graphs.

## Notes

2026-10-08: an issue now goes one column after the farthest issue that points to it, and takes the mean row of its predecessors in its directory. An issue that an arrow reaches from two or more columns back takes a row that is free in every column between. On `graph plugin-protocol` in the sisyphus repo, every arrow is now one straight line with one turn at most, and no arrow loops back. The drawing is a staircase: about 150 columns wide and 23 lines tall for 7 issues.

## Resolution

Completed. sisyphus [[0.0.22]] on bookmark samw/ai/graph-directory-boxes, on top of 0.0.21; it waits for a human merge. [[0003-place-issues-by-their-longest-chain]] records the placement. TestGraphArrowsTakeNoDetourWithoutACycle checks it.
