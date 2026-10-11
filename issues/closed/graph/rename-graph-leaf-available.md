---
title: "Call ready-to-start issues \"available\" instead of \"leaf\" in sisyphus graph"
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

# Call ready-to-start issues "available" instead of "leaf" in sisyphus graph

## Summary

In `sisyphus graph`, use "available" instead of "leaf" for an issue that is ready to start, in the drawing, `--json`, `--mermaid`, help, and docs.

## Context

From the human, 2026-10-08: use "available" instead of "leaf" for an issue that is ready to start (not closed, with no open dependency or sub-issue). Since 0.0.22, the graph is laid out in work order, so these issues sit inside the drawing, not at an edge, and "leaf" reads wrong.

The term appears in the graph legend, the `--json` node field `leaf`, the Mermaid class `leaf`, the graph help, the `sisyphus_graph` MCP description, and docs/sisyphus-graph.md. As of 2026-10-08, nothing on this host reads the `leaf` JSON field.

## Acceptance criteria

- [x] The legend says `available (ready to start)`, and `(available)` after an available focus issue
- [x] `--json` nodes have `available` instead of `leaf`; `--mermaid` uses the class `available`
- [x] The graph help, the `sisyphus_graph` MCP description, and the graph docs use "available"
- [x] The rule that decides availability does not change

## Out of scope

- A `--json` alias for the old `leaf` field.

## Notes

2026-10-08: renamed in code, tests, help, MCP description, and docs. [[0004-call-ready-issues-available]] records it.

## Resolution

Completed. sisyphus [[0.0.23]] on bookmark samw/ai/graph-directory-boxes, on top of 0.0.22; it waits for a human merge.
