---
title: "Color the focus issue and its arrows in sisyphus graph"
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

# Color the focus issue and its arrows in sisyphus graph

## Summary

On a terminal, draw the focus issue of `sisyphus graph` and its direct arrows in a bold bright color, and the rest of the drawing in a dim color, as an experiment.

## Context

From the human, 2026-10-08, as an experiment: in the terminal drawing, use bold colors for the issue given as the argument and for every arrow attached directly to it, and a dimmer color for everything else. Color must not reach output that is piped, `--json`, `--mermaid`, or the MCP tool, so a `--color auto|always|never` flag decides, and `auto` follows the terminal and `NO_COLOR`.

## Acceptance criteria

- [x] The focus box, and each arrow that leaves or enters it with its arrowhead, are bold bright cyan
- [x] Every other cell of the drawing is gray, and the legend has no color
- [x] `--color auto|always|never` decides; `auto` colors only on a terminal, and not with `NO_COLOR` or `TERM=dumb`
- [x] Piped output, `--json`, `--mermaid`, and the MCP tool have no color

## Out of scope

- Colors for other kinds of issues, such as available or closed ones.
- A configurable palette.

## Notes

2026-10-08: implemented. A canvas cell is bright if any bright drawing touches it, so on a line that arrows of other issues share, only the part that leads to or from the focus is bright. [[0007-color-the-focus-and-its-arrows]] records it.

## Resolution

Completed. sisyphus [[0.0.26]] on bookmark samw/ai/graph-directory-boxes, on top of 0.0.25; it waits for a human merge.
