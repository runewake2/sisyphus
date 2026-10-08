---
title: "Draw each issue subdirectory as a box in sisyphus graph"
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

# Draw each issue subdirectory as a box in sisyphus graph

## Summary

In `sisyphus graph`, draw each issue subdirectory as a box that holds all issues below it, and label each node with only the last part of its name.

## Context

From the human, 2026-10-08: when sisyphus renders a graph, each subdirectory must be a box that contains all issues under it. Then each node needs only the last part of its name, and the drawing shows the grouping that the directory structure makes.

Issue subdirectories exist since 0.0.18 (bookmark samw/ai/issue-subdirectories, not merged). The graph code is cmd/sisyphus/graph.go on samw/ai/index-issues-once (0.0.19). drawGraph builds Mermaid source and renders it with github.com/AlexanderGrooff/mermaid-ascii.

mermaid-ascii parses `subgraph` blocks, nested too, and draws a box around them. But its layout does not know about subgraphs: it places nodes on a grid, then draws each box around the area of its members. So a box can enclose a node from another directory. Measured 2026-10-08 with small samples: nested boxes, and sibling boxes that an edge joins, draw correctly. A sibling box whose member is placed in another box's rows is drawn inside that box, and a box with no edges can draw over its own node.

## Acceptance criteria

- [x] Each directory of a drawn issue is a box, nested as the directories nest, and labeled with its last path segment
- [x] A box holds exactly the issues below its directory, for any links between directories
- [x] Sibling boxes do not overlap, and no box border is hidden by an arrow
- [x] A node shows only the file name of its issue; the legend still shows the full name of the focus issue
- [x] The focus issue and the leaves have their own box line styles instead of emoji marks (added 2026-10-08, from the human)
- [x] `--mermaid` output has the same subgraphs; `--json` does not change
- [x] A graph with no subdirectories draws no directory box. Its drawing changed with the new renderer, as [[0001-draw-graph-without-mermaid-ascii]] records

## Out of scope

- Collapsing a chain of directories that hold one subdirectory each into one box.

## Notes

2026-10-08: prototype on samw/ai/graph-directory-boxes. mermaidSource emits one nested `subgraph` per directory, and a node label keeps only the file name. Flat graphs (the sisyphus repo's own issues) draw byte-identical to 0.0.20.

2026-10-08: with stock mermaid-ascii, the drawing is wrong for 1 of 3 realistic graphs. The test was a copy of the sisyphus repo issues that was moved into `plugins/`, `plugins/github/`, `github/`, `github/actions/`, and `cli/`. In `graph integration-plugins --full`, the two `github/` issues are drawn inside the `plugins` box, and the top-level `github` box is not drawn. The cause is in mermaid-ascii createMapping: the grid placement ignores subgraphs. Then calculateSubgraphBoundingBox draws each box around its members' area. The pinned version (2026-09-08) is the latest.

2026-10-08: a scratch patch of mermaid-ascii gives each innermost subgraph its own band of grid rows, and sizes the gap row between bands for the borders and labels at that boundary. With it, every box holds exactly its own issues in the same test. A remaining defect: the edge router does not know about boxes, so an arrow can run along a box border and hide it.

Open question Q1 for the human: patch mermaid-ascii (fork plus an upstream PR), write a renderer in sisyphus, or draw boxes only in `--mermaid` output.

2026-10-08, from the human: Q1 is "own renderer in sisyphus". sisyphus lays out and draws the terminal graph itself, and stops using mermaid-ascii. `--mermaid` output does not need the library.

2026-10-08, from the human: mark the focus issue and the leaves with different box line styles, not emoji. Done as double (focus), heavy (leaf), and light (other) issue boxes, and rounded directory boxes. No box is both double and heavy, so a focus leaf is double and the legend adds "(a leaf)".

## Resolution

Completed. sisyphus [[0.0.21]] on bookmark samw/ai/graph-directory-boxes, stacked on samw/ai/slug-ascii; it waits for a human merge. sisyphus lays out and draws the terminal graph itself ([[0001-draw-graph-without-mermaid-ascii]]) and marks the focus and leaves with box styles ([[0002-mark-focus-and-leaves-with-box-styles]]). TestGraphDirectoryBoxHoldsOnlyItsIssues checks the case that mermaid-ascii drew wrong.
