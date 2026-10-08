---
title: "Show the issue state on a second line in sisyphus graph boxes"
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

# Show the issue state on a second line in sisyphus graph boxes

## Summary

In `sisyphus graph`, show the state of an issue on a second line below its name, so boxes and graphs are narrower.

## Context

From the human, 2026-10-08: move the issue state onto a second line below the issue name in each graph box, to save horizontal space. Since 0.0.22 each dependency step adds a column, so a graph is often wide, and `[in-progress]` adds 14 columns to a box label.

## Acceptance criteria

- [x] An issue box shows the file name on one line and the state on the line below
- [x] A box is as wide as the longer of the two lines
- [x] Arrows leave and enter on the line of the name
- [x] `--mermaid` and `--json` output do not change

## Out of scope

- Abbreviated states.

## Notes

2026-10-08: implemented, and the state has no brackets on its own line. The doc example on an issue inside an epic went from 117 to 77 columns. [[0006-show-the-state-below-the-name]] records it.

## Resolution

Completed. sisyphus [[0.0.25]] on bookmark samw/ai/graph-directory-boxes, on top of 0.0.24; it waits for a human merge.
