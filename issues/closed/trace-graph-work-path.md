---
title: "Mark start-now issues and trace the work path to the focus in sisyphus graph"
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

# Mark start-now issues and trace the work path to the focus in sisyphus graph

## Summary

In `sisyphus graph`, mark the issues to start now and draw the path of waiting from them to the focus issue, with and without color.

## Context

From the human, 2026-10-08: mark the ready-to-start work in the graph in a noticeable way, and trace the path from that work to the focus issue, so it is clear what to work on now and what later.

The focus issue waits on its open dependencies and open sub-issues, and they wait on theirs. Following those links through issues that are not closed ends at available issues: the work to start now. The path is every arrow on those chains. It must show without color (heavy lines), and with color.

## Acceptance criteria

- [x] The work path follows what the focus waits on (dependencies and sub-issues that are not closed) down to available issues
- [x] Work-path arrows have heavy lines, so the path shows without color
- [x] A legend line names the issues to start now
- [x] With color, the work path and the start-now borders are bold magenta
- [x] An available or closed focus has no work path and no path legend

## Out of scope

- Ordering the start-now issues by priority or effort.

## Notes

2026-10-08: implemented. A cell now stores which of its directions are heavy, and a table made from the Unicode names draws every light and heavy mix. The table also showed that the join at a heavy box border was wrong (`┝`; it is `┠`). [[0009-trace-the-work-path]] records it.

## Resolution

Completed. sisyphus [[0.0.28]] on bookmark samw/ai/graph-directory-boxes, on top of 0.0.27; it waits for a human merge.
