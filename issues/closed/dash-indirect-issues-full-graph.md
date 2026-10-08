---
title: "Draw indirectly linked issues with dashed boxes in sisyphus graph --full"
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

# Draw indirectly linked issues with dashed boxes in sisyphus graph --full

## Summary

With `sisyphus graph --full`, draw each issue that is linked to the focus issue but is not below it or on its path above it in a dashed box, so a reader can tell it apart from the direct family.

## Context

From the human, 2026-10-08: with `--full`, give issues that are linked to the focus issue but are not below it or on its path above it their own outline. Agreed design: a light dashed box (`┌╌┐`, `╎`) for such an issue, and a heavy dashed box (`┏╍┓`, `╏`) for one that is also available. Unicode has no dashed corners, so corners stay solid. The two-dash pattern differs from the three-dash `┄`/`┆` of dependency arrows. `--mermaid` gives these issues a dashed class, and `--json` marks them.

## Acceptance criteria

- [x] With `--full`, an indirect issue has a light dashed box (`┌╌┐`), or a heavy dashed box (`┏╍┓`) if it is available
- [x] The legend lists each dashed style that the drawing uses
- [x] `--json` nodes have `indirect: true` for such issues; `--mermaid` gives them the dashed class `indirect`
- [x] Without `--full`, no issue is indirect and the drawing does not change
- [x] The graph help, the `sisyphus_graph` MCP description, and the graph docs describe the dashed boxes

## Out of scope

- ANSI color or faint text.

## Notes

2026-10-08: implemented. [[0005-dash-indirect-issues]] records it. A dashed box has solid corners, and a solid `├` or `┝` where an arrow leaves it, because Unicode has no dashed corners or joins.

## Resolution

Completed. sisyphus [[0.0.24]] on bookmark samw/ai/graph-directory-boxes, on top of 0.0.23; it waits for a human merge.
